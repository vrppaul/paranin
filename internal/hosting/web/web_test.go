package web

import (
	"bytes"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/release"
)

// The build's fixture's stylesheet and script, which every page a browser reads includes
const (
	stylesheetPath = "/assets/site-fixture.css"
	scriptPath     = "/assets/site-fixture.js"
)

// The fixture post's slug, and its picture: its file in the content's folder, and where the site serves it
const (
	fixturePostSlug = "reading-slowly"
	pictureFile     = "posts/reading-slowly/margin.png"
	picturePath     = "/blog/reading-slowly/margin.png"
)

// How long a test's preview link lasts from the tests' day, which its links are opened on
const linkLasts = 24 * time.Hour

// Each page is served at its address in each format, and the index at llms.txt as plain text, as agents read it.
func TestEveryPageIsServedAtItsAddress(test *testing.T) {
	server := New(fixtureStore(test, fixtures.ContentFolder(test)), false, nil)
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
		{name: "a post", path: "/blog/reading-slowly.md", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "# Reading slowly"},
		{name: "a post as a manual", path: "/blog/reading-slowly.txt", status: http.StatusOK, contentType: "text/plain; charset=utf-8", holds: "READING-SLOWLY(1)"},
		{name: "an older version", path: "/blog/reading-slowly/2026-11-10.md", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "# Slow reading"},
		{
			name: "an older version as a manual", path: "/blog/reading-slowly/2026-11-10.txt", status: http.StatusOK, contentType: "text/plain; charset=utf-8",
			holds: "newer: 2026-11-12",
		},
		{name: "an older version, chosen", path: "/blog/reading-slowly/2026-11-10", status: http.StatusOK, contentType: "text/plain; charset=utf-8", holds: "Slow reading"},
		{name: "a post's picture beside its older version", path: "/blog/reading-slowly/margin.png", status: http.StatusOK, contentType: "image/png", holds: ""},
		{name: "the blog's list", path: "/blog.md", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "## 2026"},
		{name: "the blog's list as a manual", path: "/blog.txt", status: http.StatusOK, contentType: "text/plain; charset=utf-8", holds: "BLOG(1)"},
		{name: "the list by series", path: "/blog/series.md", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "1 of 2 out"},
		{name: "the list by series as a manual", path: "/blog/series.txt", status: http.StatusOK, contentType: "text/plain; charset=utf-8", holds: "BLOG-SERIES(1)"},
		{name: "the list filtered by a tag", path: "/blog.md?tag=terminals", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "posts tagged terminals"},
		{
			name: "the list filtered by a series, as a manual", path: "/blog.txt?series=reading", status: http.StatusOK, contentType: "text/plain; charset=utf-8",
			holds: "blog — parts of Reading",
		},
		{name: "the list filtered, chosen", path: "/blog?tag=reading", status: http.StatusOK, contentType: "text/plain; charset=utf-8", holds: "blog — posts tagged reading"},
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
		{name: "the feed of no series", path: "/blog/series/aloud.atom", status: http.StatusNotFound, contentType: "text/plain; charset=utf-8", holds: ""},
		{name: "a translation", path: "/ru/blog/reading-slowly.md", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "# Медленное чтение"},
		{name: "a translation as a manual", path: "/ru/blog/reading-slowly.txt", status: http.StatusOK, contentType: "text/plain; charset=utf-8", holds: "Медленное чтение"},
		{name: "a translation's picture", path: "/ru/blog/reading-slowly/margin.png", status: http.StatusOK, contentType: "image/png", holds: ""},
		{name: "the list in a language", path: "/ru/blog.md", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "# блог"},
		{name: "the list in a language, filtered", path: "/ru/blog.md?tag=reading", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "посты с тегом reading"},
		{
			name: "the blog's feed in a language", path: "/ru/blog.atom", status: http.StatusOK, contentType: "application/atom+xml; charset=utf-8",
			holds: `<id>tag:example.test,2026:ru/blog/1</id>`,
		},
		{name: "a post in a language it has no translation in", path: "/ru/blog/notes-on-tools.md", status: http.StatusNotFound, contentType: "text/plain; charset=utf-8", holds: ""},
		{name: "the list in a language the site has no words for", path: "/de/blog.md", status: http.StatusNotFound, contentType: "text/plain; charset=utf-8", holds: ""},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := serve(server, http.MethodGet, testCase.path, "")
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
	server := New(fixtureStore(test, fixtures.ContentFolder(test)), false, nil)
	const (
		markdown = "text/markdown; charset=utf-8"
		manual   = "text/plain; charset=utf-8"
		browsed  = "text/html; charset=utf-8"
		// a line of the résumé's opening tab, and one of a project's opening tab
		resumeOpening  = "Writes software for people who read it."
		projectOpening = "go install example.test/tidy@latest"
	)
	cases := []struct {
		name        string
		path        string
		accept      string
		contentType string
		holds       string
		lacks       string
	}{
		{name: "a tab of the résumé in markdown", path: "/experience.md", accept: "", contentType: markdown, holds: "# Experience\n\n## Northwind — Senior Engineer", lacks: resumeOpening},
		{name: "a tab of the résumé as a manual", path: "/experience.txt", accept: "", contentType: manual, holds: "Northwind", lacks: resumeOpening},
		{name: "a tab of the résumé, chosen by curl", path: "/experience", accept: "*/*", contentType: manual, holds: "Northwind", lacks: resumeOpening},
		{name: "a tab of the résumé in a browser", path: "/experience", accept: "text/html", contentType: browsed, holds: "Northwind", lacks: resumeOpening},
		{name: "a tab of a project's page in markdown", path: "/p/tidy-tool/why.md", accept: "", contentType: markdown, holds: "# Why\n\nFiles were untidy.", lacks: projectOpening},
		{name: "a tab of a project's page, chosen by curl", path: "/p/tidy-tool/why", accept: "*/*", contentType: manual, holds: "Files were untidy.", lacks: projectOpening},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := serve(server, http.MethodGet, testCase.path, testCase.accept)
			body := response.Body.String()
			if response.Code != http.StatusOK || response.Header().Get("Content-Type") != testCase.contentType ||
				!strings.Contains(body, testCase.holds) || strings.Contains(body, testCase.lacks) {
				test.Errorf("GET %s, Accept %q = %d %q %q; want %d %q holding %q and not %q", testCase.path, testCase.accept,
					response.Code, response.Header().Get("Content-Type"), body, http.StatusOK, testCase.contentType, testCase.holds, testCase.lacks)
			}
		})
	}
}

