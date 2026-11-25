// Package golden holds what a face writes to a golden file in its package's testdata folder, rewritten where a run
// asks for it with -update and read in the diff (.claude/rules/tests.md). Tests alone import it, as depguard's rule
// tests-alone-read-golden holds: each face's test package runs through Main, which knows the flag, and holds a text
// to its file with Match, a picture with MatchPicture. Every file is read and written inside the testdata folder,
// opened as a root, so no path leads out of it.
package golden

import (
	"bytes"
	"flag"
	"os"
	"path"
	"strings"
	"testing"
)

// Where the golden files live, what opens a path written from the root, the flag that asks for the files to be
// rewritten, and who may read and write them
const (
	folder     = "testdata"
	pathRoot   = "/"
	updateFlag = "update"
	updateHelp = "rewrite the golden files from what the face writes now"
	folderMode = 0o750
	fileMode   = 0o600
)

// Main runs a package's tests, with the flag that asks for the golden files to be rewritten: a test package's
// TestMain hands its tests to it.
func Main(tests *testing.M) {
	flag.Bool(updateFlag, false, updateHelp)
	flag.Parse()
	os.Exit(tests.Run())
}

// Match holds a text to the golden file at a path under the testdata folder, written with slashes, as routes
// writes a page's path, its opening slash or not; after rewriting the file where the run asks to.
func Match(test *testing.T, written, got string) {
	test.Helper()
	if want := held(test, written, []byte(got)); got != string(want) {
		test.Errorf("%s =\n%s\nwant\n%s", written, got, want)
	}
}

// MatchPicture holds a picture to its golden file, as Match holds a text. A picture's bytes read as nothing in a log,
// so a failure names the file alone: rewrite it with -update and look at the two.
func MatchPicture(test *testing.T, written string, got []byte) {
	test.Helper()
	if want := held(test, written, got); !bytes.Equal(got, want) {
		test.Errorf("%s differs from its golden picture, %d bytes where it holds %d: rewrite it with -%s and look at both",
			written, len(got), len(want), updateFlag)
	}
}

// held is the golden file at a path, after rewriting it with what came where the run asks to.
func held(test *testing.T, written string, got []byte) []byte {
	test.Helper()
	rewriting := updating()
	root := openFolder(test, rewriting)
	inside := strings.TrimPrefix(written, pathRoot)
	if rewriting {
		write(test, root, inside, got)
	}
	want, err := root.ReadFile(inside)
	if err != nil {
		test.Fatalf("read %s/%s: %v", folder, inside, err)
	}
	return want
}

// openFolder is the testdata folder opened as a root, closed when the test ends; made first where the run rewrites
// the golden files, so a package's first one needs no folder made by hand.
func openFolder(test *testing.T, rewriting bool) *os.Root {
	test.Helper()
	if rewriting {
		if err := os.MkdirAll(folder, folderMode); err != nil {
			test.Fatal(err)
		}
	}
	root, err := os.OpenRoot(folder)
	if err != nil {
		test.Fatal(err)
	}
	test.Cleanup(func() { closeFolder(test, root) })
	return root
}

func closeFolder(test *testing.T, root *os.Root) {
	if err := root.Close(); err != nil {
		test.Errorf("close %s: %v", folder, err)
	}
}

func updating() bool {
	return flag.Lookup(updateFlag).Value.String() == "true"
}

func write(test *testing.T, root *os.Root, inside string, got []byte) {
	test.Helper()
	if err := root.MkdirAll(path.Dir(inside), folderMode); err != nil {
		test.Fatalf("make the folder of %s/%s: %v", folder, inside, err)
	}
	if err := root.WriteFile(inside, got, fileMode); err != nil {
		test.Fatalf("write %s/%s: %v", folder, inside, err)
	}
}
