package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vrppaul/paranin/internal/content"
)

// The content package's fixture
const fixturePath = "../../content/testdata/content"

// Each page is served at its address in each format, and the index at llms.txt as plain text, as agents read it.
func TestEveryPageIsServedAtItsAddress(test *testing.T) {
	server := New(fixtureStore(test, fixtureCopy(test)))
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
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := serve(server, http.MethodGet, testCase.path)
			if response.Code != testCase.status || response.Header().Get("Content-Type") != testCase.contentType ||
				!strings.Contains(response.Body.String(), testCase.holds) {
				test.Errorf("GET %s = %d %q holding %q, want %d %q holding %q", testCase.path, response.Code,
					response.Header().Get("Content-Type"), response.Body.String(), testCase.status, testCase.contentType, testCase.holds)
			}
		})
	}
}

// At a page's own address the reader's request chooses the format: HTML for a browser, markdown for whoever asks
// for it, the manual for everyone else, `curl` among them, and for a request that ranks two of them alike; and a
// cache keeps one copy for each Accept.
func TestTheReaderChoosesTheFormat(test *testing.T) {
	server := New(fixtureStore(test, fixtureCopy(test)))
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
		{name: "a project's page in a browser", path: "/p/tidy-tool", accept: "text/html", contentType: browsed, holds: "<h1>tidy-tool</h1>"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			request := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			request.Header.Set("Accept", testCase.accept)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)
			header := response.Header()
			if header.Get("Content-Type") != testCase.contentType || !strings.Contains(response.Body.String(), testCase.holds) ||
				header.Get("Vary") != "Accept" {
				test.Errorf("GET %s, Accept %q = %q %q, Vary %q; want %q holding %q, Vary %q", testCase.path, testCase.accept,
					header.Get("Content-Type"), response.Body.String(), header.Get("Vary"), testCase.contentType, testCase.holds, "Accept")
			}
		})
	}
}

// Each format of a page names the address the page is known by, whole, so a search engine counts its markdown and
// its manual as that page; the index, which is no page a person reads, names none.
func TestEveryFormatOfAPageNamesItsCanonicalAddress(test *testing.T) {
	server := New(fixtureStore(test, fixtureCopy(test)))
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
		{name: "the index", path: "/llms.txt", accept: "", want: ""},
		{name: "no page", path: "/nothing", accept: "", want: ""},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			request := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			request.Header.Set("Accept", testCase.accept)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)
			if got := response.Header().Get("Link"); got != testCase.want {
				test.Errorf("GET %s, Accept %q: Link = %q, want %q", testCase.path, testCase.accept, got, testCase.want)
			}
		})
	}
}

// An address of no page is not found, and says so with the status it is answered with and the path asked for, as
// the request decoded it: in plain text to a terminal, an agent and a request naming nothing, in HTML, in the
// site's language, to a browser; and a cache keeps one copy for each Accept.
func TestAnAddressOfNoPageIsNotFound(test *testing.T) {
	server := New(fixtureStore(test, fixtureCopy(test)))
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
			request := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			request.Header.Set("Accept", testCase.accept)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)
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

// A page tells a browser not to guess its type, and a cache to ask again each time, since a reload can change it.
func TestAPageTellsBrowsersNotToGuessAndCachesToAskAgain(test *testing.T) {
	header := serve(New(fixtureStore(test, fixtureCopy(test))), http.MethodGet, "/index.md").Header()
	cases := []struct {
		name   string
		header string
		want   string
	}{
		{name: "no guessing its type", header: "X-Content-Type-Options", want: "nosniff"},
		{name: "a cache asks again", header: "Cache-Control", want: "no-cache"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := header.Get(testCase.header); got != testCase.want {
				test.Errorf("%s = %q, want %q", testCase.header, got, testCase.want)
			}
		})
	}
}

// A page is drawn again once the store holds new content: a reload reaches the next request.
func TestAReloadReachesTheNextRequest(test *testing.T) {
	folder := fixtureCopy(test)
	store := fixtureStore(test, folder)
	server := New(store)
	serve(server, http.MethodGet, "/index.md")
	profilePath := filepath.Join(folder, "profile.yaml")
	profile, err := os.ReadFile(profilePath)
	if err != nil {
		test.Fatalf("read the copy's profile: %v", err)
	}
	renamed := strings.Replace(string(profile), "name: Mira Example", "name: Mira Other", 1)
	if err := os.WriteFile(profilePath, []byte(renamed), 0o600); err != nil {
		test.Fatalf("write the copy's profile: %v", err)
	}
	if err := store.Reload(); err != nil {
		test.Fatalf("Reload = %v, want no error", err)
	}
	if body := serve(server, http.MethodGet, "/index.md").Body.String(); !strings.Contains(body, "# Mira Other") {
		test.Errorf("GET /index.md after a reload = %q, want the new name", body)
	}
}

// A page is only read: a request to change it is refused.
func TestAPageIsOnlyRead(test *testing.T) {
	server := New(fixtureStore(test, fixtureCopy(test)))
	if response := serve(server, http.MethodPost, "/index.md"); response.Code != http.StatusMethodNotAllowed {
		test.Errorf("POST /index.md = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

func serve(server http.Handler, method, path string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequest(method, path, nil))
	return response
}

// fixtureCopy is a copy of the fixture content on disk, which a test may change.
func fixtureCopy(test *testing.T) string {
	folder := test.TempDir()
	if err := os.CopyFS(folder, os.DirFS(fixturePath)); err != nil {
		test.Fatalf("copy the fixture: %v", err)
	}
	return folder
}

func fixtureStore(test *testing.T, folder string) *content.Store {
	store, err := content.Open(os.DirFS(folder))
	if err != nil {
		test.Fatalf("Open(fixture) = %v, want no error", err)
	}
	return store
}
