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

// At a page's own address the reader's request chooses the format: markdown for whoever asks for it, the manual
// for everyone else, `curl` among them; and a cache keeps one copy for each Accept.
func TestTheReaderChoosesTheFormat(test *testing.T) {
	server := New(fixtureStore(test, fixtureCopy(test)))
	cases := []struct {
		name   string
		path   string
		accept string
		holds  string
	}{
		{name: "an agent asking for markdown", path: "/", accept: "text/markdown", holds: "# Mira Example"},
		{name: "curl, taking anything", path: "/", accept: "*/*", holds: "EXAMPLE.TEST(1)"},
		{name: "a request asking for nothing", path: "/", accept: "", holds: "EXAMPLE.TEST(1)"},
		{name: "markdown refused", path: "/", accept: "text/markdown;q=0, */*", holds: "EXAMPLE.TEST(1)"},
		{name: "markdown among others", path: "/", accept: "text/html;q=0.9, text/markdown", holds: "# Mira Example"},
		{name: "plain text ranked above markdown", path: "/", accept: "text/plain, text/markdown;q=0.1", holds: "EXAMPLE.TEST(1)"},
		{name: "markdown ranked above plain text", path: "/", accept: "text/markdown, text/plain;q=0.5", holds: "# Mira Example"},
		{name: "both of a family", path: "/", accept: "text/*", holds: "EXAMPLE.TEST(1)"},
		{name: "markdown below anything", path: "/", accept: "text/markdown;q=0.5, */*", holds: "EXAMPLE.TEST(1)"},
		{name: "markdown below its family", path: "/", accept: "text/markdown;q=0.5, text/*", holds: "EXAMPLE.TEST(1)"},
		{name: "a browser asking for HTML", path: "/", accept: "text/html", holds: "EXAMPLE.TEST(1)"},
		{name: "a project's page in markdown", path: "/p/tidy-tool", accept: "text/markdown", holds: "# tidy-tool"},
		{name: "a project's page as a manual", path: "/p/tidy-tool", accept: "*/*", holds: "TIDY-TOOL(1)"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			request := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			request.Header.Set("Accept", testCase.accept)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)
			if !strings.Contains(response.Body.String(), testCase.holds) || response.Header().Get("Vary") != "Accept" {
				test.Errorf("GET %s, Accept %q = %q, Vary %q; want it to hold %q, Vary %q", testCase.path, testCase.accept,
					response.Body.String(), response.Header().Get("Vary"), testCase.holds, "Accept")
			}
		})
	}
}

// An address of no page is not found, in plain text opening with the status it is answered with and the path
// asked for, as the request decoded it.
func TestAnAddressOfNoPageIsNotFound(test *testing.T) {
	server := New(fixtureStore(test, fixtureCopy(test)))
	cases := []struct {
		name  string
		path  string
		opens string
	}{
		{name: "a project with no page", path: "/p/other-tool.md", opens: `404 Not Found: "/p/other-tool.md"`},
		{name: "no page at all", path: "/nothing", opens: `404 Not Found: "/nothing"`},
		{name: "a path written escaped", path: "/no%20thing", opens: `404 Not Found: "/no thing"`},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			response := serve(server, http.MethodGet, testCase.path)
			if response.Code != http.StatusNotFound || response.Header().Get("Content-Type") != "text/plain; charset=utf-8" ||
				!strings.HasPrefix(response.Body.String(), testCase.opens) {
				test.Errorf("GET %s = %d %q, %q; want %d %q, opening %q", testCase.path, response.Code,
					response.Header().Get("Content-Type"), response.Body.String(), http.StatusNotFound, "text/plain; charset=utf-8", testCase.opens)
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
