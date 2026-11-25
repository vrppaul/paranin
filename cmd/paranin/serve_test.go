package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"syscall"
	"testing"
	"testing/synctest"

	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/release"
)

// A reload, asked by SIGHUP or by a change in the folders, takes good content and refuses content with faults,
// keeping what the site served, and tells IndexNow's notifier of a reload that took alone; a stop then ends the site
// cleanly.
func TestTheSiteAnswersItsSignals(test *testing.T) {
	cases := []struct {
		name        string
		replacement string
		ask         func(tested siteUnderTest)
		wantLog     string
		wantName    string
		wantTold    bool
	}{
		{
			name: "a reload of good content", replacement: "name: Mira Other", ask: askBySignal, wantLog: "the release is loaded again",
			wantName: "Mira Other", wantTold: true,
		},
		{
			name: "a reload of faulty content", replacement: `name: ""`, ask: askBySignal, wantLog: "the new release is refused",
			wantName: "Mira Example", wantTold: false,
		},
		{
			name: "a change in the folders", replacement: "name: Mira Other", ask: askByChange, wantLog: "the release is loaded again",
			wantName: "Mira Other", wantTold: true,
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			// In a bubble, a site that never takes what the test asks leaves every goroutine waiting, which synctest
			// reports, where outside one the test would hang
			synctest.Test(test, func(test *testing.T) {
				folder := fixtures.Copy(test, fixtures.ContentFolder(test))
				tested := testSite(test, folder)
				stopping, stop := context.WithCancel(context.Background())
				exits := make(chan int, 1)
				go func() { exits <- tested.running.watch(stopping) }()
				fixtures.ReplaceIn(test, folder, "profile.yaml", "name: Mira Example", testCase.replacement)
				testCase.ask(tested)
				stop()
				code := <-exits
				name, told := tested.running.store.Current().Content.Profile.Person.Name, len(tested.running.reloaded) > 0
				if code != succeeded || name != testCase.wantName || !strings.Contains(tested.log.String(), testCase.wantLog) || told != testCase.wantTold {
					test.Errorf("exit %d with the name %q, logging %q, IndexNow's notifier told %t; want exit %d with %q, logging %q, told %t",
						code, name, tested.log.String(), told, succeeded, testCase.wantName, testCase.wantLog, testCase.wantTold)
				}
			})
		})
	}
}

// How many records a test of the site's log keeps, more than a start, a reload and a stop write
const recordsKept = 16

// seenRecord is a record of the site's log as a test reads it: its level, and how many values it names.
type seenRecord struct {
	level  slog.Level
	values int
}

// The site logs what its content warns of, the fixture's translation older than its original, as it starts and again
// at each load, among what else it logs.
func TestTheSiteLogsWhatItsContentWarnsOfAtEachLoad(test *testing.T) {
	synctest.Test(test, func(test *testing.T) {
		tested := testSite(test, fixtures.Copy(test, fixtures.ContentFolder(test)))
		records := make(chan slog.Record, recordsKept)
		tested.running.logger = slog.New(fixtures.RecordedLog{Records: records})
		stopping, stop := context.WithCancel(context.Background())
		exits := make(chan int, 1)
		go func() { exits <- tested.running.watch(stopping) }()
		askBySignal(tested)
		stop()
		<-exits
		close(records)
		seen := []seenRecord{}
		for record := range records {
			seen = append(seen, seenRecord{level: record.Level, values: record.NumAttrs()})
		}
		// the warnings as it starts, the reload and its warnings, and the stop
		want := []seenRecord{{level: slog.LevelWarn, values: 1}, {level: slog.LevelInfo, values: 0}, {level: slog.LevelWarn, values: 1}, {level: slog.LevelInfo, values: 0}}
		if !slices.Equal(seen, want) {
			test.Errorf("the site's log = %+v, want %+v", seen, want)
		}
	})
}

// askBySignal asks a site to load its release again by SIGHUP.
func askBySignal(tested siteUnderTest) {
	tested.running.reloads <- syscall.SIGHUP
}

// askByChange tells a site its folders changed and held still, as its watch does.
func askByChange(tested siteUnderTest) {
	tested.changes <- struct{}{}
}

