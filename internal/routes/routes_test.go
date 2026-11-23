package routes

import (
	"testing"

	"github.com/vrppaul/paranin/internal/document"
)

// The site's addresses are a contract with whoever links to it: each page and each tab of it shown alone has its
// address in each format, the address a reader's request chooses the format at among them, the index at the address
// agents look for it, llms.txt, and in markdown alone; and an address outside the site is left as it is.
func TestEveryTargetHasItsAddress(test *testing.T) {
	site := New(writtenAddress{written: "https://example.test", host: "example.test"})
	resume := document.Page{Kind: document.PageResume, Slug: "", Tab: ""}
	resumeTab := document.Page{Kind: document.PageResume, Slug: "", Tab: "experience"}
	project := document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""}
	projectTab := document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: "why"}
	index := document.Page{Kind: document.PageIndex, Slug: "", Tab: ""}
	post := document.Page{Kind: document.PagePost, Slug: "reading-slowly", Tab: ""}
	list := document.Page{Kind: document.PageBlog, Slug: "", Tab: ""}
	listBySeries := document.Page{Kind: document.PageBlogSeries, Slug: "", Tab: ""}
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
		{name: "a tab of the résumé in markdown", target: resumeTab, format: FormatMarkdown, want: "https://example.test/experience.md"},
		{name: "a tab of the résumé as a manual", target: resumeTab, format: FormatManual, want: "https://example.test/experience.txt"},
		{name: "a tab of the résumé as the reader asks", target: resumeTab, format: FormatChosen, want: "https://example.test/experience"},
		{name: "a tab of a project's page in markdown", target: projectTab, format: FormatMarkdown, want: "https://example.test/p/tidy-tool/why.md"},
		{name: "a tab of a project's page as a manual", target: projectTab, format: FormatManual, want: "https://example.test/p/tidy-tool/why.txt"},
		{name: "a tab of a project's page as the reader asks", target: projectTab, format: FormatChosen, want: "https://example.test/p/tidy-tool/why"},
		{name: "the index in markdown", target: index, format: FormatMarkdown, want: "https://example.test/llms.txt"},
		{name: "the index has no manual", target: index, format: FormatManual, want: ""},
		{name: "the index has no address to choose at", target: index, format: FormatChosen, want: ""},
		{name: "a post in markdown", target: post, format: FormatMarkdown, want: "https://example.test/blog/reading-slowly.md"},
		{name: "a post as a manual", target: post, format: FormatManual, want: "https://example.test/blog/reading-slowly.txt"},
		{name: "a post as the reader asks", target: post, format: FormatChosen, want: "https://example.test/blog/reading-slowly"},
		{name: "the blog's list in markdown", target: list, format: FormatMarkdown, want: "https://example.test/blog.md"},
		{name: "the blog's list as a manual", target: list, format: FormatManual, want: "https://example.test/blog.txt"},
		{name: "the blog's list as the reader asks", target: list, format: FormatChosen, want: "https://example.test/blog"},
		{name: "the list by series in markdown", target: listBySeries, format: FormatMarkdown, want: "https://example.test/blog/series.md"},
		{name: "the list by series as a manual", target: listBySeries, format: FormatManual, want: "https://example.test/blog/series.txt"},
		{name: "the list by series as the reader asks", target: listBySeries, format: FormatChosen, want: "https://example.test/blog/series"},
		{name: "no page", target: document.Page{Kind: document.PageNone, Slug: "", Tab: ""}, format: FormatMarkdown, want: ""},
		{name: "an address outside the site", target: document.Address{URL: "https://github.test/mira"}, format: FormatMarkdown, want: "https://github.test/mira"},
		{name: "a mailbox, under the scheme that writes to it", target: document.Mail{Address: "mira@example.test"}, format: FormatMarkdown, want: "mailto:mira@example.test"},
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
		{name: "the résumé as the reader asks", target: document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, format: FormatChosen, want: "/"},
		{name: "a project's page in markdown", target: document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""}, format: FormatMarkdown, want: "/p/tidy-tool.md"},
		{name: "a hidden page", target: document.Page{Kind: document.PageHealth, Slug: "", Tab: ""}, format: FormatChosen, want: ""},
		{name: "an address outside the site", target: document.Address{URL: "https://github.test/mira"}, format: FormatChosen, want: "https://github.test/mira"},
		{name: "a mailbox, under the scheme that writes to it", target: document.Mail{Address: "mira@example.test"}, format: FormatChosen, want: "mailto:mira@example.test"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := Reference(testCase.target, testCase.format); got != testCase.want {
				test.Errorf("Reference(%+v) = %q, want %q", testCase.target, got, testCase.want)
			}
		})
	}
}

// A page of the site is read in a terminal with curl and its whole address, which answers a terminal with the manual.
func TestATerminalReadsAPageByItsWholeAddress(test *testing.T) {
	site := New(writtenAddress{written: "https://example.test", host: "example.test"})
	cases := []struct {
		name string
		page document.Page
		want string
	}{
		{name: "the résumé", page: document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, want: "curl https://example.test/"},
		{name: "a project's page", page: document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""}, want: "curl https://example.test/p/tidy-tool"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := site.ReadCommand(testCase.page); got != testCase.want {
				test.Errorf("ReadCommand(%+v) = %q, want %q", testCase.page, got, testCase.want)
			}
		})
	}
}

// In a terminal a page is reached by its path with no slash before it, as an SSH user name and the TUI's :e write it
// (D26): the blog's list, a post by its slug and by its number.
func TestATerminalNamesAPageByItsPath(test *testing.T) {
	cases := []struct {
		name string
		path string
		want string
	}{
		{name: "the blog's list", path: Path(document.Page{Kind: document.PageBlog, Slug: "", Tab: ""}, FormatChosen), want: "blog"},
		{name: "a post by its slug", path: Path(document.Page{Kind: document.PagePost, Slug: "reading-slowly", Tab: ""}, FormatChosen), want: "blog/reading-slowly"},
		{name: "a post by its number", path: NumberPath(12), want: "blog/12"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := TerminalName(testCase.path); got != testCase.want {
				test.Errorf("TerminalName(%q) = %q, want %q", testCase.path, got, testCase.want)
			}
		})
	}
}

// A name a terminal writes is the blog's where it is the list's or one in the blog's folder, whether or not a page goes
// by it, and no other's.
func TestABlogsNameIsItsListsOrInItsFolder(test *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{name: "blog", want: true},
		{name: "blog/99", want: true},
		{name: "blogger", want: false},
		{name: "reader", want: false},
		{name: "p/tidy-tool", want: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := IsBlogName(testCase.name); got != testCase.want {
				test.Errorf("IsBlogName(%q) = %t, want %t", testCase.name, got, testCase.want)
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
