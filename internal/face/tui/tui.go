// Package tui draws a document in a terminal as a Bubble Tea program laid out like Neovim (docs/ARCHITECTURE.md,
// ADR-0003): a line of tabs above, the page between two rows of corner crosses, a status line below. The head of the
// document and what stands before its first section are the first tab; each section after them is a tab of its own,
// titled as the document titles it. The reader moves between tabs and scrolls each one, and each keeps its place.
//
// A tab is laid out once for a size and painted once for a theme and the terminal's colours, with the frame around it:
// a key press only chooses which painted lines show, and writes the ruler that says where it scrolled. What a piece of
// text is (text, muted, a command, a result worth seeing first) is the layout's to say; how it looks is the paint's,
// from the design's colours in look, for the colours the terminal shows. Every address is written by routes; a link
// carries its address to a terminal that follows links.
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

// The keys a reader presses, as Bubble Tea names them
const (
	quitKey        = "q"
	interruptKey   = "ctrl+c"
	downKey        = "j"
	downArrow      = "down"
	upKey          = "k"
	upArrow        = "up"
	topKey         = "g"
	homeKey        = "home"
	bottomKey      = "G"
	endKey         = "end"
	nextKey        = "l"
	rightArrow     = "right"
	tabKey         = "tab"
	previousKey    = "h"
	leftArrow      = "left"
	backwardTabKey = "shift+tab"
)

// Model is one reader's session: the tabs of the document, the terminal it is drawn in, and where the reader is.
type Model struct {
	tabs []tab
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
	// topPending is a first g pressed, which a second one makes the top of the tab
	topPending bool
}

// terminal is what the session knows of the reader's terminal: its size, its background, the colours it shows.
type terminal struct {
	width   int
	height  int
	theme   look.Theme
	profile colorprofile.Profile
}

// New is a session that draws a document in a terminal of a size, a dark one of full colour until it says what
// it is. A size of nothing is taken for a common terminal's.
func New(composed document.Document, site routes.Site, width, height int) Model {
	if width <= 0 || height <= 0 {
		width, height = assumedWidth, assumedHeight
	}
	tabs := tabsOf(composed)
	model := Model{
		tabs:       tabs,
		site:       site,
		terminal:   terminal{width: width, height: height, theme: look.ThemeDark, profile: colorprofile.TrueColor},
		frame:      frame{width: 0, height: 0, left: 0, inner: 0, rows: 0, crosses: false},
		laid:       nil,
		paint:      paint{kinds: nil, parts: nil, highlightGround: nil, highlightText: nil, fewColours: false, noColours: false},
		drawn:      drawnScreen{tabs: nil, crossRow: "", blankRow: ""},
		chosen:     0,
		scrolls:    make([]int, len(tabs)),
		topPending: false,
	}
	return model.relaid()
}

// Init asks the terminal for its background, so the theme follows it.
func (model Model) Init() tea.Cmd {
	return tea.RequestBackgroundColor
}

// Update takes what the terminal says: its size, its background, its colours, or a key.
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
		return model.pressed(typed.String())
	}
	return model, nil
}

// View is the screen as the reader sees it, over the whole terminal.
func (model Model) View() tea.View {
	view := tea.NewView(model.screen())
	view.AltScreen = true
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
	return model.repainted()
}

// repainted is the model with every tab and the frame around it painted again, for the terminal's background
// and colours.
func (model Model) repainted() Model {
	model.paint = paintFor(model.terminal.theme, model.terminal.profile)
	model.drawn = model.drawnScreen()
	return model.clamped()
}

// pressed is the model after a key: a step through the tab or between tabs, or the end of the session.
func (model Model) pressed(key string) (Model, tea.Cmd) {
	topPending := model.topPending
	model.topPending = false
	switch key {
	case quitKey, interruptKey:
		return model, tea.Quit
	case downKey, downArrow:
		return model.scrolledTo(model.scrolls[model.chosen] + 1), nil
	case upKey, upArrow:
		return model.scrolledTo(model.scrolls[model.chosen] - 1), nil
	case topKey:
		if topPending {
			return model.scrolledTo(0), nil
		}
		model.topPending = true
	case homeKey:
		return model.scrolledTo(0), nil
	case bottomKey, endKey:
		return model.scrolledTo(len(model.drawn.tabs[model.chosen].lines)), nil
	case nextKey, rightArrow, tabKey:
		return model.choose(model.chosen + 1), nil
	case previousKey, leftArrow, backwardTabKey:
		return model.choose(model.chosen - 1), nil
	}
	return model, nil
}

// choose shows a tab by its place, round from the last to the first and back.
func (model Model) choose(place int) Model {
	model.chosen = (place%len(model.tabs) + len(model.tabs)) % len(model.tabs)
	return model
}

// scrolledTo is the chosen tab scrolled to a line.
func (model Model) scrolledTo(line int) Model {
	model.scrolls = slices.Clone(model.scrolls)
	model.scrolls[model.chosen] = line
	return model.clamped()
}

// clamped is the model with each tab scrolled no further than between its start and the last line that fills
// the page, as the tab is drawn now.
func (model Model) clamped() Model {
	scrolls := make([]int, len(model.drawn.tabs))
	for place, drawn := range model.drawn.tabs {
		scrolls[place] = min(max(model.scrolls[place], 0), max(len(drawn.lines)-model.frame.rows, 0))
	}
	model.scrolls = scrolls
	return model
}
