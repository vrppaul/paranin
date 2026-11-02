package tui

import (
	"strconv"
	"strings"
	"unicode/utf8"

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

// The rows the frame takes: the line of tabs above the page, the status line and, last, the command line below it,
// and a row of crosses above and below the page; a screen with fewer rows than the crosses need has none
const (
	rowsAbovePage   = 1
	statusLineRows  = 1
	commandLineRows = 1
	rowsBelowPage   = statusLineRows + commandLineRows
	barRows         = rowsAbovePage + rowsBelowPage
	crossRows       = 2
	fewestForCross  = 6
)

// The columns the terminal's cursor takes on the command line, after what is typed
const cursorWidth = 1

// The marks of the frame: a corner cross, the rule between the parts of the status line, and the room a tab's
// number and title are padded with, most first
const (
	crossMark       = "+"
	statusRuleMark  = "│"
	widestTabPad    = 2
	narrowestTabPad = 1
)

// The words of the status line, as Neovim's ruler and its modes write them
const (
	normalMode   = "NORMAL"
	commandMode  = "COMMAND"
	wholeShown   = "All"
	startShown   = "Top"
	endShown     = "Bot"
	percentMark  = "%"
	wholePercent = 100
)

// frame is where the page sits on a screen of a size: its left edge and width, the rows it shows, and whether
// corner crosses stand around it.
type frame struct {
	width   int
	height  int
	left    int
	inner   int
	rows    int
	crosses bool
}

func frameOf(width, height int) frame {
	inner := min(width-pageMargins, widestPage)
	if inner < narrowestPage {
		inner = width
	}
	left := (width - inner) / sides
	crosses := left >= crossRoom && height >= fewestForCross
	rows := height - barRows
	if crosses {
		rows -= crossRows
	}
	return frame{width: width, height: height, left: left, inner: inner, rows: max(rows, 0), crosses: crosses}
}

// Tabs are numbered from one, in the line of tabs and by the number that chooses one
const firstTabNumber = 1

// hint is a key, or keys, and what they do, as the status line and the window of keys name them.
type hint struct {
	key  string
	does string
}

// What parts the two keys of a hint, and the widest the ruler ever is with the spaces around it: one of its
// words, or a percentage below a hundred, which is two digits at most
const (
	hintKeySeparator = "/"
	percentDigits    = 2
	widestRuler      = 2*spaceWidth + max(len(wholeShown), len(startShown), len(endShown), percentDigits+len(percentMark))
)

// The mark the status line names the Enter key by, the word for what both it and the slash do, and for what Esc
// does on the command line
const (
	enterMark  = "↵"
	findHint   = "find"
	cancelHint = "cancel"
)

// modeLook is how the status line shows a mode: the name it opens with, and the keys it names, the most needed
// first, as many as fit.
type modeLook struct {
	name  string
	hints []hint
}

// modeLooks are how the status line shows each mode, saying what q does on the page shown.
func modeLooks(closing string) map[mode]modeLook {
	return map[mode]modeLook{
		modeNormal: {name: normalMode, hints: []hint{
			{key: keyWindowKey, does: keysTitle},
			{key: previousKey + hintKeySeparator + nextKey, does: "tabs"},
			{key: downKey + hintKeySeparator + upKey, does: "scroll"},
			{key: searchKey, does: findHint},
			{key: quitKey, does: closing},
		}},
		modeSearch: {name: commandMode, hints: []hint{
			{key: enterMark, does: findHint},
			{key: escapeKey, does: cancelHint},
		}},
		modeCommand: {name: commandMode, hints: []hint{
			{key: enterMark, does: "run"},
			{key: tabKey, does: "complete"},
			{key: escapeKey, does: cancelHint},
		}},
		modeKeys: {name: normalMode, hints: []hint{
			{key: escapeKey, does: closeHint},
		}},
	}
}

// piece is a piece of the frame: its text and the part of the frame it is.
type piece struct {
	text string
	part part
}

// drawnScreen is the screen painted for the terminal as it is: each tab, the rows of the page that hold nothing but
// its crosses, or nothing at all, and the window of keys.
type drawnScreen struct {
	tabs      []drawnTab
	crossRow  string
	blankRow  string
	keyWindow paintedWindow
}

// drawnTab is one tab painted: its page, the line of tabs with it chosen, and the status line in each mode but for
// its ruler.
type drawnTab struct {
	paintedPage
	tabline string
	status  map[mode]drawnStatus
}

// paintedPage is a tab's lines painted with the search it marks, and the line of each match, in order: a line
// holding two matches is there twice.
type paintedPage struct {
	lines   []string
	matches []int
}

// drawnStatus is the status line but for its ruler: what stands before the ruler, and the keys after it.
type drawnStatus struct {
	opening paintedBar
	hints   paintedBar
}

// paintedBar is pieces of a bar painted, and the columns they take.
type paintedBar struct {
	text  string
	width int
}

// drawnScreen paints every tab and the frame around it, all but the ruler, which a key moves.
func (model Model) drawnScreen() drawnScreen {
	width := model.frame.width
	tabs := make([]drawnTab, 0, len(model.tabs))
	looks := modeLooks(closingHint(model.shown))
	for place := range model.tabs {
		status := make(map[mode]drawnStatus, len(looks))
		for shown, seen := range looks {
			status[shown] = model.drawnStatus(place, seen)
		}
		tabs = append(tabs, drawnTab{
			paintedPage: model.paintedPage(place),
			tabline:     model.paint.bar(model.tabline(place), width),
			status:      status,
		})
	}
	crossRow := ""
	if model.frame.crosses {
		crossRow = model.paint.crossRow(model.frame)
	}
	return drawnScreen{tabs: tabs, crossRow: crossRow, blankRow: strings.Repeat(space, width), keyWindow: model.paint.keyWindow(model.keyNotes())}
}

// paintedPage paints a tab's lines, each match of the search it marks marked.
func (model Model) paintedPage(place int) paintedPage {
	marked := markedLines(model.laid[place], model.markedQuery(place))
	return paintedPage{lines: model.paint.tab(marked, model.frame), matches: matchLines(marked)}
}

// drawnStatus paints a tab's status line in a mode, all but its ruler.
func (model Model) drawnStatus(place int, seen modeLook) drawnStatus {
	width := model.frame.width
	opening := model.paint.cut(model.statusOpening(place, seen.name), width)
	return drawnStatus{opening: opening, hints: model.paint.cut(statusHints(seen.hints, width-opening.width-widestRuler), width)}
}

// screen draws the whole screen: the line of tabs, the page between its crosses, the status line and the command
// line; and the cursor, where the reader types. All of it is painted already but the ruler and the command line.
func (model Model) screen() (string, *tea.Cursor) {
	frame, drawn := model.frame, model.drawn.tabs[model.chosen]
	rows := []string{drawn.tabline}
	if frame.crosses {
		rows = append(rows, model.drawn.crossRow)
	}
	scroll := model.scrolls[model.chosen]
	for row := range frame.rows {
		if scroll+row < len(drawn.lines) {
			rows = append(rows, drawn.lines[scroll+row])
			continue
		}
		rows = append(rows, model.drawn.blankRow)
	}
	if frame.crosses {
		rows = append(rows, model.drawn.crossRow)
	}
	commandLine, cursor := model.commandLine()
	rows = append(rows, model.statusline(drawn), commandLine)
	rows = rows[max(len(rows)-frame.height, 0):]
	switch model.keys.mode {
	case modeCommand:
		rows = model.withCompletions(rows)
	case modeKeys:
		rows = model.withKeys(rows)
	case modeNormal, modeSearch:
	}
	return strings.Join(rows, lineEnd), cursor
}

// commandLine is the last row: what the reader types after the slash or the colon, with the cursor after it, or
// what the last key said, with no cursor.
func (model Model) commandLine() (string, *tea.Cursor) {
	switch model.keys.mode {
	case modeSearch:
		return model.typedLine(searchMark)
	case modeCommand:
		return model.typedLine(commandMark)
	case modeNormal, modeKeys:
	}
	said := partCommand
	if model.said.failed {
		said = partFailure
	}
	return model.paint.plainRow([]piece{{text: space, part: partCommand}, {text: model.said.text, part: said}}, model.frame.width), nil
}

// typedLine is the command line as the reader types on it after a mark, and the terminal's own cursor after what
// they typed, held within the screen: a bar the terminal blinks, as the first version's cursor blinked, in the
// accent where the terminal shows many colours.
func (model Model) typedLine(mark string) (string, *tea.Cursor) {
	prompt := space + mark
	shown := model.shownTyping(prompt)
	row := model.paint.plainRow([]piece{{text: prompt, part: partPrompt}, {text: shown, part: partCommand}}, model.frame.width)
	position := tea.Position{X: min(textWidth(prompt+shown), model.frame.width-cursorWidth), Y: model.frame.height - commandLineRows}
	return row, &tea.Cursor{Position: position, Color: model.paint.cursorColour, Shape: tea.CursorBar, Blink: true}
}

// shownTyping is what the command line shows of what is typed after a prompt: its end, where the line is too
// narrow for all of it and the cursor after it.
func (model Model) shownTyping(prompt string) string {
	room := model.frame.width - textWidth(prompt) - cursorWidth
	typed := model.keys.typing
	for textWidth(typed) > max(room, 0) {
		_, first := utf8.DecodeRuneInString(typed)
		typed = typed[first:]
	}
	return typed
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

// tabline is each tab by its number and title, a chosen one on the page's ground, then the site's name at the
// right. Where the tabs would not fit, they are padded less, and then the first ones go until the chosen one fits.
func (model Model) tabline(chosen int) []piece {
	width := model.frame.width
	pieces := model.tabPieces(widestTabPad, 0, chosen)
	for first := 0; piecesWidth(pieces) > width && first <= chosen; first++ {
		pieces = model.tabPieces(narrowestTabPad, first, chosen)
	}
	host := space + model.site.Host() + space
	if room := width - piecesWidth(pieces) - textWidth(host); room >= 0 {
		pieces = append(pieces, piece{text: strings.Repeat(space, room), part: partBar}, piece{text: host, part: partHost})
	}
	return pieces
}

// tabPieces are the tabs from a first one on, each padded by as much, one of them chosen.
func (model Model) tabPieces(padding, first, chosen int) []piece {
	pad := strings.Repeat(space, padding)
	pieces := []piece{}
	for place, shown := range model.tabs[first:] {
		number, title := partTabNumber, partTabTitle
		if first+place == chosen {
			number, title = partChosenNumber, partChosenTitle
		}
		pieces = append(pieces, piece{text: pad + strconv.Itoa(first+place+firstTabNumber) + space, part: number}, piece{text: shown.title + pad, part: title})
	}
	return pieces
}

// statusOpening is what the status line opens with on a tab in a mode: the mode, the tab's title, and the rule
// before the ruler.
func (model Model) statusOpening(place int, modeName string) []piece {
	return []piece{
		{text: space + modeName + space, part: partMode},
		{text: space + model.tabs[place].title + space, part: partStatus},
		{text: statusRuleMark, part: partStatusRule},
	}
}

// statusHints are the keys the status line names in a mode that fit in a room, then the bar's last space; in a
// room of none, nothing.
func statusHints(named []hint, room int) []piece {
	if room < spaceWidth {
		return nil
	}
	shown := []piece{}
	for _, named := range named {
		next := append(append([]piece{}, shown...),
			piece{text: space + space, part: partStatus}, piece{text: named.key, part: partHintKey}, piece{text: space + named.does, part: partStatus})
		if piecesWidth(next)+spaceWidth > room {
			break
		}
		shown = next
	}
	return append(shown, piece{text: space, part: partStatus})
}

// statusline is the status line of a tab: what it opens with, the ruler, and the keys, a fill between them.
func (model Model) statusline(drawn drawnTab) string {
	width, status := model.frame.width, drawn.status[model.keys.mode]
	ruler := model.paint.cut([]piece{{text: space + model.position() + space, part: partStatus}}, width-status.opening.width)
	// the keys were chosen to fit beside the widest ruler, so they fit beside this one
	fill := strings.Repeat(space, max(width-status.opening.width-ruler.width-status.hints.width, 0))
	return status.opening.text + ruler.text + model.paint.parts[partBar].Render(fill) + status.hints.text
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

// bar paints pieces of the frame as one row of a width, cut at its edge and filled with the bar's ground.
func (paint paint) bar(pieces []piece, width int) string {
	painted := paint.cut(pieces, width)
	return painted.text + paint.parts[partBar].Render(strings.Repeat(space, max(width-painted.width, 0)))
}

// plainRow paints pieces of the frame as one row of a width on the page's own ground, cut at its edge.
func (paint paint) plainRow(pieces []piece, width int) string {
	painted := paint.cut(pieces, width)
	return painted.text + strings.Repeat(space, max(width-painted.width, 0))
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
