package tui

import (
	"strings"
	"testing"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
)

// What the fixture's résumé says under the name, and what the page of its project tidy-tool says under its own
const (
	resumeHeadline = "Builds small tools that last."
	projectTagline = "Tidies files."
)

// :e or :edit and a project's slug, spaces between or not, opens the project's own page in place of the résumé.
func TestEditOpensAProjectsPage(test *testing.T) {
	for _, text := range []string{"e tidy-tool", "edit tidy-tool", "e   tidy-tool "} {
		test.Run(text, func(test *testing.T) {
			screen := screenOf(test, commanded(test, fixtureModel(test, commonWidth, commonHeight), text))
			if !strings.Contains(screen, projectTagline) || strings.Contains(screen, resumeHeadline) {
				test.Errorf("after :%s the screen =\n%s\nwant the project's page, not the résumé", text, screen)
			}
		})
	}
}

// What the frame says q does follows the page shown: on a project's page q closes it, on the résumé q quits. The
// status line, the window of keys and the window of commands each say it.
func TestTheFrameSaysWhatTheQuitKeyDoesOnThePageShown(test *testing.T) {
	cases := []struct {
		name   string
		opened string
		keys   []string
		want   string
	}{
		{name: "the status line on the résumé", opened: "", keys: nil, want: "q quit"},
		{name: "the status line on a project's page", opened: "e tidy-tool", keys: nil, want: "q close"},
		{name: "the window of keys on a project's page", opened: "e tidy-tool", keys: []string{"?"}, want: "q     close"},
		{name: "the window of commands on a project's page", opened: "e tidy-tool", keys: []string{":"}, want: ":q  close"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, commanded(test, fixtureModel(test, commonWidth, commonHeight), testCase.opened), testCase.keys...)
			if screen := screenOf(test, model); !strings.Contains(screen, testCase.want) {
				test.Errorf("after :%s and %q the screen =\n%s\nwant %q", testCase.opened, testCase.keys, screen, testCase.want)
			}
		})
	}
}

// Each page keeps where its reader stood, its tab and how far down it they scrolled: q on a project's page goes back
// to the résumé as it was left, and :e opens the project's page again as it was left.
func TestEachPageKeepsWhereItsReaderStood(test *testing.T) {
	resume := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(toExperience(), "j", "j")...)
	project := pressed(test, commanded(test, resume, "e tidy-tool"), "j")
	if back := pressed(test, project, "q"); screenOf(test, back) != screenOf(test, resume) {
		test.Errorf("after a tab, a scroll, :e tidy-tool and q the screen =\n%s\nwant the résumé as it was left:\n%s",
			screenOf(test, back), screenOf(test, resume))
	}
	again := commanded(test, pressed(test, project, "q"), "e tidy-tool")
	if screenOf(test, again) != screenOf(test, project) {
		test.Errorf("after :e tidy-tool, a scroll, q and :e tidy-tool again the screen =\n%s\nwant the page as it was left:\n%s",
			screenOf(test, again), screenOf(test, project))
	}
}

// On a project's page, or the health report, q and the commands that close a window go back to the résumé, and the
// session goes on.
func TestClosingAPageGoesBackToTheResume(test *testing.T) {
	cases := []struct {
		name    string
		opening string
		typing  string
		last    string
	}{
		{name: "q", opening: "e tidy-tool", typing: "", last: "q"},
		{name: ":q", opening: "e tidy-tool", typing: ":q", last: "enter"},
		{name: ":quit", opening: "e tidy-tool", typing: ":quit", last: "enter"},
		{name: ":q!", opening: "e tidy-tool", typing: ":q!", last: "enter"},
		{name: "q on the health report", opening: "checkhealth", typing: "", last: "q"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			opened := typed(test, commanded(test, fixtureModel(test, commonWidth, commonHeight), testCase.opening), testCase.typing)
			next, command := opened.Update(keyPress(test, testCase.last))
			if command != nil {
				test.Fatalf("after :%s and %s a command came, want the session going on", testCase.opening, testCase.name)
			}
			if screen := screenOf(test, as[Model](test, next)); !strings.Contains(screen, resumeHeadline) {
				test.Errorf("after :%s and %s the screen =\n%s\nwant the résumé", testCase.opening, testCase.name, screen)
			}
		})
	}
}

