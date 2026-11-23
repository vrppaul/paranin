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

	"github.com/vrppaul/paranin/internal/compose"
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
	// how long a test's connection may take to start its session, ample on a busy machine and well within the
	// connection's deadline
	openingTimeout = 6 * time.Second
	// how soon after its session a connection is closed, on a synctest bubble's clock, which moves only on the TUI's
	// frames while the session ends
	sessionEndTolerance = time.Second
	// the subsystem a client asks for to move files
	sftpSubsystem = "sftp"
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

// variable is a variable a client sets in its session's environment, as RFC 4254 lays it out.
type variable struct {
	Name  string
	Value string
}

// A byte where a variable's payload should start with the length of its name, too short to hold one
const strayByte = 0xff

// A variable a client sets in its session's environment, as OpenSSH's SendEnv does, is taken only if the colour
// profile reads it, once, and with a value no longer than the profile reads; any other is refused, and one whose
// payload cannot be read, since charm.land/ssh would keep every one it was sent.
func TestOnlyTheColourVariablesOfTheEnvironmentAreTaken(test *testing.T) {
	noColour := variable{Name: "NO_COLOR", Value: "1"}
	cases := []struct {
		name    string
		before  []variable
		payload []byte
		refused bool
	}{
		{name: "a variable of colour", before: nil, payload: gossh.Marshal(&noColour), refused: false},
		{name: "a variable of no colour", before: nil, payload: gossh.Marshal(&variable{Name: "LANG", Value: "C.UTF-8"}), refused: true},
		{
			name: "a variable of colour again", before: []variable{noColour},
			payload: gossh.Marshal(&variable{Name: "NO_COLOR", Value: "0"}), refused: true,
		},
		{
			name: "a variable of colour too long", before: nil,
			payload: gossh.Marshal(&variable{Name: "COLORTERM", Value: strings.Repeat(floodKey, longestColourValue+1)}), refused: true,
		},
		{name: "a payload that cannot be read", before: nil, payload: []byte{strayByte}, refused: true},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			session := clientSession(test, serving(test, slog.New(slog.DiscardHandler)), "reader")
			for _, taken := range testCase.before {
				if err := session.Setenv(taken.Name, taken.Value); err != nil {
					test.Fatalf("Setenv(%q, %q) = %v, want it taken", taken.Name, taken.Value, err)
				}
			}
			accepted, err := session.SendRequest(environmentType, true, testCase.payload)
			if err != nil || accepted == testCase.refused {
				test.Errorf("an env request of %q after %+v = %t, %v; want refused %t", testCase.payload, testCase.before,
					accepted, err, testCase.refused)
			}
		})
	}
}

// The requests a session refuses reach the log once, as the session's requests end, with how many there were and the
// first, so a client sending a flood of them writes one line.
func TestASessionsRefusedRequestsAreLoggedOnce(test *testing.T) {
	language := gossh.Marshal(&variable{Name: "LANG", Value: "C.UTF-8"})
	noColour := gossh.Marshal(&variable{Name: "NO_COLOR", Value: "1"})
	tooLong := gossh.Marshal(&variable{Name: "COLORTERM", Value: strings.Repeat(floodKey, longestColourValue+1)})
	cases := []struct {
		name     string
		payloads [][]byte
		counted  refusalsCounted
	}{
		{
			name: "one", payloads: [][]byte{language},
			counted: refusalsCounted{unreadable: 0, unread: 1, setAlready: 0, tooLong: 0},
		},
		{
			name: "several, each for its reason", payloads: [][]byte{language, language, noColour, noColour, tooLong, {strayByte}},
			counted: refusalsCounted{unreadable: 1, unread: 2, setAlready: 1, tooLong: 1},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			records := make(chan slog.Record, logLines)
			session := clientSession(test, serving(test, slog.New(recordedLog{records: records})), "reader")
			for _, payload := range testCase.payloads {
				if _, err := session.SendRequest(environmentType, true, payload); err != nil {
					test.Fatalf("an env request of %q = %v, want it answered", payload, err)
				}
			}
			if err := session.Close(); err != nil {
				test.Fatalf("Close of the session = %v, want it closed", err)
			}
			record := nextRecord(test, records)
			if counted := countedIn(record); record.Level != slog.LevelWarn || counted != testCase.counted {
				test.Errorf("the log after the requests %q = %v, counting %+v; want one warning counting %+v", testCase.payloads,
					record, counted, testCase.counted)
			}
		})
	}
}

// refusalsCounted are how many requests the log says a session refused, by why.
type refusalsCounted struct {
	unreadable int64
	unread     int64
	setAlready int64
	tooLong    int64
}

