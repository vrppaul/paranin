package compose

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// The fixture's series, by its name and its title, which a case writes over with a series of its own; and how a case
// writes a post of its own and a series
const (
	fixtureSeriesName  = "reading"
	fixtureSeriesTitle = "Reading"
	postFileLayout     = "posts/%s/post.en.md"
	postTextLayout     = "---\nid: %d\ntitle: %s\ndescription: About %s.\nstatus: %s\npublished: %s\ntags: [%s]\n---\nIts text.\n"
	seriesFileLayout   = "series/%s.yaml"
	seriesTextLayout   = "title: %s\nline: About %s.\nposts: [%s]\nto_come: [%s]\n"
)

// The fixture's post's title, and the day it came out, between the days a case's posts come out on
const (
	fixturePostTitle  = "Reading slowly"
	fixturePostMoment = "2026-11-10T09:00+01:00"
	earlier           = "2026-11-02T09:00+01:00"
	earliest          = "2026-11-01T09:00+01:00"
	later             = "2026-11-20T09:00+01:00"
	latest            = "2026-11-21T09:00+01:00"
	// the zero time, as RFC 3339 writes it: when a neighbour inside a series came out, which it does not say
	noTime = "0001-01-01T00:00:00Z"
)

// writtenPost is a post a case writes into the fixture: its slug, its number, its status, when it came out and its
// tags, as YAML lists them, none where a case names none; its title is its slug, its lead says what it is about.
type writtenPost struct {
	slug      string
	id        int
	status    content.Status
	published string
	tags      string
}

// writtenSeries is a series a case writes into the fixture: its name, its title, its parts' slugs and the titles it
// names to come, each as YAML lists them; its line says what it is about.
type writtenSeries struct {
	name   string
	title  string
	parts  string
	toCome string
}

// seenPart is a part of a series as a test reads it: its number, its title, where it stands, and the slug of its
// page, none for a part to come.
type seenPart struct {
	number int
	title  string
	state  document.PartState
	page   string
}

// seenNeighbour is a post beside another as a test reads it: its part, its title, its lead, the slug of its page and
// when it came out, as RFC 3339 writes it, none inside a series.
type seenNeighbour struct {
	part      int
	title     string
	lead      string
	page      string
	published string
}

// A post of a series shows every part of it in order, numbered from one: itself as the part read, the part out after
// it as the next, any other part out as out, and as to come each title the series names to come; a part not published,
// a draft, a preview or unlisted, is listed nowhere, so the series leaves it out, and the parts after it count on
// without it. A post the series leaves out, unlisted, shows no series.
func TestAPostShowsItsSeriesPartByPart(test *testing.T) {
	cases := []struct {
		name   string
		posts  []writtenPost
		series string
		toCome string
		read   string
		want   []seenPart
	}{
		{
			name: "the first part, the rest to come", series: fixturePostSlug, toCome: "Reading aloud", read: fixturePostSlug,
			want: []seenPart{
				{number: 1, title: fixturePostTitle, state: document.PartReading, page: fixturePostSlug},
				{number: 2, title: "Reading aloud", state: document.PartToCome, page: ""},
			},
		},
		{
			name:   "a middle part, between parts out",
			posts:  []writtenPost{{slug: "first", id: 3, status: content.StatusPublished, published: earlier}, {slug: "third", id: 4, status: content.StatusPublished, published: later}, {slug: "fourth", id: 5, status: content.StatusPublished, published: latest}},
			series: "first, reading-slowly, third, fourth", read: fixturePostSlug,
			want: []seenPart{
				{number: 1, title: "first", state: document.PartOut, page: "first"},
				{number: 2, title: fixturePostTitle, state: document.PartReading, page: fixturePostSlug},
				{number: 3, title: "third", state: document.PartNext, page: "third"},
				{number: 4, title: "fourth", state: document.PartOut, page: "fourth"},
			},
		},
		{
			name:   "a part not published yet, left out",
			posts:  []writtenPost{{slug: "drafted", id: 3, status: content.StatusDraft, published: later}, {slug: "previewed", id: 4, status: content.StatusPreview, published: later}},
			series: "reading-slowly, drafted, previewed", toCome: "Reading aloud", read: fixturePostSlug,
			want: []seenPart{
				{number: 1, title: fixturePostTitle, state: document.PartReading, page: fixturePostSlug},
				{number: 2, title: "Reading aloud", state: document.PartToCome, page: ""},
			},
		},
		{
			name:   "a part unlisted, left out",
			posts:  []writtenPost{{slug: "hidden", id: 3, status: content.StatusUnlisted, published: earlier}},
			series: "hidden, reading-slowly", read: fixturePostSlug,
			want: []seenPart{{number: 1, title: fixturePostTitle, state: document.PartReading, page: fixturePostSlug}},
		},
		{
			name:   "the unlisted part itself, in no series",
			posts:  []writtenPost{{slug: "hidden", id: 3, status: content.StatusUnlisted, published: earlier}},
			series: "hidden, reading-slowly", read: "hidden",
			want: nil,
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			snapshot := withPostsAndSeries(test, testCase.posts, fixtureSeriesOf(testCase.series, testCase.toCome))
			series := postNamed(test, Posts(snapshot), testCase.read).Head.Post.Series
			if got := seenParts(series); !reflect.DeepEqual(got, testCase.want) {
				test.Errorf("the series of %q = %+v, want %+v", testCase.read, got, testCase.want)
			}
		})
	}
}

