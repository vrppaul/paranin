package compose

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// The days a case's posts of other years come out on: one in the year before the fixture's post, and one in the
// fixture's year by its own zone, though in the year before it by UTC's
const (
	yearBefore    = "2025-03-01T09:00+01:00"
	newYearsNight = "2026-01-01T00:30+01:00"
)

// How a case writes when a post came out, as its front matter does, and what joins the tags of a post the list names
const (
	writtenMomentLayout = "2006-01-02T15:04Z07:00"
	seenTagJoint        = ","
)

// seenGroup is a group of the blog's list as a test reads it: its title, the name a place reaches it by, a series'
// line, how many posts it holds, and its posts.
type seenGroup struct {
	title string
	name  string
	line  string
	count string
	posts []seenListed
}

// seenListed is a post a list names as a test reads it: whether it is out, at its page and when, or to come, as fmt
// writes the value; its title; its part, its series' title and its series' name, each where it says one; its tags.
type seenListed struct {
	readiness  string
	title      string
	part       string
	series     string
	seriesName string
	tags       string
}

// The blog's list by year holds every post the blog lists, never a draft, a preview, a post unlisted or one that comes
// later than its reader, a group a year, the newest year first and the newest post first in each, a post's year the
// one it came out in by its own zone; each says how many posts it holds, one post apart from more, and each post its
// page and when it came out, and its part and its series where it is a part of one.
func TestTheListByYearHoldsEveryPostListedNewestFirst(test *testing.T) {
	posts := []writtenPost{
		{slug: "first", id: 3, status: "published", published: earlier, tags: "reading"},
		{slug: "old-one", id: 4, status: "published", published: yearBefore, tags: ""},
		{slug: "new-year", id: 5, status: "published", published: newYearsNight, tags: ""},
		{slug: "hidden", id: 6, status: "unlisted", published: later, tags: ""},
		{slug: "previewed", id: 7, status: "preview", published: later, tags: ""},
		{slug: "coming", id: 8, status: "published", published: afterTheDay, tags: ""},
	}
	series := fixtureSeriesOf("first, hidden, reading-slowly, previewed, coming", "")
	want := []seenGroup{
		{
			title: "2026", name: "", line: "", count: "3 posts",
			posts: []seenListed{
				{
					readiness: outAt(test, fixturePostSlug, fixturePostMoment), title: fixturePostTitle, part: "part 2", series: fixtureSeriesTitle,
					seriesName: fixtureSeriesName, tags: "reading,terminals",
				},
				{readiness: outAt(test, "first", earlier), title: "first", part: "part 1", series: fixtureSeriesTitle, seriesName: fixtureSeriesName, tags: "reading"},
				{readiness: outAt(test, "new-year", newYearsNight), title: "new-year", part: "", series: "", seriesName: "", tags: ""},
			},
		},
		{
			title: "2025", name: "", line: "", count: "1 post",
			posts: []seenListed{{readiness: outAt(test, "old-one", yearBefore), title: "old-one", part: "", series: "", seriesName: "", tags: ""}},
		},
	}
	lists := SiteOf(withPostsAndSeries(test, posts, series), publicOnTheDay()).Lists
	if got := seenGroups(lists[0]); !reflect.DeepEqual(got, want) {
		test.Errorf("the list by year = %+v, want %+v", got, want)
	}
}

// The blog's list by series holds a group a series with a part out, the series whose latest part came out last first,
// each named by the series' name, as a place reaches it, with its line, how many of its parts, out and to come, are
// out, and each of them in order by its number, a part named to come saying it comes, with no page; a part not
// published, a draft, a preview, unlisted or coming later than its reader, is left out, as its series leaves it, and a
// series with no part out yet is in no list. The posts in no series come last, newest first.
func TestTheListBySeriesHoldsEverySeriesThenThePostsInNone(test *testing.T) {
	posts := []writtenPost{
		{slug: "first", id: 3, status: "published", published: earlier, tags: ""},
		{slug: "drafted", id: 4, status: "draft", published: later, tags: ""},
		{slug: "hidden", id: 5, status: "unlisted", published: later, tags: ""},
		{slug: "toolbox", id: 6, status: "published", published: latest, tags: "tools"},
		{slug: "idle", id: 7, status: "draft", published: later, tags: ""},
		{slug: "old-one", id: 8, status: "published", published: yearBefore, tags: ""},
		{slug: "new-year", id: 9, status: "published", published: newYearsNight, tags: ""},
		{slug: "coming", id: 10, status: "published", published: afterTheDay, tags: ""},
		{slug: "upcoming", id: 11, status: "published", published: afterTheDay, tags: ""},
	}
	series := []writtenSeries{
		{name: fixtureSeriesName, title: fixtureSeriesTitle, parts: "first, hidden, reading-slowly, drafted, coming", toCome: "Reading aloud"},
		{name: "tools", title: "Tools", parts: "toolbox", toCome: ""},
		{name: "quiet", title: "Quiet", parts: "idle", toCome: ""},
		{name: "later", title: "Later", parts: "upcoming", toCome: ""},
	}
	toCome := fmt.Sprint(document.ComingPart{Coming: textSpans("soon")})
	want := []seenGroup{
		{
			title: "Tools", name: "tools", line: "About Tools.", count: "1 of 1 out",
			posts: []seenListed{{readiness: outAt(test, "toolbox", latest), title: "toolbox", part: "part 1", series: "", seriesName: "tools", tags: "tools"}},
		},
		{
			title: fixtureSeriesTitle, name: fixtureSeriesName, line: "About Reading.", count: "2 of 3 out",
			posts: []seenListed{
				{readiness: outAt(test, "first", earlier), title: "first", part: "part 1", series: "", seriesName: fixtureSeriesName, tags: ""},
				{
					readiness: outAt(test, fixturePostSlug, fixturePostMoment), title: fixturePostTitle, part: "part 2", series: "",
					seriesName: fixtureSeriesName, tags: "reading,terminals",
				},
				{readiness: toCome, title: "Reading aloud", part: "part 3", series: "", seriesName: fixtureSeriesName, tags: ""},
			},
		},
		{
			title: "other posts", name: "", line: "", count: "2 posts",
			posts: []seenListed{
				{readiness: outAt(test, "new-year", newYearsNight), title: "new-year", part: "", series: "", seriesName: "", tags: ""},
				{readiness: outAt(test, "old-one", yearBefore), title: "old-one", part: "", series: "", seriesName: "", tags: ""},
			},
		},
	}
	lists := SiteOf(withPostsAndSeries(test, posts, series), publicOnTheDay()).Lists
	if got := seenGroups(lists[1]); !reflect.DeepEqual(got, want) {
		test.Errorf("the list by series = %+v, want %+v", got, want)
	}
}

