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
	"github.com/vrppaul/paranin/internal/look"
)

// A terminal of a common size, and one tall enough to show any tab whole
const (
	commonWidth  = 100
	commonHeight = 30
	tallHeight   = 400
	narrowWidth  = 40
	// a terminal too short for a row of nothing above the page shows two lines of it above the command line's row
	shortHeight = 4
	// a phone's terminal, narrower than the page with its margins, and the smallest that still draws
	phoneWidth     = 28
	phoneHeight    = 20
	smallestWidth  = 10
	smallestHeight = 3
	// a page as wide as the label Email, its gap and one column
	oneColumnLeft = 8
)

// The résumé's first tab holds its facts as a brief beside the name, as neofetch prints a system's, each once, as
// written: the address, the subtitle and the place, the facts; the career's line under them, each role by its
// company, the one still held running to the present; then the summary.
func TestTheFirstTabHoldsTheBriefAndTheCareer(test *testing.T) {
	screen := screenOf(test, fixtureModel(test, commonWidth, tallHeight))
	cases := []struct {
		name string
		want string
	}{
		{name: "a fact", want: "5 years"},
		{name: "the address", want: "mira@example.test"},
		{name: "the subtitle and the place", want: "Software Engineer · Brno, Czechia"},
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
	if page := aboveTheFooter(screenOf(test, model)); !strings.Contains(page, "the last value") {
		test.Errorf("the first tab =\n%s\nwant the brief's last fact, %q", page, "the last value")
	}
}

// The blocks the first section holds beside its facts, before them and after them, stand on the first tab under the
// brief.
func TestTheBlocksBesideTheBriefStandOnTheFirstTab(test *testing.T) {
	section := document.Section{Title: textOf("Facts"), Blocks: []document.Block{
		document.Paragraph{Spans: textOf("said before")},
		document.Facts{List: []document.Fact{{Label: "experience", Value: textOf("5 years")}}},
		document.Paragraph{Spans: textOf("said after")},
	}}
	page := aboveTheFooter(screenOf(test, handResume(test, "home", commonWidth, tallHeight, section)))
	for _, want := range []string{"said before", "said after"} {
		if !strings.Contains(page, want) {
			test.Errorf("the first tab =\n%s\nholds no %q", page, want)
		}
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
			role:  document.Entry{Kind: document.EntryRole, Period: document.Period{Start: knownMonth(2027, time.January), End: document.CurrentMonth{Text: "present"}}, Title: textOf("Futurewind")},
			width: commonWidth, holds: "2027",
		},
		{
			name:  "a word for the present as wide as the page",
			role:  document.Entry{Kind: document.EntryRole, Period: document.Period{Start: knownMonth(2023, time.April), End: document.CurrentMonth{Text: strings.Repeat("now", presentWidth/len("now"))}}, Title: textOf("Northwind")},
			width: presentWidth, holds: "5 years",
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			work := document.Section{Title: textOf("Work"), Blocks: []document.Block{testCase.role}}
			if page := aboveTheFooter(screenOf(test, handResume(test, "home", testCase.width, tallHeight, brief, work))); !strings.Contains(page, testCase.holds) {
				test.Errorf("the first tab =\n%s\nwant it to hold %q", page, testCase.holds)
			}
		})
	}
}