// At a page's own address the reader's request chooses the format: HTML for a browser, markdown for whoever asks
// for it, the manual for everyone else, `curl` among them, and for a request that ranks two of them alike; and a
// cache keeps one copy for each Accept.
func TestTheReaderChoosesTheFormat(test *testing.T) {
	server := New(fixtureStore(test, fixtures.ContentFolder(test)), false, nil)
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
		contentType string
		holds       string
	}{
		{name: "an agent asking for markdown", path: "/", accept: "text/markdown", contentType: markdown, holds: "# Mira Example"},
		{name: "curl, taking anything", path: "/", accept: "*/*", contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "a request asking for nothing", path: "/", accept: "", contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "markdown refused", path: "/", accept: "text/markdown;q=0, */*", contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "markdown among others", path: "/", accept: "text/html;q=0.9, text/markdown", contentType: markdown, holds: "# Mira Example"},
		{name: "plain text ranked above markdown", path: "/", accept: "text/plain, text/markdown;q=0.1", contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "markdown ranked above plain text", path: "/", accept: "text/markdown, text/plain;q=0.5", contentType: markdown, holds: "# Mira Example"},
		{name: "all of a family", path: "/", accept: "text/*", contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "markdown below anything", path: "/", accept: "text/markdown;q=0.5, */*", contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "markdown below its family", path: "/", accept: "text/markdown;q=0.5, text/*", contentType: manual, holds: "EXAMPLE.TEST(1)"},
		{name: "a browser asking for HTML", path: "/", accept: "text/html", contentType: browsed, holds: pageOpening},
		{
			name: "a browser's own Accept", path: "/", accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			contentType: browsed, holds: pageOpening,
		},
		{name: "HTML and markdown alike", path: "/", accept: "text/html, text/markdown, text/plain;q=0.1", contentType: markdown, holds: "# Mira Example"},
		{name: "a project's page in markdown", path: "/p/tidy-tool", accept: "text/markdown", contentType: markdown, holds: "# tidy-tool"},
		{name: "a project's page as a manual", path: "/p/tidy-tool", accept: "*/*", contentType: manual, holds: "TIDY-TOOL(1)"},
		{name: "a project's page in a browser", path: "/p/tidy-tool", accept: "text/html", contentType: browsed, holds: `<h1 aria-label="tidy-tool">`},
		{name: "a post in markdown", path: "/blog/reading-slowly", accept: "text/markdown", contentType: markdown, holds: "# Reading slowly"},
		{name: "a post as a manual", path: "/blog/reading-slowly", accept: "*/*", contentType: manual, holds: "READING-SLOWLY(1)"},
		{
			name: "a post in a browser", path: "/blog/reading-slowly", accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			contentType: browsed, holds: `<h1 class="post-title">`,
		},
		{name: "the blog's list in markdown", path: "/blog", accept: "text/markdown", contentType: markdown, holds: "# blog"},
		{name: "the blog's list in a browser", path: "/blog", accept: "text/html", contentType: browsed, holds: `<a class="switch-way" href="/blog" aria-current="page">`},
		{
			name: "the list by series in a browser", path: "/blog/series", accept: "text/html", contentType: browsed,
			holds: `<a class="switch-way" href="/blog/series" aria-current="page">`,
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := serve(server, http.MethodGet, testCase.path, testCase.accept)
			header := response.Header()
			if header.Get("Content-Type") != testCase.contentType || !strings.Contains(response.Body.String(), testCase.holds) ||
				header.Get("Vary") != "Accept" {
				test.Errorf("GET %s, Accept %q = %q %q, Vary %q; want %q holding %q, Vary %q", testCase.path, testCase.accept,
					header.Get("Content-Type"), response.Body.String(), header.Get("Vary"), testCase.contentType, testCase.holds, "Accept")
			}
		})
	}
}

