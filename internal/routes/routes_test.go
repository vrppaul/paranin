package routes

import (
	"net/url"
	"testing"

	"github.com/vrppaul/paranin/internal/document"
)

// The site's addresses are a contract with whoever links to it: each page and each tab of it shown alone has its
// address in each format, the address a reader's request chooses the format at among them, the index at the address
// agents look for it, llms.txt, and in markdown alone; and an address outside the site is left as it is.
func TestEveryTargetHasItsAddress(test *testing.T) {
	site := New(writtenAddress{written: "https://example.test", host: "example.test"}, testedAuthority, testedTranslations())
	resume := document.Page{Kind: document.PageResume, Slug: "", Tab: ""}
	resumeTab := document.Page{Kind: document.PageResume, Slug: "", Tab: "experience"}
	project := document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""}
	projectTab := document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: "why"}
	index := document.Page{Kind: document.PageIndex, Slug: "", Tab: ""}
	post := document.Page{Kind: document.PagePost, Slug: "reading-slowly", Tab: ""}
	older := post.OlderVersion("2026-11-10")
	translated := post.InLanguage("ru")
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
		{name: "an older version in markdown", target: older, format: FormatMarkdown, want: "https://example.test/blog/reading-slowly/2026-11-10.md"},
		{name: "an older version as a manual", target: older, format: FormatManual, want: "https://example.test/blog/reading-slowly/2026-11-10.txt"},
		{name: "an older version as the reader asks", target: older, format: FormatChosen, want: "https://example.test/blog/reading-slowly/2026-11-10"},
		{name: "a heading of an older version", target: document.Anchor{Page: older, Name: "in-a-terminal"}, format: FormatChosen, want: "https://example.test/blog/reading-slowly/2026-11-10#in-a-terminal"},
		{name: "a picture an older version shows", target: document.File{Page: older, Name: "margin.png"}, format: FormatChosen, want: "https://example.test/blog/reading-slowly/margin.png"},
		{name: "the blog's list in markdown", target: list, format: FormatMarkdown, want: "https://example.test/blog.md"},
		{name: "the blog's list as a manual", target: list, format: FormatManual, want: "https://example.test/blog.txt"},
		{name: "the blog's list as the reader asks", target: list, format: FormatChosen, want: "https://example.test/blog"},
		{name: "the list by series in markdown", target: listBySeries, format: FormatMarkdown, want: "https://example.test/blog/series.md"},
		{name: "the list by series as a manual", target: listBySeries, format: FormatManual, want: "https://example.test/blog/series.txt"},
		{name: "the list by series as the reader asks", target: listBySeries, format: FormatChosen, want: "https://example.test/blog/series"},
		{name: "the list filtered by a tag in markdown", target: document.TagFilter("go"), format: FormatMarkdown, want: "https://example.test/blog.md?tag=go"},
		{name: "the list filtered by a tag as a manual", target: document.TagFilter("go"), format: FormatManual, want: "https://example.test/blog.txt?tag=go"},
		{name: "the list filtered by a tag as the reader asks", target: document.TagFilter("go"), format: FormatChosen, want: "https://example.test/blog?tag=go"},
		{name: "a filtered list in no format", target: document.TagFilter("go"), format: 0, want: ""},
		{name: "no page", target: document.Page{Kind: document.PageNone, Slug: "", Tab: ""}, format: FormatMarkdown, want: ""},
		{name: "a post's picture, in any format", target: document.File{Page: post, Name: "margin.png"}, format: FormatManual, want: "https://example.test/blog/reading-slowly/margin.png"},
		{name: "a file of a page that serves none", target: document.File{Page: project, Name: "margin.png"}, format: FormatChosen, want: ""},
		{name: "a heading of a post in markdown", target: document.Anchor{Page: post, Name: "in-a-terminal"}, format: FormatMarkdown, want: "https://example.test/blog/reading-slowly.md#in-a-terminal"},
		{name: "a heading of a post as the reader asks", target: document.Anchor{Page: post, Name: "in-a-terminal"}, format: FormatChosen, want: "https://example.test/blog/reading-slowly#in-a-terminal"},
		{name: "a section of a page with no address", target: document.Anchor{Page: index, Name: "in-a-terminal"}, format: FormatChosen, want: ""},
		{name: "an address outside the site", target: document.Address{URL: "https://github.test/mira"}, format: FormatMarkdown, want: "https://github.test/mira"},
		{name: "a mailbox, under the scheme that writes to it", target: document.Mail{Address: "mira@example.test"}, format: FormatMarkdown, want: "mailto:mira@example.test"},
		{name: "the blog's feed, in any format", target: document.Feed{Series: ""}, format: FormatMarkdown, want: "https://example.test/blog.atom"},
		{name: "a series' feed, in any format", target: document.Feed{Series: "stint"}, format: FormatChosen, want: "https://example.test/blog/series/stint.atom"},
		{name: "a translation as the reader asks", target: translated, format: FormatChosen, want: "https://example.test/ru/blog/reading-slowly"},
		{name: "a translation in markdown", target: translated, format: FormatMarkdown, want: "https://example.test/ru/blog/reading-slowly.md"},
		{name: "a translation as a manual", target: translated, format: FormatManual, want: "https://example.test/ru/blog/reading-slowly.txt"},
		{name: "a translation's older version", target: translated.OlderVersion("2026-11-16"), format: FormatChosen, want: "https://example.test/ru/blog/reading-slowly/2026-11-16"},
		{name: "a translation's heading", target: document.Anchor{Page: translated, Name: "потом"}, format: FormatChosen, want: "https://example.test/ru/blog/reading-slowly#потом"},
		{name: "a translation's picture", target: document.File{Page: translated, Name: "margin.png"}, format: FormatChosen, want: "https://example.test/ru/blog/reading-slowly/margin.png"},
		{name: "the blog's list in a language", target: list.InLanguage("ru"), format: FormatChosen, want: "https://example.test/ru/blog"},
		{name: "the list by series in a language, in markdown", target: listBySeries.InLanguage("ru"), format: FormatMarkdown, want: "https://example.test/ru/blog/series.md"},
		{name: "the list in a language filtered by a tag", target: document.TagFilter("go").InLanguage("ru"), format: FormatChosen, want: "https://example.test/ru/blog?tag=go"},
		{name: "the blog's feed in a language", target: document.Feed{Series: "", Translation: "ru"}, format: FormatChosen, want: "https://example.test/ru/blog.atom"},
		{name: "a series' feed in a language", target: document.Feed{Series: "stint", Translation: "ru"}, format: FormatChosen, want: "https://example.test/ru/blog/series/stint.atom"},
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
		{
			name: "a post's picture", target: document.File{Page: document.Page{Kind: document.PagePost, Slug: "reading-slowly", Tab: ""}, Name: "margin.png"},
			format: FormatChosen, want: "/blog/reading-slowly/margin.png",
		},
		{
			name: "a heading of a post", target: document.Anchor{Page: document.Page{Kind: document.PagePost, Slug: "reading-slowly", Tab: ""}, Name: "in-a-terminal"},
			format: FormatChosen, want: "/blog/reading-slowly#in-a-terminal",
		},
		{name: "a section of a hidden page", target: document.Anchor{Page: document.Page{Kind: document.PageHealth, Slug: "", Tab: ""}, Name: "go"}, format: FormatChosen, want: ""},
		{name: "an address outside the site", target: document.Address{URL: "https://github.test/mira"}, format: FormatChosen, want: "https://github.test/mira"},
		{name: "a mailbox, under the scheme that writes to it", target: document.Mail{Address: "mira@example.test"}, format: FormatChosen, want: "mailto:mira@example.test"},
		{name: "a series' feed", target: document.Feed{Series: "stint"}, format: FormatChosen, want: "/blog/series/stint.atom"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := Reference(testCase.target, testCase.format); got != testCase.want {
				test.Errorf("Reference(%+v) = %q, want %q", testCase.target, got, testCase.want)
			}
		})
	}
}

