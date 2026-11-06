package tui

import (
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/vrppaul/paranin/internal/document"
)

// How the name stands in the block font: how many rows a glyph is tall, a column between two glyphs, a row of nothing
// between its two lines, and the cursor after the second line, a column apart from it, as wide as the font's strokes
const (
	glyphRows    = 3
	glyphGap     = 1
	blockLineGap = 1
	blockCursor  = "▌"
)

// The room between the name's column and the brief's, as the storyboard set them apart, and the least the brief
// takes beside the name, where the page has not the room for both, the brief standing under the name; the least room
// the career's line is drawn in, beside the site's word for the present; and how the career's line counts and spaces
// what it draws
const (
	columnsApart    = 4
	narrowestBrief  = 36
	narrowestCareer = 24
	monthsInYear    = 12
	yearLabelGap    = 1
	presentGap      = 1
	shortestCutName = 2
)

// The marks of the brief: the rule under the address, the at the address parts at, and a swatch of the design's
// colours; and of the career's line: a role's span, the line the spans stand on, a year's tick, and the end of a
// company's name cut to its span
const (
	briefRuleMark  = "─"
	addressAt      = "@"
	swatchMark     = "███"
	spanMark       = "█"
	careerLineMark = "─"
	yearMark       = "╵"
	cutMark        = "…"
)

// The rows of the career's line, from the top: the companies, the spans, the years' ticks, the years; and how many
// they are
const (
	companyRow = iota
	spanRow
	tickRow
	yearRow
	careerRows
)

// blockFont is the first version's banner font: each capital, digit, space and hyphen three rows tall, of half blocks,
// each glyph's rows written as one text, a line's end between each two. A name with a letter it has not is drawn as
// text.
func blockFont() map[rune][]string {
	glyphs := map[rune]string{
		'A': "█▀▀█\n█▀▀█\n▀  ▀", 'B': "█▀▀▄\n█▀▀█\n▀▀▀ ", 'C': "█▀▀▀\n█   \n▀▀▀▀", 'D': "█▀▀▄\n█  █\n▀▀▀ ",
		'E': "█▀▀▀\n█▀▀ \n▀▀▀▀", 'F': "█▀▀▀\n█▀▀ \n▀   ", 'G': "█▀▀▀\n█ ▀█\n▀▀▀▀", 'H': "█  █\n█▀▀█\n▀  ▀",
		'I': "▀█▀\n █ \n▀▀▀", 'J': "   █\n▄  █\n▀▀▀▀", 'K': "█ ▄▀\n█▀▄ \n▀  ▀", 'L': "█   \n█   \n▀▀▀▀",
		'M': "█▄ ▄█\n█ ▀ █\n▀   ▀", 'N': "█▄ █\n█ ▀█\n▀  ▀", 'O': "█▀▀█\n█  █\n▀▀▀▀", 'P': "█▀▀█\n█▀▀▀\n▀   ",
		'Q': "█▀▀█\n█ ▄█\n▀▀▀▀", 'R': "█▀▀█\n█▀▀▄\n▀  ▀", 'S': "█▀▀▀\n▀▀▀█\n▀▀▀▀", 'T': "▀█▀\n █ \n ▀ ",
		'U': "█  █\n█  █\n▀▀▀▀", 'V': "█  █\n█  █\n ▀▀ ", 'W': "█   █\n█▄▀▄█\n▀   ▀", 'X': "█  █\n▄▀▀▄\n▀  ▀",
		'Y': "█  █\n ██ \n ▀▀ ", 'Z': "▀▀▀█\n▄▀▀ \n▀▀▀▀", ' ': "  \n  \n  ", '-': "   \n▀▀▀\n   ",
		'1': "▀█ \n █ \n▀▀▀", '2': "▀▀▀█\n█▀▀▀\n▀▀▀▀", '3': "▀▀▀█\n ▀▀█\n▀▀▀▀", '4': "█  █\n▀▀▀█\n   ▀",
		'6': "█▀▀▀\n█▀▀█\n▀▀▀▀", '7': "▀▀▀█\n  █ \n  ▀ ", '8': "█▀▀█\n█▀▀█\n▀▀▀▀", '9': "█▀▀█\n▀▀▀█\n▀▀▀▀",
	}
	// the font draws a zero as the O and a five as the S
	glyphs['0'], glyphs['5'] = glyphs['O'], glyphs['S']
	font := make(map[rune][]string, len(glyphs))
	for letter, glyph := range glyphs {
		font[letter] = strings.Split(glyph, lineEnd)
	}
	return font
}

