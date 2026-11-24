package compose

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// The days a case's posts of other years come out on: one in the year before the fixture's post, and one in the
// fixture's year by its own zone, though in the year before it by UTC's
const (
	yearBefore    = "2025-03-01T09:00+01:00"
	newYearsNight = "2026-01-01T00:30+01:00"
)

// What joins the tags of a post the list names
const seenTagJoint = ","

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
	lists := SiteOf(Prepare(withPostsAndSeries(test, posts, series)), publicOnTheDay()).Own.Lists
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
	lists := SiteOf(Prepare(withPostsAndSeries(test, posts, series)), publicOnTheDay()).Own.Lists
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
	for index, list := range SiteOf(Prepare(snapshot), publicOnTheDay()).Own.Lists {
		head := list.Head
		if head.Page.Kind != pages[index] || !reflect.DeepEqual(head.ListedWords(), want) || document.PlainText(head.Title) != snapshot.Own.Words.Text(content.WordBlog) ||
			document.PlainText(head.Lead) != "What the writer reads, and how." || document.PlainText(head.Description) != "What the writer reads, and how." {
			test.Errorf("the list's head = %+v, want the page of kind %d, the tags %q and the blog's name and line", head, pages[index], want)
		}
	}
}

// The blog's list is filtered by each tag its head names, in its order, then by each series with a part out, the one
// whose latest part came out last first: each a list by year of the posts the list names that carry the tag or are
// parts of the series, newest first, its head the list's, every tag of the blog among it, with the filter in the
// site's words beside the blog's name. No tag of a post the list leaves out, and no series with no part out, filters
// it.
func TestTheListIsFilteredByEachTagThenEachSeries(test *testing.T) {
	posts := []writtenPost{
		{slug: "first", id: 3, status: "published", published: earlier, tags: "terminals, tools"},
		{slug: "old-one", id: 4, status: "published", published: yearBefore, tags: "tools"},
		{slug: "toolbox", id: 5, status: "published", published: latest, tags: ""},
		{slug: "drafted", id: 6, status: "draft", published: later, tags: "secret"},
		{slug: "idle", id: 7, status: "draft", published: later, tags: ""},
	}
	series := []writtenSeries{
		{name: fixtureSeriesName, title: fixtureSeriesTitle, parts: "first, reading-slowly", toCome: "Reading aloud"},
		{name: "tools", title: "Tools", parts: "toolbox, old-one", toCome: ""},
		{name: "quiet", title: "Quiet", parts: "idle", toCome: ""},
	}
	tags := []string{"terminals", "tools", "reading"}
	fixturePost := seenListed{
		readiness: outAt(test, fixturePostSlug, fixturePostMoment), title: fixturePostTitle, part: "part 2", series: fixtureSeriesTitle,
		seriesName: fixtureSeriesName, tags: "reading,terminals",
	}
	first := seenListed{
		readiness: outAt(test, "first", earlier), title: "first", part: "part 1", series: fixtureSeriesTitle, seriesName: fixtureSeriesName, tags: "terminals,tools",
	}
	oldOne := seenListed{readiness: outAt(test, "old-one", yearBefore), title: "old-one", part: "part 2", series: "Tools", seriesName: "tools", tags: "tools"}
	toolbox := seenListed{readiness: outAt(test, "toolbox", latest), title: "toolbox", part: "part 1", series: "Tools", seriesName: "tools", tags: ""}
	cases := []struct {
		page     document.Page
		subtitle string
		groups   []seenGroup
	}{
		{
			page: document.TagFilter("terminals"), subtitle: "posts tagged terminals",
			groups: []seenGroup{{title: "2026", name: "", line: "", count: "2 posts", posts: []seenListed{fixturePost, first}}},
		},
		{
			page: document.TagFilter("tools"), subtitle: "posts tagged tools",
			groups: []seenGroup{
				{title: "2026", name: "", line: "", count: "1 post", posts: []seenListed{first}},
				{title: "2025", name: "", line: "", count: "1 post", posts: []seenListed{oldOne}},
			},
		},
		{
			page: document.TagFilter("reading"), subtitle: "posts tagged reading",
			groups: []seenGroup{{title: "2026", name: "", line: "", count: "1 post", posts: []seenListed{fixturePost}}},
		},
		{
			page: document.SeriesFilter("tools"), subtitle: "parts of Tools",
			groups: []seenGroup{
				{title: "2026", name: "", line: "", count: "1 post", posts: []seenListed{toolbox}},
				{title: "2025", name: "", line: "", count: "1 post", posts: []seenListed{oldOne}},
			},
		},
		{
			page: document.SeriesFilter(fixtureSeriesName), subtitle: "parts of Reading",
			groups: []seenGroup{{title: "2026", name: "", line: "", count: "2 posts", posts: []seenListed{fixturePost, first}}},
		},
	}
	filtered := SiteOf(Prepare(withPostsAndSeries(test, posts, series)), publicOnTheDay()).Own.Filtered
	if len(filtered) != len(cases) {
		test.Fatalf("the list is filtered %d ways, %+v, want %d", len(filtered), filtered, len(cases))
	}
	for index, testCase := range cases {
		test.Run(testCase.subtitle, func(test *testing.T) {
			head := filtered[index].Head
			if head.Page != testCase.page || document.PlainText(head.Subtitle) != testCase.subtitle || !reflect.DeepEqual(head.ListedWords(), tags) ||
				document.PlainText(head.Title) != "blog" {
				test.Errorf("the filtered list's head = %+v, want the page %+v, the subtitle %q and the tags %q", head, testCase.page, testCase.subtitle, tags)
			}
			if got := seenGroups(filtered[index]); !reflect.DeepEqual(got, testCase.groups) {
				test.Errorf("the list filtered as %+v = %+v, want %+v", testCase.page, got, testCase.groups)
			}
		})
	}
}

