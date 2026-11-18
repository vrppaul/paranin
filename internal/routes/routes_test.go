package routes

import (
	"testing"

	"github.com/vrppaul/paranin/internal/document"
)

// The site's addresses are a contract with whoever links to it: each page has its address in each format, the
// address a reader's request chooses the format at among them, the index at the address agents look for it,
// llms.txt, and in markdown alone; and an address outside the site is left as it is.
func TestEveryTargetHasItsAddress(test *testing.T) {
	site := New(writtenAddress{written: "https://example.test", host: "example.test"})
	resume := document.Page{Kind: document.PageResume, Slug: ""}
	project := document.Page{Kind: document.PageProject, Slug: "tidy-tool"}
	index := document.Page{Kind: document.PageIndex, Slug: ""}
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
		{name: "the index in markdown", target: index, format: FormatMarkdown, want: "https://example.test/llms.txt"},
		{name: "the index has no manual", target: index, format: FormatManual, want: ""},
		{name: "the index has no address to choose at", target: index, format: FormatChosen, want: ""},
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

// A link inside one of the site's own pages leads to a page of the site by its path from the root, which a browser
// reads against the address the page came from, so a page served from this machine leads to this machine; an
// address outside the site is left as it is, and a page with no address leads nowhere.
func TestALinkInsideThePageLeadsByItsPath(test *testing.T) {
	cases := []struct {
		name   string
		target document.Target
		format Format
		want   string
	}{
		{name: "the résumé as the reader asks", target: document.Page{Kind: document.PageResume, Slug: ""}, format: FormatChosen, want: "/"},
		{name: "a project's page in markdown", target: document.Page{Kind: document.PageProject, Slug: "tidy-tool"}, format: FormatMarkdown, want: "/p/tidy-tool.md"},
		{name: "a hidden page", target: document.Page{Kind: document.PageHealth, Slug: ""}, format: FormatChosen, want: ""},
		{name: "an address outside the site", target: document.Address{URL: "https://github.test/mira"}, format: FormatChosen, want: "https://github.test/mira"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := Reference(testCase.target, testCase.format); got != testCase.want {
				test.Errorf("Reference(%+v) = %q, want %q", testCase.target, got, testCase.want)
			}
		})
	}
}

// A link to a place on a page leads to it by its id: on a page of the site by the page's path then the place, on the
// page the link stands on by the place alone.
func TestALinkToAPlaceLeadsByItsID(test *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{name: "a section of the résumé", got: PlaceReference(document.Page{Kind: document.PageResume, Slug: ""}, FormatChosen, "about"), want: "/#about"},
		{name: "a place on a project's page", got: PlaceReference(document.Page{Kind: document.PageProject, Slug: "tidy-tool"}, FormatChosen, "usage"), want: "/p/tidy-tool#usage"},
		{name: "a place on the page itself", got: Place("main"), want: "#main"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if testCase.got != testCase.want {
				test.Errorf("the reference = %q, want %q", testCase.got, testCase.want)
			}
		})
	}
}

// writtenAddress is a base address as a test writes it, its host beside it, as content reads one from a profile.
type writtenAddress struct {
	written string
	host    string
}

func (address writtenAddress) Written() string {
	return address.written
}

func (address writtenAddress) Host() string {
	return address.host
}