// about lays out the résumé's opening as neofetch prints a system: the name's column, the greeting and the name,
// beside the brief where the page has room for both, above it where not; then the career's line, and the lead.
func (layout layout) about(head document.Head, brief []document.Fact, career []document.Entry) []line {
	parts := layout.headParts(head)
	name := slices.Concat(parts.greeting, blankLines(), layout.name(document.PlainText(head.Title)))
	nameWidth := linesWidth(name)
	var lines []line
	if room := layout.width - nameWidth - columnsApart; room >= narrowestBrief {
		// the brief lays out in the column beside the name
		briefColumn := layout
		briefColumn.width = room
		lines = besideLines(name, nameWidth+columnsApart, briefColumn.brief(head, brief))
	} else {
		lines = slices.Concat(name, blankLines(), layout.brief(head, brief))
	}
	if careerLine := layout.career(career); len(careerLine) > 0 {
		lines = slices.Concat(lines, blankLines(), careerLine)
	}
	if len(parts.lead) > 0 {
		lines = slices.Concat(lines, blankLines(), parts.lead)
	}
	return lines
}

// name is the author's name: its first word above the rest in the block font, the cursor after them, where the font
// draws every letter of it and both lines fit the page; else as text in the same colours, the site's cursor after
// it, wrapped where the page is narrower than the name.
func (layout layout) name(title string) []line {
	if banner := bannerOf(strings.Fields(strings.ToUpper(title))); len(banner) > 0 && linesWidth(banner) <= layout.width {
		return banner
	}
	return layout.fill(withCursor(textNameWords(title)), 0)
}

// bannerOf is a name's words in the block font: the first above the rest, the cursor after them; nothing where the
// name has no words or the font does not draw every letter of them.
func bannerOf(words []string) []line {
	font := blockFont()
	if len(words) == 0 || !drawsEvery(font, words) {
		return nil
	}
	rows := [][]string{blockRows(font, words[0])}
	if len(words) > 1 {
		rows = append(rows, blockRows(font, strings.Join(words[1:], space)))
	}
	// the cursor stands after the last line, whose width it adds to
	last := rows[len(rows)-1]
	width := max(rowsWidth(rows[0]), rowsWidth(last)+textWidth(space+blockCursor))
	lines := tonedRows(rows[0], width)
	for _, written := range rows[1:] {
		lines = slices.Concat(lines, slices.Repeat(blankLines(), blockLineGap), tonedRows(written, width))
	}
	for row := len(lines) - len(last); row < len(lines); row++ {
		lines[row] = append(lines[row], gapRun(spaceWidth), run{text: blockCursor, kind: kindCursor, highlighted: false, address: "", copied: "", tone: 0})
	}
	return lines
}

// drawsEvery is whether the block font has a glyph for every letter of words.
func drawsEvery(font map[rune][]string, words []string) bool {
	for _, written := range words {
		for _, letter := range written {
			if _, drawn := font[letter]; !drawn {
				return false
			}
		}
	}
	return true
}

// blockRows are a text's rows in the block font, a column between each two glyphs; every letter of it is one the
// font draws.
func blockRows(font map[rune][]string, text string) []string {
	rows := make([]string, glyphRows)
	for index, letter := range text {
		glyph := font[letter]
		for row := range rows {
			if index > 0 {
				rows[row] += strings.Repeat(space, glyphGap)
			}
			rows[row] += glyph[row]
		}
	}
	return rows
}

// linesWidth is the widest of lines.
func linesWidth(lines []line) int {
	width := 0
	for _, shown := range lines {
		width = max(width, runsWidth(shown))
	}
	return width
}

// rowsWidth is the widest of rows of text.
func rowsWidth(rows []string) int {
	width := 0
	for _, row := range rows {
		width = max(width, textWidth(row))
	}
	return width
}

// tonedRows are rows of the block font as lines, each cell in the tone its column stands at along a width, cells of
// one tone side by side one run.
func tonedRows(rows []string, width int) []line {
	lines := make([]line, 0, len(rows))
	for _, row := range rows {
		shown := line{}
		for column, letter := range []rune(row) {
			tone := toneAt(column, width)
			if last := len(shown) - 1; last >= 0 && shown[last].tone == tone {
				shown[last].text += string(letter)
				continue
			}
			shown = append(shown, run{text: string(letter), kind: kindName, highlighted: false, address: "", copied: "", tone: tone})
		}
		lines = append(lines, shown)
	}
	return lines
}

// textNameWords are the words of the name as text, each letter in the tone its place along the whole name stands at,
// so the colours run on across the words wherever a line breaks them.
func textNameWords(title string) []word {
	letters := []rune(title)
	words, current := []word{}, word{runs: nil, highlightBefore: false}
	for index, letter := range letters {
		if unicode.IsSpace(letter) {
			words, current = appendWord(words, current), word{runs: nil, highlightBefore: false}
			continue
		}
		piece := run{text: string(letter), kind: kindName, highlighted: false, address: "", copied: "", tone: toneAt(index, len(letters))}
		current.runs = append(current.runs, piece)
	}
	return appendWord(words, current)
}

