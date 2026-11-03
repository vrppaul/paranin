package tui

import (
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/vrppaul/paranin/internal/document"
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

// The marks and words of the layout itself: a point's bullet, what parts a title from its subtitle and a period
// from its place, the start of a period from its end; as Neovim's :checkhealth writes them, the mark after a
// group's name and the word before a check's advice; and as lazy.nvim marks a plugin loaded or not, the mark of a
// skill in use and of one still being learnt
const (
	bulletMark   = "-"
	subtitleMark = "—"
	placeMark    = "·"
	rangeMark    = "–"
	monthLayout  = "2006-01"
	groupMark    = "~"
	adviceWord   = "ADVICE:"
	usedMark     = "●"
	learningMark = "○"
)

// levelLook is how a check's level reads: the word Neovim's :checkhealth writes it in, and the kind of that word.
type levelLook struct {
	word string
	kind kind
}

// levelLooks are how each level of a check reads. exhaustive holds the literal to every level, so a level the
// document gains reads its own way here too.
func levelLooks() map[document.CheckLevel]levelLook {
	return map[document.CheckLevel]levelLook{
		document.CheckOK:      {word: "OK", kind: kindPassed},
		document.CheckWarning: {word: "WARNING", kind: kindWarned},
		document.CheckError:   {word: "ERROR", kind: kindFailed},
	}
}

// The site's cursor: the bar after the author's name, as the page, the PDF and the card draw it
// (docs/FOUNDATION.md §3.4)
const cursorBar = "▎"

// tab is one screen of a document a reader moves between: the head and the blocks before the first section, or
// a section's blocks under its title.
type tab struct {
	title string
	// tally is the tab's checks counted by level, its sections' included, as the status line names them after the
	// title, and empty where the tab holds none
	tally string
	// head is the document's head on the tab that opens with it, and nothing on a section's
	head   *document.Head
	blocks []document.Block
}

// tabsOf parts a document into its tabs: the head with what stands before the first section, then each section.
func tabsOf(composed document.Document) []tab {
	opening := tab{title: document.PlainText(composed.Head.Title), tally: "", head: &composed.Head, blocks: nil}
	sections := []tab{}
	for _, block := range composed.Blocks {
		section, isSection := block.(document.Section)
		if !isSection {
			opening.blocks = append(opening.blocks, block)
			continue
		}
		sections = append(sections, tab{title: document.PlainText(section.Title), tally: checkTally(section.Blocks), head: nil, blocks: section.Blocks})
	}
	opening.tally = checkTally(opening.blocks)
	return append([]tab{opening}, sections...)
}

// checkTally is how many checks a tab's blocks hold at each level that has one, its sections' included, in the
// order the document gives the levels, the count before Neovim's word for the level: "1 OK · 1 WARNING".
func checkTally(blocks []document.Block) string {
	looks, counted := levelLooks(), []string{}
	for _, level := range slices.Sorted(maps.Keys(looks)) {
		if count := checksAt(blocks, level); count > 0 {
			counted = append(counted, strconv.Itoa(count)+space+looks[level].word)
		}
	}
	return strings.Join(counted, tallySeparator)
}

// checksAt is how many checks blocks hold at a level, the blocks of their sections included.
func checksAt(blocks []document.Block, level document.CheckLevel) int {
	count := 0
	for _, block := range blocks {
		switch typed := block.(type) {
		case document.Section:
			count += checksAt(typed.Blocks, level)
		case document.Checks:
			count += checksOfLevel(typed.List, level)
		case document.Paragraph, document.Table, document.CodeBlock, document.Facts, document.Entry, document.Card,
			document.Contents, document.Skills:
		}
	}
	return count
}

// checksOfLevel is how many of a group's checks read at a level.
func checksOfLevel(checks []document.Check, level document.CheckLevel) int {
	count := 0
	for _, check := range checks {
		if check.Level == level {
			count++
		}
	}
	return count
}

// layout lays a tab out in lines of a width; the site writes the address each link leads to.
type layout struct {
	site  routes.Site
	width int
}

// tab lays out a tab: the head, where it is the opening one, then its blocks.
func (layout layout) tab(shown tab) []line {
	if shown.head == nil {
		return layout.blocks(shown.blocks, 0)
	}
	lines := layout.head(*shown.head)
	if len(shown.blocks) == 0 {
		return lines
	}
	return append(append(lines, blankLines()...), layout.blocks(shown.blocks, 0)...)
}

// head lays out the title, with the site's cursor after it where the page carries one, the subtitle and place, the
// lead, and the ways to reach the author.
func (layout layout) head(head document.Head) []line {
	title := layout.words(head.Title, kindStrong)
	if carriesCursor(head.Page.Kind) {
		title = withCursor(title)
	}
	lines := layout.fill(title, 0)
	beside := placedWords(layout.words(head.Subtitle, kindText), textWords(head.Place, kindMuted))
	lines = append(lines, layout.fill(beside, 0)...)
	if lead := layout.fill(layout.words(head.Lead, kindText), 0); len(lead) > 0 {
		lines = append(append(lines, blankLines()...), lead...)
	}
	if len(head.Ways) > 0 {
		lines = append(append(lines, blankLines()...), layout.facts(head.Ways, 0)...)
	}
	return lines
}

// carriesCursor is whether a page's head ends in the site's cursor: the résumé's does, after the author's name, as
// the first version's page drew it; a project's, the index's and the hidden pages' do not.
func carriesCursor(kind document.PageKind) bool {
	switch kind {
	case document.PageResume:
		return true
	case document.PageNone, document.PageProject, document.PageIndex, document.PageHealth, document.PageSkills:
	}
	return false
}

// withCursor is a title's words with the site's cursor at the end of the last, so a line parts the two only where
// that word is wider than the line and breaks, bar and all, as fill breaks any; a title of no words is left with
// none.
func withCursor(words []word) []word {
	last := len(words) - 1
	if last < 0 {
		return words
	}
	marked := slices.Clone(words)
	cursor := run{text: cursorBar, kind: kindCursor, highlighted: false, address: "", copied: ""}
	marked[last].runs = append(slices.Clone(marked[last].runs), cursor)
	return marked
}

// blocks lays out blocks at an indent, a blank line between each two.
func (layout layout) blocks(blocks []document.Block, indent int) []line {
	lines := []line{}
	for index, block := range blocks {
		if index > 0 {
			lines = append(lines, blankLines()...)
		}
		lines = append(lines, layout.block(block, indent)...)
	}
	return lines
}

func (layout layout) block(block document.Block, indent int) []line {
	switch typed := block.(type) {
	case document.Section:
		title := layout.fill(layout.words(typed.Title, kindStrong), indent)
		return append(title, layout.blocks(typed.Blocks, indent+nestIndent)...)
	case document.Paragraph:
		return layout.fill(layout.words(typed.Spans, kindText), indent)
	case document.Table:
		return layout.table(typed, indent)
	case document.CodeBlock:
		return codeLines(typed.Text, indent+nestIndent)
	case document.Facts:
		return layout.facts(typed.List, indent)
	case document.Entry:
		return layout.entry(typed, indent)
	case document.Card:
		return layout.card(typed, indent)
	case document.Contents:
		return layout.contents(typed, indent)
	case document.Checks:
		return layout.checks(typed, indent)
	case document.Skills:
		return layout.skills(typed, indent)
	}
	return nil
}

// checks lays out a group of checks as Neovim's :checkhealth does: the group's name and its mark, then each check
// as a point opening with the word of its level, and its advice as a point under it.
func (layout layout) checks(checks document.Checks, indent int) []line {
	lines := layout.fill(append(layout.words(checks.Title, kindStrong), textWords(groupMark, kindFaint)...), indent)
	looks := levelLooks()
	for _, check := range checks.List {
		level := looks[check.Level]
		lines = append(lines, layout.bullet(append(textWords(level.word, level.kind), layout.words(check.Spans, kindText)...), indent)...)
		if len(check.Advice) > 0 {
			advice := append(textWords(adviceWord, kindMuted), layout.words(check.Advice, kindText)...)
			lines = append(lines, layout.bullet(advice, indent+bulletHang)...)
		}
	}
	return lines
}

// skills lays out each skill as lazy.nvim lays out a plugin, a table's row after the mark of whether the author uses
// it or is still learning it: its name, how long, and where it was proven.
func (layout layout) skills(skills document.Skills, indent int) []line {
	rows := make([]tableRow, 0, len(skills.List))
	for _, skill := range skills.List {
		cells := []document.Cell{{Spans: skill.Name}, {Spans: skill.Duration}, {Spans: skill.Sources}}
		rows = append(rows, tableRow{opening: []run{skillMark(skill.Learning)}, cells: cells})
	}
	return layout.rows(rows, indent)
}

// skillMark is the mark before a skill: lazy.nvim's mark of a plugin loaded for one in use, of one not loaded for
// one still being learnt.
func skillMark(learning bool) run {
	if learning {
		return run{text: learningMark + space, kind: kindFaint, highlighted: false, address: "", copied: ""}
	}
	return run{text: usedMark + space, kind: kindPassed, highlighted: false, address: "", copied: ""}
}

// entry lays out an entry's title and subtitle with its period and place at the right, then its points.
func (layout layout) entry(entry document.Entry, indent int) []line {
	left := joinedWords(layout.words(entry.Title, kindStrong), layout.words(entry.Subtitle, kindMuted))
	right := placedWords(textWords(periodText(entry.Period), kindMuted), textWords(entry.Place, kindMuted))
	lines := layout.titleLine(left, right, indent)
	for _, point := range entry.Points {
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
	lines := layout.titleLine(joinedWords(title, layout.words(card.Subtitle, kindMuted)), textWords(card.Tag, kindMuted), indent)
	lines = append(lines, layout.fill(layout.words(card.Summary, kindText), indent)...)
	return append(lines, layout.facts(card.Facts, indent)...)
}

// contents lays out each listed page as a point: its title, leading to it, and the line on what it is.
func (layout layout) contents(contents document.Contents, indent int) []line {
	lines := []line{}
	for _, listing := range contents.Listings {
		title := layout.words([]document.Span{document.Link{Target: listing.Page, Spans: listing.Title}}, kindStrong)
		lines = append(lines, layout.bullet(joinedWords(title, layout.words(listing.Note, kindMuted)), indent)...)
	}
	return lines
}

// facts lays out labels and their values, each value in a column after the widest label.
func (layout layout) facts(facts []document.Fact, indent int) []line {
	labelWidth := 0
	for _, fact := range facts {
		labelWidth = max(labelWidth, textWidth(fact.Label))
	}
	column := labelWidth + columnGap
	lines := []line{}
	for _, fact := range facts {
		lead := line{indentRun(indent), {text: padded(fact.Label, column), kind: kindMuted, highlighted: false, address: "", copied: ""}}
		lines = append(lines, layout.continued(lead, layout.valueWords(fact.Value), indent+column)...)
	}
	return lines
}

// table lays out a table's rows, none opening with a mark.
func (layout layout) table(table document.Table, indent int) []line {
	rows := make([]tableRow, 0, len(table.Rows))
	for _, row := range table.Rows {
		rows = append(rows, tableRow{opening: nil, cells: row.Cells})
	}
	return layout.rows(rows, indent)
}

// tableRow is a row as rows lays it out: the runs it opens with, a mark or none, and its cells.
type tableRow struct {
	opening []run
	cells   []document.Cell
}

// rows lays out rows a line each after its opening, the openings as wide as the widest, and each cell in a column as
// wide as the widest cell of its own, the last wrapping under itself. Where the columns would leave the last less
// room than it needs, each row is words that wrap after its opening, its cells parted by the place mark and its
// lines after the first standing in under it, as a point's do, so a narrow terminal cuts no cell off and each row
// shows where it starts.
func (layout layout) rows(rows []tableRow, indent int) []line {
	widths, start := layout.columnWidths(rows), indent+openingsWidth(rows)
	lastColumn := start
	for _, width := range widths[:max(len(widths)-1, 0)] {
		lastColumn += width + columnGap
	}
	columned := layout.width-lastColumn >= narrowestLastColumn
	lines := []line{}
	for _, row := range rows {
		lead := append(line{indentRun(indent)}, row.opening...)
		if pad := start - indent - runsWidth(row.opening); pad > 0 {
			lead = append(lead, gapRun(pad))
		}
		if columned {
			lines = append(lines, layout.columnedRow(lead, row.cells, widths, start)...)
			continue
		}
		lines = append(lines, layout.continued(lead, layout.flowingCells(row.cells), start+bulletHang)...)
	}
	return lines
}

// columnedRow lays out a row's cells after its lead, each but the last padded to its column's width, the last
// wrapping under itself.
func (layout layout) columnedRow(lead line, cells []document.Cell, widths []int, column int) []line {
	for place, cell := range cells {
		words := layout.words(cell.Spans, cellKind(place))
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
		words = markedWords(words, layout.words(cell.Spans, cellKind(place)), placeMark)
	}
	return words
}

// cellKind is what a cell of a row is: the first is read as text, and the rest as what stands beside it.
func cellKind(place int) kind {
	if place == 0 {
		return kindText
	}
	return kindMuted
}

// columnWidths are the widths of the rows' columns, each as wide as its widest cell laid in a line.
func (layout layout) columnWidths(rows []tableRow) []int {
	widths := []int{}
	for _, row := range rows {
		for place, cell := range row.cells {
			if place == len(widths) {
				widths = append(widths, 0)
			}
			widths[place] = max(widths[place], wordsWidth(layout.words(cell.Spans, kindText)))
		}
	}
	return widths
}

// openingsWidth is how wide the widest opening of the rows is.
func openingsWidth(rows []tableRow) int {
	widest := 0
	for _, row := range rows {
		widest = max(widest, runsWidth(row.opening))
	}
	return widest
}

// bullet lays out words as a point: its mark, then the words, the lines after the first standing in under them.
func (layout layout) bullet(words []word, indent int) []line {
	lead := line{indentRun(indent), {text: bulletMark + space, kind: kindFaint, highlighted: false, address: "", copied: ""}}
	return layout.continued(lead, words, indent+bulletHang)
}

// continued lays out words after a lead already on the line, the words' lines standing in at a column.
func (layout layout) continued(lead line, words []word, column int) []line {
	lines := layout.fill(words, column)
	if len(lines) == 0 {
		return []line{lead}
	}
	lines[0] = append(lead, lines[0][1:]...)
	return lines
}

// titleLine lays out left words with right words at the right of the same line, or each on lines of its own
// where both do not fit.
func (layout layout) titleLine(left, right []word, indent int) []line {
	leftWidth, rightWidth := wordsWidth(left), wordsWidth(right)
	gap := layout.width - indent - leftWidth - rightWidth
	if len(right) == 0 || gap < columnGap {
		return append(layout.fill(left, indent), layout.fill(right, indent)...)
	}
	return []line{append(append(line{indentRun(indent)}, wordRuns(left)...), append(line{gapRun(gap)}, wordRuns(right)...)...)}
}

// codeLines lays out text line for line at an indent, as it is written: never wrapped.
func codeLines(code string, indent int) []line {
	lines := []line{}
	for written := range strings.SplitSeq(code, lineEnd) {
		lines = append(lines, line{indentRun(indent), {text: written, kind: kindMuted, highlighted: false, address: "", copied: ""}})
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

// joinedWords are a title's words and its subtitle's, the subtitle's mark between them where both are there.
func joinedWords(title, subtitle []word) []word {
	return markedWords(title, subtitle, subtitleMark)
}

// placedWords are a period's or a subtitle's words and a place's, the place's mark between them where both are
// there.
func placedWords(period, place []word) []word {
	return markedWords(period, place, placeMark)
}

func markedWords(first, second []word, mark string) []word {
	if len(first) == 0 || len(second) == 0 {
		return append(first, second...)
	}
	return append(append(first, textWords(mark, kindMuted)...), second...)
}

func blankLines() []line {
	return make([]line, blockGap)
}
