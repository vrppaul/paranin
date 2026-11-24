package compose

import (
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// The fixture post's versions: its current one, dated by its update, and its one revision, dated by its file's name
const (
	fixtureCurrentDay  = "2026-11-12"
	fixtureRevisionDay = "2026-11-10"
)

// dayOf is a day as the content writes it, at its first moment in UTC.
func dayOf(test *testing.T, written string) time.Time {
	test.Helper()
	day, err := time.Parse(time.DateOnly, written)
	if err != nil {
		test.Fatalf("read the day %q: %v", written, err)
	}
	return day
}

// fixtureVersions are the fixture post's versions, newest first, the version read at a page of them.
func fixtureVersions(test *testing.T, read document.Page) []document.PostVersion {
	current, older := postPageOf(fixturePostSlug), document.OlderVersion(fixturePostSlug, fixtureRevisionDay)
	return []document.PostVersion{
		{Page: current, Day: dayOf(test, fixtureCurrentDay), Name: fixtureCurrentDay, Current: true, Read: read == current},
		{Page: older, Day: dayOf(test, fixtureRevisionDay), Name: fixtureRevisionDay, Current: false, Read: read == older},
	}
}

// A post's head lists every version newest first: its current one, then each older one, the newest revision first,
// whatever order its files stand in.
func TestAPostsHeadListsItsVersionsNewestFirst(test *testing.T) {
	site := SiteOf(Prepare(changedFixture(test, secondRevisionWritten)), publicOnTheDay())
	names := []string{}
	for _, version := range document.VersionsOf(postNamed(test, site.Posts, fixturePostSlug).Head.Version) {
		names = append(names, version.Name)
	}
	if want := []string{fixtureCurrentDay, "2026-11-11", fixtureRevisionDay}; !slices.Equal(names, want) {
		test.Errorf("the versions = %q, want %q", names, want)
	}
}

// A post of one version alone lists that version, its current one, read.
func TestAPostOfOneVersionListsThatVersionAlone(test *testing.T) {
	site := SiteOf(Prepare(changedFixture(test, fixtureRevisionRemoved)), publicOnTheDay())
	versions := document.VersionsOf(postNamed(test, site.Posts, fixturePostSlug).Head.Version)
	want := []document.PostVersion{{Page: postPageOf(fixturePostSlug), Day: dayOf(test, fixtureCurrentDay), Name: fixtureCurrentDay, Current: true, Read: true}}
	if !reflect.DeepEqual(versions, want) || len(site.Versions) > 0 {
		test.Errorf("the versions = %+v and %d older ones served, want %+v alone", versions, len(site.Versions), want)
	}
}

// An older version's code is split into its kinds of word, as its post's is.
func TestAnOlderVersionsCodeIsSplitIntoItsKindsOfWord(test *testing.T) {
	site := SiteOf(Prepare(changedFixture(test, revisionOfCode)), publicOnTheDay())
	code := codeOf(test, fixtures.DocumentOf(test, site.Versions, document.OlderVersion(fixturePostSlug, "2026-11-11")))
	if kinds := kindsOfRuns(code.Runs); !kinds[document.CodeKeyword] {
		test.Errorf("the older version's code runs of kinds %v, want a keyword among them", kinds)
	}
}

// secondRevisionWritten writes a second revision of the fixture's post, between its first and its current version.
//
//wrapper:a test names the change it writes
func secondRevisionWritten(test *testing.T, folder string) {
	fixtures.WriteIn(test, folder, sampleRevisionFile, sampleRevisionHead+"A second text.\n")
}

// fixtureRevisionRemoved takes the fixture post's one revision away.
//
//wrapper:a test names the change it writes
func fixtureRevisionRemoved(test *testing.T, folder string) {
	fixtures.RemoveIn(test, folder, fixtureRevisionFile)
}

// revisionOfCode writes a revision of the fixture's post whose body is a block of Go.
//
//wrapper:a test names the change it writes
func revisionOfCode(test *testing.T, folder string) {
	fixtures.WriteIn(test, folder, sampleRevisionFile, sampleRevisionHead+"```go\nfunc slowly() {}\n```\n")
}

// An older version of a post is a page of its own at its revision's address, its head the post's but for the title,
// the lead and the description that version had, and its version, which says a newer stands, the current one, in the
// site's words round its day; its body its own.
func TestAnOlderVersionIsAPageOfItsOwnSayingANewerStands(test *testing.T) {
	site := SiteOf(Prepare(fixtures.LoadedContent(test)), publicOnTheDay())
	page := document.OlderVersion(fixturePostSlug, fixtureRevisionDay)
	older := fixtures.DocumentOf(test, site.Versions, page)
	head := postNamed(test, site.Posts, fixturePostSlug).Head
	lead := textSpans("Why a page read twice is read once, first said.")
	versions := fixtureVersions(test, page)
	head.Page, head.Title, head.Lead, head.Description = page, textSpans("Slow reading"), lead, lead
	head.Version = document.VersionOlder{
		Day: dayOf(test, fixtureRevisionDay), Newer: versions[0], NewerWords: document.Around{Before: "newer: ", After: ""}, Versions: versions,
	}
	if !reflect.DeepEqual(older.Head, head) {
		test.Errorf("the older version's head = %+v, want %+v", older.Head, head)
	}
	if first := older.Blocks[0]; !reflect.DeepEqual(first, paragraphOf(document.Text{Text: "A page read twice is read once."})) {
		test.Errorf("the older version's first block = %+v, want its own first paragraph", first)
	}
}

// Only a post out has older versions a reader reads: a post that comes later than its reader, a preview until then,
// lists its current version alone, and a draft as its author reads it none; neither serves an older one, though their
// revisions stay for when they are out.
func TestAPostNotOutServesNoOlderVersion(test *testing.T) {
	current := document.PostVersion{Page: postPageOf(fixturePostSlug), Day: dayOf(test, fixtureCurrentDay), Name: fixtureCurrentDay, Current: true, Read: true}
	cases := []struct {
		name     string
		audience Audience
		change   func(test *testing.T, folder string)
		want     []document.PostVersion
	}{
		{name: "a post that comes later", audience: Audience{At: fixtures.MomentOf(test, "2026-11-09T09:00+01:00"), Drafts: false}, change: fixtures.KeepFolder, want: []document.PostVersion{current}},
		{name: "a draft", audience: Audience{At: fixtures.Day(), Drafts: true}, change: fixturePostDrafted, want: nil},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			site := SiteOf(Prepare(changedFixture(test, testCase.change)), testCase.audience)
			read := postNamed(test, postsAndPreviews(site), fixturePostSlug).Head.Version
			if versions := document.VersionsOf(read); len(site.Versions) > 0 || !reflect.DeepEqual(versions, testCase.want) {
				test.Errorf("%s serves %d older versions and lists %+v, want none and %+v", testCase.name, len(site.Versions), versions, testCase.want)
			}
		})
	}
}

