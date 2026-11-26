package compose

import (
	"cmp"
	"maps"
	"slices"
	"strconv"
	"time"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// blogLists are the blog's list, a document every face draws (D25), in its two groupings: every post the blog lists
// by year, newest first, and the same posts by series, the posts in no series last. Each opens with the blog's head.
func blogLists(read blog) []document.Document {
	words, translation := read.words, read.words.Translation()
	return []document.Document{
		{Head: read.listHead(document.PageOf(document.PageBlog).InLanguage(translation), nil, words), Blocks: read.byYear(read.newest, words)},
		{Head: read.listHead(document.PageOf(document.PageBlogSeries).InLanguage(translation), nil, words), Blocks: read.bySeries(words)},
	}
}

// listFilter is a filter of the blog's list: the list it makes, the tag it filters by, and the posts it keeps, newest
// first.
type listFilter struct {
	page  document.Page
	tag   string
	posts []content.Post
}

// filtersOf are the filters of the blog's list (D25), by each tag its head names, in its order, each keeping the posts
// the list names that carry the tag.
func (read blog) filtersOf() []listFilter {
	translation := read.words.Translation()
	filters := make([]listFilter, 0, len(read.tags))
	for _, tag := range read.tags {
		kept := slices.DeleteFunc(slices.Clone(read.newest), func(post content.Post) bool { return !slices.Contains(post.Tags, tag) })
		filters = append(filters, listFilter{page: document.TagFilter(tag).InLanguage(translation), tag: tag, posts: kept})
	}
	return filters
}

// partsOut are the parts of a series the lists name, newest first.
func (read blog) partsOut(series content.Series) []content.Post {
	return slices.DeleteFunc(slices.Clone(read.newest), func(post content.Post) bool { return read.seriesOf[post.Slug].Name != series.Name })
}

// filteredLists are the blog's list by year filtered by each of its filters, a view of the list each, not a page of its
// own (docs/FOUNDATION.md §5.4): its head the list's, with the filter in the site's words beside the blog's name.
func filteredLists(read blog) []document.Document {
	words := read.words
	lists := make([]document.Document, 0, len(read.filters))
	for _, filter := range read.filters {
		lists = append(lists, document.Document{Head: read.listHead(filter.page, filter.words(words), words), Blocks: read.byYear(filter.posts, words)})
	}
	return lists
}

// words are a filter in the site's words: "posts tagged go".
func (filter listFilter) words(words content.Words) []document.Span {
	return textSpans(words.Phrase(content.PhraseTagFilter, filter.tag))
}

// listFilters are the filters the head of the blog's list offers on a page of it (DESIGN.md, BlogHead), every tag of
// the posts the list names. The filter the list shown is filtered by leads back to the whole list, every other to the
// list it filters.
func (read blog) listFilters(shown document.Page, words content.Words) []document.Filter {
	filters := make([]document.Filter, 0, len(read.filters))
	for _, filter := range read.filters {
		filters = append(filters, filter.offered(shown, words))
	}
	return filters
}

// offered is a filter as a head on a page offers it: leading to the list it filters, or, on that list, where it is
// the filter shown, back to the whole list, named as the way there.
func (filter listFilter) offered(shown document.Page, words content.Words) document.Filter {
	offered := document.Filter{Tag: filter.tag, Page: filter.page, Current: false, WayOut: ""}
	if filter.page == shown {
		offered.Page, _ = filter.page.Unfiltered()
		offered.Current, offered.WayOut = true, words.Phrase(content.PhraseFilterWayOut, document.PlainText(filter.words(words)))
	}
	return offered
}

// postFilters are the filters a post's head offers: each of its tags, leading to the list it filters where the lists
// name the post; a draft's and an unlisted post's lead nowhere, since a tag of theirs may filter none.
func (read blog) postFilters(post content.Post) []document.Filter {
	filters := make([]document.Filter, 0, len(post.Tags))
	for _, tag := range post.Tags {
		leads := document.PageOf(document.PageNone)
		if read.audience.lists(post) {
			leads = document.TagFilter(tag).InLanguage(read.words.Translation())
		}
		filters = append(filters, document.Filter{Tag: tag, Page: leads, Current: false, WayOut: ""})
	}
	return filters
}

// blogListings are the blog's pages as the site's index lists them: the blog's list in each grouping, by its own
// head's title named with the grouping, and its description, the blog's line; then each post the list names, newest
// first, by its title and its description.
func blogListings(read blog, words content.Words) []document.Listing {
	listings := []document.Listing{}
	for _, grouping := range read.listWords().Groupings(read.words.Translation()) {
		head := read.listHead(grouping.List, nil, words)
		title := words.Phrase(content.PhraseGroupedList, document.PlainText(head.Title), grouping.Words)
		listings = append(listings, document.Listing{Target: head.Page, Title: textSpans(title), Note: head.Description})
	}
	for _, post := range read.newest {
		listings = append(listings, document.Listing{Target: postPage(post), Title: textSpans(post.Title), Note: textSpans(post.Description)})
	}
	return listings
}

// listHead is the head of the blog's list as a page shows it, in a grouping or filtered: the blog's name its title, the
// filter beside it, none where it filters none, its line the lead and the page's description, the filters it offers,
// the tags of every post the blog lists among them, and the feeds that follow it.
func (read blog) listHead(page document.Page, filter []document.Span, words content.Words) document.Head {
	line := textSpans(words.Text(content.WordBlogLine))
	return document.Head{
		Page:        page,
		Number:      0,
		Language:    words.Language(),
		Greeting:    "",
		Title:       textSpans(words.Text(content.WordBlog)),
		Subtitle:    filter,
		Place:       nil,
		Filters:     read.listFilters(page, words),
		Lead:        line,
		Version:     nil,
		Project:     nil,
		Post:        nil,
		Description: line,
		Feeds:       read.feedLinks(noSeries),
		Languages:   nil,
		Latest:      nil,
	}
}

// listedTags are the tags of the posts the blog lists, each once, the most used first, those used alike by name.
func listedTags(posts []content.Post) []string {
	uses := map[string]int{}
	for _, post := range posts {
		for _, tag := range post.Tags {
			uses[tag]++
		}
	}
	return slices.SortedFunc(maps.Keys(uses), func(first, second string) int {
		return cmp.Or(cmp.Compare(uses[second], uses[first]), cmp.Compare(first, second))
	})
}

// byYear are posts the blog lists, newest first, a group a year, as each came out in its own zone; each post says
// where it stands in its series, which its group does not.
func (read blog) byYear(newest []content.Post, words content.Words) []document.Block {
	years := []int{}
	postsOf := map[int][]document.ListedPost{}
	for _, post := range newest {
		year := post.Published.At().Year()
		if _, seen := postsOf[year]; !seen {
			years = append(years, year)
		}
		postsOf[year] = append(postsOf[year], read.listedOutside(post, words))
	}
	groups := make([]document.Block, 0, len(years))
	for _, year := range years {
		posts := postsOf[year]
		groups = append(groups, postGroup(strconv.Itoa(year), "", nil, postCount(len(posts), words), posts))
	}
	return groups
}

// bySeries are the posts the blog lists, a group a series, the series whose latest part came out last first, then
// the posts in no series, newest first. A series' group holds every part in order, those to come among them, and
// says how many are out; a series with no part out yet is in no list.
func (read blog) bySeries(words content.Words) []document.Block {
	groups := []document.Block{}
	for _, series := range read.seriesNewestFirst() {
		parts := read.partsOf(series, read.audience.lists)
		posts := make([]document.ListedPost, 0, len(parts))
		out := 0
		for index, part := range parts {
			posts = append(posts, partInItsGroup(part, partNumber(index), read.seriesGroup(series), words))
			if part.post != nil {
				out++
			}
		}
		count := textSpans(words.Count(content.CountPartsOut, out, strconv.Itoa(len(parts))))
		groups = append(groups, postGroup(series.Title, series.Name, textSpans(series.Line), count, posts))
	}
	others := []document.ListedPost{}
	for _, post := range read.newest {
		if _, inSeries := read.seriesOf[post.Slug]; !inSeries {
			others = append(others, read.listedOutside(post, words))
		}
	}
	if len(others) > 0 {
		groups = append(groups, postGroup(words.Text(content.WordOutsideSeries), "", nil, postCount(len(others), words), others))
	}
	return groups
}

// postGroup is a group of the blog's list: a section under its title, holding its posts, with a series' name, which a
// place reaches the group by, and its line where it is a series' and how many posts it holds.
func postGroup(title, name string, line, count []document.Span, posts []document.ListedPost) document.Section {
	return document.Section{Title: textSpans(title), Name: name, Blocks: []document.Block{document.Posts{Line: line, Count: count, List: posts}}}
}

// postCount is how many posts a group holds, in the site's words.
func postCount(posts int, words content.Words) []document.Span {
	return textSpans(words.Count(content.CountPosts, posts))
}

// newestFirst are posts in the order they came out turned round, the one out last first.
func newestFirst(listed []content.Post) []content.Post {
	posts := slices.Clone(listed)
	slices.Reverse(posts)
	return posts
}

// seriesNewestFirst are the series with a part out, the one whose latest part came out last first, those of one moment
// in their names' order.
func (read blog) seriesNewestFirst() []content.Series {
	latest := map[string]time.Time{}
	for _, post := range read.listed {
		if series, inSeries := read.seriesOf[post.Slug]; inSeries {
			latest[series.Name] = post.Published.At()
		}
	}
	out := slices.DeleteFunc(slices.Clone(read.series), func(series content.Series) bool { return latest[series.Name].IsZero() })
	slices.SortStableFunc(out, func(first, second content.Series) int { return latest[second.Name].Compare(latest[first.Name]) })
	return out
}

// listedOutside is a post as a group that is no series' own names it: when it came out, its title, its lead, its
// tags, and its part, its series' title and its series' name where it is a part of one.
func (read blog) listedOutside(post content.Post, words content.Words) document.ListedPost {
	listed := listedPost(post)
	series, inSeries := read.seriesOf[post.Slug]
	if !inSeries {
		return listed
	}
	parts := read.partsOf(series, read.audience.lists)
	index := slices.IndexFunc(parts, func(part listedPart) bool { return part.post != nil && part.post.Slug == post.Slug })
	listed.Part = textSpans(words.Phrase(content.PhrasePart, strconv.Itoa(partNumber(index))))
	listed.Series, listed.SeriesGroup = textSpans(series.Title), read.seriesGroup(series)
	return listed
}

// seriesGroup is the group of a series on the blog's list by series, in the blog's language.
func (read blog) seriesGroup(series content.Series) document.Anchor {
	return document.Anchor{Page: document.PageOf(document.PageBlogSeries).InLanguage(read.words.Translation()), Name: series.Name}
}

// partInItsGroup is a part of a series as its own group names it, by its number: a part out as the post it is, a part
// to come by its title alone, saying it comes.
func partInItsGroup(part listedPart, number int, group document.Anchor, words content.Words) document.ListedPost {
	partWords := textSpans(words.Phrase(content.PhrasePart, strconv.Itoa(number)))
	if part.post == nil {
		return document.ListedPost{
			Readiness: document.ComingPart{Coming: textSpans(words.Text(content.WordSoon))},
			Title:     textSpans(part.title), Lead: nil, Part: partWords, Series: nil, SeriesGroup: group, Tags: nil,
		}
	}
	listed := listedPost(*part.post)
	listed.Part, listed.SeriesGroup = partWords, group
	return listed
}

// latestPosts are the blog's latest posts as the résumé's first screen names them (DESIGN.md, LatestPosts): the blog's
// name and the way to its list by year, in its words, then the latest posts its lists name, newest first, each a post
// out by its title alone, as a small tile holds it; none where the lists name none.
func (read blog) latestPosts() *document.LatestPosts {
	if len(read.newest) == 0 {
		return nil
	}
	posts := make([]document.ListedPost, 0, document.LatestPostsPlaces)
	for _, post := range read.newest[:min(document.LatestPostsPlaces, len(read.newest))] {
		listed := listedPost(post)
		listed.Lead, listed.Tags = nil, nil
		posts = append(posts, listed)
	}
	words := read.words
	list := document.PageOf(document.PageBlog).InLanguage(words.Translation())
	return &document.LatestPosts{
		Title: textSpans(words.Text(content.WordBlog)),
		All:   document.Link{Target: list, Spans: textSpans(words.Text(content.WordAllPosts))},
		Posts: posts,
	}
}

// listedPost is a post out as a list names it, in no series: its page, when it came out, its title, its lead and its
// tags.
func listedPost(post content.Post) document.ListedPost {
	outside := document.Anchor{Page: document.PageOf(document.PageNone), Name: ""}
	return document.ListedPost{
		Readiness: document.PostOut{Page: postPage(post), Published: post.Published.At()},
		Title:     textSpans(post.Title), Lead: textSpans(post.Description), Part: nil, Series: nil, SeriesGroup: outside, Tags: slices.Clone(post.Tags),
	}
}