// Both groupings of the blog's list open with one head: the blog's name, its line as the lead and the page's
// description, and the tags of the posts it lists, each once, the most used first, those used alike by name; a draft's
// tag is none of them. Each stands at its own page.
func TestTheListsHeadNamesTheTagsOfTheListedPosts(test *testing.T) {
	posts := []writtenPost{
		{slug: "first", id: 3, status: "published", published: earlier, tags: "writing, terminals"},
		{slug: "drafted", id: 4, status: "draft", published: later, tags: "secret, writing"},
	}
	snapshot := withPostsAndSeries(test, posts, fixtureSeriesOf(fixturePostSlug, ""))
	want := []string{"terminals", "reading", "writing"}
	pages := []document.PageKind{document.PageBlog, document.PageBlogSeries}
	for index, list := range SiteOf(snapshot, publicOnTheDay()).Lists {
		head := list.Head
		if head.Page.Kind != pages[index] || !reflect.DeepEqual(head.Tags, want) || document.PlainText(head.Title) != snapshot.Words.Text(content.WordBlog) ||
			document.PlainText(head.Lead) != "What the writer reads, and how." || document.PlainText(head.Description) != "What the writer reads, and how." {
			test.Errorf("the list's head = %+v, want the page of kind %d, the tags %q and the blog's name and line", head, pages[index], want)
		}
	}
}

// Each grouping of the blog's list has for places each post out it names, in its own order, by its slug at its page,
// then each series with a group, by its name at its group on the list by series, in the groups' order; never a post
// the list leaves out, nor a series with no part out, nor one for a post in none.
func TestBothListsPlacesAreItsPostsThenItsSeries(test *testing.T) {
	posts := []writtenPost{
		{slug: "first", id: 3, status: "published", published: earlier, tags: ""},
		{slug: "toolbox", id: 4, status: "published", published: latest, tags: ""},
		{slug: "hidden", id: 5, status: "unlisted", published: later, tags: ""},
		{slug: "idle", id: 6, status: "draft", published: later, tags: ""},
		{slug: "loner", id: 7, status: "published", published: earliest, tags: ""},
	}
	series := []writtenSeries{
		{name: fixtureSeriesName, title: fixtureSeriesTitle, parts: "first, hidden, reading-slowly", toCome: ""},
		{name: "tools", title: "Tools", parts: "toolbox", toCome: ""},
		{name: "quiet", title: "Quiet", parts: "idle", toCome: ""},
	}
	bySeries := document.PageOf(document.PageBlogSeries)
	toolbox := document.Place{Name: "toolbox", Kind: document.PlacePost, Page: document.Page{Kind: document.PagePost, Slug: "toolbox", Tab: ""}, Section: ""}
	read := document.Place{Name: fixturePostSlug, Kind: document.PlacePost, Page: document.Page{Kind: document.PagePost, Slug: fixturePostSlug, Tab: ""}, Section: ""}
	first := document.Place{Name: "first", Kind: document.PlacePost, Page: document.Page{Kind: document.PagePost, Slug: "first", Tab: ""}, Section: ""}
	loner := document.Place{Name: "loner", Kind: document.PlacePost, Page: document.Page{Kind: document.PagePost, Slug: "loner", Tab: ""}, Section: ""}
	seriesPlaces := []document.Place{
		{Name: "tools", Kind: document.PlaceSeries, Page: bySeries, Section: "tools"},
		{Name: fixtureSeriesName, Kind: document.PlaceSeries, Page: bySeries, Section: fixtureSeriesName},
	}
	lists := SiteOf(withPostsAndSeries(test, posts, series), publicOnTheDay()).Lists
	cases := []struct {
		name string
		list document.Document
		want []document.Place
	}{
		{name: "by year, newest first", list: lists[0], want: slices.Concat([]document.Place{toolbox, read, first, loner}, seriesPlaces)},
		{name: "by series, by its groups", list: lists[1], want: slices.Concat([]document.Place{toolbox, first, read, loner}, seriesPlaces)},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := document.PlacesOf(testCase.list); !reflect.DeepEqual(got, testCase.want) {
				test.Errorf("the places of the list %+v = %+v, want %+v", testCase.list.Head.Page, got, testCase.want)
			}
		})
	}
}

