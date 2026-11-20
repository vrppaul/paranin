package document

import (
	"reflect"
	"testing"
)

// A document as a reader of the finished work sees it leaves out a fact or a point that holds nothing but gaps still
// to fill, and a list of facts left with none; a gap among written words stays, as does every other block.
func TestWrittenLeavesOutWhatIsOnlyAGap(test *testing.T) {
	later := Fact{Label: "looking for", Value: []Span{Gap{Text: "[later]"}}}
	where := Fact{Label: "where", Value: []Span{Text{Text: "Prague · "}, Gap{Text: "[remote?]"}}}
	unwrittenResult := Point{Spans: []Span{Highlight{Spans: []Span{Gap{Text: "[N]"}, Text{Text: " "}, Gap{Text: "[requests]"}}}}}
	writtenResult := Point{Spans: []Span{Highlight{Spans: []Span{Text{Text: "40% faster"}}}}}
	amongWords := Point{Spans: []Span{Text{Text: "Owned "}, Gap{Text: "[service]"}}}
	onlyAGap := Point{Spans: []Span{Gap{Text: "[What was built]"}}}
	gapParagraph := Paragraph{Spans: []Span{Gap{Text: "[summary]"}}}
	cases := []struct {
		name  string
		block Block
		want  []Block
	}{
		{
			name: "a fact only a gap", block: Facts{Kind: FactsDetails, List: []Fact{later, where}},
			want: []Block{Facts{Kind: FactsDetails, List: []Fact{where}}},
		},
		{name: "a list of facts left with none", block: Facts{Kind: FactsDetails, List: []Fact{later}}, want: []Block{}},
		{
			name:  "a fact only a link to a gap",
			block: Facts{Kind: FactsWays, List: []Fact{where, {Label: "site", Value: []Span{Link{Target: Address{URL: "https://example.test"}, Spans: []Span{Gap{Text: "[site]"}}}}}}},
			want:  []Block{Facts{Kind: FactsWays, List: []Fact{where}}},
		},
		{
			name:  "a point only a gap, in words or in a result",
			block: Entry{Kind: EntryRole, Period: Period{Start: nil, End: nil}, Title: nil, Subtitle: nil, Place: "", Points: []Point{onlyAGap, amongWords, unwrittenResult, writtenResult}},
			want:  []Block{Entry{Kind: EntryRole, Period: Period{Start: nil, End: nil}, Title: nil, Subtitle: nil, Place: "", Points: []Point{amongWords, writtenResult}}},
		},
		{
			name:  "a card's fact",
			block: Card{Title: nil, Subtitle: nil, Tag: "", Summary: nil, Facts: []Fact{where, later}, Page: Page{Kind: PageNone, Slug: "", Tab: ""}},
			want:  []Block{Card{Title: nil, Subtitle: nil, Tag: "", Summary: nil, Facts: []Fact{where}, Page: Page{Kind: PageNone, Slug: "", Tab: ""}}},
		},
		{
			name:  "a fact in a section inside a section",
			block: Section{Title: nil, Blocks: []Block{Section{Title: nil, Blocks: []Block{Facts{Kind: FactsDetails, List: []Fact{where, later}}}}}},
			want:  []Block{Section{Title: nil, Blocks: []Block{Section{Title: nil, Blocks: []Block{Facts{Kind: FactsDetails, List: []Fact{where}}}}}}},
		},
		{name: "a paragraph, which is no fact and no point", block: gapParagraph, want: []Block{gapParagraph}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			got := Written(Document{Head: Head{}, Blocks: []Block{testCase.block}}).Blocks
			if !reflect.DeepEqual(got, testCase.want) {
				test.Errorf("Written of %+v = %+v, want %+v", testCase.block, got, testCase.want)
			}
		})
	}
}

// The document Written is handed keeps every fact and point it had, in its order, for the faces that draw it whole.
func TestWrittenKeepsTheDocumentHandedWhole(test *testing.T) {
	facts := []Fact{{Label: "looking for", Value: []Span{Gap{Text: "[later]"}}}, {Label: "now", Value: []Span{Text{Text: "here"}}}}
	points := []Point{{Spans: []Span{Gap{Text: "[What was built]"}}}, {Spans: []Span{Text{Text: "Wrote it"}}}}
	role := Entry{Kind: EntryRole, Period: Period{Start: nil, End: nil}, Title: nil, Subtitle: nil, Place: "", Points: points}
	handed := Document{Head: Head{}, Blocks: []Block{Facts{Kind: FactsDetails, List: facts}, role}}
	want := Phrases(handed)
	Written(handed)
	if got := Phrases(handed); !reflect.DeepEqual(got, want) {
		test.Errorf("after Written the document handed holds %q, want %q", got, want)
	}
}
