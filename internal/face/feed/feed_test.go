package feed

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/golden"
	"github.com/vrppaul/paranin/internal/routes"
)

func TestMain(tests *testing.M) {
	golden.Main(tests)
}

// Atom's namespace, which a feed's every element is in; the relations a feed's links name, its own address and the
// page a person reads; what a test's body writes round a post's slug; and how many of an element RFC 4287 asks for
const (
	atomSpace           = "http://www.w3.org/2005/Atom"
	toItself            = "self"
	toItsPage           = "alternate"
	bodyOpening         = "<p>The body of "
	bodyClosing         = ", with <b>markup</b> & all.</p>"
	expectedSingleValue = 1
)

// What a case moves: the site's address, and the fixture post's folder, with the series naming it
const (
	profileFile       = "profile.yaml"
	fixtureBase       = "base_url: https://example.test"
	movedBase         = "base_url: https://moved.example.test"
	postsFolder       = "posts"
	fixturePostSlug   = "reading-slowly"
	movedSlug         = "read-slowly"
	fixtureSeriesFile = "series/reading.yaml"
)

// bodyStub writes a post's body as a test reads it: its slug among markup, which the feed must carry as text.
type bodyStub struct{}

func (bodyStub) Body(post document.Document) string {
	return bodyOpening + post.Head.Page.Slug + bodyClosing
}

// readFeed is a feed as RFC 4287 reads it, each element a list so a test counts how many there are.
type readFeed struct {
	XMLName  xml.Name     `xml:"feed"`
	IDs      []string     `xml:"id"`
	Titles   []string     `xml:"title"`
	Updates  []string     `xml:"updated"`
	Authors  []readPerson `xml:"author"`
	Links    []readLink   `xml:"link"`
	Entries  []readEntry  `xml:"entry"`
	Language string       `xml:"http://www.w3.org/XML/1998/namespace lang,attr"`
}

// readPerson is a feed's author as RFC 4287 reads one: a name.
type readPerson struct {
	Names []string `xml:"name"`
}

// readLink is a link of a feed or an entry: its relation and its address.
type readLink struct {
	Relation string `xml:"rel,attr"`
	Address  string `xml:"href,attr"`
}

// readEntry is an entry as RFC 4287 reads it.
type readEntry struct {
	IDs     []string   `xml:"id"`
	Titles  []string   `xml:"title"`
	Updates []string   `xml:"updated"`
	Links   []readLink `xml:"link"`
}

// The text of each feed is what a feed reader polls, so the text itself is the behaviour: each feed of the fixture is
// held to its golden file, named by the feed's own path.
func TestEveryFeedReadsAsItsGoldenAtom(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	for _, feed := range feedsOf(snapshot) {
		path := routes.FeedPath(feed.Target)
		test.Run(path, func(test *testing.T) {
			golden.Match(test, path, Atom(feed, fixtures.Site(snapshot), bodyStub{}))
		})
	}
}

// Every feed keeps to what RFC 4287 requires of one, as a validator would hold it: its root a feed in Atom's
// namespace, in a language; one id, one title and one moment it was updated, an author named, and a link to itself;
// and in each entry one id, one title and one moment, and a link to the page a person reads it on.
func TestEveryFeedKeepsToAtom(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	for _, feed := range feedsOf(snapshot) {
		test.Run(routes.FeedPath(feed.Target), func(test *testing.T) {
			read := readAtom(test, Atom(feed, fixtures.Site(snapshot), bodyStub{}))
			if read.XMLName.Space != atomSpace || read.Language == "" || len(read.IDs) != expectedSingleValue ||
				len(read.Titles) != expectedSingleValue || !oneMoment(read.Updates) || len(read.Authors) == 0 || len(read.Authors[0].Names) != expectedSingleValue ||
				!linksTo(read.Links, toItself) {
				test.Errorf("the feed reads as %+v, want a feed in Atom's namespace, in a language, of one id, title and moment, a named author and a link to itself", read)
			}
			for _, entry := range read.Entries {
				if len(entry.IDs) != expectedSingleValue || len(entry.Titles) != expectedSingleValue || !oneMoment(entry.Updates) || !linksTo(entry.Links, toItsPage) {
					test.Errorf("an entry reads as %+v, want one id, title and moment, and a link to its page", entry)
				}
			}
		})
	}
}

// A feed and its posts keep their names wherever they move, the site's address and a post's folder, so a reader never
// shows a post as new: the blog's feed of the site moved names itself and its posts as the fixture's does.
func TestAFeedAndItsPostsKeepTheirNamesWhereverTheyMove(test *testing.T) {
	folder := fixtures.Copy(test, fixtures.ContentFolder(test))
	fixtures.ReplaceIn(test, folder, profileFile, fixtureBase, movedBase)
	if err := os.Rename(filepath.Join(folder, postsFolder, fixturePostSlug), filepath.Join(folder, postsFolder, movedSlug)); err != nil {
		test.Fatalf("rename the fixture post's folder: %v", err)
	}
	fixtures.ReplaceIn(test, folder, fixtureSeriesFile, fixturePostSlug, movedSlug)
	want, got := namesOf(test, fixtures.LoadedContent(test)), namesOf(test, fixtures.LoadedFrom(test, folder))
	if !reflect.DeepEqual(got, want) {
		test.Errorf("moved, the blog's feed and its posts are named %q, want %q", got, want)
	}
}

// feedsOf are the feeds of a snapshot's site as the public reads it on the tests' day: the blog's, and its series'.
func feedsOf(snapshot *content.Snapshot) []compose.Feed {
	return compose.SiteOf(compose.Prepare(snapshot), compose.Audience{At: fixtures.Day(), Drafts: false}).Own.Feeds
}

// namesOf are the ids of the blog's feed of a snapshot's site, its own and then its posts'.
func namesOf(test *testing.T, snapshot *content.Snapshot) []string {
	read := readAtom(test, Atom(feedsOf(snapshot)[0], fixtures.Site(snapshot), bodyStub{}))
	names := read.IDs
	for _, entry := range read.Entries {
		names = append(names, entry.IDs...)
	}
	return names
}

// readAtom is a feed's text as RFC 4287 reads it.
func readAtom(test *testing.T, written string) readFeed {
	test.Helper()
	var read readFeed
	if err := xml.Unmarshal([]byte(written), &read); err != nil {
		test.Fatalf("the feed reads as no XML: %v", err)
	}
	return read
}

// oneMoment is whether a list holds one moment, as RFC 3339 writes it.
func oneMoment(moments []string) bool {
	if len(moments) != expectedSingleValue {
		return false
	}
	_, err := time.Parse(time.RFC3339, moments[0])
	return err == nil
}

// linksTo is whether links hold one of a relation, to an address.
func linksTo(links []readLink, relation string) bool {
	for _, link := range links {
		if link.Relation == relation && link.Address != "" {
			return true
		}
	}
	return false
}
