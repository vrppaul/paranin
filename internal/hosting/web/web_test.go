package web

import (
	"bytes"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/vrppaul/paranin/internal/assets"
	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/release"
)

// The build's fixture's stylesheet and script, which every page a browser reads includes
const (
	stylesheetPath = "/assets/site-fixture.css"
	scriptPath     = "/assets/site-fixture.js"
)

// The cards' fixture's card of every page but the résumé, where the site serves it
const cardPath = "/cards/page-fixture.png"

// The fixture post's slug, and its picture: its file in the content's folder, and where the site serves it
const (
	fixturePostSlug = "reading-slowly"
	pictureFile     = "posts/reading-slowly/margin.png"
	picturePath     = "/blog/reading-slowly/margin.png"
)

// Each page is served at its address in each format, the index at llms.txt as plain text, as agents read it, and
// robots.txt and the sitemap at the paths a crawler asks for them by.
func TestEveryPageIsServedAtItsAddress(test *testing.T) {
	server := serverOf(fixtureStore(test, fixtures.ContentFolder(test)), false)
	cases := []struct {
		name        string
		path        string
		status      int
		contentType string
		holds       string
	}{
		{name: "the résumé", path: "/index.md", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "# Mira Example"},
		{name: "a project's page", path: "/p/tidy-tool.md", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "# tidy-tool"},
		{name: "the résumé as a manual", path: "/index.txt", status: http.StatusOK, contentType: "text/plain; charset=utf-8", holds: "EXAMPLE.TEST(1)"},
		{name: "a project's page as a manual", path: "/p/tidy-tool.txt", status: http.StatusOK, contentType: "text/plain; charset=utf-8", holds: "TIDY-TOOL(1)"},
		{name: "the index", path: "/llms.txt", status: http.StatusOK, contentType: "text/plain; charset=utf-8", holds: "- [tidy-tool](https://example.test/p/tidy-tool.md): Tidies files."},
		{name: "robots.txt", path: "/robots.txt", status: http.StatusOK, contentType: "text/plain; charset=utf-8", holds: "Sitemap: https://example.test/sitemap.xml"},
		{
			name: "the sitemap", path: "/sitemap.xml", status: http.StatusOK, contentType: "application/xml; charset=utf-8",
			holds: "<loc>https://example.test/blog/reading-slowly</loc>",
		},
		{name: "a post", path: "/blog/reading-slowly.md", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "# Reading slowly"},
		{name: "a post as a manual", path: "/blog/reading-slowly.txt", status: http.StatusOK, contentType: "text/plain; charset=utf-8", holds: "READING-SLOWLY(1)"},
		{name: "an older version", path: "/blog/reading-slowly/2026-11-10.md", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "# Slow reading"},
		{
			name: "an older version as a manual", path: "/blog/reading-slowly/2026-11-10.txt", status: http.StatusOK, contentType: "text/plain; charset=utf-8",
			holds: "newer: 2026-11-12",
		},
		{name: "an older version, chosen", path: "/blog/reading-slowly/2026-11-10", status: http.StatusOK, contentType: "text/html; charset=utf-8", holds: "Slow reading"},
		{name: "a post's picture beside its older version", path: "/blog/reading-slowly/margin.png", status: http.StatusOK, contentType: "image/png", holds: ""},
		{name: "the blog's list", path: "/blog.md", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "## 2026"},
		{name: "the blog's list as a manual", path: "/blog.txt", status: http.StatusOK, contentType: "text/plain; charset=utf-8", holds: "BLOG(1)"},
		{name: "the list by series", path: "/blog/series.md", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "1 of 2 out"},
		{name: "the list by series as a manual", path: "/blog/series.txt", status: http.StatusOK, contentType: "text/plain; charset=utf-8", holds: "BLOG-SERIES(1)"},
		{name: "the list filtered by a tag", path: "/blog.md?tag=terminals", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "posts tagged terminals"},
		{name: "the list filtered, chosen", path: "/blog?tag=reading", status: http.StatusOK, contentType: "text/html; charset=utf-8", holds: "blog — posts tagged reading"},
		{
			name: "the list filtered, its other keys passed over", path: "/blog.md?ref=feed&tag=reading", status: http.StatusOK,
			contentType: "text/markdown; charset=utf-8", holds: "posts tagged reading",
		},
		{name: "the list, its other keys passed over", path: "/blog.md?ref=feed", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "## 2026"},
		{
			name: "the blog's feed", path: "/blog.atom", status: http.StatusOK, contentType: "application/atom+xml; charset=utf-8",
			holds: `<feed xmlns="http://www.w3.org/2005/Atom" xml:lang="en">`,
		},
		{
			name: "a series' feed, each post's body whole", path: "/blog/series/reading.atom", status: http.StatusOK,
			contentType: "application/atom+xml; charset=utf-8", holds: "&lt;p&gt;A page read twice is read once",
		},
		{name: "the feed of no series", path: "/blog/series/aloud.atom", status: http.StatusNotFound, contentType: "text/html; charset=utf-8", holds: ""},
		{name: "a translation", path: "/ru/blog/reading-slowly.md", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "# Медленное чтение"},
		{name: "a translation as a manual", path: "/ru/blog/reading-slowly.txt", status: http.StatusOK, contentType: "text/plain; charset=utf-8", holds: "Медленное чтение"},
		{name: "a translation's picture", path: "/ru/blog/reading-slowly/margin.png", status: http.StatusOK, contentType: "image/png", holds: ""},
		{name: "the list in a language", path: "/ru/blog.md", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "# блог"},
		{name: "the list in a language, filtered", path: "/ru/blog.md?tag=reading", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "посты с тегом reading"},
		{
			name: "the blog's feed in a language", path: "/ru/blog.atom", status: http.StatusOK, contentType: "application/atom+xml; charset=utf-8",
			holds: `<id>tag:example.test,2026:ru/blog/1</id>`,
		},
		{name: "a post in a language it has no translation in", path: "/ru/blog/notes-on-tools.md", status: http.StatusNotFound, contentType: "text/html; charset=utf-8", holds: ""},
		{name: "the list in a language the site has no words for", path: "/de/blog.md", status: http.StatusNotFound, contentType: "text/html; charset=utf-8", holds: ""},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := serve(server, http.MethodGet, testCase.path, "", "")
			if response.Code != testCase.status || response.Header().Get("Content-Type") != testCase.contentType ||
				!strings.Contains(response.Body.String(), testCase.holds) {
				test.Errorf("GET %s = %d %q holding %q, want %d %q holding %q", testCase.path, response.Code,
					response.Header().Get("Content-Type"), response.Body.String(), testCase.status, testCase.contentType, testCase.holds)
			}
		})
	}
}

