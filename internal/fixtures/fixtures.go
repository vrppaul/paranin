// Package fixtures is the tests' way to the fixtures the packages' tests share (.claude/rules/tests.md): the
// content's, in internal/content/testdata, and the build's, in internal/assets/testdata, each found from the root
// of the module, which go test's folder for a package lies under however the test was built, so no test writes a
// path that climbs out of its package; each loaded, as the site loads it; and a copy of either on disk, which a test
// may change. Tests alone import it, as depguard's rule tests-alone-read-fixtures holds; a test of content or
// assets, which it loads with, imports it from outside its package.
package fixtures

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vrppaul/paranin/internal/assets"
	"github.com/vrppaul/paranin/internal/content"
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

// BuildFolder is the folder of the build's fixture: a manifest naming the stylesheet the site names, and its file.
func BuildFolder(test *testing.T) string {
	return filepath.Join(moduleRoot(test), buildFromRoot)
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
// and written inside the folder, opened as a root, so no path leads out of it.
func ReplaceIn(test *testing.T, folder, file, original, replacement string) {
	test.Helper()
	root, err := os.OpenRoot(folder)
	if err != nil {
		test.Fatalf("open %s: %v", folder, err)
	}
	test.Cleanup(func() { closeFolder(test, root) })
	written, err := root.ReadFile(file)
	if err != nil {
		test.Fatalf("read %s: %v", file, err)
	}
	changed := strings.Replace(string(written), original, replacement, 1)
	if changed == string(written) {
		test.Fatalf("%s holds no %q", file, original)
	}
	if err := root.WriteFile(file, []byte(changed), fileMode); err != nil {
		test.Fatalf("write %s: %v", file, err)
	}
}

func closeFolder(test *testing.T, root *os.Root) {
	if err := root.Close(); err != nil {
		test.Errorf("close %s: %v", root.Name(), err)
	}
}
