package compose

import (
	"fmt"
	"reflect"
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

// seenGroup is a group of the blog's list as a test reads it: its name, a series' line, how many posts it holds, and
// its posts.
type seenGroup struct {
	title string
	line  string
	count string
	posts []seenListed
}

// seenListed is a post a list names as a test reads it: whether it is out, at its page and when, or to come, as fmt
// writes the value; its title; its part and its series' title, each where it says one; its tags.
type seenListed struct {
	readiness string
	title     string
	part      string
	series    string
	tags      string
}

// The blog's list by year holds every post the blog lists, never a draft, a preview or a post unlisted, a group a
// year, the newest year first and the newest post first in each, a post's year the one it came out in by its own
// zone; each says how many posts it holds, one post apart from more, and each post its page and when it came out,
// and its part and its series where it is a part of one.
func TestTheListByYearHoldsEveryPostListedNewestFirst(test *testing.T) {
	posts := []writtenPost{
		{slug: "first", id: 3, status: "published", published: earlier, tags: "reading"},
		{slug: "old-one", id: 4, status: "published", published: yearBefore, tags: ""},
		{slug: "new-year", id: 5, status: "published", published: newYearsNight, tags: ""},
		{slug: "hidden", id: 6, status: "unlisted", published: later, tags: ""},
		{slug: "previewed", id: 7, status: "preview", published: later, tags: ""},
	}
	series := fixtureSeriesOf("first, hidden, reading-slowly, previewed", "")
	want := []seenGroup{
		{
			title: "2026", line: "", count: "3 posts",
			posts: []seenListed{
				{readiness: outAt(test, fixturePostSlug, fixturePostMoment), title: fixturePostTitle, part: "part 2", series: fixtureSeriesTitle, tags: "reading,terminals"},
				{readiness: outAt(test, "first", earlier), title: "first", part: "part 1", series: fixtureSeriesTitle, tags: "reading"},
				{readiness: outAt(test, "new-year", newYearsNight), title: "new-year", part: "", series: "", tags: ""},
			},
		},
		{
			title: "2025", line: "", count: "1 post",
			posts: []seenListed{{readiness: outAt(test, "old-one", yearBefore), title: "old-one", part: "", series: "", tags: ""}},
		},
	}
	lists := Lists(withPostsAndSeries(test, posts, series))
	if got := seenGroups(lists[0]); !reflect.DeepEqual(got, want) {
		test.Errorf("the list by year = %+v, want %+v", got, want)
	}
}

// The blog's list by series holds a group a series with a part out, the series whose latest part came out last first,
// each with its line, how many of its parts are out, and every part in order by its number, a part not published yet
// or named to come saying it comes, with no page; a part unlisted is left out, as its series leaves it. The posts in
// no series come last, newest first.
func TestTheListBySeriesHoldsEverySeriesThenThePostsInNone(test *testing.T) {
	posts := []writtenPost{
		{slug: "first", id: 3, status: "published", published: earlier, tags: ""},
		{slug: "drafted", id: 4, status: "draft", published: later, tags: ""},
		{slug: "hidden", id: 5, status: "unlisted", published: later, tags: ""},
		{slug: "toolbox", id: 6, status: "published", published: latest, tags: "tools"},
		{slug: "idle", id: 7, status: "draft", published: later, tags: ""},
		{slug: "old-one", id: 8, status: "published", published: yearBefore, tags: ""},
		{slug: "new-year", id: 9, status: "published", published: newYearsNight, tags: ""},
	}
	series := []writtenSeries{
		{name: fixtureSeriesName, title: fixtureSeriesTitle, parts: "first, hidden, reading-slowly, drafted", toCome: "Reading aloud"},
		{name: "tools", title: "Tools", parts: "toolbox", toCome: ""},
		{name: "quiet", title: "Quiet", parts: "idle", toCome: ""},
	}
	toCome := fmt.Sprint(document.ComingPart{Coming: textSpans("soon")})
	want := []seenGroup{
		{
			title: "Tools", line: "About Tools.", count: "1 of 1 parts out",
			posts: []seenListed{{readiness: outAt(test, "toolbox", latest), title: "toolbox", part: "part 1", series: "", tags: "tools"}},
		},
		{
			title: fixtureSeriesTitle, line: "About Reading.", count: "2 of 4 parts out",
			posts: []seenListed{
				{readiness: outAt(test, "first", earlier), title: "first", part: "part 1", series: "", tags: ""},
				{readiness: outAt(test, fixturePostSlug, fixturePostMoment), title: fixturePostTitle, part: "part 2", series: "", tags: "reading,terminals"},
				{readiness: toCome, title: "drafted", part: "part 3", series: "", tags: ""},
				{readiness: toCome, title: "Reading aloud", part: "part 4", series: "", tags: ""},
			},
		},
		{
			title: "other posts", line: "", count: "2 posts",
			posts: []seenListed{
				{readiness: outAt(test, "new-year", newYearsNight), title: "new-year", part: "", series: "", tags: ""},
				{readiness: outAt(test, "old-one", yearBefore), title: "old-one", part: "", series: "", tags: ""},
			},
		},
	}
	lists := Lists(withPostsAndSeries(test, posts, series))
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
	for index, list := range Lists(snapshot) {
		head := list.Head
		if head.Page.Kind != pages[index] || !reflect.DeepEqual(head.Tags, want) || document.PlainText(head.Title) != snapshot.Words.Text(content.WordBlog) ||
			document.PlainText(head.Lead) != "What the writer reads, and how." || document.PlainText(head.Description) != "What the writer reads, and how." {
			test.Errorf("the list's head = %+v, want the page of kind %d, the tags %q and the blog's name and line", head, pages[index], want)
		}
	}
}

// seenGroups are the groups of a list as a test reads them, each a section holding its posts.
func seenGroups(list document.Document) []seenGroup {
	groups := []seenGroup{}
	for _, block := range list.Blocks {
		section := block.(document.Section)
		posts := section.Blocks[0].(document.Posts)
		group := seenGroup{
			title: document.PlainText(section.Title), line: document.PlainText(posts.Line), count: document.PlainText(posts.Count),
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
		series: document.PlainText(post.Series), tags: strings.Join(post.Tags, seenTagJoint),
	}
}

// outAt is a post out as a test expects a list to name it, by its slug and when it came out as a case writes it, as
// fmt writes the value.
func outAt(test *testing.T, slug, moment string) string {
	test.Helper()
	published, err := time.Parse(writtenMomentLayout, moment)
	if err != nil {
		test.Fatalf("read the moment %q: %v", moment, err)
	}
	return fmt.Sprint(document.PostOut{Page: document.Page{Kind: document.PagePost, Slug: slug, Tab: ""}, Published: published})
}
