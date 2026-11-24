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

// format is one format this package writes, the function that writes it in a language's manual words, and how a
// reader finds a phrase in it.
type format struct {
	format routes.Format
	write  func(document.Document, routes.Site, compose.ManualWords) string
	holds  func(text, phrase string) bool
}

func formats() []format {
	return []format{
		{
			format: routes.FormatMarkdown, holds: strings.Contains,
			write: func(composed document.Document, site routes.Site, _ compose.ManualWords) string {
				return Markdown(composed, site)
			},
		},
		{format: routes.FormatManual, write: Manual, holds: manualHolds},
	}
}

// framedPage is a page a text face writes, and the manual's own words in its language.
type framedPage struct {
	document document.Document
	manual   compose.ManualWords
}

// ownManual is the manual's own words in the site's own language, as the fixture writes them.
func ownManual(test *testing.T) compose.ManualWords {
	test.Helper()
	return compose.SiteOf(compose.Prepare(fixtures.LoadedContent(test)), compose.Audience{At: fixtures.Day(), Drafts: false}).Own.Frame.Manual
}

// manualHolds tells a manual that holds a phrase: its lines may break inside the phrase, and a section's title
// stands in capitals.
func manualHolds(text, phrase string) bool {
	words, wanted := strings.Join(strings.Fields(text), " "), strings.Join(strings.Fields(phrase), " ")
	return strings.Contains(words, wanted) || strings.Contains(words, strings.ToUpper(wanted))
}

// textPages are the pages a text face writes of a snapshot as its author reads it on the tests' day: the pages of the
// résumé and its projects, the posts, its draft among them, their older versions, and the blog's list in each
// grouping, in every language, each with the manual's own words in its language.
func textPages(snapshot *content.Snapshot) []framedPage {
	site := compose.SiteOf(compose.Prepare(snapshot), compose.Audience{At: fixtures.Day(), Drafts: true})
	pages := framedPages(compose.DocumentsOf(site.Pages), site.Own.Frame.Manual)
	for _, blog := range site.Blogs() {
		pages = append(pages, framedPages(slices.Concat(compose.DocumentsOf(blog.Posts), blog.Versions, blog.Lists), blog.Frame.Manual)...)
	}
	return pages
}

// framedPages are documents, each with the manual's own words of their language.
func framedPages(documents []document.Document, manual compose.ManualWords) []framedPage {
	pages := make([]framedPage, 0, len(documents))
	for _, composed := range documents {
		pages = append(pages, framedPage{document: composed, manual: manual})
	}
	return pages
}

// The text of each page is what an agent or `curl` reads, so the text itself is the behaviour: each page of the
// fixture, a post's among them, in each format, is held to its golden file, named by the page's own path in that
// format.
func TestEveryPageReadsAsItsGoldenText(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	pages, site := textPages(snapshot), fixtures.Site(snapshot)
	for _, written := range formats() {
		for _, page := range pages {
			path := routes.Path(page.document.Head.Page, written.format)
			test.Run(path, func(test *testing.T) {
				golden.Match(test, path, written.write(page.document, site, page.manual))
			})
		}
	}
}

// A filtered list is written as the list it filters, its manual named as the list's, with the filter in the site's
// words beside the blog's name, and the posts it keeps; in each format, it is held to its golden file, named by the
// filter, since the list's path holds a query.
func TestAFilteredListReadsAsItsGoldenText(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	filtered := compose.SiteOf(compose.Prepare(snapshot), compose.Audience{At: fixtures.Day(), Drafts: false}).Own.Filtered
	cases := []struct {
		page document.Page
		name string
	}{
		{page: document.TagFilter("reading"), name: "filtered/tag-reading"},
	}
	for _, testCase := range cases {
		for _, written := range formats() {
			file := testCase.name + path.Ext(routes.Path(document.PageOf(document.PageBlog), written.format))
			test.Run(file, func(test *testing.T) {
				golden.Match(test, file, written.write(fixtures.DocumentOf(test, filtered, testCase.page), fixtures.Site(snapshot), ownManual(test)))
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
	filtered := compose.SiteOf(compose.Prepare(snapshot), compose.Audience{At: fixtures.Day(), Drafts: false}).Own.Filtered
	pages, site := slices.Concat(textPages(snapshot), framedPages(filtered, ownManual(test))), fixtures.Site(snapshot)
	for _, written := range formats() {
		for _, page := range pages {
			text := written.write(page.document, site, page.manual)
			for _, phrase := range document.Phrases(page.document) {
				if !written.holds(text, phrase) {
					test.Errorf("the %s of %+v holds no %q", routes.Path(page.document.Head.Page, written.format), page.document.Head.Page, phrase)
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
