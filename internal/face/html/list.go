package html

import (
	"github.com/a-h/templ"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// postGroup is a group of the blog's list as its page draws it: its name, and the posts it holds.
type postGroup struct {
	title []document.Span
	posts document.Posts
}

// grouping is a way the blog's list groups its posts, as its head's Switch offers it (DESIGN.md, Switch): its word,
// the address of the list grouped so, and whether it is the grouping shown.
type grouping struct {
	words   string
	address string
	current bool
}

// List writes the blog's list as a page of HTML (DESIGN.md, Pages): what the page says of itself, then its head, with
// the Switch between its groupings, its groups of tiles, and the foot line, in the blog's session, the list its one
// window.
func (writer PageWriter) List(composed document.Document) string {
	head := composed.Head
	page := head.Page
	windows := []window{windowOf(firstWindowNumber, writer.frame.listWords.Window, page, page)}
	foot := footLineOf(writer.frame, writer.site, writer.frame.sessionsOf(true), windows, page)
	main := writer.listMain(head, writer.groupingsOf(page.Kind), composed.Blocks)
	return rendered(writer.page(head.Language, writer.metadataOf(head, page, head.Title), nil, main, foot))
}

// groupingsOf are the ways the blog's list groups its posts, by year and by series, the one of a page's kind shown.
func (writer PageWriter) groupingsOf(shown document.PageKind) []grouping {
	words := writer.frame.listWords
	return []grouping{groupingOf(words.Years, document.PageBlog, shown), groupingOf(words.Series, document.PageBlogSeries, shown)}
}

// groupingOf is the grouping of the list of a kind, by its word, the current one where it is the kind shown.
func groupingOf(words string, kind, shown document.PageKind) grouping {
	return grouping{words: words, address: routes.Reference(listPage(kind), routes.FormatChosen), current: kind == shown}
}

// listBlock is a block of the blog's list: a section holding the posts it names, as a group; any other block, which
// compose writes none of into the list, as any page draws it.
func (writer PageWriter) listBlock(block document.Block) templ.Component {
	if group, isGroup := groupOf(block); isGroup {
		return writer.postGroup(group)
	}
	return writer.block(block, firstHeadingDepth)
}

// groupOf is a block as a group of the blog's list, where it is one: a section holding the posts it names alone.
func groupOf(block document.Block) (postGroup, bool) {
	section, isSection := block.(document.Section)
	if !isSection || len(section.Blocks) != 1 {
		return postGroup{title: nil, posts: document.Posts{Line: nil, Count: nil, List: nil}}, false
	}
	posts, isPosts := section.Blocks[0].(document.Posts)
	return postGroup{title: section.Title, posts: posts}, isPosts
}

// tile is a post the list names as its tile, by whether it is out: a link to it, or a part still to come, off.
func (writer PageWriter) tile(post document.ListedPost) templ.Component {
	switch typed := post.Readiness.(type) {
	case document.PostOut:
		return writer.tileOut(post, typed)
	case document.ComingPart:
		return writer.tileToCome(post, typed)
	}
	return templ.NopComponent
}

// listPage is the page of the blog's list of a kind, by year or by series.
func listPage(kind document.PageKind) document.Page {
	return document.Page{Kind: kind, Slug: "", Tab: ""}
}

// partInEdge is whether a tile names its post's part in its top edge: in its series' own group, which names the
// series, so the part alone says where it stands.
func partInEdge(post document.ListedPost) bool {
	return len(post.Part) > 0 && len(post.Series) == 0
}
