package tui

import (
	"slices"
	"strings"
	"time"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/look"
	"github.com/vrppaul/paranin/internal/routes"
)

// How a tab lays its blocks out: how far what nests under a title stands in, how far a point's lines after its
// first stand in under its text, the gap after the widest label and between columns, the least room a table's last
// column may have, about two words of a place it names, and how many lines of nothing part two blocks
const (
	nestIndent          = 2
	bulletHang          = 2
	columnGap           = 2
	narrowestLastColumn = 16
	blockGap            = 1
)

// The least room the words of an entry may have beside its period's column, where the period stands above them
const narrowestEntry = 30

// The marks and words of the layout itself: what parts a title from its subtitle and a subtitle from its place, the
// start of a period from its end
const (
	subtitleMark = "—"
	placeMark    = "·"
	rangeMark    = "–"
	monthLayout  = "2006-01"
)

// The site's cursor: the bar after the author's name, as the page, the PDF and the card draw it
// (docs/FOUNDATION.md §3.4)
const cursorBar = "▎"

// tab is one screen of a document a reader moves between: the page it is of and its place among the page's tabs, which
// a reader's place in it is kept by; its title, the name it goes by, as the footer names it and the command line goes
// to it, and its body, which says what the tab shows and so how it is laid out.
type tab struct {
	page  document.Page
	place int
	title string
	slug  string
	body  tabBody
}

// laidTab is a tab laid out: its lines; on the blog's list the links a key goes to, in reading order, the head's then
// the tiles', the place among them of the one the keys stand on as the list opens, and the groups whose names stay in
// sight while their tiles scroll; the column what stays in sight over the lines' left part ends at, gap and all, the
// groups' names' column or a post's Outline's, where the tab has one; and the line each named section starts at, a
// post's heading's and a series' group's, which a place of the page goes to; and on a post its Outline, none on any
// other tab.
type laidTab struct {
	lines    []line
	links    []laidLink
	opening  int
	groups   []laidGroup
	split    int
	sections map[string]int
	outline  laidOutline
}

// overLines are the spans of a tab's lines what stays in sight may stand over: each group's of the blog's list, and
// a post's body beside its Outline.
func (laid laidTab) overLines() []lineSpan {
	spans := make([]lineSpan, 0, len(laid.groups)+1)
	for _, group := range laid.groups {
		spans = append(spans, lineSpan{first: group.first, end: group.end})
	}
	return append(spans, lineSpan{first: laid.outline.first, end: laid.outline.end})
}

// tabBody is what a tab shows: a document's opening, beside the brief, as a README, as a report of checks, as a list
// of skills or as it is, or a section, under its title or as it is; or the blog's list or a post, whole.
//
//sumtype:decl
type tabBody interface {
	tabBody()
}

// briefOpening is the résumé's opening as neofetch prints a system, where its first section holds facts: the head,
// beside it those facts as the brief, headed by the address, the first way to reach the author; the roles for the
// career's line under them; then the blocks before the first section and those the section holds beside its facts.
type briefOpening struct {
	head    document.Head
	brief   []document.Fact
	address []document.Span
	career  []document.Entry
	blocks  []document.Block
}

// readmeOpening is a project's opening as its forge shows its README: the head, then all its blocks, its
// sections among them.
type readmeOpening struct {
	head   document.Head
	blocks []document.Block
}

// reportOpening is the health report's opening as Neovim's :checkhealth, laid out anew: the head, then the blocks
// before the first section.
type reportOpening struct {
	head   document.Head
	blocks []document.Block
}

// skillsOpening is the list of skills' opening as lazy.nvim's :Lazy, laid out anew: the head's title in the titles'
// font, then the blocks before the first section, the skills among them.
type skillsOpening struct {
	head   document.Head
	blocks []document.Block
}

// plainOpening is any other opening: the head, then the blocks before the first section.
type plainOpening struct {
	head   document.Head
	blocks []document.Block
}

// titledSection is a section of the résumé, which stands under its title.
type titledSection struct {
	blocks []document.Block
}

// plainSection is a section of another document, which stands as it is, named in the footer alone.
type plainSection struct {
	blocks []document.Block
}

// listBody is the blog's list, its head and its groups, as the page's: one tab, the blog's first window.
type listBody struct {
	head   document.Head
	groups []document.Block
}