// Each page a browser reads carries its record as linked data: the résumé its person, a project's page its code, a
// tab its page's, the person and the code at their page's address whole.
func TestEveryPageCarriesItsRecord(test *testing.T) {
	server := New(fixtureStore(test, fixtures.ContentFolder(test)), false, nil)
	cases := []struct {
		name  string
		path  string
		holds string
	}{
		{name: "the résumé", path: "/", holds: `"@type":"ProfilePage"`},
		{name: "a project's page", path: "/p/tidy-tool", holds: `"@type":"SoftwareSourceCode","name":"tidy-tool"`},
		{name: "a tab of the résumé, its page's", path: "/experience", holds: `"email":"mira@example.test","url":"https://example.test/","worksFor"`},
		{name: "a tab of a project's page, its page's", path: "/p/tidy-tool/why", holds: `"programmingLanguage":"Lua","url":"https://example.test/p/tidy-tool","codeRepository"`},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := serve(server, http.MethodGet, testCase.path, "text/html")
			if !strings.Contains(response.Body.String(), testCase.holds) {
				test.Errorf("GET %s as a browser = %q, want it to hold %q", testCase.path, response.Body.String(), testCase.holds)
			}
		})
	}
}

// Each format of a page names the address the page is known by, whole, so a search engine counts its markdown and
// its manual as that page; the index, which is no page a person reads, names none, nor a filtered list, a part of the
// list and no copy of it.
func TestEveryFormatOfAPageNamesItsCanonicalAddress(test *testing.T) {
	server := New(fixtureStore(test, fixtures.ContentFolder(test)), false, nil)
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
		{name: "a tab of a project's page, chosen", path: "/p/tidy-tool/why", accept: "text/html", want: `<https://example.test/p/tidy-tool/why>; rel="canonical"`},
		{name: "a post in markdown", path: "/blog/reading-slowly.md", accept: "", want: `<https://example.test/blog/reading-slowly>; rel="canonical"`},
		{name: "a post, chosen", path: "/blog/reading-slowly", accept: "*/*", want: `<https://example.test/blog/reading-slowly>; rel="canonical"`},
		{name: "a post in a browser", path: "/blog/reading-slowly", accept: "text/html", want: `<https://example.test/blog/reading-slowly>; rel="canonical"`},
		{name: "a translation in markdown", path: "/ru/blog/reading-slowly.md", accept: "", want: `<https://example.test/ru/blog/reading-slowly>; rel="canonical"`},
		{name: "a filtered list in markdown, a part of the list and no copy of it", path: "/blog.md?tag=reading", accept: "", want: ""},
		{name: "a filtered list as a manual", path: "/blog.txt?series=reading", accept: "", want: ""},
		{name: "a filtered list in a browser", path: "/blog?tag=terminals", accept: "text/html", want: ""},
		{name: "the index", path: "/llms.txt", accept: "", want: ""},
		{name: "no page", path: "/nothing", accept: "", want: ""},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := serve(server, http.MethodGet, testCase.path, testCase.accept)
			if got := response.Header().Get("Link"); got != testCase.want {
				test.Errorf("GET %s, Accept %q: Link = %q, want %q", testCase.path, testCase.accept, got, testCase.want)
			}
		})
	}
}

// A filtered list, a view of the list and no page of its own, and an older version of a post, which a newer stands
// over, tell a search engine not to index them, in each format; the list itself, and any other page, leaves it free to.
func TestAPageOfNoIndexIsSentForNoSearchEngineToIndex(test *testing.T) {
	server := New(fixtureStore(test, fixtures.ContentFolder(test)), false, nil)
	cases := []struct {
		name   string
		path   string
		accept string
		want   string
	}{
		{name: "filtered by a tag, in a browser", path: "/blog?tag=reading", accept: "text/html", want: "noindex"},
		{name: "filtered by a series, chosen", path: "/blog?series=reading", accept: "*/*", want: "noindex"},
		{name: "filtered, in markdown", path: "/blog.md?tag=terminals", accept: "", want: "noindex"},
		{name: "filtered, as a manual", path: "/blog.txt?tag=terminals", accept: "", want: "noindex"},
		{name: "the list", path: "/blog", accept: "text/html", want: ""},
		{name: "the list in markdown", path: "/blog.md", accept: "", want: ""},
		{name: "a post", path: "/blog/reading-slowly", accept: "text/html", want: ""},
		{name: "an older version, in a browser", path: "/blog/reading-slowly/2026-11-10", accept: "text/html", want: "noindex"},
		{name: "an older version, in markdown", path: "/blog/reading-slowly/2026-11-10.md", accept: "", want: "noindex"},
		{name: "an older version, as a manual", path: "/blog/reading-slowly/2026-11-10.txt", accept: "", want: "noindex"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := serve(server, http.MethodGet, testCase.path, testCase.accept)
			if got := response.Header().Get("X-Robots-Tag"); response.Code != http.StatusOK || got != testCase.want {
				test.Errorf("GET %s, Accept %q = %d, X-Robots-Tag %q; want %d, %q", testCase.path, testCase.accept, response.Code, got, http.StatusOK, testCase.want)
			}
		})
	}
}

