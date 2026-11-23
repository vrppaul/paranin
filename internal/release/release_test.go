package release

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// The files of the fixtures a test changes, from each one's root, and the author's name the profile holds
const (
	profileFile  = "profile.yaml"
	manifestFile = "manifest.json"
	authorName   = "name: Mira Example"
)

// The fixture post's picture, by its name and its file in the content's folder, and who may read the copy a case
// writes outside it
const (
	pictureName = "margin.png"
	pictureFile = "posts/reading-slowly/" + pictureName
	pictureMode = 0o600
)

// A site cannot open on a release with a fault in either part, and says each fault: the content's, among them those the
// check it is handed finds, and the build's.
func TestASiteCannotOpenOnAFaultyRelease(test *testing.T) {
	cases := []struct {
		name   string
		change func(test *testing.T, folders Folders)
		check  Check
		want   string
	}{
		{name: "faulty content", change: faultyContent, check: passes, want: "person.name"},
		{name: "content its check refuses", change: keepFolders, check: refuses, want: refusal},
		{name: "a faulty build", change: faultyBuild, check: passes, want: manifestFile},
		{name: "a picture linking out of the content's folder", change: pictureLinkedOut, check: passes, want: pictureName},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			folders := fixtureCopies(test)
			testCase.change(test, folders)
			if store, err := Open(folders, testCase.check); store != nil || err == nil || !strings.Contains(err.Error(), testCase.want) {
				test.Errorf("Open = %v, %v; want no store and a fault naming %q", store, err, testCase.want)
			}
		})
	}
}

// A reload that fails in either part keeps the last good snapshot whole, so a page and the files it names are never
// swapped apart; a good one swaps in the new snapshot.
func TestTheStoreKeepsTheLastGoodSnapshot(test *testing.T) {
	folders := fixtureCopies(test)
	store, err := Open(folders, passes)
	if err != nil {
		test.Fatalf("Open(fixtures) = %v, want no error", err)
	}
	fixtures.ReplaceIn(test, folders.Content, profileFile, authorName, "name: Mira Other")
	faultyBuild(test, folders)
	if err := store.Reload(); err == nil || store.Current().Content.Profile.Person.Name != "Mira Example" {
		test.Errorf("a reload with a faulty build = %v with the name %q, want faults and the old name kept", err,
			store.Current().Content.Profile.Person.Name)
	}
	fixtures.ReplaceIn(test, folders.Build, manifestFile, "[", "{")
	if err := store.Reload(); err != nil || store.Current().Content.Profile.Person.Name != "Mira Other" {
		test.Errorf("a good reload = %v with the name %q, want the new name", err, store.Current().Content.Profile.Person.Name)
	}
}

// A value derived from a release is worked out once for each snapshot: readers of one snapshot share it, and a
// reload brings the next reader a value of the new one.
func TestADerivedValueIsWorkedOutOnceForEachSnapshot(test *testing.T) {
	folders := fixtureCopies(test)
	store, err := Open(folders, passes)
	if err != nil {
		test.Fatalf("Open(fixtures) = %v, want no error", err)
	}
	counter := &workCounter{times: 0}
	names := Derive(store, counter.name)
	first, again := names.Current(), names.Current()
	fixtures.ReplaceIn(test, folders.Content, profileFile, authorName, "name: Mira Other")
	if err := store.Reload(); err != nil {
		test.Fatalf("Reload = %v, want no error", err)
	}
	reloaded := names.Current()
	if first != "Mira Example" || again != first || reloaded != "Mira Other" || counter.times != 2 {
		test.Errorf("values %q, %q, then %q after a reload, worked out %d times; want %q twice, then %q, worked out twice",
			first, again, reloaded, counter.times, "Mira Example", "Mira Other")
	}
}

// workCounter counts how many times a value is worked out.
type workCounter struct {
	times int
}

// name works out the author's name, and counts it.
func (counter *workCounter) name(snapshot *Snapshot) string {
	counter.times++
	return snapshot.Content.Profile.Person.Name
}

// fixtureCopies are copies of both fixtures on disk, which a test may change.
func fixtureCopies(test *testing.T) Folders {
	return Folders{Content: fixtures.Copy(test, fixtures.ContentFolder(test)), Build: fixtures.Copy(test, fixtures.BuildFolder(test))}
}

// What the check a case hands in says where it refuses the content
const refusal = "the check refuses it"

// passes is a check that finds no fault.
func passes(*content.Snapshot) error {
	return nil
}

// refuses is a check that finds a fault in any content.
//
//wrapper:a case of the table names the check it hands the release
func refuses(*content.Snapshot) error {
	return errors.New(refusal)
}

// keepFolders leaves both fixtures as they are.
func keepFolders(*testing.T, Folders) {}

// faultyContent leaves the author with no name.
//
//wrapper:a case of the table names the fault it writes
func faultyContent(test *testing.T, folders Folders) {
	fixtures.ReplaceIn(test, folders.Content, profileFile, authorName, `name: ""`)
}

// pictureLinkedOut makes the fixture post's picture a link to a copy of it outside the content's folder, which a read
// of the folder must not follow.
func pictureLinkedOut(test *testing.T, folders Folders) {
	picture := filepath.Join(folders.Content, pictureFile)
	outside := filepath.Join(test.TempDir(), pictureName)
	written, err := os.ReadFile(picture)
	if err == nil {
		err = os.WriteFile(outside, written, pictureMode)
	}
	if err == nil {
		err = os.Remove(picture)
	}
	if err == nil {
		err = os.Symlink(outside, picture)
	}
	if err != nil {
		test.Fatalf("link %s out of the content's folder: %v", pictureFile, err)
	}
}

// faultyBuild leaves a manifest that does not read.
//
//wrapper:a case of the table names the fault it writes
func faultyBuild(test *testing.T, folders Folders) {
	fixtures.ReplaceIn(test, folders.Build, manifestFile, "{", "[")
}