// A tab of a page is served alone at its own address in each format, and as the reader asks at the address it is
// known by: what it holds, with none of the tabs beside it.
func TestATabIsServedAlone(test *testing.T) {
	server := serverOf(fixtureStore(test, fixtures.ContentFolder(test)), false)
	const (
		markdown = "text/markdown; charset=utf-8"
		manual   = "text/plain; charset=utf-8"
		browsed  = "text/html; charset=utf-8"
		// a line of the résumé's opening tab
		resumeOpening = "Writes software for people who read it."
	)
	cases := []struct {
		name        string
		path        string
		accept      string
		contentType string
		holds       string
		lacks       string
	}{
		{name: "a tab of the résumé in markdown", path: "/experience.md", accept: "", contentType: markdown, holds: "# Experience\n\nShips tools a team", lacks: resumeOpening},
		{name: "a tab of the résumé as a manual", path: "/experience.txt", accept: "", contentType: manual, holds: "Northwind", lacks: resumeOpening},
		{name: "a tab of the résumé, chosen as a manual", path: "/experience", accept: "text/plain", contentType: manual, holds: "Northwind", lacks: resumeOpening},
		{name: "a tab of the résumé in a browser", path: "/experience", accept: "text/html", contentType: browsed, holds: "Northwind", lacks: resumeOpening},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := serve(server, http.MethodGet, testCase.path, testCase.accept, "")
			body := response.Body.String()
			if response.Code != http.StatusOK || response.Header().Get("Content-Type") != testCase.contentType ||
				!strings.Contains(body, testCase.holds) || strings.Contains(body, testCase.lacks) {
				test.Errorf("GET %s, Accept %q = %d %q %q; want %d %q holding %q and not %q", testCase.path, testCase.accept,
					response.Code, response.Header().Get("Content-Type"), body, http.StatusOK, testCase.contentType, testCase.holds, testCase.lacks)
			}
		})
	}
}

// The agents a test's request is sent by: a terminal's client, and one no list names; and the Accept a browser sends
const (
	curlAgent     = "curl/8.5.0"
	unknownAgent  = "ExampleBot/1.0 (+https://example.test/bot)"
	browserAccept = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
)

// What a negotiated answer says a cache keeps one copy for each of
const negotiatedVary = "Accept, User-Agent"

// At a page's own address the reader's request chooses the format: where it names none by its exact type but to refuse
// it, its agent's own, the manual for a terminal's client and HTML for anyone else; otherwise the format it ranks
// highest, one it names before one a wildcard reaches, markdown before the manual before HTML among those it names, and
// the agent's own first among those a wildcard reaches; and a cache keeps one copy for each Accept and User-Agent.
func TestTheReaderChoosesTheFormat(test *testing.T) {
	server := serverOf(fixtureStore(test, fixtures.ContentFolder(test)), false)
	const (
		markdown = "text/markdown; charset=utf-8"
		manual   = "text/plain; charset=utf-8"
		browsed  = "text/html; charset=utf-8"
		// how a page of HTML opens, which face/html's tests hold whole
		pageOpening = "<!doctype html>"
	)
	cases := []struct {
		name        string
		path        string
		accept      string
		userAgent   string
		contentType string
		holds       string
	}{
		{name: "anything, from an agent no list names", path: "/", accept: "*/*", userAgent: unknownAgent, contentType: browsed, holds: pageOpening},
		{name: "nothing asked, from an agent no list names", path: "/", accept: "", userAgent: unknownAgent, contentType: browsed, holds: pageOpening},
		{name: "nothing asked, from no agent", path: "/", accept: "", userAgent: "", contentType: browsed, holds: pageOpening},
		{name: "curl, taking anything", path: "/", accept: "*/*", userAgent: curlAgent, contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "curl, asking for nothing", path: "/", accept: "", userAgent: curlAgent, contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "Wget, taking anything", path: "/", accept: "*/*", userAgent: "Wget/1.21", contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "Wget, asking for nothing", path: "/", accept: "", userAgent: "Wget/1.21", contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "HTTPie, taking anything", path: "/", accept: "*/*", userAgent: "HTTPie/3.2", contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "HTTPie, asking for nothing", path: "/", accept: "", userAgent: "HTTPie/3.2", contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "xh, taking anything", path: "/", accept: "*/*", userAgent: "xh/0.22", contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "xh, asking for nothing", path: "/", accept: "", userAgent: "xh/0.22", contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "a terminal's client in capitals", path: "/", accept: "*/*", userAgent: "CURL/8.5.0", contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "an agent whose name opens with a terminal's", path: "/", accept: "*/*", userAgent: "curlish/1.0", contentType: browsed, holds: pageOpening},
		{name: "an agent asking for markdown", path: "/", accept: "text/markdown", userAgent: "", contentType: markdown, holds: "# Mira Example"},
		{name: "a terminal asking for markdown", path: "/", accept: "text/markdown", userAgent: curlAgent, contentType: markdown, holds: "# Mira Example"},
		{name: "a request asking for plain text", path: "/", accept: "text/plain", userAgent: "", contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "a terminal asking for HTML before anything", path: "/", accept: "text/html, */*", userAgent: curlAgent, contentType: browsed, holds: pageOpening},
		{name: "HTML before anything, from an agent no list names", path: "/", accept: "text/html, */*", userAgent: unknownAgent, contentType: browsed, holds: pageOpening},
		{name: "markdown and HTML before anything", path: "/", accept: "text/markdown, text/html, */*", userAgent: "", contentType: markdown, holds: "# Mira Example"},
		{name: "markdown refused", path: "/", accept: "text/markdown;q=0, */*", userAgent: unknownAgent, contentType: browsed, holds: pageOpening},
		{name: "markdown refused, by a terminal", path: "/", accept: "text/markdown;q=0, */*", userAgent: curlAgent, contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "HTML refused, and anything taken", path: "/", accept: "text/html;q=0, */*", userAgent: unknownAgent, contentType: markdown, holds: "# Mira Example"},
		{name: "HTML refused by a terminal, and nothing named", path: "/", accept: "text/html;q=0", userAgent: curlAgent, contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "markdown among others", path: "/", accept: "text/html;q=0.9, text/markdown", userAgent: "", contentType: markdown, holds: "# Mira Example"},
		{name: "plain text ranked above markdown", path: "/", accept: "text/plain, text/markdown;q=0.1", userAgent: "", contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "markdown ranked above plain text", path: "/", accept: "text/markdown, text/plain;q=0.5", userAgent: "", contentType: markdown, holds: "# Mira Example"},
		{name: "all of a family", path: "/", accept: "text/*", userAgent: unknownAgent, contentType: browsed, holds: pageOpening},
		{name: "all of a family, by a terminal", path: "/", accept: "text/*", userAgent: curlAgent, contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "markdown below anything", path: "/", accept: "text/markdown;q=0.5, */*", userAgent: unknownAgent, contentType: browsed, holds: pageOpening},
		{name: "markdown below anything, by a terminal", path: "/", accept: "text/markdown;q=0.5, */*", userAgent: curlAgent, contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "markdown below its family", path: "/", accept: "text/markdown;q=0.5, text/*", userAgent: unknownAgent, contentType: browsed, holds: pageOpening},
		{name: "a browser asking for HTML", path: "/", accept: "text/html", userAgent: "", contentType: browsed, holds: pageOpening},
		{name: "a browser's own Accept", path: "/", accept: browserAccept, userAgent: "", contentType: browsed, holds: pageOpening},
		{name: "HTML and markdown alike", path: "/", accept: "text/html, text/markdown, text/plain;q=0.1", userAgent: "", contentType: markdown, holds: "# Mira Example"},
		{name: "a project's page in markdown", path: "/p/tidy-tool", accept: "text/markdown", userAgent: "", contentType: markdown, holds: "# tidy-tool"},
		{name: "a project's page as a manual", path: "/p/tidy-tool", accept: "*/*", userAgent: curlAgent, contentType: manual, holds: "TIDY-TOOL(1)"},
		{name: "a project's page in a browser", path: "/p/tidy-tool", accept: "text/html", userAgent: "", contentType: browsed, holds: `<h1 aria-label="tidy-tool">`},
		{name: "a post in markdown", path: "/blog/reading-slowly", accept: "text/markdown", userAgent: "", contentType: markdown, holds: "# Reading slowly"},
		{name: "a post as a manual", path: "/blog/reading-slowly", accept: "*/*", userAgent: curlAgent, contentType: manual, holds: "READING-SLOWLY(1)"},
		{name: "a post in a browser", path: "/blog/reading-slowly", accept: browserAccept, userAgent: "", contentType: browsed, holds: `<h1 class="post-title">`},
		{name: "the blog's list in markdown", path: "/blog", accept: "text/markdown", userAgent: "", contentType: markdown, holds: "# blog"},
		{name: "the blog's list in a browser", path: "/blog", accept: "text/html", userAgent: "", contentType: browsed, holds: `<a class="switch-way" href="/blog" aria-current="page">`},
		{
			name: "the list by series in a browser", path: "/blog/series", accept: "text/html", userAgent: "", contentType: browsed,
			holds: `<a class="switch-way" href="/blog/series" aria-current="page">`,
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := serve(server, http.MethodGet, testCase.path, testCase.accept, testCase.userAgent)
			header := response.Header()
			if header.Get("Content-Type") != testCase.contentType || !strings.Contains(response.Body.String(), testCase.holds) ||
				header.Get("Vary") != negotiatedVary {
				test.Errorf("GET %s, Accept %q, User-Agent %q = %q %q, Vary %q; want %q holding %q, Vary %q", testCase.path,
					testCase.accept, testCase.userAgent, header.Get("Content-Type"), response.Body.String(), header.Get("Vary"),
					testCase.contentType, testCase.holds, negotiatedVary)
			}
		})
	}
}

