package compose

import (
	"cmp"
	"slices"
	"time"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// The number of a series' first part, as a reader counts them
const firstPart = 1

// blog is what each post's place in the blog and the blog's lists are read from, read once from a snapshot for an
// audience: the audience; every post by its slug; the posts the site serves them, in the snapshot's order, and those
// it serves them at their preview links alone; the posts the lists name, those published and out, in the order they
// came out and newest first, and their tags; every series in their names' order; each series by the slug of every post
// it names; the filters of the list; and its feeds, and the posts each holds.
type blog struct {
	audience  Audience
	posts     map[string]content.Post
	served    []content.Post
	previewed []content.Post
	listed    []content.Post
	newest    []content.Post
	tags      []string
	series    []content.Series
	seriesOf  map[string]content.Series
	filters   []listFilter
	feeds     []Feed
	feedPosts map[document.Feed][]content.Post
}

// listedPart is a part of a series as its reader holds it: its title, and its post where it is out or is the part
// read.
type listedPart struct {
	title string
	post  *content.Post
}

// blogOf is the blog of a snapshot's posts and series, as an audience reads it.
func blogOf(snapshot *content.Snapshot, audience Audience) blog {
	read := blog{
		audience: audience, posts: map[string]content.Post{}, served: []content.Post{}, previewed: []content.Post{}, listed: []content.Post{},
		newest: nil, tags: nil, series: snapshot.Series, seriesOf: map[string]content.Series{}, filters: nil, feeds: nil,
		feedPosts: nil,
	}
	for _, post := range snapshot.Posts {
		read.posts[post.Slug] = post
		if audience.serves(post) {
			read.served = append(read.served, post)
		}
		if audience.previews(post) {
			read.previewed = append(read.previewed, post)
		}
		if audience.lists(post) {
			read.listed = append(read.listed, post)
		}
	}
	// the snapshot holds the posts in their slugs' order, which a stable sort keeps among posts out at one moment
	slices.SortStableFunc(read.listed, func(first, second content.Post) int { return first.Published.At().Compare(second.Published.At()) })
	for _, series := range snapshot.Series {
		for _, slug := range series.Posts {
			read.seriesOf[slug] = series
		}
	}
	read.newest, read.tags = newestFirst(read.listed), listedTags(read.listed)
	read.filters = read.filtersOf()
	read.feeds, read.feedPosts = read.feedsOf(snapshot)
	return read
}

// placeOf is where a post stands in the blog: its series and the posts beside it, and where it is discussed.
func (read blog) placeOf(post content.Post) *document.PostPlace {
	series, neighbours := read.seriesAndNeighbours(post)
	return &document.PostPlace{Series: series, Neighbours: neighbours, Discussions: discussionLinks(post.Discussions)}
}

// seriesAndNeighbours are the series a post is a part of, with every part where it stands for the post's reader, and
// the posts beside it, the parts before and after it; outside a series, or as a part the series leaves out, no
// series, and the posts the lists name before and after it by the time it comes out.
func (read blog) seriesAndNeighbours(post content.Post) (*document.Series, document.Neighbours) {
	series, inSeries := read.seriesOf[post.Slug]
	parts := read.partsOf(series, func(part content.Post) bool {
		return read.audience.lists(part) || part.Slug == post.Slug && read.audience.showsItsSeries(part)
	})
	reading := slices.IndexFunc(parts, func(part listedPart) bool { return part.post != nil && part.post.Slug == post.Slug })
	if !inSeries || reading < 0 {
		return nil, read.byDate(post)
	}
	documentParts := make([]document.Part, 0, len(parts))
	for index, part := range parts {
		documentParts = append(documentParts, document.Part{
			Number: partNumber(index),
			Title:  textSpans(part.title),
			State:  partState(parts, index, reading),
			Page:   part.page(),
		})
	}
	neighbours := document.Neighbours{Before: partNeighbour(parts, reading-1), After: partNeighbour(parts, reading+1)}
	return &document.Series{Title: textSpans(series.Title), Parts: documentParts}, neighbours
}

// partsOf are a series' parts in their order, as its list or the reader of one of its posts holds them: each post it
// names that shows, then each title it names to come. A list shows the posts it names; a post's page those and itself.
// Any other part not published, a draft, a preview or unlisted, is listed nowhere, so the series leaves it out and
// keeps its working title off the site.
func (read blog) partsOf(series content.Series, shows func(content.Post) bool) []listedPart {
	parts := []listedPart{}
	for _, slug := range series.Posts {
		if post := read.posts[slug]; shows(post) {
			parts = append(parts, listedPart{title: post.Title, post: &post})
		}
	}
	for _, title := range series.ToCome {
		parts = append(parts, listedPart{title: title, post: nil})
	}
	return parts
}

// byDate are the posts the lists name before and after a post by the time it comes out, or at the same moment and
// before or after it by slug, as the lists order them; a post no list names, unlisted or a draft, stands among them
// where that time puts it.
func (read blog) byDate(post content.Post) document.Neighbours {
	others := slices.DeleteFunc(slices.Clone(read.listed), func(other content.Post) bool { return other.Slug == post.Slug })
	comes := read.comesOutAt(post)
	after := slices.IndexFunc(others, func(other content.Post) bool {
		return cmp.Or(other.Published.At().Compare(comes), cmp.Compare(other.Slug, post.Slug)) > 0
	})
	if after < 0 {
		after = len(others)
	}
	return document.Neighbours{Before: datedNeighbour(others, after-1), After: datedNeighbour(others, after)}
}

// comesOutAt is when a post comes out: when it is dated, or for a draft not yet dated as the audience reads it, which
// is when it would come out were it published then.
func (read blog) comesOutAt(post content.Post) time.Time {
	if dated := post.Published.At(); !dated.IsZero() {
		return dated
	}
	return read.audience.At
}

// partState is where a part stands for the reader of the part read: the part read, the part out after it, which is
// read next, any other part out, or a part to come.
func partState(parts []listedPart, index, reading int) document.PartState {
	switch {
	case index == reading:
		return document.PartReading
	case parts[index].post == nil:
		return document.PartToCome
	case index == reading+1:
		return document.PartNext
	}
	return document.PartOut
}

// partNeighbour is the part at an index as the post beside the part read: its number, its title, and, once it is out,
// its lead and its page; none where the series has no part there. Inside a series the parts are read in order, so
// when one came out is none of the reader's.
func partNeighbour(parts []listedPart, index int) *document.Neighbour {
	if index < 0 || index >= len(parts) {
		return nil
	}
	part := parts[index]
	var lead []document.Span
	if part.post != nil {
		lead = textSpans(part.post.Description)
	}
	return &document.Neighbour{Part: partNumber(index), Title: textSpans(part.title), Lead: lead, Page: part.page(), Published: time.Time{}}
}

// datedNeighbour is the post at an index of those the lists name as the post beside another: its title, its lead, its
// page and when it came out, which the blog orders them by; none where no post stands there.
func datedNeighbour(posts []content.Post, index int) *document.Neighbour {
	if index < 0 || index >= len(posts) {
		return nil
	}
	post := posts[index]
	return &document.Neighbour{Part: 0, Title: textSpans(post.Title), Lead: textSpans(post.Description), Page: postPage(post), Published: post.Published.At()}
}

// page is a part's page, none for a part to come.
func (part listedPart) page() document.Page {
	if part.post == nil {
		return document.PageOf(document.PageNone)
	}
	return postPage(*part.post)
}

// partNumber is the number of a part at an index of its series, counted from one.
func partNumber(index int) int {
	return index + firstPart
}

// postPage is a post's page, its current version's, by its slug.
func postPage(post content.Post) document.Page {
	return document.Page{Kind: document.PagePost, Slug: post.Slug, Tab: "", Revision: ""}
}

// discussionLinks are where a post is discussed, each a link to the thread by its site's name.
func discussionLinks(discussions []content.Discussion) []document.Link {
	links := make([]document.Link, 0, len(discussions))
	for _, discussion := range discussions {
		links = append(links, document.Link{Target: document.Address{URL: discussion.Address.Written()}, Spans: textSpans(discussion.Site)})
	}
	return links
}
