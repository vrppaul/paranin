// Package fixtures is the tests' one way to the fixtures the packages' tests share (.claude/rules/tests.md), so no
// test writes a path that climbs out of its package. Tests alone import it (depguard's tests-alone-read-fixtures);
// a test of a package it imports reaches it from an external test package.
package fixtures

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/vrppaul/paranin/internal/assets"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// Where the fixtures live, from the module's root, the file the root is known by, and who may write a changed file of
// a copy and a folder made in it
const (
	contentFromRoot = "internal/content/testdata/content"
	buildFromRoot   = "internal/assets/testdata/build"
	cardsFromRoot   = "internal/assets/testdata/cards"
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
// deep and one of them twice, and the code of the widget the fixture post shows, ReadingPace, which the entry imports
// lazily; the file each was built into, with the stylesheets each imports and a font the entry's stylesheet loads; and
// the widget's props' schema, as the build writes it from web/tests/widgets/ReadingPace.props.ts.
func BuildFolder(test *testing.T) string {
	return filepath.Join(moduleRoot(test), buildFromRoot)
}

// CardsFolder is the folder of the cards' fixture: a manifest naming a card of each page of the fixture content that a
// preview shows, at the fixture's day and after, the résumé's a file of its own and every other page's one file they
// share; and a file the manifest names not.
func CardsFolder(test *testing.T) string {
	return filepath.Join(moduleRoot(test), cardsFromRoot)
}

// InModule is a path from the module's root made whole, for a test that reads a folder of the repository itself, as the
// faces the cards are set in, which a card's drawing is held to as a release draws it.
//
//wrapper:names a path of the repository a test reads, whatever folder go test runs it in
func InModule(test *testing.T, fromRoot string) string {
	return filepath.Join(moduleRoot(test), fromRoot)
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
	return routes.New(snapshot.Profile.Site.BaseURL, snapshot.Profile.Site.TagAuthority, snapshot.TranslatedInto())
}

// HeadOf is the head of a page a test builds by hand: the page and its title, nothing else, so a field the head gains
// reaches no test that has no use for it.
func HeadOf(page document.Page, title []document.Span) document.Head {
	return document.Head{
		Page: page, Number: 0, Language: "", Greeting: "", Title: title, Subtitle: nil, Place: "", Filters: nil, Lead: nil,
		Version: nil, Project: nil, Post: nil, Description: nil, Feeds: nil, Languages: nil,
	}
}

// TabPhrases are the phrases of a document a face that shows it tab by tab (document.TabsOf) must show, as the page
// and the TUI show it: the head's; on an opening with a brief, the brief's facts and the ways to reach the author; then
// every tab's blocks, each past the opening under its title.
func TabPhrases(composed document.Document) []string {
	tabs := document.TabsOf(composed)
	blocks := slices.Clone(tabs.Opening.Blocks)
	if briefed := tabs.Opening.Brief; briefed != nil {
		blocks = append(blocks, briefed.Brief, document.Facts{Kind: document.FactsWays, List: briefed.Ways})
	}
	for _, tab := range tabs.Rest {
		blocks = append(blocks, document.Section{Title: tab.Title, Name: "", Blocks: tab.Blocks})
	}
	return document.Phrases(document.Document{Head: composed.Head, Blocks: blocks})
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
//
//wrapper:names the fixture content every package's test reads
func LoadedContent(test *testing.T) *content.Snapshot {
	test.Helper()
	return LoadedFrom(test, ContentFolder(test))
}

// WithTwoProjectPages is the fixture content, loaded, with a second project of a page of its own after tidy-tool,
// though its slug sorts before it: tame-tool, of a name, a tagline and a language, so a project's page has another
// to lead to.
func WithTwoProjectPages(test *testing.T) *content.Snapshot {
	test.Helper()
	// the second project's slug, which is its name too
	const tameTool = "tame-tool"
	tame := content.Project{
		Slug: tameTool, Name: tameTool, Tagline: "Tames files.", Description: "", Language: "Go", Stack: nil, Repo: "",
		Install: nil, Release: nil, Licence: "", Page: true, Sections: nil, Details: nil,
	}
	snapshot := *LoadedContent(test)
	snapshot.Profile.Projects = append(slices.Clone(snapshot.Profile.Projects), tame)
	return &snapshot
}

// LoadedFrom is the content of a folder, loaded: the fixture's, or a copy of it a test changed.
func LoadedFrom(test *testing.T, folder string) *content.Snapshot {
	test.Helper()
	snapshot, err := content.Load(os.DirFS(folder))
	if err != nil {
		test.Fatalf("load the content of %s: %v", folder, err)
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

// LoadedCards is the cards' fixture, loaded.
func LoadedCards(test *testing.T) *assets.Cards {
	test.Helper()
	cards, err := assets.LoadCards(os.DirFS(CardsFolder(test)))
	if err != nil {
		test.Fatalf("load the cards' fixture: %v", err)
	}
	return cards
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
// says; the folders it stands in are made where they are missing.
func WriteIn(test *testing.T, folder, file, text string) {
	test.Helper()
	root := openFolder(test, folder)
	if err := root.MkdirAll(filepath.Dir(file), folderMode); err != nil {
		test.Fatalf("make the folders of %s: %v", file, err)
	}
	writeAt(test, root, file, text)
}

// How a test writes a moment, as a post's front matter writes when it was published: to the minute, with its zone
const momentLayout = "2006-01-02T15:04Z07:00"

// MomentOf is a moment as a test writes it, as front matter does, or the test's end where it does not read.
func MomentOf(test *testing.T, written string) time.Time {
	test.Helper()
	moment, err := time.Parse(momentLayout, written)
	if err != nil {
		test.Fatalf("read the moment %q: %v", written, err)
	}
	return moment
}

// KeepFolder leaves a folder as it is: the change of a test's case that changes nothing.
func KeepFolder(*testing.T, string) {}

// RemoveIn takes a file of a folder away, inside the folder, as ReplaceIn writes one.
func RemoveIn(test *testing.T, folder, file string) {
	test.Helper()
	if err := openFolder(test, folder).Remove(file); err != nil {
		test.Fatalf("remove %s: %v", file, err)
	}
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

// RecordedLog hands each record a logger writes to a channel, so a test reads what a server or the site logs by its
// values, not by its words. Their loggers add no attribute or group of their own, so neither is kept.
type RecordedLog struct {
	Records chan<- slog.Record
}

func (RecordedLog) Enabled(context.Context, slog.Level) bool { return true }

func (recorded RecordedLog) Handle(_ context.Context, record slog.Record) error {
	recorded.Records <- record.Clone()
	return nil
}

func (recorded RecordedLog) WithAttrs([]slog.Attr) slog.Handler { return recorded }

func (recorded RecordedLog) WithGroup(string) slog.Handler { return recorded }
