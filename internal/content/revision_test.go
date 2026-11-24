package content

import (
	"bytes"
	"errors"
	"slices"
	"testing"
	"time"
)

// A revision that loads beside the fixture's post, published on the 10th and updated on the 12th, which a case writes
// changed: the version of the 11th, its body's line the file's 5th
const sampleRevision = `---
title: Reading slowly, the second time
description: A lead since rewritten.
---
A text since rewritten.
`

// Where a case writes the sample revision unless it names another place, and the fixture's own revision
const (
	sampleRevisionFile  = "posts/reading-slowly/revisions/2026-11-11.en.md"
	fixtureRevisionName = "2026-11-10"
	fixtureRevisions    = "posts/reading-slowly/revisions"
)

// revisionFields are the fields of a loaded revision a test reads.
type revisionFields struct {
	Name        string
	Day         string
	Title       string
	Description string
}

// A post's revisions load from its folder oldest first, each by the day its version came out, as its file's name
// writes it, with the title, the lead and the body that version had.
func TestAPostsRevisionsLoadOldestFirstWithTheirOwnText(test *testing.T) {
	snapshot, err := Load(writtenFile(fixtureFolder(test), sampleRevisionFile, sampleRevision))
	if err != nil {
		test.Fatalf("Load = %v, want no error", err)
	}
	post := postNamed(test, snapshot, fixturePostSlug)
	got := []revisionFields{}
	for _, revision := range post.Revisions {
		got = append(got, revisionFields{
			Name: revision.Day.Written(), Day: revision.Day.Start().Format(time.DateOnly), Title: revision.Title, Description: revision.Description,
		})
	}
	want := []revisionFields{
		{Name: fixtureRevisionName, Day: fixtureRevisionName, Title: "Slow reading", Description: "Why a page read twice is read once, first said."},
		{Name: "2026-11-11", Day: "2026-11-11", Title: "Reading slowly, the second time", Description: "A lead since rewritten."},
	}
	if !slices.Equal(got, want) {
		test.Errorf("the revisions of %s = %+v, want %+v", fixturePostSlug, got, want)
	}
	if last := post.Revisions[len(post.Revisions)-1]; !bytes.Equal(last.Body.Source, []byte("A text since rewritten.\n")) {
		test.Errorf("the newest revision's body = %q, want the text after its front matter", last.Body.Source)
	}
}

// Every fault of a revision stands at its file, and at its field or its line where it names one; a revision is held to
// what a post's text is held to, and to a day from the post's first version's on and before its current version's.
func TestEveryFaultOfARevisionStandsAtItsPlace(test *testing.T) {
	cases := []struct {
		name        string
		file        string
		original    string
		replacement string
		want        faultPlace
	}{
		{name: "raw HTML", file: sampleRevisionFile, original: "A text since rewritten.", replacement: "<b>A text</b>.", want: faultPlace{file: sampleRevisionFile, field: "", line: 5}},
		{name: "a link to no address", file: sampleRevisionFile, original: "A text since rewritten.", replacement: "A [text](p/tidy-tool).", want: faultPlace{file: sampleRevisionFile, field: "", line: 5}},
		{name: "a picture of no file", file: sampleRevisionFile, original: "A text since rewritten.", replacement: "![A page](missing.png)", want: faultPlace{file: sampleRevisionFile, field: "", line: 5}},
		{name: "no title", file: sampleRevisionFile, original: "title: Reading slowly, the second time\n", replacement: "", want: faultPlace{file: sampleRevisionFile, field: "title", line: 0}},
		{name: "no description", file: sampleRevisionFile, original: "description: A lead since rewritten.\n", replacement: "", want: faultPlace{file: sampleRevisionFile, field: "description", line: 0}},
		{name: "a field of the post's", file: sampleRevisionFile, original: "title:", replacement: "id: 1\ntitle:", want: faultPlace{file: sampleRevisionFile, field: "", line: 0}},
		{name: "an escape in a field", file: sampleRevisionFile, original: "title: Reading slowly, the second time", replacement: `title: "Reading \e[2J"`, want: faultPlace{file: sampleRevisionFile, field: "title", line: 0}},
		{name: "a day before the post was published", file: "posts/reading-slowly/revisions/2026-10-01.en.md", original: "title:", replacement: "title:", want: faultPlace{file: "posts/reading-slowly/revisions/2026-10-01.en.md", field: "", line: 0}},
		{name: "the current version's day", file: "posts/reading-slowly/revisions/2026-11-12.en.md", original: "title:", replacement: "title:", want: faultPlace{file: "posts/reading-slowly/revisions/2026-11-12.en.md", field: "", line: 0}},
		{name: "a day after the current version's", file: "posts/reading-slowly/revisions/2026-11-13.en.md", original: "title:", replacement: "title:", want: faultPlace{file: "posts/reading-slowly/revisions/2026-11-13.en.md", field: "", line: 0}},
		{name: "a name of no day", file: "posts/reading-slowly/revisions/first.en.md", original: "title:", replacement: "title:", want: faultPlace{file: "posts/reading-slowly/revisions/first.en.md", field: "", line: 0}},
		{name: "a day of no calendar", file: "posts/reading-slowly/revisions/2026-11-31.en.md", original: "title:", replacement: "title:", want: faultPlace{file: "posts/reading-slowly/revisions/2026-11-31.en.md", field: "", line: 0}},
		{name: "a translation's revision", file: "posts/reading-slowly/revisions/2026-11-11.ru.md", original: "title:", replacement: "title:", want: faultPlace{file: "posts/reading-slowly/revisions/2026-11-11.ru.md", field: "", line: 0}},
		{name: "a folder among the revisions", file: "posts/reading-slowly/revisions/old/2026-11-11.en.md", original: "title:", replacement: "title:", want: faultPlace{file: "posts/reading-slowly/revisions/old", field: "", line: 0}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			written := sampleWith(test, sampleRevision, testCase.original, testCase.replacement)
			faults := faultsOf(test, writtenFile(fixtureFolder(test), testCase.file, written))
			if !slices.ContainsFunc(faults, func(fault Fault) bool { return placeOf(fault) == testCase.want }) {
				test.Errorf("faults with %q at %s = %q, want one at %+v", testCase.replacement, testCase.file, faults, testCase.want)
			}
		})
	}
}

