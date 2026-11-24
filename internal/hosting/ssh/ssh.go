// Package ssh serves the TUI over SSH (docs/REQUIREMENTS.md): a kiosk, not a shell. Anyone may connect, with no key
// and no password, since no way to authenticate is installed (FR-1); a session with a terminal reads the résumé,
// the projects' pages, the TUI's extras and the blog, opening on the blog's page its user's name reaches (D26), and
// nothing else is served: no shell, no command, no file transfer, no forwarding (SEC-1, SEC-2). A session with no
// terminal is refused (SEC-3), a panic ends its own session and never the server (SEC-4), a connection too long in
// starting its session, its handshake with it, and a session idle too long or open too long are closed (SEC-5), a
// connection opens one session, a terminal larger than a bound is drawn at the bound, a session's input reaches the
// TUI at a reader's pace and its environment holds only what the colour profile reads, so no connection takes memory
// or time without end (SEC-7), and each session is logged as it opens and closes, and its refused requests once
// (SEC-13), what its reader sent cut short and quoted by the logger's handler, as slog's handlers quote a value with a
// control character in it (SEC-15).
//
// A panic in the TUI's Init, Update or View is logged with its session and its stack, and goes on up to Bubble Tea,
// which ends the session and gives the reader's terminal back as it found it. Bubble Tea and wish take no logger:
// Bubble Tea still writes the panic and its stack to standard error, and wish a line of the error the program ended
// with; a panic in a command Bubble Tea runs reaches only those (TODO.md). The server alone ends its sessions: no
// session's program answers the server's own signals, so a stop waits for the sessions open before it closes them.
//
// The documents are composed once for each snapshot the store holds (release.Derived), from its content alone, its
// posts' code split once (compose.Prepare), and again as a post dated later comes out; each session draws the site as
// it stood when it opened, so a reload, or a post coming out, reaches the sessions that open after it.
package ssh

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"runtime/debug"
	"slices"
	"time"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	charmssh "charm.land/ssh"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	"charm.land/wish/v2/bubbletea"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/face/tui"
	"github.com/vrppaul/paranin/internal/release"
	"github.com/vrppaul/paranin/internal/routes"
)

// The keys of what the log says of a session
const (
	remoteKey = "remote"
	userKey   = "user"
	lastedKey = "lasted"
	panicKey  = "panic"
	stackKey  = "stack"
	// how many requests the session refused for each reason: a payload it could not read, a variable of the
	// environment the colour profile never reads, one set again, and one too long
	unreadableKey     = "unreadable"
	unreadVariableKey = "unread_variable"
	setAlreadyKey     = "set_already"
	tooLongKey        = "too_long"
)

// The most of a client's text the log holds, in bytes, and the mark where it is cut: twice the longest user name a
// system takes, so a name is whole and a long one costs the log no more
const (
	longestLoggedText = 64
	textCut           = "…"
)

// Settings are what the server is run with: where its host key lives, how long a connection may take to start its
// session from when it connects, its handshake within that, how long a session may stay idle and stay open, and
// whether its sessions read the drafts.
type Settings struct {
	// HostKeyPath is the file of the server's key, made there on the first start when it is missing (SEC-10)
	HostKeyPath    string
	OpeningTimeout time.Duration
	IdleTimeout    time.Duration
	MaxTimeout     time.Duration
	// Drafts is whether the sessions read the drafts too, as their author runs the site to read them
	Drafts bool
}

// Server serves the résumé, and the projects' pages a reader opens from it, to each session with a terminal.
type Server struct {
	server *charmssh.Server
}

// sessions open each session's TUI on what the current snapshot draws, its panics logged through the logger.
type sessions struct {
	logger *slog.Logger
	drawn  *release.Derived[tui.Drawing]
}

// guard logs each session and ends one that panics.
type guard struct {
	logger *slog.Logger
}