// A post's places are its headings, then each other part of its series out, by its slug, at its page: never the part
// read, which is the page shown, a part to come, which has no page, or a part the series leaves out.
func TestAPostsPlacesAreItsHeadingsThenItsOtherParts(test *testing.T) {
	posts := []writtenPost{
		{slug: "first", id: 3, status: content.StatusPublished, published: earlier},
		{slug: "drafted", id: 4, status: content.StatusDraft, published: later},
		{slug: "third", id: 5, status: content.StatusPublished, published: latest},
	}
	snapshot := withPostsAndSeries(test, posts, fixtureSeriesOf("first, reading-slowly, drafted, third", "Reading aloud"))
	read := document.Page{Kind: document.PagePost, Slug: fixturePostSlug, Tab: ""}
	want := []document.Place{
		{Name: "what-a-reader-keeps", Kind: document.PlaceHeading, Page: read, Section: "what-a-reader-keeps"},
		{Name: "in-a-terminal", Kind: document.PlaceHeading, Page: read, Section: "in-a-terminal"},
		{Name: "afterwards", Kind: document.PlaceHeading, Page: read, Section: "afterwards"},
		{Name: "first", Kind: document.PlacePart, Page: document.Page{Kind: document.PagePost, Slug: "first", Tab: ""}, Section: ""},
		{Name: "third", Kind: document.PlacePart, Page: document.Page{Kind: document.PagePost, Slug: "third", Tab: ""}, Section: ""},
	}
	if got := document.PlacesOf(postNamed(test, Posts(snapshot), fixturePostSlug)); !reflect.DeepEqual(got, want) {
		test.Errorf("the post's places = %+v, want %+v", got, want)
	}
}

// A heading named as another part of its post's series takes its count, so its name leads to one place.
func TestAHeadingNamedAsAPartTakesItsCount(test *testing.T) {
	posts := []writtenPost{{slug: "what-a-reader-keeps", id: 3, status: content.StatusPublished, published: earlier}}
	snapshot := withPostsAndSeries(test, posts, fixtureSeriesOf("what-a-reader-keeps, reading-slowly", ""))
	places := document.PlacesOf(postNamed(test, Posts(snapshot), fixturePostSlug))
	want := []string{"what-a-reader-keeps-2", "in-a-terminal", "afterwards", "what-a-reader-keeps"}
	got := []string{}
	for _, place := range places {
		got = append(got, place.Name)
	}
	if !reflect.DeepEqual(got, want) {
		test.Errorf("the post's places = %q, want %q", got, want)
	}
}

