package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// A terminal narrower than the search line's prompt, the space and the slash before what is typed
const narrowerThanPrompt = 1

// While the reader types on the command line, the terminal's own cursor stands on its last row after what they
// typed; on a line too narrow for all of it, or for the slash, in the last column.
func TestTheCursorStandsAfterWhatIsTyped(test *testing.T) {
	cases := []struct {
		name    string
		width   int
		opening string
		text    string
		column  int
	}{
		{name: "a search", width: commonWidth, opening: "/", text: "pa", column: len(" /pa")},
		{name: "a command", width: commonWidth, opening: ":", text: "", column: len(" :")},
		{name: "a line too narrow", width: smallestWidth, opening: "/", text: "parser parser", column: smallestWidth - 1},
		{name: "a line narrower than the slash", width: narrowerThanPrompt, opening: "/", text: "pa", column: narrowerThanPrompt - 1},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := typed(test, pressed(test, fixtureModel(test, testCase.width, commonHeight), testCase.opening), testCase.text)
			cursor := model.View().Cursor
			if cursor == nil {
				test.Fatalf("after %s%s no cursor shows, want one", testCase.opening, testCase.text)
			}
			if want := (tea.Position{X: testCase.column, Y: commonHeight - 1}); cursor.Position != want {
				test.Errorf("after %s%s the cursor stands at %+v, want %+v", testCase.opening, testCase.text, cursor.Position, want)
			}
		})
	}
}

// Where the reader types nothing, in normal mode and in the window of keys, no cursor shows: none before a key,
// and none once the search line or the command line is left.
func TestNoCursorShowsWhereNothingIsTyped(test *testing.T) {
	cases := []struct {
		name string
		keys []string
	}{
		{name: "normal mode", keys: nil},
		{name: "the window of keys", keys: []string{"?"}},
		{name: "a search left by Esc", keys: []string{"/", "esc"}},
		{name: "a search left by a backspace on nothing", keys: []string{"/", "backspace"}},
		{name: "the command line left by Esc", keys: []string{":", "esc"}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if cursor := pressed(test, fixtureModel(test, commonWidth, commonHeight), testCase.keys...).View().Cursor; cursor != nil {
				test.Errorf("after %q a cursor shows at %+v, want none", testCase.keys, *cursor)
			}
		})
	}
}
