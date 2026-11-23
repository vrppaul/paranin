package tui

import (
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/look"
	"github.com/vrppaul/paranin/internal/routes"
)

// A post's columns (DESIGN.md, The TUI), beside its Outline's and its sources', look.Rail and look.Aside wide: the gap
// after the Outline; its paragraphs' numbers and the gap after them; its text at the measure a paragraph is read at;
// the gap before the sources beside the text; and the width its title wraps at, a few words a line, as the page holds
// it
const (
	outlineGap   = 4
	numberWidth  = 4
	numberGap    = 2
	textMeasure  = 66
	sourcesGap   = 4
	titleMeasure = 44
)

// How a post's columns add up: the column its text starts at, after the numbers; the numbers and the text together;
// the width the sources stand beside the text from; and the width the Outline stands beside them from
const (
	textColumn   = numberWidth + numberGap
	articleWidth = textColumn + textMeasure
	sourcesFrom  = articleWidth + sourcesGap + look.Aside
	outlineFrom  = look.Rail + outlineGap + sourcesFrom
)

// The marks of a post's body: what follows an ordered list's number, the bar beside a quote and the room it takes
// before the quote's words, the rows of air a pause keeps above and below it on top of the row its block keeps, and
// what opens each row of a line of code that goes on under itself, as Neovim's showbreak marks it
const (
	orderedMark    = "."
	quoteBar       = "▎"
	quoteRoom      = 2
	pauseAir       = 1
	codeGoesOnMark = "↪ "
)

// How a post's parts stand apart: the rows between its head and its body and between its body and its panes; the
// room inside a frame of code at each side, and above and below what it holds; how far into a frame's top edge its
// name stands; and the gap between the panes' heading and the way to all of them, and between the word before where
// the post is discussed and the places
const (
	partRows       = 2
	framePadding   = 2
	frameRowsRound = 1
	frameNameInset = 2
	headingGap     = 4
	discussGap     = 2
)

// postShape is how a post stands, by the columns its page has room for (DESIGN.md, The TUI).
type postShape int

const (
	// narrowPost is its numbers and its text alone, its sources at the text's end, its panes one under the other
	narrowPost postShape = iota
	// sourcedPost has its sources beside its text and its head's facts beside its title, its panes side by side
	sourcedPost
	// outlinedPost has its Outline at its left as well
	outlinedPost
)

// postShapeOf is how a post stands on a page of a width: with its Outline and its sources beside its text each from
// the width it needs.
func postShapeOf(width int) postShape {
	switch {
	case width >= outlineFrom:
		return outlinedPost
	case width >= sourcesFrom:
		return sourcedPost
	}
	return narrowPost
}

// width is how wide a post of the shape stands on a page of a width: its numbers and its text alone no wider than the
// page.
func (shape postShape) width(page int) int {
	switch shape {
	case sourcedPost:
		return sourcesFrom
	case outlinedPost:
		return outlineFrom
	case narrowPost:
	}
	return min(page, articleWidth)
}

// articleStart is the column the post's numbers stand from: after the Outline and its gap where it stands.
func (shape postShape) articleStart() int {
	switch shape {
	case outlinedPost:
		return look.Rail + outlineGap
	case narrowPost, sourcedPost:
	}
	return 0
}

// sourcesBeside is whether the post's sources stand beside its text.
func (shape postShape) sourcesBeside() bool {
	switch shape {
	case sourcedPost, outlinedPost:
		return true
	case narrowPost:
	}
	return false
}

// sourcesStart is the column the sources beside the post's text stand from.
func (shape postShape) sourcesStart() int {
	return shape.articleStart() + articleWidth + sourcesGap
}