// A post leads at its foot to the posts beside it: in a series the parts before and after it, a part to come by its
// title alone, with no page, past any part not published; outside one the posts the blog lists before and after it by
// the time they came out, each with its lead and that time, posts out at one moment by their slugs, so never a draft,
// a preview or a post unlisted; none on a side with nothing there.
// An unlisted post, in no list itself, leads to the posts listed before and after the time it came out.
func TestAPostLeadsToThePostsBesideThePost(test *testing.T) {
	cases := []struct {
		name   string
		posts  []writtenPost
		series string
		toCome string
		read   string
		before *seenNeighbour
		after  *seenNeighbour
	}{
		{
			name: "the first part, the next to come", series: fixturePostSlug, toCome: "Reading aloud", read: fixturePostSlug,
			before: nil, after: &seenNeighbour{part: 2, title: "Reading aloud", lead: "", page: "", published: noTime},
		},
		{
			name:   "the part before a draft, the next to come past it",
			posts:  []writtenPost{{slug: "drafted", id: 3, status: content.StatusDraft, published: later}},
			series: "reading-slowly, drafted", toCome: "Reading aloud", read: fixturePostSlug,
			before: nil, after: &seenNeighbour{part: 2, title: "Reading aloud", lead: "", page: "", published: noTime},
		},
		{
			name:   "a middle part, between parts out",
			posts:  []writtenPost{{slug: "first", id: 3, status: content.StatusPublished, published: later}, {slug: "third", id: 4, status: content.StatusPublished, published: earliest}},
			series: "first, reading-slowly, third", read: fixturePostSlug,
			before: &seenNeighbour{part: 1, title: "first", lead: "About first.", page: "first", published: noTime},
			after:  &seenNeighbour{part: 3, title: "third", lead: "About third.", page: "third", published: noTime},
		},
		{
			name:   "the last part",
			posts:  []writtenPost{{slug: "first", id: 3, status: content.StatusPublished, published: earlier}},
			series: "first, reading-slowly", read: fixturePostSlug,
			before: &seenNeighbour{part: 1, title: "first", lead: "About first.", page: "first", published: noTime},
			after:  nil,
		},
		{
			name: "outside a series, between the posts listed by date",
			posts: []writtenPost{
				{slug: "oldest", id: 3, status: content.StatusPublished, published: earliest},
				{slug: "older", id: 4, status: content.StatusPublished, published: earlier},
				{slug: "newer", id: 5, status: content.StatusPublished, published: later},
				{slug: "newest", id: 6, status: content.StatusPublished, published: latest},
				{slug: "hidden", id: 7, status: content.StatusUnlisted, published: earlier},
				{slug: "drafted", id: 8, status: content.StatusDraft, published: later},
			},
			series: "oldest", read: fixturePostSlug,
			before: &seenNeighbour{part: 0, title: "older", lead: "About older.", page: "older", published: "2026-11-02T09:00:00+01:00"},
			after:  &seenNeighbour{part: 0, title: "newer", lead: "About newer.", page: "newer", published: "2026-11-20T09:00:00+01:00"},
		},
		{
			name:   "posts out at the same moment, in their slugs' order",
			posts:  []writtenPost{{slug: "another", id: 3, status: content.StatusPublished, published: fixturePostMoment}, {slug: "zeta", id: 4, status: content.StatusPublished, published: fixturePostMoment}},
			series: "zeta", read: fixturePostSlug,
			before: &seenNeighbour{part: 0, title: "another", lead: "About another.", page: "another", published: "2026-11-10T09:00:00+01:00"},
			after:  &seenNeighbour{part: 0, title: "zeta", lead: "About zeta.", page: "zeta", published: "2026-11-10T09:00:00+01:00"},
		},
		{
			name:   "the newest post, outside a series",
			posts:  []writtenPost{{slug: "older", id: 3, status: content.StatusPublished, published: earlier}, {slug: "preview", id: 4, status: content.StatusPreview, published: later}},
			series: "older", read: fixturePostSlug,
			before: &seenNeighbour{part: 0, title: "older", lead: "About older.", page: "older", published: "2026-11-02T09:00:00+01:00"},
			after:  nil,
		},
		{
			name:   "an unlisted post, between the posts listed before and after it",
			posts:  []writtenPost{{slug: "hidden", id: 3, status: content.StatusUnlisted, published: later}, {slug: "newest", id: 4, status: content.StatusPublished, published: latest}},
			series: "newest", read: "hidden",
			before: &seenNeighbour{part: 0, title: fixturePostTitle, lead: "Why a page read twice is read once.", page: fixturePostSlug, published: "2026-11-10T09:00:00+01:00"},
			after:  &seenNeighbour{part: 0, title: "newest", lead: "About newest.", page: "newest", published: "2026-11-21T09:00:00+01:00"},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			snapshot := withPostsAndSeries(test, testCase.posts, fixtureSeriesOf(testCase.series, testCase.toCome))
			neighbours := postNamed(test, Posts(snapshot), testCase.read).Head.Post.Neighbours
			before, after := seenNeighbourOf(neighbours.Before), seenNeighbourOf(neighbours.After)
			if !reflect.DeepEqual(before, testCase.before) || !reflect.DeepEqual(after, testCase.after) {
				test.Errorf("the neighbours of %q = %+v and %+v, want %+v and %+v", testCase.read, before, after, testCase.before, testCase.after)
			}
		})
	}
}