// Each page a browser reads carries its record as linked data: the résumé its person, a tab of it a page about the
// person, by the @id the résumé's address gives them; a project's page its code.
func TestEveryPageCarriesItsRecord(test *testing.T) {
	server := serverOf(fixtureStore(test, fixtures.ContentFolder(test)), false)
	cases := []struct {
		name  string
		path  string
		holds string
	}{
		{name: "the résumé", path: "/", holds: `"@type":"ProfilePage"`},
		{name: "a project's page", path: "/p/tidy-tool", holds: `"@type":"SoftwareSourceCode","name":"tidy-tool"`},
		{name: "a tab of the résumé, a page about its person", path: "/experience", holds: `"@type":"WebPage","url":"https://example.test/experience","about":{"@type":"Person","@id":"https://example.test/#person"`},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := serve(server, http.MethodGet, testCase.path, "text/html", "")
			if !strings.Contains(response.Body.String(), testCase.holds) {
				test.Errorf("GET %s as a browser = %q, want it to hold %q", testCase.path, response.Body.String(), testCase.holds)
			}
		})
	}
}

// Each format of a page names the address the page is known by, whole, so a search engine counts its markdown and
// its manual as that page; the index, which is no page a person reads, names none, nor a filtered list, a part of the
// list and no copy of it, nor an unlisted post, which no search engine indexes.
func TestEveryFormatOfAPageNamesItsCanonicalAddress(test *testing.T) {
	server := unlistedPostServer(test)
	cases := []struct {
		name   string
		path   string
		accept string
		want   string
	}{
		{name: "the résumé in markdown", path: "/index.md", accept: "", want: `<https://example.test/>; rel="canonical"`},
		{name: "the résumé as a manual", path: "/index.txt", accept: "", want: `<https://example.test/>; rel="canonical"`},
		{name: "the résumé, chosen", path: "/", accept: "text/html", want: `<https://example.test/>; rel="canonical"`},
		{name: "a project's page in markdown", path: "/p/tidy-tool.md", accept: "", want: `<https://example.test/p/tidy-tool>; rel="canonical"`},
		{name: "a project's page, chosen", path: "/p/tidy-tool", accept: "*/*", want: `<https://example.test/p/tidy-tool>; rel="canonical"`},
		{name: "a tab of the résumé in markdown", path: "/experience.md", accept: "", want: `<https://example.test/experience>; rel="canonical"`},
		{name: "a post in markdown", path: "/blog/reading-slowly.md", accept: "", want: `<https://example.test/blog/reading-slowly>; rel="canonical"`},
		{name: "a post, chosen", path: "/blog/reading-slowly", accept: "*/*", want: `<https://example.test/blog/reading-slowly>; rel="canonical"`},
		{name: "a post in a browser", path: "/blog/reading-slowly", accept: "text/html", want: `<https://example.test/blog/reading-slowly>; rel="canonical"`},
		{name: "a translation in markdown", path: "/ru/blog/reading-slowly.md", accept: "", want: `<https://example.test/ru/blog/reading-slowly>; rel="canonical"`},
		{name: "a filtered list in markdown, a part of the list and no copy of it", path: "/blog.md?tag=reading", accept: "", want: ""},
		{name: "a filtered list as a manual", path: "/blog.txt?tag=reading", accept: "", want: ""},
		{name: "a filtered list in a browser", path: "/blog?tag=terminals", accept: "text/html", want: ""},
		{name: "an unlisted post in a browser", path: unlistedPostPath, accept: "text/html", want: ""},
		{name: "an unlisted post in markdown", path: unlistedPostPath + ".md", accept: "", want: ""},
		{name: "an unlisted post as a manual", path: unlistedPostPath + ".txt", accept: "", want: ""},
		{name: "the index", path: "/llms.txt", accept: "", want: ""},
		{name: "no page", path: "/nothing", accept: "", want: ""},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := serve(server, http.MethodGet, testCase.path, testCase.accept, "")
			if got := response.Header().Get("Link"); got != testCase.want {
				test.Errorf("GET %s, Accept %q: Link = %q, want %q", testCase.path, testCase.accept, got, testCase.want)
			}
		})
	}
}