// postBody is a post, as the page draws it: one tab, the blog's window after its list's.
type postBody struct {
	post document.Document
}

func (briefOpening) tabBody()  {}
func (readmeOpening) tabBody() {}
func (reportOpening) tabBody() {}
func (skillsOpening) tabBody() {}
func (plainOpening) tabBody()  {}
func (titledSection) tabBody() {}
func (plainSection) tabBody()  {}
func (listBody) tabBody()      {}
func (postBody) tabBody()      {}

// tabsOf lays a document's tabs out as the TUI shows them, as document.TabsOf parts it: the blog's list, filtered or
// not, and a post whole are one tab each, which the blog's session names as its windows (Model.windows).
func tabsOf(composed document.Document) []tab {
	head := composed.Head
	page := head.Page
	parted := document.TabsOf(composed)
	opening := parted.Opening
	tabs := []tab{{page: page, place: 0, title: document.PlainText(opening.Title), slug: opening.Slug, body: openingBody(composed, opening)}}
	for _, shown := range parted.Rest {
		body := sectionBody(page.Kind, shown.Blocks)
		tabs = append(tabs, tab{page: page, place: len(tabs), title: document.PlainText(shown.Title), slug: shown.Slug, body: body})
	}
	return tabs
}

// openingBody is what a document's first tab shows: its brief as neofetch prints a system, or by the kind of its page
// a project's README, the health report, the list of skills, a post whole, the blog's list, or the head as it is.
func openingBody(composed document.Document, opening document.Opening) tabBody {
	head, blocks := composed.Head, opening.Blocks
	if briefed := opening.Brief; briefed != nil {
		return briefOpening{head: head, brief: briefed.Brief.List, address: briefed.Address(), career: briefed.Roles, blocks: blocks}
	}
	switch head.Page.Kind {
	case document.PageProject:
		return readmeOpening{head: head, blocks: blocks}
	case document.PageHealth:
		return reportOpening{head: head, blocks: blocks}
	case document.PageSkills:
		return skillsOpening{head: head, blocks: blocks}
	case document.PagePost:
		return postBody{post: composed}
	case document.PageBlog, document.PageBlogSeries, document.PageTagFilter:
		return listBody{head: head, groups: blocks}
	case document.PageNone, document.PageResume, document.PageIndex:
	}
	return plainOpening{head: head, blocks: blocks}
}

// sectionBody is what a section's tab shows of its blocks: on the résumé under its title, on another page as they are;
// a project's page has no section tabs.
func sectionBody(kind document.PageKind, blocks []document.Block) tabBody {
	switch kind {
	case document.PageResume:
		return titledSection{blocks: blocks}
	case document.PageNone, document.PageProject, document.PageIndex, document.PageHealth, document.PageSkills, document.PagePost, document.PageBlog, document.PageBlogSeries,
		document.PageTagFilter:
	}
	return plainSection{blocks: blocks}
}

// layout lays a tab out in lines of a width, on a page of a number of rows, on a day; the site writes the address
// each link leads to, and siteWords are the site's words of the frame a post and the blog's list lay out round them.
type layout struct {
	site      routes.Site
	siteWords compose.Frame
	width     int
	// pageRows are how many rows the page shows
	pageRows int
	// today is the day a role still held runs to on the career's line
	today time.Time
}

// tab lays out a tab as its body says: an opening's head, then its blocks; a section under its title, in the titles'
// font; another section as it is; the blog's list and a post as list.go and post.go lay them out.
func (layout layout) tab(shown tab) laidTab {
	switch body := shown.body.(type) {
	case briefOpening:
		return linesAlone(stacked(layout.about(body.head, body.brief, body.address, body.career), layout.blocks(body.blocks, 0)))
	case readmeOpening:
		return linesAlone(stacked(layout.readme(body.head), layout.blocks(body.blocks, 0)))
	case reportOpening:
		return linesAlone(stacked(layout.report(body.head, body.blocks), layout.blocks(body.blocks, 0)))
	case skillsOpening:
		return linesAlone(stacked(layout.titleRows(document.PlainText(body.head.Title)), layout.blocks(body.blocks, 0)))
	case plainOpening:
		return linesAlone(stacked(layout.head(body.head), layout.blocks(body.blocks, 0)))
	case titledSection:
		return linesAlone(slices.Concat(layout.heading(shown.title), layout.blocks(body.blocks, 0)))
	case plainSection:
		return linesAlone(layout.blocks(body.blocks, 0))
	case listBody:
		return layout.list(body.head, body.groups)
	case postBody:
		return layout.post(body.post)
	}
	return linesAlone(nil)
}