// The career's line lays out the roles, the entries of a role's kind in any top-level section, whose months are known:
// not a role whose start or end is still to fill, nor a school, even one in a section before the roles.
func TestTheCareersLineHoldsTheRolesOfKnownMonths(test *testing.T) {
	role := document.Entry{Kind: document.EntryRole, Period: document.Period{Start: knownMonth(2023, time.April), End: document.CurrentMonth{Text: "present"}}, Title: textOf("Northwind")}
	unfilled := document.Entry{Kind: document.EntryRole, Period: document.Period{Start: knownMonth(2020, time.January), End: document.GapMonth{Text: "[20XX]"}}, Title: textOf("Eastwind")}
	unstarted := document.Entry{Kind: document.EntryRole, Period: document.Period{Start: document.GapMonth{Text: "[20XX]"}, End: knownMonth(2019, time.June)}, Title: textOf("Westwind")}
	school := document.Entry{Kind: document.EntrySchool, Period: document.Period{Start: knownMonth(2016, time.September), End: knownMonth(2020, time.June)}, Title: textOf("A School")}
	facts := document.Facts{List: []document.Fact{{Label: "experience", Value: textOf("5 years")}}}
	model := handResume(test, "home", commonWidth, tallHeight,
		document.Section{Title: textOf("Facts"), Blocks: []document.Block{facts}},
		document.Section{Title: textOf("Study"), Blocks: []document.Block{school}},
		document.Section{Title: textOf("Work"), Blocks: []document.Block{role, unfilled, unstarted}},
	)
	page := aboveTheFooter(screenOf(test, model))
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

// Only the résumé's first section, where it holds facts, joins the first tab, its facts as the brief: one that holds
// no fact, one after another section, and one after a first section of facts keep a tab of their own.
func TestOnlyAFirstSectionOfFactsJoinsTheFirstTab(test *testing.T) {
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
			name:   "facts after a paragraph, first",
			blocks: []document.Block{document.Section{Title: textOf("Facts"), Blocks: []document.Block{paragraph, facts}}},
			listed: false,
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
			if footer := footerOf(screenOf(test, model)); strings.Contains(footer, "1:facts*") == testCase.listed {
				test.Errorf("the footer = %q, want the facts on a tab of their own, not the first: %t", footer, testCase.listed)
			}
		})
	}
}

// The first section the first tab takes in names the tab, and what it holds beside its facts stands on the tab
// under the career's line; where none joins, the tab is named by the author.
func TestTheSectionTakenIntoTheFirstTabNamesTheTab(test *testing.T) {
	facts := document.Facts{List: []document.Fact{{Label: "experience", Value: textOf("5 years")}}}
	paragraph := document.Paragraph{Spans: textOf("the door")}
	cases := []struct {
		name       string
		blocks     []document.Block
		window     string
		doorInside bool
	}{
		{name: "a section of facts", blocks: []document.Block{document.Section{Title: textOf("Brief"), Blocks: []document.Block{paragraph, facts}}}, window: "1:brief*", doorInside: true},
		{name: "no section of facts", blocks: []document.Block{document.Section{Title: textOf("Work"), Blocks: []document.Block{paragraph}}}, window: "1:home*", doorInside: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			screen := screenOf(test, handResume(test, "home", commonWidth, tallHeight, testCase.blocks...))
			if footer := footerOf(screen); !strings.Contains(footer, testCase.window) {
				test.Errorf("the footer = %q, want the first tab named %q", footer, testCase.window)
			}
			if inside := strings.Contains(aboveTheFooter(screen), document.PlainText(paragraph.Spans)); inside != testCase.doorInside {
				test.Errorf("the first tab =\n%s\nholds the section's paragraph: %t, want %t", screen, inside, testCase.doorInside)
			}
		})
	}
}

// A project's page keeps a first section of facts alone as a tab of its own: only the résumé opens with a brief.
func TestAProjectsPageKeepsItsSectionOfFacts(test *testing.T) {
	facts := document.Facts{List: []document.Fact{{Label: "language", Value: textOf("Go")}}}
	project := document.Document{
		Head:   fixtures.HeadOf(document.Page{Kind: document.PageProject, Slug: "a-tool", Tab: ""}, textOf("a-tool")),
		Blocks: []document.Block{document.Section{Title: textOf("Facts"), Blocks: []document.Block{facts}}},
	}
	resume := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, textOf("home")), Blocks: nil}
	model := commanded(test, handModel(test, commonWidth, commonHeight, resume, project), "e a-tool")
	if footer := footerOf(screenOf(test, model)); !strings.Contains(footer, "2:facts") {
		test.Errorf("on the project's page the footer = %q, want its tab of the facts", footer)
	}
}

