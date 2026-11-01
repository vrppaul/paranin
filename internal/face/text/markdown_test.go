package text

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// The content package's fixture, the address it is served at, and where the golden files live
const (
	fixturePath  = "../../content/testdata/content"
	fixtureBase  = "https://example.test"
	goldenFolder = "testdata"
	updateFlag   = "update"
)

func TestMain(tests *testing.M) {
	flag.Bool(updateFlag, false, "rewrite the golden files from what the face writes now")
	flag.Parse()
	os.Exit(tests.Run())
}

// format is one format this package writes, the function that writes it, and how a reader finds a phrase in it.
type format struct {
	format routes.Format
	write  func(document.Document, routes.Site) string
	holds  func(text, phrase string) bool
}

func formats() []format {
	return []format{
		{format: routes.FormatMarkdown, write: Markdown, holds: strings.Contains},
		{format: routes.FormatManual, write: Manual, holds: manualHolds},
	}
}

// manualHolds tells a manual that holds a phrase: its lines may break inside the phrase, and a section's title
// stands in capitals.
func manualHolds(text, phrase string) bool {
	words, wanted := strings.Join(strings.Fields(text), " "), strings.Join(strings.Fields(phrase), " ")
	return strings.Contains(words, wanted) || strings.Contains(words, strings.ToUpper(wanted))
}

// The text of each page is what an agent or `curl` reads, so the text itself is the behaviour: each page of the
// fixture, in each format, is held to its golden file, named by the page's own path in that format.
func TestEveryPageReadsAsItsGoldenText(test *testing.T) {
	documents := compose.Documents(fixtureSnapshot(test))
	for _, written := range formats() {
		for _, composed := range documents {
			path := routes.Path(composed.Head.Page, written.format)
			test.Run(path, func(test *testing.T) {
				matchesGolden(test, path, written.write(composed, routes.New(fixtureBase)))
			})
		}
	}
}

// The index is what an agent reads first, llms.txt: markdown as llmstxt.org lays it out, a title, a quote of what
// the site is, and each page listed under its section. It is held to its golden file too.
func TestTheIndexReadsAsItsGoldenText(test *testing.T) {
	index := compose.Index(fixtureSnapshot(test))
	path := routes.Path(index.Head.Page, routes.FormatMarkdown)
	matchesGolden(test, path, Markdown(index, routes.New(fixtureBase)))
}

// Every phrase of a document reaches its text in each format: the facts a reader reads are the facts the
// document holds, found as each format writes them.
func TestEveryPhraseReachesEveryFormat(test *testing.T) {
	documents := compose.Documents(fixtureSnapshot(test))
	for _, written := range formats() {
		for _, composed := range documents {
			text := written.write(composed, routes.New(fixtureBase))
			for _, phrase := range document.Phrases(composed) {
				if !written.holds(text, phrase) {
					test.Errorf("the %s of %+v holds no %q", routes.Path(composed.Head.Page, written.format), composed.Head.Page, phrase)
				}
			}
		}
	}
}

// Every phrase of the index reaches llms.txt, the markdown it is served as: an agent reads every page the index
// lists, by its title and its line.
func TestEveryPhraseOfTheIndexReachesItsMarkdown(test *testing.T) {
	index := compose.Index(fixtureSnapshot(test))
	text := Markdown(index, routes.New(fixtureBase))
	for _, phrase := range document.Phrases(index) {
		if !strings.Contains(text, phrase) {
			test.Errorf("the index's markdown %q holds no %q", text, phrase)
		}
	}
}

func fixtureSnapshot(test *testing.T) *content.Snapshot {
	snapshot, err := content.Load(os.DirFS(fixturePath))
	if err != nil {
		test.Fatalf("Load(%q) = %v, want no error", fixturePath, err)
	}
	return snapshot
}

// matchesGolden holds a text to the golden file of its path, after rewriting the file where the run asks to.
func matchesGolden(test *testing.T, path, got string) {
	golden := filepath.Join(goldenFolder, filepath.FromSlash(path))
	if updating() {
		writeGolden(test, golden, got)
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		test.Fatalf("read %s: %v", golden, err)
	}
	if got != string(want) {
		test.Errorf("%s =\n%s\nwant\n%s", path, got, want)
	}
}

func updating() bool {
	return flag.Lookup(updateFlag).Value.String() == "true"
}

func writeGolden(test *testing.T, golden, text string) {
	if err := os.MkdirAll(filepath.Dir(golden), 0o750); err != nil {
		test.Fatalf("make the folder of %s: %v", golden, err)
	}
	if err := os.WriteFile(golden, []byte(text), 0o600); err != nil {
		test.Fatalf("write %s: %v", golden, err)
	}
}