// linesAlone is a tab laid out in lines, with no link, no group of the blog's list and no named section.
func linesAlone(lines []line) laidTab {
	return laidTab{lines: lines, links: nil, opening: 0, groups: nil, split: 0, sections: nil, outline: noOutline()}
}

// stacked is lines above others, a row of nothing between where there are lines on both sides.
func stacked(above, below []line) []line {
	if len(above) == 0 || len(below) == 0 {
		return slices.Concat(above, below)
	}
	return slices.Concat(above, blankLines(), below)
}

// headParts are the parts of a head every opening lays out alike, about its title: the greeting and the lead; each
// nothing where the head has none.
type headParts struct {
	greeting []line
	lead     []line
}

// headParts lays out the parts of a head every opening lays out alike.
func (layout layout) headParts(head document.Head) headParts {
	return headParts{
		greeting: layout.fill(textWords(head.Greeting, kindMuted), 0),
		lead:     layout.fill(layout.words(head.Lead, kindText), 0),
	}
}

// subtitleLines are a head's subtitle with its place beside it, under the title or in the brief.
func (layout layout) subtitleLines(head document.Head) []line {
	return layout.fill(markedWords(layout.words(head.Subtitle, kindText), textWords(head.Place, kindMuted), placeMark), 0)
}

// head lays out the greeting, where the head has one, then the title, with the site's cursor after it where the
// page carries one, the subtitle and place, and the lead. A project's head, the one with tags and a repository, is a
// README's (readme).
func (layout layout) head(head document.Head) []line {
	parts := layout.headParts(head)
	title := layout.words(head.Title, kindStrong)
	if carriesCursor(head.Page.Kind) {
		title = withCursor(title)
	}
	lines := slices.Concat(parts.greeting, layout.fill(title, 0), layout.subtitleLines(head))
	return stacked(lines, parts.lead)
}

// carriesCursor is whether a page's head ends in the site's cursor: the résumé's does, after the author's name, as
// the first version's page drew it; a project's, the index's and the hidden pages' do not.
func carriesCursor(kind document.PageKind) bool {
	switch kind {
	case document.PageResume:
		return true
	case document.PageNone, document.PageProject, document.PageIndex, document.PageHealth, document.PageSkills, document.PagePost,
		document.PageBlog, document.PageBlogSeries, document.PageTagFilter:
	}
	return false
}

// withCursor is a title's words with the site's cursor at the end of the last, so a line parts the two only where
// that word is wider than the line and breaks, bar and all, as fill breaks any; a title of no words is left with
// none. It stands still: a blink would paint the head again each half second, where a tab is painted once.
func withCursor(words []word) []word {
	last := len(words) - 1
	if last < 0 {
		return words
	}
	marked := slices.Clone(words)
	marked[last].runs = append(slices.Clone(marked[last].runs), plainRun(cursorBar, kindCursor))
	return marked
}

// heading is a tab's title over its page as titleRows lays it out, a row of nothing under it; nothing where there is
// no title to lay out.
func (layout layout) heading(title string) []line {
	rows := layout.titleRows(title)
	if len(rows) == 0 {
		return nil
	}
	return slices.Concat(rows, blankLines())
}

// titleRows are a tab's title: in the titles' font, smaller than the name's, its colours going from the accent to the
// host's, where the font draws every letter of it and it and a row of nothing under it fit the page; as text in the
// same colours where that fits; nothing where neither does, or the tab has no title, the footer naming the tab.
func (layout layout) titleRows(title string) []line {
	if title == "" {
		return nil
	}
	if rows := look.TitleFont().Banner([]string{strings.ToUpper(title)}); len(rows) > 0 {
		if banner := tonedBanner(rows, kindHeading, rowsWidth(rows[0])); layout.fits(slices.Concat(banner, blankLines())) {
			return banner
		}
	}
	if text := layout.fill(gradientWords(title, kindHeading), 0); layout.fits(slices.Concat(text, blankLines())) {
		return text
	}
	return nil
}

