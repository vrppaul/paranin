package compose

import (
	"slices"
	"testing"

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
// those the author uses first, then those they are still learning, each in the profile's order.
func TestTheListOfSkillsHoldsEverySkill(test *testing.T) {
	list := Skills(fixtures.LoadedContent(test))
	if list.Head.Page.Kind != document.PageSkills {
		test.Errorf("the list's page = %+v, want the list of skills'", list.Head.Page)
	}
	phrases := []string{
		"Skills", "Go", "3y", "tidy-tool, a hobby", "Shell scripts for the build machines", "2y", "Northwind, Southwind",
		"Rust", "weeks", "a hobby", "Docker", "side", "Northwind",
	}
	if got := document.Phrases(list); !slices.Equal(got, phrases) {
		test.Errorf("the list's phrases = %q, want %q", got, phrases)
	}
	learning := []bool{false, false, true, true}
	if got := learningOf(test, list); !slices.Equal(got, learning) {
		test.Errorf("the list's skills still being learnt = %v, want %v", got, learning)
	}
}

// learningOf is whether each skill of a list is one the author is still learning, in the list's order.
func learningOf(test *testing.T, list document.Document) []bool {
	learning := []bool{}
	for _, block := range list.Blocks {
		for _, skill := range as[document.Skills](test, block).List {
			learning = append(learning, skill.Learning)
		}
	}
	return learning
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
