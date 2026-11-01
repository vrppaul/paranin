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

// The markdown of each page is what an agent or `curl` reads, so the text itself is the behaviour: each page of
// the fixture is held to its golden file, named by the page's own path.
func TestEveryPageReadsAsItsGoldenMarkdown(test *testing.T) {
	for _, composed := range fixtureDocuments(test) {
		path := routes.Path(composed.Head.Page, routes.FormatMarkdown)
		test.Run(path, func(test *testing.T) {
			golden := filepath.Join(goldenFolder, filepath.FromSlash(path))
			got := Markdown(composed, routes.New(fixtureBase))
			if updating() {
				writeGolden(test, golden, got)
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				test.Fatalf("read %s: %v", golden, err)
			}
			if got != string(want) {
				test.Errorf("Markdown of %s =\n%s\nwant\n%s", path, got, want)
			}
		})
	}
}

// Every phrase of a document reaches its markdown: the facts an agent reads are the facts the document holds.
func TestEveryPhraseReachesTheMarkdown(test *testing.T) {
	for _, composed := range fixtureDocuments(test) {
		markdown := Markdown(composed, routes.New(fixtureBase))
		for _, phrase := range document.Phrases(composed) {
			if !strings.Contains(markdown, phrase) {
				test.Errorf("the markdown of %+v holds no %q", composed.Head.Page, phrase)
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