// blocks lays out blocks at an indent, a blank line between each two, the periods of the entries among them in one
// column, as wide as the widest of them.
func (layout layout) blocks(blocks []document.Block, indent int) []line {
	periodWidth := periodWidthOf(blocks)
	lines := []line{}
	for index, block := range blocks {
		if index > 0 {
			lines = append(lines, blankLines()...)
		}
		lines = append(lines, layout.block(block, indent, periodWidth)...)
	}
	return lines
}

// block lays out a block at an indent, an entry's period in a column of a width.
func (layout layout) block(block document.Block, indent, periodWidth int) []line {
	switch typed := block.(type) {
	case document.Section:
		title := layout.fill(layout.words(typed.Title, kindStrong), indent)
		return append(title, layout.blocks(typed.Blocks, indent+nestIndent)...)
	case document.Paragraph:
		return layout.fill(layout.words(typed.Spans, kindText), indent)
	case document.Table:
		return layout.table(typed, indent)
	case document.CodeBlock:
		return codeLines(typed.Text(), indent+nestIndent)
	case document.Facts:
		return layout.factBlock(typed.List, indent)
	case document.Entry:
		return layout.entry(typed, indent, periodWidth)
	case document.Card:
		return layout.card(typed, indent)
	case document.Contents:
		return layout.contents(typed, indent)
	case document.Checks:
		return layout.checks(typed, indent)
	case document.Skills:
		return layout.skills(typed, indent)
	case document.List, document.Quote, document.Pause, document.Figure, document.Widget, document.Posts:
		// a post's and the blog's list's alone, which post.go and list.go lay out
	}
	return nil
}

// entry lays out an entry: its period in a column of a width, the widest period of those laid out with it, the one
// still going on standing out, and beside it its title, its subtitle and its place, its points under them; where the
// page leaves the words beside the column too little room, the period stands above them.
func (layout layout) entry(entry document.Entry, indent, periodWidth int) []line {
	periodKind := kindMuted
	if entry.Period.GoesOn() {
		periodKind = kindCurrent
	}
	period := textWords(periodText(entry.Period), periodKind)
	titleWords := markedWords(markedWords(layout.words(entry.Title, kindStrong), layout.words(entry.Subtitle, kindMuted), subtitleMark), textWords(entry.Place, kindMuted), placeMark)
	column := indent + periodWidth + columnGap
	if layout.width-column < narrowestEntry {
		return slices.Concat(layout.fill(period, indent), layout.fill(titleWords, indent), layout.points(entry.Points, indent))
	}
	lead := append(line{gapRun(indent)}, wordRuns(period)...)
	lead = append(lead, gapRun(column-runsWidth(lead)))
	return append(layout.continued(lead, titleWords, column), layout.points(entry.Points, column)...)
}

// periodWidthOf is how wide the widest period of the entries among blocks is.
func periodWidthOf(blocks []document.Block) int {
	width := 0
	for _, block := range blocks {
		if entry, isEntry := block.(document.Entry); isEntry {
			width = max(width, textWidth(periodText(entry.Period)))
		}
	}
	return width
}

// points lays out an entry's points at an indent, each a bullet.
func (layout layout) points(points []document.Point, indent int) []line {
	lines := []line{}
	for _, point := range points {
		lines = append(lines, layout.bullet(layout.words(point.Spans, kindText), indent)...)
	}
	return lines
}

// card lays out a card's title, leading to its page where it has one, and its subtitle, with its tag at the
// right; then its summary and its facts.
func (layout layout) card(card document.Card, indent int) []line {
	title := layout.words(card.Title, kindStrong)
	if card.Page.Kind != document.PageNone {
		title = layout.words([]document.Span{document.Link{Target: card.Page, Spans: card.Title}}, kindStrong)
	}
	lines := layout.titleLine(markedWords(title, layout.words(card.Subtitle, kindMuted), subtitleMark), textWords(card.Tag, kindMuted), indent)
	lines = append(lines, layout.fill(layout.words(card.Summary, kindText), indent)...)
	return append(lines, layout.facts(card.Facts, indent, kindMuted)...)
}

// contents lays out each listed page as a point: its title, leading to it, and the line on what it is.
func (layout layout) contents(contents document.Contents, indent int) []line {
	lines := []line{}
	for _, listing := range contents.Listings {
		title := layout.words([]document.Span{document.Link{Target: listing.Page, Spans: listing.Title}}, kindStrong)
		lines = append(lines, layout.bullet(markedWords(title, layout.words(listing.Note, kindMuted), subtitleMark), indent)...)
	}
	return lines
}

