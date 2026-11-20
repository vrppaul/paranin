// Package fixtures is the tests' way to the fixtures the packages' tests share (.claude/rules/tests.md): the
// content's, in internal/content/testdata, and the build's, in internal/assets/testdata, each found from the root
// of the module, which go test's folder for a package lies under however the test was built, so no test writes a
// path that climbs out of its package; each loaded, as the site loads it; a copy of either on disk, which a test
// may change; the day the faces draw on; the head of a page a test builds by hand; and the document of a page among
// documents. Tests alone import it, as depguard's rule tests-alone-read-fixtures holds; a test of content or assets,
// which it loads with, imports it from outside its package.
package fixtures

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vrppaul/paranin/internal/assets"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// Where the fixtures live, from the module's root, the file the root is known by, and who may write a changed file
// of a copy
const (
	contentFromRoot = "internal/content/testdata/content"
	buildFromRoot   = "internal/assets/testdata/build"
	moduleFile      = "go.mod"
	fileMode        = 0o600
)

// ContentFolder is the folder of the fixture content: a profile and the site's words.
func ContentFolder(test *testing.T) string {
	return filepath.Join(moduleRoot(test), contentFromRoot)
}

// BuildFolder is the folder of the build's fixture: a manifest naming the site's entry, the chunks it imports, two
// deep and one of them twice, and the file each was built into, with the stylesheets each imports.
func BuildFolder(test *testing.T) string {
	return filepath.Join(moduleRoot(test), buildFromRoot)
}

// The year and the day of the month the faces' tests draw on, in November
const (
	dayYear       = 2026
	dayOfTheMonth = 1
)

// Day is the day the faces' tests draw on, a TUI's session opening and the page's career line, the fixture's role
// still held running to it.
//
//wrapper:names the one day every face's test draws on
func Day() time.Time {
	return time.Date(dayYear, time.November, dayOfTheMonth, 0, 0, 0, 0, time.UTC)
}

// HeadOf is the head of a page a test builds by hand: the page and its title, nothing else, so a field the head gains
// reaches no test that has no use for it.
func HeadOf(page document.Page, title []document.Span) document.Head {
	return document.Head{
		Page: page, Language: "", Greeting: "", Title: title, Subtitle: nil, Place: "", Tags: nil, Lead: nil,
		Repository: nil, Description: nil,
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
// says.
func WriteIn(test *testing.T, folder, file, text string) {
	test.Helper()
	writeAt(test, openFolder(test, folder), file, text)
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
