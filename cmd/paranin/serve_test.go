package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"syscall"
	"testing"
	"testing/synctest"

	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/release"
)

// A reload, asked by SIGHUP or by a change in the folders, takes good content and refuses content with faults,
// keeping what the site served; a stop then ends the site cleanly.
func TestTheSiteAnswersItsSignals(test *testing.T) {
	cases := []struct {
		name        string
		replacement string
		ask         func(tested siteUnderTest)
		wantLog     string
		wantName    string
	}{
		{name: "a reload of good content", replacement: "name: Mira Other", ask: askBySignal, wantLog: "the release is loaded again", wantName: "Mira Other"},
		{name: "a reload of faulty content", replacement: `name: ""`, ask: askBySignal, wantLog: "the new release is refused", wantName: "Mira Example"},
		{name: "a change in the folders", replacement: "name: Mira Other", ask: askByChange, wantLog: "the release is loaded again", wantName: "Mira Other"},
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
				if name := tested.running.store.Current().Content.Profile.Person.Name; code != succeeded || name != testCase.wantName ||
					!strings.Contains(tested.log.String(), testCase.wantLog) {
					test.Errorf("exit %d with the name %q, logging %q; want exit %d with %q, logging %q", code, name,
						tested.log.String(), succeeded, testCase.wantName, testCase.wantLog)
				}
			})
		})
	}
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

// testSite is a site on a folder of content and the build's fixture that listens nowhere, its reloads and changes
// unbuffered so a test knows when one is taken.
func testSite(test *testing.T, folder string) siteUnderTest {
	store, err := release.Open(release.FoldersAt(folder, fixtures.BuildFolder(test)))
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
		changes:     changes,
		failures:    make(chan error, 1),
		stopSignals: keepSignals,
	}
	return siteUnderTest{running: running, log: log, changes: changes}
}

// keepSignals gives no signal back: a test sends none.
func keepSignals() {}
