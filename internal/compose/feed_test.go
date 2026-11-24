package compose

import (
	"fmt"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// When the fixture's post came out and when its current version did, the start of its updated: day in the post's own
// zone
const (
	fixturePostPublished = "2026-11-10T09:00:00+01:00"
	fixtureVersionOut    = "2026-11-12T00:00:00+01:00"
)

// How a case writes a post of its own past the fixture's: its slug and when it came out, a minute after the one before
const (
	manyPostsSlug   = "post-%d"
	manyPostsMoment = "2026-11-11T09:%02d+01:00"
	firstManyID     = 10
)

// seenFeed is a feed as a test reads it: the feed a link names, its title, when it last changed, and its posts, newest
// first.
type seenFeed struct {
	feed    document.Feed
	title   string
	updated string
	posts   []seenFeedPost
}

// seenFeedPost is a post as a feed holds it, as a test reads it: its slug, and when the post and its current version
// came out, as RFC 3339 writes them.
type seenFeedPost struct {
	slug      string
	published string
	updated   string
}

// The blog's feed holds the posts its list names, newest first, never a draft, a preview, a post unlisted or one that
// comes later than its reader; a series' holds its parts out, and a series with no part out has none. A feed changed
// last when the newest version among its posts came out, which may be an older post's.
func TestEachFeedHoldsTheListedPostsNewestFirst(test *testing.T) {
	posts := []writtenPost{
		{slug: "first", id: 3, status: "published", published: earlier, tags: ""},
		{slug: "old-one", id: 4, status: "published", published: yearBefore, tags: ""},
		{slug: "fresh", id: 9, status: "published", published: "2026-11-11T09:00+01:00", tags: ""},
		{slug: "hidden", id: 5, status: "unlisted", published: later, tags: ""},
		{slug: "previewed", id: 6, status: "preview", published: later, tags: ""},
		{slug: "coming", id: 7, status: "published", published: afterTheDay, tags: ""},
		{slug: "drafted", id: 8, status: "draft", published: later, tags: ""},
	}
	series := []writtenSeries{
		{name: fixtureSeriesName, title: fixtureSeriesTitle, parts: "first, hidden, reading-slowly, previewed, coming", toCome: ""},
		{name: "quiet", title: "Quiet", parts: "drafted", toCome: ""},
	}
	fresh := seenFeedPost{slug: "fresh", published: "2026-11-11T09:00:00+01:00", updated: "2026-11-11T09:00:00+01:00"}
	fixturePost := seenFeedPost{slug: fixturePostSlug, published: fixturePostPublished, updated: fixtureVersionOut}
	first := seenFeedPost{slug: "first", published: "2026-11-02T09:00:00+01:00", updated: "2026-11-02T09:00:00+01:00"}
	oldOne := seenFeedPost{slug: "old-one", published: "2025-03-01T09:00:00+01:00", updated: "2025-03-01T09:00:00+01:00"}
	want := []seenFeed{
		{feed: document.Feed{Series: ""}, title: "Mira Example's blog", updated: fixtureVersionOut, posts: []seenFeedPost{fresh, fixturePost, first, oldOne}},
		{feed: document.Feed{Series: fixtureSeriesName}, title: "Reading — Mira Example's blog", updated: fixtureVersionOut, posts: []seenFeedPost{fixturePost, first}},
	}
	feeds := SiteOf(Prepare(withPostsAndSeries(test, posts, series)), publicOnTheDay()).Own.Feeds
	if got := seenFeeds(feeds); !reflect.DeepEqual(got, want) {
		test.Errorf("the feeds = %+v, want %+v", got, want)
	}
}

// A feed holds its newest posts alone, as many as feedLength says, so a poll fetches no more as the posts grow.
func TestAFeedHoldsItsNewestPostsAlone(test *testing.T) {
	posts := postsAMinuteApart(feedLength + 1)
	newest := slices.Clone(posts)
	slices.Reverse(newest)
	want := writtenSlugs(newest[:feedLength])
	feeds := SiteOf(Prepare(withPostsAndSeries(test, posts, nil)), publicOnTheDay()).Own.Feeds
	if got := feedSlugs(feeds[0]); !reflect.DeepEqual(got, want) {
		test.Errorf("the blog's feed holds %q, want %q", got, want)
	}
}

// A post in a feed is updated when its current version came out: when the post did, where it was never updated or
// was updated on the day it came out, and else at the start of the day its updated: says, in the post's own zone.
func TestAPostInAFeedIsUpdatedWhenItsCurrentVersionCameOut(test *testing.T) {
	cases := []struct {
		name    string
		updated string
		want    string
	}{
		{name: "never updated", updated: "", want: fixturePostPublished},
		{name: "updated on the day it came out", updated: "\nupdated: 2026-11-10", want: fixturePostPublished},
		{name: "updated on a later day", updated: "\nupdated: 2026-11-15", want: "2026-11-15T00:00:00+01:00"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			folder := fixtures.Copy(test, fixtures.ContentFolder(test))
			fixtures.ReplaceIn(test, folder, fixturePostFile, fixturePostUpdated, testCase.updated)
			fixtures.RemoveIn(test, folder, fixtureRevisionFile)
			feeds := SiteOf(Prepare(fixtures.LoadedFrom(test, folder)), publicOnTheDay()).Own.Feeds
			if got := feeds[0].Posts[0].Updated.Format(time.RFC3339); got != testCase.want {
				test.Errorf("the fixture post, %s, is updated in its feed at %s, want %s", testCase.name, got, testCase.want)
			}
		})
	}
}

