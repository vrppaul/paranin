package tui

import (
	"slices"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/fixtures"
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

// A project's head names where its code lives as the page does: the owner and the name its repository's address
// gives, a link to the repository; where the address has no path to read them from, the address as written.
func TestAProjectsPageNamesTheOwnerAndNameOfItsRepository(test *testing.T) {
	const address = "https://github.test/mira/tidy"
	cases := []struct {
		name    string
		address string
		want    string
	}{
		{name: "an address with a path", address: address, want: "mira / tidy"},
		{name: "an address of a host alone", address: "https://tidy.test", want: "https://tidy.test"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := commanded(test, projectModel(test, testCase.address), "e tidy-tool")
			view := model.View().Content
			page := aboveTheFooter(screenOf(test, model))
			rows := strings.Split(page, lineEnd)
			if !slices.ContainsFunc(rows, func(row string) bool { return strings.TrimSpace(row) == testCase.want }) || !strings.Contains(view, ansi.SetHyperlink(testCase.address)) {
				test.Errorf("the project's page =\n%s\nwant a row of %q alone, a link to %q", page, testCase.want, testCase.address)
			}
		})
	}
}

// A project's page stands in one column in the page's order (DESIGN.md, The TUI): what the project is, its tagline
// and description; what gets it, its repository and each way to install it, a later one under its note; what tells
// about it, its stack, its releases and its licence; its sections, each side under its section's words, its clip by
// what it shows and its address, as a picture stands, its links by their names and lines, then its details; then its
// end, the heading over the other projects and the way to them all, each other project by its name, its tagline and
// its stack, and the author's mail.
func TestAProjectsPageStandsInThePagesOrder(test *testing.T) {
	inOrder := []string{
		"Tidies files.", "A tool that tidies files.", "mira / tidy",
		"go install example.test/tidy@latest", "On a machine with no Go", "curl -fsSL https://example.test/tidy.sh | sh",
		"Lua · Penlight · LPeg", "latest · changelog", "MIT", "Why", "Files were untidy.",
		"How it tidies", "tidy . reads each file once.", "Keeps the blank ones", "tidy . tidying a folder",
		"https://example.test/p/tidy-tool/tidying.mp4", "source", "where the code lives", "manual",
		"Flags", "--check reports what it would tidy", "flags", "each flag in full",
		"Other projects", "all projects", "tame-tool", "Tames files.", "Go", "mira@example.test",
	}
	page := aboveTheFooter(screenOf(test, commanded(test, sessionModel(fixtures.WithTwoProjectPages(test), commonWidth, tallHeight), "e tidy-tool")))
	rest := page
	for _, want := range inOrder {
		_, after, found := strings.Cut(rest, want)
		if !found {
			test.Fatalf("the project's page =\n%s\nwant %q after what stands before it, in the order %q", page, want, inOrder)
		}
		rest = after
	}
}

// A project's details open with a quiet rule after its story, and its end with one after its last detail, as the
// page's, whether the project has others or not.
func TestAProjectsDetailsAndEndOpenWithARule(test *testing.T) {
	cases := []struct {
		name    string
		content *content.Snapshot
		after   string
		before  string
	}{
		{name: "the details", content: fixtures.LoadedContent(test), after: "every flag it takes", before: "Flags"},
		{name: "the end of a project with others", content: fixtures.WithTwoProjectPages(test), after: "each flag in full", before: "Other projects"},
		{name: "the end of a project alone", content: fixtures.LoadedContent(test), after: "each flag in full", before: "mira@example.test"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			page := aboveTheFooter(screenOf(test, commanded(test, sessionModel(testCase.content, commonWidth, tallHeight), "e tidy-tool")))
			_, rest, _ := strings.Cut(page, testCase.after)
			between, _, _ := strings.Cut(rest, testCase.before)
			if !slices.ContainsFunc(strings.Split(between, "\n"), isRule) {
				test.Errorf("the project's page =\n%s\nwant a rule between %q and %q", page, testCase.after, testCase.before)
			}
		})
	}
}

