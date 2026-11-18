package tui

import (
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/vrppaul/paranin/internal/look"
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

// footerRow is a row of the footer under the page.
type footerRow int

const (
	// footerRim is the rim that parts the footer from the page
	footerRim footerRow = iota
	// footerGap is a row of the panel with nothing on it
	footerGap
	// footerPrompt is the prompt, and footerPlaces the places under it that Tab completes to
	footerPrompt
	footerPlaces
	// footerStatus is the status line
	footerStatus
)

// roomyFooter is the footer's rows, in order, on a screen with room for them: the rim, the panel holding the prompt
// and the places with a row of nothing above, between and below the two, and the status line last.
func roomyFooter() []footerRow {
	return []footerRow{footerRim, footerGap, footerPrompt, footerGap, footerPlaces, footerGap, footerStatus}
}

// tightFooter is the footer's rows on a screen too short to spare rows for nothing: the prompt, the places and the
// status line alone.
func tightFooter() []footerRow {
	return []footerRow{footerPrompt, footerPlaces, footerStatus}
}

// A screen roomy enough for the roomy footer keeps a row of nothing above the page and has a few lines of the page
// besides; a row of crosses stands above and below the page where the screen has two rows more to spare for them
const (
	rowsAbovePage = 1
	roomyPageRows = 4
	crossRows     = 2
)

// The scrollbar stands this many columns in from the screen's right edge, where the page leaves it room
const scrollbarInset = 2

// The page's rows faded where more of the tab lies below and the rim says so, counted up from the page's last: the
// last the most, and the one above it the less
const (
	fadedMostFromBottom = 1
	fadedLessFromBottom = 2
)

// The marks of the frame: a corner cross; the scrollbar's track and its thumb; and what parts the rim's count of the
// lines below from the key that scrolls to them
const (
	crossMark     = "+"
	trackMark     = "│"
	thumbMark     = "┃"
	moreSeparator = space + placeMark + space
)

// The words of the ruler, as Neovim's writes them
const (
	wholeShown   = "All"
	startShown   = "Top"
	endShown     = "Bot"
	percentMark  = "%"
	wholePercent = 100
)

// frame is where the page sits on a screen of a size: its left edge and width, the rows of nothing above it and the
// rows it shows, whether corner crosses stand around it, and the footer's rows under it.
type frame struct {
	width   int
	height  int
	left    int
	inner   int
	above   int
	rows    int
	crosses bool
	footer  []footerRow
}

func frameOf(width, height int) frame {
	inner := min(width-pageMargins, widestPage)
	if inner < narrowestPage {
		inner = width
	}
	left := (width - inner) / sides
	footer, above := tightFooter(), 0
	if roomy := roomyFooter(); height >= rowsAbovePage+len(roomy)+roomyPageRows {
		footer, above = roomy, rowsAbovePage
	}
	crosses := above > 0 && left >= crossRoom && height >= above+len(footer)+roomyPageRows+crossRows
	rows := height - above - len(footer)
	if crosses {
		rows -= crossRows
	}
	return frame{width: width, height: height, left: left, inner: inner, above: above, rows: max(rows, 0), crosses: crosses, footer: footer}
}

// rowOf is the row a row of the footer stands on, from the screen's top, and whether the footer has it.
func (frame frame) rowOf(wanted footerRow) (int, bool) {
	index := slices.Index(frame.footer, wanted)
	return frame.height - len(frame.footer) + index, index >= 0
}

// scrollbarColumn is the column the scrollbar stands in, and whether the page leaves it room there.
func (frame frame) scrollbarColumn() (int, bool) {
	column := frame.width - scrollbarInset
	return column, column >= frame.left+frame.inner
}

// Tabs are numbered from one, by the number that chooses one
const firstTabNumber = 1

// hint is a key, or keys, and what they do, as the footer and the window of keys name them.
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
// its crosses, or nothing at all; the rim, the panel's row of nothing, the prompt where nothing is typed or said, and
// the scrollbar's two cells; and the window of keys.
type drawnScreen struct {
	tabs       []drawnTab
	crossRow   string
	blankRow   string
	rim        string
	panelGap   string
	idlePrompt string
	track      string
	thumb      string
	keyWindow  paintedWindow
}

// drawnTab is one tab painted: its page; the status line but for its ruler; and the places listed in normal mode.
type drawnTab struct {
	paintedPage
	status paintedBar
	places string
}

// paintedPage is a tab's lines painted with the search it marks, and marked, unpainted, for the two rows a key fades;
// and the line of each match, in order: a line holding two matches is there twice.
type paintedPage struct {
	lines   []paintedLine
	marked  []markedLine
	matches []int
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

// drawnScreen paints every tab and the frame around it, all but what a key moves: the ruler, the rim's word of the
// lines below, the scrollbar's place, the prompt and the places typed.
func (model Model) drawnScreen() drawnScreen {
	width := model.frame.width
	tabs := make([]drawnTab, 0, len(model.tabs))
	for place := range model.tabs {
		tabs = append(tabs, drawnTab{
			paintedPage: model.paintedPage(place),
			status:      model.paint.cut(model.statusOpening(place), width),
			places:      model.paint.panelRow(model.placePieces(model.placesFor(place, ""), noneChosen, ""), model.frame),
		})
	}
	crossRow := ""
	if model.frame.crosses {
		crossRow = model.paint.crossRow(model.frame)
	}
	idle := append(model.promptOpening(look.PromptMark+space), piece{text: cursorBar, part: partSiteCursor})
	return drawnScreen{
		tabs:       tabs,
		crossRow:   crossRow,
		blankRow:   strings.Repeat(space, width),
		rim:        model.paint.rim(width),
		panelGap:   model.paint.panelRow(nil, model.frame),
		idlePrompt: model.paint.panelRow(idle, model.frame),
		track:      model.paint.parts[partTrack].Render(trackMark),
		thumb:      model.paint.parts[partThumb].Render(thumbMark),
		keyWindow:  model.paint.keyWindow(model.keyNotes()),
	}
}

// paintedPage paints a tab's lines, each match of the search it marks marked.
func (model Model) paintedPage(place int) paintedPage {
	marked := markedLines(model.laid[place], model.markedQuery(place))
	return paintedPage{lines: model.paint.tab(marked, model.frame), marked: marked, matches: matchLines(marked)}
}

// screen draws the whole screen: the page between its crosses, the footer under it, and over the page the window of
// keys where it is open; and the cursor, where the reader types. All of it is painted already but what a key moves.
func (model Model) screen() (string, *tea.Cursor) {
	frame, drawn := model.frame, model.drawn.tabs[model.chosen]
	rows := slices.Repeat([]string{model.drawn.blankRow}, frame.above)
	if frame.crosses {
		rows = append(rows, model.drawn.crossRow)
	}
	rows = append(rows, model.pageRows(drawn)...)
	if frame.crosses {
		rows = append(rows, model.drawn.crossRow)
	}
	prompt, cursor := model.promptLine()
	for _, shown := range frame.footer {
		rows = append(rows, model.footerLine(shown, drawn, prompt))
	}
	rows = rows[max(len(rows)-frame.height, 0):]
	if model.keys.mode == modeKeys {
		rows = model.withKeys(rows)
	}
	return strings.Join(rows, lineEnd), cursor
}

// footerLine is a row of the footer as the chosen tab and the keys pressed leave it, the prompt drawn already.
func (model Model) footerLine(shown footerRow, drawn drawnTab, prompt string) string {
	switch shown {
	case footerRim:
		return model.rimRow()
	case footerGap:
		return model.drawn.panelGap
	case footerPrompt:
		return prompt
	case footerPlaces:
		return model.placesRow(drawn)
	case footerStatus:
		return model.statusline(drawn)
	}
	return model.drawn.blankRow
}

// pageRows are the rows of the page the chosen tab shows at its scroll: its lines, the two above the rim faded where
// more lies below, the rim says so and the terminal shows the colours to fade in, and the scrollbar at the right
// where the tab is taller than the page and there is room; rows of nothing past its end.
func (model Model) pageRows(drawn drawnTab) []string {
	frame, scroll := model.frame, model.scrolls[model.chosen]
	_, rimShown := frame.rowOf(footerRim)
	fading, bar := rimShown && model.linesBelow() > 0 && !model.paint.fewColours, model.scrollbar()
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
			text = model.paint.faded(drawn.marked[scroll+row], frame, model.paint.fadedMost)
		case fading && frame.rows-row == fadedLessFromBottom:
			text = model.paint.faded(drawn.marked[scroll+row], frame, model.paint.fadedLess)
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
// the page leaves it room; the column it stands in; where its thumb stands among the page's rows, from its first row
// to the row after its last; the screen's width it ends each row at; and its two cells painted.
type scrollbar struct {
	shown      bool
	column     int
	thumbStart int
	thumbEnd   int
	width      int
	track      string
	thumb      string
}

// scrollbar is the scrollbar of the chosen tab at its scroll: its thumb as long as the share of the tab the page
// shows, at least a row, and as far down as the page is.
func (model Model) scrollbar() scrollbar {
	rows, lines := model.frame.rows, len(model.drawn.tabs[model.chosen].lines)
	column, room := model.frame.scrollbarColumn()
	bar := scrollbar{shown: false, column: column, thumbStart: 0, thumbEnd: rows, width: model.frame.width, track: model.drawn.track, thumb: model.drawn.thumb}
	if !room || lines <= rows {
		return bar
	}
	length := max(rows*rows/lines, 1)
	bar.shown, bar.thumbStart = true, (rows-length)*model.scrolls[model.chosen]/(lines-rows)
	bar.thumbEnd = bar.thumbStart + length
	return bar
}

// rest is the rest of a page's row after a line of a width: blanks to the screen's edge, the scrollbar's cell among
// them where it shows.
func (bar scrollbar) rest(lineWidth, row int) string {
	if !bar.shown {
		return strings.Repeat(space, max(bar.width-lineWidth, 0))
	}
	cell := bar.track
	if row >= bar.thumbStart && row < bar.thumbEnd {
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

// statusOpening is what the status line opens with on a tab: its title, as the block that opens Neovim's status line
// names the mode, then its tally where it has one and it fits whole beside the title with room left for the ruler.
func (model Model) statusOpening(place int) []piece {
	shown := model.tabs[place]
	title := []piece{{text: space + shown.title + space, part: partTitle}}
	tally := []piece{{text: space + shown.tally, part: partTally}}
	if shown.tally == "" || piecesWidth(title)+piecesWidth(tally)+widestRuler > model.frame.width {
		return title
	}
	return append(title, tally...)
}

// The widest the ruler ever is with the spaces around it: one of its words, or a percentage below a hundred, which
// is two digits at most
const (
	percentDigits = 2
	widestRuler   = 2*spaceWidth + max(len(wholeShown), len(startShown), len(endShown), percentDigits+len(percentMark))
)

// statusline is the status line of a tab: what it opens with, a fill, and the ruler at the right.
func (model Model) statusline(drawn drawnTab) string {
	width := model.frame.width
	ruler := model.paint.cut([]piece{{text: space + model.position() + space, part: partRuler}}, width-drawn.status.width)
	fill := strings.Repeat(space, max(width-drawn.status.width-ruler.width, 0))
	return drawn.status.text + model.paint.parts[partStatus].Render(fill) + ruler.text
}

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
