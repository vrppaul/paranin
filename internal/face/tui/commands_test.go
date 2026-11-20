package tui

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// The presses of Tab that reach the last place from the résumé's first tab: the fixture's four other tabs, then help
const lastPlace = 5

// A colon opens the command line over the footer, where what the reader types shows after it, the footer still under
// it.
func TestTheColonOpensTheCommandLineOverTheFooter(test *testing.T) {
	screen := screenOf(test, typed(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), ":"), "qa"))
	if line, footer := commandLineOf(screen), footerOf(screen); line != ":qa" || !strings.HasPrefix(footer, "[example.test]") {
		test.Errorf("after a colon and qa the command line = %q and the footer %q, want %q over the footer", line, footer, ":qa")
	}
}

// Each of Neovim's ways to quit, typed on the command line, ends the session on the résumé, as q does, spaces around
// it or not; :qa ends it on a project's page too.
func TestAQuitCommandEndsTheSession(test *testing.T) {
	cases := []struct {
		name   string
		opened string
		text   string
	}{
		{name: ":q", opened: "", text: "q"},
		{name: ":quit", opened: "", text: "quit"},
		{name: ":qa", opened: "", text: "qa"},
		{name: ":q!", opened: "", text: "q!"},
		{name: ":q with spaces around", opened: "", text: " q "},
		{name: ":qa on a project's page", opened: "e tidy-tool", text: "qa"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := commanded(test, fixtureModel(test, commonWidth, commonHeight), testCase.opened)
			_, command := typed(test, pressed(test, model, ":"), testCase.text).Update(keyPress(test, "enter"))
			if command == nil {
				test.Fatalf("after :%s, :%s gave no command, want the end of the session", testCase.opened, testCase.text)
			}
			if message := command(); message != tea.Quit() {
				test.Errorf("after :%s, :%s gave %T, want the end of the session", testCase.opened, testCase.text, message)
			}
		})
	}
}

// A command given an argument it cannot take is named back with Neovim's errors: one that takes an argument needs
// it, and one that takes none refuses what follows it.
func TestCommandsRefuseArgumentsTheyCannotTake(test *testing.T) {
	cases := []struct {
		name string
		text string
		want string
	}{
		{name: ":e with nothing", text: "e", want: "E471: Argument required"},
		{name: ":q with something after", text: "q now", want: "E488: Trailing characters: now"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			screen := screenOf(test, commanded(test, fixtureModel(test, commonWidth, commonHeight), testCase.text))
			if line := commandLineOf(screen); line != testCase.want || !strings.Contains(screen, resumeHeadline) {
				test.Errorf("after :%s the screen =\n%s\nwant the command line to say %q, and the résumé still shown", testCase.text, screen, testCase.want)
			}
		})
	}
}

// What names no place and no command is named back with Neovim's error, alone on the command line as Neovim says it,
// and the keys go back to normal, where no cursor shows; the start of a command's name is no command.
func TestAnUnknownCommandIsNamedBack(test *testing.T) {
	for _, name := range []string{"nope", "qui"} {
		test.Run(name, func(test *testing.T) {
			model := pressed(test, typed(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), ":"), name), "enter")
			want := "E492: Not an editor command: " + name
			if line := commandLineOf(screenOf(test, model)); line != want || model.View().Cursor != nil {
				test.Errorf("after :%s the command line = %q and the cursor %+v, want %q and no cursor", name, line, model.View().Cursor, want)
			}
		})
	}
}

