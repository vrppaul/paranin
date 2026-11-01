package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/vrppaul/paranin/internal/content"
)

// The content package's fixture
const fixturePath = "../../internal/content/testdata/content"

// A reload takes good content and refuses content with faults, keeping what the site served; a stop then ends
// the site cleanly.
func TestTheSiteAnswersItsSignals(test *testing.T) {
	cases := []struct {
		name        string
		replacement string
		wantLog     string
		wantName    string
	}{
		{name: "a reload of good content", replacement: "name: Mira Other", wantLog: "the content is loaded again", wantName: "Mira Other"},
		{name: "a reload of faulty content", replacement: `name: ""`, wantLog: "the new content is refused", wantName: "Mira Example"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			folder := fixtureCopy(test)
			running, log := testSite(test, folder)
			stopping, stop := context.WithCancel(context.Background())
			exits := make(chan int, 1)
			go func() { exits <- running.watch(stopping) }()
			renameAuthor(test, folder, testCase.replacement)
			running.reloads <- syscall.SIGHUP
			stop()
			code := <-exits
			if name := running.store.Current().Profile.Person.Name; code != succeeded || name != testCase.wantName ||
				!strings.Contains(log.String(), testCase.wantLog) {
				test.Errorf("exit %d with the name %q, logging %q; want exit %d with %q, logging %q", code, name,
					log.String(), succeeded, testCase.wantName, testCase.wantLog)
			}
		})
	}
}

// A server that fails ends the site, and the program says so in its exit.
func TestASiteWhoseServerFailsExitsFailed(test *testing.T) {
	running, _ := testSite(test, fixtureCopy(test))
	running.failures <- errors.New("the port went away")
	if code := running.watch(context.Background()); code != failed {
		test.Errorf("exit %d, want %d", code, failed)
	}
}

// testSite is a site on a folder that listens nowhere, its reloads unbuffered so a test knows when one is taken,
// and the log it writes.
func testSite(test *testing.T, folder string) (serving, *bytes.Buffer) {
	store, err := content.Open(os.DirFS(folder))
	if err != nil {
		test.Fatalf("Open(fixture) = %v, want no error", err)
	}
	log := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(log, nil))
	return serving{
		logger:      logger,
		store:       store,
		server:      newServer(http.NotFoundHandler(), logger),
		reloads:     make(chan os.Signal),
		failures:    make(chan error, 1),
		stopSignals: keepSignals,
	}, log
}

// keepSignals gives no signal back: a test sends none.
func keepSignals() {}

func fixtureCopy(test *testing.T) string {
	folder := test.TempDir()
	if err := os.CopyFS(folder, os.DirFS(fixturePath)); err != nil {
		test.Fatalf("copy the fixture: %v", err)
	}
	return folder
}

// renameAuthor writes the profile of a folder with its author's name written otherwise.
func renameAuthor(test *testing.T, folder, replacement string) {
	path := filepath.Join(folder, "profile.yaml")
	profile, err := os.ReadFile(path)
	if err != nil {
		test.Fatalf("read the copy's profile: %v", err)
	}
	if err := os.WriteFile(path, []byte(strings.Replace(string(profile), "name: Mira Example", replacement, 1)), 0o600); err != nil {
		test.Fatalf("write the copy's profile: %v", err)
	}
}
