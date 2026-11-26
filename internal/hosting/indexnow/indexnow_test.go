package indexnow

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/vrppaul/paranin/internal/assets"
	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/release"
)

// A key is IndexNow's only where it has 8 to 128 characters, each a letter or a digit of ASCII or a hyphen.
func TestAKeyIsHeldToIndexNowsForm(test *testing.T) {
	cases := []struct {
		name      string
		key       string
		isRefused bool
	}{
		{name: "letters, digits and hyphens", key: "a1B2-c3D4", isRefused: false},
		{name: "the shortest", key: strings.Repeat("a", 8), isRefused: false},
		{name: "the longest", key: strings.Repeat("a", 128), isRefused: false},
		{name: "too short", key: strings.Repeat("a", 7), isRefused: true},
		{name: "too long", key: strings.Repeat("a", 129), isRefused: true},
		{name: "an underscore", key: "abcd_efgh", isRefused: true},
		{name: "a dot", key: "abcd.efgh", isRefused: true},
		{name: "a slash", key: "abcd/efgh", isRefused: true},
		{name: "a letter outside ASCII", key: "abcdéfgh", isRefused: true},
		{name: "none", key: "", isRefused: true},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if err := CheckKey(testCase.key); (err != nil) != testCase.isRefused {
				test.Errorf("CheckKey(%q) = %v, want it refused: %t", testCase.key, err, testCase.isRefused)
			}
		})
	}
}

// The key the tests' site is known by, and where IndexNow is told its file is, at the fixture site's root; the fixture
// site's host, and the addresses of its blog's list by year and by series
const (
	testKey           = "fixture-key-3f9a"
	testKeyFile       = "https://example.test/fixture-key-3f9a.txt"
	fixtureHost       = "example.test"
	yearListAddress   = "https://example.test/blog"
	seriesListAddress = "https://example.test/blog/series"
)

// The fixture's profile, which names the address the site is served at
const profileFile = "profile.yaml"

// How the fixture's profile names the address the site is served at, and how it names another host the site moves to;
// that host, and where IndexNow is told the key's file is there
const (
	fixtureBaseAddress = "base_url: https://example.test"
	movedBaseAddress   = "base_url: https://moved.test"
	movedHost          = "moved.test"
	movedKeyFile       = "https://moved.test/fixture-key-3f9a.txt"
)

// How the post added is unlisted: its status, as written, and as changed
const (
	publishedStatus = "status: published"
	unlistedStatus  = "status: unlisted"
)

// Posts a test adds to the fixture content, each by its file and its text: one out as the bubble's clock starts, at
// midnight on the first of January, 2000; another out then too; and one dated an hour later
const (
	addedPostFile    = "posts/added/post.en.md"
	addedPostText    = "---\nid: 3\ntitle: Added\ndescription: About what was added.\nstatus: published\npublished: 2000-01-01T00:00+00:00\n---\nIts text.\n"
	addedPostAddress = "https://example.test/blog/added"
	otherPostFile    = "posts/other/post.en.md"
	otherPostText    = "---\nid: 4\ntitle: Other\ndescription: About another thing.\nstatus: published\npublished: 2000-01-01T00:00+00:00\n---\nIts text.\n"
	otherPostAddress = "https://example.test/blog/other"
	comingPostFile   = "posts/coming/post.en.md"
	comingPostText   = "---\nid: 3\ntitle: Coming\ndescription: About what comes.\nstatus: published\npublished: 2000-01-01T01:00+00:00\n---\nIts text.\n"
	comingAddress    = "https://example.test/blog/coming"
	comingIn         = time.Hour
)

// How many submissions and log records a test keeps, more than any test makes
const keptPerTest = 16

// endpoint stands for IndexNow's: it hands the test each submission it takes, and answers with the status the test
// sets. It runs outside the test's bubble, since a socket listened at in one would hold the bubble's clock still.
type endpoint struct {
	server      *httptest.Server
	submissions chan submission
	status      *atomic.Int64
}

// endpointOf is an endpoint that answers 200 until the test says otherwise, closed when the test ends.
func endpointOf(test *testing.T) endpoint {
	status := &atomic.Int64{}
	status.Store(http.StatusOK)
	submissions := make(chan submission, keptPerTest)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		takeSubmission(writer, request, submissions, int(status.Load()))
	}))
	test.Cleanup(server.Close)
	return endpoint{server: server, submissions: submissions, status: status}
}

// takeSubmission reads a submission off a request, hands it on and answers with a status.
func takeSubmission(writer http.ResponseWriter, request *http.Request, submissions chan<- submission, status int) {
	var taken submission
	// a body that does not read is handed on empty, which no test wants
	_ = json.NewDecoder(request.Body).Decode(&taken)
	submissions <- taken
	writer.WriteHeader(status)
}