// The author's name reads as text where the name's font cannot draw it: on a page narrower than its banner, on one
// too short to show the tab under it, wrapped whole on a page narrower than the name, and where the font has not every
// letter of it.
func TestTheNameReadsAsTextWhereNoBannerFits(test *testing.T) {
	// a page narrower than the name written in one line; a screen whose page shows four rows; and one whose page has the
	// rows for the banner and the tab under it, not for the blank row over the banner as well
	const (
		narrowerThanTheName = 10
		shorterThanTheName  = 6
		noRowOverTheName    = 15
	)
	brief := document.Section{Title: textOf("Facts"), Blocks: []document.Block{document.Facts{List: []document.Fact{{Label: "experience", Value: textOf("5 years")}}}}}
	cases := []struct {
		name   string
		title  string
		width  int
		height int
		holds  string
	}{
		{name: "a page narrower than the banner", title: "Mira Example", width: phoneWidth, height: tallHeight, holds: "Mira Example"},
		{name: "a page shorter than the banner", title: "Mira Example", width: commonWidth, height: shorterThanTheName, holds: "Mira Example"},
		{name: "a page shorter than what stands over the name", title: "Mira Example", width: commonWidth, height: noRowOverTheName, holds: "Mira Example"},
		{name: "a page narrower than the name", title: "Mira Example", width: narrowerThanTheName, height: tallHeight, holds: "Example"},
		{name: "letters the font has not", title: "José Núñez", width: commonWidth, height: tallHeight, holds: "José Núñez"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := handResume(test, testCase.title, testCase.width, testCase.height, brief)
			if page := aboveTheFooter(screenOf(test, model)); !strings.Contains(page, testCase.holds) {
				test.Errorf("at %d×%d the page =\n%s\nwant the name's text %q", testCase.width, testCase.height, page, testCase.holds)
			}
		})
	}
}

// The footer names the sessions, as tmux names its own, the site's by its host and the blog's, then each tab by its
// number and its name, the one shown marked, as tmux marks the window shown; then the key that opens the window of
// keys, and the ruler.
func TestTheFooterNamesTheSiteAndEveryTab(test *testing.T) {
	want := []string{"[example.test]", "[blog]", "1:about", "2:experience*", "3:projects", "4:skills", "5:contact", "?", "keys", "All"}
	footer := strings.Fields(footerOf(screenOf(test, pressed(test, fixtureModel(test, wideWidth, commonHeight), "l"))))
	if !slices.Equal(footer, want) {
		test.Errorf("on the second tab the footer reads %q, want %q", footer, want)
	}
}

// Where the footer does not fit the screen, the key to the window of keys gives way, then the tabs farthest from the
// one shown, a mark standing for them, then the session, the ruler kept.
func TestTheFooterGivesWayAroundTheTabShown(test *testing.T) {
	// a terminal wide enough for the sessions and one tab in the footer; one wide enough for every tab of the
	// fixture's résumé, not for the key beside them; and one wide enough for the third tab and one tab beside it
	const (
		oneTabWidth    = 47
		everyTabWidth  = 92
		oneBesideWidth = 62
	)
	cases := []struct {
		name  string
		width int
		keys  []string
		holds []string
		lacks []string
	}{
		{name: "the last tab", width: oneTabWidth, keys: []string{"h"}, holds: []string{"[example.test] [blog] … 5:contact*"}, lacks: []string{"4:skills", "keys"}},
		{name: "the first tab at its end", width: oneTabWidth, keys: []string{"G"}, holds: []string{"[example.test] [blog] 1:about* … Bot"}, lacks: []string{"2:experience", "keys"}},
		{name: "a tab between", width: oneTabWidth, keys: []string{"l", "l"}, holds: []string{"[example.test] [blog] … 3:projects* …"}, lacks: []string{"2:experience", "4:skills"}},
		{name: "a screen too narrow for the sessions beside the tab shown", width: phoneWidth, keys: []string{"l"}, holds: []string{"2:experience*"}, lacks: []string{"[example.test]"}},
		{name: "the later of two as far gives way first", width: oneBesideWidth, keys: []string{"l", "l"}, holds: []string{"… 2:experience 3:projects* …"}, lacks: []string{"4:skills"}},
		{name: "a screen roomy enough for all but the key", width: everyTabWidth, keys: nil, holds: []string{"1:about* 2:experience 3:projects 4:skills 5:contact"}, lacks: []string{"keys"}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			footer := strings.Join(strings.Fields(footerOf(screenOf(test, pressed(test, fixtureModel(test, testCase.width, commonHeight), testCase.keys...)))), space)
			for _, want := range testCase.holds {
				if !strings.Contains(footer, want) {
					test.Errorf("at %d columns after %q the footer = %q, want it to hold %q", testCase.width, testCase.keys, footer, want)
				}
			}
			for _, unwanted := range testCase.lacks {
				if strings.Contains(footer, unwanted) {
					test.Errorf("at %d columns after %q the footer = %q, want it not to hold %q", testCase.width, testCase.keys, footer, unwanted)
				}
			}
		})
	}
}