// A filtered list, a view of the list and no page of its own, an older version of a post, which a newer stands over,
// and an unlisted post, read by whoever has its address, tell a search engine not to index them, in each format; the
// list itself, and any other page, leaves it free to.
func TestAPageOfNoIndexIsSentForNoSearchEngineToIndex(test *testing.T) {
	server := unlistedPostServer(test)
	cases := []struct {
		name   string
		path   string
		accept string
		want   string
	}{
		{name: "filtered by a tag, in a browser", path: "/blog?tag=reading", accept: "text/html", want: "noindex"},
		{name: "filtered, in markdown", path: "/blog.md?tag=terminals", accept: "", want: "noindex"},
		{name: "filtered, as a manual", path: "/blog.txt?tag=terminals", accept: "", want: "noindex"},
		{name: "the list", path: "/blog", accept: "text/html", want: ""},
		{name: "the list in markdown", path: "/blog.md", accept: "", want: ""},
		{name: "a post", path: "/blog/reading-slowly", accept: "text/html", want: ""},
		{name: "an older version, in a browser", path: "/blog/reading-slowly/2026-11-10", accept: "text/html", want: "noindex"},
		{name: "an older version, in markdown", path: "/blog/reading-slowly/2026-11-10.md", accept: "", want: "noindex"},
		{name: "an older version, as a manual", path: "/blog/reading-slowly/2026-11-10.txt", accept: "", want: "noindex"},
		{name: "an unlisted post, in a browser", path: unlistedPostPath, accept: "text/html", want: "noindex"},
		{name: "an unlisted post, in markdown", path: unlistedPostPath + ".md", accept: "", want: "noindex"},
		{name: "an unlisted post, as a manual", path: unlistedPostPath + ".txt", accept: "", want: "noindex"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := serve(server, http.MethodGet, testCase.path, testCase.accept, "")
			if got := response.Header().Get("X-Robots-Tag"); response.Code != http.StatusOK || got != testCase.want {
				test.Errorf("GET %s, Accept %q = %d, X-Robots-Tag %q; want %d, %q", testCase.path, testCase.accept, response.Code, got, http.StatusOK, testCase.want)
			}
		})
	}
}

// Readers who first ask for a filtered list at the same moment, in each format, each get the whole list as one drawing
// of it, which -race holds to drawing it once without a race.
func TestAFilteredListAskedForAtOnceIsServedWhole(test *testing.T) {
	server := serverOf(fixtureStore(test, fixtures.ContentFolder(test)), false)
	formats := []struct {
		path   string
		accept string
		holds  string
	}{
		{path: "/blog?tag=reading", accept: "text/html", holds: "<title>blog — posts tagged reading</title>"},
		{path: "/blog?tag=reading", accept: "text/plain", holds: "blog — posts tagged reading"},
		{path: "/blog.md?tag=reading", accept: "", holds: "posts tagged reading"},
	}
	// how many readers ask for each format at the same moment
	const readersEach = 8
	answered := make(chan readersAnswer, readersEach*len(formats))
	var readers sync.WaitGroup
	for range readersEach {
		for _, format := range formats {
			readers.Go(func() {
				answered <- readersAnswer{holds: format.holds, body: serve(server, http.MethodGet, format.path, format.accept, "").Body.String()}
			})
		}
	}
	readers.Wait()
	close(answered)
	for got := range answered {
		if !strings.Contains(got.body, got.holds) {
			test.Errorf("a reader asking at once got %q, want it holding %q", got.body, got.holds)
		}
	}
}

// readersAnswer is what one of the readers asking at once got, and what it should hold.
type readersAnswer struct {
	holds string
	body  string
}

// A post's number sends the reader on for good to the post's own address, by its path from the root, and a cache asks
// again, since the post's slug is its folder's name, which may change; the number of a post not served is no page. The
// day of a post's current version, the name its versions list it by, sends the reader on for a while to the same face
// at the post's own address, since the day names an older version once the post is updated.
func TestAPostsNumberSendsTheReaderOnToItsAddress(test *testing.T) {
	server := serverOf(fixtureStore(test, fixtures.ContentFolder(test)), false)
	cases := []struct {
		name     string
		path     string
		status   int
		location string
	}{
		{name: "a post published", path: "/blog/1", status: http.StatusMovedPermanently, location: "/blog/reading-slowly"},
		{name: "a translation published", path: "/ru/blog/1", status: http.StatusMovedPermanently, location: "/ru/blog/reading-slowly"},
		{name: "a draft", path: "/blog/2", status: http.StatusNotFound, location: ""},
		{name: "the current version by its day", path: "/blog/reading-slowly/2026-11-12", status: http.StatusFound, location: "/blog/reading-slowly"},
		{name: "the current version by its day, in markdown", path: "/blog/reading-slowly/2026-11-12.md", status: http.StatusFound, location: "/blog/reading-slowly.md"},
		{name: "the current version by its day, as a manual", path: "/blog/reading-slowly/2026-11-12.txt", status: http.StatusFound, location: "/blog/reading-slowly.txt"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := serve(server, http.MethodGet, testCase.path, "*/*", "")
			header := response.Header()
			if response.Code != testCase.status || header.Get("Location") != testCase.location || header.Get("Cache-Control") != "no-cache" {
				test.Errorf("GET %s = %d, Location %q, Cache-Control %q; want %d, Location %q, Cache-Control %q", testCase.path,
					response.Code, header.Get("Location"), header.Get("Cache-Control"), testCase.status, testCase.location, "no-cache")
			}
		})
	}
}

// A path with a slash at its end, whose path without it is served, sends the reader on for good to that path, its query
// kept as written, or where that path sends the reader on to, straight there and for as long; a path with a slash at
// its end whose path without it is not served is no page, and the site's root is the résumé. A cache asks again, as for
// every move.
func TestAPathWithASlashAtItsEndSendsTheReaderOnToThePathWithout(test *testing.T) {
	server := serverOf(fixtureStore(test, fixtures.ContentFolder(test)), false)
	cases := []struct {
		name     string
		path     string
		status   int
		location string
	}{
		{name: "the résumé", path: "/", status: http.StatusOK, location: ""},
		{name: "a page", path: "/p/tidy-tool/", status: http.StatusMovedPermanently, location: "/p/tidy-tool"},
		{name: "a tab", path: "/experience/", status: http.StatusMovedPermanently, location: "/experience"},
		{name: "the blog's list", path: "/blog/", status: http.StatusMovedPermanently, location: "/blog"},
		{name: "the list in a language", path: "/ru/blog/", status: http.StatusMovedPermanently, location: "/ru/blog"},
		{
			name: "the list, its query kept as written", path: "/blog/?ref=feed&tag=reading", status: http.StatusMovedPermanently,
			location: "/blog?ref=feed&tag=reading",
		},
		{name: "a page's markdown", path: "/index.md/", status: http.StatusMovedPermanently, location: "/index.md"},
		{
			name: "an older version, drawn when first asked for", path: "/blog/reading-slowly/2026-11-10/", status: http.StatusMovedPermanently,
			location: "/blog/reading-slowly/2026-11-10",
		},
		{
			name: "an older version's markdown, drawn when first asked for", path: "/blog/reading-slowly/2026-11-10.md/",
			status: http.StatusMovedPermanently, location: "/blog/reading-slowly/2026-11-10.md",
		},
		{name: "a picture of the site", path: "/favicon.svg/", status: http.StatusMovedPermanently, location: "/favicon.svg"},
		{name: "a post's number, straight to the post", path: "/blog/1/", status: http.StatusMovedPermanently, location: "/blog/reading-slowly"},
		{
			name: "the current version by its day, straight to the post for a while", path: "/blog/reading-slowly/2026-11-12/",
			status: http.StatusFound, location: "/blog/reading-slowly",
		},
		{name: "no page", path: "/nothing/", status: http.StatusNotFound, location: ""},
		{name: "a language's folder", path: "/ru/", status: http.StatusNotFound, location: ""},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := serve(server, http.MethodGet, testCase.path, "", "")
			header := response.Header()
			if response.Code != testCase.status || header.Get("Location") != testCase.location || header.Get("Cache-Control") != "no-cache" {
				test.Errorf("GET %s = %d, Location %q, Cache-Control %q; want %d, Location %q, Cache-Control %q", testCase.path,
					response.Code, header.Get("Location"), header.Get("Cache-Control"), testCase.status, testCase.location, "no-cache")
			}
		})
	}
}

