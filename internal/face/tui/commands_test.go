package tui

import (
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// The prompt as it opens, on the fixture's site, before anything is typed
const emptyPrompt = "example.test ~ ❯"

// A colon opens the prompt, where what the reader types shows after it.
func TestTheColonOpensThePrompt(test *testing.T) {
	model := typed(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), ":"), "qa")
	if line := footerLine(screenOf(test, model), footerPrompt); line != emptyPrompt+" qa" {
		test.Errorf("after a colon and qa the prompt = %q, want %q", line, emptyPrompt+" qa")
	}
}

// Each of Neovim's ways to quit, typed on the prompt, ends the session on the résumé, as q does, spaces around it or
// not; :qa ends it on a project's page too.
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
			if line := footerLine(screen, footerPrompt); line != testCase.want || !strings.Contains(screen, resumeHeadline) {
				test.Errorf("after :%s the screen =\n%s\nwant the prompt to say %q, and the résumé still shown", testCase.text, screen, testCase.want)
			}
		})
	}
}

// What names no place and no command is named back with Neovim's error, alone on the prompt's row as Neovim's command
// line says it, and the keys go back to normal, where no cursor shows; the start of a command's name is no command.
func TestAnUnknownCommandIsNamedBack(test *testing.T) {
	for _, name := range []string{"nope", "qui"} {
		test.Run(name, func(test *testing.T) {
			model := pressed(test, typed(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), ":"), name), "enter")
			want := "E492: Not an editor command: " + name
			if line := footerLine(screenOf(test, model), footerPrompt); line != want || model.View().Cursor != nil {
				test.Errorf("after :%s the prompt = %q and the cursor %+v, want %q and no cursor", name, line, model.View().Cursor, want)
			}
		})
	}
}

// Esc and Ctrl+C leave the prompt, as do a backspace and Enter on nothing, running nothing: the prompt is back as it
// opens, with the site's cursor and no terminal's cursor.
func TestThePromptIsLeftUnrun(test *testing.T) {
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
			if line := footerLine(screenOf(test, model), footerPrompt); line != emptyPrompt+" "+cursorBar || model.View().Cursor != nil {
				test.Errorf("after :%s and %q the prompt = %q and the cursor %+v, want %q and no cursor", testCase.text, testCase.keys,
					line, model.View().Cursor, emptyPrompt+" "+cursorBar)
			}
		})
	}
}

// Under the prompt the places it goes to are listed: before anything is typed, every tab but the one shown and the
// window of keys; once the reader types, every place whose name starts with it, a project's page among them; where
// nothing does, what Enter and Esc do.
func TestThePlacesListWhatIsTyped(test *testing.T) {
	cases := []struct {
		name  string
		text  string
		holds []string
		lacks []string
	}{
		{name: "nothing typed", text: "", holds: []string{"experience", "contact", "help"}, lacks: []string{"about", "tidy-tool"}},
		{name: "the start of a tab", text: "e", holds: []string{"experience"}, lacks: []string{"projects", "help"}},
		{name: "the start of a project's page", text: "ti", holds: []string{"tidy-tool"}, lacks: []string{"experience"}},
		{name: "the start of nothing", text: "x", holds: []string{"↵ run", "esc cancel"}, lacks: []string{"experience"}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			places := footerLine(screenOf(test, typed(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), ":"), testCase.text)), footerPlaces)
			for _, want := range testCase.holds {
				if !strings.Contains(places, want) {
					test.Errorf("after :%s the places = %q, want them to hold %q", testCase.text, places, want)
				}
			}
			for _, unwanted := range testCase.lacks {
				if strings.Contains(places, unwanted) {
					test.Errorf("after :%s the places = %q, want them not to hold %q", testCase.text, places, unwanted)
				}
			}
		})
	}
}

// A second colon opens the prompt on nothing typed, and the places listed before anything is typed.
func TestASecondColonOpensOnNothing(test *testing.T) {
	screen := screenOf(test, pressed(test, typed(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), ":"), "x"), "esc", ":"))
	if line, places := footerLine(screen, footerPrompt), footerLine(screen, footerPlaces); line != emptyPrompt || !strings.Contains(places, "experience") {
		test.Errorf("after :x, Esc and a colon the prompt = %q and the places %q, want %q and every tab listed", line, places, emptyPrompt)
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

// The places follow the page shown: a project's page lists its own tabs, not the résumé's, and q brings the
// résumé's back.
func TestThePlacesFollowThePageShown(test *testing.T) {
	cases := []struct {
		name  string
		keys  []string
		holds string
		lacks string
	}{
		{name: "a project's page", keys: nil, holds: "why", lacks: "experience"},
		{name: "the résumé again", keys: []string{"q"}, holds: "experience", lacks: "why"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, commanded(test, fixtureModel(test, commonWidth, commonHeight), "e tidy-tool"), append(testCase.keys, ":")...)
			if places := footerLine(screenOf(test, model), footerPlaces); !strings.Contains(places, testCase.holds) || strings.Contains(places, testCase.lacks) {
				test.Errorf("after :e tidy-tool, %q and a colon the places = %q, want %q and not %q", testCase.keys, places, testCase.holds, testCase.lacks)
			}
		})
	}
}

