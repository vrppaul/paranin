package ssh

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"testing/synctest"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	gossh "golang.org/x/crypto/ssh"

	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/release"
)

// A terminal's size, and timeouts long enough never to end a test's session
const (
	terminalWidth  = 100
	terminalHeight = 30
	longTimeout    = time.Hour
	// more lines than a session logs, so the server never waits on the test
	logLines = 16
	// how long a test's connection lasts before the client gives it up, rather than hanging until go test does
	testDeadline = 10 * time.Second
	// how long a test waits for the server to log, well within the connection's deadline, so what closes a
	// session the test waits on is the server
	logWait = 3 * time.Second
)

// commonTerminal is the size of a reader's terminal, well within the bound.
func commonTerminal() terminalSize {
	return terminalSize{width: terminalWidth, height: terminalHeight}
}

// A session with a terminal reads the résumé: the places of its tabs reach the reader's screen.
func TestASessionWithATerminalReadsTheResume(test *testing.T) {
	address := serving(test, slog.New(slog.DiscardHandler))
	output, _ := openTerminal(test, clientSession(test, address, "reader"), commonTerminal())
	if read := readUntil(test, output, "contact"); !strings.Contains(read, "experience") {
		test.Errorf("the screen = %q, want the places of the résumé's tabs", read)
	}
}

// Terminals larger than the bound one way, within the most Bubble Tea takes a change of size to (4096 a side, about
// a million cells); and a terminal too narrow for the footer to name the last tab
const (
	widestChange  = 4096
	tallestChange = 2048
	narrowWidth   = 30
)

// What a session may allocate, in bytes, from its opening to its first screen: one drawn at the bound takes about
// 20 MB, one drawn a size larger than the bound one way more than 140 MB.
const sessionAllocation = 64 << 20

// A terminal larger than the bound, as it opens or as it grows, takes a session no more memory than the bound's,
// and its reader still reads the résumé. A narrow terminal opened first names the last tab only once it has grown.
func TestATerminalLargerThanTheBoundTakesTheBoundsMemory(test *testing.T) {
	narrow := terminalSize{width: narrowWidth, height: terminalHeight}
	wider := terminalSize{width: widestChange, height: tallestTerminal}
	taller := terminalSize{width: widestTerminal, height: tallestChange}
	cases := []struct {
		name    string
		opening terminalSize
		changes []terminalSize
	}{
		{name: "opening wider", opening: wider, changes: nil},
		{name: "opening taller", opening: taller, changes: nil},
		{name: "growing wider", opening: narrow, changes: []terminalSize{wider}},
		{name: "growing taller", opening: narrow, changes: []terminalSize{taller}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			address := serving(test, slog.New(slog.DiscardHandler))
			before := allocated()
			session := clientSession(test, address, "reader")
			output, _ := openTerminal(test, session, testCase.opening)
			for _, size := range testCase.changes {
				if err := session.WindowChange(size.height, size.width); err != nil {
					test.Fatalf("WindowChange(%+v) = %v, want the terminal grown", size, err)
				}
			}
			readUntil(test, output, "contact")
			if taken := allocated() - before; taken > sessionAllocation {
				test.Errorf("a terminal opening at %+v and changing to %+v took %d bytes to draw, want %d at most", testCase.opening,
					testCase.changes, taken, sessionAllocation)
			}
		})
	}
}

// strayTerminalRequest is a request for a terminal as RFC 4254 lays it out, with a byte after its end.
type strayTerminalRequest struct {
	Term         string
	Columns      uint32
	Rows         uint32
	WidthPixels  uint32
	HeightPixels uint32
	Modes        string
	Stray        uint8
}

// A request for a terminal whose size cannot be read is refused, and the session has no terminal, however much of
// the request charm.land/ssh would read: what it read would be the size unbounded.
func TestATerminalOfAnUnreadableSizeIsRefused(test *testing.T) {
	address := serving(test, slog.New(slog.DiscardHandler))
	session := clientSession(test, address, "reader")
	asked := strayTerminalRequest{
		Term: "xterm-256color", Columns: widestChange, Rows: tallestChange, WidthPixels: 0, HeightPixels: 0,
		Modes: "", Stray: 0,
	}
	if accepted, err := session.SendRequest(terminalRequestType, true, gossh.Marshal(&asked)); accepted || err != nil {
		test.Fatalf("a terminal request with a stray byte = %t, %v; want it refused", accepted, err)
	}
	if output, _ := session.CombinedOutput(""); !strings.Contains(string(output), "Requires an active PTY") {
		test.Errorf("the session after its terminal was refused = %q, want it refused for want of a terminal", output)
	}
}

