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

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/look"
	"github.com/vrppaul/paranin/internal/routes"
)

// The size a terminal is taken to be until it says its own
const (
	assumedWidth  = 80
	assumedHeight = 24
)

// Drawing is what a session draws, as its host hands it: the documents it holds, the site's words of the frame round
// them, and the site that writes their addresses.
type Drawing struct {
	Documents []document.Document
	Frame     compose.Frame
	Site      routes.Site
}

// Model is one reader's session: the buffers of the documents it holds, the tabs of the one shown, the terminal it
// is drawn in, and where the reader is.
type Model struct {
	// buffers are the tabs of each page the session holds. sitePage is the page the site's session shows, blog what
	// the blog's session shows, and inBlog whether the blog's is the session shown; tabs are its windows. left is where
	// the reader stood in each page, as they left it, and chosen, scrolls and tiles below where they stand in the tabs
	// shown
	buffers  map[document.Page][]tab
	sitePage document.Page
	blog     blogSession
	inBlog   bool
	tabs     []tab
	left     map[document.Page]readerPlace
	// slugs are the slugs of the projects' pages the session holds, in the documents' order, as Tab goes round them
	slugs []string
	// blogPages are the pages of the blog by the names a terminal reaches them by, as :e and an SSH user name write
	// them
	blogPages map[string]document.Page
	// siteWords are the site's words of the frame, and site writes the addresses links lead to
	siteWords compose.Frame
	site      routes.Site
	terminal  terminal
	// today is the day the session opened, which a role still held runs to on the career's line
	today time.Time
	// frame and laid are where the page sits and each tab laid out in it, worked out again when the terminal's
	// size changes; paint and drawn are how it looks and each tab painted, worked out again when its size, its
	// background or its colours change. Never for a key
	frame  frame
	laid   []laidTab
	paint  paint
	drawn  drawnScreen
	chosen int
	// scrolls is how far down each tab the reader has scrolled, and tiles the tile the keys stand on in each, on the
	// blog's list, by the tab's place
	scrolls []int
	tiles   []int
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

// New is a session that draws what a host hands it, the résumé among its documents, opened on the page the name it is
// reached by names, an SSH session's user's (D26): a page of the blog; the blog's list, saying it finds no page, where
// the name is under the blog's but names none; the résumé where it names nothing of the blog. In a terminal of a size,
// a dark one of full colour until it says what it is, on a day, which a role still held runs to on the career's line.
// A size of nothing is taken for a common terminal's.
func New(drawing Drawing, reachedBy string, width, height int, today time.Time) Model {
	if width <= 0 || height <= 0 {
		width, height = assumedWidth, assumedHeight
	}
	buffers := buffersOf(drawing.Documents)
	model := Model{
		buffers:   buffers,
		sitePage:  document.PageOf(document.PageResume),
		blog:      blogSession{list: document.PageOf(document.PageBlog), post: document.PageOf(document.PageNone), atPost: false},
		inBlog:    false,
		tabs:      nil,
		left:      placesLeft(buffers),
		slugs:     projectSlugs(drawing.Documents),
		blogPages: blogNames(drawing.Documents),
		siteWords: drawing.Frame,
		site:      drawing.Site,
		terminal:  terminal{width: width, height: height, theme: look.ThemeDark, profile: colorprofile.TrueColor},
		today:     today,
		frame:     frame{width: 0, height: 0, left: 0, inner: 0, above: 0, rows: 0, under: 0, crosses: false},
		laid:      nil,
		paint:     paint{kinds: nil, parts: nil, gradient: nil, swatches: nil, fadedLess: nil, fadedMost: nil, highlightGround: nil, highlightText: nil, pointedGround: nil, pointedUnderline: nil, matchGround: nil, matchText: nil, matchUnderline: nil, cursorColour: nil, fewColours: false, noColours: false},
		drawn: drawnScreen{
			tabs: nil, crossRow: "", blankRow: "", margin: "", track: "", thumb: "",
			footer: paintedFooter{place: footerPlace{margin: 0, keysShown: false, room: 0}, gutter: "", keys: paintedBar{text: "", width: 0}},
		},
		chosen:     0,
		scrolls:    nil,
		tiles:      nil,
		actions:    normalActions(),
		commands:   commands(buffers),
		places:     nil,
		keys:       keyState{mode: modeNormal, typing: "", topPending: false, count: 0, countSequence: 0},
		search:     searchState{begunAt: 0, query: "", shown: nil},
		completion: completionState{typed: "", listed: nil, matched: "", widest: menuWidths{name: 0, kind: 0}, chosen: noneChosen, first: 0},
		said:       commandMessage{text: "", failed: false},
	}
	return model.reached(reachedBy)
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

// relaid is the model with every tab laid out again for the terminal's size, then painted, each list scrolled to
// keep the tile the keys stand on in sight.
func (model Model) relaid() Model {
	model.frame = frameOf(model.terminal.width, model.terminal.height, widestPageOf(model.inBlog))
	layout := model.layout()
	model.laid = make([]laidTab, 0, len(model.tabs))
	for _, shown := range model.tabs {
		model.laid = append(model.laid, layout.tab(shown))
	}
	// the matches shown were places among the matches of the layout gone
	model.search.shown = make([]shownMatch, len(model.tabs))
	model = model.repainted().withTilesInSight()
	// the line a search began at is a scroll of the chosen tab, held to the new layout as every scroll is
	model.search.begunAt = model.scrollWithin(model.chosen, model.search.begunAt)
	// the menu over the command line keeps the item chosen in sight and stays full in the rows the new size has
	completion := model.completion
	model.completion.first = firstInSight(completion.first, completion.chosen, model.frame.rows, len(completion.listed))
	return model.windowOfKeysSettled()
}

// windowOfKeysSettled is the model with an open window of keys closed where the screen cannot show the chosen tab's
// whole, so no mode stays open unseen.
func (model Model) windowOfKeysSettled() Model {
	if model.keys.mode == modeKeys && !model.shows(model.drawn.tabs[model.chosen].keyWindow) {
		model.keys.mode = modeNormal
	}
	return model
}

// layout is how the tabs shown are laid out, for the terminal's size as the frame has it.
func (model Model) layout() layout {
	return layout{site: model.site, siteWords: model.siteWords, width: model.frame.inner, pageRows: model.frame.rows, today: model.today}
}

// repainted is the model with every tab and the frame around it painted again, for the terminal's background
// and colours.
func (model Model) repainted() Model {
	model.paint = paintFor(model.terminal.theme, model.terminal.profile)
	pages := make([]paintedPage, 0, len(model.tabs))
	for place := range model.tabs {
		pages = append(pages, model.paintedPage(place))
	}
	model.drawn = model.drawnScreen(pages)
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

// scrolledTo is the chosen tab scrolled to a line; on the blog's list, where the tile the keys stand on goes out of
// sight, they stand on the tile in sight nearest the line, the later of a row going down and the earlier going up.
func (model Model) scrolledTo(line int) Model {
	column := 0
	if line > model.scrolls[model.chosen] {
		column = model.frame.inner
	}
	return model.scrolledToward(line, spot{line: line, column: column})
}

// scrolledToward is the chosen tab scrolled to a line; on the blog's list, where the tile the keys stand on goes out
// of sight, they stand on the tile in sight nearest a spot the move went to.
func (model Model) scrolledToward(line int, near spot) Model {
	model.scrolls = slices.Clone(model.scrolls)
	model.scrolls[model.chosen] = line
	return model.clamped().withChosenTileNear(near)
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
