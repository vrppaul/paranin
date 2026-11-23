package html

import (
	"slices"
	"strconv"

	"github.com/a-h/templ"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
)

// How a post's day is written (DESIGN.md, Content): as a reader reads it, and as a time's datetime holds it
const (
	dayLayout        = "2 Jan 2006"
	machineDayLayout = "2006-01-02"
)

// The window a post is shown in, in the blog's session: the second, after the list of posts, as the TUI numbers them
const postWindowNumber = 2

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

// What stands between two words templ writes no space between: a note's number and its words, a part's title and the
// word saying it comes, the word before where a post is discussed and the first place
const wordSpace = " "

// The names of the parts of a post's page a link or a button names, beside its article's places: the series' pick in
// its head and in its panes' heading, each window's title after the pick's name, and the panes' heading
const (
	headPickID     = "parts"
	footPickID     = "parts-foot"
	pickTitleEnd   = "-title"
	panesHeadingID = "more"
)

// The parts of a post's page the site's script finds (web/src/controllers/outline.ts): the outline, each link of it,
// and each heading it leads to, the article's top level's
const (
	outlinePart        = "outline"
	outlineLinkPart    = "outline-link"
	outlineHeadingPart = "outline-heading"
)

// Post writes a post as a page of HTML (DESIGN.md, Pages): what the page says of itself, then its head, its outline,
// its article, the panes at its foot and the foot line. A post has no tabs: it is drawn whole, in the blog's session,
// its window the post's slug. The places of its article a reader links to take names no part of the page holds, the
// foot line's and the picks' among them.
func (writer PageWriter) Post(composed document.Document) string {
	head := composed.Head
	page := head.Page
	foot := writer.postFoot(page)
	held := slices.Concat(foot.ids(), pickIDs(headPickID), pickIDs(footPickID), []string{panesHeadingID})
	drawn := writer.articleOf(composed.Blocks, held)
	parts := postParts{
		head:    head,
		outline: drawn.outline,
		article: templ.Join(drawn.parts...),
		notes:   document.NotesOf(composed.Blocks),
		panes:   writer.panesOf(head.Post),
	}
	// a post's record, the BlogPosting its linked data would write, comes with the site's SEO (TODO.md)
	return rendered(writer.page(head.Language, writer.metadataOf(head, page, head.Title), nil, writer.postMain(parts), foot))
}

// postParts are the parts of a post's main part, drawn from its document: its head, the headings its outline lists,
// its article, every note it cites, and the panes at its foot, none where the document holds no place in the blog.
type postParts struct {
	head    document.Head
	outline []outlined
	article templ.Component
	notes   []document.Note
	panes   *panesFoot
}

// outlined is a heading the outline lists: the name it is linked to by, and its title.
type outlined struct {
	id    string
	title []document.Span
}

// pick is the series' Pick (DESIGN.md, Pick) as a page draws it: its chip's words and the count after them, none in
// the panes' heading, and the window it opens, known by its id, listing the series' parts.
type pick struct {
	words   string
	count   string
	id      string
	titleID string
	series  document.Series
}

// panesFoot is the foot of a post's article (DESIGN.md, Panes): its heading, the series' pick beside it where the post
// is a part of one, or the way to the blog's list where it is not, the two sides, and where the post is discussed.
type panesFoot struct {
	heading     string
	pick        *pick
	allPosts    string
	before      paneSide
	after       paneSide
	discussions []document.Link
}

// paneSide is a side of a post's foot: its name, the post there, none on a side with nothing there, and what the side
// says then; and whether it is the side after the post.
type paneSide struct {
	name      string
	neighbour *document.Neighbour
	none      string
	after     bool
}

// pickOf is a series' pick known by an id, its chip saying words and a count after them.
func pickOf(id, words, count string, series document.Series) pick {
	return pick{words: words, count: count, id: id, titleID: pickTitleID(id), series: series}
}

// pickIDs are the names a pick known by an id gives its parts: its window, then the window's title.
func pickIDs(id string) []string {
	return []string{id, pickTitleID(id)}
}

// pickTitleID is the name of the title of the window of a pick known by an id: the id, then the title's ending.
func pickTitleID(id string) string {
	return id + pickTitleEnd
}

// headPick is the pick of a post's series in its head: the series' title, then the post's place in it.
func (writer PageWriter) headPick(series document.Series) pick {
	count := writer.frame.seriesWords.PartOf(readPart(series), len(series.Parts))
	return pickOf(headPickID, document.PlainText(series.Title), count, series)
}