// Readers who first ask for a filtered list at the same moment, in each format, each get the whole list as one drawing
// of it, which -race holds to drawing it once without a race.
func TestAFilteredListAskedForAtOnceIsServedWhole(test *testing.T) {
	server := New(fixtureStore(test, fixtures.ContentFolder(test)), false, nil)
	formats := []struct {
		path   string
		accept string
		holds  string
	}{
		{path: "/blog?tag=reading", accept: "text/html", holds: "<title>blog — posts tagged reading</title>"},
		{path: "/blog?tag=reading", accept: "", holds: "blog — posts tagged reading"},
		{path: "/blog.md?tag=reading", accept: "", holds: "posts tagged reading"},
	}
	// how many readers ask for each format at the same moment
	const readersEach = 8
	answered := make(chan readersAnswer, readersEach*len(formats))
	var readers sync.WaitGroup
	for range readersEach {
		for _, format := range formats {
			readers.Go(func() {
				answered <- readersAnswer{holds: format.holds, body: serve(server, http.MethodGet, format.path, format.accept).Body.String()}
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
	server := New(fixtureStore(test, fixtures.ContentFolder(test)), false, nil)
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
			response := serve(server, http.MethodGet, testCase.path, "*/*")
			header := response.Header()
			if response.Code != testCase.status || header.Get("Location") != testCase.location || header.Get("Cache-Control") != "no-cache" {
				test.Errorf("GET %s = %d, Location %q, Cache-Control %q; want %d, Location %q, Cache-Control %q", testCase.path,
					response.Code, header.Get("Location"), header.Get("Cache-Control"), testCase.status, testCase.location, "no-cache")
			}
		})
	}
}

// An address of no page is not found, and says so with the status it is answered with and the address asked for, its
// path as the request decoded it and its query as written: in plain text to a terminal, an agent and a request naming
// nothing, in HTML, in the site's language, to a browser; and a cache keeps one copy for each Accept.
func TestAnAddressOfNoPageIsNotFound(test *testing.T) {
	server := New(fixtureStore(test, fixtures.ContentFolder(test)), false, nil)
	const (
		plain   = "text/plain; charset=utf-8"
		browsed = "text/html; charset=utf-8"
		browser = "text/html,application/xhtml+xml,*/*;q=0.8"
	)
	cases := []struct {
		name        string
		path        string
		accept      string
		contentType string
		holds       string
	}{
		{name: "a project with no page", path: "/p/other-tool.md", accept: "", contentType: plain, holds: `404 Not Found: "/p/other-tool.md"`},
		{name: "no page at all", path: "/nothing", accept: "", contentType: plain, holds: `404 Not Found: "/nothing"`},
		{name: "a path written escaped", path: "/no%20thing", accept: "", contentType: plain, holds: `404 Not Found: "/no thing"`},
		{
			name: "a post's own text beside its picture", path: "/blog/reading-slowly/post.en.md", accept: "",
			contentType: plain, holds: `404 Not Found: "/blog/reading-slowly/post.en.md"`,
		},
		{
			name: "a path out of a post's folder, its slashes escaped", path: "/blog/reading-slowly/..%2F..%2Fprofile.yaml", accept: "",
			contentType: plain, holds: `404 Not Found: "/blog/reading-slowly/../../profile.yaml"`,
		},
		{name: "the list filtered by a tag no post carries", path: "/blog?tag=nothing", accept: "", contentType: plain, holds: `404 Not Found: "/blog?tag=nothing"`},
		{name: "the list filtered by no tag", path: "/blog.md?tag=", accept: "", contentType: plain, holds: `404 Not Found: "/blog.md?tag="`},
		{
			name: "the list filtered by a tag twice", path: "/blog?tag=reading&tag=terminals", accept: "", contentType: plain,
			holds: `404 Not Found: "/blog?tag=reading&tag=terminals"`,
		},
		{
			name: "the list filtered by a tag and a series", path: "/blog?tag=reading&series=reading&ref=feed", accept: "", contentType: plain,
			holds: `404 Not Found: "/blog?tag=reading&series=reading&ref=feed"`,
		},
		{name: "no page, with a query", path: "/nothing?ref=feed", accept: "", contentType: plain, holds: `404 Not Found: "/nothing?ref=feed"`},
		{name: "curl, taking anything", path: "/nothing", accept: "*/*", contentType: plain, holds: `404 Not Found: "/nothing"`},
		{name: "an agent asking for markdown", path: "/nothing", accept: "text/markdown", contentType: plain, holds: `404 Not Found: "/nothing"`},
		{name: "a browser", path: "/nothing", accept: browser, contentType: browsed, holds: `<code>&#34;/nothing&#34;</code>`},
		{name: "a browser, in the site's language", path: "/nothing", accept: browser, contentType: browsed, holds: `<html lang="en">`},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := serve(server, http.MethodGet, testCase.path, testCase.accept)
			header := response.Header()
			if response.Code != http.StatusNotFound || header.Get("Content-Type") != testCase.contentType ||
				!strings.Contains(response.Body.String(), testCase.holds) || header.Get("Vary") != "Accept" {
				test.Errorf("GET %s, Accept %q = %d %q, Vary %q, %q; want %d %q, Vary %q, holding %q", testCase.path,
					testCase.accept, response.Code, header.Get("Content-Type"), header.Get("Vary"), response.Body.String(),
					http.StatusNotFound, testCase.contentType, "Accept", testCase.holds)
			}
		})
	}
}

