package tui

import (
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

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

// The rows the frame takes: the line of tabs, the status line and the command line, and a row of crosses above
// and below the page; a screen with fewer rows than the crosses need has none
const (
	barRows        = 3
	crossRows      = 2
	fewestForCross = 6
)

// The marks of the frame: a corner cross, the rule between the parts of the status line, and the room a tab's
// number and title are padded with, most first
const (
	crossMark       = "+"
	statusRuleMark  = "│"
	widestTabPad    = 2
	narrowestTabPad = 1
)

// The words of the status line, as Neovim's ruler and its modes write them, and the mark of the command line's
// cursor
const (
	normalMode   = "NORMAL"
	commandMode  = "COMMAND"
	cursorMark   = "▎"
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

// hint is a key the status line names, and what it does.
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

// The mark the status line names the Enter key by, and the word for what both it and the slash do
const (
	enterMark = "↵"
	findHint  = "find"
)

// modeLook is how the status line shows a mode: the name it opens with, and the keys it names, the most needed
// first, as many as fit.
type modeLook struct {
	name  string
	hints []hint
}

// modeLooks are how the status line shows each mode.
func modeLooks() map[mode]modeLook {
	return map[mode]modeLook{
		modeNormal: {name: normalMode, hints: []hint{
			{key: previousKey + hintKeySeparator + nextKey, does: "tabs"},
			{key: downKey + hintKeySeparator + upKey, does: "scroll"},
			{key: searchKey, does: findHint},
			{key: quitKey, does: "quit"},
		}},
		modeSearch: {name: commandMode, hints: []hint{
			{key: enterMark, does: findHint},
			{key: escapeKey, does: "cancel"},
		}},
	}
}

// piece is a piece of the frame: its text and the part of the frame it is.
type piece struct {
	text string
	part part
}

// drawnScreen is the screen painted for the terminal as it is: each tab, and the rows of the page that hold
// nothing but its crosses, or nothing at all.
type drawnScreen struct {
	tabs     []drawnTab
	crossRow string
	blankRow string
}

// drawnTab is one tab painted: its lines, those that hold a match of the search, the line of tabs with it chosen,
// and the status line in each mode but for its ruler.
type drawnTab struct {
	lines   []string
	matches []int
	tabline string
	status  map[mode]drawnStatus
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
	looks := modeLooks()
	for place := range model.tabs {
		status := make(map[mode]drawnStatus, len(looks))
		for shown, seen := range looks {
			status[shown] = model.drawnStatus(place, seen)
		}
		marked := markedLines(model.laid[place], model.search.query)
		tabs = append(tabs, drawnTab{
			lines:   model.paint.tab(marked, model.frame),
			matches: matchingLines(marked),
			tabline: model.paint.bar(model.tabline(place), width),
			status:  status,
		})
	}
	crossRow := ""
	if model.frame.crosses {
		crossRow = model.paint.crossRow(model.frame)
	}
	return drawnScreen{tabs: tabs, crossRow: crossRow, blankRow: strings.Repeat(space, width)}
}

// drawnStatus paints a tab's status line in a mode, all but its ruler.
func (model Model) drawnStatus(place int, seen modeLook) drawnStatus {
	width := model.frame.width
	opening := model.paint.cut(model.statusOpening(place, seen.name), width)
	return drawnStatus{opening: opening, hints: model.paint.cut(statusHints(seen.hints, width-opening.width-widestRuler), width)}
}

// screen draws the whole screen: the line of tabs, the page between its crosses, and the status line. All of it
// is painted already but the ruler.
func (model Model) screen() string {
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
	rows = append(rows, model.statusline(drawn), model.commandLine())
	return strings.Join(rows[max(len(rows)-frame.height, 0):], lineEnd)
}

// commandLine is the last row: the search as the reader types it after the slash, or what the last key said.
func (model Model) commandLine() string {
	width := model.frame.width
	if model.keys.mode == modeSearch {
		// what is typed shows its end where the line is too narrow for all of it, so the cursor stays in sight
		room := width - textWidth(space+searchMark+cursorMark)
		typed := model.search.typing
		for textWidth(typed) > max(room, 0) {
			_, first := utf8.DecodeRuneInString(typed)
			typed = typed[first:]
		}
		typing := []piece{{text: space, part: partCommand}, {text: searchMark, part: partPrompt}, {text: typed, part: partCommand}, {text: cursorMark, part: partCursor}}
		return model.paint.plainRow(typing, width)
	}
	said := partCommand
	if model.said.failed {
		said = partFailure
	}
	return model.paint.plainRow([]piece{{text: space, part: partCommand}, {text: model.said.text, part: said}}, width)
}

// markedLines are a tab's lines with each match of a query parted into a run of its own, worked out once for both
// the painting and the list of the lines that hold a match.
func markedLines(lines []line, query string) [][]matchedRun {
	marked := make([][]matchedRun, 0, len(lines))
	for _, shown := range lines {
		marked = append(marked, matchedRuns(shown, query))
	}
	return marked
}

// matchingLines are the places of the marked lines that hold a match.
func matchingLines(marked [][]matchedRun) []int {
	matching := []int{}
	for place, runs := range marked {
		if slices.ContainsFunc(runs, func(piece matchedRun) bool { return piece.matched }) {
			matching = append(matching, place)
		}
	}
	return matching
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
		pieces = append(pieces, piece{text: pad + strconv.Itoa(first+place+1) + space, part: number}, piece{text: shown.title + pad, part: title})
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
