package ssh

import (
	"context"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	gossh "golang.org/x/crypto/ssh"

	"github.com/vrppaul/paranin/internal/content"
)

// The content package's fixture, a terminal's size, and timeouts long enough never to end a test's session
const (
	fixturePath    = "../../content/testdata/content"
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

// A session with a terminal reads the résumé: the site's tabs reach the reader's screen.
func TestASessionWithATerminalReadsTheResume(test *testing.T) {
	address := serving(test, slog.New(slog.DiscardHandler))
	output, _ := openTerminal(test, clientSession(test, address, "reader"))
	if read := readUntil(test, output, "Languages"); !strings.Contains(read, "Experience") {
		test.Errorf("the screen = %q, want the résumé's tabs", read)
	}
}

// A stop whose wait runs out closes the sessions still open, and stops cleanly: a reader may keep a session open
// for as long as the server lets them. The server's end of the session is what closes, well before the client's
// own deadline would.
func TestAStopClosesTheSessionsStillOpen(test *testing.T) {
	lines := make(chan string, logLines)
	server, address := servingServer(test, slog.New(slog.NewTextHandler(lineWriter{lines: lines}, nil)))
	output, _ := openTerminal(test, clientSession(test, address, "reader"))
	readUntil(test, output, "Languages")
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
	output, input := openTerminal(test, clientSession(test, address, "reader"))
	readUntil(test, output, "Languages")
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
	readUntil(test, output, "5 years")
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

// serving is the address of a server of the fixture, which serves until the test ends.
func serving(test *testing.T, logger *slog.Logger) string {
	_, address := servingServer(test, logger)
	return address
}

// servingServer is a server of the fixture, which serves until the test ends, and its address.
func servingServer(test *testing.T, logger *slog.Logger) (*Server, string) {
	store, err := content.Open(os.DirFS(fixturePath))
	if err != nil {
		test.Fatalf("Open(fixture) = %v, want no error", err)
	}
	settings := Settings{HostKeyPath: filepath.Join(test.TempDir(), "host_key"), IdleTimeout: longTimeout, MaxTimeout: longTimeout}
	server, err := New(store, logger, settings)
	if err != nil {
		test.Fatalf("New = %v, want a server", err)
	}
	listener, err := (&net.ListenConfig{}).Listen(test.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		test.Fatalf("listen: %v", err)
	}
	go func() { _ = server.Serve(listener) }()
	test.Cleanup(func() { _ = server.Shutdown(context.Background()) })
	return server, listener.Addr().String()
}

// openTerminal asks for a terminal and the TUI in it: what the server writes to it, and what the reader types.
func openTerminal(test *testing.T, session *gossh.Session) (io.Reader, io.Writer) {
	if err := session.RequestPty("xterm-256color", terminalHeight, terminalWidth, gossh.TerminalModes{}); err != nil {
		test.Fatalf("RequestPty = %v, want a terminal", err)
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
	config := &gossh.ClientConfig{User: user, Auth: nil, HostKeyCallback: gossh.InsecureIgnoreHostKey()} //nolint:gosec // a test's own server
	connection, err := (&net.Dialer{}).DialContext(test.Context(), "tcp", address)
	if err != nil {
		test.Fatalf("Dial = %v, want a connection", err)
	}
	if err := connection.SetDeadline(time.Now().Add(testDeadline)); err != nil {
		test.Fatalf("SetDeadline = %v", err)
	}
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
	return session
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