// A post's revisions are a folder of its own: a file in its place is a fault of the post's folder.
func TestARevisionsFileInPlaceOfItsFolderIsAFault(test *testing.T) {
	faults := faultsOf(test, writtenFile(fixtureFolder(test), "posts/notes-on-tools/revisions", sampleRevision))
	want := faultPlace{file: "posts/notes-on-tools/revisions", field: "", line: 0}
	if !slices.ContainsFunc(faults, func(fault Fault) bool { return placeOf(fault) == want }) {
		test.Errorf("faults with a file of revisions = %q, want one at %+v", faults, want)
	}
}

// A picture a revision alone shows is a part of its post, served with it: the post's folder keeps it for the older
// version that shows it.
func TestAPictureARevisionAloneShowsIsAPartOfItsPost(test *testing.T) {
	folder := writtenFile(fixtureFolder(test), sampleRevisionFile, sampleWith(test, sampleRevision, "A text since rewritten.", "![A page](earlier.png)"))
	folder = writtenFile(folder, "posts/reading-slowly/earlier.png", string(fixturePicture(test)))
	snapshot, err := Load(folder)
	if err != nil {
		test.Fatalf("Load with a picture a revision alone shows = %v, want no error", err)
	}
	if _, isShown := postNamed(test, snapshot, fixturePostSlug).Pictures["earlier.png"]; !isShown {
		test.Errorf("the pictures of %s hold no earlier.png, which its revision shows", fixturePostSlug)
	}
}

// A post's version out is dated by the day its text last changed: the day it was updated, where it says one, or else
// the day it was published, in its own zone.
func TestAPostsVersionIsDatedByItsLastChange(test *testing.T) {
	cases := []struct {
		name     string
		original string
		dates    string
		want     string
	}{
		{name: "never updated", original: "published: 2026-11-10T09:00+01:00", dates: "published: 2026-11-10T23:30-02:00", want: "2026-11-10"},
		{name: "updated", original: "published: 2026-11-10T09:00+01:00", dates: "published: 2026-11-10T09:00+01:00\nupdated: 2026-11-14", want: "2026-11-14"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			snapshot, err := Load(writtenFile(fixtureFolder(test), samplePostFile, sampleWith(test, samplePost, testCase.original, testCase.dates)))
			if err != nil {
				test.Fatalf("Load = %v, want no error", err)
			}
			day := postNamed(test, snapshot, samplePostSlug).VersionDay()
			if got := []string{day.Written(), day.Start().Format(time.DateOnly)}; !slices.Equal(got, []string{testCase.want, testCase.want}) {
				test.Errorf("the version's day with %q = %q, want %q", testCase.dates, got, testCase.want)
			}
		})
	}
}