// appendWord is words with one more after them, where that one holds any text.
func appendWord(words []word, next word) []word {
	if len(next.runs) == 0 {
		return words
	}
	return append(words, next)
}

// toneAt is the tone of a column along a width: the first at its start, the last at its end.
func toneAt(column, width int) int {
	return column * (nameTones - 1) / max(width-1, 1)
}

// besideLines are two columns of lines side by side, the second standing a width in from the first's left edge.
func besideLines(left []line, rightStart int, right []line) []line {
	lines := make([]line, 0, max(len(left), len(right)))
	for row := range max(len(left), len(right)) {
		shown := line{}
		if row < len(left) {
			shown = append(shown, left[row]...)
		}
		if row < len(right) {
			shown = slices.Concat(shown, line{gapRun(rightStart - runsWidth(shown))}, right[row])
		}
		lines = append(lines, shown)
	}
	return lines
}

// brief is the brief as neofetch prints a system's: the first way to reach the head's author, the address, parted
// into its mailbox and its host, a rule under it; the subtitle and the place; then the facts and the other ways, each
// by its label; then a row of swatches of the design's colours.
func (layout layout) brief(head document.Head, facts []document.Fact) []line {
	lines, ways := []line{}, head.Ways
	if len(ways) > 0 {
		address := addressRuns(layout.valueWords(ways[0].Value))
		rule := run{text: strings.Repeat(briefRuleMark, runsWidth(address)), kind: kindFaint, highlighted: false, address: "", copied: "", tone: 0}
		lines = append(lines, address, line{rule})
		ways = ways[1:]
	}
	lines = slices.Concat(lines, layout.subtitleLines(head), layout.facts(slices.Concat(facts, ways), 0, kindLabel))
	swatches := line{}
	for tone := range len(swatchColours()) {
		swatches = append(swatches, run{text: swatchMark, kind: kindSwatch, highlighted: false, address: "", copied: "", tone: tone})
	}
	return slices.Concat(lines, blankLines(), []line{swatches})
}

// addressRuns are an address's words as one line, parted at its at: the mailbox, the at faint, the host, each part
// keeping its link.
func addressRuns(words []word) line {
	shown := line{}
	partKind := kindMailbox
	for _, piece := range wordRuns(words) {
		before, after, parted := strings.Cut(piece.text, addressAt)
		shown = append(shown, retold(piece, before, partKind))
		if parted {
			partKind = kindHost
			shown = append(shown, retold(piece, addressAt, kindFaint), retold(piece, after, partKind))
		}
	}
	return shown
}

// retold is a run's piece of text in a kind of its own, its highlight, its link and the value it copies kept.
func retold(piece run, text string, shown kind) run {
	piece.text, piece.kind = text, shown
	return piece
}

// careerSpan is a role on the career's line: its company, its first and last month, whether it is still held, and
// the site's word for what still goes on where it is.
type careerSpan struct {
	company string
	start   time.Time
	end     time.Time
	current bool
	present string
}

// career is the career's line: each role a span on a line of months from the first role's start, or the day where
// every role starts after it, to the day or the last month a role reaches past it, the one still held standing out;
// each company over its span, cut to it; and under them each year's tick and the year, as many as have room. Roles
// with a month still to fill are left out, and the line where the page has not the room for it and the site's word
// for the present.
func (layout layout) career(entries []document.Entry) []line {
	spans := layout.careerSpans(entries)
	present := presentWord(spans)
	if len(spans) == 0 || layout.width < narrowestCareer+presentGap+textWidth(present) {
		return nil
	}
	// the spans stand in the order of their starts, so the first starts the line, unless it starts after the day
	first, last := spans[0].start, lastMonth(spans, layout.today)
	if layout.today.Before(first) {
		first = layout.today
	}
	// the last column is the line's last month, so the columns count one fewer than the room the line takes
	scale := careerScale{first: first, months: max(monthsBetween(first, last), 1), columns: layout.width - presentGap - textWidth(present) - 1}
	rows := make([][]cell, careerRows)
	for row := range rows {
		rows[row] = slices.Repeat([]cell{{letter: space, kind: kindText}}, layout.width)
	}
	lastColumn := scale.column(last)
	put(rows[spanRow], 0, strings.Repeat(careerLineMark, lastColumn+1), kindFaint)
	putSpans(rows, scale, spans)
	put(rows[spanRow], lastColumn+1+presentGap, present, kindCurrent)
	putYears(rows, scale, first, last)
	lines := make([]line, 0, careerRows)
	for _, row := range rows {
		lines = append(lines, cellRuns(row))
	}
	return lines
}

