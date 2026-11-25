package tui

import (
	"image/color"
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// How wide the page stands at most: the site's at a text's measure, the blog's as wide as a post's columns, its
// Outline, its paragraphs' numbers, its text and its sources beside it (DESIGN.md, The TUI)
const (
	widestPage     = 100
	widestBlogPage = outlineFrom
)

// How the page sits on the screen: no wider than the session shown has it (widestPageOf), standing in from each
// side by at least this much; narrower than the least, the page takes the whole width and the crosses go
const (
	pageMargins   = 8
	narrowestPage = 24
	crossRoom     = 2
	sides         = 2
)

// The footer, one line at the screen's foot, under the page always; and over it the row the command line stands on,
// which the page never takes, so what the command line says never hides a line of the page
const (
	footerRows      = 1
	commandLineRows = 1
)

// A screen with room for a few lines of the page keeps a row of nothing above it, and fades the page's last two rows
// where more lies below; a row of crosses stands above the page where the screen has a row more for it, and the
// command line's row holds the crosses under the page
const (
	rowsAbovePage = 1
	roomyPageRows = 4
	topCrossRows  = 1
)

// The scrollbar stands this many columns in from the screen's right edge, where the page leaves it room
const scrollbarInset = 2

// The page's rows faded where more of the tab lies below, counted up from the page's last: the last the most, and the
// one above it the less
const (
	fadedMostFromBottom = 1
	fadedLessFromBottom = 2
)

// The marks of the frame: a corner cross; and the scrollbar's track and its thumb
const (
	crossMark = "+"
	trackMark = "│"
	thumbMark = "┃"
)

// The words of the ruler, as Neovim's writes them
const (
	wholeShown   = "All"
	startShown   = "Top"
	endShown     = "Bot"
	percentMark  = "%"
	wholePercent = 100
)

// frame is where the page sits on a screen of a size: its left edge and width, the rows of nothing above it, the rows
// it shows and the command line's row under them, which a screen of one row has not; and whether corner crosses stand
// around it.
type frame struct {
	width   int
	height  int
	left    int
	inner   int
	above   int
	rows    int
	under   int
	crosses bool
}

// widestPageOf is how wide the page stands at most in the session shown: the blog's wider than the site's.
func widestPageOf(inBlog bool) int {
	if inBlog {
		return widestBlogPage
	}
	return widestPage
}

func frameOf(width, height, widest int) frame {
	inner := min(width-pageMargins, widest)
	if inner < narrowestPage {
		inner = width
	}
	left := (width - inner) / sides
	under := min(commandLineRows, max(height-footerRows, 0))
	above := 0
	if height >= rowsAbovePage+roomyPageRows+commandLineRows+footerRows {
		above = rowsAbovePage
	}
	crosses := above > 0 && left >= crossRoom && height >= above+topCrossRows+roomyPageRows+commandLineRows+footerRows
	rows := height - above - under - footerRows
	if crosses {
		rows -= topCrossRows
	}
	return frame{width: width, height: height, left: left, inner: inner, above: above, rows: max(rows, 0), under: under, crosses: crosses}
}

// pageTop is the row the page's first row stands on, from the screen's top: under the rows of nothing and the crosses.
func (frame frame) pageTop() int {
	if frame.crosses {
		return frame.above + topCrossRows
	}
	return frame.above
}

// footerAt is the row the footer stands on, from the screen's top: the last.
func (frame frame) footerAt() int {
	return frame.height - footerRows
}

// commandRow is the row the command line stands on, over the footer, from the screen's top, and whether the screen
// has it.
func (frame frame) commandRow() (int, bool) {
	return frame.height - footerRows - frame.under, frame.under > 0
}

// scrollbarColumn is the column the scrollbar stands in, and whether the page leaves it room there.
func (frame frame) scrollbarColumn() (int, bool) {
	column := frame.width - scrollbarInset
	return column, column >= frame.left+frame.inner
}

// Tabs are numbered from one, by the number that chooses one
const firstTabNumber = 1

// hint is a key, or keys, and what they do, as the command line and the window of keys name them.
type hint struct {
	key  string
	does string
}

// piece is a piece of the frame: its text and the part of the frame it is.
type piece struct {
	text string
	part part
}

// drawnScreen is the screen painted for the terminal as it is: each tab; the rows of the page that hold nothing but
// its crosses, or nothing at all; the scrollbar's two cells; and what the footer holds on any tab.
type drawnScreen struct {
	tabs     []drawnTab
	crossRow string
	blankRow string
	track    string
	thumb    string
	footer   paintedFooter
}

// drawnTab is one tab painted: its page; what stays in sight over its lines' left part; and while it is shown the
// footer's sessions and tabs, where a click on each part of the footer goes, and the window of keys.
type drawnTab struct {
	paintedPage
	stuck       stuckPaint
	footerTabs  paintedBar
	footerSpots []footerSpot
	keyWindow   paintedWindow
}

// paintedTab is a tab's page painted, and what stays in sight over it.
type paintedTab struct {
	page  paintedPage
	stuck stuckPaint
}

// paintedPage is a tab's lines painted with the search it marks, and marked, unpainted, for the two rows a key fades;
// the line of each match, in order: a line holding two matches is there twice; and what sets runs of its lines apart,
// the match its count names and the link the keys stand on.
type paintedPage struct {
	lines   []paintedLine
	marked  []markedLine
	matches []int
	marks   pageMarks
}

// stuckPaint is what stays in sight over a tab's lines' left part, painted alone, which hangs on its layout, the
// paint and the frame and never on the search: the names of the blog's list's groups, and a post's Outline.
type stuckPaint struct {
	groups  []paintedGroup
	outline paintedOutline
}

// paintedLine is a line of a page painted from the screen's left edge to its own end: its part up to the column what
// stays in sight over it may end at, the margin with it, and its part after that column, painted apart, which what
// stays in sight stands beside; a line nothing may stand over is its first part whole.
type paintedLine struct {
	left  paintedBar
	right paintedBar
}

// paintedBar is pieces of a bar painted, and the columns they take.
type paintedBar struct {
	text  string
	width int
}

// keysPainted is a window of keys painted, and the keys it names.
type keysPainted struct {
	notes  []hint
	window paintedWindow
}

// drawnScreen paints the frame around the tabs' pages, painted already, all but what a key moves: the ruler, the
// scrollbar's place and the command line.
func (model Model) drawnScreen(painted []paintedTab) drawnScreen {
	footer := model.paintedFooter()
	tabs, windows := make([]drawnTab, 0, len(painted)), []keysPainted{}
	for place, page := range painted {
		notes := model.keyNotes(place)
		// tabs that name the same keys share one window of them, painted once
		found := slices.IndexFunc(windows, func(painted keysPainted) bool { return slices.Equal(painted.notes, notes) })
		if found < 0 {
			found, windows = len(windows), append(windows, keysPainted{notes: notes, window: model.paint.keyWindow(notes)})
		}
		footerTabs := model.footerTabs(place, footer.place)
		tabs = append(tabs, drawnTab{
			paintedPage: page.page, stuck: page.stuck, footerTabs: footerTabs.painted, footerSpots: slices.Concat(footerTabs.spots, footer.keys.spots),
			keyWindow: windows[found].window,
		})
	}
	crossRow := ""
	if model.frame.crosses {
		crossRow = model.paint.crossRow(model.frame)
	}
	return drawnScreen{
		tabs:     tabs,
		crossRow: crossRow,
		blankRow: strings.Repeat(space, model.frame.width),
		track:    model.paint.parts[partTrack].Render(trackMark),
		thumb:    model.paint.parts[partThumb].Render(thumbMark),
		footer:   footer,
	}
}

// paintedTab paints a tab's page and what stays in sight over it.
func (model Model) paintedTab(place int) paintedTab {
	return paintedTab{page: model.paintedPage(place), stuck: model.stuckPaint(place)}
}

// paintedPage paints a tab's lines, each match of the search it marks marked, the one its count names apart, and on
// the blog's list the link the keys stand on.
func (model Model) paintedPage(place int) paintedPage {
	laid := model.laid[place]
	marked := markedLines(laid.lines, model.markedQuery(place))
	matches := matchLines(marked)
	marks := pageMarks{named: model.namedMatchOf(place, matches), pointed: model.pointedLink(place), split: laid.split, overLines: laid.overLines()}
	return paintedPage{lines: model.paint.tab(marked, model.frame, marks), marked: marked, matches: matches, marks: marks}
}

// stuckPaint paints what stays in sight over a tab's lines' left part: the names of the blog's list's groups and a
// post's Outline.
func (model Model) stuckPaint(place int) stuckPaint {
	laid := model.laid[place]
	return stuckPaint{groups: model.paint.groupNames(laid.groups, laid.split, model.frame), outline: model.paint.outline(laid.outline, laid.split, model.frame)}
}

// screen draws the whole screen: the page between its crosses, the footer under it, the command line over the footer
// where it shows, and over the page its menu, or the window of keys where it is open; and the cursor, where the reader
// types. All of it is painted already but what a key moves.
func (model Model) screen() (string, *tea.Cursor) {
	frame, drawn := model.frame, model.drawn.tabs[model.chosen]
	rows := slices.Repeat([]string{model.drawn.blankRow}, frame.above)
	if frame.crosses {
		rows = append(rows, model.drawn.crossRow)
	}
	rows = append(rows, model.pageRows(drawn)...)
	under := model.drawn.blankRow
	if frame.crosses {
		under = model.drawn.crossRow
	}
	rows = append(rows, slices.Repeat([]string{under}, frame.under)...)
	rows = append(rows, model.footerRow(drawn))
	command := model.commandLine()
	if row, shown := frame.commandRow(); shown && command.row != "" {
		rows[row] = command.row
	}
	switch model.keys.mode {
	case modeCommand:
		rows = model.withMenu(rows, command)
	case modeKeys:
		rows = model.withKeys(rows)
	case modeNormal, modeSearch:
	}
	return strings.Join(rows, lineEnd), command.cursor
}

// pageRows are the rows of the page the chosen tab shows at its scroll: its lines, the last two faded where more
// lies below, the page shows a few lines more than those and the terminal shows the colours to fade in; what stays in
// sight, a group's name on the blog's list or a post's Outline, painted already, beside the rest of its rows; and the
// scrollbar at the right where the tab is taller than the page and there is room; rows of nothing past its end.
func (model Model) pageRows(drawn drawnTab) []string {
	frame, scroll := model.frame, model.positions[model.chosen].scroll
	fading, bar := frame.rows >= roomyPageRows && model.linesBelow() > 0 && !model.paint.fewColours, model.scrollbar()
	stuck := stuckRows(drawn.stuck, readerView{scroll: scroll, rows: frame.rows, lines: len(drawn.lines), goneTo: model.goneToLine()})
	rows := make([]string, 0, frame.rows)
	for row := range frame.rows {
		if scroll+row >= len(drawn.lines) {
			rows = append(rows, model.drawn.blankRow)
			continue
		}
		shown, faded := drawn.lines[scroll+row], model.fadedAt(row, fading)
		if faded != nil {
			shown = model.paint.pageLine(drawn.marked[scroll+row], frame, faded, drawn.marks.on(scroll+row))
		}
		left := model.leftOf(shown, stuck[row], faded, drawn.marks.split)
		rows = append(rows, left.text+shown.right.text+bar.rest(left.width+shown.right.width, row))
	}
	return rows
}

// leftOf is a row's part up to the column what stays in sight ends at, a split of the page's lines: the line's own,
// or what stays in sight over it where anything does, faded as the row fades.
func (model Model) leftOf(shown paintedLine, over stuckLine, faded color.Color, split int) paintedBar {
	switch {
	case over.painted == "":
		return shown.left
	case faded != nil:
		return paintedBar{text: model.paint.stuckRow(over.marked, over.hang, split, model.frame, faded), width: model.frame.left + split}
	}
	return paintedBar{text: over.painted, width: model.frame.left + split}
}

// fadedAt is the colour a row of the page fades its text to where the page fades its last two, none on any other.
func (model Model) fadedAt(row int, fading bool) color.Color {
	switch {
	case fading && model.frame.rows-row == fadedMostFromBottom:
		return model.paint.fadedMost
	case fading && model.frame.rows-row == fadedLessFromBottom:
		return model.paint.fadedLess
	}
	return nil
}

// linesBelow is how many lines of the chosen tab lie below the page.
func (model Model) linesBelow() int {
	return max(len(model.drawn.tabs[model.chosen].lines)-model.positions[model.chosen].scroll-model.frame.rows, 0)
}

// scrollbar is the scrollbar as the chosen tab stands: whether it shows, where the tab is taller than the page and
// the page leaves it room; the column it stands in; the rows its thumb takes; the screen's width it ends each row at;
// and its two cells painted.
type scrollbar struct {
	shown     bool
	column    int
	thumbRows thumbSpan
	width     int
	track     string
	thumb     string
}

// scrollbar is the scrollbar of the chosen tab at its scroll.
func (model Model) scrollbar() scrollbar {
	rows, lines := model.frame.rows, len(model.drawn.tabs[model.chosen].lines)
	column, room := model.frame.scrollbarColumn()
	bar := scrollbar{shown: false, column: column, thumbRows: thumbSpan{start: 0, end: rows}, width: model.frame.width, track: model.drawn.track, thumb: model.drawn.thumb}
	if !room || lines <= rows {
		return bar
	}
	bar.shown, bar.thumbRows = true, thumbOf(rows, lines, model.positions[model.chosen].scroll)
	return bar
}

// thumbSpan is the rows a scrollbar's thumb takes: from its first to the row after its last.
type thumbSpan struct {
	start int
	end   int
}

// thumbOf is the rows a scrollbar's thumb takes where rows show more lines, scrolled down so many: as many as the
// share of the lines the rows show, at least one, and as far down as the scroll is.
func thumbOf(rows, lines, scroll int) thumbSpan {
	length := max(rows*rows/lines, 1)
	start := (rows - length) * scroll / (lines - rows)
	return thumbSpan{start: start, end: start + length}
}

// covers is whether the thumb takes a row.
func (span thumbSpan) covers(row int) bool {
	return row >= span.start && row < span.end
}

// rest is the rest of a page's row after a line of a width: blanks to the screen's edge, the scrollbar's cell among
// them where it shows.
func (bar scrollbar) rest(lineWidth, row int) string {
	if !bar.shown {
		return strings.Repeat(space, max(bar.width-lineWidth, 0))
	}
	cell := bar.track
	if bar.thumbRows.covers(row) {
		cell = bar.thumb
	}
	return strings.Repeat(space, bar.column-lineWidth) + cell + strings.Repeat(space, bar.width-bar.column-textWidth(trackMark))
}

// markedLines are a tab's lines with each match of a query marked, worked out once for both the painting and the
// list of the matches.
func markedLines(lines []line, query string) []markedLine {
	marked := make([]markedLine, 0, len(lines))
	for _, shown := range lines {
		marked = append(marked, markedLineOf(shown, query))
	}
	return marked
}

// matchLines are the place of the line of each match among marked lines, in order.
func matchLines(marked []markedLine) []int {
	lines := []int{}
	for place, shown := range marked {
		for range shown.matchCount {
			lines = append(lines, place)
		}
	}
	return lines
}

// The widest the ruler ever is: one of its words, or a percentage below a hundred, which is two digits at most
const (
	percentDigits = 2
	widestRuler   = max(len(wholeShown), len(startShown), len(endShown), percentDigits+len(percentMark))
)

// position says how far down the tab the page shows, as Neovim's ruler does: all of it, its top, its bottom,
// or how far between them.
func (model Model) position() string {
	last := len(model.drawn.tabs[model.chosen].lines) - model.frame.rows
	scroll := model.positions[model.chosen].scroll
	switch {
	case last <= 0:
		return wholeShown
	case scroll <= 0:
		return startShown
	case scroll >= last:
		return endShown
	}
	return strconv.Itoa(scroll*wholePercent/last) + percentMark
}

// cut paints pieces of the frame cut at a width, and says how many columns they take.
func (paint paint) cut(pieces []piece, width int) paintedBar {
	var builder strings.Builder
	used := 0
	for _, shown := range pieces {
		text := ansi.Truncate(shown.text, width-used, "")
		if text == "" {
			continue
		}
		builder.WriteString(paint.parts[shown.part].Render(text))
		used += textWidth(text)
	}
	return paintedBar{text: builder.String(), width: used}
}

// crossRow is a row of nothing with a cross at each corner of the page.
func (paint paint) crossRow(frame frame) string {
	cross := paint.parts[partCross].Render(crossMark)
	between := strings.Repeat(space, frame.inner+crossRoom)
	row := strings.Repeat(space, frame.left-crossRoom) + cross + between + cross
	return row + strings.Repeat(space, max(frame.width-lipgloss.Width(row), 0))
}

func piecesWidth(pieces []piece) int {
	width := 0
	for _, shown := range pieces {
		width += textWidth(shown.text)
	}
	return width
}