// The status line names :e where the session holds a project's page to open; a session of the résumé alone names
// none.
func TestTheStatusLineNamesEditWhereAPageIsThere(test *testing.T) {
	snapshot := fixtureSnapshot(test)
	cases := []struct {
		name  string
		model Model
		named bool
	}{
		{name: "a project's page", model: fixtureModel(test, commonWidth, commonHeight), named: true},
		{name: "the résumé alone", model: New([]document.Document{compose.Resume(snapshot)}, fixtureSite(snapshot), commonWidth, commonHeight), named: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			below := belowThePage(screenOf(test, testCase.model))
			if named := strings.Contains(below, ":e pages"); named != testCase.named {
				test.Errorf("on a session of %s the status and command lines = %q, want :e pages named: %t", testCase.name, below,
					testCase.named)
			}
		})
	}
}

// :e finds only a project with a page of its own, and names back what it does not find with Neovim's error.
func TestEditFindsOnlyAProjectsPage(test *testing.T) {
	cases := []struct {
		name string
		text string
		want string
	}{
		{name: "no such page", text: "e nope", want: `E447: Can't find file "nope" in path`},
		{name: "a project of no page", text: "e other-tool", want: `E447: Can't find file "other-tool" in path`},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			screen := screenOf(test, commanded(test, fixtureModel(test, commonWidth, commonHeight), testCase.text))
			if below := belowThePage(screen); !strings.Contains(below, testCase.want) || !strings.Contains(screen, resumeHeadline) {
				test.Errorf("after :%s the screen =\n%s\nwant %q, and the résumé still shown", testCase.text, screen, testCase.want)
			}
		})
	}
}

// After :e and a space the window of commands lists the projects with a page whose slugs start with what follows,
// in the order of their letters, which Tab goes through; after a command that takes no argument it lists nothing,
// and Tab leaves the line as typed. The session holds a second project's page, which sorts first.
func TestTabGoesRoundThePagesAfterEdit(test *testing.T) {
	cases := []struct {
		name string
		text string
		keys []string
		line string
	}{
		{name: "the first page after :e", text: "e ", keys: []string{"tab"}, line: " :e another-tool"},
		{name: "the second page after it", text: "e ", keys: []string{"tab", "tab"}, line: " :e tidy-tool"},
		{name: "the pages a start names", text: "e ti", keys: []string{"tab"}, line: " :e tidy-tool"},
		{name: "what was typed after the one page a start names", text: "e ti", keys: []string{"tab", "tab"}, line: " :e ti"},
		{name: "spaces around", text: " e  ti", keys: []string{"tab"}, line: " :e tidy-tool"},
		{name: "a project of no page", text: "e other", keys: []string{"tab"}, line: " :e other"},
		{name: "none after :q", text: "q ", keys: []string{"tab"}, line: " :q"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := typed(test, pressed(test, twoPagesModel(test), ":"), testCase.text)
			if line := commandLineOf(screenOf(test, pressed(test, model, testCase.keys...))); line != testCase.line {
				test.Errorf("after :%s and %q the command line = %q, want %q", testCase.text, testCase.keys, line, testCase.line)
			}
		})
	}
}

// twoPagesModel is a session on the fixture's documents and a second project's page, another-tool, of a title alone.
func twoPagesModel(test *testing.T) Model {
	another := document.Document{
		Head:   document.Head{Page: document.Page{Kind: document.PageProject, Slug: "another-tool"}, Title: []document.Span{document.Text{Text: "another-tool"}}, Subtitle: nil, Place: "", Lead: nil, Ways: nil},
		Blocks: nil,
	}
	snapshot := fixtureSnapshot(test)
	documents := append(compose.Documents(snapshot), another)
	return New(documents, fixtureSite(snapshot), commonWidth, commonHeight)
}

// Tab after :e and a space writes the page the window lists first, which Enter opens.
func TestTabCompletesAPage(test *testing.T) {
	model := pressed(test, typed(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), ":"), "e "), "tab")
	if line := commandLineOf(screenOf(test, model)); line != " :e tidy-tool" {
		test.Errorf("after :e, a space and Tab the command line = %q, want %q", line, " :e tidy-tool")
	}
	if screen := screenOf(test, pressed(test, model, "enter")); !strings.Contains(screen, projectTagline) {
		test.Errorf("after :e, a space, Tab and Enter the screen =\n%s\nwant the project's page", screen)
	}
}

// commanded is a model after a command typed on the command line and run.
func commanded(test *testing.T, model Model, text string) Model {
	return pressed(test, typed(test, pressed(test, model, ":"), text), "enter")
}
