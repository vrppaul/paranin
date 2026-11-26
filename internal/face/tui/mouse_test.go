package tui

import (
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/vrppaul/paranin/internal/fixtures"
)

// screenCell is a cell of the screen: its row from the top and its column from the left.
type screenCell struct {
	row    int
	column int
}

// cellOf is the cell of a text's first letter where the text first shows on a screen, read from the top row down.
func cellOf(test *testing.T, screen, text string) screenCell {
	test.Helper()
	for row, shown := range strings.Split(screen, lineEnd) {
		if before, _, found := strings.Cut(shown, text); found {
			return screenCell{row: row, column: ansi.StringWidth(before)}
		}
	}
	test.Fatalf("the screen =\n%s\nholds no %q to click", screen, text)
	return screenCell{row: 0, column: 0}
}

// clickOn is a press of a button of the mouse on a cell, as the terminal reports it.
func clickOn(cell screenCell, button tea.MouseButton) tea.MouseClickMsg {
	return tea.MouseClickMsg{X: cell.column, Y: cell.row, Button: button, Mod: 0}
}

// wheelStep is a step of the wheel the way a button names, over the screen's top left cell.
func wheelStep(button tea.MouseButton) tea.MouseWheelMsg {
	return tea.MouseWheelMsg{X: 0, Y: 0, Button: button, Mod: 0}
}

// updated is a model after what the terminal says, each in turn.
func updated(test *testing.T, model Model, messages ...tea.Msg) Model {
	for _, message := range messages {
		next, _ := model.Update(message)
		model = as[Model](test, next)
	}
	return model
}

// clickedOn is a model after a left click on the first letter of a text where it first shows on its screen.
func clickedOn(test *testing.T, model Model, text string) Model {
	return updated(test, model, clickOn(cellOf(test, screenOf(test, model), text), tea.MouseLeft))
}

// The TUI asks the terminal for the mouse, its clicks and its wheel, as Neovim and tmux do with their mouse on.
func TestTheTUITakesTheMouse(test *testing.T) {
	if mode := fixtureModel(test, commonWidth, commonHeight).View().MouseMode; mode != tea.MouseModeCellMotion {
		test.Errorf("the view asks for the mouse in mode %v, want %v", mode, tea.MouseModeCellMotion)
	}
}

