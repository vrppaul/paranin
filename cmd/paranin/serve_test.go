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

	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/release"
)

// A reload takes good content and refuses content with faults, keeping what the site served; a stop then ends
// the site cleanly.
func TestTheSiteAnswersItsSignals(test *testing.T) {
	cases := []struct {
		name        string
		replacement string
		wantLog     string
		wantName    string
	}{
		{name: "a reload of good content", replacement: "name: Mira Other", wantLog: "the release is loaded again", wantName: "Mira Other"},
		{name: "a reload of faulty content", replacement: `name: ""`, wantLog: "the new release is refused", wantName: "Mira Example"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			folder := fixtures.Copy(test, fixtures.ContentFolder(test))
			running, log := testSite(test, folder)
			stopping, stop := context.WithCancel(context.Background())
			exits := make(chan int, 1)
			go func() { exits <- running.watch(stopping) }()
			fixtures.ReplaceIn(test, folder, "profile.yaml", "name: Mira Example", testCase.replacement)
			running.reloads <- syscall.SIGHUP
			stop()
			code := <-exits
			if name := running.store.Current().Content.Profile.Person.Name; code != succeeded || name != testCase.wantName ||
				!strings.Contains(log.String(), testCase.wantLog) {
				test.Errorf("exit %d with the name %q, logging %q; want exit %d with %q, logging %q", code, name,
					log.String(), succeeded, testCase.wantName, testCase.wantLog)
			}
		})
	}
}

// A server that fails ends the site, and the program says so in its exit.
func TestASiteWhoseServerFailsExitsFailed(test *testing.T) {
	running, _ := testSite(test, fixtures.ContentFolder(test))
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
			running, _ := testSite(test, fixtures.ContentFolder(test))
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

// testSite is a site on a folder of content and the build's fixture that listens nowhere, its reloads unbuffered so
// a test knows when one is taken, and the log it writes.
func testSite(test *testing.T, folder string) (serving, *bytes.Buffer) {
	store, err := release.Open(release.FoldersAt(folder, fixtures.BuildFolder(test)))
	if err != nil {
		test.Fatalf("Open(fixtures) = %v, want no error", err)
	}
	log := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(log, nil))
	return serving{
		logger:      logger,
		store:       store,
		servers:     []server{newServer(http.NotFoundHandler(), logger)},
		reloads:     make(chan os.Signal),
		failures:    make(chan error, 1),
		stopSignals: keepSignals,
	}, log
}

// keepSignals gives no signal back: a test sends none.
func keepSignals() {}
