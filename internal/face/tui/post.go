package tui

import (
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/look"
)

// A post's columns (DESIGN.md, The TUI): its paragraphs' numbers, the gap after them, and its text at the measure a
// paragraph is read at; the column the numbers and the text take together; and the width its title wraps at, a few
// words a line, as the page holds it
const (
	numberWidth  = 4
	numberGap    = 2
	textMeasure  = 66
	textColumn   = numberWidth + numberGap
	postWidth    = textColumn + textMeasure
	titleMeasure = 44
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
// room inside a frame of code or a pane at each side, and above and below what it holds; how far into a frame's top
// edge its name stands; and the gap between the panes' heading and the way to all of them, and between the word
// before where the post is discussed and the places
const (
	partRows       = 2
	framePadding   = 2
	frameRowsRound = 1
	frameNameInset = 2
	headingGap     = 4
	discussGap     = 2
)

// post lays out a post (DESIGN.md, The TUI) in a column of its numbers and its text, in the middle of the page: its
// head, its body numbered in the margin, its sources at its end under a rule, and the panes at its foot one under the
// other, with the line each heading stands at. The outline, the sources beside the text and the panes side by side
// wait for the page wide enough (TODO.md).
func (layout layout) post(post document.Document) laidTab {
	column := layout
	column.width = min(layout.width, postWidth)
	head := slices.Concat(column.postHead(post.Head), make([]line, partRows))
	body := article{layout: column, paragraphs: 0, headings: map[string]int{}}
	lines := slices.Concat(head, body.body(post.Blocks, len(head)))
	lines = stacked(lines, column.sources(document.NotesOf(post.Blocks)))
	if place := post.Head.Post; place != nil {
		lines = slices.Concat(lines, make([]line, partRows), column.panes(*place))
	}
	return laidTab{lines: shiftedLines(lines, (layout.width-column.width)/sides), tiles: nil, groups: nil, split: 0, sections: body.headings}
}

// shiftedLines are lines standing a number of columns further in.
func shiftedLines(lines []line, by int) []line {
	moved := make([]line, 0, len(lines))
	for _, shown := range lines {
		moved = append(moved, append(line{gapRun(by)}, shown...))
	}
	return moved
}

// postHead is a post's head (DESIGN.md, PostHead): its title bold, its lead, and its facts under the lead: the day it
// came out, its series and where it stands in it, and its tags.
func (layout layout) postHead(head document.Head) []line {
	titled := layout
	titled.width = min(layout.width, titleMeasure)
	lines := stacked(titled.fill(layout.words(head.Title, kindStrong), 0), layout.fill(layout.words(head.Lead, kindText), 0))
	return stacked(lines, layout.postFacts(head))
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
	if !head.Published.IsZero() {
		facts = append(facts, postFact{label: labels.Version, value: textWords(head.Published.Format(look.DayLayout), kindText)})
	}
	if place := head.Post; place != nil && place.Series != nil {
		series := place.Series
		partOf := layout.siteWords.SeriesWords.PartOf(series.Reading(), len(series.Parts))
		facts = append(facts, postFact{label: labels.Series, value: append(layout.words(series.Title, kindText), textWords(partOf, kindMuted)...)})
	}
	if len(head.Tags) > 0 {
		facts = append(facts, postFact{label: labels.Tags, value: tagWords(head.Tags, kindMuted)})
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
// it is numbered, a table, code in a frame, a list, a quote and a pause.
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
		shown := run{text: "", kind: codeKindOf(written.Kind), highlighted: false, address: "", copied: copied, tone: 0}
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

// sources lays out a post's sources at the text's end (DESIGN.md, Article), each note once under a quiet rule: its
// number, then its words muted, standing in under themselves; nothing where the post cites none.
func (layout layout) sources(notes []document.Note) []line {
	if len(notes) == 0 {
		return nil
	}
	rule := line{gapRun(textColumn), plainRun(strings.Repeat(frameAcross, max(layout.width-textColumn, 0)), kindQuietLine)}
	lines := []line{rule}
	for _, note := range notes {
		number := strconv.Itoa(note.Number)
		lead := line{gapRun(textColumn), plainRun(number, kindLabel), gapRun(spaceWidth)}
		lines = append(lines, layout.continued(lead, layout.words(note.Spans, kindMuted), runsWidth(lead))...)
	}
	return lines
}

// paneSide is a side of a post's foot: the name set into its pane's top edge, the post there, none on a side with
// nothing there, and what the side says then.
type paneSide struct {
	name      string
	neighbour *document.Neighbour
	none      string
}

// panes lays out a post's foot (DESIGN.md, Panes): a heading saying where its panes lead and the way to all of them,
// then the two panes one under the other, the side after first, as tmux splits a window; then where the post is
// discussed.
func (layout layout) panes(place document.PostPlace) []line {
	paneWords, seriesWords := layout.siteWords.PaneWords, layout.siteWords.SeriesWords
	sides, heading, all := paneWords.Posts, paneWords.OtherPosts, layout.siteWords.ListWords.AllPosts
	if series := place.Series; series != nil {
		sides, heading, all = paneWords.Parts, seriesWords.OtherParts(document.PlainText(series.Title)), seriesWords.AllParts(len(series.Parts))
	}
	after := paneSide{name: seriesWords.PaneName(sides.After, place.Neighbours.After), neighbour: place.Neighbours.After, none: sides.After.None}
	before := paneSide{name: seriesWords.PaneName(sides.Before, place.Neighbours.Before), neighbour: place.Neighbours.Before, none: sides.Before.None}
	title := line{gapRun(textColumn), plainRun(heading, kindStrong), gapRun(headingGap), plainRun(all, kindMuted)}
	lines := slices.Concat([]line{title}, blankLines(), layout.stackedPanes(after, before))
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

// stackedPanes are two panes of a post's foot, one under the other across the text's column, sharing the line between
// them, each piece of the frame read as the panes it bounds have it (frameKind).
func (layout layout) stackedPanes(first, second paneSide) []line {
	across := layout.width - textColumn - sides*textWidth(frameDown)
	inside := layout
	inside.width = across - sides*framePadding
	firstLeads, secondLeads := first.neighbour != nil, second.neighbour != nil
	lines := []line{paneEdge(frameTopLeft, frameTopRight, first, frameKind(firstLeads), across)}
	lines = append(lines, paneRows(inside.paneWordsOf(first), frameKind(firstLeads), across)...)
	lines = append(lines, paneEdge(frameRuleLeft, frameRuleRight, second, frameKind(firstLeads, secondLeads), across))
	lines = append(lines, paneRows(inside.paneWordsOf(second), frameKind(secondLeads), across)...)
	bottom := plainRun(frameBottomLeft+strings.Repeat(frameAcross, max(across, 0))+frameBottomRight, frameKind(secondLeads))
	return append(lines, line{gapRun(textColumn), bottom})
}

// paneEdge is a pane's top edge across a width, between two corners, in a kind: its name set in at the left, bold
// where the pane leads somewhere, faint where not.
func paneEdge(left, right string, side paneSide, edge kind, across int) line {
	nameKind := kindStrong
	if side.neighbour == nil {
		nameKind = kindFaint
	}
	name := space + side.name + space
	rest := max(across-frameNameInset-textWidth(name), 0)
	return line{gapRun(textColumn), plainRun(left+strings.Repeat(frameAcross, frameNameInset), edge), plainRun(name, nameKind), plainRun(strings.Repeat(frameAcross, rest)+right, edge)}
}

// paneRows are a pane's rows across a width, its words between a row of room above and below, its edges in a kind.
func paneRows(words []line, edge kind, across int) []line {
	rows := []line{}
	for _, shown := range slices.Concat(make([]line, frameRowsRound), words, make([]line, frameRowsRound)) {
		cell := append(line{gapRun(framePadding)}, shown...)
		rows = append(rows, slices.Concat(line{gapRun(textColumn), plainRun(frameDown, edge)}, cell, line{gapRun(across - runsWidth(cell)), plainRun(frameDown, edge)}))
	}
	return rows
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