// post lays out a post (DESIGN.md, The TUI) in the middle of the page, in the columns it has room for: its head; its
// body numbered in the margin, its sources beside the lines that cite them or at its end under a rule; the panes at
// its foot; with the line each heading stands at, and the Outline, which stands in sight beside its body as it
// scrolls.
func (layout layout) post(post document.Document) laidTab {
	shape := postShapeOf(layout.width)
	width := shape.width(layout.width)
	headed, numbered := layout, layout
	headed.width, numbered.width = width, min(width, articleWidth)
	head := slices.Concat(headed.postHead(post.Head, shape), make([]line, partRows))
	body := article{layout: numbered, paragraphs: 0, headings: map[string]int{}}
	lines := slices.Concat(head, shiftedLines(body.body(post.Blocks, len(head)), shape.articleStart()))
	notes := document.NotesOf(post.Blocks)
	if shape.sourcesBeside() {
		sourced := layout
		sourced.width = look.Aside
		lines = besideLines(lines, shape.sourcesStart(), sourced.sourcesBeside(lines, len(head), notes))
	} else {
		lines = stacked(lines, numbered.sources(notes))
	}
	bodyEnd := len(lines)
	if place := post.Head.Post; place != nil {
		footed := layout
		footed.width = width - shape.articleStart()
		lines = slices.Concat(lines, make([]line, partRows), shiftedLines(footed.panes(*place, shape.sourcesBeside()), shape.articleStart()))
	}
	shift := (layout.width - width) / sides
	laid := laidTab{lines: shiftedLines(lines, shift), tiles: nil, groups: nil, split: 0, sections: body.headings, outline: noOutline()}
	if shape == outlinedPost {
		outlined := layout
		outlined.width = outlineHang + look.Rail
		laid.split, laid.outline = shift+shape.articleStart(), outlined.outline(post.Blocks, body.headings, len(head), bodyEnd)
	}
	return laid
}

// shiftedLines are lines standing a number of columns further in.
func shiftedLines(lines []line, by int) []line {
	moved := make([]line, 0, len(lines))
	for _, shown := range lines {
		moved = append(moved, append(line{gapRun(by)}, shown...))
	}
	return moved
}

// postHead is a post's head (DESIGN.md, PostHead): its title bold, a few words a line, its lead under it, and its
// facts, the day it came out, its series and where it stands in it, and its tags: beside the title in the sources'
// column where the post has one, under the lead where not.
func (layout layout) postHead(head document.Head, shape postShape) []line {
	titled, led := layout, layout
	titled.width, led.width = min(layout.width, titleMeasure), min(layout.width, look.PostLead)
	intro := stacked(titled.fill(layout.words(head.Title, kindStrong), 0), led.fill(layout.words(head.Lead, kindText), 0))
	if !shape.sourcesBeside() {
		return stacked(intro, layout.postFacts(head))
	}
	beside := layout
	beside.width = look.Aside
	return besideLines(intro, shape.sourcesStart(), beside.postFacts(head))
}

// postFact is a fact of a post's head: its label, and its value's words.
type postFact struct {
	label string
	value []word
}

// postFacts lays out a post's facts, each value in a column after the widest label, those the post has.
func (layout layout) postFacts(head document.Head) []line {
	labels := layout.siteWords.PostLabels
	facts := []postFact{}
	switch version := head.Version.(type) {
	case document.VersionOut:
		facts = append(facts, postFact{label: labels.Version, value: textWords(version.Published.Format(look.DayLayout), kindText)})
	case document.VersionDraft:
		facts = append(facts, postFact{label: labels.Version, value: textWords(version.Word, kindText)})
	}
	if place := head.Post; place != nil && place.Series != nil {
		series := place.Series
		partOf := layout.siteWords.SeriesWords.PartOf(series.Reading(), len(series.Parts))
		facts = append(facts, postFact{label: labels.Series, value: append(layout.words(series.Title, kindText), textWords(partOf, kindMuted)...)})
	}
	if len(head.Filters) > 0 {
		tags := []word{}
		for _, filter := range head.Filters {
			tags = append(tags, filterWords(filter)...)
		}
		facts = append(facts, postFact{label: labels.Tags, value: tags})
	}
	column := 0
	for _, fact := range facts {
		column = max(column, textWidth(fact.label)+columnGap)
	}
	lines := []line{}
	for _, fact := range facts {
		lines = append(lines, layout.continued(line{gapRun(0), plainRun(padded(fact.label, column), kindMuted)}, fact.value, column)...)
	}
	return lines
}

// article lays out a post's body: its layout, how many of its paragraphs it has numbered so far, and the line of the
// post each heading stands at, by its anchor.
type article struct {
	layout     layout
	paragraphs int
	headings   map[string]int
}

// body lays out the blocks of a post's body or of a section in the text's column, starting at a line of the post, a
// row of nothing between each two: each section under its heading, bold, whose line it keeps by the heading's anchor,
// and the paragraphs numbered in the margin.
func (writing *article) body(blocks []document.Block, from int) []line {
	lines := []line{}
	for _, block := range blocks {
		section, isSection := block.(document.Section)
		if !isSection {
			lines = stacked(lines, writing.block(block, textColumn, true))
			continue
		}
		start := stackedFrom(from, lines)
		writing.headings[section.Name] = start
		heading := writing.layout.fill(writing.layout.words(section.Title, kindStrong), textColumn)
		lines = stacked(lines, stacked(heading, writing.body(section.Blocks, stackedFrom(start, heading))))
	}
	return lines
}

