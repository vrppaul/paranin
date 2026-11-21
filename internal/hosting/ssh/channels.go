package ssh

import (
	"context"
	"fmt"
	"net"
	"time"

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

// The kind of channel a session is, the requests in one that say the size of its terminal, and the one that asks
// for a subsystem (RFC 4254, 6.1, 6.2, 6.7 and 6.5)
const (
	sessionChannel       = "session"
	terminalRequestType  = "pty-req"
	windowChangeType     = "window-change"
	subsystemRequestType = "subsystem"
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

// The keys under which a connection's context says it has opened the channel of its session, and started the
// session, its shell or command taken
type (
	sessionOpened  struct{}
	sessionStarted struct{}
)

// limitingSessions has the server open one session a connection, with the size of its terminal bounded before
// charm.land/ssh, Bubble Tea or the TUI reads it, so none of them draws more; no channel of another kind is opened,
// as by default. Refusing a larger terminal would turn its reader away, and a bound in Bubble Tea's options or in the
// TUI would come after charm.land/ssh and wish hand the size on.
func limitingSessions(server *charmssh.Server) error {
	server.ChannelHandlers = map[string]charmssh.ChannelHandler{sessionChannel: openLimited}
	return nil
}

// openLimited opens a connection's first session as charm.land/ssh does, its client's requests bounded on their
// way, refuses every session after it, and closes the connection once its session ends: a reader needs one, and
// each costs the TUI's memory again. A client that shares one connection among its sessions (OpenSSH's ControlMaster)
// sees the second refused and opens a connection.
func openLimited(server *charmssh.Server, connection *gossh.ServerConn, channel gossh.NewChannel, ctx charmssh.Context) {
	if !isFirstSession(ctx) {
		_ = channel.Reject(gossh.ResourceShortage, secondSessionRefusal)
		return
	}
	session, end := context.WithCancel(ctx)
	defer end()
	charmssh.DefaultSessionHandler(server, connection, boundedChannel{NewChannel: channel}, sessionContext{Context: ctx, session: session})
	_ = connection.Close()
}

// sessionContext is a connection's context as one of its sessions sees it, which ends, its Done and Err, as the
// session's requests do, when its channel closes: charm.land/ssh hands a session its connection's, so wish would keep
// its TUI, and read its closed window changes without end, for as long as the connection stays open.
type sessionContext struct {
	charmssh.Context
	session context.Context
}

func (sessionCtx sessionContext) Done() <-chan struct{} {
	return sessionCtx.session.Done()
}

func (sessionCtx sessionContext) Err() error {
	return sessionCtx.session.Err() //nolint:wrapcheck // a context's Err is context.Canceled itself, as its contract asks
}

// isFirstSession tells whether a connection opens its first session, and marks its context as having opened it. The
// lock is the connection's own, which charm.land/ssh holds for limits per connection: its channels open at once. A
// session closed leaves its place taken: its TUI ends only after its requests have, so a place freed with them would
// let a client opening and closing sessions hold more than one.
func isFirstSession(ctx charmssh.Context) bool {
	ctx.Lock()
	defer ctx.Unlock()
	if ctx.Value(sessionOpened{}) != nil {
		return false
	}
	ctx.SetValue(sessionOpened{}, true)
	return true
}

// openingLimit is how long a connection may take from connecting to starting its session, an option of the server:
// past its handshake, charm.land/ssh holds a connection that starts none, with a channel for one or not, until it is
// idle too long.
type openingLimit time.Duration

// apply has the server watch each connection for the limit, from before its handshake, so the one limit bounds the
// handshake too, and mark each session as it starts.
func (limit openingLimit) apply(server *charmssh.Server) error {
	server.ConnCallback = limit.watch
	server.SessionRequestCallback = markStarted
	return nil
}

// watch closes a connection that has started no session once the limit has passed since it connected, and forgets
// it once the connection ends.
func (limit openingLimit) watch(ctx charmssh.Context, connection net.Conn) net.Conn {
	timer := time.AfterFunc(time.Duration(limit), func() { closeUnstarted(ctx, connection) })
	// a timer left to fire would hold the connection and its context for the whole limit, so a flood of short
	// connections would add them up
	context.AfterFunc(ctx, func() { timer.Stop() })
	return connection
}

// closeUnstarted closes a connection unless it has started its session.
func closeUnstarted(ctx charmssh.Context, connection net.Conn) {
	ctx.Lock()
	defer ctx.Unlock()
	if ctx.Value(sessionStarted{}) == nil {
		_ = connection.Close()
	}
}

// markStarted marks a session's connection as having started its session, its shell or command taken, and takes
// it: charm.land/ssh asks as each comes, the one moment a session starts. A subsystem is refused and starts
// nothing, since a kiosk serves none.
func markStarted(session charmssh.Session, requestType string) bool {
	if requestType == subsystemRequestType {
		return false
	}
	ctx := session.Context()
	ctx.Lock()
	defer ctx.Unlock()
	ctx.SetValue(sessionStarted{}, true)
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
