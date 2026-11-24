package release

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

// A value derived from a release is worked out once for each snapshot: readers of one snapshot share it, whether they
// ask for the store's current one or hand that snapshot in, and a reload brings the next reader a value of the new one.
func TestADerivedValueIsWorkedOutOnceForEachSnapshot(test *testing.T) {
	folders := fixtureCopies(test)
	store, err := Open(folders, passes)
	if err != nil {
		test.Fatalf("Open(fixtures) = %v, want no error", err)
	}
	counter := &workCounter{times: 0, lapses: time.Time{}}
	names := Derive(store, counter.name)
	first, again := names.Current(fixtures.Day()), names.Of(store.Current(), fixtures.Day())
	fixtures.ReplaceIn(test, folders.Content, profileFile, authorName, "name: Mira Other")
	if err := store.Reload(); err != nil {
		test.Fatalf("Reload = %v, want no error", err)
	}
	reloaded := names.Current(fixtures.Day())
	if first != "Mira Example" || again != first || reloaded != "Mira Other" || counter.times != 2 {
		test.Errorf("values %q, %q, then %q after a reload, worked out %d times; want %q twice, then %q, worked out twice",
			first, again, reloaded, counter.times, "Mira Example", "Mira Other")
	}
}

// A value derived is held until the moment its work says it lapses at and worked out again from then; one that never
// lapses is held for as long as its snapshot.
func TestADerivedValueIsWorkedOutAgainOnceLapsed(test *testing.T) {
	store, err := Open(fixtureCopies(test), passes)
	if err != nil {
		test.Fatalf("Open(fixtures) = %v, want no error", err)
	}
	lapse := fixtures.Day().Add(time.Hour)
	counter := &workCounter{times: 0, lapses: lapse}
	names := Derive(store, counter.name)
	names.Current(fixtures.Day())
	names.Current(lapse.Add(-time.Minute))
	beforeLapse := counter.times
	counter.lapses = time.Time{}
	names.Current(lapse)
	names.Current(lapse.Add(time.Hour))
	if beforeLapse != 1 || counter.times != 2 {
		test.Errorf("worked out %d times before the value lapsed and %d times after; want once, then twice", beforeLapse, counter.times)
	}
}

// A value derived from a snapshot's content alone is worked out once for each snapshot and held at every moment after,
// until a reload brings another.
func TestAValueOfTheContentAloneLapsesWithItsSnapshot(test *testing.T) {
	folders := fixtureCopies(test)
	store, err := Open(folders, passes)
	if err != nil {
		test.Fatalf("Open(fixtures) = %v, want no error", err)
	}
	counter := &workCounter{times: 0, lapses: time.Time{}}
	names := DeriveFromContent(store, counter.contentName)
	names.Current(fixtures.Day())
	names.Current(fixtures.Day().AddDate(1, 0, 0))
	beforeReload := counter.times
	fixtures.ReplaceIn(test, folders.Content, profileFile, authorName, "name: Mira Other")
	if err := store.Reload(); err != nil {
		test.Fatalf("Reload = %v, want no error", err)
	}
	reloaded := names.Current(fixtures.Day())
	if beforeReload != 1 || counter.times != 2 || reloaded != "Mira Other" {
		test.Errorf("worked out %d times before a reload and %d times after it, %q; want once, then twice, %q",
			beforeReload, counter.times, reloaded, "Mira Other")
	}
}

// workCounter counts how many times a value is worked out, and says when each value it works out lapses.
type workCounter struct {
	times  int
	lapses time.Time
}

// name works out the author's name, lapsing when the counter says, and counts it.
func (counter *workCounter) name(snapshot *Snapshot, _ time.Time) (string, time.Time) {
	counter.times++
	return snapshot.Content.Profile.Person.Name, counter.lapses
}

// contentName works out the author's name from the content alone, and counts it.
func (counter *workCounter) contentName(loaded *content.Snapshot) string {
	counter.times++
	return loaded.Profile.Person.Name
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