// An address of no page is not found, and says so with the status it is answered with and the address asked for, its
// path as the request decoded it and its query as written: its format chosen as a page's is, in text where the manual
// or markdown would be, in HTML where HTML would, in the language whose folder the path opens with, the site's own
// where it opens with none the site is translated into; and a cache keeps one copy for each Accept and User-Agent.
func TestAnAddressOfNoPageIsNotFound(test *testing.T) {
	server := serverOf(fixtureStore(test, fixtures.ContentFolder(test)), false)
	const (
		plain   = "text/plain; charset=utf-8"
		browsed = "text/html; charset=utf-8"
	)
	cases := []struct {
		name        string
		path        string
		accept      string
		userAgent   string
		contentType string
		holds       string
	}{
		{name: "a project with no page", path: "/p/other-tool.md", accept: "text/plain", userAgent: "", contentType: plain, holds: `404 Not Found: "/p/other-tool.md"`},
		{name: "no page at all", path: "/nothing", accept: "text/plain", userAgent: "", contentType: plain, holds: `404 Not Found: "/nothing"`},
		{name: "a path written escaped", path: "/no%20thing", accept: "text/plain", userAgent: "", contentType: plain, holds: `404 Not Found: "/no thing"`},
		{
			name: "a post's own text beside its picture", path: "/blog/reading-slowly/post.en.md", accept: "text/plain", userAgent: "",
			contentType: plain, holds: `404 Not Found: "/blog/reading-slowly/post.en.md"`,
		},
		{
			name: "a path out of a post's folder, its slashes escaped", path: "/blog/reading-slowly/..%2F..%2Fprofile.yaml", accept: "text/plain",
			userAgent: "", contentType: plain, holds: `404 Not Found: "/blog/reading-slowly/../../profile.yaml"`,
		},
		{
			name: "the list filtered by a tag no post carries", path: "/blog?tag=nothing", accept: "text/plain", userAgent: "", contentType: plain,
			holds: `404 Not Found: "/blog?tag=nothing"`,
		},
		{name: "the list filtered by no tag", path: "/blog.md?tag=", accept: "text/plain", userAgent: "", contentType: plain, holds: `404 Not Found: "/blog.md?tag="`},
		{
			name: "the list filtered by a tag twice", path: "/blog?tag=reading&tag=terminals", accept: "text/plain", userAgent: "", contentType: plain,
			holds: `404 Not Found: "/blog?tag=reading&tag=terminals"`,
		},
		{name: "no page, with a query", path: "/nothing?ref=feed", accept: "text/plain", userAgent: "", contentType: plain, holds: `404 Not Found: "/nothing?ref=feed"`},
		{name: "anything, from an agent no list names", path: "/nothing", accept: "*/*", userAgent: unknownAgent, contentType: browsed, holds: `<code>&#34;/nothing&#34;</code>`},
		{name: "nothing asked, from no agent", path: "/nothing", accept: "", userAgent: "", contentType: browsed, holds: `<code>&#34;/nothing&#34;</code>`},
		{name: "curl, taking anything", path: "/nothing", accept: "*/*", userAgent: curlAgent, contentType: plain, holds: `404 Not Found: "/nothing"`},
		{name: "curl, asking for nothing", path: "/nothing", accept: "", userAgent: curlAgent, contentType: plain, holds: `404 Not Found: "/nothing"`},
		{name: "an agent asking for markdown", path: "/nothing", accept: "text/markdown", userAgent: "", contentType: plain, holds: `404 Not Found: "/nothing"`},
		{name: "a browser", path: "/nothing", accept: browserAccept, userAgent: "", contentType: browsed, holds: `<code>&#34;/nothing&#34;</code>`},
		{name: "a browser, in the site's language", path: "/nothing", accept: browserAccept, userAgent: "", contentType: browsed, holds: `<html lang="en">`},
		{name: "a browser, in a translation's folder", path: "/ru/blog/nothing", accept: browserAccept, userAgent: "", contentType: browsed, holds: `<html lang="ru">`},
		{
			name: "a browser, in the folder of a language the site has no words for", path: "/de/blog", accept: browserAccept, userAgent: "",
			contentType: browsed, holds: `<html lang="en">`,
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := serve(server, http.MethodGet, testCase.path, testCase.accept, testCase.userAgent)
			header := response.Header()
			if response.Code != http.StatusNotFound || header.Get("Content-Type") != testCase.contentType ||
				!strings.Contains(response.Body.String(), testCase.holds) || header.Get("Vary") != negotiatedVary {
				test.Errorf("GET %s, Accept %q, User-Agent %q = %d %q, Vary %q, %q; want %d %q, Vary %q, holding %q", testCase.path,
					testCase.accept, testCase.userAgent, response.Code, header.Get("Content-Type"), header.Get("Vary"), response.Body.String(),
					http.StatusNotFound, testCase.contentType, negotiatedVary, testCase.holds)
			}
		})
	}
}