// A post's places end with its other versions, each by the day it came out: an older version's with the current one.
func TestAPostsPlacesEndWithItsOtherVersions(test *testing.T) {
	site := SiteOf(Prepare(fixtures.LoadedContent(test)), publicOnTheDay())
	older := document.OlderVersion(fixturePostSlug, fixtureRevisionDay)
	cases := []struct {
		name  string
		shown document.Document
		want  document.Place
	}{
		{name: "the current version", shown: postNamed(test, site.Posts, fixturePostSlug), want: document.Place{Name: fixtureRevisionDay, Kind: document.PlaceVersion, Page: older, Section: ""}},
		{name: "an older version", shown: fixtures.DocumentOf(test, site.Versions, older), want: document.Place{Name: fixtureCurrentDay, Kind: document.PlaceVersion, Page: postPageOf(fixturePostSlug), Section: ""}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			places := document.PlacesOf(testCase.shown)
			if last := places[len(places)-1]; last != testCase.want {
				test.Errorf("the last place of %s = %+v, want %+v", testCase.name, last, testCase.want)
			}
		})
	}
}

// A heading named as a version of its post takes its count, so its name leads to one place.
func TestAHeadingNamedAsAVersionTakesItsCount(test *testing.T) {
	snapshot := changedFixture(test, headingNamedAsAVersion)
	post := postNamed(test, SiteOf(Prepare(snapshot), publicOnTheDay()).Posts, fixturePostSlug)
	want := document.Place{Name: fixtureRevisionDay + "-2", Kind: document.PlaceHeading, Page: postPageOf(fixturePostSlug), Section: fixtureRevisionDay + "-2"}
	if places := document.PlacesOf(post); !slices.Contains(places, want) {
		test.Errorf("the post's places = %+v, want %+v among them", places, want)
	}
}

// headingNamedAsAVersion titles a heading of the fixture's post as the day of its older version.
//
//wrapper:a test names the change it writes
func headingNamedAsAVersion(test *testing.T, folder string) {
	fixtures.ReplaceIn(test, folder, fixturePostFile, "## Afterwards", "## "+fixtureRevisionDay)
}

// :e opens an older version by the name a terminal reaches it by, though Tab never offers it.
func TestEditOpensAnOlderVersionByItsAddress(test *testing.T) {
	want := OpenedPage{
		Place: document.Place{
			Name: "blog/reading-slowly/" + fixtureRevisionDay, Kind: document.PlaceVersion, Page: document.OlderVersion(fixturePostSlug, fixtureRevisionDay), Section: "",
		},
		Offered: false,
	}
	if pages := SiteOf(Prepare(fixtures.LoadedContent(test)), publicOnTheDay()).Frame.Pages; !slices.Contains(pages, want) {
		test.Errorf("the pages :e opens = %+v, want %+v among them", pages, want)
	}
}

// changedFixture is the fixture content with a change a case writes, loaded.
func changedFixture(test *testing.T, change func(test *testing.T, folder string)) *content.Snapshot {
	test.Helper()
	folder := fixtures.Copy(test, fixtures.ContentFolder(test))
	change(test, folder)
	return fixtures.LoadedFrom(test, folder)
}

// A revision the fixture post's revisions folder gains, the version of the 11th, whose body is its file's 5th line
const (
	sampleRevisionFile = "posts/reading-slowly/revisions/2026-11-11.en.md"
	sampleRevisionHead = "---\ntitle: Reading slowly\ndescription: A lead since rewritten.\n---\n"
	sampleRevisionLine = 5
)

// revisionLinkingNowhere writes a revision of the fixture's post whose link leads to no page the site serves.
//
//wrapper:a case of the table names the change it writes
func revisionLinkingNowhere(test *testing.T, folder string) {
	fixtures.WriteIn(test, folder, sampleRevisionFile, sampleRevisionHead+"[a](/nowhere)\n")
}
