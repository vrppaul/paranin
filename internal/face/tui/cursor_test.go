package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// A terminal narrower than the prompt's opening
const narrowerThanPrompt = 1

// While the reader types on the prompt, the terminal's own cursor stands on the row of what they typed, right after
// it, on a screen too short to spare rows for nothing too.
func TestTheCursorStandsAfterWhatIsTyped(test *testing.T) {
	cases := []struct {
		name    string
		height  int
		opening string
		text    string
		after   string
	}{
		{name: "a search", height: commonHeight, opening: "/", text: "pa", after: "~ /pa"},
		{name: "a place or a command", height: commonHeight, opening: ":", text: "", after: "~ ❯ "},
		{name: "a search on a screen of the fewest rows", height: smallestHeight, opening: "/", text: "pa", after: "~ /pa"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := typed(test, pressed(test, fixtureModel(test, commonWidth, testCase.height), testCase.opening), testCase.text)
			cursor := model.View().Cursor
			if cursor == nil {
				test.Fatalf("after %s%s no cursor shows, want one", testCase.opening, testCase.text)
			}
			row := strings.Split(screenOf(test, model), lineEnd)[cursor.Position.Y]
			end := strings.Index(row, testCase.after)
			if end < 0 || cursor.Position.X != textWidth(row[:end+len(testCase.after)]) {
				test.Errorf("after %s%s the cursor stands at %+v on the row %q, want it right after %q", testCase.opening, testCase.text,
					cursor.Position, row, testCase.after)
			}
		})
	}
}

// On a screen too narrow for what is typed, or for the prompt's opening, the cursor stands in the last column of the
// prompt's row.
func TestTheCursorStaysOnTheScreen(test *testing.T) {
	cases := []struct {
		name  string
		width int
		text  string
	}{
		{name: "a line too narrow", width: smallestWidth, text: "parser parser"},
		{name: "a line narrower than the opening", width: narrowerThanPrompt, text: "pa"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := typed(test, pressed(test, fixtureModel(test, testCase.width, commonHeight), "/"), testCase.text)
			cursor := model.View().Cursor
			if cursor == nil {
				test.Fatalf("at %d columns after /%s no cursor shows, want one", testCase.width, testCase.text)
			}
			prompt, _ := frameOf(testCase.width, commonHeight).rowOf(footerPrompt)
			if want := (tea.Position{X: testCase.width - 1, Y: prompt}); cursor.Position != want {
				test.Errorf("at %d columns after /%s the cursor stands at %+v, want %+v", testCase.width, testCase.text, cursor.Position, want)
			}
		})
	}
}

// Where the reader types nothing, in normal mode and in the window of keys, no cursor shows: none before a key,
// and none once the search or the prompt is left.
func TestNoCursorShowsWhereNothingIsTyped(test *testing.T) {
	cases := []struct {
		name string
		keys []string
	}{
		{name: "normal mode", keys: nil},
		{name: "the window of keys", keys: []string{"?"}},
		{name: "a search left by Esc", keys: []string{"/", "esc"}},
		{name: "a search left by a backspace on nothing", keys: []string{"/", "backspace"}},
		{name: "the prompt left by Esc", keys: []string{":", "esc"}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if cursor := pressed(test, fixtureModel(test, commonWidth, commonHeight), testCase.keys...).View().Cursor; cursor != nil {
				test.Errorf("after %q a cursor shows at %+v, want none", testCase.keys, *cursor)
			}
		})
	}
}
