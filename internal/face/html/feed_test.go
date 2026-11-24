package html

import (
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/golden"
)

// How a test finds every address a body leads to or loads, each match's address after the whole match, and every
// match there is
const (
	addressAttribute = `(?:href|src)="([^"]*)"`
	addressMatched   = 1
	everyMatch       = -1
)

// A post's body in a feed is what a feed reader shows of it: its headings, paragraphs, table, lists, quote, picture
// and code as plain HTML, every note once at its end, numbered as its citations are. It is held to its golden file,
// named after the post's address in the feed's folder.
func TestAPostsBodyInAFeedReadsAsItsGoldenMarkup(test *testing.T) {
	golden.Match(test, "feed/blog/reading-slowly.html", lined(fixtureFeedBody(test)))
}

// A feed reader shows a post's body away from the site, so every address the body leads to or loads is whole.
func TestAPostsBodyInAFeedLeadsByWholeAddresses(test *testing.T) {
	for _, attribute := range regexp.MustCompile(addressAttribute).FindAllStringSubmatch(fixtureFeedBody(test), everyMatch) {
		if address, err := url.Parse(attribute[addressMatched]); err != nil || !address.IsAbs() {
			test.Errorf("the post's body in a feed leads to %q, which is no whole address", attribute[addressMatched])
		}
	}
}

// A feed reader shows a post's body with no stylesheet of the site's and runs no script, so the body holds none of the
// page's own parts, which would show there as stray words or do nothing; a class it passes over.
func TestAPostsBodyInAFeedCarriesNoPartOfThePage(test *testing.T) {
	cases := []struct {
		name    string
		refused string
	}{
		{name: "a paragraph's number", refused: `class="paragraph-number"`},
		{name: "a point's mark", refused: `class="point-mark"`},
		{name: "a note beside its line", refused: `class="notes"`},
		{name: "an id", refused: " id="},
		{name: "a part the site's script finds", refused: "data-"},
		{name: "a style", refused: "style="},
		{name: "a button", refused: "<button"},
		{name: "a script", refused: "<script"},
	}
	body := fixtureFeedBody(test)
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if strings.Contains(body, testCase.refused) {
				test.Errorf("the post's body in a feed %q holds %q", body, testCase.refused)
			}
		})
	}
}

// fixtureFeedBody is the fixture post's body as a feed holds it.
func fixtureFeedBody(test *testing.T) string {
	return NewFeedWriter(fixtures.Site(fixtures.LoadedContent(test))).Body(fixturePost(test).Document)
}
