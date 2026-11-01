package routes

import (
	"testing"

	"github.com/vrppaul/paranin/internal/document"
)

// The site's addresses are a contract with whoever links to it: each page has its address in each format, the
// address a reader's request chooses the format at among them, and an address outside the site is left as it is.
func TestEveryTargetHasItsAddress(test *testing.T) {
	site := New("https://example.test")
	resume := document.Page{Kind: document.PageResume, Slug: ""}
	project := document.Page{Kind: document.PageProject, Slug: "tidy-tool"}
	cases := []struct {
		name   string
		target document.Target
		format Format
		want   string
	}{
		{name: "the résumé in markdown", target: resume, format: FormatMarkdown, want: "https://example.test/index.md"},
		{name: "the résumé as a manual", target: resume, format: FormatManual, want: "https://example.test/index.txt"},
		{name: "the résumé as the reader asks", target: resume, format: FormatChosen, want: "https://example.test/"},
		{name: "a project's page in markdown", target: project, format: FormatMarkdown, want: "https://example.test/p/tidy-tool.md"},
		{name: "a project's page as a manual", target: project, format: FormatManual, want: "https://example.test/p/tidy-tool.txt"},
		{name: "a project's page as the reader asks", target: project, format: FormatChosen, want: "https://example.test/p/tidy-tool"},
		{name: "no page", target: document.Page{Kind: document.PageNone, Slug: ""}, format: FormatMarkdown, want: ""},
		{name: "an address outside the site", target: document.Address{URL: "mailto:mira@example.test"}, format: FormatMarkdown, want: "mailto:mira@example.test"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := site.Address(testCase.target, testCase.format); got != testCase.want {
				test.Errorf("Address(%+v) = %q, want %q", testCase.target, got, testCase.want)
			}
		})
	}
}

// The site's name is the host of its base address alone, whatever path the address goes on with.
func TestTheSiteIsNamedByItsHost(test *testing.T) {
	cases := []struct {
		name string
		base string
		want string
	}{
		{name: "a bare address", base: "https://example.test", want: "example.test"},
		{name: "an address with a path", base: "https://example.test/cv", want: "example.test"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := New(testCase.base).Host(); got != testCase.want {
				test.Errorf("New(%q).Host() = %q, want %q", testCase.base, got, testCase.want)
			}
		})
	}
}
