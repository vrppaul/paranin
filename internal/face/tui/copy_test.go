package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// A terminal far wider than the widest page: a long value still wraps on the page, and the prompt shows what it says
// whole
const wideWidth = 160

// y copies the value of the fact on the line the search found, whole, to the reader's clipboard over OSC 52, and the
// prompt says what it copied, for a terminal that takes no OSC 52; a link gives its text, as its line shows it,
// and a value wrapped over lines is copied whole from any of them.
func TestYankCopiesTheValueOnTheLineFound(test *testing.T) {
	cases := []struct {
		name  string
		keys  []string
		query string
		value string
	}{
		{name: "the mail", keys: nil, query: "mira@", value: "mira@example.test"},
		{name: "a link", keys: nil, query: "github.test", value: "github.test/mira"},
		{name: "the ssh command", keys: nil, query: "ssh ex", value: "ssh example.test"},
		{
			name: "a value from its second line", keys: []string{"l"}, query: "next",
			value: "Small tools for people who live in the terminal, written to last and to be read by whoever comes next",
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, pressed(test, fixtureModel(test, wideWidth, commonHeight), testCase.keys...), "/")
			model = pressed(test, typed(test, model, testCase.query), "enter")
			next, command := model.Update(keyPress(test, "y"))
			if command == nil {
				test.Fatalf("after /%s and y no command came, want the value copied", testCase.query)
			}
			if got, want := command(), tea.SetClipboard(testCase.value)(); got != want {
				test.Errorf("after /%s and y the command gave %v, want %v", testCase.query, got, want)
			}
			if below := belowThePage(screenOf(test, as[Model](test, next))); !strings.Contains(below, "copied: "+testCase.value) {
				test.Errorf("after /%s and y the footer = %q, want %q", testCase.query, below, "copied: "+testCase.value)
			}
		})
	}
}

// y copies nothing where no line holding a value is found and shown: with no search, on a found line of no value,
// once the reader scrolled the page from where the search left it, or once Esc has taken the search away; the
// prompt says so.
func TestYankWithNoValueFoundCopiesNothing(test *testing.T) {
	cases := []struct {
		name  string
		query string
		keys  []string
	}{
		{name: "no search", query: "", keys: nil},
		{name: "a found line of no value", query: "Writes software", keys: nil},
		{name: "a page scrolled from the line found", query: "mira@", keys: []string{"j"}},
		{name: "the search taken away", query: "mira@", keys: []string{"esc"}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := typed(test, pressed(test, fixtureModel(test, commonWidth, shortHeight), "/"), testCase.query)
			model = pressed(test, pressed(test, model, "enter"), testCase.keys...)
			next, command := model.Update(keyPress(test, "y"))
			if command != nil {
				test.Errorf("after /%s, %q and y a command came, want nothing copied", testCase.query, testCase.keys)
			}
			if below := belowThePage(screenOf(test, as[Model](test, next))); !strings.Contains(below, "nothing to copy") {
				test.Errorf("after /%s, %q and y the footer = %q, want nothing to copy said", testCase.query,
					testCase.keys, below)
			}
		})
	}
}
