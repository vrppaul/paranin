package tui

import (
	"strings"
	"testing"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// :checkhealth shows the health report in place of the résumé, as Neovim writes its own: each check after the word
// of its level, and its advice after Neovim's word for advice. TestEveryPhraseReachesTheTUI holds its groups' names.
func TestCheckhealthShowsTheHealthReport(test *testing.T) {
	screen := screenOf(test, commanded(test, fixtureModel(test, commonWidth, commonHeight), "checkhealth"))
	cases := []struct {
		name string
		want string
	}{
		{name: "a check that passed", want: "- OK all fine"},
		{name: "a check with something to look at", want: "- WARNING one thing"},
		{name: "its advice", want: "- ADVICE: look at it"},
		{name: "a check that failed, in a group of its own", want: "- ERROR no tests yet"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !strings.Contains(screen, testCase.want) || strings.Contains(screen, resumeHeadline) {
				test.Errorf("after :checkhealth the screen =\n%s\nwant %q, and the résumé gone", screen, testCase.want)
			}
		})
	}
}

// Where the count does not fit beside the ruler it drops whole, the ruler kept; and at every width it shows in
// every mode or in none.
func TestTheCountOfChecksShowsWholeOrNotAtAll(test *testing.T) {
	phone := belowThePage(screenOf(test, commanded(test, fixtureModel(test, phoneWidth, commonHeight), "checkhealth")))
	if strings.Contains(phone, "OK") || !strings.Contains(phone, "All") {
		test.Errorf("at a phone's width the status and command lines = %q, want no part of the count, and the ruler", phone)
	}
	for width := phoneWidth; width <= commonWidth; width++ {
		normal := commanded(test, fixtureModel(test, width, commonHeight), "checkhealth")
		inNormal := strings.Contains(belowThePage(screenOf(test, normal)), "ERROR")
		if afterColon := strings.Contains(belowThePage(screenOf(test, pressed(test, normal, ":"))), "ERROR"); afterColon != inNormal {
			test.Errorf("at %d columns the count shows in normal mode: %t, after a colon: %t; want the same", width, inNormal, afterColon)
		}
	}
}

// The status line counts a tab's checks by level after its title, as the first version's did, each level that has
// a check, its sections' included; a page of no checks counts none.
func TestTheStatusLineCountsTheChecks(test *testing.T) {
	cases := []struct {
		name    string
		model   Model
		count   string
		counted bool
	}{
		{
			name:  "the health report",
			model: commanded(test, fixtureModel(test, commonWidth, commonHeight), "checkhealth"),
			count: "health · 1 OK · 1 WARNING · 1 ERROR", counted: true,
		},
		{name: "the résumé", model: fixtureModel(test, commonWidth, commonHeight), count: "OK", counted: false},
		{
			name:  "a tab whose section holds a group of its own",
			model: pressed(test, commanded(test, nestedReportModel(test), "checkhealth"), "l"),
			count: "outer · 2 WARNING", counted: true,
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			below := belowThePage(screenOf(test, testCase.model))
			if strings.Contains(below, testCase.count) != testCase.counted {
				test.Errorf("on %s the status and command lines = %q, want %q counted: %t", testCase.name, below, testCase.count,
					testCase.counted)
			}
		})
	}
}

// nestedReportModel is a session on the fixture's résumé and a health report of a section, outer, that holds a
// warning and a section of its own holding another.
func nestedReportModel(test *testing.T) Model {
	warning := document.Checks{Title: textOf("group"), List: []document.Check{{Level: document.CheckWarning, Spans: textOf("look"), Advice: nil}}}
	inner := document.Section{Title: textOf("inner"), Blocks: []document.Block{warning}}
	report := document.Document{
		Head:   fixtures.HeadOf(document.Page{Kind: document.PageHealth, Slug: ""}, textOf("report")),
		Blocks: []document.Block{document.Section{Title: textOf("outer"), Blocks: []document.Block{warning, inner}}},
	}
	snapshot := fixtures.LoadedContent(test)
	return New([]document.Document{compose.Resume(snapshot), report}, fixtureSite(snapshot), commonWidth, commonHeight)
}

// A session handed no health report knows no :checkhealth, and one handed no list of skills no :Lazy; each says so
// as Neovim says it of a command it does not know, the résumé still shown.
func TestAHiddenCommandWithNoPageIsNoCommand(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	model := New([]document.Document{compose.Resume(snapshot)}, fixtureSite(snapshot), commonWidth, commonHeight)
	cases := []struct {
		name    string
		command string
	}{
		{name: "the health report", command: "checkhealth"},
		{name: "the list of skills", command: "Lazy"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			screen := screenOf(test, commanded(test, model, testCase.command))
			want := "E492: Not an editor command: " + testCase.command
			if !strings.Contains(screen, want) || !strings.Contains(screen, resumeHeadline) {
				test.Errorf("after :%s with no page of it the screen =\n%s\nwant %q, and the résumé still shown", testCase.command,
					screen, want)
			}
		})
	}
}
