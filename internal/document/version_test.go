package document

import "testing"

// linksTo is whether spans hold a link to a page.
func linksTo(spans []Span, page Page) bool {
	for _, span := range spans {
		if link, isLink := span.(Link); isLink && link.Target == page {
			return true
		}
	}
	return false
}

// What an older version says of the newer one sets the site's words round that version's day, as a language puts
// them, before it or after it, the day a link to the newer version.
func TestTheNewerVersionsDayStandsAmongTheSitesWords(test *testing.T) {
	newer := PostVersion{Page: Page{Kind: PagePost, Slug: "slow", Tab: "", Revision: ""}, Name: "2026-11-21"}
	cases := []struct {
		name  string
		words Around
		want  string
	}{
		{name: "words before the day", words: Around{Before: "newer: ", After: ""}, want: "newer: 2026-11-21"},
		{name: "words after the day", words: Around{Before: "", After: " is newer"}, want: "2026-11-21 is newer"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			spans := VersionOlder{Newer: newer, NewerWords: testCase.words}.NewerSpans("2026-11-21")
			if got := PlainText(spans); got != testCase.want || !linksTo(spans, newer.Page) {
				test.Errorf("the newer spans = %+v, reading %q; want %q, the day a link to %+v", spans, got, testCase.want, newer.Page)
			}
		})
	}
}