// Every answer tells a browser not to guess its type, to load nothing and send no form but from and to the site,
// and to let no site frame it: a page in each format, the page of no page, and a refusal net/http writes itself.
func TestEveryAnswerTellsBrowsersWhatTheyMayDo(test *testing.T) {
	server := serverOf(fixtureStore(test, fixtures.ContentFolder(test)), false)
	const policy = "default-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'"
	cases := []struct {
		name   string
		method string
		path   string
		accept string
	}{
		{name: "a page in a browser", method: http.MethodGet, path: "/", accept: "text/html"},
		{name: "a page as a manual", method: http.MethodGet, path: "/", accept: "text/plain"},
		{name: "a page in markdown at its own path", method: http.MethodGet, path: "/index.md", accept: ""},
		{name: "no page in a browser", method: http.MethodGet, path: "/nothing", accept: "text/html"},
		{name: "no page as text", method: http.MethodGet, path: "/nothing", accept: "text/plain"},
		{name: "a method refused", method: http.MethodPost, path: "/", accept: ""},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			header := serve(server, testCase.method, testCase.path, testCase.accept, "").Header()
			if sentSniffing, sentPolicy := header.Get("X-Content-Type-Options"), header.Get("Content-Security-Policy"); sentSniffing != "nosniff" || sentPolicy != policy {
				test.Errorf("%s %s, Accept %q: X-Content-Type-Options %q, Content-Security-Policy %q; want %q, %q",
					testCase.method, testCase.path, testCase.accept, sentSniffing, sentPolicy, "nosniff", policy)
			}
		})
	}
}

// A page and a feed tell a cache to ask again each time, since a reload, or a post coming out, can change them.
func TestAPageTellsCachesToAskAgain(test *testing.T) {
	server := serverOf(fixtureStore(test, fixtures.ContentFolder(test)), false)
	for _, path := range []string{"/index.md", "/blog.atom"} {
		test.Run(path, func(test *testing.T) {
			if got := serve(server, http.MethodGet, path, "", "").Header().Get("Cache-Control"); got != "no-cache" {
				test.Errorf("GET %s: Cache-Control = %q, want %q", path, got, "no-cache")
			}
		})
	}
}

// A file of the build is served at its path as the type its name says, whatever the request asks for, and a cache
// keeps it for good: a file changed is built under another name.
func TestABuiltFileIsServedForCachesToKeepForGood(test *testing.T) {
	response := serve(serverOf(fixtureStore(test, fixtures.ContentFolder(test)), false), http.MethodGet, stylesheetPath, "*/*", "")
	header := response.Header()
	const (
		wantType  = "text/css; charset=utf-8"
		wantCache = "public, max-age=31536000, immutable"
	)
	if response.Code != http.StatusOK || header.Get("Content-Type") != wantType || header.Get("Cache-Control") != wantCache ||
		response.Body.String() != "body{margin:0}\n" {
		test.Errorf("GET %s = %d %q, Cache-Control %q, %q; want %d %q, %q, the fixture's stylesheet", stylesheetPath,
			response.Code, header.Get("Content-Type"), header.Get("Cache-Control"), response.Body.String(), http.StatusOK,
			wantType, wantCache)
	}
}

// A card is served at its address as a PNG, whatever the request asks for, for a cache to keep for good: a card changed
// is drawn under another name.
func TestACardIsServedForCachesToKeepForGood(test *testing.T) {
	response := serve(serverOf(fixtureStore(test, fixtures.ContentFolder(test)), false), http.MethodGet, cardPath, "text/plain", curlAgent)
	header := response.Header()
	const (
		wantType  = "image/png"
		wantCache = "public, max-age=31536000, immutable"
		wantBody  = "PNG page fixture\n"
	)
	if response.Code != http.StatusOK || header.Get("Content-Type") != wantType || header.Get("Cache-Control") != wantCache ||
		response.Body.String() != wantBody {
		test.Errorf("GET %s = %d %q, Cache-Control %q, %q; want %d %q, %q, %q", cardPath, response.Code, header.Get("Content-Type"),
			header.Get("Cache-Control"), response.Body.String(), http.StatusOK, wantType, wantCache, wantBody)
	}
}

// Nothing of the cards' folder is served but the cards: neither its manifest nor a file it names not.
func TestNothingOfTheCardsFolderIsServedButTheCards(test *testing.T) {
	server := serverOf(fixtureStore(test, fixtures.ContentFolder(test)), false)
	for _, path := range []string{"/cards/manifest.json", "/cards/stray-fixture.png"} {
		test.Run(path, func(test *testing.T) {
			if got := serve(server, http.MethodGet, path, "", "").Code; got != http.StatusNotFound {
				test.Errorf("GET %s = %d, want %d", path, got, http.StatusNotFound)
			}
		})
	}
}

// A page a browser reads, and the page of no page, are drawn in the build's stylesheet and run its script.
func TestABrowsedPageIsDrawnInTheStylesheetAndRunsTheScript(test *testing.T) {
	server := serverOf(fixtureStore(test, fixtures.ContentFolder(test)), false)
	wanted := []string{`<link rel="stylesheet" href="` + stylesheetPath + `">`, `<script type="module" src="` + scriptPath + `"`}
	cases := []struct {
		name string
		path string
	}{
		{name: "a page", path: "/"},
		{name: "no page", path: "/nothing"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			body := serve(server, http.MethodGet, testCase.path, "text/html", "").Body.String()
			for _, want := range wanted {
				if !strings.Contains(body, want) {
					test.Errorf("GET %s as a browser = %q, want it to hold %q", testCase.path, body, want)
				}
			}
		})
	}
}

// A browser finds the site's icon: each picture of it a page's head links, among them the one a phone keeps on its home
// screen and the one a browser asks for by name on a page with no head, as the manual, is served at its path as its
// type, for a cache to ask for again, since a new build of the server may draw it anew under the same path.
func TestABrowserFindsTheSiteIcon(test *testing.T) {
	server := serverOf(fixtureStore(test, fixtures.ContentFolder(test)), false)
	page := serve(server, http.MethodGet, "/", "text/html", "").Body.String()
	cases := []struct {
		name        string
		path        string
		contentType string
	}{
		{name: "the drawing", path: "/favicon.svg", contentType: "image/svg+xml"},
		{name: "the home screen's", path: "/apple-touch-icon.png", contentType: "image/png"},
		{name: "the one asked for by name", path: "/favicon.ico", contentType: "image/vnd.microsoft.icon"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := serve(server, http.MethodGet, testCase.path, "image/*", "")
			header := response.Header()
			linked := `href="` + testCase.path + `"`
			const wantCache = "no-cache"
			if response.Code != http.StatusOK || header.Get("Content-Type") != testCase.contentType ||
				header.Get("Cache-Control") != wantCache || response.Body.Len() == 0 || !strings.Contains(page, linked) {
				test.Errorf("GET %s = %d %q, Cache-Control %q, of %d bytes, the page linking it %t; want %d %q, %q, "+
					"a picture, linked by %q", testCase.path, response.Code, header.Get("Content-Type"),
					header.Get("Cache-Control"), response.Body.Len(), strings.Contains(page, linked), http.StatusOK,
					testCase.contentType, wantCache, linked)
			}
		})
	}
}

