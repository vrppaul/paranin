package release

import (
	"strings"
	"testing"

	"github.com/vrppaul/paranin/internal/fixtures"
)

// The files of the fixtures a test changes, from each one's root, and the author's name the profile holds
const (
	profileFile  = "profile.yaml"
	manifestFile = "manifest.json"
	authorName   = "name: Mira Example"
)

// A site cannot open on a release with a fault in either part, and says each fault: the content's and the build's.
func TestASiteCannotOpenOnAFaultyRelease(test *testing.T) {
	cases := []struct {
		name   string
		change func(test *testing.T, folders folderPaths)
		want   string
	}{
		{name: "faulty content", change: faultyContent, want: "person.name"},
		{name: "a faulty build", change: faultyBuild, want: manifestFile},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			folders := fixtureCopies(test)
			testCase.change(test, folders)
			if store, err := Open(FoldersAt(folders.content, folders.build)); store != nil || err == nil || !strings.Contains(err.Error(), testCase.want) {
				test.Errorf("Open = %v, %v; want no store and a fault naming %q", store, err, testCase.want)
			}
		})
	}
}

// A reload that fails in either part keeps the last good snapshot whole, so a page and the files it names are never
// swapped apart; a good one swaps in the new snapshot.
func TestTheStoreKeepsTheLastGoodSnapshot(test *testing.T) {
	folders := fixtureCopies(test)
	store, err := Open(FoldersAt(folders.content, folders.build))
	if err != nil {
		test.Fatalf("Open(fixtures) = %v, want no error", err)
	}
	fixtures.ReplaceIn(test, folders.content, profileFile, authorName, "name: Mira Other")
	faultyBuild(test, folders)
	if err := store.Reload(); err == nil || store.Current().Content.Profile.Person.Name != "Mira Example" {
		test.Errorf("a reload with a faulty build = %v with the name %q, want faults and the old name kept", err,
			store.Current().Content.Profile.Person.Name)
	}
	fixtures.ReplaceIn(test, folders.build, manifestFile, "[", "{")
	if err := store.Reload(); err != nil || store.Current().Content.Profile.Person.Name != "Mira Other" {
		test.Errorf("a good reload = %v with the name %q, want the new name", err, store.Current().Content.Profile.Person.Name)
	}
}

// A value derived from a release is worked out once for each snapshot: readers of one snapshot share it, and a
// reload brings the next reader a value of the new one.
func TestADerivedValueIsWorkedOutOnceForEachSnapshot(test *testing.T) {
	folders := fixtureCopies(test)
	store, err := Open(FoldersAt(folders.content, folders.build))
	if err != nil {
		test.Fatalf("Open(fixtures) = %v, want no error", err)
	}
	counter := &workCounter{times: 0}
	names := Derive(store, counter.name)
	first, again := names.Current(), names.Current()
	fixtures.ReplaceIn(test, folders.content, profileFile, authorName, "name: Mira Other")
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

// folderPaths are the folders of a release's two parts on disk, which a test may change.
type folderPaths struct {
	content string
	build   string
}

// fixtureCopies are copies of both fixtures on disk.
func fixtureCopies(test *testing.T) folderPaths {
	return folderPaths{content: fixtures.Copy(test, fixtures.ContentFolder(test)), build: fixtures.Copy(test, fixtures.BuildFolder(test))}
}

// faultyContent leaves the author with no name.
//
//wrapper:a case of the table names the fault it writes
func faultyContent(test *testing.T, folders folderPaths) {
	fixtures.ReplaceIn(test, folders.content, profileFile, authorName, `name: ""`)
}

// faultyBuild leaves a manifest that does not read.
//
//wrapper:a case of the table names the fault it writes
func faultyBuild(test *testing.T, folders folderPaths) {
	fixtures.ReplaceIn(test, folders.build, manifestFile, "{", "[")
}
