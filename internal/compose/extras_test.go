package compose

import (
	"os"
	"slices"
	"testing"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// The health report is a page of its own, under the site's word for it, and holds each group of checks the profile
// lists, in its order: each check at its level, with its advice where it has one.
func TestTheHealthReportHoldsEachCheck(test *testing.T) {
	report := Health(fixtures.LoadedContent(test))
	if report.Head.Page.Kind != document.PageHealth {
		test.Errorf("the report's page = %+v, want the health report's", report.Head.Page)
	}
	phrases := []string{"health", "paranin", "all fine", "one thing", "look at it", "tooling", "no tests yet"}
	if got := document.Phrases(report); !slices.Equal(got, phrases) {
		test.Errorf("the report's phrases = %q, want %q", got, phrases)
	}
	levels := []document.CheckLevel{document.CheckOK, document.CheckWarning, document.CheckError}
	if got := checkLevelsOf(test, report); !slices.Equal(got, levels) {
		test.Errorf("the report's levels = %v, want %v", got, levels)
	}
}

// The list of skills is a page of its own, under the site's word for skills, and holds every skill of every group:
// those the author uses first, then those they are still learning, each in the profile's order, how long each was
// used in the site's words where the profile writes a word for it, and in months where it writes a span.
func TestTheListOfSkillsHoldsEverySkill(test *testing.T) {
	list := Skills(fixtures.LoadedContent(test))
	if list.Head.Page.Kind != document.PageSkills {
		test.Errorf("the list's page = %+v, want the list of skills'", list.Head.Page)
	}
	phrases := []string{
		"Skills", "Go", "3y", "tidy-tool, a hobby", "Shell scripts for the build machines", "2y", "Northwind, Southwind",
		"Rust", "still learning", "a hobby", "Docker", "on the side", "Northwind",
	}
	if got := document.Phrases(list); !slices.Equal(got, phrases) {
		test.Errorf("the list's phrases = %q, want %q", got, phrases)
	}
	want := []listedTime{{Learning: false, Months: 36}, {Learning: false, Months: 24}, {Learning: true, Months: 0}, {Learning: true, Months: 0}}
	if got := listedTimesOf(test, list); !slices.Equal(got, want) {
		test.Errorf("whether each skill of the list is still being learnt, and its months, = %v, want %v", got, want)
	}
}

// Where the profile leaves how long a skill was used a gap, the list holds it as a gap still to fill.
func TestAGapForHowLongASkillWasUsedReachesTheListAsAGap(test *testing.T) {
	const gap = "[2y?]"
	folder := fixtures.Copy(test, fixtures.ContentFolder(test))
	fixtures.ReplaceIn(test, folder, "profile.yaml", "duration: 3y", `duration: "`+gap+`"`)
	snapshot, err := content.Load(os.DirFS(folder))
	if err != nil {
		test.Fatalf("load the fixture with a gap for Go's duration: %v", err)
	}
	want := []document.Span{document.Gap{Text: gap}}
	if got := listedSkillNamed(test, Skills(snapshot), "Go").Duration; !slices.Equal(got, want) {
		test.Errorf("Go's duration = %#v, want %#v", got, want)
	}
}

// listedSkillNamed is the skill of a name in a list of skills.
func listedSkillNamed(test *testing.T, list document.Document, name string) document.Skill {
	for _, block := range list.Blocks {
		for _, skill := range as[document.Skills](test, block).List {
			if document.PlainText(skill.Name) == name {
				return skill
			}
		}
	}
	test.Fatalf("the list holds no skill %q", name)
	return document.Skill{}
}

// listedTime is what a skill of the list says of the author's time with it: whether they are still learning it, and
// how many months they used it.
type listedTime struct {
	Learning bool
	Months   int
}

// listedTimesOf are what each skill of a list says of the author's time with it, in the list's order.
func listedTimesOf(test *testing.T, list document.Document) []listedTime {
	times := []listedTime{}
	for _, block := range list.Blocks {
		for _, skill := range as[document.Skills](test, block).List {
			times = append(times, listedTime{Learning: skill.Learning, Months: skill.Months})
		}
	}
	return times
}

// checkLevelsOf are the levels of a report's checks, group after group.
func checkLevelsOf(test *testing.T, report document.Document) []document.CheckLevel {
	levels := []document.CheckLevel{}
	for _, block := range report.Blocks {
		for _, check := range as[document.Checks](test, block).List {
			levels = append(levels, check.Level)
		}
	}
	return levels
}