// stackedFrom is the line of the post that lines stacked under others, which start at a line, start at.
func stackedFrom(from int, above []line) int {
	if len(above) == 0 {
		return from
	}
	return from + len(above) + blockGap
}

// blocks lays out the blocks a quote holds from a column, a row of nothing between each two, no paragraph numbered.
func (writing *article) blocks(blocks []document.Block, column int) []line {
	lines := []line{}
	for _, block := range blocks {
		lines = stacked(lines, writing.block(block, column, false))
	}
	return lines
}

// block lays out a block of a post that is no section from a column (DESIGN.md, Article): a paragraph numbered where
// it is numbered, a table, code in a frame, a list, a quote, a pause and a picture.
func (writing *article) block(block document.Block, column int, numbered bool) []line {
	layout := writing.layout
	switch typed := block.(type) {
	case document.Paragraph:
		words := layout.words(typed.Spans, kindText)
		if !numbered {
			return layout.fill(words, column)
		}
		writing.paragraphs++
		number := strconv.Itoa(writing.paragraphs)
		lead := line{gapRun(column - numberGap - textWidth(number)), plainRun(number, kindFaint), gapRun(numberGap)}
		return layout.continued(lead, words, column)
	case document.Table:
		return layout.table(document.Table{Header: typed.Header, Rows: slices.Concat(headerRows(typed.Header), typed.Rows)}, column)
	case document.CodeBlock:
		return layout.codeFrame(typed, column)
	case document.List:
		return writing.list(typed, column)
	case document.Quote:
		return writing.quote(typed, column)
	case document.Pause:
		return layout.pause(column)
	case document.Figure:
		return layout.figure(typed, column)
	case document.Section, document.Facts, document.Entry, document.Card, document.Contents, document.Checks, document.Skills,
		document.Posts:
		// body lays out the body's sections, and the load refuses a heading inside a list's item or a quote; a post's
		// body holds none of the résumé's blocks nor the blog's list
	}
	return nil
}

// headerRows are a table's header as rows over its own, none where it has none.
func headerRows(header document.Row) []document.Row {
	if len(header.Cells) == 0 {
		return nil
	}
	return []document.Row{header}
}

// list lays out a list from a column, inside the text's, so its marks never read as the paragraphs' numbers: each
// item's mark, a faint dash or its number muted, the numbers standing to their right edge; its words after the widest
// mark, standing in under themselves.
func (writing *article) list(list document.List, column int) []line {
	marks := make([]string, 0, len(list.Items))
	for index := range list.Items {
		mark := look.PointMark
		if list.Ordered {
			mark = strconv.Itoa(list.Start+index) + orderedMark
		}
		marks = append(marks, mark)
	}
	markKind, markWidth := kindFaint, 0
	if list.Ordered {
		markKind = kindMuted
	}
	for _, mark := range marks {
		markWidth = max(markWidth, textWidth(mark))
	}
	lines := []line{}
	for index, item := range list.Items {
		lead := line{gapRun(column + markWidth - textWidth(marks[index])), plainRun(marks[index], markKind), gapRun(spaceWidth)}
		lines = append(lines, writing.item(item, lead, column+markWidth+spaceWidth)...)
	}
	return lines
}

// item lays out an item of a list after its mark's lead, its words standing from a column: its first paragraph's words
// on the mark's line, the blocks after them under them, a list among them from the same column. The mark stands alone
// on its line where the item opens with no paragraph.
func (writing *article) item(item document.Item, lead line, column int) []line {
	words, blocks := []word{}, item.Blocks
	if paragraph, isParagraph := firstParagraph(blocks); isParagraph {
		words, blocks = writing.layout.words(paragraph.Spans, kindText), blocks[1:]
	}
	lines := writing.layout.continued(lead, words, column)
	for _, block := range blocks {
		lines = append(lines, writing.block(block, column, false)...)
	}
	return lines
}

// firstParagraph is the paragraph blocks open with, where they open with one.
func firstParagraph(blocks []document.Block) (document.Paragraph, bool) {
	if len(blocks) == 0 {
		return document.Paragraph{Spans: nil}, false
	}
	paragraph, isParagraph := blocks[0].(document.Paragraph)
	return paragraph, isParagraph
}

