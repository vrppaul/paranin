package text

import (
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
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

// textPages are the pages a text face writes of a snapshot as its author reads it on the tests' day: the pages of the
// résumé and its projects, the posts, its draft among them, and the blog's list in each grouping.
func textPages(snapshot *content.Snapshot) []document.Document {
	site := compose.SiteOf(compose.Prepare(snapshot), compose.Audience{At: fixtures.Day(), Drafts: true})
	return slices.Concat(site.Documents(), site.Posts, site.Lists)
}

// The text of each page is what an agent or `curl` reads, so the text itself is the behaviour: each page of the
// fixture, a post's among them, in each format, is held to its golden file, named by the page's own path in that
// format.
func TestEveryPageReadsAsItsGoldenText(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	documents, site := textPages(snapshot), fixtures.Site(snapshot)
	for _, written := range formats() {
		for _, composed := range documents {
			path := routes.Path(composed.Head.Page, written.format)
			test.Run(path, func(test *testing.T) {
				golden.Match(test, path, written.write(composed, site))
			})
		}
	}
}

// A filtered list is written as the list it filters, its manual named as the list's, with the filter in the site's
// words beside the blog's name, and the posts it keeps; each filter, by tag and by series, in each format, is held to
// its golden file, named by the filter, since the list's path holds a query.
func TestAFilteredListReadsAsItsGoldenText(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	filtered := compose.SiteOf(compose.Prepare(snapshot), compose.Audience{At: fixtures.Day(), Drafts: false}).Filtered
	cases := []struct {
		page document.Page
		name string
	}{
		{page: document.TagFilter("reading"), name: "filtered/tag-reading"},
		{page: document.SeriesFilter("reading"), name: "filtered/series-reading"},
	}
	for _, testCase := range cases {
		for _, written := range formats() {
			file := testCase.name + path.Ext(routes.Path(document.PageOf(document.PageBlog), written.format))
			test.Run(file, func(test *testing.T) {
				golden.Match(test, file, written.write(fixtures.DocumentOf(test, filtered, testCase.page), fixtures.Site(snapshot)))
			})
		}
	}
}

// The index is what an agent reads first, llms.txt: markdown as llmstxt.org lays it out, a title, a quote of what
// the site is, and each page listed under its section. It is held to its golden file too.
func TestTheIndexReadsAsItsGoldenText(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	index := compose.SiteOf(compose.Prepare(snapshot), compose.Audience{At: fixtures.Day(), Drafts: false}).Index
	path := routes.Path(index.Head.Page, routes.FormatMarkdown)
	golden.Match(test, path, Markdown(index, fixtures.Site(snapshot)))
}

// Every phrase of a document reaches its text in each format: the facts a reader reads are the facts the
// document holds, found as each format writes them.
func TestEveryPhraseReachesEveryFormat(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	filtered := compose.SiteOf(compose.Prepare(snapshot), compose.Audience{At: fixtures.Day(), Drafts: false}).Filtered
	documents, site := slices.Concat(textPages(snapshot), filtered), fixtures.Site(snapshot)
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
	index := compose.SiteOf(compose.Prepare(snapshot), compose.Audience{At: fixtures.Day(), Drafts: false}).Index
	text := Markdown(index, fixtures.Site(snapshot))
	for _, phrase := range document.Phrases(index) {
		if !strings.Contains(text, phrase) {
			test.Errorf("the index's markdown %q holds no %q", text, phrase)
		}
	}
}

// A bar in a cell of a table under a header stays in its cell in markdown, escaped, since GFM would end the cell at
// it.
func TestABarInATablesCellStaysInItsCell(test *testing.T) {
	table := document.Table{
		Header: document.Row{Cells: []document.Cell{cellOf("Pipe"), cellOf("Reads")}},
		Rows:   []document.Row{{Cells: []document.Cell{cellOf("a | b"), cellOf("both")}}},
	}
	composed := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PagePost, Slug: "post", Tab: ""}, nil), Blocks: []document.Block{table}}
	if got, want := Markdown(composed, fixtures.Site(fixtures.LoadedContent(test))), `| a \| b | both |`; !strings.Contains(got, want) {
		test.Errorf("Markdown = %q, want %q", got, want)
	}
}

// cellOf is a cell of a table a test builds by hand: a text.
func cellOf(text string) document.Cell {
	return document.Cell{Kind: document.CellText, Spans: []document.Span{document.Text{Text: text}}}
}