// Every answer tells a browser not to guess its type, to load nothing and send no form but from and to the site,
// and to let no site frame it: a page in each format, the page of no page, and a refusal net/http writes itself.
func TestEveryAnswerTellsBrowsersWhatTheyMayDo(test *testing.T) {
	server := New(fixtureStore(test, fixtures.ContentFolder(test)), false, nil)
	const policy = "default-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'"
	cases := []struct {
		name   string
		method string
		path   string
		accept string
	}{
		{name: "a page in a browser", method: http.MethodGet, path: "/", accept: "text/html"},
		{name: "a page as a manual", method: http.MethodGet, path: "/", accept: "*/*"},
		{name: "a page in markdown at its own path", method: http.MethodGet, path: "/index.md", accept: ""},
		{name: "no page in a browser", method: http.MethodGet, path: "/nothing", accept: "text/html"},
		{name: "no page as text", method: http.MethodGet, path: "/nothing", accept: ""},
		{name: "a method refused", method: http.MethodPost, path: "/", accept: ""},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			header := serve(server, testCase.method, testCase.path, testCase.accept).Header()
			if sentSniffing, sentPolicy := header.Get("X-Content-Type-Options"), header.Get("Content-Security-Policy"); sentSniffing != "nosniff" || sentPolicy != policy {
				test.Errorf("%s %s, Accept %q: X-Content-Type-Options %q, Content-Security-Policy %q; want %q, %q",
					testCase.method, testCase.path, testCase.accept, sentSniffing, sentPolicy, "nosniff", policy)
			}
		})
	}
}

// A page and a feed tell a cache to ask again each time, since a reload, or a post coming out, can change them.
func TestAPageTellsCachesToAskAgain(test *testing.T) {
	server := New(fixtureStore(test, fixtures.ContentFolder(test)), false, nil)
	for _, path := range []string{"/index.md", "/blog.atom"} {
		test.Run(path, func(test *testing.T) {
			if got := serve(server, http.MethodGet, path, "").Header().Get("Cache-Control"); got != "no-cache" {
				test.Errorf("GET %s: Cache-Control = %q, want %q", path, got, "no-cache")
			}
		})
	}
}