// lastMonth is where the career's line ends: the day, or the last month a role starts or ends in after it.
func lastMonth(spans []careerSpan, today time.Time) time.Time {
	last := today
	for _, span := range spans {
		for _, month := range []time.Time{span.start, span.end} {
			if month.After(last) {
				last = month
			}
		}
	}
	return last
}

// presentWord is the site's word for the present, which the line ends in where a role is still held; none where none
// is.
func presentWord(spans []careerSpan) string {
	for _, span := range spans {
		if span.current {
			return span.present
		}
	}
	return ""
}

// putSpans writes each role into the career's rows: its span, the one still held standing out from those that ended,
// and over it its company, cut to end before the next role's span starts, the last role's before the page's edge.
func putSpans(rows [][]cell, scale careerScale, spans []careerSpan) {
	for index, span := range spans {
		start, next := scale.column(span.start), len(rows[companyRow])+1
		if index+1 < len(spans) {
			next = scale.column(spans[index+1].start)
		}
		spanKind, companyKind := kindPast, kindMuted
		if span.current {
			spanKind, companyKind = kindCurrent, kindCurrent
		}
		put(rows[spanRow], start, strings.Repeat(spanMark, max(scale.column(span.end)-start, 1)), spanKind)
		put(rows[companyRow], start, cutTo(span.company, max(next-start-1, 1)), companyKind)
	}
}

// putYears writes under the career's line a tick where each year after the first month starts, and the year beside
// each tick that leaves room after the year before.
func putYears(rows [][]cell, scale careerScale, first, last time.Time) {
	free := 0
	for year := first.Year() + 1; year <= last.Year(); year++ {
		at := scale.column(time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC))
		put(rows[tickRow], at, yearMark, kindFaint)
		if label := strconv.Itoa(year); at >= free && at+textWidth(label) <= len(rows[yearRow]) {
			put(rows[yearRow], at, label, kindFaint)
			free = at + textWidth(label) + yearLabelGap
		}
	}
}

// careerSpans are the roles of entries as spans, in the order of their starts; a role whose start or end is a
// month still to fill is left out, one that ended runs to the end of its last month, and one still held to the day.
func (layout layout) careerSpans(entries []document.Entry) []careerSpan {
	spans := []careerSpan{}
	for _, entry := range entries {
		start, known := entry.Period.Start.(document.KnownMonth)
		if !known {
			continue
		}
		span := careerSpan{company: document.PlainText(entry.Title), start: start.Start, end: layout.today, current: false, present: ""}
		switch end := entry.Period.End.(type) {
		case document.KnownMonth:
			span.end = end.Start.AddDate(0, 1, 0)
		case document.CurrentMonth:
			span.current, span.present = true, end.Text
		case document.GapMonth:
			continue
		}
		spans = append(spans, span)
	}
	slices.SortFunc(spans, func(first, second careerSpan) int { return first.start.Compare(second.start) })
	return spans
}

// careerScale is how the career's line lays months along columns: from its first month, over how many months, across
// how many columns.
type careerScale struct {
	first   time.Time
	months  int
	columns int
}

// column is the column a month stands at.
func (scale careerScale) column(month time.Time) int {
	return monthsBetween(scale.first, month) * scale.columns / scale.months
}

// monthsBetween is how many months lie from one month to another.
func monthsBetween(from, to time.Time) int {
	return (to.Year()-from.Year())*monthsInYear + int(to.Month()-from.Month())
}

// cutTo is a text cut to a width, its end marked where it was cut and the width leaves room for a letter and the mark.
func cutTo(text string, width int) string {
	letters := []rune(text)
	if len(letters) <= width {
		return text
	}
	if width < shortestCutName {
		return string(letters[:width])
	}
	return string(letters[:width-textWidth(cutMark)]) + cutMark
}

// cell is a cell of a row the career's line draws, a letter and its kind.
type cell struct {
	letter string
	kind   kind
}

// put writes a text into a row of cells from a column, as much of it as the row holds.
func put(row []cell, column int, text string, written kind) {
	for offset, letter := range []rune(text) {
		if at := column + offset; at >= 0 && at < len(row) {
			row[at] = cell{letter: string(letter), kind: written}
		}
	}
}

// cellRuns are a row of cells as a line, each run of cells of one kind one run.
func cellRuns(row []cell) line {
	shown := line{}
	for _, written := range row {
		if last := len(shown) - 1; last >= 0 && shown[last].kind == written.kind {
			shown[last].text += written.letter
			continue
		}
		shown = append(shown, run{text: written.letter, kind: written.kind, highlighted: false, address: "", copied: "", tone: 0})
	}
	return shown
}
