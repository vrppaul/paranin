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

// Drawing is what a session draws, as its host hands it: the documents it holds, in every language; the site's words of
// the frame round them in each language, by the translation its pages are in (compose.Site's Frames); and the site
// that writes their addresses.
type Drawing struct {
	Documents []document.Document
	Frames    map[string]compose.Frame
	Site      routes.Site
}

// Model is one reader's session: the buffers of the documents it holds, the tabs of the one shown, the terminal it
// is drawn in, and where the reader is.
type Model struct {
	// buffers are the tabs of each page the session holds, in every language. sitePage is the page the site's session
	// shows, blog what the blog's session shows, in one language, and inBlog whether the blog's is the session shown;
	// tabs are its windows. left is where the reader stood in each page, as they left it, and chosen, scrolls and links
	// below where they stand in the tabs shown
	buffers  map[document.Page][]tab
	sitePage document.Page
	blog     blogSession
	inBlog   bool
	tabs     []tab
	left     map[document.Page]readerPlace
	// pagePlaces are each page's own places (document.PlacesOf), and sitePlaces those every page's command line goes
	// to after them, the projects' pages and the window of keys
	pagePlaces map[document.Page][]document.Place
	sitePlaces []place
	// opened are the pages :e opens in every language by every name it reads, as compose names them, an SSH user name
	// among them
	opened map[string]document.Page
	// frames are the site's words of the frame in each language, by the language a page is translated into
	// (document.Page.Translation), and site writes the addresses links lead to
	frames   map[string]compose.Frame
	site     routes.Site
	terminal terminal
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
	// positions are where the reader stands in each tab, by the tab's place
	positions []tabPosition
	// actions are what each key does in normal mode, commands what the command line runs, and places where it goes
	// from the tabs shown
	actions    map[string]action
	commands   []command
	places     shownPlaces
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
// reached by names, an SSH session's user's (D26): a page of the blog in any language; that language's list, saying it
// finds no page, where the name is under its blog's but names none; the résumé where it names nothing of a blog. In a
// terminal of a size, a dark one of full colour until it says what it is, on a day, which a role still held runs to on
// the career's line. A size of nothing is taken for a common terminal's.
func New(drawing Drawing, reachedBy string, width, height int, today time.Time) Model {
	if width <= 0 || height <= 0 {
		width, height = assumedWidth, assumedHeight
	}
	buffers, frames := buffersOf(drawing.Documents), drawing.Frames
	own := frames[document.Untranslated]
	model := Model{
		buffers:    buffers,
		sitePage:   document.PageOf(document.PageResume),
		blog:       blogSession{list: document.PageOf(document.PageBlog), post: document.PageOf(document.PageNone), atPost: false},
		inBlog:     false,
		tabs:       nil,
		left:       placesLeft(buffers),
		pagePlaces: pagePlacesOf(drawing.Documents),
		sitePlaces: sitePlaces(own.Projects),
		opened:     openedPages(own.Pages),
		frames:     frames,
		site:       drawing.Site,
		terminal:   terminal{width: width, height: height, theme: look.ThemeDark, profile: colorprofile.TrueColor},
		today:      today,
		frame:      frame{width: 0, height: 0, left: 0, inner: 0, above: 0, rows: 0, under: 0, crosses: false},
		laid:       nil,
		paint:      paint{kinds: nil, parts: nil, gradient: nil, swatches: nil, fadedLess: nil, fadedMost: nil, highlightGround: nil, highlightText: nil, pointedGround: nil, pointedUnderline: nil, matchGround: nil, matchText: nil, matchUnderline: nil, cursorColour: nil, fewColours: false, noColours: false},
		drawn: drawnScreen{
			tabs: nil, crossRow: "", blankRow: "", track: "", thumb: "",
			footer: paintedFooter{place: footerPlace{margin: 0, keysShown: false, room: 0}, gutter: "", keys: paintedBar{text: "", width: 0}},
		},
		chosen:     0,
		positions:  nil,
		actions:    normalActions(),
		commands:   commands(buffers),
		places:     shownPlaces{tabs: nil, own: nil, site: nil, named: nil},
		keys:       keyState{mode: modeNormal, typing: "", topPending: false},
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
		return model.pressed(typed)
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
// keep the link the keys stand on in sight.
func (model Model) relaid() Model {
	model.frame = frameOf(model.terminal.width, model.terminal.height, widestPageOf(model.inBlog))
	model.laid = make([]laidTab, 0, len(model.tabs))
	for _, shown := range model.tabs {
		model.laid = append(model.laid, model.layoutIn(shown.page.Translation).tab(shown))
	}
	// the matches shown were places among the matches of the layout gone
	model.search.shown = make([]shownMatch, len(model.tabs))
	model = model.repainted().withLinksInSight()
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

// layoutIn is how the tabs shown of a page in a language are laid out, by the language a page is translated into: in
// the site's words in it, for the terminal's size as the frame has it.
func (model Model) layoutIn(translation string) layout {
	return layout{site: model.site, siteWords: model.frames[translation], width: model.frame.inner, pageRows: model.frame.rows, today: model.today}
}

// repainted is the model with every tab and the frame around it painted again, for the terminal's background
// and colours.
func (model Model) repainted() Model {
	model.paint = paintFor(model.terminal.theme, model.terminal.profile)
	painted := make([]paintedTab, 0, len(model.tabs))
	for place := range model.tabs {
		painted = append(painted, model.paintedTab(place))
	}
	model.drawn = model.drawnScreen(painted)
	return model.clamped()
}

// repaintedChosen is the model with the chosen tab's page painted again, for the search it marks.
func (model Model) repaintedChosen() Model {
	model.drawn.tabs = slices.Clone(model.drawn.tabs)
	model.drawn.tabs[model.chosen].paintedPage = model.paintedPage(model.chosen)
	return model
}

// repaintedPages is the model with every tab's page painted again, for the search they mark.
func (model Model) repaintedPages() Model {
	model.drawn.tabs = slices.Clone(model.drawn.tabs)
	for place := range model.drawn.tabs {
		model.drawn.tabs[place].paintedPage = model.paintedPage(place)
	}
	return model.clamped()
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

// scrolledTo is the chosen tab scrolled to a line; on the blog's list, where the link the keys stand on goes out of
// sight, they stand on the link in sight nearest the line, the later of a row going down and the earlier going up.
func (model Model) scrolledTo(line int) Model {
	column := 0
	if line > model.positions[model.chosen].scroll {
		column = model.frame.inner
	}
	return model.scrolledToward(line, spot{line: line, column: column})
}

// scrolledToward is the chosen tab scrolled to a line; on the blog's list, where the link the keys stand on goes out
// of sight, they stand on the link in sight nearest a spot the move went to.
func (model Model) scrolledToward(line int, near spot) Model {
	model.positions = slices.Clone(model.positions)
	model.positions[model.chosen].scroll = line
	return model.clamped().withChosenLinkNear(near)
}

// clamped is the model with each tab scrolled within it, as the tab is drawn now.
func (model Model) clamped() Model {
	positions := slices.Clone(model.positions)
	for place := range positions {
		positions[place].scroll = model.scrollWithin(place, positions[place].scroll)
	}
	model.positions = positions
	return model
}

// scrollWithin is a scroll of a tab held between its start and the last line that fills the page, as the tab is
// drawn now.
func (model Model) scrollWithin(place, scroll int) int {
	return min(max(scroll, 0), max(len(model.drawn.tabs[place].lines)-model.frame.rows, 0))
}
