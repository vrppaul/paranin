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
	for _, written := range formats() {
		for _, composed := range fixtureDocuments(test) {
			path := routes.Path(composed.Head.Page, written.format)
			test.Run(path, func(test *testing.T) {
				golden := filepath.Join(goldenFolder, filepath.FromSlash(path))
				got := written.write(composed, routes.New(fixtureBase))
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
			})
		}
	}
}

// Every phrase of a document reaches its text in each format: the facts a reader reads are the facts the
// document holds, found as each format writes them.
func TestEveryPhraseReachesEveryFormat(test *testing.T) {
	for _, written := range formats() {
		for _, composed := range fixtureDocuments(test) {
			text := written.write(composed, routes.New(fixtureBase))
			for _, phrase := range document.Phrases(composed) {
				if !written.holds(text, phrase) {
					test.Errorf("the %s of %+v holds no %q", routes.Path(composed.Head.Page, written.format), composed.Head.Page, phrase)
				}
			}
		}
	}
}

func fixtureDocuments(test *testing.T) []document.Document {
	snapshot, err := content.Load(os.DirFS(fixturePath))
	if err != nil {
		test.Fatalf("Load(%q) = %v, want no error", fixturePath, err)
	}
	return compose.Documents(snapshot)
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