// readPart is the number of the part read in a series, which compose marks once.
func readPart(series document.Series) int {
	for _, part := range series.Parts {
		if part.State == document.PartReading {
			return part.Number
		}
	}
	return 0
}

// panesOf are the panes at the foot of a post: in a series, under the heading of the series' other parts, its pick
// beside it, each side named by its part's number; outside one, under the heading of the blog's other posts, the way
// to the blog's list beside it, each side named older or newer; none where a document holds no place in the blog.
func (writer PageWriter) panesOf(place *document.PostPlace) *panesFoot {
	if place == nil {
		return nil
	}
	words, seriesWords := writer.frame.paneWords, writer.frame.seriesWords
	sides, heading, allPosts := words.Posts, words.OtherPosts, writer.frame.blogStart
	var allParts *pick
	if series := place.Series; series != nil {
		sides, heading, allPosts = words.Parts, seriesWords.OtherParts(document.PlainText(series.Title)), ""
		footPick := pickOf(footPickID, seriesWords.AllParts(len(series.Parts)), "", *series)
		allParts = &footPick
	}
	return &panesFoot{
		heading:     heading,
		pick:        allParts,
		allPosts:    allPosts,
		before:      writer.paneSideOf(sides.Before, place.Neighbours.Before, false),
		after:       writer.paneSideOf(sides.After, place.Neighbours.After, true),
		discussions: place.Discussions,
	}
}

// paneSideOf is a side of a post's foot holding a neighbour: named by the neighbour's part where it is one, by the
// side's own name otherwise.
func (writer PageWriter) paneSideOf(words compose.SideWords, neighbour *document.Neighbour, after bool) paneSide {
	name := words.Name
	if neighbour != nil && neighbour.Part > 0 {
		name = writer.frame.seriesWords.Part(neighbour.Part)
	}
	return paneSide{name: name, neighbour: neighbour, none: words.None, after: after}
}

// postFoot is the foot line of a post's page, in the blog's session: the blog's list its first window and the post
// its second, by its slug, and the post to read in a terminal and as markdown.
func (writer PageWriter) postFoot(page document.Page) footLine {
	windows := []window{
		windowOf(firstWindowNumber, writer.frame.listWords.Window, listPage(document.PageBlog), page),
		windowOf(postWindowNumber, page.Slug, page, page),
	}
	return footLineOf(writer.frame, writer.site, writer.frame.sessionsOf(true), windows, page)
}

// article draws a post's body, each part as it is met: each paragraph at the body's top or a section's numbered, from
// one; each block of code numbered, from one, its button named by its number; each note beside the first block that
// cites it; each heading and numbered paragraph named for a reader to link to, by a name the page holds nowhere else;
// and each heading of the body's top level listed for the outline, by its name. What it drew are its parts.
type article struct {
	writer     PageWriter
	paragraphs int
	codes      int
	cited      map[int]bool
	names      map[string]bool
	outline    []outlined
	parts      []templ.Component
}

// articleOf is a post's body as the article draws it, on a page whose other parts hold names already: the article
// drawn, its parts and the headings its outline lists.
func (writer PageWriter) articleOf(blocks []document.Block, held []string) *article {
	names := make(map[string]bool, len(held))
	for _, name := range held {
		names[name] = true
	}
	drawing := &article{writer: writer, paragraphs: 0, codes: 0, cited: map[int]bool{}, names: names, outline: []outlined{}, parts: nil}
	drawing.parts = drawing.partsOf(blocks, firstHeadingDepth)
	return drawing
}

// partsOf are blocks as the article lays them out, a section's at the depth of its heading: each block, then the notes
// it is the first to cite, beside it; a section's notes stand beside the blocks it holds. A heading of the body's top
// level is one the outline lists, and the site's script finds.
func (drawing *article) partsOf(blocks []document.Block, depth int) []templ.Component {
	parts := make([]templ.Component, 0, len(blocks))
	for _, block := range blocks {
		if section, isSection := block.(document.Section); isSection {
			id := drawing.name(document.SlugOf(document.PlainText(section.Title)))
			named := templ.Attributes{idAttribute: id}
			if depth == firstHeadingDepth {
				named[partAttribute] = outlineHeadingPart
				drawing.outline = append(drawing.outline, outlined{id: id, title: section.Title})
			}
			parts = append(parts, drawing.writer.articleSection(named, depth, section.Title, drawing.partsOf(section.Blocks, depth+1)))
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
	case document.Section, document.Facts, document.Entry, document.Card, document.Contents, document.Checks, document.Skills,
		document.Posts:
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