// A blog whose list names no post has no feed, and no page names one.
func TestABlogListingNothingHasNoFeed(test *testing.T) {
	folder := fixtures.Copy(test, fixtures.ContentFolder(test))
	fixturePostDrafted(test, folder)
	site := SiteOf(Prepare(fixtures.LoadedFrom(test, folder)), publicOnTheDay())
	if len(site.Own.Feeds) != 0 || len(site.Own.Lists[0].Head.Feeds) != 0 || len(site.Pages[0].Document.Head.Feeds) != 0 {
		test.Errorf("a blog listing nothing has the feeds %+v, its list names %+v and the résumé %+v, want none", site.Own.Feeds,
			site.Own.Lists[0].Head.Feeds, site.Pages[0].Document.Head.Feeds)
	}
}

// Each page of the blog names in its head the feeds that follow it: the blog's, and first a series' on the series'
// filtered list and on a post whose page shows its series. The résumé and its tabs name the blog's, so a feed reader
// handed the site's address finds it; a project's page names none.
func TestEachPageNamesTheFeedsThatFollowThePage(test *testing.T) {
	posts := []writtenPost{{slug: "hidden", id: 5, status: "unlisted", published: later, tags: ""}}
	series := fixtureSeriesOf("reading-slowly, hidden", "")
	blogFeed := "Mira Example's blog at {Series: Translation:}"
	readingFeed := "Reading — Mira Example's blog at {Series:reading Translation:}"
	cases := []struct {
		name string
		page document.Page
		want []string
	}{
		{name: "the list by year", page: document.PageOf(document.PageBlog), want: []string{blogFeed}},
		{name: "the list by series", page: document.PageOf(document.PageBlogSeries), want: []string{blogFeed}},
		{name: "the list filtered by a tag", page: document.TagFilter("reading"), want: []string{blogFeed}},
		{name: "the list filtered by a series", page: document.SeriesFilter(fixtureSeriesName), want: []string{readingFeed, blogFeed}},
		{name: "a part of a series", page: document.PostPage(fixturePostSlug), want: []string{readingFeed, blogFeed}},
		{name: "an unlisted part, whose page shows no series", page: document.PostPage("hidden"), want: []string{blogFeed}},
		{name: "the résumé", page: document.PageOf(document.PageResume), want: []string{blogFeed}},
		{name: "a tab of the résumé", page: document.Page{Kind: document.PageResume, Slug: "", Tab: "experience", Revision: ""}, want: []string{blogFeed}},
		{name: "a project's page", page: document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: "", Revision: ""}, want: []string{}},
	}
	site := SiteOf(Prepare(withPostsAndSeries(test, posts, series)), publicOnTheDay())
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := seenFeedLinks(fixtures.DocumentOf(test, site.served, testCase.page).Head.Feeds); !reflect.DeepEqual(got, testCase.want) {
				test.Errorf("the head of %+v names the feeds %q, want %q", testCase.page, got, testCase.want)
			}
		})
	}
}

// postsAMinuteApart are as many published posts as a case asks for, each a minute after the one before, after the
// fixture's.
func postsAMinuteApart(count int) []writtenPost {
	posts := make([]writtenPost, 0, count)
	for index := range count {
		posts = append(posts, writtenPost{
			slug: fmt.Sprintf(manyPostsSlug, index), id: firstManyID + index, status: content.StatusPublished, published: fmt.Sprintf(manyPostsMoment, index), tags: "",
		})
	}
	return posts
}

// seenFeeds are feeds as a test reads them.
func seenFeeds(feeds []Feed) []seenFeed {
	seen := make([]seenFeed, 0, len(feeds))
	for _, feed := range feeds {
		posts := make([]seenFeedPost, 0, len(feed.Posts))
		for _, post := range feed.Posts {
			posts = append(posts, seenFeedPost{
				slug: post.Document.Head.Page.Slug, published: post.Published.Format(time.RFC3339), updated: post.Updated.Format(time.RFC3339),
			})
		}
		seen = append(seen, seenFeed{feed: feed.Target, title: document.PlainText(feed.Title), updated: feed.Updated.Format(time.RFC3339), posts: posts})
	}
	return seen
}

// writtenSlugs are the slugs of posts a case writes, in their order.
func writtenSlugs(posts []writtenPost) []string {
	slugs := make([]string, 0, len(posts))
	for _, post := range posts {
		slugs = append(slugs, post.slug)
	}
	return slugs
}

// feedSlugs are the slugs of a feed's posts, in its order.
func feedSlugs(feed Feed) []string {
	slugs := make([]string, 0, len(feed.Posts))
	for _, post := range feed.Posts {
		slugs = append(slugs, post.Document.Head.Page.Slug)
	}
	return slugs
}

// seenFeedLinks are the feeds a head names, each by its title at the feed it leads to.
func seenFeedLinks(links []document.Link) []string {
	seen := make([]string, 0, len(links))
	for _, link := range links {
		seen = append(seen, fmt.Sprintf("%s at %+v", document.PlainText(link.Spans), link.Target))
	}
	return seen
}