// Esc and Ctrl+C leave the command line, as do a backspace and Enter on nothing, running nothing: the screen is back
// as it was before the colon, with no command line and no cursor.
func TestTheCommandLineIsLeftUnrun(test *testing.T) {
	before := screenOf(test, fixtureModel(test, commonWidth, commonHeight))
	cases := []struct {
		name string
		text string
		keys []string
	}{
		{name: "Esc", text: "q", keys: []string{"esc"}},
		{name: "Ctrl+C", text: "q", keys: []string{"ctrl+c"}},
		{name: "a backspace on nothing", text: "q", keys: []string{"backspace", "backspace"}},
		{name: "Enter on nothing", text: "", keys: []string{"enter"}},
		{name: "Esc from the menu of the places", text: "", keys: []string{"esc"}},
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
			if screen := screenOf(test, model); screen != before || model.View().Cursor != nil {
				test.Errorf("after :%s and %q the screen =\n%s\nand the cursor %+v, want the screen as before the colon and no cursor",
					testCase.text, testCase.keys, screen, model.View().Cursor)
			}
		})
	}
}

// At its right the command line names the keys that finish what the reader types: Tab where it has places to go
// round, Enter, and Esc.
func TestTheCommandLineNamesTheKeysThatFinishTyping(test *testing.T) {
	// a line typed that leaves the keys some room on a common terminal, not all they take
	const tooLongForTheKeys = 80
	cases := []struct {
		name    string
		opening string
		text    string
		holds   []string
		lacks   []string
	}{
		{name: "a place to go round", opening: ":", text: "", holds: []string{"tab next", "↵ go", "esc cancel"}, lacks: []string{"find"}},
		{name: "no place to go round", opening: ":", text: "x", holds: []string{"↵ go", "esc cancel"}, lacks: []string{"tab next"}},
		{name: "a search", opening: "/", text: "", holds: []string{"↵ find", "esc cancel"}, lacks: []string{"tab next"}},
		{name: "a line too long to leave them room", opening: ":", text: strings.Repeat("x", tooLongForTheKeys), holds: nil, lacks: []string{"↵"}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := typed(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), testCase.opening), testCase.text)
			foot := atTheFoot(screenOf(test, model))
			for _, want := range testCase.holds {
				if !strings.Contains(foot, want) {
					test.Errorf("after %s%s the foot =\n%s\nwant it to hold %q", testCase.opening, testCase.text, foot, want)
				}
			}
			for _, unwanted := range testCase.lacks {
				if strings.Contains(foot, unwanted) {
					test.Errorf("after %s%s the foot =\n%s\nwant it not to hold %q", testCase.opening, testCase.text, foot, unwanted)
				}
			}
		})
	}
}

// A second colon opens the command line on nothing typed, where Tab goes round the places from the first again.
func TestASecondColonOpensOnNothing(test *testing.T) {
	screen := screenOf(test, pressed(test, typed(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), ":"), "x"), "esc", ":", "tab"))
	if line := commandLineOf(screen); line != ":experience" {
		test.Errorf("after :x, Esc, a colon and Tab the command line = %q, want %q", line, ":experience")
	}
}

// Enter on a place's name goes there: a tab chosen, a project's page opened, the window of keys opened.
func TestEnterOnAPlaceGoesThere(test *testing.T) {
	cases := []struct {
		name  string
		text  string
		holds string
	}{
		{name: "a tab", text: "experience", holds: "Northwind"},
		{name: "a project's page", text: "tidy-tool", holds: projectTagline},
		{name: "the window of keys", text: "help", holds: "next / prior match"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if screen := screenOf(test, commanded(test, fixtureModel(test, commonWidth, commonHeight), testCase.text)); !strings.Contains(screen, testCase.holds) {
				test.Errorf("after :%s the screen =\n%s\nwant it to hold %q", testCase.text, screen, testCase.holds)
			}
		})
	}
}

// The places follow the page shown: on a project's page Tab goes round its own tabs, not the résumé's, and after q
// the résumé's again.
func TestThePlacesFollowThePageShown(test *testing.T) {
	cases := []struct {
		name string
		keys []string
		line string
	}{
		{name: "a project's page", keys: nil, line: ":why"},
		{name: "the résumé again", keys: []string{"q"}, line: ":experience"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, commanded(test, fixtureModel(test, commonWidth, commonHeight), "e tidy-tool"), append(testCase.keys, ":", "tab")...)
			if line := commandLineOf(screenOf(test, model)); line != testCase.line {
				test.Errorf("after :e tidy-tool, %q, a colon and Tab the command line = %q, want %q", testCase.keys, line, testCase.line)
			}
		})
	}
}