// quote lays out a quote from a column: its blocks, none numbered, each line after a bar as the page draws it.
func (writing *article) quote(quote document.Quote, column int) []line {
	inside := article{layout: writing.layout, paragraphs: 0, headings: nil}
	inside.layout.width = writing.layout.width - column - quoteRoom
	bar := line{gapRun(column), plainRun(quoteBar, kindQuoteBar), gapRun(quoteRoom - textWidth(quoteBar))}
	lines := []line{}
	for _, shown := range inside.blocks(quote.Blocks, 0) {
		lines = append(lines, slices.Concat(bar, shown))
	}
	return lines
}

// pause lays out a pause (DESIGN.md, Article): three stars faint in the middle of the text's column, a row of air more
// above and below.
func (layout layout) pause(column int) []line {
	stars := line{gapRun(column + (layout.width-column-textWidth(look.PauseMark))/sides), plainRun(look.PauseMark, kindFaint)}
	return slices.Concat(make([]line, pauseAir), []line{stars}, make([]line, pauseAir))
}

// figure lays out a picture from a column as a terminal, which draws none, shows it (docs/FOUNDATION.md §5.4), each
// part from a line of its own: its alt text, then its address, faint, each a link to the picture, so an address broken
// across lines leads there from each; its caption muted, as the page sets it.
func (layout layout) figure(figure document.Figure, column int) []line {
	address := layout.site.Address(figure.Picture, routes.FormatChosen)
	lines := layout.fill(layout.words(pictureLink(figure, figure.Alt), kindText), column)
	lines = append(lines, layout.fill(layout.words(pictureLink(figure, address), kindFaint), column)...)
	return append(lines, layout.fill(textWords(figure.Caption, kindMuted), column)...)
}

// pictureLink is words as a link to a figure's picture.
func pictureLink(figure document.Figure, words string) []document.Span {
	return []document.Span{document.Link{Target: figure.Picture, Spans: []document.Span{document.Text{Text: words}}}}
}

// codeFrame lays out a block of code from a column in a quiet frame across the text's column (DESIGN.md, The TUI): its
// language set into the top edge, a row of room above and below the code, each kind of word in its colour, a tab
// standing to the next of look.TabStop's stops, and a line wider than the frame going on under itself after the mark
// that says so, so no letter is lost; each line carries the whole block for y to copy once a search finds it.
func (layout layout) codeFrame(block document.CodeBlock, column int) []line {
	across := layout.width - column - sides*textWidth(frameDown)
	label := line{}
	if block.Language != "" {
		label = line{plainRun(space+block.Language+space, kindFaint)}
	}
	top := slices.Concat(line{gapRun(column), plainRun(frameTopLeft+strings.Repeat(frameAcross, frameNameInset), kindQuietLine)}, label,
		line{plainRun(strings.Repeat(frameAcross, max(across-frameNameInset-runsWidth(label), 0))+frameTopRight, kindQuietLine)})
	lines := []line{top}
	for _, code := range slices.Concat(make([]line, frameRowsRound), codeRows(block, across-sides*framePadding), make([]line, frameRowsRound)) {
		row := append(line{gapRun(framePadding)}, code...)
		lines = append(lines, slices.Concat(line{gapRun(column), plainRun(frameDown, kindQuietLine)}, row, line{gapRun(across - runsWidth(row)), plainRun(frameDown, kindQuietLine)}))
	}
	bottom := plainRun(frameBottomLeft+strings.Repeat(frameAcross, max(across, 0))+frameBottomRight, kindQuietLine)
	return append(lines, line{gapRun(column), bottom})
}

// codeRows are a block of code's rows within a width: its runs, each in the kind its kind of word reads as, parted at
// each line's end, a tab written as the spaces to the next stop of its line, and a line going on under itself where it
// is wider, each row it goes on in opening with the mark that says so, faint.
func codeRows(block document.CodeBlock, width int) []line {
	copied := block.Text()
	rows := codeRowing{width: max(width, textWidth(codeGoesOnMark)+spaceWidth), rows: []line{}, current: line{}, used: 0, column: 0}
	for _, written := range block.Runs {
		shown := run{text: "", kind: codeKindOf(written.Kind), highlighted: false, address: "", copied: copied, tone: 0, cites: noCitation}
		for _, letter := range written.Text {
			rows.add(letter, shown)
		}
	}
	return append(rows.rows, rows.current)
}