// countedIn is how many refused requests a record counts, by why, none where it counts nothing.
func countedIn(record slog.Record) refusalsCounted {
	return refusalsCounted{
		unreadable: countOf(record, unreadableKey), unread: countOf(record, unreadVariableKey),
		setAlready: countOf(record, setAlreadyKey), tooLong: countOf(record, tooLongKey),
	}
}

// countOf is the count a record holds under a key, or none.
func countOf(record slog.Record, key string) int64 {
	value, _ := valueOf(record, key)
	count, _ := value.Any().(int64)
	return count
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

// A connection that starts no session, however far it goes, is closed once its time to start one is up, so a client
// that only connects, or only shakes hands, holds a file of the server's for that time at most.
func TestAConnectionThatStartsNoSessionIsClosed(test *testing.T) {
	cases := []struct {
		name    string
		reaches func(test *testing.T, connection net.Conn)
	}{
		{name: "sending nothing", reaches: sendingNothing},
		{name: "shaking hands", reaches: shakingHands},
		{name: "opening a channel with no shell", reaches: openingAChannel},
		{name: "asking for a subsystem", reaches: askingForASubsystem},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			synctest.Test(test, func(test *testing.T) {
				server := newServer(test, slog.New(slog.DiscardHandler))
				clientEnd, serverEnd := memoryConnections()
				go server.server.HandleConn(serverEnd)
				opened := time.Now()
				testCase.reaches(test, clientEnd)
				<-serverEnd.own.closed
				if held := time.Since(opened); held > openingTimeout {
					test.Errorf("a connection %s was held %s, want it closed within %s", testCase.name, held, openingTimeout)
				}
			})
		})
	}
}

// sendingNothing is a client that connects and sends nothing.
func sendingNothing(*testing.T, net.Conn) {}

// shakingHands is a client that finishes its handshake and opens no channel.
//
//wrapper:a case of the table, in the shape the table hands every case
func shakingHands(test *testing.T, connection net.Conn) {
	clientOver(test, connection, memoryAddress, "reader")
}

// openingAChannel is a client that opens a channel for a session and asks it for nothing.
func openingAChannel(test *testing.T, connection net.Conn) {
	if _, err := clientOver(test, connection, memoryAddress, "reader").NewSession(); err != nil {
		test.Fatalf("NewSession = %v, want a channel for a session", err)
	}
}

// askingForASubsystem is a client that asks its session for a subsystem, which a kiosk refuses.
func askingForASubsystem(test *testing.T, connection net.Conn) {
	session, err := clientOver(test, connection, memoryAddress, "reader").NewSession()
	if err != nil {
		test.Fatalf("NewSession = %v, want a channel for a session", err)
	}
	if err := session.RequestSubsystem(sftpSubsystem); err == nil {
		test.Errorf("RequestSubsystem(%q) = %v, want it refused", sftpSubsystem, err)
	}
}

// A connection ends with its one session, since it may open no other: a client that keeps its connection, as
// OpenSSH's ControlPersist does, holds no file of the server's for a session it has closed.
func TestAConnectionEndsWithItsSession(test *testing.T) {
	synctest.Test(test, func(test *testing.T) {
		server := newServer(test, slog.New(slog.DiscardHandler))
		clientEnd, serverEnd := memoryConnections()
		go server.server.HandleConn(serverEnd)
		client, session := sessionOver(test, clientEnd, memoryAddress, "reader")
		output, _ := openTerminal(test, session, commonTerminal())
		readUntil(test, output, "contact")
		closed := time.Now()
		if err := session.Close(); err != nil {
			test.Fatalf("Close of the session = %v, want it closed", err)
		}
		if err := client.Wait(); err == nil {
			test.Fatalf("Wait on the connection after its session = %v, want it ended by the server", err)
		}
		if held := time.Since(closed); held > sessionEndTolerance {
			test.Errorf("a connection was held %s after its session closed, want it closed within %s", held, sessionEndTolerance)
		}
	})
}

// A connection that starts its session in time keeps it past the time to start one: its reader still reads.
func TestASessionStartedInTimeOutlivesItsTimeToStart(test *testing.T) {
	synctest.Test(test, func(test *testing.T) {
		server := newServer(test, slog.New(slog.DiscardHandler))
		clientEnd, serverEnd := memoryConnections()
		go server.server.HandleConn(serverEnd)
		client, session := sessionOver(test, clientEnd, memoryAddress, "reader")
		output, input := openTerminal(test, session, commonTerminal())
		readUntil(test, output, "contact")
		time.Sleep(openingTimeout)
		if _, err := io.WriteString(input, ":checkhealth\r"); err != nil {
			test.Fatalf("type :checkhealth past the time to start = %v, want the session open", err)
		}
		readUntil(test, output, "no tests yet")
		closeClient(client, session)
		if err := server.Shutdown(test.Context()); err != nil {
			test.Errorf("Shutdown after the session closed = %v, want the server stopped", err)
		}
	})
}

