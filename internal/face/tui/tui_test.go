package tui

import (
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// The content package's fixture, and the address it is served at
const (
	fixturePath = "../../content/testdata/content"
	fixtureBase = "https://example.test"
)

// A terminal of a common size, and one tall enough to show any tab whole
const (
	commonWidth  = 100
	commonHeight = 30
	tallHeight   = 400
	narrowWidth  = 40
	// a terminal of six rows shows two lines of a page
	shortHeight = 6
	// a phone's terminal, narrower than the page with its margins, and the smallest that still draws
	phoneWidth     = 28
	phoneHeight    = 20
	smallestWidth  = 10
	smallestHeight = 3
)

// The first screen is the head of the résumé, and the tabs name every section after it.
func TestTheFirstScreenShowsTheHeadAndEveryTab(test *testing.T) {
	screen := screenOf(test, fixtureModel(test, commonWidth, commonHeight))
	cases := []struct {
		name string
		want string
	}{
		{name: "the name", want: "Mira Example"},
		{name: "the title", want: "Software Engineer"},
		{name: "the headline", want: "Builds small tools that last."},
		{name: "the first section's tab", want: "Facts"},
		{name: "the last section's tab", want: "Languages"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !strings.Contains(screen, testCase.want) {
				test.Errorf("the first screen =\n%s\nwant it to hold %q", screen, testCase.want)
			}
		})
	}
}

// The tab keys move to the next tab and the previous one, round from the last to the first.
func TestTheTabKeysMoveBetweenTabs(test *testing.T) {
	cases := []struct {
		name  string
		keys  []string
		holds string
		lacks string
	}{
		{name: "l goes to the next tab", keys: []string{"l"}, holds: "5 years", lacks: "Builds small tools that last."},
		{name: "tab goes to the next tab", keys: []string{"tab"}, holds: "5 years", lacks: "Builds small tools that last."},
		{name: "the right arrow goes to the next tab", keys: []string{"right"}, holds: "5 years", lacks: "Builds small tools that last."},
		{name: "two steps go two tabs on", keys: []string{"l", "l"}, holds: "Northwind", lacks: "5 years"},
		{name: "h goes round from the first tab to the last", keys: []string{"h"}, holds: "fluent", lacks: "Builds small tools that last."},
		{name: "shift and tab go back", keys: []string{"l", "shift+tab"}, holds: "Builds small tools that last.", lacks: "5 years"},
		{name: "the left arrow goes back", keys: []string{"l", "left"}, holds: "Builds small tools that last.", lacks: "5 years"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			screen := screenOf(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), testCase.keys...))
			if !strings.Contains(screen, testCase.holds) || strings.Contains(screen, testCase.lacks) {
				test.Errorf("after %q the screen =\n%s\nwant it to hold %q and not %q", testCase.keys, screen, testCase.holds, testCase.lacks)
			}
		})
	}
}

// The scrolling keys move a tab taller than the terminal a line at a time, to its end and back to its start.
// The experience tab opens with a role, then its first point.
func TestTheScrollingKeysMoveThroughATab(test *testing.T) {
	toExperience := []string{"l", "l"}
	cases := []struct {
		name  string
		keys  []string
		holds string
		lacks string
	}{
		{name: "the tab opens at its start", keys: nil, holds: "Northwind", lacks: "one long winter"},
		{name: "j goes a line down", keys: []string{"j"}, holds: "Kept the queue moving", lacks: "Northwind"},
		{name: "the down arrow goes a line down", keys: []string{"down"}, holds: "Kept the queue moving", lacks: "Northwind"},
		{name: "k goes a line back up", keys: []string{"j", "k"}, holds: "Northwind", lacks: "one long winter"},
		{name: "the up arrow goes a line back up", keys: []string{"j", "up"}, holds: "Northwind", lacks: "one long winter"},
		{name: "G goes to the end", keys: []string{"G"}, holds: "one long winter", lacks: "Northwind"},
		{name: "gg goes back to the start", keys: []string{"G", "g", "g"}, holds: "Northwind", lacks: "one long winter"},
		{name: "one g alone stays", keys: []string{"G", "g"}, holds: "one long winter", lacks: "Northwind"},
		{name: "a key between two g's parts them", keys: []string{"G", "g", "k", "g"}, holds: "one long winter", lacks: "Northwind"},
		{name: "k at the start stays there", keys: []string{"k"}, holds: "Northwind", lacks: "one long winter"},
		{name: "j at the end stays there", keys: []string{"G", "j"}, holds: "one long winter", lacks: "Northwind"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(toExperience, testCase.keys...)...)
			screen := screenOf(test, model)
			if !strings.Contains(screen, testCase.holds) || strings.Contains(screen, testCase.lacks) {
				test.Errorf("after %q the screen =\n%s\nwant it to hold %q and not %q", testCase.keys, screen, testCase.holds, testCase.lacks)
			}
		})
	}
}

