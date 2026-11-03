package tui

import (
	"strings"
	"testing"
)

// :Lazy shows the list of skills in place of the résumé, as lazy.nvim lists its plugins: each skill after a mark of
// whether the author uses it or is still learning it.
func TestLazyListsEverySkill(test *testing.T) {
	screen := screenOf(test, commanded(test, fixtureModel(test, commonWidth, commonHeight), "Lazy"))
	if strings.Contains(screen, resumeHeadline) {
		test.Errorf("after :Lazy the screen =\n%s\nwant the résumé gone", screen)
	}
	cases := []struct {
		name string
		want string
	}{
		{name: "a skill in use", want: "● Go"},
		{name: "a skill still being learnt", want: "○ Rust"},
		{name: "another skill still being learnt", want: "○ Docker"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !strings.Contains(screen, testCase.want) {
				test.Errorf("after :Lazy the screen =\n%s\nwant %q", screen, testCase.want)
			}
		})
	}
}