// A connection opens one session: a second is refused while the first is open, so one connection cannot hold the
// TUI's memory many times over.
func TestASecondSessionOnAConnectionIsRefused(test *testing.T) {
	address := serving(test, slog.New(slog.DiscardHandler))
	client, first := sessionOver(test, dialed(test, address), address, "reader")
	output, _ := openTerminal(test, first, commonTerminal())
	readUntil(test, output, "contact")
	var refusal *gossh.OpenChannelError
	if _, err := client.NewSession(); !errors.As(err, &refusal) || refusal.Reason != gossh.ResourceShortage {
		test.Errorf("a second session on the connection = %v, want it refused for want of resources", err)
	}
}

// A session its reader closes leaves no goroutine behind once the server has stopped: the bubble fails on one still
// blocked as it ends.
func TestAClosedSessionLeavesNoGoroutineBehind(test *testing.T) {
	synctest.Test(test, func(test *testing.T) {
		server := newServer(test, slog.New(slog.DiscardHandler))
		clientEnd, serverEnd := memoryConnections()
		go server.server.HandleConn(serverEnd)
		client, session := sessionOver(test, clientEnd, memoryAddress, "reader")
		output, _ := openTerminal(test, session, commonTerminal())
		readUntil(test, output, "contact")
		closeClient(client, session)
		if err := server.Shutdown(test.Context()); err != nil {
			test.Errorf("Shutdown after the session closed = %v, want the server stopped", err)
		}
	})
}

// A session its reader closes ends there, its TUI with it, though its connection stays open: a client that keeps
// its connection, as OpenSSH's ControlPersist does, holds no TUI it has closed. It runs on a socket: a TUI left
// running once its session is closed keeps wish reading the session's closed window changes, which no synctest
// bubble sees blocked.
func TestASessionClosedOnAnOpenConnectionEnds(test *testing.T) {
	lines := make(chan string, logLines)
	address := serving(test, slog.New(slog.NewTextHandler(lineWriter{lines: lines}, nil)))
	session := clientSession(test, address, "reader")
	output, _ := openTerminal(test, session, commonTerminal())
	readUntil(test, output, "contact")
	nextLine(test, lines)
	if err := session.Close(); err != nil {
		test.Fatalf("Close of the session = %v, want it closed", err)
	}
	if closed := nextLine(test, lines); !strings.Contains(closed, "the session closed") {
		test.Errorf("the log after the session closed = %q, want the session closed with its connection open", closed)
	}
}

// allocated is how many bytes the process has allocated since it started, freed or not.
func allocated() uint64 {
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	return memory.TotalAlloc
}

// A stop whose wait runs out closes the sessions still open, and stops cleanly: a reader may keep a session open
// for as long as the server lets them. The server's end of the session is what closes, well before the client's
// own deadline would.
func TestAStopClosesTheSessionsStillOpen(test *testing.T) {
	lines := make(chan string, logLines)
	server, address := servingServer(test, slog.New(slog.NewTextHandler(lineWriter{lines: lines}, nil)))
	output, _ := openTerminal(test, clientSession(test, address, "reader"), commonTerminal())
	readUntil(test, output, "contact")
	ended, end := context.WithCancel(context.Background())
	end()
	if err := server.Shutdown(ended); err != nil {
		test.Errorf("Shutdown with the wait run out = %v, want the sessions closed and no error", err)
	}
	if opened, closed := nextLine(test, lines), nextLine(test, lines); !strings.Contains(opened, "the session opened") ||
		!strings.Contains(closed, "the session closed") {
		test.Errorf("the log = %q then %q, want the session opened, then closed by the stop", opened, closed)
	}
}

// The signal that stops the server leaves its sessions open: the server alone ends them, once it has waited for
// them, so no session's program may take the signal for its own.
func TestTheServersSignalLeavesItsSessionsOpen(test *testing.T) {
	address := serving(test, slog.New(slog.DiscardHandler))
	output, input := openTerminal(test, clientSession(test, address, "reader"), commonTerminal())
	readUntil(test, output, "contact")
	stops := make(chan os.Signal, 1)
	signal.Notify(stops, syscall.SIGTERM)
	defer signal.Stop(stops)
	if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
		test.Fatalf("send the stop signal: %v", err)
	}
	<-stops
	if _, err := io.WriteString(input, "l"); err != nil {
		test.Fatalf("press l = %v, want the session still open", err)
	}
	readUntil(test, output, "Kept the queue")
}

// A session opens the health report with :checkhealth: the server hands the TUI the extras beside the site's pages.
func TestASessionOpensTheHealthReport(test *testing.T) {
	address := serving(test, slog.New(slog.DiscardHandler))
	output, input := openTerminal(test, clientSession(test, address, "reader"), commonTerminal())
	readUntil(test, output, "contact")
	if _, err := io.WriteString(input, ":checkhealth\r"); err != nil {
		test.Fatalf("type :checkhealth = %v, want the session open", err)
	}
	readUntil(test, output, "no tests yet")
}