// A section's links read as the page's contents (DESIGN.md, Contents): each on a row of its own, its name, then a
// colon and its line, with no mark before it, so under the section's points they read as no more of them.
func TestASectionsLinksReadAsContents(test *testing.T) {
	page := aboveTheFooter(screenOf(test, commanded(test, fixtureModel(test, commonWidth, tallHeight), "e tidy-tool")))
	rows := strings.Split(page, "\n")
	for _, want := range []string{"source: where the code lives", "manual: every flag it takes"} {
		if !slices.ContainsFunc(rows, func(row string) bool { return strings.TrimSpace(row) == want }) {
			test.Errorf("the project's page =\n%s\nwant a row %q alone", page, want)
		}
	}
}

// isNameRow is whether a screen's row holds the fixture project's name alone, as text, the site's cursor after it.
func isNameRow(row string) bool {
	return strings.TrimSuffix(strings.TrimSpace(row), cursorBar) == "tidy-tool"
}

// isRule is whether a screen's row is a rule alone: a line across and nothing else.
func isRule(row string) bool {
	return strings.Contains(row, frameAcross) && strings.Trim(row, " "+frameAcross) == ""
}

// A tile with nothing set into its top edge, as another project at a project's end, draws its top edge whole, with no
// gap in the line.
func TestATileWithNothingInItsEdgeDrawsTheEdgeWhole(test *testing.T) {
	page := aboveTheFooter(screenOf(test, commanded(test, sessionModel(fixtures.WithTwoProjectPages(test), commonWidth, tallHeight), "e tidy-tool")))
	_, others, _ := strings.Cut(page, "all projects")
	edgeStart := strings.Index(others, frameAcross)
	lineStart := strings.LastIndex(others[:edgeStart], "\n") + 1
	edge, _, _ := strings.Cut(others[lineStart:], "\n")
	if strings.Contains(strings.TrimSpace(edge), " ") {
		test.Errorf("the project's page =\n%s\nwant the other projects' top edge whole, not %q", page, edge)
	}
}

// A project's page stands under the project's name, in the titles' font where the page has the room for it, as text
// on a page too narrow for that. A banner is a look, so the cases read only whether the name stands as text, on a row
// of its own, as a title does, apart from the addresses on the page that hold it.
func TestAProjectsPageStandsUnderItsName(test *testing.T) {
	cases := []struct {
		name       string
		width      int
		nameAsText bool
	}{
		{name: "a page with room for the banner", width: commonWidth, nameAsText: false},
		{name: "a page as narrow as a phone", width: phoneWidth, nameAsText: true},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			page := aboveTheFooter(screenOf(test, commanded(test, fixtureModel(test, testCase.width, tallHeight), "e tidy-tool")))
			if slices.ContainsFunc(strings.Split(page, "\n"), isNameRow) != testCase.nameAsText {
				test.Errorf("the project's page =\n%s\nwant its name %q on it as text: %t", page, "tidy-tool", testCase.nameAsText)
			}
		})
	}
}

// projectModel is a session on the fixture content, the code of each of its projects living at an address, on a screen
// tall enough for a project's whole page, so no scrollbar stands beside its rows.
func projectModel(test *testing.T, address string) Model {
	snapshot := *fixtures.LoadedContent(test)
	snapshot.Profile.Projects = slices.Clone(snapshot.Profile.Projects)
	for index := range snapshot.Profile.Projects {
		snapshot.Profile.Projects[index].Repo = address
	}
	return sessionModel(&snapshot, commonWidth, tallHeight)
}