// takenSubmissions are the submissions an endpoint has taken and not yet handed to the test, in the order it took them.
func (taken endpoint) takenSubmissions() []submission {
	var submissions []submission
	for range len(taken.submissions) {
		submissions = append(submissions, <-taken.submissions)
	}
	return submissions
}

// submissionOf is the submission the tests' site makes of the addresses that changed.
func submissionOf(addresses ...string) submission {
	return submission{Host: fixtureHost, Key: testKey, KeyLocation: testKeyFile, URLList: addresses}
}

// movedSubmissionOf is the submission the tests' site makes of the addresses that changed, once moved to another host.
func movedSubmissionOf(addresses ...string) submission {
	return submission{Host: movedHost, Key: testKey, KeyLocation: movedKeyFile, URLList: addresses}
}

// notifiedSite is a site under test: the folder of its content, its store, how it is told of a reload and what its
// notifier logs.
type notifiedSite struct {
	folder  string
	store   *release.Store
	reloads chan struct{}
	records chan slog.Record
}

// addPost writes a post into the site's content and tells the notifier the reload that takes it took.
func (site notifiedSite) addPost(test *testing.T, file, text string) {
	fixtures.WriteIn(test, site.folder, file, text)
	site.reload(test)
}

// reload loads the site's content again and tells the notifier it took.
func (site notifiedSite) reload(test *testing.T) {
	if err := site.store.Reload(); err != nil {
		test.Fatalf("Reload = %v, want no error", err)
	}
	site.reloads <- struct{}{}
}

// notify starts a notifier of the fixture content, changed first as start changes it, telling an endpoint, until the
// test ends; then waits for it to stop.
func notify(test *testing.T, toldAt endpoint, start func(test *testing.T, folder string)) notifiedSite {
	folder := fixtures.Copy(test, fixtures.ContentFolder(test))
	start(test, folder)
	folders := release.Folders{Content: folder, Build: fixtures.BuildFolder(test), Cards: fixtures.CardsFolder(test)}
	checks := release.Checks{
		Content: func(loaded *content.Snapshot) error { return compose.Check(loaded, time.Now()) },
		// a post a test adds has no card in the cards' fixture, which no test here reads
		Cards: func(*content.Snapshot, *assets.Cards) error { return nil },
	}
	store, err := release.Open(folders, checks)
	if err != nil {
		test.Fatalf("Open(fixtures) = %v, want no error", err)
	}
	site := notifiedSite{folder: folder, store: store, reloads: make(chan struct{}, 1), records: make(chan slog.Record, keptPerTest)}
	notifier := New(store, release.DeriveFromContent(store, compose.Prepare), slog.New(fixtures.RecordedLog{Records: site.records}),
		Settings{Key: testKey, Endpoint: toldAt.server.URL, Client: &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}})
	running, stop := context.WithCancel(context.Background())
	stopped := make(chan struct{})
	go runNotifier(running, notifier, site.reloads, stopped)
	test.Cleanup(func() { stopNotifier(stop, stopped) })
	return site
}

// runNotifier runs a notifier until its context ends, then says it stopped.
func runNotifier(ctx context.Context, notifier *Notifier, reloads <-chan struct{}, stopped chan<- struct{}) {
	notifier.Run(ctx, reloads)
	close(stopped)
}

// How long the clock goroutine of chroma's regexps (regexp2's fastclock), which a snapshot prepared in the bubble
// starts in it, runs past the last match: the match's timeout of 250 ms, the clock's period of 100 ms and its second of
// slop. A bubble ends only once every goroutine in it has.
const lexersClockLasts = 2 * time.Second

// stopNotifier ends a notifier's run, waits for it to return, and waits out the lexers' clock.
func stopNotifier(stop context.CancelFunc, stopped <-chan struct{}) {
	stop()
	<-stopped
	time.Sleep(lexersClockLasts)
}

// writeTheComingPost writes the post dated later into a content folder.
//
//wrapper:a case of the table names the post it starts with
func writeTheComingPost(test *testing.T, folder string) {
	fixtures.WriteIn(test, folder, comingPostFile, comingPostText)
}

