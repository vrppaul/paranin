package text

import (
	"strings"
	"testing"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/golden"
	"github.com/vrppaul/paranin/internal/routes"
)

func TestMain(tests *testing.M) {
	golden.Main(tests)
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
	snapshot := fixtures.LoadedContent(test)
	documents, site := compose.Documents(snapshot), fixtures.Site(snapshot)
	for _, written := range formats() {
		for _, composed := range documents {
			path := routes.Path(composed.Head.Page, written.format)
			test.Run(path, func(test *testing.T) {
				golden.Match(test, path, written.write(composed, site))
			})
		}
	}
}

// The index is what an agent reads first, llms.txt: markdown as llmstxt.org lays it out, a title, a quote of what
// the site is, and each page listed under its section. It is held to its golden file too.
func TestTheIndexReadsAsItsGoldenText(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	index := compose.Index(snapshot)
	path := routes.Path(index.Head.Page, routes.FormatMarkdown)
	golden.Match(test, path, Markdown(index, fixtures.Site(snapshot)))
}

// Every phrase of a document reaches its text in each format: the facts a reader reads are the facts the
// document holds, found as each format writes them.
func TestEveryPhraseReachesEveryFormat(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	documents, site := compose.Documents(snapshot), fixtures.Site(snapshot)
	for _, written := range formats() {
		for _, composed := range documents {
			text := written.write(composed, site)
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
	snapshot := fixtures.LoadedContent(test)
	index := compose.Index(snapshot)
	text := Markdown(index, fixtures.Site(snapshot))
	for _, phrase := range document.Phrases(index) {
		if !strings.Contains(text, phrase) {
			test.Errorf("the index's markdown %q holds no %q", text, phrase)
		}
	}
}