// codeRowing is a block of code being laid in rows of a width: the rows filled, the row being laid and how many
// columns it takes, and how far into its line of code the row being laid has come, which a tab's stops count from.
type codeRowing struct {
	width   int
	rows    []line
	current line
	used    int
	column  int
}

// add lays a letter of code of a run's form on the row being laid: a line's end opens the next line, a tab the spaces
// to the next stop, and a letter past the width opens a row it goes on in.
func (rowing *codeRowing) add(letter rune, form run) {
	text := string(letter)
	switch text {
	case lineEnd:
		rowing.open(line{}, 0)
		rowing.column = 0
		return
	case tabCharacter:
		text = strings.Repeat(space, look.TabStop-rowing.column%look.TabStop)
	}
	width := ansi.StringWidth(text)
	if rowing.used > 0 && rowing.used+width > rowing.width {
		rowing.open(line{plainRun(codeGoesOnMark, kindFaint)}, textWidth(codeGoesOnMark))
	}
	if last := len(rowing.current) - 1; last >= 0 && sameForm(rowing.current[last], form) {
		rowing.current[last].text += text
	} else {
		form.text = text
		rowing.current = append(rowing.current, form)
	}
	rowing.used += width
	rowing.column += width
}

// open ends the row being laid and opens the next, holding a lead of a width.
func (rowing *codeRowing) open(lead line, leadWidth int) {
	rowing.rows = append(rowing.rows, rowing.current)
	rowing.current, rowing.used = lead, leadWidth
}

// The tab a line of code may hold, which a face writes as the spaces to its next stop
const tabCharacter = "\t"

// codeKindOf is the kind a kind of word in code reads as (DESIGN.md, Article): a comment muted, punctuation faint,
// each other kind of colour its own, and every other word text. exhaustive holds the switch to every kind.
func codeKindOf(codeKind document.CodeKind) kind {
	switch codeKind {
	case document.CodeKeyword:
		return kindCodeKeyword
	case document.CodeString:
		return kindCodeString
	case document.CodeNumber:
		return kindCodeNumber
	case document.CodeFunction:
		return kindCodeFunction
	case document.CodeType:
		return kindCodeType
	case document.CodeBuiltin:
		return kindCodeBuiltin
	case document.CodeOperator:
		return kindCodeOperator
	case document.CodeVariable:
		return kindCodeVariable
	case document.CodeComment:
		return kindMuted
	case document.CodePunctuation:
		return kindFaint
	case document.CodePlain:
	}
	return kindText
}

// sources lays out a post's sources at the text's end (DESIGN.md, Article), in the text's column, each note once under
// a quiet rule; nothing where the post cites none.
func (layout layout) sources(notes []document.Note) []line {
	if len(notes) == 0 {
		return nil
	}
	noted := layout
	noted.width = layout.width - textColumn
	lines := []line{{gapRun(textColumn), plainRun(strings.Repeat(frameAcross, max(noted.width, 0)), kindQuietLine)}}
	for _, note := range notes {
		lines = append(lines, shiftedLines(noted.noteLines(note), textColumn)...)
	}
	return lines
}

// sourcesBeside are the lines a post's sources stand in beside its lines, from a line on (DESIGN.md, Article): each
// note at the line that first cites it, or a row under the note before where that one runs on past it; none beside
// any other line.
func (layout layout) sourcesBeside(lines []line, from int, notes []document.Note) []line {
	cited := citedLines(lines)
	beside, free := []line{}, from
	for _, note := range notes {
		at := max(cited[note.Number], free)
		beside = slices.Concat(beside, make([]line, at-len(beside)), layout.noteLines(note))
		free = len(beside) + blockGap
	}
	return beside
}

// citedLines are the line among lines that first cites each note, by the note's number.
func citedLines(lines []line) map[int]int {
	cited := map[int]int{}
	for index, shown := range lines {
		for _, piece := range shown {
			if _, seen := cited[piece.cites]; !seen && piece.cites != noCitation {
				cited[piece.cites] = index
			}
		}
	}
	return cited
}

// noteLines are a note as a source stands (DESIGN.md, Article): its number, then its words muted, standing in under
// themselves.
func (layout layout) noteLines(note document.Note) []line {
	lead := line{gapRun(0), plainRun(strconv.Itoa(note.Number), kindLabel), gapRun(spaceWidth)}
	return layout.continued(lead, layout.words(note.Spans, kindMuted), runsWidth(lead))
}

