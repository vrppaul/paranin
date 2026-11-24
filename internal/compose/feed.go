package compose

import (
	"slices"
	"time"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// The most posts a feed holds, its newest: a reader that polls keeps what it read, so it needs only what came since,
// and one who subscribes the last months'; every poll fetches the feed whole, which this bounds as the posts grow
const feedLength = 20

// Feed is a feed of the blog as a feed reader polls it (docs/FOUNDATION.md §5.4, Feeds): the feed a link names; the
// language it is written in; its title and its line; its author and the page that is theirs; where a person reads what
// it holds; when it last changed, as its newest version did; and its posts, newest first.
type Feed struct {
	Target     document.Feed
	Language   string
	Title      []document.Span
	Subtitle   []document.Span
	Author     string
	AuthorPage document.Page
	Shown      document.Target
	Updated    time.Time
	Posts      []FeedPost
}

// FeedPost is a post as a feed holds it: the document of its current version, and when the post came out and when
// that version did, which is no sooner.
type FeedPost struct {
	Document  document.Document
	Published time.Time
	Updated   time.Time
}

// feedsOf are the blog's feeds (docs/FOUNDATION.md §5.4, Feeds), their posts still to compose, and the posts each
// holds, its newest the lists name: the blog's, where its list names a post, then each series' with a part out, in the
// order its filter stands, holding its parts out; titled in the blog's words by the author's name.
func (read blog) feedsOf(snapshot *content.Snapshot) ([]Feed, map[document.Feed][]content.Post) {
	if len(read.newest) == 0 {
		return nil, nil
	}
	words, translation, author := read.words, read.words.Translation(), plainText(snapshot.Profile.Person.Name)
	blogTitle := words.Phrase(content.PhraseBlogFeed, author)
	ofBlog := document.Feed{Series: "", Translation: translation}
	feeds := []Feed{feedOf(ofBlog, blogTitle, words.Text(content.WordBlogLine), document.PageOf(document.PageBlog).InLanguage(translation), read.words, author)}
	posts := map[document.Feed][]content.Post{ofBlog: newestOf(read.newest)}
	for _, series := range read.seriesNewestFirst() {
		ofSeries := document.Feed{Series: series.Name, Translation: translation}
		title := words.Phrase(content.PhraseSeriesFeed, series.Title, blogTitle)
		feeds = append(feeds, feedOf(ofSeries, title, series.Line, read.seriesGroup(series), read.words, author))
		posts[ofSeries] = newestOf(read.partsOut(series))
	}
	return feeds, posts
}

// feedOf is a feed by its title, its line and where a person reads what it holds, written by its author in the words'
// language, holding no post yet.
func feedOf(target document.Feed, title, line string, shown document.Target, words content.Words, author string) Feed {
	return Feed{
		Target: target, Language: words.Language(), Title: textSpans(title), Subtitle: textSpans(line),
		Author: author, AuthorPage: document.PageOf(document.PageResume), Shown: shown, Updated: time.Time{}, Posts: nil,
	}
}

// newestOf are the newest of posts, newest first, as many as a feed holds.
func newestOf(newest []content.Post) []content.Post {
	return newest[:min(len(newest), feedLength)]
}

// feedLinks are the feeds a page names, each by its title (document.Head's Feeds): the feed of the series it shows,
// where it shows one with a feed, then the blog's, where the blog has one.
func (read blog) feedLinks(seriesShown string) []document.Link {
	ofSeries, ofBlog := []document.Link{}, []document.Link{}
	for _, feed := range read.feeds {
		switch feed.Target.Series {
		// the blog's feed follows no series
		case "":
			ofBlog = append(ofBlog, document.Link{Target: feed.Target, Spans: feed.Title})
		case seriesShown:
			ofSeries = append(ofSeries, document.Link{Target: feed.Target, Spans: feed.Title})
		}
	}
	return slices.Concat(ofSeries, ofBlog)
}

// seriesShownOn is the name of the series a post's own page shows, none where it shows none (Audience's
// showsItsSeries) or the post is a part of none.
func (read blog) seriesShownOn(post content.Post) string {
	if !read.audience.showsItsSeries(post) {
		return ""
	}
	return read.seriesOf[post.Slug].Name
}

// seriesShownBy is the name of the series a page of the blog's list shows, a series' filter's; none on any other.
// exhaustive holds the switch to every kind.
func seriesShownBy(page document.Page) string {
	switch page.Kind {
	case document.PageSeriesFilter:
		return page.Slug
	case document.PageNone, document.PageResume, document.PageProject, document.PageIndex, document.PageHealth, document.PageSkills,
		document.PagePost, document.PageBlog, document.PageBlogSeries, document.PageTagFilter:
	}
	return ""
}

// feedsAt are the blog's feeds by their paths, which a post's link inside the site may name.
func feedsAt(feeds []Feed) map[string]document.Target {
	targets := make(map[string]document.Target, len(feeds))
	for _, feed := range feeds {
		targets[routes.FeedPath(feed.Target)] = feed.Target
	}
	return targets
}

// withPosts are the blog's feeds, each holding its posts by the document of each one's current version among the
// posts composed, and changed last when the newest of their versions came out.
func (read blog) withPosts(composed []document.Document) []Feed {
	documents := make(map[string]document.Document, len(composed))
	for _, post := range composed {
		documents[post.Head.Page.Slug] = post
	}
	feeds := make([]Feed, 0, len(read.feeds))
	for _, feed := range read.feeds {
		posts := make([]FeedPost, 0, len(read.feedPosts[feed.Target]))
		for _, post := range read.feedPosts[feed.Target] {
			posts = append(posts, FeedPost{Document: documents[post.Slug], Published: post.Published.At(), Updated: versionOutAt(post)})
		}
		feed.Posts, feed.Updated = posts, lastUpdate(posts)
		feeds = append(feeds, feed)
	}
	return feeds
}

// versionOutAt is when a post's current version came out: the start of its day in the post's own zone, but never
// before the post itself came out, so a version of the day it came out came out with it.
func versionOutAt(post content.Post) time.Time {
	published, day := post.Published.At(), post.VersionDay().Start()
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, published.Location())
	if start.Before(published) {
		return published
	}
	return start
}

// lastUpdate is when the newest version among a feed's posts came out, the zero time where it holds none.
func lastUpdate(posts []FeedPost) time.Time {
	last := time.Time{}
	for _, post := range posts {
		if post.Updated.After(last) {
			last = post.Updated
		}
	}
	return last
}
