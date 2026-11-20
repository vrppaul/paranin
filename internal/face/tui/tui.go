// Package tui draws the site's documents for a reader who comes over SSH, as a Bubble Tea program that reads as
// Neovim inside tmux (docs/ARCHITECTURE.md, ADR-0003): each tab of a page a screen of its own, one line of tmux at
// the foot naming the tabs as windows, and over it Neovim's command line while the reader types after a colon or a
// slash, or a key said something. A tab is laid out once for a size and painted once for a theme, so a key press only
// chooses which painted lines show (.claude/rules/tui.md).
package tui

import (
	"slices"
	"time"

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
	// slugs are the slugs of the projects' pages the session holds, in order, as Tab goes round them
	slugs []string
	// site writes the addresses links lead to, and names the session in the footer
	site     routes.Site
	terminal terminal
	// today is the day the session opened, which a role still held runs to on the career's line
	today time.Time
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
	// actions are what each key does in normal mode, commands what the command line runs, and places where it goes
	// from the page shown
	actions    map[string]action
	commands   []command
	places     []place
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
// of full colour until it says what it is, opened on a day, which a role still held runs to on the career's line. A
// size of nothing is taken for a common terminal's.
func New(documents []document.Document, site routes.Site, width, height int, today time.Time) Model {
	if width <= 0 || height <= 0 {
		width, height = assumedWidth, assumedHeight
	}
	buffers, shown := buffersOf(documents), resumePage()
	slugs := projectSlugs(buffers)
	model := Model{
		buffers:  buffers,
		shown:    shown,
		tabs:     buffers[shown],
		left:     placesLeft(buffers, shown),
		slugs:    slugs,
		site:     site,
		terminal: terminal{width: width, height: height, theme: look.ThemeDark, profile: colorprofile.TrueColor},
		today:    today,
		frame:    frame{width: 0, height: 0, left: 0, inner: 0, above: 0, rows: 0, under: 0, crosses: false},
		laid:     nil,
		paint:    paint{kinds: nil, parts: nil, gradient: nil, swatches: nil, fadedLess: nil, fadedMost: nil, highlightGround: nil, highlightText: nil, matchGround: nil, matchText: nil, cursorColour: nil, fewColours: false, noColours: false},
		drawn: drawnScreen{
			tabs: nil, crossRow: "", blankRow: "", track: "", thumb: "",
			footer:    paintedFooter{place: footerPlace{margin: 0, keysShown: false, room: 0}, gutter: "", keys: paintedBar{text: "", width: 0}},
			keyWindow: paintedWindow{rows: nil, width: 0},
		},
		chosen:     0,
		scrolls:    make([]int, len(buffers[shown])),
		actions:    normalActions(),
		commands:   commands(buffers),
		places:     placesOf(buffers[shown], slugs),
		keys:       keyState{mode: modeNormal, typing: "", topPending: false, count: 0, countSequence: 0},
		search:     searchState{begunAt: 0, query: "", shown: nil},
		completion: completionState{typed: "", listed: nil, matched: "", widest: menuWidths{name: 0, kind: 0}, chosen: noneChosen, first: 0},
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
	layout := layout{site: model.site, width: model.frame.inner, pageRows: model.frame.rows, today: model.today}
	model.laid = make([][]line, 0, len(model.tabs))
	for _, shown := range model.tabs {
		model.laid = append(model.laid, layout.tab(shown))
	}
	// the matches shown were places among the matches of the layout gone
	model.search.shown = make([]shownMatch, len(model.tabs))
	model = model.repainted()
	// the line a search began at is a scroll of the chosen tab, held to the new layout as every scroll is
	model.search.begunAt = model.scrollWithin(model.chosen, model.search.begunAt)
	// the menu over the command line keeps the item chosen in sight and stays full in the rows the new size has
	completion := model.completion
	model.completion.first = firstInSight(completion.first, completion.chosen, model.frame.rows, len(completion.listed))
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
