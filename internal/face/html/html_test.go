package html

import (
	"html"
	"os"
	"strings"
	"testing"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/golden"
	"github.com/vrppaul/paranin/internal/routes"
)

// The content package's fixture
const fixturePath = "../../content/testdata/content"

func TestMain(tests *testing.M) {
	golden.Main(tests)
}

// The markup of a page is what a browser, a screen reader and a search engine read: the language it names and the
// title it is called by, a heading's depth, a link's address, a month as a time, a result marked, a command as
// code, a fact as a term and its description, and the words between the parts of a line. Each page of the fixture
// is held to its golden file, named after the page's own address, with a line break between each two tags that
// meet, so the file reads in a diff.
func TestEveryPageReadsAsItsGoldenMarkup(test *testing.T) {
	snapshot := fixtureSnapshot(test)
	site := fixtureSite(snapshot)
	cases := []struct {
		page document.Page
		file string
	}{
		{page: document.Page{Kind: document.PageResume, Slug: ""}, file: "index.html"},
		{page: document.Page{Kind: document.PageProject, Slug: "tidy-tool"}, file: "p/tidy-tool.html"},
	}
	for _, testCase := range cases {
		test.Run(testCase.file, func(test *testing.T) {
			golden.Match(test, testCase.file, lined(Page(documentOf(test, snapshot, testCase.page), site)))
		})
	}
}

// Every phrase of a document reaches its page, as HTML writes text: the facts a browser shows are the facts the
// document holds.
func TestEveryPhraseReachesThePage(test *testing.T) {
	snapshot := fixtureSnapshot(test)
	site := fixtureSite(snapshot)
	for _, composed := range compose.Documents(snapshot) {
		page := Page(composed, site)
		for _, phrase := range document.Phrases(composed) {
			if !strings.Contains(page, html.EscapeString(phrase)) {
				test.Errorf("the page of %+v holds no %q", composed.Head.Page, phrase)
			}
		}
	}
}

// A link whose target has no address, as a hidden page has none, reads as its words, never as a link to nowhere.
func TestALinkWithNoAddressReadsAsItsWords(test *testing.T) {
	nowhere := document.Link{Target: document.Page{Kind: document.PageHealth, Slug: ""}, Spans: []document.Span{document.Text{Text: "the report"}}}
	composed := document.Document{
		Head:   document.Head{Page: document.Page{Kind: document.PageNone, Slug: ""}, Language: "", Title: nil, Subtitle: nil, Place: "", Lead: nil, Ways: nil},
		Blocks: []document.Block{document.Paragraph{Spans: []document.Span{nowhere}}},
	}
	if page, want := Page(composed, fixtureSite(fixtureSnapshot(test))), "<p>the report</p>"; !strings.Contains(page, want) {
		test.Errorf("the page %q holds no %q", page, want)
	}
}

// lined is a page with a line break between each two tags that meet, so its golden file reads in a diff; a space
// or a word between two tags stays where it is.
func lined(page string) string {
	return strings.ReplaceAll(page, "><", ">\n<") + "\n"
}

func fixtureSnapshot(test *testing.T) *content.Snapshot {
	snapshot, err := content.Load(os.DirFS(fixturePath))
	if err != nil {
		test.Fatalf("Load(%q) = %v, want no error", fixturePath, err)
	}
	return snapshot
}

// fixtureSite is the site of the fixture, at the address its profile gives.
func fixtureSite(snapshot *content.Snapshot) routes.Site {
	return routes.New(snapshot.Profile.Site.BaseURL)
}

// documentOf is the document compose makes of a page of the fixture.
func documentOf(test *testing.T, snapshot *content.Snapshot, page document.Page) document.Document {
	for _, composed := range compose.Documents(snapshot) {
		if composed.Head.Page == page {
			return composed
		}
	}
	test.Fatalf("compose made no page %+v", page)
	return document.Document{Head: document.Head{Page: page, Language: "", Title: nil, Subtitle: nil, Place: "", Lead: nil, Ways: nil}, Blocks: nil}
}
