package tui

import (
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// How the page sits on the screen: no wider than a text is read at, standing in from each side by at least
// this much; narrower than the least, the page takes the whole width and the crosses go
const (
	widestPage    = 100
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

func frameOf(width, height int) frame {
	inner := min(width-pageMargins, widestPage)
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
// its crosses, or nothing at all; the scrollbar's two cells; what the footer holds on any tab; and the window of keys.
type drawnScreen struct {
	tabs      []drawnTab
	crossRow  string
	blankRow  string
	track     string
	thumb     string
	footer    paintedFooter
	keyWindow paintedWindow
}

// drawnTab is one tab painted: its page, and the footer's session and tabs while it is shown.
type drawnTab struct {
	paintedPage
	footerTabs paintedBar
}

// paintedPage is a tab's lines painted with the search it marks, and marked, unpainted, for the two rows a key fades;
// the line of each match, in order: a line holding two matches is there twice; and the match its count names.
type paintedPage struct {
	lines   []paintedLine
	marked  []markedLine
	matches []int
	named   namedMatch
}

// paintedLine is a line of a page painted from the screen's left edge to its own end, and the columns it takes.
type paintedLine struct {
	text  string
	width int
}

// paintedBar is pieces of a bar painted, and the columns they take.
type paintedBar struct {
	text  string
	width int
}

// drawnScreen paints every tab and the frame around it, all but what a key moves: the ruler, the scrollbar's place
// and the command line.
func (model Model) drawnScreen() drawnScreen {
	footer := model.paintedFooter()
	tabs := make([]drawnTab, 0, len(model.tabs))
	for place := range model.tabs {
		tabs = append(tabs, drawnTab{paintedPage: model.paintedPage(place), footerTabs: model.footerTabs(place, footer.place)})
	}
	crossRow := ""
	if model.frame.crosses {
		crossRow = model.paint.crossRow(model.frame)
	}
	return drawnScreen{
		tabs:      tabs,
		crossRow:  crossRow,
		blankRow:  strings.Repeat(space, model.frame.width),
		track:     model.paint.parts[partTrack].Render(trackMark),
		thumb:     model.paint.parts[partThumb].Render(thumbMark),
		footer:    footer,
		keyWindow: model.paint.keyWindow(model.keyNotes()),
	}
}

// paintedPage paints a tab's lines, each match of the search it marks marked, the one its count names apart.
func (model Model) paintedPage(place int) paintedPage {
	marked := markedLines(model.laid[place], model.markedQuery(place))
	matches := matchLines(marked)
	named := model.namedMatchOf(place, matches)
	return paintedPage{lines: model.paint.tab(marked, model.frame, named), marked: marked, matches: matches, named: named}
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
// lies below, the page shows a few lines more than those and the terminal shows the colours to fade in, and the
// scrollbar at the right where the tab is taller than the page and there is room; rows of nothing past its end.
func (model Model) pageRows(drawn drawnTab) []string {
	frame, scroll := model.frame, model.scrolls[model.chosen]
	fading, bar := frame.rows >= roomyPageRows && model.linesBelow() > 0 && !model.paint.fewColours, model.scrollbar()
	rows := make([]string, 0, frame.rows)
	for row := range frame.rows {
		if scroll+row >= len(drawn.lines) {
			rows = append(rows, model.drawn.blankRow)
			continue
		}
		shown := drawn.lines[scroll+row]
		text := shown.text
		switch {
		case fading && frame.rows-row == fadedMostFromBottom:
			text = model.paint.faded(drawn.marked[scroll+row], frame, model.paint.fadedMost, drawn.named.occurrenceOn(scroll+row))
		case fading && frame.rows-row == fadedLessFromBottom:
			text = model.paint.faded(drawn.marked[scroll+row], frame, model.paint.fadedLess, drawn.named.occurrenceOn(scroll+row))
		}
		rows = append(rows, text+bar.rest(shown.width, row))
	}
	return rows
}

// linesBelow is how many lines of the chosen tab lie below the page.
func (model Model) linesBelow() int {
	return max(len(model.drawn.tabs[model.chosen].lines)-model.scrolls[model.chosen]-model.frame.rows, 0)
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
	bar.shown, bar.thumbRows = true, thumbOf(rows, lines, model.scrolls[model.chosen])
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
	scroll := model.scrolls[model.chosen]
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