// The first screen is the head of the résumé: its greeting, its title and its headline.
func TestTheFirstScreenShowsTheHead(test *testing.T) {
	screen := screenOf(test, fixtureModel(test, commonWidth, commonHeight))
	cases := []struct {
		name string
		want string
	}{
		{name: "the greeting", want: "Hello, my name is"},
		{name: "the title", want: "Software Engineer"},
		{name: "the headline", want: "Builds small tools that last."},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !strings.Contains(screen, testCase.want) {
				test.Errorf("the first screen =\n%s\nwant it to hold %q", screen, testCase.want)
			}
		})
	}
}

// A digit chooses the tab of its number at once, waiting for no key after it, as the page's vim keys do: a key after
// it does what it does alone, never counted; a digit of no tab, zero among them, chooses none.
func TestADigitChoosesItsTabAtOnce(test *testing.T) {
	cases := []struct {
		name  string
		keys  []string
		holds string
		lacks string
	}{
		{name: "the third tab", keys: []string{"3"}, holds: "other-tool", lacks: "Kept the queue"},
		{name: "the tab, then a line down", keys: []string{"2", "j"}, holds: "Kept the queue", lacks: "other-tool"},
		{name: "a digit of no tab", keys: []string{"2", "9"}, holds: "Kept the queue", lacks: "other-tool"},
		{name: "zero", keys: []string{"2", "0"}, holds: "Kept the queue", lacks: "other-tool"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := fixtureModel(test, commonWidth, shortHeight)
			for _, key := range testCase.keys {
				next, command := model.Update(keyPress(test, key))
				if command != nil {
					test.Fatalf("after %q a command came, want the digit taken at once", key)
				}
				model = as[Model](test, next)
			}
			if screen := screenOf(test, model); !strings.Contains(screen, testCase.holds) || strings.Contains(screen, testCase.lacks) {
				test.Errorf("after %q the screen =\n%s\nwant it to hold %q and not %q", testCase.keys, screen, testCase.holds, testCase.lacks)
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
		{name: "h goes round from the first tab to the last", keys: []string{"h"}, holds: "github.test/mira", lacks: "Builds small tools that last."},
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
		{name: "a key between two g's parts them", keys: []string{"G", "g", "k", "g"}, holds: "Kept the build green", lacks: "Northwind"},
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

// A row faded at the page's foot still carries its link to a terminal that follows links: on a narrow screen of
// thirteen rows the first tab's last row on the page is the fixture's address, faded, as more lies below it.
func TestALinkOnAFadedRowStillLeads(test *testing.T) {
	const fadedLinkHeight = 13
	if screen := fixtureModel(test, narrowWidth, fadedLinkHeight).View().Content; !strings.Contains(screen, "mailto:mira@example.test") {
		test.Errorf("the screen =\n%q\nwant the address row's link to mailto:mira@example.test", screen)
	}
}

// Each tab keeps where it was scrolled to while the reader looks at another.
func TestEachTabKeepsItsPlace(test *testing.T) {
	model := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(toExperience(), "G", "l", "h")...)
	if screen := screenOf(test, model); !strings.Contains(screen, "one long winter") {
		test.Errorf("back on the experience tab the screen =\n%s\nwant it still at the end", screen)
	}
}

// A role reads with its period, each month its year and its number, and with its title, its subtitle and its place,
// on a common terminal as on a phone.
func TestARoleReadsWithItsPeriodAndPlace(test *testing.T) {
	cases := []struct {
		name  string
		width int
		want  string
	}{
		{name: "the period of a role still held", width: commonWidth, want: "2023-04 – present"},
		{name: "a role still held", width: commonWidth, want: "Northwind — Senior Engineer · Brno"},
		{name: "the period of a role that ended", width: commonWidth, want: "2020-01 – 2023-03"},
		{name: "a role that ended", width: commonWidth, want: "Southwind — Engineer · Prague"},
		{name: "the period of a role on a phone", width: phoneWidth, want: "2023-04 – present"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			screen := screenOf(test, pressed(test, fixtureModel(test, testCase.width, tallHeight), toExperience()...))
			if !strings.Contains(screen, testCase.want) {
				test.Errorf("at %d columns the experience tab =\n%s\nwant it to hold %q", testCase.width, screen, testCase.want)
			}
		})
	}
}

// A table's row that gives way to words, its columns leaving the last too little room, fills its first line to the
// page's edge.
func TestARowGivenWayToWordsFillsItsFirstLine(test *testing.T) {
	// the words of the row's first line, its cells parted by the place mark, exactly as wide as a phone's page
	const firstLine = "abcdefghijklmn · opqrstuvwxy"
	row := document.Row{Cells: []document.Cell{
		{Kind: document.CellText, Spans: textOf("abcdefghijklmn")},
		{Kind: document.CellText, Spans: textOf("opqrstuvwxy z")},
	}}
	model := handResume(test, "home", phoneWidth, tallHeight, document.Table{Rows: []document.Row{row}})
	if page := aboveTheFooter(screenOf(test, model)); !strings.Contains(page, firstLine) {
		test.Errorf("at %d columns the page =\n%s\nwant the row's first line %q whole", phoneWidth, page, firstLine)
	}
}

// A word of letters two columns wide, where the line leaves one column for it, does not stop the tab from being laid
// out: each line takes a letter, though the page cuts it, a letter of one code point or of several.
func TestAWordOfWideLettersLeavesTheTabLaidOut(test *testing.T) {
	cases := []struct {
		name  string
		value string
	}{
		{name: "letters of one code point", value: "中文"},
		{name: "letters of two code points", value: "❤️❤️"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			facts := document.Facts{List: []document.Fact{{Label: "Email", Value: textOf(testCase.value)}}}
			if page := aboveTheFooter(screenOf(test, handResume(test, "home", oneColumnLeft, tallHeight, facts))); !strings.Contains(page, "Email") {
				test.Errorf("at %d columns with %q the page =\n%s\nwant the fact's label %q", oneColumnLeft, testCase.value, page, "Email")
			}
		})
	}
}

// A letter wider than the room its line leaves stands on that line, the next word on the line after: no line is
// left blank between them.
func TestALetterWiderThanItsRoomLeavesNoBlankLine(test *testing.T) {
	facts := document.Facts{List: []document.Fact{{Label: "Email", Value: textOf("中 ab")}}}
	lines := strings.Split(aboveTheFooter(screenOf(test, handResume(test, "home", oneColumnLeft, tallHeight, facts))), lineEnd)
	label := slices.IndexFunc(lines, func(shown string) bool { return strings.Contains(shown, "Email") })
	if label < 0 || label+1 >= len(lines) {
		test.Fatalf("the page =\n%s\nwant a line holding the label %q and one after it", strings.Join(lines, lineEnd), "Email")
	}
	if next := strings.TrimSpace(lines[label+1]); next != "a" {
		test.Errorf("the page =\n%s\nthe line after the label's holds %q, want %q", strings.Join(lines, lineEnd), next, "a")
	}
}

// A role's period stands apart from its title on the line they share: a word of its own.
func TestAPeriodStandsApartFromItsTitle(test *testing.T) {
	lines := strings.Split(screenOf(test, pressed(test, fixtureModel(test, commonWidth, tallHeight), toExperience()...)), lineEnd)
	shared := slices.IndexFunc(lines, func(shown string) bool { return strings.Contains(shown, "Northwind —") })
	if shared < 0 {
		test.Fatalf("the experience tab =\n%s\nwant a line holding the role %q", strings.Join(lines, lineEnd), "Northwind —")
	}
	if words := strings.Fields(lines[shared]); !slices.Contains(words, "present") || !slices.Contains(words, "Northwind") {
		test.Errorf("the role's line = %q, want its period's %q and its title's %q apart", lines[shared], "present", "Northwind")
	}
}

// A section stands under its title. On the résumé and on a project's page the title stands in the titles' font where
// the page has room for it and rows of the tab under it, as text where it has room for that alone, and on a page too
// short for either the section stands as it is, named in the footer alone; on another page it stands as it is.
func TestASectionStandsUnderItsTitle(test *testing.T) {
	// a screen whose page has the rows for the title as text and the tab under it, not for its banner; and one a row
	// taller, whose page has them for the banner. A banner is a look, so no case tells it from no title: they read
	// only whether the title stands as text
	const (
		shorterThanABanner   = 10
		tallEnoughForABanner = shorterThanABanner + 1
	)
	cases := []struct {
		name        string
		model       Model
		title       string
		titleAsText bool
	}{
		{name: "the résumé's experience", model: pressed(test, fixtureModel(test, phoneWidth, tallHeight), toExperience()...), title: "Experience", titleAsText: true},
		{name: "a page too short for the banner", model: pressed(test, fixtureModel(test, commonWidth, shorterThanABanner), toExperience()...), title: "Experience", titleAsText: true},
		{name: "a page tall enough for the banner holds no title as text", model: pressed(test, fixtureModel(test, commonWidth, tallEnoughForABanner), toExperience()...), title: "Experience", titleAsText: false},
		{name: "a short screen", model: pressed(test, fixtureModel(test, phoneWidth, shortHeight), toExperience()...), title: "Experience", titleAsText: false},
		{
			name: "a project's section on a page too short for the banner", title: "Why", titleAsText: true,
			model: pressed(test, commanded(test, fixtureModel(test, commonWidth, shorterThanABanner), "e tidy-tool"), "l"),
		},
		{
			name: "a project's section on a page tall enough for the banner", title: "Why", titleAsText: false,
			model: pressed(test, commanded(test, fixtureModel(test, commonWidth, tallEnoughForABanner), "e tidy-tool"), "l"),
		},
		{
			name: "a section of another page", title: "Growing", titleAsText: false,
			model: pressed(test, commanded(test, handHealth(test, phoneWidth, tallHeight, "Growing"), "checkhealth"), "l"),
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			page := aboveTheFooter(screenOf(test, testCase.model))
			if strings.Contains(page, testCase.title) != testCase.titleAsText {
				test.Errorf("the page =\n%s\nwant the title %q over it as text: %t", page, testCase.title, testCase.titleAsText)
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
		{name: "the smallest terminal still draws", width: smallestWidth, height: smallestHeight, holds: "2023-04", lacks: "Builds small tools"},
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
		Head:   fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, nil),
		Blocks: []document.Block{document.Facts{List: []document.Fact{{Label: "available", Value: nil}}}},
	}
	if screen := screenOf(test, handModel(test, commonWidth, commonHeight, composed)); !strings.Contains(screen, "available") {
		test.Errorf("the screen =\n%s\nwant the label %q", screen, "available")
	}
}

// A block of facts frames a command in a box, as the page's Command stands: its label in the top edge, and inside the
// prompt's mark before the command. Each fact round it stands once, and a fact whose value is code and more is no
// command, standing beside its label.
func TestEachFactRoundACommandStandsOnce(test *testing.T) {
	read := commandFactsRead(test)
	cases := []struct {
		name string
		text string
	}{
		{name: "a fact before the command", text: "status done"},
		{name: "the command after the prompt, between the box's sides", text: frameDown + space + look.PromptMark + space + "go install example.test/tool" + space + frameDown},
		{name: "the command's label, in the box's top edge", text: frameTopLeft + frameAcross + space + "setup" + space + frameAcross},
		{name: "a value of code and more", text: "steps make then run"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if holding := strings.Count(read, testCase.text); holding != 1 {
				test.Errorf("the page =\n%s\nholds %q %d times, want once", read, testCase.text, holding)
			}
		})
	}
}

// A block of facts keeps the document's order round a command it frames in a box, as no face picks the facts' order
// (ADR-0003): the fact before it above it, the one after below.
func TestAFactBlockKeepsItsOrderRoundACommand(test *testing.T) {
	read := commandFactsRead(test)
	places := []int{strings.Index(read, "status done"), strings.Index(read, "go install example.test/tool"), strings.Index(read, "steps make then run")}
	if slices.Min(places) < 0 || !slices.IsSorted(places) {
		test.Errorf("the page =\n%s\nholds its facts at %v, want each, in the document's order", read, places)
	}
}

// A command stands under its label, as any fact does, where the page is too narrow for its box: for the top edge round
// its label, or for the row of the prompt's mark and the command.
func TestACommandWiderThanItsBoxStandsUnderItsLabel(test *testing.T) {
	cases := []struct {
		name    string
		label   string
		command string
	}{
		{name: "a label too wide for the top edge", label: "set up the tool and run", command: "ls"},
		{name: "a command too wide for its row", label: "run", command: "go install example.tool"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			facts := document.Facts{List: []document.Fact{{Label: testCase.label, Value: []document.Span{document.Code{Text: testCase.command}}}}}
			read := readRows(aboveTheFooter(screenOf(test, handResume(test, "Mira", phoneWidth, commonHeight, facts))))
			if want := testCase.label + space + testCase.command; !strings.Contains(read, want) || strings.Contains(read, look.PromptMark) {
				test.Errorf("at %d columns the page =\n%s\nwant %q under its label, with no box", phoneWidth, read, want)
			}
		})
	}
}

