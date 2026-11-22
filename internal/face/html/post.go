package html

import (
	"strconv"

	"github.com/a-h/templ"

	"github.com/vrppaul/paranin/internal/document"
)

// How a post's day is written (DESIGN.md, Content): as a reader reads it, and as a time's datetime holds it
const (
	dayLayout        = "2 Jan 2006"
	machineDayLayout = "2006-01-02"
)

// The window a post is shown in, in the blog's session: the second, after the list of posts, as the TUI numbers them;
// and where the blog's session leads while its list is still to come: nowhere
const (
	postWindowNumber = 2
	stillToCome      = ""
)

// How the places of an article a reader links to are named: a paragraph by its number after paragraphIDOpening, a
// heading by its title's words (document.SlugOf), or headingID where its title has none; a name the page holds already
// takes its count after it, from the second
const (
	paragraphIDOpening = "p"
	headingID          = "heading"
	nameCountJoint     = "-"
	secondCount        = 2
)

// The place an address names inside a page, after its path
const fragmentMark = "#"

// What stands between a note's number and its words, which templ writes no space between
const noteSpace = " "

// Post writes a post as a page of HTML (DESIGN.md, Pages): what the page says of itself, then its head, its article
// and the foot line. A post has no tabs: it is drawn whole, in the blog's session, its window the post's slug. The
// places of its article a reader links to take names no part of the page holds, the foot line's among them.
func (writer PageWriter) Post(composed document.Document) string {
	head := composed.Head
	page := head.Page
	foot := writer.postFoot(page)
	main := writer.postMain(head, writer.articleOf(composed.Blocks, foot.ids()), document.NotesOf(composed.Blocks))
	// a post's record, the BlogPosting its linked data would write, comes with the site's SEO (TODO.md)
	return rendered(writer.page(head.Language, writer.metadataOf(head, page, head.Title), nil, main, foot))
}

// postFoot is the foot line of a post's page: the site's session, then the blog's, which the post belongs to and whose
// list is still to come; the post as its one window, and the post to read in a terminal and as markdown.
func (writer PageWriter) postFoot(page document.Page) footLine {
	sessions := []session{
		{name: writer.frame.siteName, address: writer.frame.start, current: false},
		{name: writer.frame.blog, address: stillToCome, current: true},
	}
	windows := []window{windowOf(postWindowNumber, page.Slug, page, page)}
	return footLineOf(writer.frame, writer.site, sessions, windows, page)
}

// article draws a post's body, each part as it is met: each paragraph at the body's top or a section's numbered, from
// one; each block of code numbered, from one, its button named by its number; each note beside the first block that
// cites it; each heading and numbered paragraph named for a reader to link to, by a name the page holds nowhere else.
type article struct {
	writer     PageWriter
	paragraphs int
	codes      int
	cited      map[int]bool
	names      map[string]bool
}

// articleOf is a post's body as the article draws it, on a page whose other parts hold names already.
func (writer PageWriter) articleOf(blocks []document.Block, held []string) templ.Component {
	names := make(map[string]bool, len(held))
	for _, name := range held {
		names[name] = true
	}
	drawing := &article{writer: writer, paragraphs: 0, codes: 0, cited: map[int]bool{}, names: names}
	return templ.Join(drawing.parts(blocks, firstHeadingDepth)...)
}

// parts are blocks as the article lays them out, a section's at the depth of its heading: each block, then the notes it
// is the first to cite, beside it; a section's notes stand beside the blocks it holds.
func (drawing *article) parts(blocks []document.Block, depth int) []templ.Component {
	parts := make([]templ.Component, 0, len(blocks))
	for _, block := range blocks {
		if section, isSection := block.(document.Section); isSection {
			id := drawing.name(document.SlugOf(document.PlainText(section.Title)))
			parts = append(parts, drawing.writer.articleSection(id, depth, section.Title, drawing.parts(section.Blocks, depth+1)))
			continue
		}
		parts = append(parts, drawing.part(block))
		if notes := drawing.firstCited(block); len(notes) > 0 {
			parts = append(parts, drawing.writer.notesBeside(notes))
		}
	}
	return parts
}

// part is a block of the body that is no section: a paragraph numbered and named, any other as the article draws it.
func (drawing *article) part(block document.Block) templ.Component {
	paragraph, isParagraph := block.(document.Paragraph)
	if !isParagraph {
		return drawing.block(block)
	}
	drawing.paragraphs++
	number := strconv.Itoa(drawing.paragraphs)
	return drawing.writer.numberedParagraph(number, drawing.name(paragraphIDOpening+number), paragraph.Spans)
}

// block is a block of a post as its article draws it, inside a list's item or a quote as at the body's top: a list and
// a quote holding blocks drawn so too, code numbered, with its language and the button that copies it.
func (drawing *article) block(block document.Block) templ.Component {
	switch typed := block.(type) {
	case document.Paragraph:
		return drawing.writer.paragraph(typed)
	case document.Table:
		return drawing.writer.table(typed)
	case document.CodeBlock:
		drawing.codes++
		return copyableCode(typed.Language, drawing.writer.frame.copyWords.CodeOf(typed.Text, drawing.codes))
	case document.List:
		items := make([]templ.Component, 0, len(typed.Items))
		for _, item := range typed.Items {
			items = append(items, drawing.item(item))
		}
		return postList(typed.Ordered, typed.Start, items)
	case document.Quote:
		return postQuote(drawing.blocks(typed.Blocks))
	case document.Section, document.Facts, document.Entry, document.Card, document.Contents, document.Checks, document.Skills:
		// a post's body nests its headings into sections, which parts draws, and holds none of the résumé's blocks; a
		// list's item and a quote hold no heading
	}
	return templ.NopComponent
}

// blocks are blocks a list's item or a quote holds, each as the article draws it.
func (drawing *article) blocks(blocks []document.Block) []templ.Component {
	drawn := make([]templ.Component, 0, len(blocks))
	for _, block := range blocks {
		drawn = append(drawn, drawing.block(block))
	}
	return drawn
}

// item is what an item of a post's list holds: its first block's words on the line of its mark, where that is a
// paragraph, and the blocks after them under them.
func (drawing *article) item(item document.Item) templ.Component {
	if len(item.Blocks) == 0 {
		return templ.NopComponent
	}
	if paragraph, isParagraph := item.Blocks[0].(document.Paragraph); isParagraph {
		return templ.Join(append([]templ.Component{drawing.writer.spans(paragraph.Spans)}, drawing.blocks(item.Blocks[1:])...)...)
	}
	return templ.Join(drawing.blocks(item.Blocks)...)
}

// firstCited are the notes a block cites that no block before it has, in the order it cites them.
func (drawing *article) firstCited(block document.Block) []document.Note {
	notes := []document.Note{}
	for _, note := range document.NotesOf([]document.Block{block}) {
		if !drawing.cited[note.Number] {
			drawing.cited[note.Number] = true
			notes = append(notes, note)
		}
	}
	return notes
}

// name is the name a place of the article is linked to by: the one it asks for, or that with its count after it where
// the page holds the name already, a heading's of the same words or a part's of the page among them; a heading of no
// words asks for headingID.
func (drawing *article) name(asked string) string {
	if asked == unnamed {
		asked = headingID
	}
	given := asked
	for count := secondCount; drawing.names[given]; count++ {
		given = asked + nameCountJoint + strconv.Itoa(count)
	}
	drawing.names[given] = true
	return given
}