// The notifier tells IndexNow each address a crawler should read again, in one submission, as the site changes: a post
// out by a reload, and a post dated later at its moment, with no request to the site; never as it starts, nor where a
// reload changed nothing.
func TestTheNotifierTellsIndexNowWhatChanged(test *testing.T) {
	cases := []struct {
		name   string
		start  func(test *testing.T, folder string)
		change func(site notifiedSite, test *testing.T)
		want   []submission
	}{
		{name: "as it starts", start: fixtures.KeepFolder, change: keepSite, want: nil},
		{name: "a reload that changed nothing", start: fixtures.KeepFolder, change: notifiedSite.reload, want: nil},
		{
			name: "a reload adding a post", start: fixtures.KeepFolder, change: addThePost,
			want: []submission{submissionOf(yearListAddress, addedPostAddress, seriesListAddress)},
		},
		{
			name: "a post dated later at its moment", start: writeTheComingPost, change: waitForThePost,
			want: []submission{submissionOf(yearListAddress, comingAddress, seriesListAddress)},
		},
		{
			name: "a reload removing a post", start: writeTheAddedPost, change: removeThePost,
			want: []submission{submissionOf(yearListAddress, addedPostAddress, seriesListAddress)},
		},
		{
			name: "a reload unlisting a post", start: writeTheAddedPost, change: unlistThePost,
			want: []submission{submissionOf(yearListAddress, addedPostAddress, seriesListAddress)},
		},
		// the addresses on the host the site left are not its to tell of, and IndexNow would refuse them all
		{
			name: "a reload moving the site to another host", start: fixtures.KeepFolder, change: moveTheSite,
			want: []submission{movedSubmissionOf(
				"https://moved.test/", "https://moved.test/blog", "https://moved.test/blog/series", "https://moved.test/experience",
				"https://moved.test/p/tidy-tool", "https://moved.test/projects", "https://moved.test/ru/blog",
				"https://moved.test/ru/blog/series",
			)},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			toldAt := endpointOf(test)
			synctest.Test(test, func(test *testing.T) {
				site := notify(test, toldAt, testCase.start)
				synctest.Wait()
				testCase.change(site, test)
				synctest.Wait()
				if got := toldAt.takenSubmissions(); !reflect.DeepEqual(got, testCase.want) {
					test.Errorf("IndexNow was told %+v, want %+v", got, testCase.want)
				}
			})
		})
	}
}

// keepSite changes nothing of a site.
func keepSite(notifiedSite, *testing.T) {}

// addThePost adds a post out at once to a site.
//
//wrapper:a case of the table names the post it adds
func addThePost(site notifiedSite, test *testing.T) {
	site.addPost(test, addedPostFile, addedPostText)
}

// writeTheAddedPost writes the post out at once into a content folder.
//
//wrapper:a case of the table names the post it starts with
func writeTheAddedPost(test *testing.T, folder string) {
	fixtures.WriteIn(test, folder, addedPostFile, addedPostText)
}

// removeThePost takes the post added away from a site, its folder with it.
func removeThePost(site notifiedSite, test *testing.T) {
	fixtures.RemoveIn(test, site.folder, addedPostFile)
	fixtures.RemoveIn(test, site.folder, path.Dir(addedPostFile))
	site.reload(test)
}

// unlistThePost unlists the post added to a site.
func unlistThePost(site notifiedSite, test *testing.T) {
	fixtures.ReplaceIn(test, site.folder, addedPostFile, publishedStatus, unlistedStatus)
	site.reload(test)
}

// moveTheSite moves a site to another host.
func moveTheSite(site notifiedSite, test *testing.T) {
	fixtures.ReplaceIn(test, site.folder, profileFile, fixtureBaseAddress, movedBaseAddress)
	site.reload(test)
}

// waitForThePost waits for the post dated later to come out.
//
//wrapper:a case of the table names the moment it waits for
func waitForThePost(notifiedSite, *testing.T) {
	time.Sleep(comingIn)
}

// A submission IndexNow refuses is logged and never sent again, and the next change is sent all the same: a post out
// after it alone, since the lists' day did not change.
func TestARefusedSubmissionIsLoggedAndTheNextChangeSent(test *testing.T) {
	toldAt := endpointOf(test)
	synctest.Test(test, func(test *testing.T) {
		site := notify(test, toldAt, fixtures.KeepFolder)
		toldAt.status.Store(http.StatusServiceUnavailable)
		site.addPost(test, addedPostFile, addedPostText)
		synctest.Wait()
		toldAt.status.Store(http.StatusOK)
		site.addPost(test, otherPostFile, otherPostText)
		synctest.Wait()
		want := []submission{submissionOf(yearListAddress, addedPostAddress, seriesListAddress), submissionOf(otherPostAddress)}
		got := toldAt.takenSubmissions()
		levels := []slog.Level{}
		for range len(site.records) {
			levels = append(levels, (<-site.records).Level)
		}
		if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(levels, []slog.Level{slog.LevelError}) {
			test.Errorf("IndexNow was told %+v, logging %v; want %+v, logging %v", got, levels, want, []slog.Level{slog.LevelError})
		}
	})
}
