// Package tui draws the site's documents in a terminal as a Bubble Tea program laid out like Neovim
// (docs/ARCHITECTURE.md, ADR-0003): a line of tabs above, the page between two rows of corner crosses, a status line
// and a command line below. The head of a document and what stands before its first section are the first tab; each
// section after them is a tab of its own, titled as the document titles it. The reader moves between tabs and
// scrolls each one, by a count typed before a key where they type one, and each tab keeps its place; a slash
// searches the tab, and n and N go through its matches. While the reader types a search, the tab marks what they
// type and its page goes to the first match, as Neovim's incsearch does, but for the slash, which takes the marks of
// the search before away at once; the command line counts the matches, as Neovim does. A colon opens the command
// line, which runs a command, or names one it does not know with Neovim's error. The session holds the résumé and
// each project's page as buffers, as Neovim holds the files it opened: :e and a project's slug shows the project's
// page in place of the résumé, and q, or :q, shows the résumé again, as the first version closed a report; each
// buffer keeps the tab and the scrolls its reader left. The status line names :e last, where a project's page is
// there to open; the page itself holds the document's words alone. While the reader types, a window over the page
// lists the commands whose names start with what is typed, as the first version showed them, where Neovim waits for
// Tab; after :e and a space it lists the projects' pages whose slugs start with what follows. Tab and Shift+Tab go
// round what it lists, writing the item chosen on the line, and past either end what was typed, as Neovim's wildmenu
// does. A question mark or a space opens the window of keys, which names the keys of normal mode and what each does,
// and which the next key closes; the frame says q closes a project's page and quits on the résumé. A search stands in
// for the cursor this TUI has not: y copies the value of the fact on the line found, as Neovim's yy yanks a line,
// to the reader's clipboard over OSC 52, as the first version copied a value of its head; a link gives its text, as
// its line shows it. A window shows whole or not at all: the window of keys opens only where it fits, where it does
// not the command line says Neovim's E36, and the window closes when a resize leaves it no room. On the command
// line the cursor is the terminal's own, a bar the terminal blinks, as the first version's blinked, in the accent
// where the terminal shows many colours: the program writes no frame to blink it, so a session left alone writes
// nothing and its idle timeout runs out. The résumé's title, the author's name, ends in the site's cursor, the bar
// in the accent the page, the PDF and the card draw there. It stands still, where the first version's blinked for
// half a minute after a key: a blink would need a tick, a count and the head painted again each half second, where
// a tab is painted once.
//
// A tab is laid out once for a size and painted once for a theme, the terminal's colours and the search, with the frame
// around it and the window of keys: a key press only chooses which painted lines show, and writes the ruler that says
// where it scrolled, the command line and the window of commands. Lipgloss lays a window over the rows it covers
// alone, as the screen is drawn. A key that changes what the search line holds, or opens or leaves it, paints the
// chosen tab's page again; a search confirmed or taken away paints the tabs again, once; a buffer shown in place of
// another is laid out and painted, once. What a piece of text is (text, muted, a command, a result worth seeing
// first) is the layout's to say; how it looks is the paint's, from the design's colours in look, for the colours
// the terminal shows. Every address is written by routes; a link carries its address to a terminal that follows
// links.
package tui

import (
	"slices"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/look"
	"github.com/vrppaul/paranin/internal/routes"
)

// The size a terminal is taken to be until it says its own
const (
	assumedWidth  = 80
	assumedHeight = 24
)

// Model is one reader's session: the buffers of the documents it holds, the tabs of the one shown, the terminal it
// is drawn in, and where the reader is.
type Model struct {
	// buffers are the tabs of each page the session holds; shown is the page shown, and tabs its tabs; left is
	// where the reader stood in each page not shown, and chosen and scrolls below where they stand in the one shown
	buffers map[document.Page][]tab
	shown   document.Page
	tabs    []tab
	left    map[document.Page]readerPlace
	// slugs are the slugs of the projects' pages the session holds, in order, as the window of commands lists them
	slugs []string
	// site writes the addresses links lead to, and names the site in the line of tabs
	site     routes.Site
	terminal terminal
	// frame and laid are where the page sits and each tab laid out in it, worked out again when the terminal's
	// size changes; paint and drawn are how it looks and each tab painted, worked out again when its size, its
	// background or its colours change. Never for a key
	frame  frame
	laid   [][]line
	paint  paint
	drawn  drawnScreen
	chosen int
	// scrolls is how far down each tab the reader has scrolled, by the tab's place
	scrolls []int
	// actions are what each key does in normal mode, and commands what the command line runs
	actions    map[string]action
	commands   []command
	keys       keyState
	search     searchState
	completion completionState
	// said is what the command line says after the last key
	said commandMessage
}

// terminal is what the session knows of the reader's terminal: its size, its background, the colours it shows.
type terminal struct {
	width   int
	height  int
	theme   look.Theme
	profile colorprofile.Profile
}

