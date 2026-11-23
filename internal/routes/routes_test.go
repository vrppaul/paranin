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
		{name: "the list filtered by a tag in markdown", target: document.TagFilter("go"), format: FormatMarkdown, want: "https://example.test/blog.md?tag=go"},
		{name: "the list filtered by a tag as a manual", target: document.TagFilter("go"), format: FormatManual, want: "https://example.test/blog.txt?tag=go"},
		{name: "the list filtered by a tag as the reader asks", target: document.TagFilter("go"), format: FormatChosen, want: "https://example.test/blog?tag=go"},
		{name: "the list filtered by a series in markdown", target: document.SeriesFilter("stint"), format: FormatMarkdown, want: "https://example.test/blog.md?series=stint"},
		{name: "the list filtered by a series as a manual", target: document.SeriesFilter("stint"), format: FormatManual, want: "https://example.test/blog.txt?series=stint"},
		{
			name: "the list filtered by a series as the reader asks", target: document.SeriesFilter("stint"), format: FormatChosen,
			want: "https://example.test/blog?series=stint",
		},
		{name: "a filtered list in no format", target: document.TagFilter("go"), format: 0, want: ""},
		{name: "no page", target: document.Page{Kind: document.PageNone, Slug: "", Tab: ""}, format: FormatMarkdown, want: ""},
		{name: "a post's picture, in any format", target: document.File{Page: post, Name: "margin.png"}, format: FormatManual, want: "https://example.test/blog/reading-slowly/margin.png"},
		{name: "a file of a page that serves none", target: document.File{Page: project, Name: "margin.png"}, format: FormatChosen, want: ""},
		{name: "a heading of a post in markdown", target: document.Anchor{Page: post, Name: "in-a-terminal"}, format: FormatMarkdown, want: "https://example.test/blog/reading-slowly.md#in-a-terminal"},
		{name: "a heading of a post as the reader asks", target: document.Anchor{Page: post, Name: "in-a-terminal"}, format: FormatChosen, want: "https://example.test/blog/reading-slowly#in-a-terminal"},
		{name: "a section of a page with no address", target: document.Anchor{Page: index, Name: "in-a-terminal"}, format: FormatChosen, want: ""},
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
	}
	site := New(writtenAddress{written: "https://example.test", host: "example.test"})
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := site.Reference(testCase.target, testCase.format); got != testCase.want {
				test.Errorf("Reference(%+v) = %q, want %q", testCase.target, got, testCase.want)
			}
		})
	}
}

// The address a test's site is served at, and the token a test's preview link holds
const (
	testedBase   = "https://example.test"
	previewToken = "c2lnbmVk"
)

// A post read at its preview link has every address of its own in the preview's folder, under its token, in each
// format, its sections and its pictures among them, whole and as a path from the root alike; every other page keeps
// its own address.
func TestAPreviewIsAddressedUnderItsToken(test *testing.T) {
	site := New(writtenAddress{written: testedBase, host: "example.test"}).Previewing("coming", previewToken)
	previewed := document.Page{Kind: document.PagePost, Slug: "coming", Tab: ""}
	cases := []struct {
		name     string
		target   document.Target
		format   Format
		wantPath string
	}{
		{name: "the post as the reader asks", target: previewed, format: FormatChosen, wantPath: "/preview/c2lnbmVk/coming"},
		{name: "the post in markdown", target: previewed, format: FormatMarkdown, wantPath: "/preview/c2lnbmVk/coming.md"},
		{name: "the post as a manual", target: previewed, format: FormatManual, wantPath: "/preview/c2lnbmVk/coming.txt"},
		{name: "a heading of the post", target: document.Anchor{Page: previewed, Name: "why"}, format: FormatChosen, wantPath: "/preview/c2lnbmVk/coming#why"},
		{name: "a picture of the post", target: document.File{Page: previewed, Name: "margin.png"}, format: FormatChosen, wantPath: "/preview/c2lnbmVk/coming/margin.png"},
		{name: "another post", target: document.Page{Kind: document.PagePost, Slug: "reading-slowly", Tab: ""}, format: FormatChosen, wantPath: "/blog/reading-slowly"},
		{name: "the résumé", target: document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, format: FormatChosen, wantPath: "/"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			address, reference := site.Address(testCase.target, testCase.format), site.Reference(testCase.target, testCase.format)
			if address != testedBase+testCase.wantPath || reference != testCase.wantPath {
				test.Errorf("Address and Reference of %+v = %q and %q, want %q and %q", testCase.target, address, reference,
					testedBase+testCase.wantPath, testCase.wantPath)
			}
		})
	}
}

