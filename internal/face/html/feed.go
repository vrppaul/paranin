package html

import (
	"github.com/a-h/templ"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// FeedWriter writes a post's body as a feed reader shows it, inside a feed (docs/WEB.md, "What each format holds"):
// its text as the page writes it, every address whole, since a reader shows it away from the site; its blocks plain,
// with none of the page's own parts, which a reader has no stylesheet to draw and runs no script for, so they would
// show as stray words: no paragraph's number, no point's mark, no button. Its code stands in no colour, and its notes
// at its end. The page's article is no fit: it sets each note beside its line and again at the end, which the
// stylesheet shows one of.
type FeedWriter struct {
	spanWriter
}

// NewFeedWriter writes the bodies of a site's posts, at the whole addresses the site writes.
func NewFeedWriter(site routes.Site) FeedWriter {
	return FeedWriter{spanWriter: spanWriter{addressOf: site.Address}}
}

// Body is a post's body as a feed holds it: its blocks, then every note its lines cite.
func (writer FeedWriter) Body(post document.Document) string {
	notes := []templ.Component{}
	for _, note := range document.NotesOf(post.Blocks) {
		notes = append(notes, writer.feedNote(note))
	}
	return rendered(feedBody(writer.blocks(post.Blocks, firstHeadingDepth), notes))
}

// blocks are blocks of a post's body as a feed draws them, a section's headings at its depth.
func (writer FeedWriter) blocks(blocks []document.Block, depth int) []templ.Component {
	drawn := make([]templ.Component, 0, len(blocks))
	for _, block := range blocks {
		drawn = append(drawn, writer.block(block, depth))
	}
	return drawn
}

// block is a block of a post's body as a feed draws it, a list's items and a quote's blocks drawn so too.
func (writer FeedWriter) block(block document.Block, depth int) templ.Component {
	switch typed := block.(type) {
	case document.Section:
		return writer.feedSection(depth, typed.Title, writer.blocks(typed.Blocks, depth+1))
	case document.Paragraph:
		return writer.paragraph(typed)
	case document.Table:
		return writer.feedTable(typed)
	case document.CodeBlock:
		return feedCode(typed.Text())
	case document.List:
		items := make([]templ.Component, 0, len(typed.Items))
		for _, item := range typed.Items {
			items = append(items, listItem(item, writer.spans, func(blocks []document.Block) []templ.Component { return writer.blocks(blocks, depth) }))
		}
		return feedList(typed.Ordered, typed.Start, items)
	case document.Quote:
		return postQuote(writer.blocks(typed.Blocks, depth))
	case document.Pause:
		return feedPause()
	case document.Figure:
		return writer.feedFigure(typed)
	case document.Widget:
		return writer.feedWidget(typed)
	case document.Facts, document.ExperienceHead, document.Entry, document.Points, document.Side, document.Reference, document.Card,
		document.Contents, document.Checks, document.Skills, document.Posts:
		// a post's body holds none of the résumé's blocks, a project's, the index's or the list's
	}
	return templ.NopComponent
}
