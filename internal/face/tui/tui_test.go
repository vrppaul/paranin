package tui

import (
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/routes"
)

// A terminal of a common size, and one tall enough to show any tab whole
const (
	commonWidth  = 100
	commonHeight = 30
	tallHeight   = 400
	narrowWidth  = 40
	// a terminal too short for the roomy footer shows two lines of a page above the tight one, of three rows
	shortHeight = 5
	// a terminal just tall enough for the roomy footer, of seven rows, the row above the page and four lines of it
	roomyHeight = 12
	// a phone's terminal, narrower than the page with its margins, and the smallest that still draws
	phoneWidth     = 28
	phoneHeight    = 20
	smallestWidth  = 10
	smallestHeight = 3
)

// The résumé's first tab holds its facts as a brief beside the name, as neofetch prints a system's, each once, as
// written: the address, the subtitle and the place, the facts, the other ways to reach the author; the career's line
// under them, each role by its company, the one still held running to the present; then the summary.
func TestTheFirstTabHoldsTheBriefAndTheCareer(test *testing.T) {
	screen := screenOf(test, fixtureModel(test, commonWidth, tallHeight))
	cases := []struct {
		name string
		want string
	}{
		{name: "a fact", want: "5 years"},
		{name: "the address", want: "mira@example.test"},
		{name: "the subtitle and the place", want: "Software Engineer · Brno, Czechia"},
		{name: "another way", want: "github.test/mira"},
		{name: "the role still held", want: "Northwind"},
		{name: "a role before it", want: "Southwind"},
		{name: "the present", want: "present"},
		{name: "the summary", want: "Writes software for people who read it."},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if count := strings.Count(screen, testCase.want); count != 1 {
				test.Errorf("the first tab =\n%s\nholds %q %d times, want it once", screen, testCase.want, count)
			}
		})
	}
}

// A brief taller than the name's column stands whole beside it.
func TestABriefTallerThanTheNameStandsWhole(test *testing.T) {
	// more facts than the name's column has rows
	const keptFacts = 12
	facts := append(slices.Repeat([]document.Fact{{Label: "kept", Value: textOf("a value")}}, keptFacts), document.Fact{Label: "last", Value: textOf("the last value")})
	model := handResume(test, "home", commonWidth, tallHeight, document.Section{Title: textOf("Facts"), Blocks: []document.Block{document.Facts{List: facts}}})
	if page := aboveTheStatusLine(screenOf(test, model)); !strings.Contains(page, "the last value") {
		test.Errorf("the first tab =\n%s\nwant the brief's last fact, %q", page, "the last value")
	}
}

// The first tab opens whatever the career holds: a role that starts after the day the session opens, and a word for
// the present so wide it leaves the career's line no room.
func TestTheFirstTabOpensWhateverTheCareerHolds(test *testing.T) {
	// a page as wide as the word for the present the second case writes
	const presentWidth = 30
	brief := document.Section{Title: textOf("Facts"), Blocks: []document.Block{document.Facts{List: []document.Fact{{Label: "experience", Value: textOf("5 years")}}}}}
	cases := []struct {
		name  string
		role  document.Entry
		width int
		holds string
	}{
		{
			name:  "a role that starts after the day",
			role:  document.Entry{Period: document.Period{Start: knownMonth(2027, time.January), End: document.CurrentMonth{Text: "present"}}, Title: textOf("Futurewind")},
			width: commonWidth, holds: "2027",
		},
		{
			name:  "a word for the present as wide as the page",
			role:  document.Entry{Period: document.Period{Start: knownMonth(2023, time.April), End: document.CurrentMonth{Text: strings.Repeat("now", presentWidth/len("now"))}}, Title: textOf("Northwind")},
			width: presentWidth, holds: "5 years",
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			work := document.Section{Title: textOf("Work"), Blocks: []document.Block{testCase.role}}
			if page := aboveTheStatusLine(screenOf(test, handResume(test, "home", testCase.width, tallHeight, brief, work))); !strings.Contains(page, testCase.holds) {
				test.Errorf("the first tab =\n%s\nwant it to hold %q", page, testCase.holds)
			}
		})
	}
}

