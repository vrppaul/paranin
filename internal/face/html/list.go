package html

import (
	"github.com/a-h/templ"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

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
	windows := blogWindowsOf(writer.frame.listWords.BlogWindows(page, document.PageOf(document.PageNone)), page)
	foot := footLineOf(writer.frame, writer.site, writer.frame.sessionsOf(page), windows, page)
	main := writer.listMain(head, groupingsOf(writer.frame.listWords.Groupings(), page), composed.Blocks)
	return rendered(writer.page(head.Language, writer.metadataOf(head, page, head.Title), nil, main, foot))
}

// groupingsOf are the ways the blog's list groups its posts as its head's Switch links them, the list shown current.
func groupingsOf(groupings []compose.Grouping, shown document.Page) []grouping {
	linked := make([]grouping, 0, len(groupings))
	for _, way := range groupings {
		linked = append(linked, grouping{words: way.Words, address: routes.Reference(way.List, routes.FormatChosen), current: way.List == shown})
	}
	return linked
}

// listBlock is a block of the blog's list: a section holding the posts it names, as a group; any other block, which
// compose writes none of into the list, as any page draws it.
func (writer PageWriter) listBlock(block document.Block) templ.Component {
	if group, isGroup := document.GroupOf(block); isGroup {
		return writer.postGroup(group)
	}
	return writer.block(block, firstHeadingDepth)
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
