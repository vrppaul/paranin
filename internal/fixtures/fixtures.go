// Package fixtures is the tests' one way to the fixtures the packages' tests share (.claude/rules/tests.md), so no
// test writes a path that climbs out of its package. Tests alone import it (depguard's tests-alone-read-fixtures);
// a test of a package it imports reaches it from an external test package.
package fixtures

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/vrppaul/paranin/internal/assets"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// Where the fixtures live, from the module's root, the file the root is known by, and who may write a changed file
// of a copy and a folder made in it
const (
	contentFromRoot = "internal/content/testdata/content"
	buildFromRoot   = "internal/assets/testdata/build"
	moduleFile      = "go.mod"
	fileMode        = 0o600
	folderMode      = 0o700
)

// ContentFolder is the folder of the fixture content: a profile, the site's words, a post published and a draft, and
// a series.
func ContentFolder(test *testing.T) string {
	return filepath.Join(moduleRoot(test), contentFromRoot)
}

// BuildFolder is the folder of the build's fixture: a manifest naming the site's entry, the chunks it imports, two
// deep and one of them twice, and the file each was built into, with the stylesheets each imports and a font the
// entry's stylesheet loads.
func BuildFolder(test *testing.T) string {
	return filepath.Join(moduleRoot(test), buildFromRoot)
}

// The year and the day of the month the tests draw on, in November
const (
	dayYear       = 2026
	dayOfTheMonth = 23
)

// Day is the day the tests draw on, a TUI's session opening and the page's career line, the fixture's role still
// held running to it, and read the site on, after the fixture's post, and every post a test dates, has come out.
//
//wrapper:names the one day every test draws on
func Day() time.Time {
	return time.Date(dayYear, time.November, dayOfTheMonth, 0, 0, 0, 0, time.UTC)
}

// Site is the site of a snapshot, at the address its profile gives.
//
//wrapper:names the site every face's test writes its pages and sessions for
func Site(snapshot *content.Snapshot) routes.Site {
	return routes.New(snapshot.Profile.Site.BaseURL)
}

// HeadOf is the head of a page a test builds by hand: the page and its title, nothing else, so a field the head gains
// reaches no test that has no use for it.
func HeadOf(page document.Page, title []document.Span) document.Head {
	return document.Head{
		Page: page, Number: 0, Language: "", Greeting: "", Title: title, Subtitle: nil, Place: "", Tags: nil, Lead: nil,
		Version: nil, Repository: nil, Post: nil, Description: nil,
	}
}

// DocumentOf is the document among documents that is a page, or the test's end where none is.
func DocumentOf(test *testing.T, documents []document.Document, page document.Page) document.Document {
	test.Helper()
	for _, found := range documents {
		if found.Head.Page == page {
			return found
		}
	}
	test.Fatalf("no document is the page %+v", page)
	return document.Document{Head: HeadOf(page, nil), Blocks: nil}
}

// LoadedContent is the fixture content, loaded.
func LoadedContent(test *testing.T) *content.Snapshot {
	test.Helper()
	snapshot, err := content.Load(os.DirFS(ContentFolder(test)))
	if err != nil {
		test.Fatalf("load the fixture content: %v", err)
	}
	return snapshot
}

// LoadedBuild is the build's fixture, loaded.
func LoadedBuild(test *testing.T) *assets.Build {
	test.Helper()
	build, err := assets.Load(os.DirFS(BuildFolder(test)))
	if err != nil {
		test.Fatalf("load the build's fixture: %v", err)
	}
	return build
}

// moduleRoot is the folder of the module: the nearest above the test's own folder, where go test runs it, that
// holds go.mod.
func moduleRoot(test *testing.T) string {
	test.Helper()
	folder, err := os.Getwd()
	if err != nil {
		test.Fatalf("find the test's folder: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(folder, moduleFile)); err == nil {
			return folder
		}
		parent := filepath.Dir(folder)
		if parent == folder {
			test.Fatalf("no folder above the test's holds %s", moduleFile)
		}
		folder = parent
	}
}

// A preview the tests write into a copy of the fixture content: its slug, its file, its picture's and the fixture
// post's picture it is a copy of, and its text, which shows the picture and leads to a heading of its own; and the
// name of a preview key's file a test writes
const (
	PreviewSlug        = "previewed"
	previewFile        = "posts/previewed/post.en.md"
	previewPictureFile = "posts/previewed/margin.png"
	fixturePictureFile = "posts/reading-slowly/margin.png"
	previewText        = "---\nid: 3\ntitle: Previewed\ndescription: About what is previewed.\nstatus: preview\npublished: 2026-11-20T09:00+01:00\n---\n" +
		"## Its heading\n\n![A page](margin.png)\n\n[Back up](/blog/previewed#its-heading)\n"
	keyFileName = "preview_key"
)

