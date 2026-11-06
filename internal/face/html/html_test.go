package html

import (
	"html"
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

// The markup of a page is what a browser, a screen reader and a search engine read: the language it names and the
// title it is called by, a heading's depth, a link's address, a month as a time, a result marked, a command as
// code, a fact as a term and its description, and the words between the parts of a line. Each page of the fixture
// is held to its golden file, named after the page's own address, with a line break between each two tags that
// meet, so the file reads in a diff.
func TestEveryPageReadsAsItsGoldenMarkup(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	cases := []struct {
		page document.Page
		file string
	}{
		{page: document.Page{Kind: document.PageResume, Slug: ""}, file: "index.html"},
		{page: document.Page{Kind: document.PageProject, Slug: "tidy-tool"}, file: "p/tidy-tool.html"},
	}
	for _, testCase := range cases {
		test.Run(testCase.file, func(test *testing.T) {
			composed := composedOf(test, snapshot, testCase.page)
			golden.Match(test, testCase.file, lined(fixtureWriter(test).Page(composed.Document, composed.Record)))
		})
	}
}

// Every phrase of a document reaches its page, as HTML writes text: the facts a browser shows are the facts the
// document holds.
func TestEveryPhraseReachesThePage(test *testing.T) {
	for _, composed := range compose.Pages(fixtures.LoadedContent(test)) {
		page := fixtureWriter(test).Page(composed.Document, composed.Record)
		for _, phrase := range document.Phrases(composed.Document) {
			if !strings.Contains(page, html.EscapeString(phrase)) {
				test.Errorf("the page of %+v holds no %q", composed.Document.Head.Page, phrase)
			}
		}
	}
}

// A link whose target has no address, as a hidden page has none, reads as its words, never as a link to nowhere.
func TestALinkWithNoAddressReadsAsItsWords(test *testing.T) {
	nowhere := document.Link{Target: document.Page{Kind: document.PageHealth, Slug: ""}, Spans: []document.Span{document.Text{Text: "the report"}}}
	composed := document.Document{
		Head:   fixtures.HeadOf(document.Page{Kind: document.PageNone, Slug: ""}, nil),
		Blocks: []document.Block{document.Paragraph{Spans: []document.Span{nowhere}}},
	}
	if page, want := fixtureWriter(test).Page(composed, blankPerson()), "<p>the report</p>"; !strings.Contains(page, want) {
		test.Errorf("the page %q holds no %q", page, want)
	}
}

// A page whose content gives no description names none, so a search engine or a preview writes its own rather
// than an empty one.
func TestAPageWithNoDescriptionNamesNone(test *testing.T) {
	composed := document.Document{
		Head:   fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: ""}, nil),
		Blocks: nil,
	}
	page := fixtureWriter(test).Page(composed, blankPerson())
	for _, unwanted := range []string{`name="description"`, `property="og:description"`} {
		if strings.Contains(page, unwanted) {
			test.Errorf("the page %q holds %q", page, unwanted)
		}
	}
}

// The page of an address the site has no page at opens its head as every page does, names its language, and its
// status in its title and on the page; quotes the path asked for as Go quotes a string, so no control character
// in it reaches the page; and leads to where the site begins, by its path, under the site's name. It is held to
// its golden file whole.
func TestAMissingPageReadsAsItsGoldenMarkup(test *testing.T) {
	missing := NewNotFound(fixtureSite(fixtures.LoadedContent(test)), "404 Not Found", "en", fixtures.LoadedBuild(test)).For("/nothing\x1b[2J<b>")
	golden.Match(test, "notfound.html", lined(missing))
}

// A person who works nowhere now is written with no employer, rather than with an employer of no name.
func TestAPersonWorkingNowhereNamesNoEmployer(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	resume := document.Page{Kind: document.PageResume, Slug: ""}
	record := blankPerson()
	record.Name = "Mira Example"
	if page := fixtureWriter(test).Page(composedOf(test, snapshot, resume).Document, record); strings.Contains(page, "worksFor") {
		test.Errorf("the page %q names an employer", page)
	}
}

// lined is a page with a line break between each two tags that meet, so its golden file reads in a diff; a space
// or a word between two tags stays where it is.
func lined(page string) string {
	return strings.ReplaceAll(page, "><", ">\n<") + "\n"
}

// fixtureWriter writes the pages of the fixture's site, with the build's fixture and the fixture's words.
func fixtureWriter(test *testing.T) PageWriter {
	snapshot := fixtures.LoadedContent(test)
	return NewPageWriter(fixtureSite(snapshot), fixtures.LoadedBuild(test), compose.ThemeWordsOf(snapshot))
}

// fixtureSite is the site of the fixture, at the address its profile gives.
func fixtureSite(snapshot *content.Snapshot) routes.Site {
	return routes.New(snapshot.Profile.Site.BaseURL)
}

// composedOf is a page of the fixture as compose makes it, its document and its record.
func composedOf(test *testing.T, snapshot *content.Snapshot, page document.Page) compose.Composed {
	for _, composed := range compose.Pages(snapshot) {
		if composed.Document.Head.Page == page {
			return composed
		}
	}
	test.Fatalf("compose made no page %+v", page)
	return compose.Composed{Document: document.Document{Head: fixtures.HeadOf(page, nil), Blocks: nil}, Record: nil}
}

// blankPerson is a record of a person with no facts, for a page a test builds by hand: every page a person reads
// has a record, and a page standing for the résumé has a person's.
func blankPerson() compose.PersonRecord {
	return compose.PersonRecord{
		Name: "", AlternateNames: nil, Title: "", Email: "", Employer: "", Schools: nil, Skills: nil, Languages: nil,
		Profiles: nil,
	}
}