// The career's line lays out the roles, the entries of the first section that holds any, whose months are known: not
// a role whose start or end is still to fill, nor the schools of a section after it.
func TestTheCareersLineHoldsTheRolesOfKnownMonths(test *testing.T) {
	role := document.Entry{Period: document.Period{Start: knownMonth(2023, time.April), End: document.CurrentMonth{Text: "present"}}, Title: textOf("Northwind")}
	unfilled := document.Entry{Period: document.Period{Start: knownMonth(2020, time.January), End: document.GapMonth{Text: "[20XX]"}}, Title: textOf("Eastwind")}
	unstarted := document.Entry{Period: document.Period{Start: document.GapMonth{Text: "[20XX]"}, End: knownMonth(2019, time.June)}, Title: textOf("Westwind")}
	school := document.Entry{Period: document.Period{Start: knownMonth(2016, time.September), End: knownMonth(2020, time.June)}, Title: textOf("A School")}
	facts := document.Facts{List: []document.Fact{{Label: "experience", Value: textOf("5 years")}}}
	model := handResume(test, "home", commonWidth, tallHeight,
		document.Section{Title: textOf("Facts"), Blocks: []document.Block{facts}},
		document.Section{Title: textOf("Work"), Blocks: []document.Block{role, unfilled, unstarted}},
		document.Section{Title: textOf("Study"), Blocks: []document.Block{school}},
	)
	page := aboveTheStatusLine(screenOf(test, model))
	cases := []struct {
		name    string
		company string
		placed  bool
	}{
		{name: "a role", company: "Northwind", placed: true},
		{name: "a role whose end is still to fill", company: "Eastwind", placed: false},
		{name: "a role whose start is still to fill", company: "Westwind", placed: false},
		{name: "a school", company: "A School", placed: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if strings.Contains(page, testCase.company) != testCase.placed {
				test.Errorf("the first tab =\n%s\nwant %q on the career's line: %t", page, testCase.company, testCase.placed)
			}
		})
	}
}

// Only the résumé's first section, where it holds facts alone, joins the first tab as its brief: one that holds more,
// or no fact, one after another section, and one after a first section of facts keep a tab of their own.
func TestOnlyAFirstSectionOfFactsAloneJoinsTheFirstTab(test *testing.T) {
	facts := document.Facts{List: []document.Fact{{Label: "experience", Value: textOf("5 years")}}}
	paragraph := document.Paragraph{Spans: textOf("the door")}
	section := document.Section{Title: textOf("Facts"), Blocks: []document.Block{facts}}
	cases := []struct {
		name   string
		blocks []document.Block
		listed bool
	}{
		{name: "facts alone, first", blocks: []document.Block{section}, listed: false},
		{
			name:   "no fact, first",
			blocks: []document.Block{document.Section{Title: textOf("Facts"), Blocks: []document.Block{document.Facts{List: nil}}}},
			listed: true,
		},
		{
			name:   "facts beside a paragraph, first",
			blocks: []document.Block{document.Section{Title: textOf("Facts"), Blocks: []document.Block{facts, paragraph}}},
			listed: true,
		},
		{
			name:   "facts alone after another section",
			blocks: []document.Block{document.Section{Title: textOf("Work"), Blocks: []document.Block{paragraph}}, section},
			listed: true,
		},
		{
			name:   "facts alone after a first section of facts",
			blocks: []document.Block{document.Section{Title: textOf("Brief"), Blocks: []document.Block{facts}}, section},
			listed: true,
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := handResume(test, "home", commonWidth, commonHeight, testCase.blocks...)
			if places := footerLine(screenOf(test, model), footerPlaces); strings.Contains(places, "facts") != testCase.listed {
				test.Errorf("the places = %q, want a tab of the facts listed: %t", places, testCase.listed)
			}
		})
	}
}

// A project's page keeps a first section of facts alone as a tab of its own: only the résumé opens with a brief.
func TestAProjectsPageKeepsItsSectionOfFacts(test *testing.T) {
	facts := document.Facts{List: []document.Fact{{Label: "language", Value: textOf("Go")}}}
	project := document.Document{
		Head:   fixtures.HeadOf(document.Page{Kind: document.PageProject, Slug: "a-tool"}, textOf("a-tool")),
		Blocks: []document.Block{document.Section{Title: textOf("Facts"), Blocks: []document.Block{facts}}},
	}
	resume := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: ""}, textOf("home")), Blocks: nil}
	model := commanded(test, New([]document.Document{resume, project}, fixtureSite(fixtures.LoadedContent(test)), commonWidth, commonHeight, fixtureDay()), "e a-tool")
	if places := footerLine(screenOf(test, model), footerPlaces); !strings.Contains(places, "facts") {
		test.Errorf("on the project's page the places = %q, want its tab of the facts", places)
	}
}