// Each tab keeps where it was scrolled to while the reader looks at another.
func TestEachTabKeepsItsPlace(test *testing.T) {
	model := pressed(test, fixtureModel(test, commonWidth, shortHeight), "l", "l", "G", "l", "h")
	if screen := screenOf(test, model); !strings.Contains(screen, "one long winter") {
		test.Errorf("back on the experience tab the screen =\n%s\nwant it still at the end", screen)
	}
}

// The chosen tab is named in the line of tabs at any width: where the tabs do not fit, the first ones give way.
func TestTheChosenTabIsNamedAtAnyWidth(test *testing.T) {
	model := pressed(test, fixtureModel(test, narrowWidth, commonHeight), "h")
	if tabline := strings.SplitN(screenOf(test, model), "\n", 2)[0]; !strings.Contains(tabline, "7 Languages") {
		test.Errorf("on the last tab the line of tabs = %q, want it to name %q", tabline, "7 Languages")
	}
}

// A role reads with its period, each month its year and its number, and its place.
func TestARoleReadsWithItsPeriod(test *testing.T) {
	screen := screenOf(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), "l", "l"))
	cases := []struct {
		name string
		want string
	}{
		{name: "a role still held", want: "2023-04 – present · Brno"},
		{name: "a role that ended", want: "2020-01 – 2023-03 · Prague"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !strings.Contains(screen, testCase.want) {
				test.Errorf("the experience tab =\n%s\nwant it to hold %q", screen, testCase.want)
			}
		})
	}
}

// A terminal resized gets its page laid out again for its new size: as many of a tab's lines as it has rows show,
// and a terminal as narrow as a phone's still reads, the page then taking its whole width (FR-13, NFR-6).
func TestAResizeLaysThePageOutAgain(test *testing.T) {
	toExperience := []string{"l", "l"}
	cases := []struct {
		name   string
		width  int
		height int
		holds  string
		lacks  string
	}{
		{name: "fewer rows show fewer lines", width: commonWidth, height: shortHeight, holds: "Northwind", lacks: "one long winter"},
		{name: "more rows show more lines", width: commonWidth, height: tallHeight, holds: "one long winter", lacks: "Builds small tools"},
		{name: "a phone's width still reads", width: phoneWidth, height: phoneHeight, holds: "Northwind", lacks: "Builds small tools"},
		{name: "the smallest terminal still draws", width: smallestWidth, height: smallestHeight, holds: "Experi", lacks: "Builds small tools"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, fixtureModel(test, commonWidth, commonHeight), toExperience...)
			resized, _ := model.Update(tea.WindowSizeMsg{Width: testCase.width, Height: testCase.height})
			screen := screenOf(test, as[Model](test, resized))
			if !strings.Contains(screen, testCase.holds) || strings.Contains(screen, testCase.lacks) {
				test.Errorf("resized to %d×%d the screen =\n%s\nwant it to hold %q and not %q", testCase.width, testCase.height,
					screen, testCase.holds, testCase.lacks)
			}
		})
	}
}

// A session opened in a terminal as narrow as a phone's draws, the page taking the whole width (NFR-6).
func TestANarrowTerminalOpensTheResume(test *testing.T) {
	if screen := screenOf(test, fixtureModel(test, phoneWidth, phoneHeight)); !strings.Contains(screen, "Mira Example") {
		test.Errorf("at %d×%d the screen =\n%s\nwant the name", phoneWidth, phoneHeight, screen)
	}
}

// A fact with no value still reads as its label: a document may leave a value out, and the tab still draws.
func TestAFactWithNoValueReadsAsItsLabel(test *testing.T) {
	composed := document.Document{
		Head:   document.Head{Page: document.Page{Kind: document.PageResume, Slug: ""}, Title: nil, Subtitle: nil, Place: "", Lead: nil, Ways: nil},
		Blocks: []document.Block{document.Facts{List: []document.Fact{{Label: "available", Value: nil}}}},
	}
	if screen := screenOf(test, New(composed, routes.New(fixtureBase), commonWidth, commonHeight)); !strings.Contains(screen, "available") {
		test.Errorf("the screen =\n%s\nwant the label %q", screen, "available")
	}
}