// A left click on a link of the blog's list stands the keys on it and follows it, as Enter on it does: a tile, its
// edges among it, to its post, a tag to the list it filters, a grouping to the list grouped so; back from the post with
// q, the keys stand on the tile clicked. A click on no link, or with another button, does nothing. So on the least
// terminal and a roomy one.
func TestAClickOnALinkOfTheListFollowsTheLink(test *testing.T) {
	// what tells the lists apart on their first screen, as TestTheKeysGoFromLinkToLink reads them; and the day set
	// into the top edge of the reader's tile, the row's last, and the last cell of its bottom edge before the corner
	const (
		byYear         = "2 posts"
		bySeries       = "1 of 2 out"
		taggedTools    = "The second post"
		readerTile     = "Reading slowly"
		readerDay      = "10 Nov 2026"
		lastBottomEdge = "─┘"
	)
	cases := []struct {
		name    string
		clicked string
		button  tea.MouseButton
		keys    []string
		footer  string
		holds   string
		lacks   string
	}{
		{name: "a tile", clicked: readerTile, button: tea.MouseLeft, keys: nil, footer: postRead, holds: readerTile, lacks: byYear},
		{name: "a tile's top edge, on its date", clicked: readerDay, button: tea.MouseLeft, keys: nil, footer: postRead, holds: readerTile, lacks: byYear},
		{name: "a tile's bottom edge, by its right corner", clicked: lastBottomEdge, button: tea.MouseLeft, keys: nil, footer: postRead, holds: readerTile, lacks: byYear},
		{name: "a tile, then q and Enter", clicked: readerTile, button: tea.MouseLeft, keys: []string{"q", "enter"}, footer: postRead, holds: readerTile, lacks: byYear},
		{name: "a tag", clicked: "#tools", button: tea.MouseLeft, keys: nil, footer: listAlone, holds: taggedTools, lacks: readerTile},
		{name: "a grouping", clicked: "series", button: tea.MouseLeft, keys: nil, footer: listAlone, holds: bySeries, lacks: byYear},
		{name: "the blog's line, no link", clicked: "What the writer reads", button: tea.MouseLeft, keys: nil, footer: listAlone, holds: byYear, lacks: bySeries},
		{name: "a tile with the right button", clicked: readerTile, button: tea.MouseRight, keys: nil, footer: listAlone, holds: byYear, lacks: bySeries},
		{name: "a tile with the middle button", clicked: readerTile, button: tea.MouseMiddle, keys: nil, footer: listAlone, holds: byYear, lacks: bySeries},
	}
	terminals := []struct {
		name   string
		width  int
		height int
	}{
		{name: "the least", width: leastWidth, height: leastHeight},
		{name: "a roomy one", width: roomyWidth, height: roomyHeight},
	}
	drawing := sessionDrawing(contentWith(test, secondPost()), false)
	for _, terminal := range terminals {
		for _, testCase := range cases {
			test.Run(terminal.name+"/"+testCase.name, func(test *testing.T) {
				list := commanded(test, New(drawing, "blog", terminal.width, terminal.height, fixtures.Day()), "e blog")
				clicked := updated(test, list, clickOn(cellOf(test, screenOf(test, list), testCase.clicked), testCase.button))
				screen := screenOf(test, pressed(test, clicked, testCase.keys...))
				if footer := footerOf(screen); !strings.Contains(footer, testCase.footer) {
					test.Errorf("after a click on %q and %q the footer = %q, want %q", testCase.clicked, testCase.keys, footer, testCase.footer)
				}
				if !strings.Contains(screen, testCase.holds) || strings.Contains(screen, testCase.lacks) {
					test.Errorf("after a click on %q and %q the screen =\n%s\nwant %q on it and no %q", testCase.clicked, testCase.keys, screen, testCase.holds, testCase.lacks)
				}
			})
		}
	}
}

// A click off the page follows no link, though a tile lies on the line past the page's edge there: on the row of
// crosses above it, or on the command line's row under it.
func TestAClickOffThePageFollowsNoLink(test *testing.T) {
	// a screen too short for the list whole; the row of crosses above the page and the command line's row under it;
	// and the steps of the wheel that take the first tile's top rows above the page, its bottom still in sight
	const (
		shortScreen          = 14
		crossRow             = 1
		commandLineRow       = shortScreen - 2
		wheelStepsPastItsTop = 4
	)
	list := handBlogModel(test, commonWidth, shortScreen)
	tileColumn := cellOf(test, screenOf(test, list), "1 Nov 2026").column
	cases := []struct {
		name  string
		shown Model
		row   int
	}{
		{name: "above the page", shown: updated(test, list, slices.Repeat([]tea.Msg{wheelStep(tea.MouseWheelDown)}, wheelStepsPastItsTop)...), row: crossRow},
		{name: "under the page", shown: list, row: commandLineRow},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			clicked := updated(test, testCase.shown, clickOn(screenCell{row: testCase.row, column: tileColumn}, tea.MouseLeft))
			if footer := footerOf(screenOf(test, clicked)); !strings.Contains(footer, listAlone) {
				test.Errorf("after a click on row %d the footer = %q, want the list still alone, %q", testCase.row, footer, listAlone)
			}
		})
	}
}

