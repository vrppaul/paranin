package html

import (
	"slices"

	"github.com/a-h/templ"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// What aria-current says of a link that is the one shown among its kind: the page shown, or the current one of a set
// whose link leads to another page than the one shown, as a filtered list's grouping and filter do
const (
	currentPage = "page"
	currentWay  = "true"
)

// grouping is a way the blog's list groups its posts, as its head's Switch offers it (DESIGN.md, Switch): its word,
// the address of the list grouped so, and what aria-current says of it, empty where it is not the grouping shown.
type grouping struct {
	words   string
	address string
	current string
}

// List writes the blog's list as a page of HTML (DESIGN.md, Pages): what the page says of itself, under its title
// (listTitle); then its head, with the Switch between its groupings and its filters, its groups of tiles, and the foot
// line, in the blog's session, the list its one window, its command line going to the list's posts, tags and series
// too.
func (writer PageWriter) List(composed document.Document) string {
	head := composed.Head
	page := head.Page
	around := writer.frame
	windows := around.blogWindowsOf(around.listWords.BlogWindows(page, document.PageOf(document.PageNone)), page)
	foot := footLineOf(around, around.sessionsOf(page), windows, page, document.PlacesOf(composed))
	main := writer.listMain(head, around.groupingsOf(around.groupings, page), composed.Blocks)
	return rendered(writer.page(head.Language, writer.metadataOf(head, page, around.listTitle(head), nil), main, foot))
}

// listTitle is the title the blog's list says of itself, so no two lists share one: the blog's name, with the filter
// after it where it is filtered, or the word of its grouping where it is grouped otherwise than the blog's own list.
func (around frame) listTitle(head document.Head) []document.Span {
	if len(head.Subtitle) > 0 {
		return joinedTitle(head.Title, dash, head.Subtitle)
	}
	shown := slices.IndexFunc(around.groupings, func(way compose.Grouping) bool { return way.List == head.Page })
	if shown < 0 || head.Page == around.blogList {
		return head.Title
	}
	return joinedTitle(head.Title, dot, []document.Span{document.Text{Text: around.groupings[shown].Words}})
}

// groupingsOf are the ways the blog's list groups its posts as its head's Switch links them, the list shown current,
// or the one a filtered list filters.
func (around frame) groupingsOf(groupings []compose.Grouping, shown document.Page) []grouping {
	grouped, isFilter := shown.Unfiltered()
	linked := make([]grouping, 0, len(groupings))
	for _, way := range groupings {
		current := ""
		switch {
		case way.List == shown:
			current = currentPage
		case way.List == grouped && isFilter:
			current = currentWay
		}
		linked = append(linked, grouping{words: way.Words, address: routes.Reference(way.List, routes.FormatChosen), current: current})
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