// A head of the blog's list offers its filters, each leading to the list it filters: every tag of the posts the list
// names, in the head's order; on a list filtered by a tag that tag current, leading back to the whole list, named to a
// screen reader as the way there; on a list filtered by a series that series first, in the site's words, current and
// leading back too.
func TestAListsHeadOffersItsFilters(test *testing.T) {
	site := SiteOf(Prepare(fixtures.LoadedContent(test)), publicOnTheDay())
	lists := slices.Concat(site.Own.Lists, site.Own.Filtered)
	whole := document.PageOf(document.PageBlog)
	reading := document.Filter{Kind: document.FilterTag, Name: "reading", Words: textSpans("reading"), Page: document.TagFilter("reading"), Current: false, WayOut: ""}
	terminals := document.Filter{Kind: document.FilterTag, Name: "terminals", Words: textSpans("terminals"), Page: document.TagFilter("terminals"), Current: false, WayOut: ""}
	readingShown := document.Filter{Kind: document.FilterTag, Name: "reading", Words: textSpans("reading"), Page: whole, Current: true, WayOut: "posts tagged reading, show all posts"}
	seriesShown := document.Filter{
		Kind: document.FilterSeries, Name: fixtureSeriesName, Words: textSpans("parts of Reading"), Page: whole, Current: true, WayOut: "parts of Reading, show all posts",
	}
	cases := []struct {
		name string
		page document.Page
		want []document.Filter
	}{
		{name: "the list", page: whole, want: []document.Filter{reading, terminals}},
		{name: "the list by series", page: document.PageOf(document.PageBlogSeries), want: []document.Filter{reading, terminals}},
		{name: "filtered by a tag", page: document.TagFilter("reading"), want: []document.Filter{readingShown, terminals}},
		{name: "filtered by a series", page: document.SeriesFilter(fixtureSeriesName), want: []document.Filter{seriesShown, reading, terminals}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := fixtures.DocumentOf(test, lists, testCase.page).Head.Filters; !reflect.DeepEqual(got, testCase.want) {
				test.Errorf("the filters of %+v = %+v, want %+v", testCase.page, got, testCase.want)
			}
		})
	}
}

