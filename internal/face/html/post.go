package html

import (
	"strconv"
	"unicode"

	"github.com/a-h/templ"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/look"
	"github.com/vrppaul/paranin/internal/routes"
)

// How a post's day is written for a machine, as a time's datetime holds it; a reader reads it in the frame's words
const machineDayLayout = "2006-01-02"

// How a paragraph a reader links to is named: by its number after this, an id of the page's own (ownID); a heading
// is named by its anchor (document.Section's Name)
const paragraphIDOpening = "p"

// The place an address names inside a page, after its path
const fragmentMark = "#"

// What stands between two words templ writes no space between: a note's number and its words, a part's title and the
// word saying it comes, the word before where a post is discussed and the first place, and a version's day and what
// follows it
const wordSpace = " "

// The names of the parts of a post's page a link, a button or a widget names, beside its article's places, each an id
// of the page's own (ownID): the versions' pick, the series' pick in its head and in its panes' heading, each window's
// title after the pick's name, the panes' heading, and a widget's alt text after its anchor, which describes the widget
// once it runs
const (
	versionsPickID = "versions"
	headPickID     = "parts"
	footPickID     = "parts-foot"
	pickTitleEnd   = "-title"
	panesHeadingID = "more"
	widgetAltEnd   = "-alt"
)

// The parts of a post's page the site's script finds (web/src/controllers/outline.ts): the outline, each link of it,
// and each heading it leads to, the article's top level's
const (
	outlinePart        = "outline"
	outlineLinkPart    = "outline-link"
	outlineHeadingPart = "outline-heading"
)

