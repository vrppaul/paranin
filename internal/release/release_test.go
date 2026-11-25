package release

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vrppaul/paranin/internal/assets"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// The files of the fixtures a test changes, from each one's root, and the author's name the profile holds
const (
	profileFile  = "profile.yaml"
	manifestFile = "manifest.json"
	schemasFile  = "props.schema.json"
	authorName   = "name: Mira Example"
	cardFile     = "page-fixture.png"
)

// The fixture post's picture, by its name and its file in the content's folder, and who may read the copy a case
// writes outside it
const (
	pictureName = "margin.png"
	pictureFile = "posts/reading-slowly/" + pictureName
	pictureMode = 0o600
)

// A site cannot open on a release with a fault in any part, and says each fault: the content's, among them those the
// check it is handed finds, the build's and the cards'.
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
		{name: "a card missing", change: cardLeftOut, check: passes, want: cardFile},
		{name: "a picture linking out of the content's folder", change: pictureLinkedOut, check: passes, want: pictureName},
		{name: "a post showing a widget the build holds no code for", change: widgetLeftOut, check: passes, want: widgetPlace},
		{name: "a build holding the component in another folder", change: componentMoved, check: passes, want: widgetPlace},
		{name: "an older version showing a widget the build holds no code for", change: revisionWidget, check: passes, want: revisionWidgetPlace},
		{name: "a translation showing a widget the build holds no code for", change: translationWidget, check: passes, want: translationWidgetPlace},
		{name: "a widget's prop of a type its schema refuses", change: propsWritten("pace: 4"), check: passes, want: widgetPlace},
		{name: "a widget's prop of a value its schema does not list", change: propsWritten("pace: brisk"), check: passes, want: widgetPlace},
		{name: "a widget's prop its schema does not name", change: propsWritten("pace: slow\ntempo: slow"), check: passes, want: widgetPlace},
		{name: "a widget's prop misspelt", change: propsWritten("Pace: slow"), check: passes, want: widgetPlace},
		{name: "a widget's prop its schema requires left out", change: propsWritten(""), check: passes, want: widgetPlace},
		{name: "an older version's widget's props its schema refuses", change: revisionProps, check: passes, want: revisionWidgetPlace},
		{name: "a translation's widget's props its schema refuses", change: translationProps, check: passes, want: translationWidgetPlace},
		{name: "a widget's prop of a format its schema refuses", change: paceOfAddresses, check: passes, want: widgetPlace},
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

// fixtureCopies are copies of the three fixtures on disk, which a test may change.
func fixtureCopies(test *testing.T) Folders {
	return Folders{
		Content: fixtures.Copy(test, fixtures.ContentFolder(test)), Build: fixtures.Copy(test, fixtures.BuildFolder(test)),
		Cards: fixtures.Copy(test, fixtures.CardsFolder(test)),
	}
}

// What the check a case hands in says where it refuses the content
const refusal = "the check refuses it"

// passes is a check that finds no fault.
func passes(*content.Snapshot, *assets.Cards) error {
	return nil
}

// refuses is a check that finds a fault in any content and cards.
//
//wrapper:a case of the table names the check it hands the release
func refuses(*content.Snapshot, *assets.Cards) error {
	return errors.New(refusal)
}

// keepFolders leaves the fixtures as they are.
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

// Where the fixture post shows its widget first, at its file and the line of the block's fence, and the widget's key in
// the build's manifest
const (
	postFile    = "posts/reading-slowly/post.en.md"
	widgetPlace = postFile + ":32:"
	widgetKey   = `"src/widgets/ReadingPace.svelte": {`
)