// The address a test's site is served at
const testedBase = "https://example.test"

// A page of the site is read in a terminal with curl and its whole address, which answers a terminal with the manual.
func TestATerminalReadsAPageByItsWholeAddress(test *testing.T) {
	site := New(writtenAddress{written: "https://example.test", host: "example.test"}, testedAuthority, testedTranslations())
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

// A thing a page's linked data names is known by the page's whole address and its name after it, in the page's
// language's folder where the page is translated.
func TestAThingOfAPagesRecordIsKnownByThePagesAddress(test *testing.T) {
	site := New(writtenAddress{written: "https://example.test", host: "example.test"}, testedAuthority, testedTranslations())
	cases := []struct {
		name  string
		page  document.Page
		thing string
		want  string
	}{
		{name: "the person on the résumé", page: document.PageOf(document.PageResume), thing: "person", want: "https://example.test/#person"},
		{name: "a thing on a translated post", page: document.PostPage("reading-slowly").InLanguage("ru"), thing: "post", want: "https://example.test/ru/blog/reading-slowly#post"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := site.Identifier(testCase.page, testCase.thing); got != testCase.want {
				test.Errorf("Identifier(%+v, %q) = %q, want %q", testCase.page, testCase.thing, got, testCase.want)
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
		{name: "a post by its number", path: NumberPath("", 12), want: "blog/12"},
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
// by it, in the site's own language or in the folder of one the site is translated into, and no other's.
func TestABlogsNameIsItsListsOrInItsFolder(test *testing.T) {
	cases := []struct {
		name        string
		translation string
		isBlog      bool
	}{
		{name: "blog", translation: "", isBlog: true},
		{name: "blog/99", translation: "", isBlog: true},
		{name: "blog?tag=nowhere", translation: "", isBlog: true},
		{name: "ru/blog", translation: "ru", isBlog: true},
		{name: "ru/blog/99", translation: "ru", isBlog: true},
		{name: "ru/blog?tag=nowhere", translation: "ru", isBlog: true},
		{name: "de/blog", translation: "", isBlog: false},
		{name: "ru/blogger", translation: "", isBlog: false},
		{name: "blogger", translation: "", isBlog: false},
		{name: "blogger?tag=go", translation: "", isBlog: false},
		{name: "reader", translation: "", isBlog: false},
		{name: "p/tidy-tool", translation: "", isBlog: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			translation, isBlog := testedSite().BlogNamed(testCase.name)
			if translation != testCase.translation || isBlog != testCase.isBlog {
				test.Errorf("BlogNamed(%q) = %q, %t, want %q, %t", testCase.name, translation, isBlog, testCase.translation, testCase.isBlog)
			}
		})
	}
}

// A request reaches the blog's list filtered by the tag its query names, in each format, at the address Path writes;
// every other key passes over, as does a query of any other path. A query naming a tag twice reaches an address of its
// tags alone, which no page has.
func TestARequestsQueryNamesTheFilterOfTheList(test *testing.T) {
	cases := []struct {
		name  string
		path  string
		query url.Values
		want  string
	}{
		{name: "the list", path: "/blog", query: url.Values{}, want: "/blog"},
		{name: "a tag", path: "/blog", query: url.Values{"tag": {"go"}}, want: Path(document.TagFilter("go"), FormatChosen)},
		{name: "a tag in markdown", path: "/blog.md", query: url.Values{"tag": {"go"}}, want: Path(document.TagFilter("go"), FormatMarkdown)},
		{name: "a tag among other keys", path: "/blog", query: url.Values{"tag": {"go"}, "ref": {"feed"}}, want: "/blog?tag=go"},
		{name: "other keys alone", path: "/blog", query: url.Values{"ref": {"feed"}}, want: "/blog"},
		{name: "a tag twice", path: "/blog", query: url.Values{"tag": {"go", "rust"}}, want: "/blog?tag=go&tag=rust"},
		{name: "a tag of the list by series", path: "/blog/series", query: url.Values{"tag": {"go"}}, want: "/blog/series"},
		{name: "a tag of the list in a language", path: "/ru/blog", query: url.Values{"tag": {"go"}}, want: Path(document.TagFilter("go").InLanguage("ru"), FormatChosen)},
		{name: "a tag in a language the site is not translated into", path: "/de/blog", query: url.Values{"tag": {"go"}}, want: "/de/blog"},
		{name: "a tag of the résumé", path: "/", query: url.Values{"tag": {"go"}}, want: "/"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := testedSite().Requested(testCase.path, testCase.query); got != testCase.want {
				test.Errorf("Requested(%q, %v) = %q, want %q", testCase.path, testCase.query, got, testCase.want)
			}
		})
	}
}

// The address a request asked for reads as its reader wrote it: its path, decoded, and its query as written, its keys
// in their order and none left out; a path alone where it has no query.
func TestTheAddressAskedForReadsAsWritten(test *testing.T) {
	cases := []struct {
		name     string
		path     string
		rawQuery string
		want     string
	}{
		{name: "a path alone", path: "/no thing", rawQuery: "", want: "/no thing"},
		{name: "a path and its query", path: "/blog", rawQuery: "tag=go&ref=feed", want: "/blog?tag=go&ref=feed"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := Asked(testCase.path, testCase.rawQuery); got != testCase.want {
				test.Errorf("Asked(%q, %q) = %q, want %q", testCase.path, testCase.rawQuery, got, testCase.want)
			}
		})
	}
}

// The tag authority a test's site names what it serves under
const testedAuthority = "example.test,2026"

// testedTranslations are the languages a test's site is translated into.
func testedTranslations() []string {
	return []string{"ru"}
}

// testedSite is a test's site, served at its base address.
func testedSite() Site {
	return New(writtenAddress{written: testedBase, host: "example.test"}, testedAuthority, testedTranslations())
}

// A feed and a post are named for good by their paths under the site's tag authority, a post by the path its number
// reaches it at, and keep the name wherever the site is served.
func TestAFeedAndAPostAreNamedForGoodWhereverTheSiteIs(test *testing.T) {
	cases := []struct {
		name string
		path string
		want string
	}{
		{name: "the blog's feed", path: FeedPath(document.Feed{Series: ""}), want: "tag:example.test,2026:blog.atom"},
		{name: "a series' feed", path: FeedPath(document.Feed{Series: "stint"}), want: "tag:example.test,2026:blog/series/stint.atom"},
		{name: "a post, by its number", path: NumberPath("", 12), want: "tag:example.test,2026:blog/12"},
		{name: "the blog's feed in a language", path: FeedPath(document.Feed{Series: "", Translation: "ru"}), want: "tag:example.test,2026:ru/blog.atom"},
		{name: "a translation, by its post's number", path: NumberPath("ru", 12), want: "tag:example.test,2026:ru/blog/12"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			for _, base := range []writtenAddress{{written: testedBase, host: "example.test"}, {written: "https://moved.test", host: "moved.test"}} {
				if got := New(base, testedAuthority, testedTranslations()).TagURI(testCase.path); got != testCase.want {
					test.Errorf("TagURI(%q) served at %s = %q, want %q", testCase.path, base.written, got, testCase.want)
				}
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
