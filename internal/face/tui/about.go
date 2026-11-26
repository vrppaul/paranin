package tui

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/look"
)

// The cursor after the name's last line in the name's font, a column apart from it, as wide as the font's strokes
const blockCursor = "▌"

// The room between the name's column and the brief's, as the storyboard set them apart, and the least the brief
// takes beside the name, where the page has not the room for both, the brief standing under the name; the least room
// the career's line is drawn in, beside the site's word for the present; and how the career's line spaces what it
// draws
const (
	columnsApart    = 4
	narrowestBrief  = 36
	narrowestCareer = 24
	yearLabelGap    = 1
	presentGap      = 1
	shortestCutName = 2
)

// The least width of a small tile under the résumé's career's line, its left edge with it: room for the day set into
// its top edge, 21 Nov 2026 with its inset and a space either side, and a dozen letters of its title, so three stand
// beside the blog's name on a terminal of 80 columns
const leastSmallTile = 18

// The marks of the brief: the rule under the address and a swatch of the design's colours; and of the career's line:
// a role's span, the line the spans stand on, a year's tick, and the end of a company's name cut to its span
const (
	briefRuleMark  = "─"
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

// about lays out the résumé's opening as neofetch prints a system: the name's column, the greeting and the name,
// beside the brief where the page has room for both, above it where not; then the career's line, the lead, and the
// blog's latest posts.
func (layout layout) about(opening briefOpening) []line {
	parts := layout.headParts(opening.head)
	name := layout.nameColumn(slices.Concat(parts.greeting, blankLines()), document.PlainText(opening.head.Title))
	nameWidth := linesWidth(name)
	var lines []line
	if room := layout.width - nameWidth - columnsApart; room >= narrowestBrief {
		// the brief lays out in the column beside the name
		briefColumn := layout
		briefColumn.width = room
		lines = besideLines(name, nameWidth+columnsApart, briefColumn.brief(opening))
	} else {
		lines = stacked(name, layout.brief(opening))
	}
	lines = stacked(stacked(lines, layout.career(opening.career)), parts.lead)
	return stacked(lines, layout.latestPosts(opening.head.Latest))
}

// latestPosts lays out the blog's latest posts under the résumé's career's line (DESIGN.md, The TUI): the blog's name
// bold over the way to its list, then a small tile for each post on a grid of a place for each post the blog names at
// most, beside them where the page has room for each place, under them one under another where not; nothing where no
// post is out, or where the page is narrower than a small tile, as the career's line is left out.
func (layout layout) latestPosts(latest *document.LatestPosts) []line {
	if latest == nil || layout.width < leastSmallTile {
		return nil
	}
	head := slices.Concat(layout.fill(layout.words(latest.Title, kindStrong), 0), layout.fill(layout.words([]document.Span{latest.All}, kindText), 0))
	headWidth := linesWidth(head)
	if room := layout.width - headWidth - nameGap; room >= document.LatestPostsPlaces*leastSmallTile+textWidth(frameDown) {
		return besideLines(head, headWidth+nameGap, layout.smallTiles(latest.Posts, filledGrid(document.LatestPostsPlaces, room)))
	}
	return stacked(head, layout.smallTiles(latest.Posts, filledGrid(oneUnderAnother, layout.width)))
}

// smallTiles lays out posts as small tiles on a grid (DESIGN.md, LatestPosts), framed as the blog's list frames its
// tiles: each its day set into its top edge, as a tile's, and its title alone, bold, on one line, cut where it runs
// longer, leading to its post.
func (layout layout) smallTiles(posts []document.ListedPost, grid tileGrid) []line {
	inside := layout
	inside.width = grid.inside()
	cells := make([]gridCell, 0, len(posts))
	for _, post := range posts {
		// compose hands the latest posts out alone; a part to come would stand off, leading nowhere
		out, isOut := post.Readiness.(document.PostOut)
		title := []document.Span{document.Link{Target: out.Page, Spans: []document.Span{document.Text{Text: cutTo(document.PlainText(post.Title), inside.width)}}}}
		cells = append(cells, gridCell{label: layout.edgeLabel(post), words: tileWords{top: inside.fill(inside.words(title, kindStrong), 0), foot: nil}, leads: isOut})
	}
	return grid.framed(cells).lines
}

// nameColumn is the lines over the author's name, then the name: its first word above the rest in the name's font, the
// cursor after them, where the font draws every letter of it and the lines over it and the banner fit the page, in
// its width and its rows; else as text in the same colours, the site's cursor after it, wrapped where the page is
// narrower than the name.
func (layout layout) nameColumn(linesAbove []line, title string) []line {
	if banner := nameBanner(look.NameLines(title)); len(banner) > 0 {
		if column := slices.Concat(linesAbove, banner); layout.fits(column) {
			return column
		}
	}
	return slices.Concat(linesAbove, layout.fill(withCursor(gradientWords(title, kindName)), 0))
}

// nameBanner is a name's lines in the name's font, the cursor after them; nothing where the name has no lines or the
// font does not draw every letter of them.
func nameBanner(texts []string) []line {
	rows := look.NameFont().Banner(texts)
	if len(rows) == 0 {
		return nil
	}
	// the cursor stands after the last line, whose width it adds to
	last := rows[len(rows)-1]
	lines := tonedBanner(rows, kindName, max(rowsWidth(rows[0]), rowsWidth(last)+textWidth(space+blockCursor)))
	for row := len(lines) - len(last); row < len(lines); row++ {
		lines[row] = append(lines[row], gapRun(spaceWidth), plainRun(blockCursor, kindCursor))
	}
	return lines
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

// brief is the brief as neofetch prints a system's: the address, the first way to reach the author, parted into its
// mailbox and its host, a rule under it; the subtitle, and the place on a line of its own, as the page's brief stands
// them; then the facts, each by its label; the other
// ways to reach the author, as the page's first screen ends its brief with them; then a row of swatches of the
// design's colours.
func (layout layout) brief(opening briefOpening) []line {
	lines := []line{}
	if address := opening.address; len(address) > 0 {
		addressLine := wordRuns(layout.valueWords(address))
		if mail, isMail := document.MailOf(address); isMail {
			addressLine = mailRuns(addressLine, mail)
		}
		rule := plainRun(strings.Repeat(briefRuleMark, runsWidth(addressLine)), kindFaint)
		lines = append(lines, addressLine, line{rule})
	}
	role, place := layout.fill(layout.words(opening.head.Subtitle, kindText), 0), layout.fill(textWords(opening.head.Place, kindMuted), 0)
	lines = stacked(slices.Concat(lines, role, place, layout.facts(opening.brief, 0, kindLabel)), layout.factBlock(opening.ways, 0))
	swatches := line{}
	for tone := range len(swatchColours()) {
		swatches = append(swatches, tonedRun(swatchMark, kindSwatch, tone))
	}
	return slices.Concat(lines, blankLines(), []line{swatches})
}

// mailRuns are a mail as one line, parted as the mail parts: the mailbox, the at faint, the host, each keeping the
// link, highlight and copied value of the run the link's words were laid out in; the address whole where it holds no
// at, as the page draws it.
func mailRuns(laid line, mail document.Mail) line {
	if len(laid) == 0 {
		return laid
	}
	piece := laid[0]
	box, host, parted := mail.Parts()
	if !parted {
		return line{retold(piece, mail.Address, kindMailbox)}
	}
	return line{retold(piece, box, kindMailbox), retold(piece, document.MailAt, kindFaint), retold(piece, host, kindHost)}
}

// retold is a run's piece of text in a kind of its own, its highlight, its link and the value it copies kept.
func retold(piece run, text string, shown kind) run {
	piece.text, piece.kind = text, shown
	return piece
}

// career is the career's line (document.CareerOf, on the session's day): each role a span on a line of months, the
// one still held standing out; each company over its span, cut to it; and under them each year's tick and the year,
// as many as have room. The line is left out where the page has not the room for it and the site's word for the
// present.
func (layout layout) career(entries []document.Entry) []line {
	career := document.CareerOf(entries, layout.today)
	if len(career.Spans) == 0 || layout.width < narrowestCareer+presentGap+textWidth(career.Present) {
		return nil
	}
	// the last column is the line's last month, so the columns count one fewer than the room the line takes
	scale := careerScale{career: career, columns: layout.width - presentGap - textWidth(career.Present) - 1}
	rows := make([][]cell, careerRows)
	for row := range rows {
		rows[row] = slices.Repeat([]cell{{letter: space, kind: kindText}}, layout.width)
	}
	lastColumn := scale.column(career.Last)
	put(rows[spanRow], 0, strings.Repeat(careerLineMark, lastColumn+1), kindFaint)
	putSpans(rows, scale, career.Spans)
	put(rows[spanRow], lastColumn+1+presentGap, career.Present, kindCurrent)
	putYears(rows, scale, career.First, career.Last)
	lines := make([]line, 0, careerRows)
	for _, row := range rows {
		lines = append(lines, cellRuns(row))
	}
	return lines
}

// putSpans writes each role into the career's rows: its span, the one still held standing out from those that ended,
// and over it its company, cut to end before the next role's span starts, the last role's before the page's edge.
func putSpans(rows [][]cell, scale careerScale, spans []document.CareerSpan) {
	for index, span := range spans {
		start, next := scale.column(span.Start), len(rows[companyRow])+1
		if index+1 < len(spans) {
			next = scale.column(spans[index+1].Start)
		}
		spanKind, companyKind := kindPast, kindMuted
		if span.Current {
			spanKind, companyKind = kindCurrent, kindCurrent
		}
		put(rows[spanRow], start, strings.Repeat(spanMark, max(scale.column(span.End)-start, 1)), spanKind)
		put(rows[companyRow], start, cutTo(span.Company, max(next-start-1, 1)), companyKind)
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

// careerScale is how the career's line lays its months along columns: across how many columns.
type careerScale struct {
	career  document.Career
	columns int
}

// column is the column a month stands at.
func (scale careerScale) column(month time.Time) int {
	return scale.career.MonthsTo(month) * scale.columns / scale.career.Months()
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
		shown = append(shown, plainRun(written.letter, written.kind))
	}
	return shown
}
