package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/release"
)

// The build's fixture's stylesheet and script, which every page a browser reads includes
const (
	stylesheetPath = "/assets/site-fixture.css"
	scriptPath     = "/assets/site-fixture.js"
)

// Each page is served at its address in each format, and the index at llms.txt as plain text, as agents read it.
func TestEveryPageIsServedAtItsAddress(test *testing.T) {
	server := New(fixtureStore(test, fixtures.ContentFolder(test)))
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
		{name: "a draft", path: "/blog/notes-on-tools.md", status: http.StatusNotFound, contentType: "text/plain; charset=utf-8", holds: "404"},
		{name: "the blog's list", path: "/blog.md", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "## 2026"},
		{name: "the blog's list as a manual", path: "/blog.txt", status: http.StatusOK, contentType: "text/plain; charset=utf-8", holds: "BLOG(1)"},
		{name: "the list by series", path: "/blog/series.md", status: http.StatusOK, contentType: "text/markdown; charset=utf-8", holds: "1 of 2 parts out"},
		{name: "the list by series as a manual", path: "/blog/series.txt", status: http.StatusOK, contentType: "text/plain; charset=utf-8", holds: "BLOG-SERIES(1)"},
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
	server := New(fixtureStore(test, fixtures.ContentFolder(test)))
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
	server := New(fixtureStore(test, fixtures.ContentFolder(test)))
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
	server := New(fixtureStore(test, fixtures.ContentFolder(test)))
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
// its manual as that page; the index, which is no page a person reads, names none.
func TestEveryFormatOfAPageNamesItsCanonicalAddress(test *testing.T) {
	server := New(fixtureStore(test, fixtures.ContentFolder(test)))
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

// A post's number sends the reader on for good to the post's own address, by its path from the root, and a cache asks
// again, since the post's slug is its folder's name, which may change; the number of a post not served is no page.
func TestAPostsNumberSendsTheReaderOnToItsAddress(test *testing.T) {
	server := New(fixtureStore(test, fixtures.ContentFolder(test)))
	cases := []struct {
		name     string
		path     string
		status   int
		location string
	}{
		{name: "a post published", path: "/blog/1", status: http.StatusMovedPermanently, location: "/blog/reading-slowly"},
		{name: "a draft", path: "/blog/2", status: http.StatusNotFound, location: ""},
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

// An address of no page is not found, and says so with the status it is answered with and the path asked for, as
// the request decoded it: in plain text to a terminal, an agent and a request naming nothing, in HTML, in the
// site's language, to a browser; and a cache keeps one copy for each Accept.
func TestAnAddressOfNoPageIsNotFound(test *testing.T) {
	server := New(fixtureStore(test, fixtures.ContentFolder(test)))
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
	server := New(fixtureStore(test, fixtures.ContentFolder(test)))
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

// A page tells a cache to ask again each time, since a reload can change it.
func TestAPageTellsCachesToAskAgain(test *testing.T) {
	response := serve(New(fixtureStore(test, fixtures.ContentFolder(test))), http.MethodGet, "/index.md", "")
	if got := response.Header().Get("Cache-Control"); got != "no-cache" {
		test.Errorf("GET /index.md: Cache-Control = %q, want %q", got, "no-cache")
	}
}

// A file of the build is served at its path as the type its name says, whatever the request asks for, and a cache
// keeps it for good: a file changed is built under another name.
func TestABuiltFileIsServedForCachesToKeepForGood(test *testing.T) {
	response := serve(New(fixtureStore(test, fixtures.ContentFolder(test))), http.MethodGet, stylesheetPath, "*/*")
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
	server := New(fixtureStore(test, fixtures.ContentFolder(test)))
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
	server := New(fixtureStore(test, fixtures.ContentFolder(test)))
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

// A page is drawn again once the store holds new content: a reload reaches the next request.
func TestAReloadReachesTheNextRequest(test *testing.T) {
	folder := fixtures.Copy(test, fixtures.ContentFolder(test))
	store := fixtureStore(test, folder)
	server := New(store)
	serve(server, http.MethodGet, "/index.md", "")
	fixtures.ReplaceIn(test, folder, "profile.yaml", "name: Mira Example", "name: Mira Other")
	if err := store.Reload(); err != nil {
		test.Fatalf("Reload = %v, want no error", err)
	}
	if body := serve(server, http.MethodGet, "/index.md", "").Body.String(); !strings.Contains(body, "# Mira Other") {
		test.Errorf("GET /index.md after a reload = %q, want the new name", body)
	}
}

// A page is only read: a request to change it is refused.
func TestAPageIsOnlyRead(test *testing.T) {
	server := New(fixtureStore(test, fixtures.ContentFolder(test)))
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

// fixtureStore is a store of a release of the content in a folder and the build's fixture.
func fixtureStore(test *testing.T, folder string) *release.Store {
	store, err := release.Open(release.FoldersAt(folder, fixtures.BuildFolder(test)), compose.SharedAddresses)
	if err != nil {
		test.Fatalf("Open(fixtures) = %v, want no error", err)
	}
	return store
}