// Each name goes to one place: a tab titled as the window of keys is named takes the name, so Tab goes round it once.
func TestANameGoesToOnePlace(test *testing.T) {
	section := document.Section{Title: textOf("Help"), Blocks: []document.Block{document.Paragraph{Spans: textOf("the door")}}}
	composed := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, textOf("home")), Blocks: []document.Block{section}}
	model := New([]document.Document{composed}, fixtures.Site(fixtures.LoadedContent(test)), commonWidth, commonHeight, fixtures.Day())
	if line := commandLineOf(screenOf(test, pressed(test, model, ":", "tab", "tab"))); line != ":" {
		test.Errorf("on a page with a tab titled Help, after a colon and Tab twice the command line = %q, want %q", line, ":")
	}
}

// A tab whose title has several words, or capitals, goes by its title in small letters, the words joined as a path
// joins them.
func TestATabOfSeveralWordsGoesByThemJoined(test *testing.T) {
	section := document.Section{Title: textOf("Ways In"), Blocks: []document.Block{document.Paragraph{Spans: textOf("the door")}}}
	composed := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, textOf("home")), Blocks: []document.Block{section}}
	model := New([]document.Document{composed}, fixtures.Site(fixtures.LoadedContent(test)), commonWidth, commonHeight, fixtures.Day())
	if screen := screenOf(test, commanded(test, model, "ways-in")); !strings.Contains(screen, "the door") {
		test.Errorf("after :ways-in the screen =\n%s\nwant the tab titled %q", screen, "Ways In")
	}
}

// Tab, in normal mode or on the command line, chooses the first place and writes it on the command line, which Enter
// goes to; Shift+Tab in normal mode chooses the last.
func TestTabCompletesAPlace(test *testing.T) {
	cases := []struct {
		name string
		keys []string
		line string
	}{
		{name: "Tab on the command line", keys: []string{":", "tab"}, line: ":experience"},
		{name: "Tab in normal mode", keys: []string{"tab"}, line: ":experience"},
		{name: "Shift+Tab in normal mode", keys: []string{"shift+tab"}, line: ":help"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, fixtureModel(test, commonWidth, commonHeight), testCase.keys...)
			if line := commandLineOf(screenOf(test, model)); line != testCase.line {
				test.Errorf("after %q the command line = %q, want %q", testCase.keys, line, testCase.line)
			}
		})
	}
}

// Enter goes to the place Tab chose, and the footer marks its tab.
func TestEnterGoesToThePlaceTabChose(test *testing.T) {
	if footer := footerOf(screenOf(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), "tab", "enter"))); !strings.Contains(footer, "2:experience*") {
		test.Errorf("after Tab and Enter the footer = %q, want the tab gone to marked, %q", footer, "2:experience*")
	}
}

// Tab goes on through the places and Shift+Tab back, writing the one chosen on the command line; past either end the
// line holds what the reader typed again, and the next step goes on round, as Neovim's wildmenu does.
func TestTabGoesRoundThePlaces(test *testing.T) {
	cases := []struct {
		name string
		keys []string
		line string
	}{
		{name: "a second Tab the next", keys: []string{"tab", "tab"}, line: ":projects"},
		{name: "the last Tab help", keys: slices.Repeat([]string{"tab"}, lastPlace), line: ":help"},
		{name: "a Tab past the last what was typed", keys: slices.Repeat([]string{"tab"}, lastPlace+1), line: ":"},
		{name: "a Tab after that the first again", keys: slices.Repeat([]string{"tab"}, lastPlace+2), line: ":experience"},
		{name: "a second Shift+Tab the one before the last", keys: []string{"shift+tab", "shift+tab"}, line: ":contact"},
		{name: "Shift+Tab after Tab what was typed", keys: []string{"tab", "shift+tab"}, line: ":"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			screen := screenOf(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), append([]string{":"}, testCase.keys...)...))
			if line := commandLineOf(screen); line != testCase.line {
				test.Errorf("after a colon and %q the command line = %q, want %q", testCase.keys, line, testCase.line)
			}
		})
	}
}

