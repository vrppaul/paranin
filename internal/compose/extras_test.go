package compose

import (
	"slices"
	"testing"

	"github.com/vrppaul/paranin/internal/document"
)

// The health report is a page of its own, under the site's word for it, and holds each group of checks the profile
// lists, in its order: each check at its level, with its advice where it has one.
func TestTheHealthReportHoldsEachCheck(test *testing.T) {
	report := Health(fixtureSnapshot(test))
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