// A widget the build holds no code for; the fixture post's older version and its translation, a line of each that a
// case writes the widget over, and where each then shows it
const (
	strangeWidget          = "```widget OtherPace\nalt: A pace of no code.\n```"
	revisionFile           = "posts/reading-slowly/revisions/2026-11-10.en.md"
	revisionLine           = "The margins, as [the tool](/p/tidy-tool) keeps them."
	revisionWidgetPlace    = revisionFile + ":9:"
	translationFile        = "posts/reading-slowly/post.ru.md"
	translationLine        = "Поля и заметки на них."
	translationWidgetPlace = translationFile + ":12:"
)

// widgetLeftOut leaves the build with the code and the props' schema of another widget than the one the fixture post
// shows.
func widgetLeftOut(test *testing.T, folders Folders) {
	fixtures.ReplaceIn(test, folders.Build, manifestFile, widgetKey, `"src/widgets/OtherPace.svelte": {`)
	fixtures.ReplaceIn(test, folders.Build, schemasFile, `"ReadingPace": {`, `"OtherPace": {`)
}

// componentMoved leaves the build with the fixture post's widget's component in a folder of web/ that holds no widget.
//
//wrapper:a case of the table names the fault it writes
func componentMoved(test *testing.T, folders Folders) {
	fixtures.ReplaceIn(test, folders.Build, manifestFile, widgetKey, `"src/ui/ReadingPace.svelte": {`)
}

// revisionWidget writes a widget the build holds no code for into the fixture post's older version.
//
//wrapper:a case of the table names the fault it writes
func revisionWidget(test *testing.T, folders Folders) {
	fixtures.ReplaceIn(test, folders.Content, revisionFile, revisionLine, strangeWidget)
}

// translationWidget writes a widget the build holds no code for into the fixture post's translation.
//
//wrapper:a case of the table names the fault it writes
func translationWidget(test *testing.T, folders Folders) {
	fixtures.ReplaceIn(test, folders.Content, translationFile, translationLine, strangeWidget)
}

// The fixture post's first widget's caption and its own props, as written, and a widget its schema refuses, of a pace
// of the wrong type
const (
	firstCaption  = "caption: Slow the pace, and the margin fills.\n"
	writtenProps  = firstCaption + "pace: slow"
	refusedWidget = "```widget ReadingPace\nalt: A pace of the wrong type.\npace: 4\n```"
)

// propsWritten writes the fixture post's first widget's own props, its alt text and caption kept.
func propsWritten(props string) func(test *testing.T, folders Folders) {
	return func(test *testing.T, folders Folders) {
		fixtures.ReplaceIn(test, folders.Content, postFile, writtenProps, firstCaption+props)
	}
}

// revisionProps writes a widget whose props its schema refuses into the fixture post's older version.
//
//wrapper:a case of the table names the fault it writes
func revisionProps(test *testing.T, folders Folders) {
	fixtures.ReplaceIn(test, folders.Content, revisionFile, revisionLine, refusedWidget)
}

// translationProps writes a widget whose props its schema refuses into the fixture post's translation.
//
//wrapper:a case of the table names the fault it writes
func translationProps(test *testing.T, folders Folders) {
	fixtures.ReplaceIn(test, folders.Content, translationFile, translationLine, refusedWidget)
}

// paceOfAddresses holds the fixture widget's pace to a mail address's format in the build's schema, which the paces the
// fixture post writes, listed by the schema still, are not.
//
//wrapper:a case of the table names the fault it writes
func paceOfAddresses(test *testing.T, folders Folders) {
	fixtures.ReplaceIn(test, folders.Build, schemasFile, `"type": "string"`, `"type": "string", "format": "email"`)
}

// faultyBuild leaves a manifest that does not read.
//
//wrapper:a case of the table names the fault it writes
func faultyBuild(test *testing.T, folders Folders) {
	fixtures.ReplaceIn(test, folders.Build, manifestFile, "{", "[")
}

// cardLeftOut takes a card the cards' manifest names away.
//
//wrapper:a case of the table names the fault it writes
func cardLeftOut(test *testing.T, folders Folders) {
	fixtures.RemoveIn(test, folders.Cards, cardFile)
}