// A file of the build is served at its path as the type its name says, whatever the request asks for, and a cache
// keeps it for good: a file changed is built under another name.
func TestABuiltFileIsServedForCachesToKeepForGood(test *testing.T) {
	response := serve(New(fixtureStore(test, fixtures.ContentFolder(test)), false, nil), http.MethodGet, stylesheetPath, "*/*")
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

// A page a browser reads, and the page of no page, are drawn in the build's stylesheet and run its script.
func TestABrowsedPageIsDrawnInTheStylesheetAndRunsTheScript(test *testing.T) {
	server := New(fixtureStore(test, fixtures.ContentFolder(test)), false, nil)
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
			body := serve(server, http.MethodGet, testCase.path, "text/html").Body.String()
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
	server := New(fixtureStore(test, fixtures.ContentFolder(test)), false, nil)
	page := serve(server, http.MethodGet, "/", "text/html").Body.String()
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
			response := serve(server, http.MethodGet, testCase.path, "image/*")
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
	response := serve(New(fixtureStore(test, fixtures.ContentFolder(test)), false, nil), http.MethodGet, picturePath, "image/*")
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
			server := New(store, false, nil)
			serve(server, http.MethodGet, testCase.path, "")
			fixtures.ReplaceIn(test, folder, testCase.file, testCase.original, testCase.replacement)
			if err := store.Reload(); err != nil {
				test.Fatalf("Reload = %v, want no error", err)
			}
			if body := serve(server, http.MethodGet, testCase.path, "").Body.String(); !strings.Contains(body, testCase.want) {
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

// A post dated later than the request is served at its preview link alone until its moment: not at its address nor its
// number, nor in the blog's list or llms.txt; from its moment on it is, with no reload, and its link sends the reader
// on to it.
func TestAPostComesOutAtItsMoment(test *testing.T) {
	key := keyOf(test, testSecret)
	cases := []struct {
		name   string
		path   string
		holds  string
		before seenAnswer
		after  seenAnswer
	}{
		{name: "its page", path: "/blog/coming.md", holds: "# Coming", before: seenAnswer{status: http.StatusNotFound, holds: false}, after: seenAnswer{status: http.StatusOK, holds: true}},
		{
			name: "its preview link", path: "/preview/" + key.Token(document.PostPage("coming"), fixtures.Day().Add(linkLasts)) + "/coming.md", holds: "# Coming",
			before: seenAnswer{status: http.StatusOK, holds: true}, after: seenAnswer{status: http.StatusFound, holds: false},
		},
		{name: "its number", path: "/blog/3", holds: "/blog/coming", before: seenAnswer{status: http.StatusNotFound, holds: false}, after: seenAnswer{status: http.StatusMovedPermanently, holds: true}},
		{name: "the blog's list", path: "/blog.md", holds: "Coming", before: seenAnswer{status: http.StatusOK, holds: false}, after: seenAnswer{status: http.StatusOK, holds: true}},
		{name: "the index", path: "/llms.txt", holds: "Coming", before: seenAnswer{status: http.StatusOK, holds: false}, after: seenAnswer{status: http.StatusOK, holds: true}},
		// the bubble's clock starts in 2000, before the fixture's post: the blog lists none until this one comes out
		{name: "the blog's feed", path: "/blog.atom", holds: "Coming", before: seenAnswer{status: http.StatusNotFound, holds: false}, after: seenAnswer{status: http.StatusOK, holds: true}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			synctest.Test(test, func(test *testing.T) {
				folder := fixtures.Copy(test, fixtures.ContentFolder(test))
				fixtures.WriteIn(test, folder, comingPostFile, comingPostText)
				server := New(fixtureStore(test, folder), false, key)
				before := seenAnswerOf(serve(server, http.MethodGet, testCase.path, ""), testCase.holds)
				time.Sleep(comingIn)
				after := seenAnswerOf(serve(server, http.MethodGet, testCase.path, ""), testCase.holds)
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
			server := New(fixtureStore(test, fixtures.ContentFolder(test)), testCase.drafts, nil)
			if got := seenAnswerOf(serve(server, http.MethodGet, "/blog/notes-on-tools.md", ""), "# Notes on tools"); got != testCase.want {
				test.Errorf("GET /blog/notes-on-tools.md = %+v, want %+v", got, testCase.want)
			}
		})
	}
}

// The secrets the tests sign preview links with: the one the server holds, and one it held before
const (
	testSecret = "a secret the tests sign preview links with"
	oldSecret  = "a secret the tests signed preview links with before"
	// a token as a key writes one, lapsing in 2100, which no key the tests hold signs
	wrongToken = "4102444800.c2lnbmVkIGJ5IG5vIGtleSB0aGUgdGVzdHMgaG9sZA"
)

// A preview is served at its link alone, under the key the server holds, until the link lapses: its page in each
// format, each leading to its picture and its own sections at the link, and its picture; at its own address, at a link
// of a wrong token, of another post's, signed with a key the server held before its own, lapsed, or whose lapse is
// moved later under the same signature, and on a server that holds no key, it is no page. Each link is opened on the
// tests' day.
func TestAPreviewOpensAtItsLinkAlone(test *testing.T) {
	store := fixtureStore(test, fixtures.PreviewedFolder(test))
	key, oldKey := keyOf(test, testSecret), keyOf(test, oldSecret)
	lapses := fixtures.Day().Add(linkLasts)
	token := key.Token(document.PostPage(fixtures.PreviewSlug), lapses)
	link := "/preview/" + token + "/" + fixtures.PreviewSlug
	lapsed := "/preview/" + key.Token(document.PostPage(fixtures.PreviewSlug), fixtures.Day().Add(-linkLasts)) + "/previewed.md"
	_, signature, _ := strings.Cut(token, tokenJoint)
	movedLater := "/preview/" + strconv.FormatInt(lapses.Add(linkLasts).Unix(), decimal) + tokenJoint + signature + "/previewed.md"
	translationToken := key.Token(document.PostPage(fixtures.PreviewSlug).InLanguage(fixtures.PreviewLanguage), lapses)
	translationLink := "/ru/preview/" + translationToken + "/" + fixtures.PreviewSlug
	opened, missing := seenAnswer{status: http.StatusOK, holds: true}, seenAnswer{status: http.StatusNotFound, holds: false}
	cases := []struct {
		name       string
		previewKey *PreviewKey
		path       string
		accept     string
		holds      string
		want       seenAnswer
	}{
		{name: "its page", previewKey: key, path: link, accept: "text/html", holds: `<h1 class="post-title"`, want: opened},
		{name: "its page, any format taken", previewKey: key, path: link, accept: "*/*", holds: "PREVIEWED(1)", want: opened},
		{name: "its page, no format named", previewKey: key, path: link, accept: "", holds: "PREVIEWED(1)", want: opened},
		{name: "its markdown", previewKey: key, path: link + ".md", accept: "", holds: "(https://example.test/preview/" + token + "/previewed/margin.png", want: opened},
		{name: "its manual", previewKey: key, path: link + ".txt", accept: "", holds: "<https://example.test/preview/" + token + "/previewed.txt>", want: opened},
		{name: "its picture", previewKey: key, path: link + "/margin.png", accept: "", holds: "PNG", want: opened},
		{name: "its own address", previewKey: key, path: "/blog/previewed.md", accept: "", holds: "# Previewed", want: missing},
		{name: "its picture at its own address", previewKey: key, path: "/blog/previewed/margin.png", accept: "", holds: "PNG", want: missing},
		{name: "a wrong token", previewKey: key, path: "/preview/" + wrongToken + "/previewed.md", accept: "", holds: "# Previewed", want: missing},
		{
			name: "the token of another post", previewKey: key, path: "/preview/" + key.Token(document.PostPage(fixturePostSlug), lapses) + "/previewed.md", accept: "",
			holds: "# Previewed", want: missing,
		},
		{
			name: "a token of an old key", previewKey: key, path: "/preview/" + oldKey.Token(document.PostPage(fixtures.PreviewSlug), lapses) + "/previewed.md", accept: "",
			holds: "# Previewed", want: missing,
		},
		{name: "a link lapsed", previewKey: key, path: lapsed, accept: "", holds: "# Previewed", want: missing},
		{name: "a lapse moved later", previewKey: key, path: movedLater, accept: "", holds: "# Previewed", want: missing},
		{name: "no key held", previewKey: nil, path: link + ".md", accept: "", holds: "# Previewed", want: missing},
		{name: "its translation's page", previewKey: key, path: translationLink, accept: "text/html", holds: `<html lang="ru">`, want: opened},
		{
			name: "its translation's markdown, leading to its picture at the link", previewKey: key, path: translationLink + ".md", accept: "",
			holds: "(https://example.test/ru/preview/" + translationToken + "/previewed/margin.png", want: opened,
		},
		{name: "its translation's picture", previewKey: key, path: translationLink + "/margin.png", accept: "", holds: "PNG", want: opened},
		{name: "its translation at the original's link", previewKey: key, path: "/ru/preview/" + token + "/previewed.md", accept: "", holds: "# Предпросмотр", want: missing},
		{name: "the original at its translation's link", previewKey: key, path: "/preview/" + translationToken + "/previewed.md", accept: "", holds: "# Previewed", want: missing},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			server := drawnServer(test, store, testCase.previewKey)
			got := seenAnswerOf(fixtures.ServedOnTheDay(test, server, http.MethodGet, testCase.path, testCase.accept), testCase.holds)
			if got != testCase.want {
				test.Errorf("GET %s = %+v, want %+v", testCase.path, got, testCase.want)
			}
		})
	}
}

// Every answer in the preview's folder asks no search engine to index it and no cache to keep it, tells a browser to
// send its address, which may hold a token, on to nothing, and names no address it is known by: a preview's page in
// each format and its picture, a path of the link that is no page, the token alone or cut short, the clean form
// net/http sends a path of an empty segment on to, and a request by a method no page takes, alike.
func TestAPreviewIsSentForItsReaderAlone(test *testing.T) {
	key := keyOf(test, testSecret)
	server := drawnServer(test, fixtureStore(test, fixtures.PreviewedFolder(test)), key)
	token := key.Token(document.PostPage(fixtures.PreviewSlug), fixtures.Day().Add(linkLasts))
	translationToken := key.Token(document.PostPage(fixtures.PreviewSlug).InLanguage(fixtures.PreviewLanguage), fixtures.Day().Add(linkLasts))
	link := "/preview/" + token + "/" + fixtures.PreviewSlug
	want := sentPrivately{robots: "noindex", referrer: "no-referrer", cache: "private, no-store", link: ""}
	cases := []struct {
		name   string
		method string
		path   string
		accept string
	}{
		{name: "its page", method: http.MethodGet, path: link, accept: "text/html"},
		{name: "its markdown", method: http.MethodGet, path: link + ".md", accept: ""},
		{name: "its manual", method: http.MethodGet, path: link + ".txt", accept: ""},
		{name: "its picture", method: http.MethodGet, path: link + "/margin.png", accept: ""},
		{name: "no page", method: http.MethodGet, path: link + "/no-such-picture.png", accept: ""},
		{name: "the token alone", method: http.MethodGet, path: "/preview/" + token, accept: "text/html"},
		{name: "the token and its folder's end", method: http.MethodGet, path: "/preview/" + token + "/", accept: "text/html"},
		{name: "a path of an empty segment", method: http.MethodGet, path: "/preview/" + token + "//" + fixtures.PreviewSlug, accept: ""},
		{name: "a method no page takes", method: http.MethodPost, path: link, accept: ""},
		{name: "its translation's page", method: http.MethodGet, path: "/ru/preview/" + translationToken + "/" + fixtures.PreviewSlug, accept: "text/html"},
		{name: "the token alone in a language's folder", method: http.MethodGet, path: "/ru/preview/" + token, accept: "text/html"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			header := fixtures.ServedOnTheDay(test, server, testCase.method, testCase.path, testCase.accept).Header()
			got := sentPrivately{robots: header.Get("X-Robots-Tag"), referrer: header.Get("Referrer-Policy"), cache: header.Get("Cache-Control"), link: header.Get("Link")}
			if got != want {
				test.Errorf("GET %s is sent %+v, want %+v", testCase.path, got, want)
			}
		})
	}
}

// A post's preview link, once the post is out, sends its reader on for a while to the same face of it at its own
// address, its page, its markdown, its manual or its picture, sent as any answer at a preview link is; a link of
// another post's token, or of a draft or no post, answers as no page, so it tells nothing of which posts there are.
func TestAPreviewLinkSendsItsReaderOnOnceThePostIsOut(test *testing.T) {
	key, lapses := keyOf(test, testSecret), fixtures.Day().Add(linkLasts)
	server := drawnServer(test, fixtureStore(test, fixtures.ContentFolder(test)), key)
	link := "/preview/" + key.Token(document.PostPage(fixturePostSlug), lapses) + "/" + fixturePostSlug
	private := sentPrivately{robots: "noindex", referrer: "no-referrer", cache: "private, no-store", link: ""}
	cases := []struct {
		name string
		path string
		want sentOn
	}{
		{name: "its page", path: link, want: sentOn{status: http.StatusFound, location: "/blog/reading-slowly", sent: private}},
		{name: "its markdown", path: link + ".md", want: sentOn{status: http.StatusFound, location: "/blog/reading-slowly.md", sent: private}},
		{name: "its manual", path: link + ".txt", want: sentOn{status: http.StatusFound, location: "/blog/reading-slowly.txt", sent: private}},
		{name: "its picture", path: link + "/margin.png", want: sentOn{status: http.StatusFound, location: picturePath, sent: private}},
		{
			name: "another post's token", path: "/preview/" + key.Token(document.PostPage("notes-on-tools"), lapses) + "/reading-slowly.md",
			want: sentOn{status: http.StatusNotFound, location: "", sent: private},
		},
		{
			name: "a draft's link", path: "/preview/" + key.Token(document.PostPage("notes-on-tools"), lapses) + "/notes-on-tools.md",
			want: sentOn{status: http.StatusNotFound, location: "", sent: private},
		},
		{
			name: "a lapsed link", path: "/preview/" + key.Token(document.PostPage(fixturePostSlug), fixtures.Day().Add(-linkLasts)) + "/reading-slowly.md",
			want: sentOn{status: http.StatusNotFound, location: "", sent: private},
		},
		{
			name: "no post's link", path: "/preview/" + key.Token(document.PostPage("no-such-post"), lapses) + "/no-such-post.md",
			want: sentOn{status: http.StatusNotFound, location: "", sent: private},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := fixtures.ServedOnTheDay(test, server, http.MethodGet, testCase.path, "")
			header := response.Header()
			got := sentOn{
				status: response.Code, location: header.Get("Location"),
				sent: sentPrivately{robots: header.Get("X-Robots-Tag"), referrer: header.Get("Referrer-Policy"), cache: header.Get("Cache-Control"), link: header.Get("Link")},
			}
			if got != testCase.want {
				test.Errorf("GET %s = %+v, want %+v", testCase.path, got, testCase.want)
			}
		})
	}
}

