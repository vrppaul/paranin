package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// The fewest rows a terminal has that shows the window of keys whole: its sixteen rows, the row above the page and
// the roomy footer's seven
const keyWindowHeight = 24

// A question mark or a space opens the window of keys over the page, which names the keys of normal mode and what
// each does, the tabs a number chooses among them, and how the window closes.
func TestTheWindowOfKeysNamesTheKeys(test *testing.T) {
	rows := []string{
		"esc close", "j k", "line down / up", "3j", "move by a count", "gg G", "top / bottom", "h l", "tab before / after",
		"1-6", "go to a tab", "type a place to go", "complete a place", "find in the tab", "n N", "next / prior match",
		"end the search", "quit",
	}
	for _, key := range []string{"?", "space"} {
		test.Run(key, func(test *testing.T) {
			above := aboveTheStatusLine(screenOf(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), key)))
			for _, want := range rows {
				if !strings.Contains(above, want) {
					test.Errorf("after %s the page =\n%s\nwant the window to hold %q", key, above, want)
				}
			}
		})
	}
}

// Esc, the question mark and the space close the window and do nothing else; any other key closes it and does what
// it does in normal mode.
func TestAKeyClosesTheWindowOfKeys(test *testing.T) {
	cases := []struct {
		name  string
		key   string
		holds string
	}{
		{name: "Esc", key: "esc", holds: "Builds small tools that last."},
		{name: "the question mark", key: "?", holds: "Builds small tools that last."},
		{name: "the space", key: "space", holds: "Builds small tools that last."},
		{name: "a key of normal mode", key: "l", holds: "Kept the queue"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			screen := screenOf(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), "?", testCase.key))
			if !strings.Contains(screen, testCase.holds) || strings.Contains(screen, "next / prior match") {
				test.Errorf("after ? and %s the screen =\n%s\nwant the window closed and %q", testCase.key, screen, testCase.holds)
			}
		})
	}
}

// The window of keys opens only where it shows whole: there Esc closes the window alone, so the search stays and n
// goes on through it; where the window would stand taller or wider than the screen, the question mark opens nothing,
// the prompt says there is no room, as Neovim does, the keys stay in normal mode, and Esc takes the search
// away. A phone's width is the window's own.
func TestTheWindowOfKeysOpensOnlyWhole(test *testing.T) {
	cases := []struct {
		name   string
		width  int
		height int
		opened bool
	}{
		{name: "a common terminal", width: commonWidth, height: commonHeight, opened: true},
		{name: "a phone", width: phoneWidth, height: tallHeight, opened: true},
		{name: "one as short as the window fits", width: commonWidth, height: keyWindowHeight, opened: true},
		{name: "one a row too short", width: commonWidth, height: keyWindowHeight - 1, opened: false},
		{name: "one a column narrower than the window", width: phoneWidth - 1, height: tallHeight, opened: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, fixtureModel(test, testCase.width, testCase.height), append(toExperience(), "/")...)
			model = pressed(test, typed(test, model, "the"), "enter", "?")
			if below := belowThePage(screenOf(test, model)); strings.Contains(below, "E36: Not enough room") == testCase.opened {
				test.Errorf("at %dx%d after a search and ? the footer = %q, want no room said: %t",
					testCase.width, testCase.height, below, !testCase.opened)
			}
			model = pressed(test, model, "esc", "n")
			if below := belowThePage(screenOf(test, model)); strings.Contains(below, "/the [2/4]") != testCase.opened {
				test.Errorf("at %dx%d after a search, ?, Esc and n the footer = %q, want the window opened: %t",
					testCase.width, testCase.height, below, testCase.opened)
			}
		})
	}
}

// A resize that leaves the open window of keys no room closes it, so the keys are back in normal mode, where Esc
// takes the search away.
func TestAResizeClosesAWindowOfKeysLeftNoRoom(test *testing.T) {
	model := pressed(test, fixtureModel(test, commonWidth, commonHeight), append(toExperience(), "/")...)
	model = pressed(test, typed(test, model, "the"), "enter", "?")
	resized, _ := model.Update(tea.WindowSizeMsg{Width: commonWidth, Height: shortHeight})
	if below := belowThePage(screenOf(test, pressed(test, as[Model](test, resized), "esc", "n"))); strings.Contains(below, "/the") {
		test.Errorf("after a search, ?, a resize too short for the window, Esc and n the footer = %q, want the search gone", below)
	}
}
