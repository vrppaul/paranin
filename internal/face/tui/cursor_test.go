package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// A terminal narrower than the command line's mark and its margins
const narrowerThanTheMark = 1

// An e and a combining acute accent: one letter on the screen, two code points
const accentedLetter = "e\U00000301"

// While the reader types on the command line, the terminal's own cursor stands on the row of what they typed, right
// after it, on a screen too short to spare rows for nothing too.
func TestTheCursorStandsAfterWhatIsTyped(test *testing.T) {
	cases := []struct {
		name    string
		height  int
		opening string
		text    string
		after   string
	}{
		{name: "a search", height: commonHeight, opening: "/", text: "pa", after: "/pa"},
		{name: "a place or a command", height: commonHeight, opening: ":", text: "", after: ":"},
		{name: "a search on a screen of the fewest rows", height: smallestHeight, opening: "/", text: "pa", after: "/pa"},
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

// On a screen too narrow for what is typed, or for the command line's mark, the cursor stands in the last column the
// command line holds, on its row, the one over the footer: a column in from the screen's edge, as the footer stands,
// where the screen has the columns for that.
func TestTheCursorStaysOnTheScreen(test *testing.T) {
	// the command line stands on the row over the footer, the screen's last
	const commandLineRow = commonHeight - 2
	cases := []struct {
		name       string
		width      int
		text       string
		lastColumn int
	}{
		{name: "a line too narrow", width: smallestWidth, text: "parser parser", lastColumn: smallestWidth - 2},
		{name: "a line narrower than the mark", width: narrowerThanTheMark, text: "pa", lastColumn: 0},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := typed(test, pressed(test, fixtureModel(test, testCase.width, commonHeight), "/"), testCase.text)
			cursor := model.View().Cursor
			if cursor == nil {
				test.Fatalf("at %d columns after /%s no cursor shows, want one", testCase.width, testCase.text)
			}
			if want := (tea.Position{X: testCase.lastColumn, Y: commandLineRow}); cursor.Position != want {
				test.Errorf("at %d columns after /%s the cursor stands at %+v, want %+v", testCase.width, testCase.text, cursor.Position, want)
			}
		})
	}
}

// A line typed wider than the command line shows its end, where the cursor stands and the next letter goes.
func TestALineTooLongShowsItsEnd(test *testing.T) {
	model := typed(test, pressed(test, fixtureModel(test, smallestWidth, commonHeight), "/"), "abcdefghij")
	if line := commandLineOf(screenOf(test, model)); line != "/efghij" {
		test.Errorf("at %d columns after /abcdefghij the command line = %q, want %q", smallestWidth, line, "/efghij")
	}
}

// A screen of one row has no row over the footer for the command line: a search typed there leaves the footer in
// sight and shows no cursor.
func TestAScreenOfOneRowKeepsTheFooter(test *testing.T) {
	const oneRow = 1
	model := typed(test, pressed(test, fixtureModel(test, commonWidth, oneRow), "/"), "pa")
	if screen := screenOf(test, model); !strings.Contains(screen, "[example.test]") || model.View().Cursor != nil {
		test.Errorf("on a screen of one row after /pa the screen = %q and the cursor %+v, want the footer and no cursor", screen, model.View().Cursor)
	}
}

// Where the reader types nothing, in normal mode and in the window of keys, no cursor shows: none before a key,
// and none once the search or the command line is left.
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