// The author's name reads as text where the block font cannot draw it: on a page narrower than its banner, wrapped
// whole on a page narrower than the name, and where the font has not every letter of it.
func TestTheNameReadsAsTextWhereNoBannerFits(test *testing.T) {
	// a page narrower than the name written in one line
	const narrowerThanTheName = 10
	brief := document.Section{Title: textOf("Facts"), Blocks: []document.Block{document.Facts{List: []document.Fact{{Label: "experience", Value: textOf("5 years")}}}}}
	cases := []struct {
		name  string
		title string
		width int
		holds string
	}{
		{name: "a page narrower than the banner", title: "Mira Example", width: phoneWidth, holds: "Mira Example"},
		{name: "a page narrower than the name", title: "Mira Example", width: narrowerThanTheName, holds: "Example"},
		{name: "letters the font has not", title: "José Núñez", width: commonWidth, holds: "José Núñez"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if page := aboveTheStatusLine(screenOf(test, handResume(test, testCase.title, testCase.width, tallHeight, brief))); !strings.Contains(page, testCase.holds) {
				test.Errorf("at %d columns the page =\n%s\nwant the name's text %q", testCase.width, page, testCase.holds)
			}
		})
	}
}

// The prompt opens with the site's name: the host of its base address, not the address whole.
func TestThePromptNamesTheSiteByItsHost(test *testing.T) {
	if line := footerLine(screenOf(test, fixtureModel(test, wideWidth, commonHeight)), footerPrompt); !strings.HasPrefix(line, "example.test ~") {
		test.Errorf("the prompt = %q, want it to open with the host %q", line, "example.test")
	}
}

// The first screen is the head of the résumé, the status line names it by the site's word, and the places under the
// prompt name every section after it.
func TestTheFirstScreenShowsTheHeadAndEveryTab(test *testing.T) {
	screen := screenOf(test, fixtureModel(test, commonWidth, commonHeight))
	cases := []struct {
		name string
		want string
	}{
		{name: "the greeting", want: "Hello, my name is"},
		{name: "the title", want: "Software Engineer"},
		{name: "the opening's tab, by the site's word", want: "About"},
		{name: "the headline", want: "Builds small tools that last."},
		{name: "the first section's place", want: "experience"},
		{name: "the last section's place", want: "languages"},
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
		{name: "l goes to the next tab", keys: []string{"l"}, holds: "Kept the queue", lacks: "Builds small tools that last."},
		{name: "the right arrow goes to the next tab", keys: []string{"right"}, holds: "Kept the queue", lacks: "Builds small tools that last."},
		{name: "two steps go two tabs on", keys: []string{"l", "l"}, holds: "other-tool", lacks: "Kept the queue"},
		{name: "h goes round from the first tab to the last", keys: []string{"h"}, holds: "fluent", lacks: "Builds small tools that last."},
		{name: "the left arrow goes back", keys: []string{"l", "left"}, holds: "Builds small tools that last.", lacks: "Kept the queue"},
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
			model := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(toExperience(), testCase.keys...)...)
			screen := screenOf(test, model)
			if !strings.Contains(screen, testCase.holds) || strings.Contains(screen, testCase.lacks) {
				test.Errorf("after %q the screen =\n%s\nwant it to hold %q and not %q", testCase.keys, screen, testCase.holds, testCase.lacks)
			}
		})
	}
}

