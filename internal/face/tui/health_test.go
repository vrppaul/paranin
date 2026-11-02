package tui

import (
	"strings"
	"testing"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
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

// A session handed no health report knows no :checkhealth, and says so as Neovim says it of a command it does not
// know, the résumé still shown.
func TestCheckhealthWithNoReportIsNoCommand(test *testing.T) {
	snapshot := fixtureSnapshot(test)
	model := New([]document.Document{compose.Resume(snapshot)}, fixtureSite(snapshot), commonWidth, commonHeight)
	screen := screenOf(test, commanded(test, model, "checkhealth"))
	if want := "E492: Not an editor command: checkhealth"; !strings.Contains(screen, want) || !strings.Contains(screen, resumeHeadline) {
		test.Errorf("after :checkhealth with no report the screen =\n%s\nwant %q, and the résumé still shown", screen, want)
	}
}