// sentOn is an answer at a preview link as a test reads it: its status, where it sends the reader on to, and what its
// headers tell a search engine, a browser and a cache.
type sentOn struct {
	status   int
	location string
	sent     sentPrivately
}

// sentPrivately is what an answer's headers tell a search engine, a browser and a cache, as a test reads them: whether
// to index it, what of its address to send on, how long to keep it, and the address it is known by.
type sentPrivately struct {
	robots   string
	referrer string
	cache    string
	link     string
}

// A preview key is read from its file; a file missing, or a key shorter than the hash it signs with, is refused.
func TestAPreviewKeyIsReadFromItsFile(test *testing.T) {
	cases := []struct {
		name     string
		file     string
		wantRead bool
	}{
		{name: "a key long enough", file: fixtures.KeyFile(test, testSecret), wantRead: true},
		{name: "a key too short", file: fixtures.KeyFile(test, "short\n"), wantRead: false},
		{name: "no file", file: filepath.Join(test.TempDir(), "missing"), wantRead: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if _, err := ReadPreviewKey(testCase.file); (err == nil) != testCase.wantRead {
				test.Errorf("ReadPreviewKey(%q) = %v, want it read: %t", testCase.file, err, testCase.wantRead)
			}
		})
	}
}

// A preview key's file signs as the key alone, the spaces and line end round it left out, so a copy written with a line
// end signs as one without.
func TestAPreviewKeySignsAsItsCopyWithoutALineEnd(test *testing.T) {
	lapses := fixtures.Day().Add(linkLasts)
	bare := keyOf(test, testSecret).Token(document.PostPage(fixtures.PreviewSlug), lapses)
	cases := []struct {
		name    string
		written string
	}{
		{name: "a key ending a line", written: testSecret + "\n"},
		{name: "a key between spaces", written: "  " + testSecret + "\r\n"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := keyOf(test, testCase.written).Token(document.PostPage(fixtures.PreviewSlug), lapses); got != bare {
				test.Errorf("the key %q signs %q, want %q as the key alone does", testCase.written, got, bare)
			}
		})
	}
}