// A left click on the footer does what the key it names does: on a window, choosing it, as its digit does; on a
// session not shown, showing it as its reader left it, the window they stood on chosen, as tmux switches a client; on
// the key to the window of keys, opening it, wherever the page stands; on the gap between two windows, the mark of
// windows given way, or the blog's session of a site with no blog, nothing.
func TestAClickOnTheFooterDoesWhatItsKeyDoes(test *testing.T) {
	// a terminal wide enough for the sessions and one tab in the footer; and what stands before a window's number, the
	// gap between two windows, then the window's own padding
	const (
		oneTabWidth = 47
		windowGap   = "  "
	)
	resume, post := openedModel(test, "", wideWidth, keyWindowHeight), openedModel(test, "blog/1", commonWidth, keyWindowHeight)
	cases := []struct {
		name    string
		opened  Model
		clicked string
		want    string
	}{
		{name: "a window", opened: resume, clicked: "3:projects", want: "3:projects*"},
		{name: "a window of the blog", opened: post, clicked: "1:posts", want: "1:posts*"},
		{name: "the gap before a window", opened: resume, clicked: windowGap + "3:projects", want: "1:about*"},
		{name: "the blog's session", opened: resume, clicked: "[blog]", want: listAlone},
		{name: "the site's session as left", opened: commanded(test, pressed(test, resume, "l"), "e blog"), clicked: "[example.test]", want: "2:experience*"},
		{name: "the blog's session as left", opened: commanded(test, post, "e tidy-tool"), clicked: "[blog]", want: postRead},
		{name: "the blog's session, where the site has no blog", opened: handResume(test, "Mira", wideWidth, keyWindowHeight), clicked: "[blog]", want: "1:mira*"},
		{name: "the mark of windows given way", opened: pressed(test, openedModel(test, "", oneTabWidth, keyWindowHeight), "h"), clicked: givenWayMark, want: "3:projects*"},
		{name: "the key to the window of keys", opened: post, clicked: "? keys", want: "esc close"},
		{name: "the key to the window of keys, the page scrolled", opened: pressed(test, post, "j"), clicked: "? keys", want: "esc close"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if screen := screenOf(test, clickedOn(test, testCase.opened, testCase.clicked)); !strings.Contains(screen, testCase.want) {
				test.Errorf("after a click on %q the screen =\n%s\nwant %q on it", testCase.clicked, screen, testCase.want)
			}
		})
	}
}

// A click opens the window of keys only on the key's own words, however wide the ruler beside them writes where the
// page stands: on the cell before them it opens nothing.
func TestAClickBeforeTheKeyToTheWindowOfKeysOpensNothing(test *testing.T) {
	post := openedModel(test, "blog/1", commonWidth, keyWindowHeight)
	cases := []struct {
		name  string
		shown Model
	}{
		{name: "the page at its top", shown: post},
		{name: "the page scrolled", shown: pressed(test, post, "j")},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if screen := screenOf(test, clickedOn(test, testCase.shown, " ? keys")); strings.Contains(screen, "esc close") {
				test.Errorf("after a click before the key to the window of keys the screen =\n%s\nwant no window of keys", screen)
			}
		})
	}
}

// A step of the wheel scrolls the page three lines, down or up, as Neovim's mousescroll ver:3 does: as j or k pressed
// three times. Its left and right do nothing.
func TestAStepOfTheWheelScrollsThreeLines(test *testing.T) {
	cases := []struct {
		name   string
		wheels []tea.MouseButton
		keys   []string
	}{
		{name: "down", wheels: []tea.MouseButton{tea.MouseWheelDown}, keys: []string{"j", "j", "j"}},
		{name: "down twice", wheels: []tea.MouseButton{tea.MouseWheelDown, tea.MouseWheelDown}, keys: []string{"j", "j", "j", "j", "j", "j"}},
		{name: "down, then up", wheels: []tea.MouseButton{tea.MouseWheelDown, tea.MouseWheelDown, tea.MouseWheelUp}, keys: []string{"j", "j", "j", "j", "j", "j", "k", "k", "k"}},
		{name: "left", wheels: []tea.MouseButton{tea.MouseWheelLeft}, keys: nil},
		{name: "right", wheels: []tea.MouseButton{tea.MouseWheelRight}, keys: nil},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			experience := pressed(test, fixtureModel(test, commonWidth, shortHeight), toExperience()...)
			wheeled := experience
			for _, button := range testCase.wheels {
				wheeled = updated(test, wheeled, wheelStep(button))
			}
			if got, want := screenOf(test, wheeled), screenOf(test, pressed(test, experience, testCase.keys...)); got != want {
				test.Errorf("after the wheel's %v the screen =\n%s\nwant it as after %q:\n%s", testCase.wheels, got, testCase.keys, want)
			}
		})
	}
}