// Over the command line a menu names what Tab goes round for what is typed, each by its kind, as Neovim's popup menu
// does: from the colon on, the places offered before anything is typed; a letter on, the places it starts, the
// projects' pages among them; after :e and a space, every project's page; and where the screen is too narrow for the
// kinds, the names alone, whole.
func TestTheMenuNamesWhatTabGoesRound(test *testing.T) {
	// a screen narrower than a project's slug, its kind and the room round them, wide enough for the slug alone
	const narrowerThanTheKinds = 16
	cases := []struct {
		name  string
		width int
		text  string
		holds []string
		lacks []string
	}{
		{name: "a colon alone", width: commonWidth, text: "", holds: []string{`experience +tab`, `contact +tab`, `help +keys`}, lacks: []string{`tidy-tool`}},
		{name: "a letter typed", width: commonWidth, text: "t", holds: []string{`tidy-tool +project`}, lacks: []string{`experience +tab`, `help +keys`}},
		{name: "an edit and a space", width: commonWidth, text: "e ", holds: []string{`tidy-tool +project`}, lacks: []string{`help +keys`}},
		{name: "a screen too narrow for the kinds", width: narrowerThanTheKinds, text: "e ", holds: []string{`tidy-tool`}, lacks: []string{`tidy-tool +pr`}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			above := aboveTheCommandLine(screenOf(test, typed(test, pressed(test, fixtureModel(test, testCase.width, commonHeight), ":"), testCase.text)))
			for _, want := range testCase.holds {
				if !regexp.MustCompile(want).MatchString(above) {
					test.Errorf("after :%s the screen above the command line =\n%s\nwant it to hold %q", testCase.text, above, want)
				}
			}
			for _, unwanted := range testCase.lacks {
				if regexp.MustCompile(unwanted).MatchString(above) {
					test.Errorf("after :%s the screen above the command line =\n%s\nwant it not to hold %q", testCase.text, above, unwanted)
				}
			}
		})
	}
}

// No menu stands over the page where nothing fits what is typed, or the screen has no row of the page to hold it: the
// screen above the command line is as it was before the colon, as many rows.
func TestNoMenuStandsWhereNothingFits(test *testing.T) {
	// a screen of the command line's row and the footer alone
	const footAlone = 2
	cases := []struct {
		name   string
		height int
		text   string
	}{
		{name: "nothing that fits", height: commonHeight, text: "x"},
		{name: "no row of the page", height: footAlone, text: ""},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := fixtureModel(test, commonWidth, testCase.height)
			before, after := screenOf(test, model), screenOf(test, typed(test, pressed(test, model, ":"), testCase.text))
			if aboveTheCommandLine(after) != aboveTheCommandLine(before) || strings.Count(after, lineEnd) != strings.Count(before, lineEnd) {
				test.Errorf("after :%s the screen =\n%s\nwant it above the command line as before the colon:\n%s", testCase.text, after, before)
			}
		})
	}
}