// What a test floods a session with, and the bytes that open and close a bracketed paste
const (
	floodBytes  = 8 << 10
	floodKey    = "z"
	pasteOpens  = "\x1b[200~"
	pasteCloses = "\x1b[201~"
)

// The pace no reader outruns, in bytes a second, and the most of it an ordinary paste sends at once
const (
	readersPace   = 200
	ordinaryPaste = 4 << 10
)

// A flood of a client's input, keys or a paste left open, reaches the TUI no faster than a reader types, so a client
// can spend the server's time and memory only at that pace: a command typed after the flood is answered no sooner.
func TestAFloodOfInputReachesTheTUIAtAReadersPace(test *testing.T) {
	cases := []struct {
		name  string
		flood string
	}{
		{name: "keys", flood: strings.Repeat(floodKey, floodBytes)},
		{name: "a paste", flood: pasteOpens + strings.Repeat(floodKey, floodBytes) + pasteCloses},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			synctest.Test(test, func(test *testing.T) {
				server := newServer(test, slog.New(slog.DiscardHandler))
				clientEnd, serverEnd := memoryConnections()
				go server.server.HandleConn(serverEnd)
				client, session := sessionOver(test, clientEnd, memoryAddress, "reader")
				output, input := openTerminal(test, session, commonTerminal())
				readUntil(test, output, "contact")
				flooded := time.Now()
				if _, err := io.WriteString(input, testCase.flood+":checkhealth\r"); err != nil {
					test.Fatalf("send a flood of %s = %v, want it taken", testCase.name, err)
				}
				readUntil(test, output, "no tests yet")
				paced := time.Duration(len(testCase.flood)-ordinaryPaste) * time.Second / readersPace
				if taken := time.Since(flooded); taken < paced {
					test.Errorf("a flood of %d bytes of %s took %s to reach the TUI, want %s at least", len(testCase.flood),
						testCase.name, taken, paced)
				}
				closeClient(client, session)
			})
		})
	}
}

// What of an ordinary paste a test leaves room for: the paste's marks and the command typed after it, and the
// terminal's answers as the TUI opens; and how soon a paste must reach the TUI to count as at once
const (
	pasteRoom  = 256
	pasteDelay = time.Second
)

// An ordinary paste reaches the TUI at once, never stalled by the pace a flood is held to: a command typed after it
// is answered within a second.
func TestAnOrdinaryPasteReachesTheTUIAtOnce(test *testing.T) {
	synctest.Test(test, func(test *testing.T) {
		server := newServer(test, slog.New(slog.DiscardHandler))
		clientEnd, serverEnd := memoryConnections()
		go server.server.HandleConn(serverEnd)
		client, session := sessionOver(test, clientEnd, memoryAddress, "reader")
		output, input := openTerminal(test, session, commonTerminal())
		readUntil(test, output, "contact")
		pasted := time.Now()
		paste := pasteOpens + strings.Repeat(floodKey, ordinaryPaste-pasteRoom) + pasteCloses
		if _, err := io.WriteString(input, paste+":checkhealth\r"); err != nil {
			test.Fatalf("paste %d bytes = %v, want them taken", len(paste), err)
		}
		readUntil(test, output, "no tests yet")
		if taken := time.Since(pasted); taken > pasteDelay {
			test.Errorf("a paste of %d bytes took %s to reach the TUI, want %s at most", len(paste), taken, pasteDelay)
		}
		closeClient(client, session)
	})
}

