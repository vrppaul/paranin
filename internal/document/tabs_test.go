package document_test

import (
	"slices"
	"testing"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// tabName is a tab as a reader names it: its slug, and how many blocks it shows.
type tabName struct {
	slug   string
	blocks int
}

// partedName is a document's tabs as a reader names them: the opening, how many facts the brief it opens with holds,
// none where it opens with no brief, the rest.
type partedName struct {
	opening tabName
	brief   int
	rest    []tabName
}

// A résumé opens on its brief alone, the first three facts its first section holds; what else that section holds, and
// the blocks before it, stand on no tab.
func TestADocumentPartsIntoAnOpeningAndATabForEachSectionLeft(test *testing.T) {
	fourFacts := []document.Fact{}
	for _, label := range []string{"experience", "stack", "languages", "now"} {
		fourFacts = append(fourFacts, document.Fact{Label: label, Value: textOf(label + " value")})
	}
	facts := document.Facts{Kind: document.FactsDetails, List: fourFacts}
	ways := document.Facts{Kind: document.FactsWays, List: []document.Fact{{Label: "mail", Value: textOf("ada@example.test")}}}
	paragraph := document.Paragraph{Spans: textOf("words")}
	about := document.Section{Title: textOf("About"), Blocks: []document.Block{paragraph, facts, paragraph}}
	aboutWithoutFacts := document.Section{Title: textOf("About"), Blocks: []document.Block{paragraph, ways}}
	experience := document.Section{Title: textOf("Experience"), Blocks: []document.Block{paragraph, paragraph}}
	install := document.Section{Title: textOf("Step 2: install & run"), Blocks: []document.Block{paragraph}}
	testCases := []struct {
		name   string
		kind   document.PageKind
		title  string
		blocks []document.Block
		want   partedName
	}{
		{
			name: "a résumé opens on its brief alone, the first three facts of its first section", kind: document.PageResume,
			title: "Ada Lovelace", blocks: []document.Block{paragraph, about, experience},
			want: partedName{opening: tabName{slug: "about", blocks: 0}, brief: 3, rest: []tabName{{slug: "experience", blocks: 2}}},
		},
		{
			name: "a résumé with no brief opens with its head", kind: document.PageResume,
			title: "Ada Lovelace", blocks: []document.Block{experience},
			want: partedName{opening: tabName{slug: "ada-lovelace", blocks: 0}, brief: 0, rest: []tabName{{slug: "experience", blocks: 2}}},
		},
		{
			name: "a résumé whose first section holds the ways alone has no brief, the ways in their tab once", kind: document.PageResume,
			title: "Ada Lovelace", blocks: []document.Block{aboutWithoutFacts, experience},
			want: partedName{opening: tabName{slug: "ada-lovelace", blocks: 0}, brief: 0, rest: []tabName{{slug: "about", blocks: 2}, {slug: "experience", blocks: 2}}},
		},
		{
			name: "a project is one tab holding its sections", kind: document.PageProject,
			title: "graph-rs", blocks: []document.Block{paragraph, about, install},
			want: partedName{opening: tabName{slug: "graph-rs", blocks: 3}, brief: 0, rest: []tabName{}},
		},
		{
			name: "a page of no section is its opening alone", kind: document.PageIndex,
			title: "Index", blocks: []document.Block{paragraph},
			want: partedName{opening: tabName{slug: "index", blocks: 1}, brief: 0, rest: []tabName{}},
		},
	}
	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			composed := document.Document{Head: fixtures.HeadOf(document.Page{Kind: testCase.kind, Slug: "", Tab: ""}, textOf(testCase.title)), Blocks: testCase.blocks}
			parted := document.TabsOf(composed)
			got := partedName{opening: nameOf(parted.Opening.Tab), brief: briefSize(parted.Opening), rest: []tabName{}}
			for _, tab := range parted.Rest {
				got.rest = append(got.rest, nameOf(tab))
			}
			if got.opening != testCase.want.opening || got.brief != testCase.want.brief || !slices.Equal(got.rest, testCase.want.rest) {
				test.Errorf("the tabs are %+v, want %+v", got, testCase.want)
			}
		})
	}
}

// briefSize is how many facts the brief an opening opens with holds, none where it opens with no brief.
func briefSize(opening document.Opening) int {
	if opening.Brief == nil {
		return 0
	}
	return len(opening.Brief.Brief.List)
}

// nameOf is a tab as a reader names it.
func nameOf(tab document.Tab) tabName {
	return tabName{slug: tab.Slug, blocks: len(tab.Blocks)}
}

// textOf is plain words as spans.
func textOf(words string) []document.Span {
	return []document.Span{document.Text{Text: words}}
}