// A server that fails ends the site, and the program says so in its exit.
func TestASiteWhoseServerFailsExitsFailed(test *testing.T) {
	running := testSite(test, fixtures.ContentFolder(test)).running
	running.failures <- errors.New("the port went away")
	if code := running.watch(context.Background()); code != failed {
		test.Errorf("exit %d, want %d", code, failed)
	}
}

// A stop shuts every server down, and a server that does not stop cleanly makes the exit say so.
func TestAStopShutsEveryServerDown(test *testing.T) {
	cases := []struct {
		name  string
		fault error
		want  int
	}{
		{name: "every server stops cleanly", fault: nil, want: succeeded},
		{name: "a server does not stop cleanly", fault: errors.New("a session would not end"), want: failed},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			running := testSite(test, fixtures.ContentFolder(test)).running
			first, second := &stoppingServer{fault: nil, stopped: false}, &stoppingServer{fault: testCase.fault, stopped: false}
			running.servers = []server{first, second}
			stopping, stop := context.WithCancel(context.Background())
			stop()
			if code := running.watch(stopping); code != testCase.want || !first.stopped || !second.stopped {
				test.Errorf("exit %d, the servers stopped %t and %t; want exit %d and both stopped", code, first.stopped,
					second.stopped, testCase.want)
			}
		})
	}
}

// How long a test's Accept header is: the longest header line nginx takes, and one over the bound and its slack
const (
	longestLineAcceptLength = 8 << 10
	oversizedAcceptLength   = 32 << 10
)

// A request whose head is over the bound is refused before any handler reads it; one nginx would pass is answered.
func TestTheServerRefusesAHeadOverItsBound(test *testing.T) {
	cases := []struct {
		name         string
		acceptLength int
		want         int
	}{
		{name: "the longest header line nginx takes", acceptLength: longestLineAcceptLength, want: http.StatusNoContent},
		{name: "a head over the bound", acceptLength: oversizedAcceptLength, want: http.StatusRequestHeaderFieldsTooLarge},
	}
	answering := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { writer.WriteHeader(http.StatusNoContent) })
	// The bound is held where net/http reads the connection, so the test reads it over a socket
	listening := httptest.NewUnstartedServer(nil)
	listening.Config = newServer(answering, slog.New(slog.DiscardHandler))
	listening.Start()
	defer listening.Close()
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			request, err := http.NewRequestWithContext(test.Context(), http.MethodGet, listening.URL, nil)
			if err != nil {
				test.Fatalf("NewRequestWithContext(%q) = %v, want no error", listening.URL, err)
			}
			request.Header.Set("Accept", strings.Repeat("a", testCase.acceptLength))
			response, err := listening.Client().Do(request)
			if err != nil {
				test.Fatalf("GET with an Accept of %d bytes = %v, want an answer", testCase.acceptLength, err)
			}
			defer response.Body.Close()
			if response.StatusCode != testCase.want {
				test.Errorf("GET with an Accept of %d bytes = %d, want %d", testCase.acceptLength, response.StatusCode, testCase.want)
			}
		})
	}
}

// stoppingServer is a server that records its shutdown and answers it with a fault, or none.
type stoppingServer struct {
	fault   error
	stopped bool
}

func (stopping *stoppingServer) Shutdown(context.Context) error {
	stopping.stopped = true
	return stopping.fault
}

// siteUnderTest is the serving a test runs, beside the log it writes and the way a test tells it its folders
// changed.
type siteUnderTest struct {
	running serving
	log     *bytes.Buffer
	changes chan<- struct{}
}

// testSite is a site on a folder of content and the build's and the cards' fixtures that listens nowhere, its reloads
// and changes unbuffered so a test knows when one is taken.
func testSite(test *testing.T, folder string) siteUnderTest {
	store, err := release.Open(fixtureRelease(test, folder), releaseChecks())
	if err != nil {
		test.Fatalf("Open(fixtures) = %v, want no error", err)
	}
	log := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(log, nil))
	changes := make(chan struct{})
	running := serving{
		logger:      logger,
		store:       store,
		servers:     []server{newServer(http.NotFoundHandler(), logger)},
		reloads:     make(chan os.Signal),
		reloaded:    make(chan struct{}, 1),
		changes:     changes,
		failures:    make(chan error, 1),
		stopSignals: keepSignals,
	}
	return siteUnderTest{running: running, log: log, changes: changes}
}

// keepSignals gives no signal back: a test sends none.
func keepSignals() {}