// A page is only read: a request to change it is refused.
func TestAPageIsOnlyRead(test *testing.T) {
	server := New(fixtureStore(test, fixtures.ContentFolder(test)), false, nil)
	if response := serve(server, http.MethodPost, "/index.md", ""); response.Code != http.StatusMethodNotAllowed {
		test.Errorf("POST /index.md = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

// serve is the server's answer to a request by a method at a path, with an Accept header.
func serve(server http.Handler, method, path, accept string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	request.Header.Set("Accept", accept)
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

// drawnServer is a server of a store, its previews opened under a key, none where there is none, that has drawn its
// snapshot outside any synctest bubble, by a request answered before the test's (fixtures.ServedOnTheDay).
func drawnServer(test *testing.T, store *release.Store, previewKey *PreviewKey) *Server {
	test.Helper()
	server := New(store, false, previewKey)
	serve(server, http.MethodGet, "/", "")
	return server
}

// keyOf is the preview key a file holding a secret is read as.
func keyOf(test *testing.T, secret string) *PreviewKey {
	test.Helper()
	key, err := ReadPreviewKey(fixtures.KeyFile(test, secret))
	if err != nil {
		test.Fatalf("ReadPreviewKey = %v, want a key", err)
	}
	return key
}

// fixtureStore is a store of a release of the content in a folder and the build's fixture.
func fixtureStore(test *testing.T, folder string) *release.Store {
	store, err := release.Open(release.Folders{Content: folder, Build: fixtures.BuildFolder(test)}, func(loaded *content.Snapshot) error {
		return compose.Check(loaded, time.Now())
	})
	if err != nil {
		test.Fatalf("Open(fixtures) = %v, want no error", err)
	}
	return store
}
