package ssh

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"slices"
	"time"

	charmssh "charm.land/ssh"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/time/rate"
)

// The largest terminal a session is drawn for, in columns and rows: a cell drawn costs the server about 240 bytes, so
// a session takes about 20 MB at most, and a 4K screen's terminal at a common font fits whole. A larger terminal is
// drawn at this size in its top left corner rather than refused, so its reader still reads the site.
const (
	widestTerminal  = 400
	tallestTerminal = 150
)

// How fast a session's input reaches the TUI, in bytes a second, and how much of it may come at once. A reader
// typing, or holding a key down, sends under two hundred bytes a second, and an ordinary paste, or the terminal's
// answers to the TUI's questions as it opens, a few kilobytes at most at once. A step of the wheel reports ten to
// fifteen bytes, so the burst carries about three hundred steps, and the wheel then goes on at about fifteen a second.
// On the largest terminal the costliest key, a letter typed in a search of a long post, costs the TUI 2.8 ms, so a
// client flooding at this pace spends up to half a core, and a paste it leaves open grows by under a megabyte an hour.
const (
	inputRate  = 192
	inputBurst = 4 << 10
)

// The kind of channel a session is, the requests in one that say the size of its terminal, the one that asks for a
// subsystem, and the one that sets a variable of its environment (RFC 4254, 6.1, 6.2, 6.7, 6.5 and 6.4)
const (
	sessionChannel       = "session"
	terminalRequestType  = "pty-req"
	windowChangeType     = "window-change"
	subsystemRequestType = "subsystem"
	environmentType      = "env"
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

// sessionLimits are the limits of a connection's sessions, and the log their refusals reach.
type sessionLimits struct {
	logger *slog.Logger
}

// apply has the server open one session a connection, with the size of its terminal bounded before
// charm.land/ssh, Bubble Tea or the TUI reads it, so none of them draws more; no channel of another kind is opened,
// as by default. Refusing a larger terminal would turn its reader away, and a bound in Bubble Tea's options or in the
// TUI would come after charm.land/ssh and wish hand the size on.
func (limits sessionLimits) apply(server *charmssh.Server) error {
	server.ChannelHandlers = map[string]charmssh.ChannelHandler{sessionChannel: limits.open}
	return nil
}

// open opens a connection's first session as charm.land/ssh does, its client's requests checked on their way,
// refuses every session after it, and closes the connection once its session ends: a reader needs one, and each
// costs the TUI's memory again. A client that shares one connection among its sessions (OpenSSH's ControlMaster)
// sees the second refused and opens a connection.
func (limits sessionLimits) open(server *charmssh.Server, connection *gossh.ServerConn, channel gossh.NewChannel, ctx charmssh.Context) {
	if !isFirstSession(ctx) {
		_ = channel.Reject(gossh.ResourceShortage, secondSessionRefusal)
		return
	}
	session, end := context.WithCancel(ctx)
	defer end()
	refusals := refusalLog{logger: limits.logger, remote: connection.RemoteAddr().String(), user: clientText(connection.User())}
	checked := checkedChannel{NewChannel: channel, session: session, refusals: refusals}
	charmssh.DefaultSessionHandler(server, connection, checked, sessionContext{Context: ctx, session: session})
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

// checkedChannel is a session's channel whose requests and input reach charm.land/ssh checked: the size of the
// terminal bounded, the environment taken only as far as the colour profile reads it, the input paced for as long as
// the session lasts.
type checkedChannel struct {
	gossh.NewChannel
	session  context.Context
	refusals refusalLog
}

func (channel checkedChannel) Accept() (gossh.Channel, <-chan *gossh.Request, error) {
	accepted, requests, err := channel.NewChannel.Accept()
	if err != nil {
		return nil, nil, fmt.Errorf("accept a session's channel: %w", err)
	}
	checked := make(chan *gossh.Request)
	go passChecked(requests, checked, channel.refusals)
	paced := pacedChannel{Channel: accepted, pace: rate.NewLimiter(inputRate, inputBurst), session: channel.session}
	return paced, checked, nil
}

// passChecked passes each request of a session on, the size of its terminal bounded, and closes the requests passed
// on once the client's requests end. One refused is answered so here and never passed on: charm.land/ssh would read
// a size it cannot read whole as far as it could, unbounded, and keep every variable of the environment it was sent.
// The refusals reach the log once, as the requests end, so a flood of them writes one line.
func passChecked(requests <-chan *gossh.Request, checked chan<- *gossh.Request, refusals refusalLog) {
	defer close(checked)
	check := requestCheck{variablesTaken: map[string]bool{}, refused: refusalCounts{unreadable: 0, unread: 0, setAlready: 0, tooLong: 0}}
	for request := range requests {
		if refused := check.checked(request); refused != notRefused {
			check.refused.add(refused)
			_ = request.Reply(false, nil)
			continue
		}
		checked <- request
	}
	refusals.write(check.refused)
}

// refusal is whether a session refuses a request, and why.
type refusal int

const (
	// notRefused is a request passed on
	notRefused refusal = iota
	// unreadablePayload is a request whose payload cannot be read whole
	unreadablePayload
	// unreadVariable sets a variable of the environment the colour profile never reads
	unreadVariable
	// variableSetAlready sets a variable of the environment taken before
	variableSetAlready
	// valueTooLong sets a variable to a value longer than any the colour profile reads
	valueTooLong
)

// requestCheck is what a session's requests are checked against, the variables of its environment taken so far,
// and how many it has refused.
type requestCheck struct {
	variablesTaken map[string]bool
	refused        refusalCounts
}

// refusalCounts are how many requests a session refused, by why.
type refusalCounts struct {
	unreadable int
	unread     int
	setAlready int
	tooLong    int
}

// add counts a refusal.
func (counts *refusalCounts) add(refused refusal) {
	switch refused {
	case notRefused:
	case unreadablePayload:
		counts.unreadable++
	case unreadVariable:
		counts.unread++
	case variableSetAlready:
		counts.setAlready++
	case valueTooLong:
		counts.tooLong++
	}
}

// refusalLog is the log a session's refused requests reach, and whose session they were.
type refusalLog struct {
	logger *slog.Logger
	remote string
	user   clientText
}

// write logs how many requests a session refused, by why, if it refused any.
func (refusals refusalLog) write(counts refusalCounts) {
	if counts != (refusalCounts{unreadable: 0, unread: 0, setAlready: 0, tooLong: 0}) {
		refusals.logger.Warn("the session refused requests", remoteKey, refusals.remote, userKey, refusals.user,
			unreadableKey, counts.unreadable, unreadVariableKey, counts.unread, setAlreadyKey, counts.setAlready,
			tooLongKey, counts.tooLong)
	}
}

// checked bounds the size of the terminal a request says, if it says one, and refuses a variable of the environment
// the session does not take.
func (check *requestCheck) checked(request *gossh.Request) refusal {
	switch request.Type {
	case terminalRequestType:
		return boundPayload(request, new(terminalAsked))
	case windowChangeType:
		return boundPayload(request, new(windowChange))
	case environmentType:
		return check.takenVariable(request)
	}
	return notRefused
}

// takenVariable takes a variable of the environment the colour profile reads, once and of a value no longer than it
// reads, and refuses any other: wish hands the colour profile the session's environment, and nothing else of the
// site reads it.
func (check *requestCheck) takenVariable(request *gossh.Request) refusal {
	var asked environmentVariable
	if gossh.Unmarshal(request.Payload, &asked) != nil {
		return unreadablePayload
	}
	switch {
	case !slices.Contains(colourVariables(), asked.Name):
		return unreadVariable
	case check.variablesTaken[asked.Name]:
		return variableSetAlready
	case len(asked.Value) > longestColourValue:
		return valueTooLong
	}
	check.variablesTaken[asked.Name] = true
	return notRefused
}

// environmentVariable is a variable a client sets in its session's environment, as RFC 4254 lays it out.
type environmentVariable struct {
	Name  string
	Value string
}

// colourVariables are the variables of the environment charmbracelet/colorprofile reads to choose a terminal's
// colours from its environment, TERM aside, which wish takes from the request for the terminal.
func colourVariables() []string {
	return []string{"NO_COLOR", "CLICOLOR", "CLICOLOR_FORCE", "COLORTERM", "WT_SESSION", "GOOGLE_CLOUD_SHELL"}
}

// The longest value of a variable of the environment taken: the longest the colour profile reads, Windows
// Terminal's session, is a GUID of 36 letters
const longestColourValue = 64

// pacedChannel is a session's channel whose input reaches charm.land/ssh at a reader's pace, so a flood of it, keys
// or a paste left open, waits in the channel's window, which x/crypto bounds, and never in the TUI.
type pacedChannel struct {
	gossh.Channel
	pace    *rate.Limiter
	session context.Context
}

func (channel pacedChannel) Read(buffer []byte) (int, error) {
	count, err := channel.Channel.Read(buffer[:min(len(buffer), inputBurst)])
	if waited := channel.pace.WaitN(channel.session, count); waited != nil {
		return count, fmt.Errorf("wait for the pace of a session's input: %w", waited)
	}
	return count, err //nolint:wrapcheck // a reader's io.EOF stays io.EOF, as its readers compare it
}

// boundPayload reads a request's payload into its layout, bounds the size in it and writes it back, and refuses one
// it cannot read whole.
func boundPayload(request *gossh.Request, payload sizedPayload) refusal {
	if gossh.Unmarshal(request.Payload, payload) != nil {
		return unreadablePayload
	}
	payload.bound()
	request.Payload = gossh.Marshal(payload)
	return notRefused
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