// Where the places listed before anything is typed do not fit, the last of them, help, stays in sight, a mark
// standing for those that give way.
func TestHelpStaysInSightOnANarrowScreen(test *testing.T) {
	places := footerLine(screenOf(test, fixtureModel(test, narrowWidth, commonHeight)), footerPlaces)
	if !strings.HasSuffix(places, "…   help") {
		test.Errorf("at %d columns the places = %q, want them to end in a mark and help", narrowWidth, places)
	}
}

// Each name goes to one place: a tab titled as the window of keys is named takes the name, listed once.
func TestANameGoesToOnePlace(test *testing.T) {
	section := document.Section{Title: textOf("Help"), Blocks: []document.Block{document.Paragraph{Spans: textOf("the door")}}}
	composed := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: ""}, textOf("home")), Blocks: []document.Block{section}}
	model := New([]document.Document{composed}, fixtures.Site(fixtures.LoadedContent(test)), commonWidth, commonHeight, fixtures.Day())
	if places := footerLine(screenOf(test, model), footerPlaces); strings.Count(places, "help") != 1 {
		test.Errorf("on a page with a tab titled Help the places = %q, want help listed once", places)
	}
}

// A tab whose title has several words, or capitals, goes by its title in small letters, the words joined as a path
// joins them.
func TestATabOfSeveralWordsGoesByThemJoined(test *testing.T) {
	section := document.Section{Title: textOf("Ways In"), Blocks: []document.Block{document.Paragraph{Spans: textOf("the door")}}}
	composed := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: ""}, textOf("home")), Blocks: []document.Block{section}}
	model := New([]document.Document{composed}, fixtures.Site(fixtures.LoadedContent(test)), commonWidth, commonHeight, fixtures.Day())
	if screen := screenOf(test, commanded(test, model, "ways-in")); !strings.Contains(screen, "the door") {
		test.Errorf("after :ways-in the screen =\n%s\nwant the tab titled %q", screen, "Ways In")
	}
}

// Tab, in normal mode or on the prompt, chooses the first place listed and writes it on the prompt, which Enter goes
// to; Shift+Tab in normal mode chooses the last.
func TestTabCompletesAPlace(test *testing.T) {
	cases := []struct {
		name string
		keys []string
		line string
	}{
		{name: "Tab on the prompt", keys: []string{":", "tab"}, line: emptyPrompt + " experience"},
		{name: "Tab in normal mode", keys: []string{"tab"}, line: emptyPrompt + " experience"},
		{name: "Shift+Tab in normal mode", keys: []string{"shift+tab"}, line: emptyPrompt + " help"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, fixtureModel(test, commonWidth, commonHeight), testCase.keys...)
			if line := footerLine(screenOf(test, model), footerPrompt); line != testCase.line {
				test.Errorf("after %q the prompt = %q, want %q", testCase.keys, line, testCase.line)
			}
		})
	}
}

// Enter goes to the place Tab chose, and the status line opens with its title.
func TestEnterGoesToThePlaceTabChose(test *testing.T) {
	if status := footerLine(screenOf(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), "tab", "enter")), footerStatus); !strings.HasPrefix(status, "Experience") {
		test.Errorf("after Tab and Enter the status line = %q, want it to open with the tab gone to, %q", status, "Experience")
	}
}

// Tab goes on through the places listed and Shift+Tab back, writing the one chosen on the prompt; past either end the
// prompt holds what the reader typed again, and the next step goes on round, as Neovim's wildmenu does. On the first
// tab the fixture lists four other tabs and help.
func TestTabGoesRoundThePlaces(test *testing.T) {
	// the presses of Tab that reach the last place: each other tab, then help
	const lastPlace = 5
	cases := []struct {
		name string
		keys []string
		line string
	}{
		{name: "a second Tab the next", keys: []string{"tab", "tab"}, line: emptyPrompt + " projects"},
		{name: "the last Tab help", keys: slices.Repeat([]string{"tab"}, lastPlace), line: emptyPrompt + " help"},
		{name: "a Tab past the last what was typed", keys: slices.Repeat([]string{"tab"}, lastPlace+1), line: emptyPrompt},
		{name: "a Tab after that the first again", keys: slices.Repeat([]string{"tab"}, lastPlace+2), line: emptyPrompt + " experience"},
		{name: "a second Shift+Tab the one before the last", keys: []string{"shift+tab", "shift+tab"}, line: emptyPrompt + " contact"},
		{name: "Shift+Tab after Tab what was typed", keys: []string{"tab", "shift+tab"}, line: emptyPrompt},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			screen := screenOf(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), append([]string{":"}, testCase.keys...)...))
			if line := footerLine(screen, footerPrompt); line != testCase.line {
				test.Errorf("after a colon and %q the prompt = %q, want %q", testCase.keys, line, testCase.line)
			}
		})
	}
}

// Editing the line takes the choice away: the places are listed anew, none chosen, so Tab chooses the first of them.
func TestAnEditTakesTheChoiceAway(test *testing.T) {
	model := pressed(test, fixtureModel(test, commonWidth, commonHeight), ":", "tab", "backspace", "tab")
	if line := footerLine(screenOf(test, model), footerPrompt); line != emptyPrompt+" experience" {
		test.Errorf("after a colon, Tab, a backspace and Tab the prompt = %q, want %q", line, emptyPrompt+" experience")
	}
}
