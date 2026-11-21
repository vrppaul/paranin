// Package ssh serves the TUI over SSH (docs/REQUIREMENTS.md): a kiosk, not a shell. Anyone may connect, with no key
// and no password, since no way to authenticate is installed (FR-1); a session with a terminal reads the résumé,
// the projects' pages and the TUI's extras, and nothing else is served: no shell, no command, no file transfer, no
// forwarding (SEC-1, SEC-2). A session with no terminal is refused (SEC-3), a panic ends its own session and never
// the server (SEC-4), a session idle too long or open too long is closed (SEC-5), a connection opens one session
// and a terminal larger than a bound is drawn at the bound, so no connection's screens take memory without end
// (SEC-7), and each session is logged as it opens and closes (SEC-13), what its reader sent quoted by the logger's
// handler, as slog's handlers quote a value with a control character in it (SEC-15).
//
// A panic in the TUI's Init, Update or View is logged with its session and its stack, and goes on up to Bubble Tea,
// which ends the session and gives the reader's terminal back as it found it. Bubble Tea and wish take no logger:
// Bubble Tea still writes the panic and its stack to standard error, and wish a line of the error the program ended
// with; a panic in a command Bubble Tea runs reaches only those (TODO.md). The server alone ends its sessions: no
// session's program answers the server's own signals, so a stop waits for the sessions open before it closes them.
//
// The documents are composed once for each snapshot the store holds (release.Derived), from its content alone, and
// each session draws the snapshot current when it opened; a reload reaches the sessions that open after it.
package ssh

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"runtime/debug"
	"time"

	tea "charm.land/bubbletea/v2"
	charmssh "charm.land/ssh"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	"charm.land/wish/v2/bubbletea"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
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
)

// Settings are what the server is run with: where its host key lives, and how long a session may stay idle and
// stay open.
type Settings struct {
	// HostKeyPath is the file of the server's key, made there on the first start when it is missing (SEC-10)
	HostKeyPath string
	IdleTimeout time.Duration
	MaxTimeout  time.Duration
}

// Server serves the résumé, and the projects' pages a reader opens from it, to each session with a terminal.
type Server struct {
	server *charmssh.Server
}

// drawing is what each session of one snapshot draws: the documents a reader reads, the résumé and the hidden pages
// among them, and the site that writes their addresses.
type drawing struct {
	documents []document.Document
	site      routes.Site
}

// sessions open each session's TUI on what the current snapshot draws, its panics logged through the logger.
type sessions struct {
	logger *slog.Logger
	drawn  *release.Derived[drawing]
}

// guard logs each session and ends one that panics.
type guard struct {
	logger *slog.Logger
}

// New is a server of the site the store holds, logging through the logger.
func New(store *release.Store, logger *slog.Logger, settings Settings) (*Server, error) {
	opening := sessions{logger: logger, drawn: release.Derive(store, draw)}
	watching := guard{logger: logger}
	// wish runs the last listed outermost: the log, then the guard against panics, then the refusal of a
	// session with no terminal, then the TUI
	server, err := wish.NewServer(
		wish.WithHostKeyPath(settings.HostKeyPath),
		wish.WithIdleTimeout(settings.IdleTimeout),
		wish.WithMaxTimeout(settings.MaxTimeout),
		limitingSessions,
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

// draw is what the sessions of a snapshot draw: the pages of its content, and the extras the TUI opens by a command.
func draw(snapshot *release.Snapshot) drawing {
	facts := snapshot.Content
	documents := append(compose.Documents(facts), compose.Extras(facts)...)
	return drawing{documents: documents, site: routes.New(facts.Profile.Site.BaseURL)}
}

// open is a session's TUI, of the size its terminal said it is.
func (opening sessions) open(session charmssh.Session) (tea.Model, []tea.ProgramOption) {
	terminal, _, _ := session.Pty()
	drawn := opening.drawn.Current()
	model := watchedModel{
		logger: opening.logger,
		model:  tui.New(drawn.documents, drawn.site, terminal.Window.Width, terminal.Window.Height, time.Now()),
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
	remote, user := session.RemoteAddr().String(), session.User()
	logged.logger.Info("the session opened", remoteKey, remote, userKey, user)
	logged.next(session)
	logged.logger.Info("the session closed", remoteKey, remote, userKey, user, lastedKey, time.Since(opened))
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