// New is a server of the site the store holds, logging through the logger.
func New(store *release.Store, logger *slog.Logger, settings Settings) (*Server, error) {
	prepared := release.Derive(store, func(snapshot *release.Snapshot, _ time.Time) (*compose.Prepared, time.Time) {
		// the zero time: it lapses with its snapshot alone (release.Work)
		return compose.Prepare(snapshot.Content), time.Time{}
	})
	drawn := release.Derive(store, func(snapshot *release.Snapshot, moment time.Time) (tui.Drawing, time.Time) {
		return draw(snapshot, prepared.Of(snapshot, moment), compose.Audience{At: moment, Drafts: settings.Drafts})
	})
	opening := sessions{logger: logger, drawn: drawn}
	watching := guard{logger: logger}
	// wish runs the last listed outermost: the log, then the guard against panics, then the refusal of a
	// session with no terminal, then the TUI
	server, err := wish.NewServer(
		wish.WithHostKeyPath(settings.HostKeyPath),
		wish.WithIdleTimeout(settings.IdleTimeout),
		wish.WithMaxTimeout(settings.MaxTimeout),
		openingLimit(settings.OpeningTimeout).apply,
		sessionLimits{logger: logger}.apply,
		wish.WithMiddleware(
			bubbletea.Middleware(opening.open),
			activeterm.Middleware(),
			watching.recovering,
			watching.logging,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("make the SSH server with the host key %q: %w", settings.HostKeyPath, err)
	}
	return &Server{server: server}, nil
}

// Serve serves the sessions that come to a listener until the server shuts down.
func (server *Server) Serve(listener net.Listener) error {
	if err := server.server.Serve(listener); err != nil {
		return fmt.Errorf("serve SSH at %s: %w", listener.Addr(), err)
	}
	return nil
}

// Shutdown stops taking sessions and waits for the open ones to end; once the context ends first, it closes the
// sessions still open and stops cleanly all the same: a reader may keep a session open for as long as the server
// lets them, so one still open at a stop is no fault, where a request still unanswered is one for the web.
func (server *Server) Shutdown(ctx context.Context) error {
	err := server.server.Shutdown(ctx)
	if err == nil {
		return nil
	}
	if ctx.Err() == nil {
		return fmt.Errorf("shut the SSH server down: %w", err)
	}
	if err := server.server.Close(); err != nil {
		return fmt.Errorf("close the sessions still open: %w", err)
	}
	return nil
}

// draw is what the sessions of a snapshot draw as an audience reads it: the pages of its content, the extras the TUI
// opens by a command, and the blog's lists, filtered or not, posts and their older versions, with the site's words of
// the frame round them; and until when it holds, as the next post dated later comes out. Prepare read the snapshot's
// content.
func draw(snapshot *release.Snapshot, prepared *compose.Prepared, audience compose.Audience) (tui.Drawing, time.Time) {
	facts := snapshot.Content
	composition := compose.SiteOf(prepared, audience)
	documents := slices.Concat(composition.Documents(), compose.Extras(facts), composition.Lists, composition.Filtered, composition.Posts,
		composition.Versions)
	return tui.Drawing{Documents: documents, Frame: composition.Frame, Site: routes.New(facts.Profile.Site.BaseURL, facts.Profile.Site.TagAuthority)}, composition.Until
}

// open is a session's TUI, of the size its terminal said it is, opened on the page its user's name reaches (D26).
func (opening sessions) open(session charmssh.Session) (tea.Model, []tea.ProgramOption) {
	terminal, _, _ := session.Pty()
	opened := time.Now()
	model := watchedModel{
		logger: opening.logger,
		model:  tui.New(opening.drawn.Current(opened), session.User(), terminal.Window.Width, terminal.Window.Height, opened),
		remote: session.RemoteAddr().String(),
	}
	return model, []tea.ProgramOption{tea.WithoutSignalHandler()}
}

// watchedModel is a session's TUI whose panics reach the server's log, with the session's address and the stack
// they came up, on their way to Bubble Tea, which ends the session and gives its terminal back.
type watchedModel struct {
	logger *slog.Logger
	model  tea.Model
	remote string
}

func (watched watchedModel) Init() tea.Cmd {
	defer watched.passPanicOn()
	return watched.model.Init()
}

func (watched watchedModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	defer watched.passPanicOn()
	next, command := watched.model.Update(message)
	watched.model = next
	return watched, command
}

func (watched watchedModel) View() tea.View {
	defer watched.passPanicOn()
	return watched.model.View()
}

// passPanicOn logs a panic of the TUI and passes it on to Bubble Tea, whose recover is in place before Init runs.
func (watched watchedModel) passPanicOn() {
	if cause := recover(); cause != nil {
		logPanic(watched.logger, watched.remote, cause)
		panic(cause) //nolint:forbidigo // one already caught, passed on to the recovery that ends the session
	}
}

// logPanic logs a panic of the session from an address, with its stack. A deferred call reads the stack before
// the panic has left the frames it came from, so the stack still holds them.
func logPanic(logger *slog.Logger, remote string, cause any) {
	logger.Error("a session panicked", remoteKey, remote, panicKey, fmt.Sprint(cause), stackKey, string(debug.Stack()))
}

// logging logs a session as it opens and as it closes, with how long it lasted.
func (watching guard) logging(next charmssh.Handler) charmssh.Handler {
	return loggedSession{guard: watching, next: next}.serve
}

// recovering ends a session that panics, logging the panic, and leaves the server and every other session be.
func (watching guard) recovering(next charmssh.Handler) charmssh.Handler {
	return recoveredSession{guard: watching, next: next}.serve
}

// loggedSession is a session handler whose sessions are logged.
type loggedSession struct {
	guard
	next charmssh.Handler
}

func (logged loggedSession) serve(session charmssh.Session) {
	opened := time.Now()
	remote, user := session.RemoteAddr().String(), clientText(session.User())
	logged.logger.Info("the session opened", remoteKey, remote, userKey, user)
	logged.next(session)
	logged.logger.Info("the session closed", remoteKey, remote, userKey, user, lastedKey, time.Since(opened))
}

// clientText is a text a client sent, which reaches the log cut to the most of it the log holds, and quoted there
// by the logger's handler.
type clientText string

func (text clientText) LogValue() slog.Value {
	if len(text) <= longestLoggedText {
		return slog.StringValue(string(text))
	}
	cut := longestLoggedText
	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}
	return slog.StringValue(string(text[:cut]) + textCut)
}

// recoveredSession is a session handler whose panics end its own session alone.
type recoveredSession struct {
	guard
	next charmssh.Handler
}

func (recovered recoveredSession) serve(session charmssh.Session) {
	defer recovered.catch(session)
	recovered.next(session)
}

// catch logs a panic of the session, which has ended it.
func (recovered recoveredSession) catch(session charmssh.Session) {
	if cause := recover(); cause != nil {
		logPanic(recovered.logger, session.RemoteAddr().String(), cause)
	}
}