// A path in the preview's folder reads back as the token it holds and the path the same face or file of its post has
// in the blog's folder; a path anywhere else, or naming no token or no post, is no preview's.
func TestAPreviewsPathReadsBackAsItsTokenAndItsPost(test *testing.T) {
	none := PreviewLink{Token: "", Slug: "", PostPath: ""}
	cases := []struct {
		name     string
		path     string
		want     PreviewLink
		wantRead bool
	}{
		{name: "the post", path: "/preview/c2lnbmVk/coming", want: PreviewLink{Token: previewToken, Slug: "coming", PostPath: "/blog/coming"}, wantRead: true},
		{name: "its markdown", path: "/preview/c2lnbmVk/coming.md", want: PreviewLink{Token: previewToken, Slug: "coming", PostPath: "/blog/coming.md"}, wantRead: true},
		{
			name: "its picture", path: "/preview/c2lnbmVk/coming/margin.png",
			want: PreviewLink{Token: previewToken, Slug: "coming", PostPath: "/blog/coming/margin.png"}, wantRead: true,
		},
		{name: "a post's own address", path: "/blog/coming", want: none, wantRead: false},
		{name: "the folder with no post", path: "/preview/c2lnbmVk/", want: none, wantRead: false},
		{name: "the folder with no token", path: "/preview//coming", want: none, wantRead: false},
		{name: "a token alone", path: "/preview/c2lnbmVk", want: none, wantRead: false},
		{name: "a picture with no post", path: "/preview/c2lnbmVk//margin.png", want: none, wantRead: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got, isRead := PreviewPath(testCase.path); got != testCase.want || isRead != testCase.wantRead {
				test.Errorf("PreviewPath(%q) = %+v, read: %t; want %+v, read: %t", testCase.path, got, isRead, testCase.want, testCase.wantRead)
			}
		})
	}
}

// A path is in the preview's folder by its opening alone, whatever follows, so every answer there is sent as a preview
// link's is; the preview's name alone, or the blog's folder, is not.
func TestAPathIsInThePreviewsFolderByItsOpening(test *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{path: "/preview/c2lnbmVk/coming", want: true},
		{path: "/preview/c2lnbmVk", want: true},
		{path: "/preview/", want: true},
		{path: "/preview", want: false},
		{path: "/previews/coming", want: false},
		{path: "/blog/coming", want: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.path, func(test *testing.T) {
			if got := InPreviewFolder(testCase.path); got != testCase.want {
				test.Errorf("InPreviewFolder(%q) = %t, want %t", testCase.path, got, testCase.want)
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
		{name: "blog?tag=nowhere", want: true},
		{name: "blogger", want: false},
		{name: "blogger?tag=go", want: false},
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

// A request reaches the blog's list filtered by the one filter its query names, by tag or by series, in each format,
// at the address Path writes; every other key passes over, as does a query of any other path. A query naming a filter
// twice, or both, reaches an address of its filters alone, which no page has.
func TestARequestsQueryNamesTheFilterOfTheList(test *testing.T) {
	cases := []struct {
		name  string
		path  string
		query url.Values
		want  string
	}{
		{name: "the list", path: "/blog", query: url.Values{}, want: "/blog"},
		{name: "a tag", path: "/blog", query: url.Values{"tag": {"go"}}, want: Path(document.TagFilter("go"), FormatChosen)},
		{name: "a series", path: "/blog", query: url.Values{"series": {"stint"}}, want: Path(document.SeriesFilter("stint"), FormatChosen)},
		{name: "a tag in markdown", path: "/blog.md", query: url.Values{"tag": {"go"}}, want: Path(document.TagFilter("go"), FormatMarkdown)},
		{name: "a series as a manual", path: "/blog.txt", query: url.Values{"series": {"stint"}}, want: Path(document.SeriesFilter("stint"), FormatManual)},
		{name: "a tag among other keys", path: "/blog", query: url.Values{"tag": {"go"}, "ref": {"feed"}}, want: "/blog?tag=go"},
		{name: "other keys alone", path: "/blog", query: url.Values{"ref": {"feed"}}, want: "/blog"},
		{name: "a tag twice", path: "/blog", query: url.Values{"tag": {"go", "rust"}}, want: "/blog?tag=go&tag=rust"},
		{name: "a tag and a series", path: "/blog", query: url.Values{"tag": {"go"}, "series": {"stint"}}, want: "/blog?series=stint&tag=go"},
		{name: "a tag of the list by series", path: "/blog/series", query: url.Values{"tag": {"go"}}, want: "/blog/series"},
		{name: "a tag of the résumé", path: "/", query: url.Values{"tag": {"go"}}, want: "/"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := Requested(testCase.path, testCase.query); got != testCase.want {
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
		{name: "a path and its query", path: "/blog", rawQuery: "tag=go&series=stint&ref=feed", want: "/blog?tag=go&series=stint&ref=feed"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := Asked(testCase.path, testCase.rawQuery); got != testCase.want {
				test.Errorf("Asked(%q, %q) = %q, want %q", testCase.path, testCase.rawQuery, got, testCase.want)
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