// A post's picture is served at its file beside the post as its type, for a cache to ask for again, since a reload may
// change it under the same name; under a policy that loads and runs nothing in it, so a drawing opened alone runs no
// script.
func TestAPostsPictureIsServedBesideThePost(test *testing.T) {
	response := serve(serverOf(fixtureStore(test, fixtures.ContentFolder(test)), false), http.MethodGet, picturePath, "image/*", "")
	header := response.Header()
	const (
		wantType   = "image/png"
		wantCache  = "no-cache"
		wantPolicy = "default-src 'none'; style-src 'unsafe-inline'; sandbox"
	)
	picture := fixturePicture(test)
	if response.Code != http.StatusOK || header.Get("Content-Type") != wantType || header.Get("Cache-Control") != wantCache ||
		header.Get("X-Content-Type-Options") != "nosniff" || header.Get("Content-Security-Policy") != wantPolicy ||
		!bytes.Equal(response.Body.Bytes(), picture) {
		test.Errorf("GET %s = %d %q, Cache-Control %q, X-Content-Type-Options %q, Content-Security-Policy %q, of %d bytes; "+
			"want %d %q, %q, %q, %q, the %d bytes of the fixture's picture", picturePath, response.Code, header.Get("Content-Type"),
			header.Get("Cache-Control"), header.Get("X-Content-Type-Options"), header.Get("Content-Security-Policy"),
			response.Body.Len(), http.StatusOK, wantType, wantCache, "nosniff", wantPolicy, len(picture))
	}
}

// A project's clip, its recording and its still, is served at its files beside the project's page, each as its type,
// for a cache to ask for again and under the policy that runs nothing in it, as a post's picture is.
func TestAProjectsClipIsServedBesideItsPage(test *testing.T) {
	server := serverOf(fixtureStore(test, fixtures.ContentFolder(test)), false)
	const (
		wantCache  = "no-cache"
		wantPolicy = "default-src 'none'; style-src 'unsafe-inline'; sandbox"
	)
	cases := []struct {
		name        string
		path        string
		file        string
		contentType string
	}{
		{name: "the recording", path: "/p/tidy-tool/tidying.mp4", file: "projects/tidy-tool/tidying.mp4", contentType: "video/mp4"},
		{name: "its still", path: "/p/tidy-tool/tidying.png", file: "projects/tidy-tool/tidying.png", contentType: "image/png"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			want, err := os.ReadFile(filepath.Join(fixtures.ContentFolder(test), testCase.file))
			if err != nil {
				test.Fatalf("read %q: %v", testCase.file, err)
			}
			response := serve(server, http.MethodGet, testCase.path, "*/*", "")
			header := response.Header()
			if response.Code != http.StatusOK || header.Get("Content-Type") != testCase.contentType || header.Get("Cache-Control") != wantCache ||
				header.Get("Content-Security-Policy") != wantPolicy || !bytes.Equal(response.Body.Bytes(), want) {
				test.Errorf("GET %s = %d %q, Cache-Control %q, Content-Security-Policy %q, of %d bytes; want %d %q, %q, %q, the %d bytes of %q",
					testCase.path, response.Code, header.Get("Content-Type"), header.Get("Cache-Control"), header.Get("Content-Security-Policy"),
					response.Body.Len(), http.StatusOK, testCase.contentType, wantCache, wantPolicy, len(want), testCase.file)
			}
		})
	}
}

// The key a site tells IndexNow of, and the path its file is served at, from the site's root (indexnow.org); and the
// path of the file of an empty key
const (
	indexNowKey      = "fixture-key-3f9a"
	indexNowKeyPath  = "/fixture-key-3f9a.txt"
	emptyKeyFilePath = "/.txt"
)

// seenKeyFile is the key's file as a test reads it: its status, its type, how long a cache keeps it, and whether its
// body is the key, all it holds.
type seenKeyFile struct {
	status       int
	contentType  string
	cacheControl string
	isKey        bool
}

// Where a site is served with an IndexNow key, its file at the site's root holds the key as plain text, whatever the
// request asks, for a cache to ask for again; where it is served with none, there is no such file, nor one of an empty
// key.
func TestTheIndexNowKeyIsServedWhereTheSiteHasOne(test *testing.T) {
	missing := seenKeyFile{status: http.StatusNotFound, contentType: "text/html; charset=utf-8", cacheControl: "no-cache", isKey: false}
	cases := []struct {
		name string
		key  string
		path string
		want seenKeyFile
	}{
		{
			name: "a key", key: indexNowKey, path: indexNowKeyPath,
			want: seenKeyFile{status: http.StatusOK, contentType: "text/plain; charset=utf-8", cacheControl: "no-cache", isKey: true},
		},
		{name: "none", key: "", path: indexNowKeyPath, want: missing},
		{name: "none, at an empty key's path", key: "", path: emptyKeyFilePath, want: missing},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			store := fixtureStore(test, fixtures.ContentFolder(test))
			server := New(store, release.DeriveFromContent(store, compose.Prepare), false, testCase.key)
			response := serve(server, http.MethodGet, testCase.path, "text/html", "")
			header := response.Header()
			got := seenKeyFile{
				status: response.Code, contentType: header.Get("Content-Type"), cacheControl: header.Get("Cache-Control"),
				isKey: response.Body.String() == indexNowKey,
			}
			if got != testCase.want {
				test.Errorf("GET %s with the key %q = %+v, want %+v", testCase.path, testCase.key, got, testCase.want)
			}
		})
	}
}

// What the store holds is drawn again once it holds new content: a reload reaches the next request, for a page and for
// a post's picture alike.
func TestAReloadReachesTheNextRequest(test *testing.T) {
	newPicture := string(pngOfTwoByOne(test))
	cases := []struct {
		name        string
		file        string
		path        string
		original    string
		replacement string
		want        string
	}{
		{name: "a page", file: "profile.yaml", path: "/index.md", original: "name: Mira Example", replacement: "name: Mira Other", want: "# Mira Other"},
		{name: "a post's picture", file: pictureFile, path: picturePath, original: string(fixturePicture(test)), replacement: newPicture, want: newPicture},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			folder := fixtures.Copy(test, fixtures.ContentFolder(test))
			store := fixtureStore(test, folder)
			server := serverOf(store, false)
			serve(server, http.MethodGet, testCase.path, "", "")
			fixtures.ReplaceIn(test, folder, testCase.file, testCase.original, testCase.replacement)
			if err := store.Reload(); err != nil {
				test.Fatalf("Reload = %v, want no error", err)
			}
			if body := serve(server, http.MethodGet, testCase.path, "", "").Body.String(); !strings.Contains(body, testCase.want) {
				test.Errorf("GET %s after a reload = %q, want it to hold %q", testCase.path, body, testCase.want)
			}
		})
	}
}

// A post dated an hour after a synctest bubble's clock starts, which every request in the bubble reads the site at,
// and its file
const (
	comingPostFile = "posts/coming/post.en.md"
	comingPostText = "---\nid: 3\ntitle: Coming\ndescription: About what comes.\nstatus: published\npublished: 2000-01-01T01:00+00:00\n---\nIts text.\n"
	comingIn       = time.Hour
)

// seenAnswer is an answer as a test reads it: its status, and whether it holds a text.
type seenAnswer struct {
	status int
	holds  bool
}