// withPostsAndSeries is the fixture content with posts of a case's own beside its post, and series of its own, one of
// them the fixture's written over, loaded.
func withPostsAndSeries(test *testing.T, posts []writtenPost, series []writtenSeries) *content.Snapshot {
	test.Helper()
	folder := fixtures.Copy(test, fixtures.ContentFolder(test))
	for _, post := range posts {
		text := fmt.Sprintf(postTextLayout, post.id, post.slug, post.slug, post.status, post.published, post.tags)
		fixtures.WriteIn(test, folder, fmt.Sprintf(postFileLayout, post.slug), text)
	}
	for _, written := range series {
		text := fmt.Sprintf(seriesTextLayout, written.title, written.title, written.parts, written.toCome)
		fixtures.WriteIn(test, folder, fmt.Sprintf(seriesFileLayout, written.name), text)
	}
	snapshot, err := content.Load(os.DirFS(folder))
	if err != nil {
		test.Fatalf("load the fixture with the posts %+v and the series %+v: %v", posts, series, err)
	}
	return snapshot
}

// fixtureSeriesOf is the fixture's series written over to hold the parts a case names, then the titles it names to
// come.
func fixtureSeriesOf(parts, toCome string) []writtenSeries {
	return []writtenSeries{{name: fixtureSeriesName, title: fixtureSeriesTitle, parts: parts, toCome: toCome}}
}

// seenParts are a series' parts as a test reads them, none where there is no series.
func seenParts(series *document.Series) []seenPart {
	if series == nil {
		return nil
	}
	parts := []seenPart{}
	for _, part := range series.Parts {
		parts = append(parts, seenPart{number: part.Number, title: document.PlainText(part.Title), state: part.State, page: part.Page.Slug})
	}
	return parts
}

// seenNeighbourOf is a neighbour as a test reads it, none where there is none.
func seenNeighbourOf(neighbour *document.Neighbour) *seenNeighbour {
	if neighbour == nil {
		return nil
	}
	return &seenNeighbour{
		part:      neighbour.Part,
		title:     document.PlainText(neighbour.Title),
		lead:      document.PlainText(neighbour.Lead),
		page:      neighbour.Page.Slug,
		published: neighbour.Published.Format(time.RFC3339),
	}
}