// commandFactsRead is a page of a block of facts round a command, as readRows reads it.
func commandFactsRead(test *testing.T) string {
	facts := document.Facts{List: []document.Fact{
		{Label: "status", Value: textOf("done")},
		{Label: "setup", Value: []document.Span{document.Code{Text: "go install example.test/tool"}}},
		{Label: "steps", Value: []document.Span{document.Code{Text: "make"}, document.Text{Text: " then run"}}},
	}}
	return readRows(aboveTheFooter(screenOf(test, handResume(test, "Mira", commonWidth, commonHeight, facts))))
}

// readRows is a screen as a test reads its rows: each with its runs of spaces read as one, so a column a row lines
// up to reads as a space.
func readRows(screen string) string {
	rows := []string{}
	for row := range strings.SplitSeq(screen, lineEnd) {
		rows = append(rows, strings.Join(strings.Fields(row), space))
	}
	return strings.Join(rows, lineEnd)
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

// Every phrase of a page reaches the TUI on one tab or another: of the résumé and of a project's page, so a terminal
// reader reads every fact a browser reader reads, and of the health report and the list of skills :checkhealth and
// :Lazy show, each at any width. A word wider than a narrow terminal breaks across lines, so the phrases are read
// with every space and break taken out. The screen is read row by row, so the résumé is read where it is narrow:
// wider, its brief stands beside the name, and a row holds a piece of each. Its opening shows two phrases by its form
// (ADR-0003): the address heads the brief without its label, and the facts stand in the brief without their
// section's title. A tab's title the page draws in the titles' font, or not at all on the opening, reaches the reader
// by form too, as the footer's name of the tab. A project's opening shows its repository by its form too: its owner
// and name under the title, a link to the address, without its label. The blog's list and a post reach it as well,
// the blog's name in the titles' font, and a post's code of several lines in a frame, an edge between each two.
func TestEveryPhraseReachesTheTUI(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	cases := []struct {
		name  string
		shown document.Document
		model Model
		// byForm are the phrases the page shows by its form, not as text
		byForm []string
	}{
		{name: "the résumé", shown: compose.Resume(snapshot), model: sessionModel(snapshot, narrowWidth, tallHeight), byForm: []string{"mail", "Facts", "About", "Projects", "Skills", "Contact"}},
		{
			name: "a project's page", shown: fixtures.DocumentOf(test, compose.Documents(snapshot), document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""}),
			model:  commanded(test, sessionModel(snapshot, narrowWidth, tallHeight), "e tidy-tool"),
			byForm: []string{"repository", "https://github.test/mira/tidy", "Why"},
		},
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
			model: commanded(test, sessionModel(snapshot, commonWidth, tallHeight), "Lazy"), byForm: []string{"Skills"},
		},
		{
			name: "the list of skills, narrow", shown: compose.Skills(snapshot),
			model: commanded(test, sessionModel(snapshot, narrowWidth, tallHeight), "Lazy"), byForm: []string{"Skills"},
		},
		{
			name: "the blog's list", shown: fixtures.DocumentOf(test, compose.Lists(snapshot), document.Page{Kind: document.PageBlog, Slug: "", Tab: ""}),
			model: New(sessionDrawing(snapshot), "blog", commonWidth, tallHeight, fixtures.Day()), byForm: []string{"blog"},
		},
		{
			name: "a post", shown: fixtures.DocumentOf(test, compose.Posts(snapshot), document.Page{Kind: document.PagePost, Slug: "reading-slowly", Tab: ""}),
			model:  New(sessionDrawing(snapshot), "blog/1", commonWidth, tallHeight, fixtures.Day()),
			byForm: []string{"// slowly reads a page twice, aloud\nfunc slowly(page string) int {\n\tfmt.Println(\"slowly\", page)\n\treturn len(page) * 2\n}"},
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
	cases := []struct {
		name    string
		model   Model
		address string
	}{
		{name: "the mail address in the brief", model: fixtureModel(test, commonWidth, commonHeight), address: "mailto:mira@example.test"},
		{name: "a public profile on the contact tab", model: commanded(test, fixtureModel(test, commonWidth, commonHeight), "contact"), address: "https://github.test/mira"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if view := testCase.model.View().Content; !strings.Contains(view, ansi.SetHyperlink(testCase.address)) {
				test.Errorf("the screen links no %q", testCase.address)
			}
		})
	}
}

