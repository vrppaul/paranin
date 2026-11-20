package history

import (
	"slices"
	"testing"
)

func TestThePhrasesThatTellWhatCodeWasAreFound(test *testing.T) {
	testCases := []struct {
		name string
		text string
		want []string
	}{
		{name: "each phrase", text: "no longer, anymore, formerly, the old way, used to be, renamed from, was changed to, now uses", want: []string{
			"no longer", "anymore", "formerly", "the old way", "used to be", "renamed from", "was changed to", "now uses",
		}},
		{name: "a renaming said four ways", text: "it was renamed, they were renamed, it has been renamed, they have been renamed", want: []string{
			"was renamed", "were renamed", "has been renamed", "have been renamed",
		}},
		{name: "any case", text: "No Longer read", want: []string{"No Longer"}},
		{name: "a measure", text: "a name no longer than a line", want: []string{}},
		{name: "a part of a word", text: "nowadays it reads the formerlyish anymoreness of theformerly", want: []string{}},
		{name: "what stands", text: "the page draws the facts the profile holds now", want: []string{}},
	}
	reader := NewReader()
	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			got := []string{}
			for _, found := range reader.Within(testCase.text) {
				got = append(got, found.Phrase)
			}
			if !slices.Equal(got, testCase.want) {
				test.Errorf("the phrases of %q are %q, want %q", testCase.text, got, testCase.want)
			}
		})
	}
}