// A session with no terminal is refused, and a command sent with it is never run: the server is a kiosk, not a
// shell.
func TestASessionWithoutATerminalIsRefused(test *testing.T) {
	address := serving(test, slog.New(slog.DiscardHandler))
	output, err := clientSession(test, address, "reader").CombinedOutput("echo ran")
	if err == nil || !strings.Contains(string(output), "Requires an active PTY") || strings.Contains(string(output), "ran") {
		test.Errorf("a command with no terminal = %q, %v; want it refused, not run", output, err)
	}
}

// A session is logged as it opens and closes, with whose it was.
func TestASessionIsLoggedOpeningAndClosing(test *testing.T) {
	lines := make(chan string, logLines)
	address := serving(test, slog.New(slog.NewTextHandler(lineWriter{lines: lines}, nil)))
	_, _ = clientSession(test, address, "mira").CombinedOutput("")
	opened, closed := nextLine(test, lines), nextLine(test, lines)
	if !strings.Contains(opened, "the session opened") || !strings.Contains(closed, "the session closed") ||
		!strings.Contains(opened+closed, "user=mira") {
		test.Errorf("the log = %q then %q, want the session of mira opened and closed", opened, closed)
	}
}

// What the test's screens panic with, and the address of the session they are drawn for
const (
	breakage      = "the model broke"
	readerAddress = "203.0.113.7:52100"
)

// A panic of a session's TUI reaches the server's log, with whose session it was, what it panicked with and the
// function it panicked in, wherever the model panics: as it starts, as it takes a message, or as it draws a frame
// after one. The panic goes on up to Bubble Tea, which ends the program with it.
func TestAPanicOfASessionsScreenIsLogged(test *testing.T) {
	cases := []struct {
		name   string
		model  tea.Model
		source string
	}{
		{name: "as it starts", model: brokenInit{}, source: "brokenInit.Init"},
		{name: "as it takes a message", model: brokenUpdate{}, source: "brokenUpdate.Update"},
		{name: "drawing a frame after a message", model: calmModel{}, source: "brokenView.View"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			lines := make(chan string, logLines)
			logger := slog.New(slog.NewTextHandler(lineWriter{lines: lines}, nil))
			screen := watchedModel{logger: logger, model: testCase.model, remote: readerAddress}
			running, stop := context.WithTimeout(test.Context(), testDeadline)
			defer stop()
			program := tea.NewProgram(screen, tea.WithContext(running), tea.WithInput(nil), tea.WithOutput(io.Discard),
				tea.WithoutSignalHandler(), tea.WithWindowSize(terminalWidth, terminalHeight))
			if _, err := program.Run(); !errors.Is(err, tea.ErrProgramPanic) {
				test.Errorf("Run of a screen that panics %s = %v, want %v", testCase.name, err, tea.ErrProgramPanic)
			}
			if line := nextLine(test, lines); !strings.Contains(line, "remote="+readerAddress) || !strings.Contains(line, breakage) ||
				!strings.Contains(line, testCase.source) {
				test.Errorf("the log = %q, want the panic %q of the session from %s, and %s in its stack", line, breakage,
					readerAddress, testCase.source)
			}
		})
	}
}

// brokenInit panics as it starts.
type brokenInit struct{}

func (brokenInit) Init() tea.Cmd { panic(breakage) }

func (brokenInit) Update(tea.Msg) (tea.Model, tea.Cmd) { return brokenInit{}, nil }

func (brokenInit) View() tea.View { return tea.NewView("") }

// brokenUpdate panics as it takes the first message Bubble Tea sends it.
type brokenUpdate struct{}

func (brokenUpdate) Init() tea.Cmd { return nil }

func (brokenUpdate) Update(tea.Msg) (tea.Model, tea.Cmd) { panic(breakage) }

func (brokenUpdate) View() tea.View { return tea.NewView("") }

// calmModel draws its first frame, and becomes a model that panics drawing the next once it takes the first
// message Bubble Tea sends it.
type calmModel struct{}

func (calmModel) Init() tea.Cmd { return nil }

func (calmModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return brokenView{}, nil }

func (calmModel) View() tea.View { return tea.NewView("") }

// brokenView panics drawing its frame.
type brokenView struct{}

func (brokenView) Init() tea.Cmd { return nil }

func (brokenView) Update(tea.Msg) (tea.Model, tea.Cmd) { return brokenView{}, nil }

func (brokenView) View() tea.View { panic(breakage) }

// serving is the address of a server of the fixture, which serves until the test ends.
func serving(test *testing.T, logger *slog.Logger) string {
	_, address := servingServer(test, logger)
	return address
}

