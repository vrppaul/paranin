package ssh

import (
	"fmt"

	charmssh "charm.land/ssh"
	gossh "golang.org/x/crypto/ssh"
)

// The largest terminal a session is drawn for, in columns and rows: a cell drawn costs the server about 240 bytes, so
// a session takes about 20 MB at most, and a 4K screen's terminal at a common font fits whole. A larger terminal is
// drawn at this size in its top left corner rather than refused, so its reader still reads the site.
const (
	widestTerminal  = 400
	tallestTerminal = 150
)

// The kind of channel a session is, and the requests in one that say the size of its terminal (RFC 4254, 6.1, 6.2
// and 6.7)
const (
	sessionChannel      = "session"
	terminalRequestType = "pty-req"
	windowChangeType    = "window-change"
)

// terminalAsked is what a client asks for with a terminal, as RFC 4254 lays it out.
type terminalAsked struct {
	Term         string
	Columns      uint32
	Rows         uint32
	WidthPixels  uint32
	HeightPixels uint32
	Modes        string
}

// windowChange is the size a client's terminal has changed to, as RFC 4254 lays it out.
type windowChange struct {
	Columns      uint32
	Rows         uint32
	WidthPixels  uint32
	HeightPixels uint32
}

// sizedPayload is a request's payload that holds the size of a terminal, which it bounds in place.
type sizedPayload interface {
	bound()
}

// Why a second session on a connection is refused, as the client is told
const secondSessionRefusal = "one session a connection"

// sessionOpened is the key under which a connection's context says it has opened its session.
type sessionOpened struct{}

// limitingSessions has the server open one session a connection, with the size of its terminal bounded before
// charm.land/ssh, Bubble Tea or the TUI reads it, so none of them draws more; no channel of another kind is opened,
// as by default. Refusing a larger terminal would turn its reader away, and a bound in Bubble Tea's options or in the
// TUI would come after charm.land/ssh and wish hand the size on.
func limitingSessions(server *charmssh.Server) error {
	server.ChannelHandlers = map[string]charmssh.ChannelHandler{sessionChannel: openLimited}
	return nil
}

// openLimited opens a connection's first session as charm.land/ssh does, its client's requests bounded on their
// way, and refuses every session after it: a reader needs one, and each costs the TUI's memory again. A client that
// shares one connection among its sessions (OpenSSH's ControlMaster) sees the second refused and opens a connection.
func openLimited(server *charmssh.Server, connection *gossh.ServerConn, channel gossh.NewChannel, ctx charmssh.Context) {
	if !isFirstSession(ctx) {
		_ = channel.Reject(gossh.ResourceShortage, secondSessionRefusal)
		return
	}
	charmssh.DefaultSessionHandler(server, connection, boundedChannel{NewChannel: channel}, ctx)
}

// isFirstSession tells whether a connection opens its first session, and marks its context as having opened it. The
// lock is the connection's own, which charm.land/ssh holds for limits per connection: its channels open at once. A
// session closed leaves its place taken, since its TUI goes on until the connection ends.
func isFirstSession(ctx charmssh.Context) bool {
	ctx.Lock()
	defer ctx.Unlock()
	if ctx.Value(sessionOpened{}) != nil {
		return false
	}
	ctx.SetValue(sessionOpened{}, true)
	return true
}

// boundedChannel is a session's channel whose requests reach charm.land/ssh with the size of the terminal bounded.
type boundedChannel struct {
	gossh.NewChannel
}

func (channel boundedChannel) Accept() (gossh.Channel, <-chan *gossh.Request, error) {
	accepted, requests, err := channel.NewChannel.Accept()
	if err != nil {
		return nil, nil, fmt.Errorf("accept a session's channel: %w", err)
	}
	bounded := make(chan *gossh.Request)
	go passBounded(requests, bounded)
	return accepted, bounded, nil
}

// passBounded passes each request of a session on with the size of its terminal bounded, and closes the requests
// passed on once the client's requests end. One whose size cannot be read is refused: charm.land/ssh reads some of
// those still, unbounded.
func passBounded(requests <-chan *gossh.Request, bounded chan<- *gossh.Request) {
	defer close(bounded)
	for request := range requests {
		if err := boundTerminal(request); err != nil {
			_ = request.Reply(false, nil)
			continue
		}
		bounded <- request
	}
}

// boundTerminal bounds the size of the terminal a request says, if it says one.
func boundTerminal(request *gossh.Request) error {
	switch request.Type {
	case terminalRequestType:
		return boundPayload(request, new(terminalAsked))
	case windowChangeType:
		return boundPayload(request, new(windowChange))
	}
	return nil
}

// boundPayload reads a request's payload into its layout, bounds the size in it and writes it back.
func boundPayload(request *gossh.Request, payload sizedPayload) error {
	if err := gossh.Unmarshal(request.Payload, payload); err != nil {
		return fmt.Errorf("read the terminal's size a %q request says: %w", request.Type, err)
	}
	payload.bound()
	request.Payload = gossh.Marshal(payload)
	return nil
}

func (asked *terminalAsked) bound() {
	asked.Columns, asked.Rows = boundedSize(asked.Columns, asked.Rows)
}

func (changed *windowChange) bound() {
	changed.Columns, changed.Rows = boundedSize(changed.Columns, changed.Rows)
}

// boundedSize is a terminal's columns and rows, each no more than the bound.
func boundedSize(columns, rows uint32) (uint32, uint32) {
	return min(columns, widestTerminal), min(rows, tallestTerminal)
}