// Of the ways to reach the author the brief keeps the mail alone: the others stand once, on the contact tab.
func TestTheBriefKeepsTheMailAloneOfTheWays(test *testing.T) {
	screen := readRows(aboveTheFooter(screenOf(test, fixtureModel(test, commonWidth, tallHeight))))
	cases := []struct {
		name string
		way  string
	}{
		{name: "a public profile", way: "github.test/mira"},
		{name: "ssh", way: "ssh example.test"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if strings.Contains(screen, testCase.way) {
				test.Errorf("the first tab =\n%s\nholds %q, want it on the contact tab alone", screen, testCase.way)
			}
		})
	}
}

// The résumé's contact tab holds every way to reach the author, each by its label.
func TestTheContactTabHoldsEveryWayToReachTheAuthor(test *testing.T) {
	screen := readRows(aboveTheFooter(screenOf(test, commanded(test, fixtureModel(test, commonWidth, tallHeight), "contact"))))
	cases := []struct {
		name string
		want string
	}{
		{name: "the mail", want: "mail mira@example.test"},
		{name: "a public profile", want: "GitHub github.test/mira"},
		{name: "ssh, a command", want: "ssh example.test"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !strings.Contains(screen, testCase.want) {
				test.Errorf("the contact tab =\n%s\nwant %q on it", screen, testCase.want)
			}
		})
	}
}