// The header a request names the formats it takes in
const acceptHeader = "Accept"

// PreviewedFolder is a copy of the fixture content holding a preview, with a picture beside it, gone when the test
// ends.
func PreviewedFolder(test *testing.T) string {
	test.Helper()
	folder := Copy(test, ContentFolder(test))
	picture, err := openFolder(test, folder).ReadFile(fixturePictureFile)
	if err != nil {
		test.Fatalf("read the fixture post's picture: %v", err)
	}
	WriteIn(test, folder, previewFile, previewText)
	WriteIn(test, folder, previewPictureFile, string(picture))
	return folder
}

// KeyFile is the path of a file of the test's own holding a text, as a preview key's file holds it, which its owner
// alone may read.
func KeyFile(test *testing.T, written string) string {
	test.Helper()
	path := filepath.Join(test.TempDir(), keyFileName)
	if err := os.WriteFile(path, []byte(written), fileMode); err != nil {
		test.Fatalf("write a preview key: %v", err)
	}
	return path
}

// ServedOnTheDay is a handler's answer to a request by a method at a path, with an Accept header, on the tests' day: in
// a synctest bubble whose clock is moved on to that day, so a link's lapse is held to it and never to the real clock.
// The handler has drawn what it draws outside the bubble: preparing a snapshot splits its posts' code by chroma
// (compose.Prepare), whose regexps start a clock goroutine of their own (regexp2's fastclock) that no bubble outlives.
func ServedOnTheDay(test *testing.T, handler http.Handler, method, path, accept string) *httptest.ResponseRecorder {
	var response *httptest.ResponseRecorder
	synctest.Test(test, func(*testing.T) { response = servedOnTheDay(handler, method, path, accept) })
	return response
}

// servedOnTheDay waits until the tests' day, which a synctest bubble's clock reaches at once, and is the handler's
// answer to a request then.
func servedOnTheDay(handler http.Handler, method, path, accept string) *httptest.ResponseRecorder {
	time.Sleep(time.Until(Day()))
	request := httptest.NewRequest(method, path, nil)
	request.Header.Set(acceptHeader, accept)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

// Copy is a copy of a fixture's folder on disk, which a test may change, gone when the test ends.
func Copy(test *testing.T, fixture string) string {
	test.Helper()
	folder := test.TempDir()
	if err := os.CopyFS(folder, os.DirFS(fixture)); err != nil {
		test.Fatalf("copy %s: %v", fixture, err)
	}
	return folder
}

// ReplaceIn writes the first of a text in a file of a folder as another; the text must be there. The file is read
// and written inside the folder, opened as a root, so no path leads out of it, and it is written at the time the
// test's clock says, which a test in synctest's bubble sets, whatever the file system's own clock and its steps.
func ReplaceIn(test *testing.T, folder, file, original, replacement string) {
	test.Helper()
	root := openFolder(test, folder)
	written, err := root.ReadFile(file)
	if err != nil {
		test.Fatalf("read %s: %v", file, err)
	}
	changed := strings.Replace(string(written), original, replacement, 1)
	if changed == string(written) {
		test.Fatalf("%s holds no %q", file, original)
	}
	writeAt(test, root, file, changed)
}

// WriteIn writes a file of a folder whole, as ReplaceIn writes it: inside the folder, at the time the test's clock
// says; the folders it stands in are made where they are missing.
func WriteIn(test *testing.T, folder, file, text string) {
	test.Helper()
	root := openFolder(test, folder)
	if err := root.MkdirAll(filepath.Dir(file), folderMode); err != nil {
		test.Fatalf("make the folders of %s: %v", file, err)
	}
	writeAt(test, root, file, text)
}

// openFolder is a folder opened as a root, closed when the test ends.
func openFolder(test *testing.T, folder string) *os.Root {
	test.Helper()
	root, err := os.OpenRoot(folder)
	if err != nil {
		test.Fatalf("open %s: %v", folder, err)
	}
	test.Cleanup(func() { closeFolder(test, root) })
	return root
}

// writeAt writes a file of a root at the time the test's clock says.
func writeAt(test *testing.T, root *os.Root, file, text string) {
	test.Helper()
	if err := root.WriteFile(file, []byte(text), fileMode); err != nil {
		test.Fatalf("write %s: %v", file, err)
	}
	now := time.Now()
	if err := root.Chtimes(file, now, now); err != nil {
		test.Fatalf("set the time of %s: %v", file, err)
	}
}

func closeFolder(test *testing.T, root *os.Root) {
	if err := root.Close(); err != nil {
		test.Errorf("close %s: %v", root.Name(), err)
	}
}
