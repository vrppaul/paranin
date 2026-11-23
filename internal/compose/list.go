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

// The count of posts a group holds that the site's words write as one post, not as posts
const onePost = 1

// Lists are the blog's list, a document every face draws (D25), in its two groupings: every post the blog lists by
// year, newest first, and the same posts by series, the posts in no series last. Each opens with the blog's head.
func Lists(snapshot *content.Snapshot) []document.Document {
	read, words := blogOf(snapshot), snapshot.Words
	return []document.Document{
		{Head: listHead(document.PageBlog, read, words), Blocks: read.byYear(words)},
		{Head: listHead(document.PageBlogSeries, read, words), Blocks: read.bySeries(words)},
	}
}

// blogListings are the blog's pages as the site's index lists them: the blog's list in each grouping, by its own
// head's title named with the grouping, and its description, the blog's line; then each post the list names, newest
// first, by its title and its description.
func blogListings(snapshot *content.Snapshot) []document.Listing {
	read, words := blogOf(snapshot), snapshot.Words
	listings := []document.Listing{}
	for _, grouping := range listWordsOf(words).Groupings() {
		head := listHead(grouping.List.Kind, read, words)
		title := words.Phrase(content.PhraseGroupedList, document.PlainText(head.Title), grouping.Words)
		listings = append(listings, document.Listing{Page: head.Page, Title: textSpans(title), Note: head.Description})
	}
	for _, post := range read.newestFirst() {
		listings = append(listings, document.Listing{Page: postPage(post), Title: textSpans(post.Title), Note: textSpans(post.Description)})
	}
	return listings
}

// listHead is the head of the blog's list in a grouping: the blog's name its title, its line the lead and the page's
// description, and the tags of the posts it lists.
func listHead(kind document.PageKind, read blog, words content.Words) document.Head {
	line := textSpans(words.Text(content.WordBlogLine))
	return document.Head{
		Page:        document.PageOf(kind),
		Number:      0,
		Language:    words.Language(),
		Greeting:    "",
		Title:       textSpans(words.Text(content.WordBlog)),
		Subtitle:    nil,
		Place:       "",
		Tags:        listedTags(read.listed),
		Lead:        line,
		Published:   time.Time{},
		Repository:  nil,
		Post:        nil,
		Description: line,
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

// byYear are the posts the blog lists, a group a year, newest first, as each came out in its own zone; each post
// says where it stands in its series, which its group does not.
func (read blog) byYear(words content.Words) []document.Block {
	years := []int{}
	postsOf := map[int][]document.ListedPost{}
	for _, post := range read.newestFirst() {
		year := post.Published.At().Year()
		if _, seen := postsOf[year]; !seen {
			years = append(years, year)
		}
		postsOf[year] = append(postsOf[year], read.listedOutside(post, words))
	}
	groups := make([]document.Block, 0, len(years))
	for _, year := range years {
		posts := postsOf[year]
		groups = append(groups, postGroup(strconv.Itoa(year), nil, postCount(len(posts), words), posts))
	}
	return groups
}

// bySeries are the posts the blog lists, a group a series, the series whose latest part came out last first, then
// the posts in no series, newest first. A series' group holds every part in order, those to come among them, and
// says how many are out; a series with no part out yet is in no list.
func (read blog) bySeries(words content.Words) []document.Block {
	groups := []document.Block{}
	for _, series := range read.seriesNewestFirst() {
		parts := read.partsOf(series)
		posts := make([]document.ListedPost, 0, len(parts))
		out := 0
		for index, part := range parts {
			posts = append(posts, partInItsGroup(part, partNumber(index), words))
			if part.post != nil {
				out++
			}
		}
		count := textSpans(words.Phrase(content.PhrasePartsOut, strconv.Itoa(out), strconv.Itoa(len(parts))))
		groups = append(groups, postGroup(series.Title, textSpans(series.Line), count, posts))
	}
	others := []document.ListedPost{}
	for _, post := range read.newestFirst() {
		if _, inSeries := read.seriesOf[post.Slug]; !inSeries {
			others = append(others, read.listedOutside(post, words))
		}
	}
	if len(others) > 0 {
		groups = append(groups, postGroup(words.Text(content.WordOutsideSeries), nil, postCount(len(others), words), others))
	}
	return groups
}

// postGroup is a group of the blog's list: a section under its name, holding its posts, with a series' line where it
// is a series' and how many posts it holds.
func postGroup(title string, line, count []document.Span, posts []document.ListedPost) document.Section {
	return document.Section{Title: textSpans(title), Blocks: []document.Block{document.Posts{Line: line, Count: count, List: posts}}}
}

// postCount is how many posts a group holds, in the site's words, one post apart from more.
func postCount(posts int, words content.Words) []document.Span {
	phrase := content.PhrasePosts
	if posts == onePost {
		phrase = content.PhraseOnePost
	}
	return textSpans(words.Phrase(phrase, strconv.Itoa(posts)))
}

// newestFirst are the posts the blog lists, the one out last first.
func (read blog) newestFirst() []content.Post {
	posts := slices.Clone(read.listed)
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
// tags, and its part and its series' title where it is a part of one.
func (read blog) listedOutside(post content.Post, words content.Words) document.ListedPost {
	listed := listedPost(post)
	series, inSeries := read.seriesOf[post.Slug]
	if !inSeries {
		return listed
	}
	parts := read.partsOf(series)
	index := slices.IndexFunc(parts, func(part listedPart) bool { return part.post != nil && part.post.Slug == post.Slug })
	listed.Part = textSpans(words.Phrase(content.PhrasePart, strconv.Itoa(partNumber(index))))
	listed.Series = textSpans(series.Title)
	return listed
}

// partInItsGroup is a part of a series as its own group names it, by its number: a part out as the post it is, a part
// to come by its title alone, saying it comes.
func partInItsGroup(part listedPart, number int, words content.Words) document.ListedPost {
	partWords := textSpans(words.Phrase(content.PhrasePart, strconv.Itoa(number)))
	if part.post == nil {
		return document.ListedPost{
			Readiness: document.ComingPart{Coming: textSpans(words.Text(content.WordSoon))},
			Title:     textSpans(part.title), Lead: nil, Part: partWords, Series: nil, Tags: nil,
		}
	}
	listed := listedPost(*part.post)
	listed.Part = partWords
	return listed
}

// listedPost is a post out as a list names it, in no series: its page, when it came out, its title, its lead and its
// tags.
func listedPost(post content.Post) document.ListedPost {
	return document.ListedPost{
		Readiness: document.PostOut{Page: postPage(post), Published: post.Published.At()},
		Title:     textSpans(post.Title), Lead: textSpans(post.Description), Part: nil, Series: nil, Tags: slices.Clone(post.Tags),
	}
}