// Each grouping of the blog's list, and each filter of it, has for places each post out it names, in its own order,
// by its slug at its page, then each tag its head names, at the list it filters, or on that list back at the list,
// then each series with a group, by its name at its group on the list by series, in the groups' order; never a post
// the list leaves out, nor a series with no part out, nor one for a post in none.
func TestAListsPlacesAreItsPostsThenItsTagsThenItsSeries(test *testing.T) {
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
	tagPlaces := []document.Place{
		{Name: "reading", Kind: document.PlaceTag, Page: document.TagFilter("reading"), Section: ""},
		{Name: "terminals", Kind: document.PlaceTag, Page: document.TagFilter("terminals"), Section: ""},
	}
	filteredTagPlaces := []document.Place{
		{Name: "reading", Kind: document.PlaceTag, Page: document.PageOf(document.PageBlog), Section: ""},
		{Name: "terminals", Kind: document.PlaceTag, Page: document.TagFilter("terminals"), Section: ""},
	}
	site := SiteOf(Prepare(withPostsAndSeries(test, posts, series)), publicOnTheDay())
	cases := []struct {
		name string
		list document.Document
		want []document.Place
	}{
		{name: "by year, newest first", list: site.Own.Lists[0], want: slices.Concat([]document.Place{toolbox, read, first, loner}, tagPlaces, seriesPlaces)},
		{name: "by series, by its groups", list: site.Own.Lists[1], want: slices.Concat([]document.Place{toolbox, first, read, loner}, tagPlaces, seriesPlaces)},
		{
			name: "filtered by a tag, which leads back to the list", list: fixtures.DocumentOf(test, site.Own.Filtered, document.TagFilter("reading")),
			want: slices.Concat([]document.Place{read}, filteredTagPlaces, seriesPlaces[1:]),
		},
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
// its number, and the list filtered by each tag and each series, which Tab never offers; never a draft, a preview or a
// post that comes later than its reader, each by the name a terminal reaches it by; then the same of the blog in each
// other language, in its folder. Every language's frame holds that one list.
func TestThePagesEditOpensAreTheProjectsThenTheBlogs(test *testing.T) {
	posts := []writtenPost{
		{slug: "first", id: 3, status: "published", published: earlier, tags: ""},
		{slug: "drafted", id: 4, status: "draft", published: later, tags: ""},
		{slug: "hidden", id: 5, status: "unlisted", published: later, tags: ""},
		{slug: "coming", id: 6, status: "published", published: afterTheDay, tags: ""},
		{slug: "coming-unlisted", id: 7, status: "unlisted", published: afterTheDay, tags: ""},
	}
	read, first, hidden := document.PostPage(fixturePostSlug), document.PostPage("first"), document.PostPage("hidden")
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
		{Place: document.Place{Name: "blog/reading-slowly/2026-11-10", Kind: document.PlaceVersion, Page: document.PostPage(fixturePostSlug).OlderVersion("2026-11-10"), Section: ""}, Offered: false},
		{Place: document.Place{Name: "blog?tag=reading", Kind: document.PlaceList, Page: document.TagFilter("reading"), Section: ""}, Offered: false},
		{Place: document.Place{Name: "blog?tag=terminals", Kind: document.PlaceList, Page: document.TagFilter("terminals"), Section: ""}, Offered: false},
		{Place: document.Place{Name: "blog?series=reading", Kind: document.PlaceList, Page: document.SeriesFilter("reading"), Section: ""}, Offered: false},
		{Place: document.Place{Name: "ru/blog", Kind: document.PlaceList, Page: document.PageOf(document.PageBlog).InLanguage(russian), Section: ""}, Offered: true},
		{Place: document.Place{Name: "ru/blog/series", Kind: document.PlaceList, Page: document.PageOf(document.PageBlogSeries).InLanguage(russian), Section: ""}, Offered: true},
		{Place: document.Place{Name: "ru/blog/" + fixturePostSlug, Kind: document.PlacePost, Page: read.InLanguage(russian), Section: ""}, Offered: true},
		{Place: document.Place{Name: "ru/blog/1", Kind: document.PlacePost, Page: read.InLanguage(russian), Section: ""}, Offered: false},
		{Place: document.Place{Name: "ru/blog?tag=reading", Kind: document.PlaceList, Page: document.TagFilter("reading").InLanguage(russian), Section: ""}, Offered: false},
		{Place: document.Place{Name: "ru/blog?tag=terminals", Kind: document.PlaceList, Page: document.TagFilter("terminals").InLanguage(russian), Section: ""}, Offered: false},
		{Place: document.Place{Name: "ru/blog?series=reading", Kind: document.PlaceList, Page: document.SeriesFilter("reading").InLanguage(russian), Section: ""}, Offered: false},
	}
	for _, blog := range SiteOf(Prepare(withPostsAndSeries(test, posts, nil)), publicOnTheDay()).Blogs() {
		if got := blog.Frame.Pages; !reflect.DeepEqual(got, want) {
			test.Errorf("the pages :e opens in the frame of %q = %+v, want %+v", blog.Frame.Language, got, want)
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
		series: document.PlainText(post.Series), seriesName: post.SeriesGroup.Name, tags: strings.Join(post.Tags, seenTagJoint),
	}
}

// outAt is a post out as a test expects a list to name it, by its slug and when it came out as a case writes it, as
// fmt writes the value.
func outAt(test *testing.T, slug, moment string) string {
	test.Helper()
	return fmt.Sprint(document.PostOut{Page: document.Page{Kind: document.PagePost, Slug: slug, Tab: ""}, Published: fixtures.MomentOf(test, moment)})
}