// The moments the tests take a revision at: the day after the fixture post's current version came out; the next day
// in a zone further east, still late on that version's own day in the post's zone; and the day after a test's rewrite,
// updated on the 20th
const (
	dayAfterTheVersion = "2026-11-13T09:00:00+01:00"
	lateOnTheVersion   = "2026-11-13T00:30:00+02:00"
	dayAfterTheRewrite = "2026-11-21T09:00:00+01:00"
)

// momentOf is a moment as RFC 3339 writes it.
func momentOf(test *testing.T, written string) time.Time {
	test.Helper()
	moment, err := time.Parse(time.RFC3339, written)
	if err != nil {
		test.Fatalf("read the moment %q: %v", written, err)
	}
	return moment
}

// A revision taken of a post holds its version's title, lead and body, under the day the version came out, and loads
// as that version once the post is dated by a later update; a post whose text is its newest revision's already has
// no version to take, though one whose title or lead alone changed has.
func TestARevisionTakenHoldsThePostsVersion(test *testing.T) {
	folder := fixtureFolder(test)
	snapshot, err := Load(folder)
	if err != nil {
		test.Fatalf("Load = %v, want no error", err)
	}
	taken, err := RevisionOf(postNamed(test, snapshot, fixturePostSlug), momentOf(test, dayAfterTheVersion))
	if err != nil {
		test.Fatalf("RevisionOf(%s) = %v, want a revision", fixturePostSlug, err)
	}
	if want := "posts/reading-slowly/revisions/2026-11-12.en.md"; taken.Path != want {
		test.Errorf("the revision taken is written at %q, want %q", taken.Path, want)
	}
	folder = changedFile(test, writtenFile(folder, taken.Path, string(taken.Text)), "posts/reading-slowly/post.en.md", "updated: 2026-11-12", "updated: 2026-11-20")
	revised, err := Load(folder)
	if err != nil {
		test.Fatalf("Load with the revision taken = %v, want no error", err)
	}
	post := postNamed(test, revised, fixturePostSlug)
	newest := post.Revisions[len(post.Revisions)-1]
	got := revisionFields{Name: newest.Day.Written(), Day: newest.Day.Start().Format(time.DateOnly), Title: newest.Title, Description: newest.Description}
	want := revisionFields{Name: "2026-11-12", Day: "2026-11-12", Title: post.Title, Description: post.Description}
	if got != want || !bytes.Equal(newest.Body.Source, post.Body.Source) {
		test.Errorf("the revision taken loads as %+v with the body %q, want %+v with the post's", got, newest.Body.Source, want)
	}
	cases := []struct {
		name  string
		title string
		lead  string
		want  error
	}{
		{name: "no change", title: post.Title, lead: post.Description, want: errNoChange},
		{name: "the title changed", title: "Reading slowly, again", lead: post.Description, want: nil},
		{name: "the lead changed", title: post.Title, lead: "Why a page read once is read twice.", want: nil},
	}
	later := momentOf(test, dayAfterTheRewrite)
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			changed := post
			changed.Title, changed.Description = testCase.title, testCase.lead
			if _, err := RevisionOf(changed, later); !errors.Is(err, testCase.want) {
				test.Errorf("RevisionOf a post titled %q, led %q = %v, want %v", testCase.title, testCase.lead, err, testCase.want)
			}
		})
	}
}

// A post's version is taken the day after it came out at the soonest, as the post's own zone counts the day: one
// version a day, and a change on the version's own day an edit of it.
func TestARevisionIsTakenTheDayAfterItsVersionAtTheSoonest(test *testing.T) {
	snapshot, err := Load(fixtureFolder(test))
	if err != nil {
		test.Fatalf("Load = %v, want no error", err)
	}
	cases := []struct {
		name   string
		moment string
		want   error
	}{
		{name: "the day after", moment: dayAfterTheVersion, want: nil},
		{name: "the version's own day, late in the post's zone", moment: lateOnTheVersion, want: errSameDay},
		{name: "before the version", moment: "2026-11-11T09:00:00+01:00", want: errSameDay},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if _, err := RevisionOf(postNamed(test, snapshot, fixturePostSlug), momentOf(test, testCase.moment)); !errors.Is(err, testCase.want) {
				test.Errorf("RevisionOf at %s = %v, want %v", testCase.moment, err, testCase.want)
			}
		})
	}
}