// q and Ctrl+C end the session.
func TestTheQuitKeysEndTheSession(test *testing.T) {
	cases := []struct {
		name string
		key  string
	}{
		{name: "q", key: "q"},
		{name: "Ctrl+C", key: "ctrl+c"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			_, command := fixtureModel(test, commonWidth, commonHeight).Update(keyPress(test, testCase.key))
			if command == nil {
				test.Fatalf("%s gave no command, want the end of the session", testCase.key)
			}
			if message := command(); message != tea.Quit() {
				test.Errorf("%s gave %T, want the end of the session", testCase.key, message)
			}
		})
	}
}

// Every phrase of the résumé reaches the TUI on one tab or another, at any width: a terminal reader reads every
// fact a browser reader reads. A word wider than a narrow terminal breaks across lines, so the phrases are read
// with every space and break taken out.
func TestEveryPhraseReachesTheTUI(test *testing.T) {
	resume := compose.Resume(fixtureSnapshot(test))
	for _, width := range []int{commonWidth, narrowWidth} {
		model := New(resume, routes.New(fixtureBase), width, tallHeight)
		screens := []string{}
		for range tabsOf(resume) {
			screens = append(screens, screenOf(test, model))
			model = pressed(test, model, "l")
		}
		read := strings.Join(strings.Fields(strings.Join(screens, " ")), "")
		for _, phrase := range document.Phrases(resume) {
			if !strings.Contains(read, strings.Join(strings.Fields(phrase), "")) {
				test.Errorf("at %d columns no tab holds %q", width, phrase)
			}
		}
	}
}

// A link leads where its target does in a terminal that follows links: its text carries its address.
func TestALinkCarriesItsAddress(test *testing.T) {
	view := fixtureModel(test, commonWidth, commonHeight).View().Content
	cases := []struct {
		name    string
		address string
	}{
		{name: "a public profile", address: "https://github.test/mira"},
		{name: "the mail address", address: "mailto:mira@example.test"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !strings.Contains(view, ansi.SetHyperlink(testCase.address)) {
				test.Errorf("the first screen links no %q", testCase.address)
			}
		})
	}
}

func fixtureModel(test *testing.T, width, height int) Model {
	return New(compose.Resume(fixtureSnapshot(test)), routes.New(fixtureBase), width, height)
}

func fixtureSnapshot(test *testing.T) *content.Snapshot {
	snapshot, err := content.Load(os.DirFS(fixturePath))
	if err != nil {
		test.Fatalf("Load(%q) = %v, want no error", fixturePath, err)
	}
	return snapshot
}

// pressed is a model after the keys, each pressed in turn.
func pressed(test *testing.T, model Model, keys ...string) Model {
	for _, key := range keys {
		next, _ := model.Update(keyPress(test, key))
		model = as[Model](test, next)
	}
	return model
}

// keyPress is a key as the terminal sends it, by the name Bubble Tea gives it.
func keyPress(test *testing.T, name string) tea.KeyPressMsg {
	keys := map[string]tea.KeyPressMsg{
		"j":         {Code: 'j', Text: "j"},
		"k":         {Code: 'k', Text: "k"},
		"h":         {Code: 'h', Text: "h"},
		"l":         {Code: 'l', Text: "l"},
		"g":         {Code: 'g', Text: "g"},
		"G":         {Code: 'g', ShiftedCode: 'G', Mod: tea.ModShift, Text: "G"},
		"q":         {Code: 'q', Text: "q"},
		"tab":       {Code: tea.KeyTab},
		"shift+tab": {Code: tea.KeyTab, Mod: tea.ModShift},
		"left":      {Code: tea.KeyLeft},
		"right":     {Code: tea.KeyRight},
		"up":        {Code: tea.KeyUp},
		"down":      {Code: tea.KeyDown},
		"ctrl+c":    {Code: 'c', Mod: tea.ModCtrl},
	}
	key, known := keys[name]
	if !known || key.String() != name {
		test.Fatalf("no key reads as %q", name)
	}
	return key
}

// screenOf is the text a reader sees, its colours gone.
func screenOf(test *testing.T, model Model) string {
	test.Helper()
	return ansi.Strip(model.View().Content)
}

// as is a value as the kind a test expects of it, or the test's end, saying what it was.
func as[Kind any](test *testing.T, value any) Kind {
	typed, isKind := value.(Kind)
	if !isKind {
		test.Fatalf("%+v is no %T", value, typed)
	}
	return typed
}