// facts lays out labels of a kind and their values, each value in a column after the widest label.
func (layout layout) facts(facts []document.Fact, indent int, labelKind kind) []line {
	column := valueColumn(facts)
	lines := []line{}
	for _, fact := range facts {
		lines = append(lines, layout.fact(fact, indent, column, labelKind)...)
	}
	return lines
}

// valueColumn is how far in from their labels facts' values stand: after the widest label, a gap between.
func valueColumn(facts []document.Fact) int {
	labelWidth := 0
	for _, fact := range facts {
		labelWidth = max(labelWidth, textWidth(fact.Label))
	}
	return labelWidth + columnGap
}

// fact lays out a label of a kind and its value, the value in a column a width in from the label.
func (layout layout) fact(fact document.Fact, indent, column int, labelKind kind) []line {
	lead := line{gapRun(indent), plainRun(padded(fact.Label, column), labelKind)}
	return layout.continued(lead, layout.valueWords(fact.Value), indent+column)
}

// factBlock lays out a block of facts at an indent, in their order, the values in one column: each fact under its
// label, but a command, a fact whose value is code alone, in a box with its label in the top edge, as a README frames
// a command to type, a row of nothing round the box.
func (layout layout) factBlock(facts []document.Fact, indent int) []line {
	column := valueColumn(facts)
	parts, standing := [][]line{}, []line{}
	for _, fact := range facts {
		if box := layout.commandBox(fact, indent); len(box) > 0 {
			parts, standing = append(parts, standing, box), nil
			continue
		}
		standing = append(standing, layout.fact(fact, indent, column, kindMuted)...)
	}
	lines := []line{}
	for _, part := range append(parts, standing) {
		lines = stacked(lines, part)
	}
	return lines
}

// commandBox is a command in a box as wide as the page, as the page's Command stands: its label set into the top edge,
// and the command on one row inside it after the prompt's mark, so a reader copies it whole; nothing for a fact that
// is no command, or where the page is too narrow for the box's edges round the label or round the command's row, the
// fact then standing under its label.
func (layout layout) commandBox(fact document.Fact, indent int) []line {
	if _, isCommand := fact.Command(); !isCommand {
		return nil
	}
	width := layout.width - indent
	topRoom := width - textWidth(frameTopLeft+frameAcross+space+fact.Label+space+frameTopRight)
	prompted := slices.Concat(line{
		boxRun(frameDown + space),
		plainRun(look.PromptMark, kindPrompt),
		gapRun(spaceWidth),
	}, wordRuns(layout.valueWords(fact.Value)))
	insideRoom := width - runsWidth(prompted) - textWidth(space+frameDown)
	if topRoom < 1 || insideRoom < 0 {
		return nil
	}
	top := line{gapRun(indent), boxRun(frameTopLeft + frameAcross + space), plainRun(fact.Label, kindMuted), boxRun(space + strings.Repeat(frameAcross, topRoom) + frameTopRight)}
	inside := slices.Concat(line{gapRun(indent)}, prompted, line{gapRun(insideRoom), boxRun(space + frameDown)})
	bottom := frameBottomLeft + strings.Repeat(frameAcross, width-textWidth(frameBottomLeft+frameBottomRight)) + frameBottomRight
	return []line{top, inside, {gapRun(indent), boxRun(bottom)}}
}

// boxRun is a piece of a box's edge, muted as the page draws the Command's hairline, so the box shows on any ground.
func boxRun(edge string) run {
	return run{text: edge, kind: kindMuted, highlighted: false, address: "", copied: "", tone: 0, cites: noCitation, opens: document.PageOf(document.PageNone)}
}