// The rim above the footer says how many lines of the tab lie below the page, as Neovim counts lines, and the key that
// scrolls to them; at the tab's end it says nothing.
func TestTheRimSaysWhatLiesBelow(test *testing.T) {
	cases := []struct {
		name  string
		keys  []string
		holds string
		lacks string
	}{
		{name: "the tab's start", keys: nil, holds: "more lines · j", lacks: "more line ·"},
		{name: "a line before the end", keys: []string{"G", "k"}, holds: "▾ 1 more line · j", lacks: "more lines"},
		{name: "the end", keys: []string{"G"}, holds: "", lacks: "more"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, fixtureModel(test, commonWidth, roomyHeight), append(toExperience(), testCase.keys...)...)
			rim := footerLine(screenOf(test, model), footerRim)
			if !strings.Contains(rim, testCase.holds) || strings.Contains(rim, testCase.lacks) {
				test.Errorf("after %q the rim = %q, want it to hold %q and not %q", testCase.keys, rim, testCase.holds, testCase.lacks)
			}
		})
	}
}

// A row faded above the rim still carries its link to a terminal that follows links: on a screen of nineteen rows the
// first tab's last row on the page is the fixture's GitHub, faded, as more lies below it.
func TestALinkOnAFadedRowStillLeads(test *testing.T) {
	const fadedLinkHeight = 19
	if screen := fixtureModel(test, commonWidth, fadedLinkHeight).View().Content; !strings.Contains(screen, "https://github.test/mira") {
		test.Errorf("the screen =\n%q\nwant the GitHub row's link to https://github.test/mira", screen)
	}
}

// Each tab keeps where it was scrolled to while the reader looks at another.
func TestEachTabKeepsItsPlace(test *testing.T) {
	model := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(toExperience(), "G", "l", "h")...)
	if screen := screenOf(test, model); !strings.Contains(screen, "one long winter") {
		test.Errorf("back on the experience tab the screen =\n%s\nwant it still at the end", screen)
	}
}

// The status line opens with the chosen tab's title at any width.
func TestTheChosenTabIsNamedAtAnyWidth(test *testing.T) {
	model := pressed(test, fixtureModel(test, narrowWidth, commonHeight), "h")
	if status := footerLine(screenOf(test, model), footerStatus); !strings.HasPrefix(status, "Languages") {
		test.Errorf("on the last tab the status line = %q, want it to open with %q", status, "Languages")
	}
}

// A role reads with its period, each month its year and its number, and its place.
func TestARoleReadsWithItsPeriod(test *testing.T) {
	screen := screenOf(test, pressed(test, fixtureModel(test, commonWidth, commonHeight), toExperience()...))
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
			model := pressed(test, fixtureModel(test, commonWidth, commonHeight), toExperience()...)
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
		Head:   fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: ""}, nil),
		Blocks: []document.Block{document.Facts{List: []document.Fact{{Label: "available", Value: nil}}}},
	}
	if screen := screenOf(test, New([]document.Document{composed}, fixtureSite(fixtures.LoadedContent(test)), commonWidth, commonHeight, fixtureDay())); !strings.Contains(screen, "available") {
		test.Errorf("the screen =\n%s\nwant the label %q", screen, "available")
	}
}

// q and Ctrl+C end the session on the résumé, and Ctrl+C on a project's page too.
func TestTheQuitKeysEndTheSession(test *testing.T) {
	cases := []struct {
		name   string
		opened string
		key    string
	}{
		{name: "q", opened: "", key: "q"},
		{name: "Ctrl+C", opened: "", key: "ctrl+c"},
		{name: "Ctrl+C on a project's page", opened: "e tidy-tool", key: "ctrl+c"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := commanded(test, fixtureModel(test, commonWidth, commonHeight), testCase.opened)
			_, command := model.Update(keyPress(test, testCase.key))
			if command == nil {
				test.Fatalf("%s gave no command, want the end of the session", testCase.key)
			}
			if message := command(); message != tea.Quit() {
				test.Errorf("%s gave %T, want the end of the session", testCase.key, message)
			}
		})
	}
}