// A menu taller than the page has room for shows the item Tab chose, and stays where it stands while that is in
// sight, as Neovim's popup menu does: on a short screen, with room for two, the first two after one Tab, the last two
// after Tab to help and still after Shift+Tab back to contact, the two before them after Shift+Tab once more; and an
// edit shows it from its first again.
func TestTheMenuKeepsTheChoiceInSight(test *testing.T) {
	cases := []struct {
		name  string
		keys  []string
		holds []string
		lacks []string
	}{
		{name: "the first chosen", keys: []string{"tab"}, holds: []string{`experience +tab`, `projects +tab`}, lacks: []string{`help +keys`}},
		{name: "the last chosen", keys: slices.Repeat([]string{"tab"}, lastPlace), holds: []string{`contact +tab`, `help +keys`}, lacks: []string{`experience +tab`}},
		{name: "one back from the last", keys: append(slices.Repeat([]string{"tab"}, lastPlace), "shift+tab"), holds: []string{`contact +tab`, `help +keys`}, lacks: []string{`skills +tab`}},
		{name: "two back from the last", keys: append(slices.Repeat([]string{"tab"}, lastPlace), "shift+tab", "shift+tab"), holds: []string{`skills +tab`, `contact +tab`}, lacks: []string{`help +keys`}},
		{name: "an edit after the last", keys: append(slices.Repeat([]string{"tab"}, lastPlace), slices.Repeat([]string{"backspace"}, len("help"))...), holds: []string{`experience +tab`, `projects +tab`}, lacks: []string{`help +keys`}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			above := aboveTheCommandLine(screenOf(test, pressed(test, fixtureModel(test, commonWidth, shortHeight), append([]string{":"}, testCase.keys...)...)))
			for _, want := range testCase.holds {
				if !regexp.MustCompile(want).MatchString(above) {
					test.Errorf("after a colon and %q the screen above the command line =\n%s\nwant it to hold %q", testCase.keys, above, want)
				}
			}
			for _, unwanted := range testCase.lacks {
				if regexp.MustCompile(unwanted).MatchString(above) {
					test.Errorf("after a colon and %q the screen above the command line =\n%s\nwant it not to hold %q", testCase.keys, above, unwanted)
				}
			}
		})
	}
}

// A menu scrolled to its end follows the screen's height: grown taller, it shows every item again; grown shorter, it
// keeps the item chosen in sight, and stays there while Shift+Tab goes to one in sight.
func TestAMenuFollowsTheScreensHeight(test *testing.T) {
	// a screen with room for four rows of the page, under its crosses
	const fourRows = 8
	cases := []struct {
		name  string
		from  int
		to    int
		keys  []string
		holds []string
		lacks []string
	}{
		{name: "taller", from: shortHeight, to: commonHeight, keys: nil, holds: []string{`experience +tab`, `help +keys`}, lacks: nil},
		{name: "shorter", from: fourRows, to: shortHeight, keys: []string{"shift+tab"}, holds: []string{`contact +tab`, `help +keys`}, lacks: []string{`skills +tab`}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, fixtureModel(test, commonWidth, testCase.from), append([]string{":"}, slices.Repeat([]string{"tab"}, lastPlace)...)...)
			resized, _ := model.Update(tea.WindowSizeMsg{Width: commonWidth, Height: testCase.to})
			above := aboveTheCommandLine(screenOf(test, pressed(test, as[Model](test, resized), testCase.keys...)))
			for _, want := range testCase.holds {
				if !regexp.MustCompile(want).MatchString(above) {
					test.Errorf("after Tab to help, %d rows and %q the screen above the command line =\n%s\nwant it to hold %q", testCase.to, testCase.keys, above, want)
				}
			}
			for _, unwanted := range testCase.lacks {
				if regexp.MustCompile(unwanted).MatchString(above) {
					test.Errorf("after Tab to help, %d rows and %q the screen above the command line =\n%s\nwant it not to hold %q", testCase.to, testCase.keys, above, unwanted)
				}
			}
		})
	}
}

// Editing the line takes the choice away: the places are worked out anew, none chosen, so Tab chooses the first.
func TestAnEditTakesTheChoiceAway(test *testing.T) {
	model := pressed(test, fixtureModel(test, commonWidth, commonHeight), ":", "tab", "backspace", "tab")
	if line := commandLineOf(screenOf(test, model)); line != ":experience" {
		test.Errorf("after a colon, Tab, a backspace and Tab the command line = %q, want %q", line, ":experience")
	}
}
