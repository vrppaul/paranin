package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// A colon opens the command line, where what the reader types shows after it.
func TestTheColonOpensTheCommandLine(test *testing.T) {
	model := typed(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), ":"), "qa")
	if below := belowThePage(screenOf(test, model)); !strings.Contains(below, "COMMAND") || !strings.Contains(below, ":qa") {
		test.Errorf("after a colon and qa the status and command lines = %q, want the command mode and :qa", below)
	}
}

// Each of Neovim's ways to quit, typed on the command line, ends the session, as q does, spaces around it or not.
func TestAQuitCommandEndsTheSession(test *testing.T) {
	for _, name := range []string{"q", "quit", "qa", "q!", " q "} {
		test.Run(name, func(test *testing.T) {
			model := typed(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), ":"), name)
			_, command := model.Update(keyPress(test, "enter"))
			if command == nil {
				test.Fatalf(":%s gave no command, want the end of the session", name)
			}
			if message := command(); message != tea.Quit() {
				test.Errorf(":%s gave %T, want the end of the session", name, message)
			}
		})
	}
}

// A command the command line does not know is named back with Neovim's error, and the keys go back to normal; the
// start of a command's name is no command.
func TestAnUnknownCommandIsNamedBack(test *testing.T) {
	for _, name := range []string{"nope", "qui"} {
		test.Run(name, func(test *testing.T) {
			model := pressed(test, typed(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), ":"), name), "enter")
			below := belowThePage(screenOf(test, model))
			if !strings.Contains(below, "E492: Not an editor command: "+name) || !strings.Contains(below, "NORMAL") {
				test.Errorf("after :%s the status and command lines = %q, want E492 and normal mode", name, below)
			}
		})
	}
}

// The window of commands at the fixture's commands is twelve columns wide: its border, the mark, ":q", the gap,
// "quit" and a space
const commandWindowWidth = 12

// Esc and Ctrl+C leave the command line, as do a backspace and Enter on nothing, running nothing.
func TestTheCommandLineIsLeftUnrun(test *testing.T) {
	cases := []struct {
		name string
		text string
		keys []string
	}{
		{name: "Esc", text: "q", keys: []string{"esc"}},
		{name: "Ctrl+C", text: "q", keys: []string{"ctrl+c"}},
		{name: "a backspace on nothing", text: "q", keys: []string{"backspace", "backspace"}},
		{name: "Enter on nothing", text: "", keys: []string{"enter"}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := typed(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), ":"), testCase.text)
			for _, key := range testCase.keys {
				next, command := model.Update(keyPress(test, key))
				if command != nil {
					test.Fatalf("after :%s and %q a command came, want none", testCase.text, testCase.keys)
				}
				model = as[Model](test, next)
			}
			if below := belowThePage(screenOf(test, model)); !strings.Contains(below, "NORMAL") || strings.Contains(below, "E492") {
				test.Errorf("after :%s and %q the status and command lines = %q, want normal mode and no error", testCase.text,
					testCase.keys, below)
			}
		})
	}
}

// While the reader types a command, a window lists the commands it may become with what each does; a command
// Neovim has under several names is listed by one, and what starts no command lists none.
func TestTheCompletionListsTheCommandsTyped(test *testing.T) {
	cases := []struct {
		name  string
		text  string
		holds string
		lacks string
	}{
		{name: "nothing typed lists every command", text: "", holds: "quit", lacks: ":qa"},
		{name: "the start of a command lists it", text: "q", holds: "quit", lacks: ":q!"},
		{name: "what starts no command lists none", text: "x", holds: ":x", lacks: "quit"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			screen := screenOf(test, typed(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), ":"), testCase.text))
			if !strings.Contains(screen, testCase.holds) || strings.Contains(screen, testCase.lacks) {
				test.Errorf("after :%s the screen =\n%s\nwant it to hold %q and not %q", testCase.text, screen, testCase.holds, testCase.lacks)
			}
		})
	}
}

// A second colon opens the command line on nothing typed, and the window lists every command again.
func TestASecondColonOpensOnNothing(test *testing.T) {
	model := pressed(test, typed(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), ":"), "x"), "esc", ":")
	if screen := screenOf(test, model); !strings.Contains(screen, "quit") || strings.Contains(screen, ":x") {
		test.Errorf("after :x, Esc and a colon the screen =\n%s\nwant every command listed and nothing typed", screen)
	}
}

// The command line opens at any size, the window of commands left out where it would cover the line of tabs or
// stand wider than the screen.
func TestTheCommandLineOpensAtAnySize(test *testing.T) {
	cases := []struct {
		name   string
		width  int
		height int
		listed bool
	}{
		{name: "a phone", width: phoneWidth, height: phoneHeight, listed: true},
		{name: "one as short as the window fits", width: commonWidth, height: shortHeight - 1, listed: true},
		{name: "one a row too short for the window", width: commonWidth, height: shortHeight - 2, listed: false},
		{name: "one as narrow as the window", width: commandWindowWidth, height: phoneHeight, listed: true},
		{name: "one a column narrower than the window", width: commandWindowWidth - 1, height: phoneHeight, listed: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			screen := screenOf(test, pressed(test, fixtureModel(test, testCase.width, testCase.height), ":"))
			if !strings.Contains(belowThePage(screen), ":▎") || strings.Contains(screen, "quit") != testCase.listed {
				test.Errorf("at %dx%d after a colon the screen =\n%s\nwant the command line, and the window listed: %t",
					testCase.width, testCase.height, screen, testCase.listed)
			}
		})
	}
}

// Tab chooses the first command the window lists, marks it there and writes it on the command line, which Enter
// runs.
func TestTabCompletesACommand(test *testing.T) {
	model := pressed(test, fixtureModel(test, commonWidth, commonHeight), ":", "tab")
	if screen := screenOf(test, model); !strings.Contains(screen, "▎:q") || !strings.Contains(belowThePage(screen), ":q▎") {
		test.Errorf("after a colon and Tab the screen =\n%s\nwant :q marked in the window and written on the line", screen)
	}
	_, command := model.Update(keyPress(test, "enter"))
	if command == nil || command() != tea.Quit() {
		test.Errorf("after a colon, Tab and Enter, want the session ended by the command chosen")
	}
}

// Editing the line takes the choice away: the window lists anew, none marked.
func TestAnEditTakesTheChoiceAway(test *testing.T) {
	model := pressed(test, fixtureModel(test, commonWidth, commonHeight), ":", "tab", "backspace")
	if screen := screenOf(test, model); !strings.Contains(screen, "quit") || strings.Contains(screen, "▎:q") {
		test.Errorf("after a colon, Tab and a backspace the screen =\n%s\nwant the window listing :q unmarked", screen)
	}
}