// Every phrase of a page reaches the TUI on one tab or another: of the résumé, so a terminal reader reads every fact
// a browser reader reads, and of the health report and the list of skills :checkhealth and :Lazy show, each at any
// width. A word wider than a narrow terminal breaks across lines, so the phrases are read with every space and break
// taken out. The screen is read row by row, so the résumé is read where it is narrow: wider, its brief stands beside
// the name, and a row holds a piece of each. Its opening shows two phrases by its form (ADR-0003): the address heads
// the brief without its label, and the facts stand in the brief without their section's title.
func TestEveryPhraseReachesTheTUI(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	cases := []struct {
		name  string
		shown document.Document
		model Model
		// byForm are the phrases the page shows by its form, not as text
		byForm []string
	}{
		{name: "the résumé", shown: compose.Resume(snapshot), model: sessionModel(snapshot, narrowWidth, tallHeight), byForm: []string{"mail", "Facts"}},
		{
			name: "the health report", shown: compose.Health(snapshot),
			model: commanded(test, sessionModel(snapshot, commonWidth, tallHeight), "checkhealth"), byForm: nil,
		},
		{
			name: "the health report, narrow", shown: compose.Health(snapshot),
			model: commanded(test, sessionModel(snapshot, narrowWidth, tallHeight), "checkhealth"), byForm: nil,
		},
		{
			name: "the list of skills", shown: compose.Skills(snapshot),
			model: commanded(test, sessionModel(snapshot, commonWidth, tallHeight), "Lazy"), byForm: nil,
		},
		{
			name: "the list of skills, narrow", shown: compose.Skills(snapshot),
			model: commanded(test, sessionModel(snapshot, narrowWidth, tallHeight), "Lazy"), byForm: nil,
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model, screens := testCase.model, []string{}
			for range tabsOf(testCase.shown) {
				screens = append(screens, screenOf(test, model))
				model = pressed(test, model, "l")
			}
			read := strings.Join(strings.Fields(strings.Join(screens, " ")), "")
			for _, phrase := range document.Phrases(testCase.shown) {
				if !strings.Contains(read, strings.Join(strings.Fields(phrase), "")) && !slices.Contains(testCase.byForm, phrase) {
					test.Errorf("no tab of %s holds %q", testCase.name, phrase)
				}
			}
		})
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
	return sessionModel(fixtures.LoadedContent(test), width, height)
}

// sessionModel is a session on what a snapshot draws, as the SSH server opens one: its pages and the TUI's extras.
func sessionModel(snapshot *content.Snapshot, width, height int) Model {
	return New(append(compose.Documents(snapshot), compose.Extras(snapshot)...), fixtureSite(snapshot), width, height, fixtureDay())
}

// fixtureDay is the day the tests' sessions open on, the fixture's role still held running to it.
func fixtureDay() time.Time {
	return time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)
}

// handResume is a session on a résumé a test builds by hand: a head titled by a name, then blocks.
func handResume(test *testing.T, name string, width, height int, blocks ...document.Block) Model {
	composed := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: ""}, textOf(name)), Blocks: blocks}
	return New([]document.Document{composed}, fixtureSite(fixtures.LoadedContent(test)), width, height, fixtureDay())
}

// knownMonth is a month of a year as a period of a document a test builds by hand names it.
func knownMonth(year int, month time.Month) document.KnownMonth {
	return document.KnownMonth{Start: time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)}
}

// textOf is a text as the spans of a document, for a document a test builds by hand.
func textOf(text string) []document.Span {
	return []document.Span{document.Text{Text: text}}
}

// fixtureSite is the site of the fixture, at the address its profile gives.
func fixtureSite(snapshot *content.Snapshot) routes.Site {
	return routes.New(snapshot.Profile.Site.BaseURL)
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
		"y":         {Code: 'y', Text: "y"},
		"n":         {Code: 'n', Text: "n"},
		"N":         {Code: 'n', ShiftedCode: 'N', Mod: tea.ModShift, Text: "N"},
		"/":         {Code: '/', Text: "/"},
		":":         {Code: ':', Text: ":"},
		"?":         {Code: '?', Text: "?"},
		"space":     {Code: tea.KeySpace, Text: " "},
		"0":         {Code: '0', Text: "0"},
		"2":         {Code: '2', Text: "2"},
		"3":         {Code: '3', Text: "3"},
		"1":         {Code: '1', Text: "1"},
		"9":         {Code: '9', Text: "9"},
		"enter":     {Code: tea.KeyEnter},
		"esc":       {Code: tea.KeyEscape},
		"backspace": {Code: tea.KeyBackspace},
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

// typed is a model after a text typed, a key for each letter.
func typed(test *testing.T, model Model, text string) Model {
	for _, letter := range text {
		next, _ := model.Update(tea.KeyPressMsg{Code: letter, Text: string(letter)})
		model = as[Model](test, next)
	}
	return model
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
