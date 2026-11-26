package tui

import (
	"strings"
	"testing"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// :checkhealth shows the health report in place of the résumé, as Neovim writes its own: all its checks counted by
// level by its title, whole wherever the title leaves them, and on a screen too short for the title alone; each
// group's name and mark, and on its row the group's checks counted, or on the next where the row has not the room;
// each check after the mark and the word of its level, and its advice after Neovim's word for advice.
// TestEveryPhraseReachesTheTUI holds its groups' names.
func TestCheckhealthShowsTheHealthReport(test *testing.T) {
	// a page too narrow for a group's name and its count on one row
	const narrowerThanAGroupsRow = 24
	report := commanded(test, fixtureModel(test, commonWidth, commonHeight), "checkhealth")
	narrow := commanded(test, fixtureModel(test, narrowerThanAGroupsRow, tallHeight), "checkhealth")
	cases := []struct {
		name  string
		model Model
		want  string
	}{
		{name: "the report's count", model: report, want: "1 OK · 1 WARNING · 1 ERROR"},
		{
			name: "the report's count whole where the title leaves it no room beside", want: "1 OK · 1 WARNING · 1 ERROR",
			model: commanded(test, fixtureModel(test, narrowWidth, tallHeight), "checkhealth"),
		},
		{
			name: "the report's count on a screen too short for the title", want: "1 OK · 1 WARNING · 1 ERROR",
			model: commanded(test, fixtureModel(test, commonWidth, shortHeight), "checkhealth"),
		},
		{name: "a group and its count", model: report, want: "paranin ~ 1 OK · 1 WARNING"},
		{name: "a check that passed", model: report, want: "✓ OK all fine"},
		{name: "a check with something to look at", model: report, want: "⚠ WARNING one thing"},
		{name: "its advice", model: report, want: "ADVICE: look at it"},
		{name: "a group of a failed check alone", model: report, want: "tooling ~ 1 ERROR"},
		{name: "a check that failed", model: report, want: "✗ ERROR no tests yet"},
		{name: "a group's count under its name on a narrow page", model: narrow, want: "paranin ~\n1 OK · 1 WARNING"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			read := readRows(aboveTheFooter(screenOf(test, testCase.model)))
			if !strings.Contains(read, testCase.want) || strings.Contains(read, resumeHeadline) {
				test.Errorf("after :checkhealth the page =\n%s\nwant %q on it, and the résumé gone", read, testCase.want)
			}
		})
	}
}

// A session handed no health report knows no :checkhealth, and one handed no list of skills no :Lazy; each says so
// as Neovim says it of a command it does not know, the résumé still shown.
func TestAHiddenCommandWithNoPageIsNoCommand(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	model := handModel(test, commonWidth, commonHeight, compose.Resume(snapshot, nil, nil))
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