// paneSide is a side of a post's foot: the name set into its pane's top edge, the post there, none on a side with
// nothing there, and what the side says then.
type paneSide struct {
	name      string
	neighbour *document.Neighbour
	none      string
}

// panes lays out a post's foot (DESIGN.md, Panes) from the text's column: a heading saying where its panes lead and
// the way to all of them, then the two panes, side by side or one under the other; then where the post is discussed.
func (layout layout) panes(place document.PostPlace, sideBySide bool) []line {
	paneWords, seriesWords := layout.siteWords.PaneWords, layout.siteWords.SeriesWords
	sides, heading, all := paneWords.Posts, paneWords.OtherPosts, layout.siteWords.ListWords.AllPosts
	if series := place.Series; series != nil {
		sides, heading, all = paneWords.Parts, seriesWords.OtherParts(document.PlainText(series.Title)), seriesWords.AllParts(len(series.Parts))
	}
	after := paneSide{name: seriesWords.PaneName(sides.After, place.Neighbours.After), neighbour: place.Neighbours.After, none: sides.After.None}
	before := paneSide{name: seriesWords.PaneName(sides.Before, place.Neighbours.Before), neighbour: place.Neighbours.Before, none: sides.Before.None}
	title := line{gapRun(textColumn), plainRun(heading, kindStrong), gapRun(headingGap), plainRun(all, kindMuted)}
	lines := slices.Concat([]line{title}, blankLines(), shiftedLines(layout.paneGrid(after, before, sideBySide), textColumn))
	if len(place.Discussions) == 0 {
		return lines
	}
	discussed := []word{}
	for _, link := range place.Discussions {
		discussed = markedWords(discussed, layout.words([]document.Span{link}, kindText), placeMark)
	}
	lead := line{gapRun(textColumn), plainRun(paneWords.Discuss, kindMuted), gapRun(discussGap)}
	return stacked(lines, layout.continued(lead, discussed, runsWidth(lead)))
}

// paneGrid are a post's two panes as tmux splits a window, across the width the layout leaves after the text's column,
// sharing the line between them as the blog's tiles do, filling the width: side by side, the side before at the left,
// as the page sets them; else one under the other, the side after first.
func (layout layout) paneGrid(after, before paneSide, sideBySide bool) []line {
	rows := [][]paneSide{{after}, {before}}
	if sideBySide {
		rows = [][]paneSide{{before, after}}
	}
	grid := filledGrid(len(rows[0]), layout.width-textColumn)
	inside := layout
	inside.width = grid.inside()
	cells := make([][]gridCell, 0, len(rows))
	for _, row := range rows {
		cells = append(cells, inside.paneCells(row))
	}
	return grid.framed(cells).lines
}

// paneCells are sides of a post's foot as a row of its grid's cells: each its name set into its top edge, bold where
// the pane leads somewhere, faint where not, and what it holds.
func (layout layout) paneCells(row []paneSide) []gridCell {
	cells := make([]gridCell, 0, len(row))
	for _, side := range row {
		nameKind := kindStrong
		if side.neighbour == nil {
			nameKind = kindFaint
		}
		cells = append(cells, gridCell{label: line{plainRun(side.name, nameKind)}, words: tileWords{top: layout.paneWordsOf(side), foot: nil}, leads: side.neighbour != nil})
	}
	return cells
}

// paneWordsOf lays out what a pane holds (DESIGN.md, Panes): the post there by its title, bold and leading to it, its
// lead muted and, outside a series, its day faint; a part to come by its title muted and the word saying it comes; a
// side with nothing there by what it says then, faint.
func (layout layout) paneWordsOf(side paneSide) []line {
	neighbour := side.neighbour
	switch {
	case neighbour == nil:
		return layout.fill(textWords(side.none, kindFaint), 0)
	case neighbour.Page.Kind == document.PageNone:
		soon := textWords(layout.siteWords.SeriesWords.State(document.PartToCome), kindFaint)
		return layout.fill(append(layout.words(neighbour.Title, kindMuted), soon...), 0)
	}
	title := layout.words([]document.Span{document.Link{Target: neighbour.Page, Spans: neighbour.Title}}, kindStrong)
	lines := slices.Concat(layout.fill(title, 0), layout.fill(layout.words(neighbour.Lead, kindMuted), 0))
	if !neighbour.Published.IsZero() {
		lines = append(lines, layout.fill(textWords(neighbour.Published.Format(look.DayLayout), kindFaint), 0)...)
	}
	return lines
}