// servingServer is a server of the fixture, which serves until the test ends, and its address.
func servingServer(test *testing.T, logger *slog.Logger) (*Server, string) {
	server := newServer(test, logger)
	listener, err := (&net.ListenConfig{}).Listen(test.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		test.Fatalf("listen: %v", err)
	}
	go func() { _ = server.Serve(listener) }()
	test.Cleanup(func() { _ = server.Shutdown(context.Background()) })
	return server, listener.Addr().String()
}

// newServer is a server of the fixture, not yet serving.
func newServer(test *testing.T, logger *slog.Logger) *Server {
	store, err := release.Open(release.FoldersAt(fixtures.ContentFolder(test), fixtures.BuildFolder(test)))
	if err != nil {
		test.Fatalf("Open(fixtures) = %v, want no error", err)
	}
	settings := Settings{HostKeyPath: filepath.Join(test.TempDir(), "host_key"), IdleTimeout: longTimeout, MaxTimeout: longTimeout}
	server, err := New(store, logger, settings)
	if err != nil {
		test.Fatalf("New = %v, want a server", err)
	}
	return server
}

// terminalSize is the size a client says its terminal is.
type terminalSize struct {
	width  int
	height int
}

// openTerminal asks for a terminal of a size and the TUI in it: what the server writes to it, and what the reader
// types.
func openTerminal(test *testing.T, session *gossh.Session, size terminalSize) (io.Reader, io.Writer) {
	if err := session.RequestPty("xterm-256color", size.height, size.width, gossh.TerminalModes{}); err != nil {
		test.Fatalf("RequestPty(%+v) = %v, want a terminal", size, err)
	}
	output, err := session.StdoutPipe()
	if err != nil {
		test.Fatalf("StdoutPipe = %v", err)
	}
	input, err := session.StdinPipe()
	if err != nil {
		test.Fatalf("StdinPipe = %v", err)
	}
	if err := session.Shell(); err != nil {
		test.Fatalf("Shell = %v, want the TUI", err)
	}
	return output, input
}

// clientSession is a session of a reader of a name, who offers no key or password, as anyone may.
func clientSession(test *testing.T, address, user string) *gossh.Session {
	_, session := sessionOver(test, dialed(test, address), address, user)
	return session
}

// dialed is a connection to an address, which the client gives up before go test would.
func dialed(test *testing.T, address string) net.Conn {
	connection, err := (&net.Dialer{}).DialContext(test.Context(), "tcp", address)
	if err != nil {
		test.Fatalf("Dial = %v, want a connection", err)
	}
	if err := connection.SetDeadline(time.Now().Add(testDeadline)); err != nil {
		test.Fatalf("SetDeadline = %v", err)
	}
	return connection
}

// sessionOver is a client over a connection to an address, of a reader of a name who offers no key or password,
// and a session of it; both close as the test ends.
func sessionOver(test *testing.T, connection net.Conn, address, user string) (*gossh.Client, *gossh.Session) {
	config := &gossh.ClientConfig{User: user, Auth: nil, HostKeyCallback: gossh.InsecureIgnoreHostKey()} //nolint:gosec // a test's own server
	clientConnection, channels, requests, err := gossh.NewClientConn(connection, address, config)
	if err != nil {
		test.Fatalf("NewClientConn = %v, want an SSH connection", err)
	}
	client := gossh.NewClient(clientConnection, channels, requests)
	session, err := client.NewSession()
	if err != nil {
		test.Fatalf("NewSession = %v, want a session", err)
	}
	test.Cleanup(func() { closeClient(client, session) })
	return client, session
}

// closeClient closes a session and its client; a test that ends has no use for a fault in closing them.
func closeClient(client *gossh.Client, session *gossh.Session) {
	_ = session.Close()
	_ = client.Close()
}

// readUntil reads what the server writes until it holds a text, its colours gone, and is what was read.
func readUntil(test *testing.T, reader io.Reader, text string) string {
	var read strings.Builder
	chunk := make([]byte, 4096)
	for !strings.Contains(ansi.Strip(read.String()), text) {
		count, err := reader.Read(chunk)
		read.Write(chunk[:count])
		if err != nil {
			test.Fatalf("read = %v after %q, want %q", err, ansi.Strip(read.String()), text)
		}
	}
	return ansi.Strip(read.String())
}

// nextLine is the next line the server logs, or the test's end once the server has logged nothing for too long.
func nextLine(test *testing.T, lines <-chan string) string {
	select {
	case line := <-lines:
		return line
	case <-time.After(logWait):
		test.Fatalf("the server logged nothing for %s", logWait)
		return ""
	}
}

// lineWriter hands each line a logger writes to a channel, so a test reads the log as the server writes it.
type lineWriter struct {
	lines chan<- string
}

func (writer lineWriter) Write(line []byte) (int, error) {
	writer.lines <- string(line)
	return len(line), nil
}