// table lays out a table's rows a line each, each cell in a column as wide as the widest cell of its own, the last
// wrapping under itself. Where the columns would leave the last less room than it needs, each row is words that
// wrap, its cells parted by the place mark and its lines after the first standing in under it, as a point's do, so a
// narrow terminal cuts no cell off and each row shows where it starts.
func (layout layout) table(table document.Table, indent int) []line {
	widths := layout.columnWidths(table.Rows)
	lastColumn := indent
	for _, width := range widths[:max(len(widths)-1, 0)] {
		lastColumn += width + columnGap
	}
	columned := layout.width-lastColumn >= narrowestLastColumn
	lines := []line{}
	for _, row := range table.Rows {
		lead := line{gapRun(indent)}
		if columned {
			lines = append(lines, layout.columnedRow(lead, row.Cells, widths, indent)...)
			continue
		}
		lines = append(lines, layout.continued(lead, layout.flowingCells(row.Cells), indent+bulletHang)...)
	}
	return lines
}

// columnedRow lays out a row's cells after its lead, each but the last padded to its column's width, the last
// wrapping under itself.
func (layout layout) columnedRow(lead line, cells []document.Cell, widths []int, column int) []line {
	for place, cell := range cells {
		words := layout.words(cell.Spans, cellLook(place, cell.Kind))
		if place == len(cells)-1 {
			return layout.continued(lead, words, column)
		}
		lead = append(append(lead, wordRuns(words)...), gapRun(widths[place]+columnGap-wordsWidth(words)))
		column += widths[place] + columnGap
	}
	return []line{lead}
}

// flowingCells are a row's cells as words in a line, the place mark between two cells.
func (layout layout) flowingCells(cells []document.Cell) []word {
	words := []word{}
	for place, cell := range cells {
		words = markedWords(words, layout.words(cell.Spans, cellLook(place, cell.Kind)), placeMark)
	}
	return words
}

// cellLook is how a cell of a row reads (DESIGN.md, Table): the first names the row, and the rest read as what they
// hold, a duration set back from the words beside it.
func cellLook(place int, cellKind document.CellKind) kind {
	if place == 0 {
		return kindStrong
	}
	switch cellKind {
	case document.CellText:
		return kindText
	case document.CellDuration:
		return kindMuted
	}
	// the zero kind, which compose never builds, reads as the words around it
	return kindText
}

// columnWidths are the widths of the rows' columns, each as wide as its widest cell laid in a line.
func (layout layout) columnWidths(rows []document.Row) []int {
	widths := []int{}
	for _, row := range rows {
		for place, cell := range row.Cells {
			if place == len(widths) {
				widths = append(widths, 0)
			}
			widths[place] = max(widths[place], wordsWidth(layout.words(cell.Spans, kindText)))
		}
	}
	return widths
}

// bullet lays out words as a point: its mark, then the words, the lines after the first standing in under them.
func (layout layout) bullet(words []word, indent int) []line {
	lead := line{gapRun(indent), plainRun(look.PointMark+space, kindFaint)}
	return layout.continued(lead, words, indent+bulletHang)
}

// titleLine lays out left words with right words at the right of the same line, or each on lines of its own
// where both do not fit.
func (layout layout) titleLine(left, right []word, indent int) []line {
	leftWidth, rightWidth := wordsWidth(left), wordsWidth(right)
	gap := layout.width - indent - leftWidth - rightWidth
	if len(right) == 0 || gap < columnGap {
		return append(layout.fill(left, indent), layout.fill(right, indent)...)
	}
	return []line{append(append(line{gapRun(indent)}, wordRuns(left)...), append(line{gapRun(gap)}, wordRuns(right)...)...)}
}

// codeLines lays out text line for line at an indent, as it is written: never wrapped.
func codeLines(code string, indent int) []line {
	lines := []line{}
	for written := range strings.SplitSeq(code, lineEnd) {
		lines = append(lines, line{gapRun(indent), plainRun(written, kindMuted)})
	}
	return lines
}

// periodText writes a period: its start and its end, a known month as its year and its number.
func periodText(period document.Period) string {
	return monthText(period.Start) + space + rangeMark + space + monthText(period.End)
}

func monthText(month document.Month) string {
	switch typed := month.(type) {
	case document.KnownMonth:
		return typed.Start.Format(monthLayout)
	case document.GapMonth:
		return typed.Text
	case document.CurrentMonth:
		return typed.Text
	}
	return ""
}

// markedWords are two runs of words with a mark between them where both are there: a title and its subtitle, words
// and their place.
func markedWords(first, second []word, mark string) []word {
	if len(first) == 0 || len(second) == 0 {
		return append(first, second...)
	}
	return append(append(first, textWords(mark, kindMuted)...), second...)
}

func blankLines() []line {
	return make([]line, blockGap)
}