// A page whose head gives its opening no title of its own titles its first tab by the head's title: a project's page
// by the project's name, which the footer names it by.
func TestAFirstTabWithNoTitleOfItsOwnTakesTheHeads(test *testing.T) {
	screen := screenOf(test, commanded(test, fixtureModel(test, commonWidth, commonHeight), "e tidy-tool"))
	if footer := footerOf(screen); !strings.Contains(footer, "1:tidy-tool*") {
		test.Errorf("after :e tidy-tool the footer = %q, want the first tab named %q", footer, "1:tidy-tool*")
	}
}

// What the window of keys says q does follows the page shown: on a project's page q closes it, on the résumé q
// quits.
func TestTheFrameSaysWhatTheQuitKeyDoesOnThePageShown(test *testing.T) {
	cases := []struct {
		name   string
		opened string
		want   string
	}{
		{name: "on the résumé", opened: "", want: "q     quit"},
		{name: "on a project's page", opened: "e tidy-tool", want: "q     close"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, commanded(test, fixtureModel(test, commonWidth, commonHeight), testCase.opened), "?")
			if screen := screenOf(test, model); !strings.Contains(screen, testCase.want) {
				test.Errorf("after :%s and ? the screen =\n%s\nwant %q", testCase.opened, screen, testCase.want)
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

// On a project's page, the health report or the list of skills, q and the commands that close a window go back to
// the résumé, and the session goes on.
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
		{name: "q on the list of skills", opening: "Lazy", typing: "", last: "q"},
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
			if below := atTheFoot(screen); !strings.Contains(below, testCase.want) || !strings.Contains(screen, resumeHeadline) {
				test.Errorf("after :%s the screen =\n%s\nwant %q, and the résumé still shown", testCase.text, screen, testCase.want)
			}
		})
	}
}

// Tab goes round the projects' pages in the profile's order: after :e and a space those whose slugs start with what
// follows, and among the places a start typed with no command names; after a command that takes no argument it leaves
// the line as typed.
func TestTabGoesRoundTheProjectsPagesInTheProfilesOrder(test *testing.T) {
	cases := []struct {
		name string
		text string
		keys []string
		line string
	}{
		{name: "the first page after :e", text: "e ", keys: []string{"tab"}, line: ":e tidy-tool"},
		{name: "the second page after it", text: "e ", keys: []string{"tab", "tab"}, line: ":e tame-tool"},
		{name: "the pages a start names", text: "e ti", keys: []string{"tab"}, line: ":e tidy-tool"},
		{name: "what was typed after the one page a start names", text: "e ti", keys: []string{"tab", "tab"}, line: ":e ti"},
		{name: "spaces around", text: " e  ti", keys: []string{"tab"}, line: ":e tidy-tool"},
		{name: "the second place a start names", text: "t", keys: []string{"tab", "tab"}, line: ":tame-tool"},
		{name: "a project of no page", text: "e other", keys: []string{"tab"}, line: ":e other"},
		{name: "none after :q", text: "q ", keys: []string{"tab"}, line: ":q"},
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

// twoPagesModel is a session on the fixture's content with a second project of a page (fixtures.WithTwoProjectPages).
func twoPagesModel(test *testing.T) Model {
	return sessionModel(fixtures.WithTwoProjectPages(test), commonWidth, commonHeight)
}

// Tab after :e and a space writes the first page whose slug fits, which Enter opens.
func TestTabCompletesAPage(test *testing.T) {
	model := pressed(test, typed(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), ":"), "e "), "tab")
	if line := commandLineOf(screenOf(test, model)); line != ":e tidy-tool" {
		test.Errorf("after :e, a space and Tab the command line = %q, want %q", line, ":e tidy-tool")
	}
	if screen := screenOf(test, pressed(test, model, "enter")); !strings.Contains(screen, projectTagline) {
		test.Errorf("after :e, a space, Tab and Enter the screen =\n%s\nwant the project's page", screen)
	}
}

// commanded is a model after a place or a command typed on the command line and run.
func commanded(test *testing.T, model Model, text string) Model {
	return pressed(test, typed(test, pressed(test, model, ":"), text), "enter")
}