// A post dated later than the request is a draft until its moment: served at its address to its author alone, who reads
// the drafts, and not at its number, nor in the blog's list or llms.txt; from its moment on it is, with no reload.
func TestAPostComesOutAtItsMoment(test *testing.T) {
	cases := []struct {
		name   string
		path   string
		drafts bool
		holds  string
		before seenAnswer
		after  seenAnswer
	}{
		{name: "its page", path: "/blog/coming.md", drafts: false, holds: "# Coming", before: seenAnswer{status: http.StatusNotFound, holds: false}, after: seenAnswer{status: http.StatusOK, holds: true}},
		{name: "its page, to its author", path: "/blog/coming.md", drafts: true, holds: "# Coming", before: seenAnswer{status: http.StatusOK, holds: true}, after: seenAnswer{status: http.StatusOK, holds: true}},
		{name: "its number", path: "/blog/3", drafts: false, holds: "/blog/coming", before: seenAnswer{status: http.StatusNotFound, holds: false}, after: seenAnswer{status: http.StatusMovedPermanently, holds: true}},
		{name: "the blog's list", path: "/blog.md", drafts: false, holds: "Coming", before: seenAnswer{status: http.StatusOK, holds: false}, after: seenAnswer{status: http.StatusOK, holds: true}},
		{name: "the blog's list, to its author", path: "/blog.md", drafts: true, holds: "Coming", before: seenAnswer{status: http.StatusOK, holds: false}, after: seenAnswer{status: http.StatusOK, holds: true}},
		{name: "the index", path: "/llms.txt", drafts: false, holds: "Coming", before: seenAnswer{status: http.StatusOK, holds: false}, after: seenAnswer{status: http.StatusOK, holds: true}},
		// the bubble's clock starts in 2000, before the fixture's post: the blog lists none until this one comes out
		{name: "the blog's feed", path: "/blog.atom", drafts: false, holds: "Coming", before: seenAnswer{status: http.StatusNotFound, holds: false}, after: seenAnswer{status: http.StatusOK, holds: true}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			synctest.Test(test, func(test *testing.T) {
				folder := fixtures.Copy(test, fixtures.ContentFolder(test))
				fixtures.WriteIn(test, folder, comingPostFile, comingPostText)
				server := serverOf(fixtureStore(test, folder), testCase.drafts)
				before := seenAnswerOf(serve(server, http.MethodGet, testCase.path, "", ""), testCase.holds)
				time.Sleep(comingIn)
				after := seenAnswerOf(serve(server, http.MethodGet, testCase.path, "", ""), testCase.holds)
				if before != testCase.before || after != testCase.after {
					test.Errorf("GET %s before the post's moment and at it = %+v and %+v, want %+v and %+v", testCase.path, before, after,
						testCase.before, testCase.after)
				}
			})
		})
	}
}

// seenAnswerOf is an answer as a test reads it, whether it holds a text in its body or its location.
func seenAnswerOf(response *httptest.ResponseRecorder, text string) seenAnswer {
	held := response.Body.String() + response.Header().Get("Location")
	return seenAnswer{status: response.Code, holds: strings.Contains(held, text)}
}

// A draft is served at its address only on a server its author runs to read the drafts.
func TestADraftIsServedToItsAuthorAlone(test *testing.T) {
	cases := []struct {
		name   string
		drafts bool
		want   seenAnswer
	}{
		{name: "to the public", drafts: false, want: seenAnswer{status: http.StatusNotFound, holds: false}},
		{name: "to its author", drafts: true, want: seenAnswer{status: http.StatusOK, holds: true}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			server := serverOf(fixtureStore(test, fixtures.ContentFolder(test)), testCase.drafts)
			if got := seenAnswerOf(serve(server, http.MethodGet, "/blog/notes-on-tools.md", "", ""), "# Notes on tools"); got != testCase.want {
				test.Errorf("GET /blog/notes-on-tools.md = %+v, want %+v", got, testCase.want)
			}
		})
	}
}

// A page is only read: a request to change it is refused.
func TestAPageIsOnlyRead(test *testing.T) {
	server := serverOf(fixtureStore(test, fixtures.ContentFolder(test)), false)
	if response := serve(server, http.MethodPost, "/index.md", "", ""); response.Code != http.StatusMethodNotAllowed {
		test.Errorf("POST /index.md = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

// serve is the server's answer to a request by a method at a path, with an Accept header and a User-Agent, either empty
// where the request sends none.
func serve(server http.Handler, method, path, accept, userAgent string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	request.Header.Set("Accept", accept)
	request.Header.Set("User-Agent", userAgent)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response
}

// fixturePicture is what the fixture post's picture holds.
func fixturePicture(test *testing.T) []byte {
	test.Helper()
	picture, err := os.ReadFile(filepath.Join(fixtures.ContentFolder(test), pictureFile))
	if err != nil {
		test.Fatalf("read the fixture's picture: %v", err)
	}
	return picture
}

// pngOfTwoByOne is a PNG two pixels by one.
func pngOfTwoByOne(test *testing.T) []byte {
	test.Helper()
	var written bytes.Buffer
	if err := png.Encode(&written, image.NewGray(image.Rect(0, 0, 2, 1))); err != nil {
		test.Fatalf("write a PNG: %v", err)
	}
	return written.Bytes()
}

// serverOf is a server of a store, its drafts read where drafts says, composed from each snapshot prepared as the
// composition root prepares it, with no IndexNow key.
func serverOf(store *release.Store, drafts bool) *Server {
	return New(store, release.DeriveFromContent(store, compose.Prepare), drafts, "")
}

// An unlisted post, out long before any test runs: its file in the content's folder, its text and the path it is
// served at
const (
	unlistedPostFile = "posts/aside/post.en.md"
	unlistedPostText = "---\nid: 3\ntitle: Aside\ndescription: A note for those sent it.\nstatus: unlisted\npublished: 2000-01-01T00:00+00:00\n---\nIts text.\n"
	unlistedPostPath = "/blog/aside"
)

// unlistedPostServer is a server of the fixture content with an unlisted post beside its own.
func unlistedPostServer(test *testing.T) *Server {
	folder := fixtures.Copy(test, fixtures.ContentFolder(test))
	fixtures.WriteIn(test, folder, unlistedPostFile, unlistedPostText)
	return serverOf(fixtureStore(test, folder), false)
}

// fixtureStore is a store of a release of the content in a folder, the build's fixture and the cards'.
func fixtureStore(test *testing.T, folder string) *release.Store {
	folders := release.Folders{Content: folder, Build: fixtures.BuildFolder(test), Cards: fixtures.CardsFolder(test)}
	checks := release.Checks{
		Content: func(loaded *content.Snapshot) error { return compose.Check(loaded, time.Now()) },
		// a post a test adds has no card in the cards' fixture, which no test here reads of that post
		Cards: func(*content.Snapshot, *assets.Cards) error { return nil },
	}
	store, err := release.Open(folders, checks)
	if err != nil {
		test.Fatalf("Open(fixtures) = %v, want no error", err)
	}
	return store
}