// The mouse acts in normal mode alone, as Neovim's mouse=nvi leaves out the command line: typing after a colon or a
// slash, a click or the wheel does nothing. With the window of keys open, a click of any button closes it and does
// nothing more, and the wheel does nothing. A click or a step of the wheel leaves what a key leaves: a first g
// dropped, what the command line said cleared.
func TestTheMouseActsInNormalModeAlone(test *testing.T) {
	// a résumé roomy enough for the window of keys, and one short enough that its experience tab scrolls
	resume, experience := fixtureModel(test, wideWidth, keyWindowHeight), pressed(test, fixtureModel(test, commonWidth, shortHeight), toExperience()...)
	projects := cellOf(test, screenOf(test, resume), "3:projects")
	onProjects, rightOnProjects := clickOn(projects, tea.MouseLeft), clickOn(projects, tea.MouseRight)
	onNothing := clickOn(screenCell{row: 0, column: 0}, tea.MouseLeft)
	cases := []struct {
		name   string
		opened Model
		inputs []tea.Msg
		sameAs []tea.Msg
	}{
		{name: "a click typing a command", opened: resume, inputs: []tea.Msg{keyPress(test, ":"), onProjects}, sameAs: []tea.Msg{keyPress(test, ":")}},
		{name: "a click typing a search", opened: resume, inputs: []tea.Msg{keyPress(test, "/"), onProjects}, sameAs: []tea.Msg{keyPress(test, "/")}},
		{name: "the wheel typing a command", opened: experience, inputs: []tea.Msg{keyPress(test, ":"), wheelStep(tea.MouseWheelDown)}, sameAs: []tea.Msg{keyPress(test, ":")}},
		{name: "the wheel typing a search", opened: experience, inputs: []tea.Msg{keyPress(test, "/"), wheelStep(tea.MouseWheelDown)}, sameAs: []tea.Msg{keyPress(test, "/")}},
		{name: "a click on the window of keys", opened: resume, inputs: []tea.Msg{keyPress(test, "?"), onProjects}, sameAs: nil},
		{name: "a right click on the window of keys", opened: resume, inputs: []tea.Msg{keyPress(test, "?"), rightOnProjects}, sameAs: nil},
		{name: "the wheel on the window of keys", opened: resume, inputs: []tea.Msg{keyPress(test, "?"), wheelStep(tea.MouseWheelDown)}, sameAs: []tea.Msg{keyPress(test, "?")}},
		{
			name:   "the wheel after a first g",
			opened: experience,
			inputs: []tea.Msg{keyPress(test, "G"), keyPress(test, "g"), wheelStep(tea.MouseWheelUp), keyPress(test, "g")},
			sameAs: []tea.Msg{keyPress(test, "G"), wheelStep(tea.MouseWheelUp)},
		},
		{
			name:   "a click after a first g",
			opened: experience,
			inputs: []tea.Msg{keyPress(test, "G"), keyPress(test, "g"), onNothing, keyPress(test, "g")},
			sameAs: []tea.Msg{keyPress(test, "G")},
		},
		{name: "a click after a command said something", opened: resume, inputs: []tea.Msg{keyPress(test, ":"), keyPress(test, "y"), keyPress(test, "enter"), onNothing}, sameAs: nil},
		{name: "the wheel after a command said something", opened: resume, inputs: []tea.Msg{keyPress(test, ":"), keyPress(test, "y"), keyPress(test, "enter"), wheelStep(tea.MouseWheelUp)}, sameAs: nil},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got, want := screenOf(test, updated(test, testCase.opened, testCase.inputs...)), screenOf(test, updated(test, testCase.opened, testCase.sameAs...)); got != want {
				test.Errorf("after %v the screen =\n%s\nwant it as after %v:\n%s", testCase.inputs, got, testCase.sameAs, want)
			}
		})
	}
}
