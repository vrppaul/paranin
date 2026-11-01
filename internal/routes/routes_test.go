package routes

import (
	"testing"

	"github.com/vrppaul/paranin/internal/document"
)

// The site's addresses are a contract with whoever links to it: each page has its address in each format, and an
// address outside the site is left as it is.
func TestEveryTargetHasItsAddress(test *testing.T) {
	site := New("https://example.test")
	cases := []struct {
		name   string
		target document.Target
		want   string
	}{
		{name: "the résumé", target: document.Page{Kind: document.PageResume, Slug: ""}, want: "https://example.test/index.md"},
		{name: "a project's page", target: document.Page{Kind: document.PageProject, Slug: "tidy-tool"}, want: "https://example.test/p/tidy-tool.md"},
		{name: "no page", target: document.Page{Kind: document.PageNone, Slug: ""}, want: ""},
		{name: "an address outside the site", target: document.Address{URL: "mailto:mira@example.test"}, want: "mailto:mira@example.test"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := site.Address(testCase.target, FormatMarkdown); got != testCase.want {
				test.Errorf("Address(%+v) = %q, want %q", testCase.target, got, testCase.want)
			}
		})
	}
}