// Post writes a post as a page of HTML (DESIGN.md, Pages): what the page says of itself, its record as linked data
// where it has one, then its head, its outline, its article, the panes at its foot and the foot line. A post has no
// tabs: it is drawn whole, in the blog's session, its window the post's slug, its command line going to its headings
// and its series' other parts too.
func (writer PageWriter) Post(composed document.Document, record compose.Record) string {
	head := composed.Head
	foot := writer.postFoot(head, document.PlacesOf(composed))
	drawn := writer.articleOf(composed.Blocks)
	parts := postParts{
		head:    head,
		outline: drawn.outline,
		article: templ.Join(drawn.parts...),
		notes:   document.NotesOf(composed.Blocks),
		panes:   writer.panesOf(head.Post),
	}
	return rendered(writer.page(head.Language, writer.metadataOf(head, head.Page, head.Title, record), writer.postMain(parts), foot))
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

// pickOf is a series' pick known by a name, its chip saying words and a count after them: its window's id the name's,
// its title's the name's with the title's ending.
func pickOf(name, words, count string, series document.Series) pick {
	return pick{words: words, count: count, id: ownID(name), titleID: ownID(name + pickTitleEnd), series: series}
}

// headPick is the pick of a post's series in its head: the series' title, then the post's place in it.
func (writer PageWriter) headPick(series document.Series) pick {
	count := writer.frame.seriesWords.PartOf(series.Reading(), len(series.Parts))
	return pickOf(headPickID, document.PlainText(series.Title), count, series)
}

// panesOf are the panes at the foot of a post: in a series, under the heading of the series' other parts, its pick
// beside it, each side named by its part's number; outside one, under the heading of the blog's other posts, the way
// to the blog's list beside it, each side named older or newer; none where a document holds no place in the blog.
func (writer PageWriter) panesOf(place *document.PostPlace) *panesFoot {
	if place == nil {
		return nil
	}
	words, seriesWords := writer.frame.paneWords, writer.frame.seriesWords
	sides, heading, allPosts := words.Posts, words.OtherPosts, routes.Reference(writer.frame.blogList, routes.FormatChosen)
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

// paneSideOf is a side of a post's foot holding a neighbour, named as compose.SeriesWords.PaneName names it.
func (writer PageWriter) paneSideOf(words compose.SideWords, neighbour *document.Neighbour, after bool) paneSide {
	return paneSide{name: writer.frame.seriesWords.PaneName(words, neighbour), neighbour: neighbour, none: words.None, after: after}
}

// postFoot is the foot line of a post's page, in the blog's session: the blog's list by year its first window, since a
// page keeps no grouping a reader chose, and the post its second, and the post to read in a terminal and as markdown;
// its command line going to the post's own places too.
func (writer PageWriter) postFoot(head document.Head, own []document.Place) footLine {
	around, page := writer.frame, head.Page
	windows := around.blogWindowsOf(around.listWords.BlogWindows(around.blogList, page), page)
	return footLineOf(around, around.sshWays.SSHWayTo(head), around.sessionsOf(page), windows, page, own)
}

// article draws a post's body, each part as it is met: each paragraph at the body's top or a section's numbered, from
// one; each block of code numbered, from one, its button named by its number; each note beside the first block that
// cites it; each heading named for a reader to link to by its anchor, and each numbered paragraph by its number; and
// each heading of the body's top level listed for the outline, by its anchor. What it drew are its parts.
type article struct {
	writer     PageWriter
	paragraphs int
	codes      int
	cited      map[int]bool
	outline    []outlined
	parts      []templ.Component
}

// articleOf is a post's body as the article draws it: the article drawn, its parts and the headings its outline lists.
func (writer PageWriter) articleOf(blocks []document.Block) *article {
	drawing := &article{writer: writer, paragraphs: 0, codes: 0, cited: map[int]bool{}, outline: []outlined{}, parts: nil}
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
			named := templ.Attributes{idAttribute: section.Name}
			if depth == firstHeadingDepth {
				named[partAttribute] = outlineHeadingPart
				drawing.outline = append(drawing.outline, outlined{id: section.Name, title: section.Title})
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
	return drawing.writer.numberedParagraph(number, ownID(paragraphIDOpening+number), paragraph.Spans)
}

// block is a block of a post as its article draws it, inside a list's item or a quote as at the body's top: a list and
// a quote holding blocks drawn so too, code numbered, with its language and the button that copies it, a picture
// under its caption, a widget in its window.
func (drawing *article) block(block document.Block) templ.Component {
	switch typed := block.(type) {
	case document.Paragraph:
		return drawing.writer.paragraph(typed)
	case document.Table:
		return drawing.writer.table(typed)
	case document.CodeBlock:
		drawing.codes++
		return copyableCode(typed, drawing.writer.frame.copyWords.CodeButton(drawing.codes))
	case document.List:
		items := make([]listedItem, 0, len(typed.Items))
		for index, item := range typed.Items {
			items = append(items, listedItem{
				mark: itemMark(typed, index), opensBlock: opensWithBlock(item), content: listItem(item, drawing.writer.spans, drawing.blocks),
			})
		}
		return postList(typed.Ordered, typed.Start, items)
	case document.Quote:
		return postQuote(drawing.blocks(typed.Blocks))
	case document.Pause:
		return pause()
	case document.Figure:
		return drawing.writer.figure(typed)
	case document.Widget:
		return drawing.writer.widget(typed)
	case document.Section, document.Facts, document.ExperienceHead, document.Entry, document.Points, document.Card, document.Contents,
		document.Checks, document.Skills, document.Posts:
		// a post's body nests its headings into sections, which parts draws, and holds none of the résumé's blocks nor
		// a project's points; the load refuses a heading inside a list's item or a quote
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

// listedItem is an item of a post's list as the page draws it: its mark, a dash or its number, which hangs in the
// margin; whether it opens with a block, which then stands on the mark's line; and what it holds.
type listedItem struct {
	mark       string
	opensBlock bool
	content    templ.Component
}

// itemMark is the mark an item of a list opens with, by its place: a dash, or in a list in order its number.
func itemMark(list document.List, index int) string {
	if !list.Ordered {
		return look.PointMark
	}
	return list.Number(index)
}

// opensWithBlock is whether an item opens with a block that is no paragraph: a picture, code, a quote, a table or a
// list.
func opensWithBlock(item document.Item) bool {
	if len(item.Blocks) == 0 {
		return false
	}
	_, isParagraph := item.Blocks[0].(document.Paragraph)
	return !isParagraph
}

// nameWords are the words of a widget's name, as a component is named, which the page lets a name too long for its line
// break between: each opens at a capital after a small letter or a digit, Page2|Cache, or at the last capital of a run
// before a small letter, HTTP|Retry.
func nameWords(name string) []string {
	letters := []rune(name)
	words, start := []string{}, 0
	for index := 1; index < len(letters); index++ {
		if opensWord(letters, index) {
			words = append(words, string(letters[start:index]))
			start = index
		}
	}
	return append(words, string(letters[start:]))
}

// opensWord is whether the letter at a place of a name, past its first, opens a word of it, as nameWords reads them.
func opensWord(letters []rune, index int) bool {
	if !unicode.IsUpper(letters[index]) {
		return false
	}
	afterWord := !unicode.IsUpper(letters[index-1])
	beforeSmall := index+1 < len(letters) && unicode.IsLower(letters[index+1])
	return afterWord || beforeSmall
}

// listItem is what an item of a post's list holds, its spans and its blocks drawn as the page or the feed draws them:
// its first block's words on the line of its mark, where that is a paragraph, and the blocks after them under them.
func listItem(item document.Item, spans func([]document.Span) templ.Component, blocks func([]document.Block) []templ.Component) templ.Component {
	if len(item.Blocks) == 0 {
		return templ.NopComponent
	}
	if paragraph, isParagraph := item.Blocks[0].(document.Paragraph); isParagraph {
		return templ.Join(append([]templ.Component{spans(paragraph.Spans)}, blocks(item.Blocks[1:])...)...)
	}
	return templ.Join(blocks(item.Blocks)...)
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