// A session its reader closes ends there, its TUI with it, though its client keeps the connection, as OpenSSH's
// ControlPersist does: the client holds no TUI it has closed. It runs on a socket: a TUI left
// running once its session is closed keeps wish reading the session's closed window changes, which no synctest
// bubble sees blocked.
func TestASessionClosedOnAnOpenConnectionEnds(test *testing.T) {
	records := make(chan slog.Record, logLines)
	address := serving(test, slog.New(recordedLog{records: records}))
	session := clientSession(test, address, "reader")
	output, _ := openTerminal(test, session, commonTerminal())
	readUntil(test, output, "contact")
	nextRecord(test, records)
	if err := session.Close(); err != nil {
		test.Fatalf("Close of the session = %v, want it closed", err)
	}
	if _, lasted := valueOf(nextRecord(test, records), lastedKey); !lasted {
		test.Errorf("the log after the session closed holds no session's close, want it closed with its connection open")
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
	cases := []struct {
		name   string
		user   string
		logged string
	}{
		{name: "a name", user: "mira", logged: "mira"},
		{name: "a name too long", user: strings.Repeat("m", longestLoggedText+1), logged: strings.Repeat("m", longestLoggedText) + textCut},
		{
			name: "a name too long cut before a letter of two bytes", user: strings.Repeat("m", longestLoggedText-1) + "üü",
			logged: strings.Repeat("m", longestLoggedText-1) + textCut,
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			records := make(chan slog.Record, logLines)
			address := serving(test, slog.New(recordedLog{records: records}))
			_, _ = clientSession(test, address, testCase.user).CombinedOutput("")
			opened, closed := nextRecord(test, records), nextRecord(test, records)
			_, openedLasted := valueOf(opened, lastedKey)
			_, closedLasted := valueOf(closed, lastedKey)
			openedUser, _ := valueOf(opened, userKey)
			closedUser, _ := valueOf(closed, userKey)
			if openedLasted || !closedLasted || openedUser.String() != testCase.logged || closedUser.String() != testCase.logged {
				test.Errorf("the log = %v then %v, want the session of %q opened, then closed with how long it lasted", opened,
					closed, testCase.logged)
			}
		})
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
	store, err := release.Open(release.FoldersAt(fixtures.ContentFolder(test), fixtures.BuildFolder(test)), compose.SharedAddresses)
	if err != nil {
		test.Fatalf("Open(fixtures) = %v, want no error", err)
	}
	settings := Settings{
		HostKeyPath: filepath.Join(test.TempDir(), "host_key"), OpeningTimeout: openingTimeout, IdleTimeout: longTimeout,
		MaxTimeout: longTimeout,
	}
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
	client := clientOver(test, connection, address, user)
	session, err := client.NewSession()
	if err != nil {
		test.Fatalf("NewSession = %v, want a session", err)
	}
	test.Cleanup(func() { closeClient(client, session) })
	return client, session
}

// clientOver is a client over a connection to an address, its handshake done, of a reader of a name who offers no
// key or password.
func clientOver(test *testing.T, connection net.Conn, address, user string) *gossh.Client {
	config := &gossh.ClientConfig{User: user, Auth: nil, HostKeyCallback: gossh.InsecureIgnoreHostKey()} //nolint:gosec // a test's own server
	clientConnection, channels, requests, err := gossh.NewClientConn(connection, address, config)
	if err != nil {
		test.Fatalf("NewClientConn = %v, want an SSH connection", err)
	}
	return gossh.NewClient(clientConnection, channels, requests)
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

// recordedLog hands each record a logger writes to a channel, so a test reads what the server logs by its values,
// not by its words. The server's loggers add no attribute or group of their own, so neither is kept.
type recordedLog struct {
	records chan<- slog.Record
}

func (recordedLog) Enabled(context.Context, slog.Level) bool { return true }

func (recorded recordedLog) Handle(_ context.Context, record slog.Record) error {
	recorded.records <- record.Clone()
	return nil
}

func (recorded recordedLog) WithAttrs([]slog.Attr) slog.Handler { return recorded }

func (recorded recordedLog) WithGroup(string) slog.Handler { return recorded }

// nextRecord is the next record the server logs, or the test's end once the server has logged nothing for too long.
func nextRecord(test *testing.T, records <-chan slog.Record) slog.Record {
	select {
	case record := <-records:
		return record
	case <-time.After(logWait):
		test.Fatalf("the server logged nothing for %s", logWait)
		return slog.Record{}
	}
}

// valueOf is the value a record holds under a key, resolved as a handler resolves it, and whether it holds one.
func valueOf(record slog.Record, key string) (slog.Value, bool) {
	search := attributeSearch{key: key, found: slog.Value{}, held: false}
	record.Attrs(search.look)
	return search.found, search.held
}

// attributeSearch is a search of a record's attributes for the one of a key, and what it found.
type attributeSearch struct {
	key   string
	found slog.Value
	held  bool
}

// look takes an attribute if it is the one sought, and goes on while it has not found it.
func (search *attributeSearch) look(attribute slog.Attr) bool {
	if attribute.Key == search.key {
		search.found, search.held = attribute.Value.Resolve(), true
	}
	return !search.held
}

// lineWriter hands each line a logger writes to a channel, so a test reads the log as the server writes it.
type lineWriter struct {
	lines chan<- string
}

func (writer lineWriter) Write(line []byte) (int, error) {
	writer.lines <- string(line)
	return len(line), nil
}