func fixtureModel(test *testing.T, width, height int) Model {
	return sessionModel(fixtures.LoadedContent(test), width, height)
}

// sessionModel is a session on what a snapshot draws, as the SSH server opens one, opened on the résumé.
func sessionModel(snapshot *content.Snapshot, width, height int) Model {
	return New(sessionDrawing(snapshot), "", width, height, fixtures.Day())
}

// sessionDrawing is what the SSH server hands a session of a snapshot: its pages, the TUI's extras, the blog's lists
// and its posts, with the site's words and its site.
func sessionDrawing(snapshot *content.Snapshot) Drawing {
	documents := slices.Concat(compose.Documents(snapshot), compose.Extras(snapshot), compose.Lists(snapshot), compose.Posts(snapshot))
	return Drawing{Documents: documents, Frame: compose.FrameOf(snapshot), Site: fixtures.Site(snapshot)}
}

// handModel is a session on documents a test builds by hand, the résumé among them, with the fixture's words and site,
// opened on the résumé.
func handModel(test *testing.T, width, height int, documents ...document.Document) Model {
	snapshot := fixtures.LoadedContent(test)
	return New(Drawing{Documents: documents, Frame: compose.FrameOf(snapshot), Site: fixtures.Site(snapshot)}, "", width, height, fixtures.Day())
}

// handHealth is a session on a résumé of nothing and a health report a test builds by hand, of one section under a
// title.
func handHealth(test *testing.T, width, height int, title string) Model {
	resume := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, nil), Blocks: nil}
	section := document.Section{Title: textOf(title), Blocks: []document.Block{document.Paragraph{Spans: textOf("Learning Go")}}}
	report := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageHealth, Slug: "", Tab: ""}, textOf("health")), Blocks: []document.Block{section}}
	return handModel(test, width, height, resume, report)
}

// handResume is a session on a résumé a test builds by hand: a head titled by a name, then blocks.
func handResume(test *testing.T, name string, width, height int, blocks ...document.Block) Model {
	composed := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, textOf(name)), Blocks: blocks}
	return handModel(test, width, height, composed)
}

// knownMonth is a month of a year as a period of a document a test builds by hand names it.
func knownMonth(year int, month time.Month) document.KnownMonth {
	return document.KnownMonth{Start: time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)}
}

// textOf is a text as the spans of a document, for a document a test builds by hand.
func textOf(text string) []document.Span {
	return []document.Span{document.Text{Text: text}}
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
