package main

import (
	"slices"
	"strings"
	"testing"
)

// placedFault is which fault a message has, and on which line.
type placedFault struct {
	kind kind
	line int
}

func TestAMessageIsHeldToAShortSubjectWithNoHistoryAndAShortWhy(test *testing.T) {
	trailers := "\n\nCo-Authored-By: Someone <someone@example.com>\nClaude-Session: https://example.com/session"
	testCases := []struct {
		name    string
		message string
		want    []placedFault
	}{
		{name: "a short subject and a why", message: "feat(page): the site has its icon\n\nWhy, in a line." + trailers, want: []placedFault{}},
		{name: "a subject of the most letters", message: "fix: " + strings.Repeat("a", 67), want: []placedFault{}},
		{name: "a subject a letter too long", message: "fix: " + strings.Repeat("a", 68), want: []placedFault{{kind: kindLongSubject, line: 1}}},
		{name: "letters, not bytes", message: "fix: " + strings.Repeat("é", 67), want: []placedFault{}},
		{
			name:    "a body of six lines",
			message: "fix: one\n\none\ntwo\nthree\n\nfour\nfive\nsix" + trailers,
			want:    []placedFault{{kind: kindLongBody, line: 3}},
		},
		{name: "trailers are no part of the body", message: "fix: one\n\none\ntwo\nthree\nfour\nfive" + trailers, want: []placedFault{}},
		{name: "a subject that tells what was", message: "refactor: the reader no longer reads twice", want: []placedFault{{kind: kindHistory, line: 1}}},
		{name: "a body says what changed", message: "refactor: one read of the profile\n\nThe reader no longer reads twice.", want: []placedFault{}},
		{
			name:    "git's own lines and the diff under the scissors",
			message: "fix: one\n" + strings.Repeat("# git's line\n", 6) + scissors + "\n" + strings.Repeat("+ a line of the diff\n", 6),
			want:    []placedFault{},
		},
	}
	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := placedFaultsOf(faultsOf(testCase.message)); !slices.Equal(got, testCase.want) {
				test.Errorf("the faults of %q are %v, want %v", testCase.message, got, testCase.want)
			}
		})
	}
}

// placedFaultsOf are which faults came and where, their words left out.
func placedFaultsOf(found []fault) []placedFault {
	placed := []placedFault{}
	for _, each := range found {
		placed = append(placed, placedFault{kind: each.kind, line: each.line})
	}
	return placed
}