// New is a session that holds documents, the résumé among them and shown first, in a terminal of a size, a dark one
// of full colour until it says what it is. A size of nothing is taken for a common terminal's.
func New(documents []document.Document, site routes.Site, width, height int) Model {
	if width <= 0 || height <= 0 {
		width, height = assumedWidth, assumedHeight
	}
	buffers, shown := buffersOf(documents), resumePage()
	model := Model{
		buffers:    buffers,
		shown:      shown,
		tabs:       buffers[shown],
		left:       placesLeft(buffers, shown),
		slugs:      projectSlugs(buffers),
		site:       site,
		terminal:   terminal{width: width, height: height, theme: look.ThemeDark, profile: colorprofile.TrueColor},
		frame:      frame{width: 0, height: 0, left: 0, inner: 0, rows: 0, crosses: false},
		laid:       nil,
		paint:      paint{kinds: nil, parts: nil, highlightGround: nil, highlightText: nil, matchGround: nil, matchText: nil, cursorColour: nil, fewColours: false, noColours: false},
		drawn:      drawnScreen{tabs: nil, crossRow: "", blankRow: "", keyWindow: paintedWindow{rows: nil, width: 0}},
		chosen:     0,
		scrolls:    make([]int, len(buffers[shown])),
		actions:    normalActions(),
		commands:   commands(closingHint(shown)),
		keys:       keyState{mode: modeNormal, typing: "", topPending: false, count: 0, countSequence: 0},
		search:     searchState{begunAt: 0, query: "", shown: nil},
		completion: completionState{typed: "", listed: nil, chosen: noneChosen},
		said:       commandMessage{text: "", failed: false},
	}
	return model.relaid()
}

// Init asks the terminal for its background, so the theme follows it.
func (model Model) Init() tea.Cmd {
	return tea.RequestBackgroundColor
}

// Update takes what the terminal says: its size, its background, its colours, or a key; and the pause after a
// digit, which the model asks for itself.
func (model Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := message.(type) {
	case tea.WindowSizeMsg:
		if typed.Width > 0 && typed.Height > 0 {
			model.terminal.width, model.terminal.height = typed.Width, typed.Height
			return model.relaid(), nil
		}
	case tea.BackgroundColorMsg:
		model.terminal.theme = themeOf(typed.IsDark())
		return model.repainted(), nil
	case tea.ColorProfileMsg:
		model.terminal.profile = typed.Profile
		return model.repainted(), nil
	case tea.KeyPressMsg:
		return model.pressed(typed)
	case countPaused:
		return model.countEnded(typed), nil
	}
	return model, nil
}

// View is the screen as the reader sees it, over the whole terminal.
func (model Model) View() tea.View {
	screen, cursor := model.screen()
	view := tea.NewView(screen)
	view.AltScreen = true
	view.Cursor = cursor
	return view
}

func themeOf(dark bool) look.Theme {
	if dark {
		return look.ThemeDark
	}
	return look.ThemeLight
}

// relaid is the model with every tab laid out again for the terminal's size, then painted.
func (model Model) relaid() Model {
	model.frame = frameOf(model.terminal.width, model.terminal.height)
	layout := layout{site: model.site, width: model.frame.inner}
	model.laid = make([][]line, 0, len(model.tabs))
	for _, shown := range model.tabs {
		model.laid = append(model.laid, layout.tab(shown))
	}
	// the matches shown were places among the matches of the layout gone
	model.search.shown = make([]shownMatch, len(model.tabs))
	model = model.repainted()
	// the line a search began at is a scroll of the chosen tab, held to the new layout as every scroll is
	model.search.begunAt = model.scrollWithin(model.chosen, model.search.begunAt)
	// an open window of keys closes where the new size cannot show it whole, so no mode stays open unseen
	if model.keys.mode == modeKeys && !model.shows(model.drawn.keyWindow) {
		model.keys.mode = modeNormal
	}
	return model
}

// repainted is the model with every tab and the frame around it painted again, for the terminal's background
// and colours.
func (model Model) repainted() Model {
	model.paint = paintFor(model.terminal.theme, model.terminal.profile)
	model.drawn = model.drawnScreen()
	return model.clamped()
}

// repaintedChosen is the model with the chosen tab's page painted again, for the search it marks.
func (model Model) repaintedChosen() Model {
	model.drawn.tabs = slices.Clone(model.drawn.tabs)
	model.drawn.tabs[model.chosen].paintedPage = model.paintedPage(model.chosen)
	return model
}

// choose shows a tab by its place, round from the last to the first and back.
func (model Model) choose(place int) Model {
	model.chosen = wrapped(place, len(model.tabs))
	return model
}

// wrapped is a place among a count of them, round from the last to the first and back.
func wrapped(place, count int) int {
	return (place%count + count) % count
}

// scrolledTo is the chosen tab scrolled to a line.
func (model Model) scrolledTo(line int) Model {
	model.scrolls = slices.Clone(model.scrolls)
	model.scrolls[model.chosen] = line
	return model.clamped()
}

// clamped is the model with each tab scrolled within it, as the tab is drawn now.
func (model Model) clamped() Model {
	scrolls := make([]int, len(model.drawn.tabs))
	for place := range model.drawn.tabs {
		scrolls[place] = model.scrollWithin(place, model.scrolls[place])
	}
	model.scrolls = scrolls
	return model
}

// scrollWithin is a scroll of a tab held between its start and the last line that fills the page, as the tab is
// drawn now.
func (model Model) scrollWithin(place, scroll int) int {
	return min(max(scroll, 0), max(len(model.drawn.tabs[place].lines)-model.frame.rows, 0))
}