// Every face's :e opens each project's page by its slug, then the blog's list in each grouping and each post the list
// names, newest first, which Tab offers; then a post served and no list names, by its slug, and every post served by
// its number, which Tab never offers; never a draft, a preview or a post that comes later than its reader, each by the
// name a terminal reaches it by.
func TestThePagesEditOpensAreTheProjectsThenTheBlogs(test *testing.T) {
	posts := []writtenPost{
		{slug: "first", id: 3, status: "published", published: earlier, tags: ""},
		{slug: "drafted", id: 4, status: "draft", published: later, tags: ""},
		{slug: "hidden", id: 5, status: "unlisted", published: later, tags: ""},
		{slug: "coming", id: 6, status: "published", published: afterTheDay, tags: ""},
		{slug: "coming-unlisted", id: 7, status: "unlisted", published: afterTheDay, tags: ""},
	}
	read, first, hidden := postPageOf(fixturePostSlug), postPageOf("first"), postPageOf("hidden")
	want := []OpenedPage{
		{Place: document.Place{Name: "tidy-tool", Kind: document.PlaceProject, Page: document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""}, Section: ""}, Offered: true},
		{Place: document.Place{Name: "blog", Kind: document.PlaceList, Page: document.PageOf(document.PageBlog), Section: ""}, Offered: true},
		{Place: document.Place{Name: "blog/series", Kind: document.PlaceList, Page: document.PageOf(document.PageBlogSeries), Section: ""}, Offered: true},
		{Place: document.Place{Name: "blog/" + fixturePostSlug, Kind: document.PlacePost, Page: read, Section: ""}, Offered: true},
		{Place: document.Place{Name: "blog/first", Kind: document.PlacePost, Page: first, Section: ""}, Offered: true},
		{Place: document.Place{Name: "blog/hidden", Kind: document.PlacePost, Page: hidden, Section: ""}, Offered: false},
		{Place: document.Place{Name: "blog/3", Kind: document.PlacePost, Page: first, Section: ""}, Offered: false},
		{Place: document.Place{Name: "blog/5", Kind: document.PlacePost, Page: hidden, Section: ""}, Offered: false},
		{Place: document.Place{Name: "blog/1", Kind: document.PlacePost, Page: read, Section: ""}, Offered: false},
	}
	if got := SiteOf(withPostsAndSeries(test, posts, nil), publicOnTheDay()).Frame.Pages; !reflect.DeepEqual(got, want) {
		test.Errorf("the pages :e opens = %+v, want %+v", got, want)
	}
}

// postPageOf is the page of a post of a slug.
func postPageOf(slug string) document.Page {
	return document.Page{Kind: document.PagePost, Slug: slug, Tab: ""}
}

// seenGroups are the groups of a list as a test reads them, each a section holding its posts.
func seenGroups(list document.Document) []seenGroup {
	groups := []seenGroup{}
	for _, block := range list.Blocks {
		section := block.(document.Section)
		posts := section.Blocks[0].(document.Posts)
		group := seenGroup{
			title: document.PlainText(section.Title), name: section.Name, line: document.PlainText(posts.Line), count: document.PlainText(posts.Count),
			posts: []seenListed{},
		}
		for _, post := range posts.List {
			group.posts = append(group.posts, seenListedOf(post))
		}
		groups = append(groups, group)
	}
	return groups
}

// seenListedOf is a post a list names as a test reads it.
func seenListedOf(post document.ListedPost) seenListed {
	return seenListed{
		readiness: fmt.Sprint(post.Readiness), title: document.PlainText(post.Title), part: document.PlainText(post.Part),
		series: document.PlainText(post.Series), seriesName: post.SeriesName, tags: strings.Join(post.Tags, seenTagJoint),
	}
}

// outAt is a post out as a test expects a list to name it, by its slug and when it came out as a case writes it, as
// fmt writes the value.
func outAt(test *testing.T, slug, moment string) string {
	test.Helper()
	return fmt.Sprint(document.PostOut{Page: document.Page{Kind: document.PagePost, Slug: slug, Tab: ""}, Published: momentOf(test, moment)})
}

// momentOf is a moment as a case writes it, as front matter does.
func momentOf(test *testing.T, written string) time.Time {
	test.Helper()
	moment, err := time.Parse(writtenMomentLayout, written)
	if err != nil {
		test.Fatalf("read the moment %q: %v", written, err)
	}
	return moment
}
