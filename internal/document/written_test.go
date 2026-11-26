package document

import (
	"reflect"
	"testing"
	"time"
)

// A document as a reader of the finished work sees it leaves out a fact, a point, a paragraph of a role's description,
// the experience's line or a technology it puts forward that holds nothing but gaps still to fill, and a list of facts
// or of points or the experience's head left with none; a gap among written words stays, as does every other block.
func TestWrittenLeavesOutWhatIsOnlyAGap(test *testing.T) {
	later := Fact{Label: "looking for", Value: []Span{Gap{Text: "[later]"}}}
	where := Fact{Label: "where", Value: []Span{Text{Text: "Prague · "}, Gap{Text: "[remote?]"}}}
	unwrittenResult := Point{Spans: []Span{Highlight{Spans: []Span{Gap{Text: "[N]"}, Text{Text: " "}, Gap{Text: "[requests]"}}}}}
	writtenResult := Point{Spans: []Span{Highlight{Spans: []Span{Text{Text: "40% faster"}}}}}
	amongWords := Point{Spans: []Span{Text{Text: "Owned "}, Gap{Text: "[service]"}}}
	onlyAGap := Point{Spans: []Span{Gap{Text: "[What was built]"}}}
	gapParagraph := Paragraph{Spans: []Span{Gap{Text: "[summary]"}}}
	writtenParagraph := Paragraph{Spans: []Span{Text{Text: "Ran the queue."}}}
	stressedGap := Point{Spans: []Span{Emphasis{Spans: []Span{Gap{Text: "[N]"}}}}}
	strongGap := Point{Spans: []Span{Strong{Spans: []Span{Gap{Text: "[N]"}}}}}
	strongWords := Point{Spans: []Span{Strong{Spans: []Span{Text{Text: "Shipped"}}}}}
	stressedWords := Point{Spans: []Span{Emphasis{Spans: []Span{Text{Text: "Twice"}}}}}
	gapLine := []Span{Gap{Text: "[the one thing]"}}
	writtenLine := []Span{Text{Text: "Ships tools."}}
	usedTechnology := Technology{Name: "Go", Used: []Span{Text{Text: "3y"}}}
	gapTechnology := Technology{Name: "Lua", Used: []Span{Gap{Text: "[1y?]"}}}
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
			block: Entry{Period: Period{Start: KnownMonth{Start: time.Time{}}, End: nil}, Title: nil, Subtitle: nil, Technologies: nil, Points: []Point{onlyAGap, amongWords, unwrittenResult, writtenResult}, Description: nil},
			want:  []Block{Entry{Period: Period{Start: KnownMonth{Start: time.Time{}}, End: nil}, Title: nil, Subtitle: nil, Technologies: nil, Points: []Point{amongWords, writtenResult}, Description: nil}},
		},
		{
			name:  "a paragraph of a description only a gap",
			block: Entry{Period: Period{Start: KnownMonth{Start: time.Time{}}, End: nil}, Title: nil, Subtitle: nil, Technologies: nil, Points: nil, Description: []Paragraph{gapParagraph, writtenParagraph}},
			want:  []Block{Entry{Period: Period{Start: KnownMonth{Start: time.Time{}}, End: nil}, Title: nil, Subtitle: nil, Technologies: nil, Points: nil, Description: []Paragraph{writtenParagraph}}},
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
		{
			name:  "a point only stressed or strong gaps, and points of words so",
			block: Entry{Period: Period{Start: KnownMonth{Start: time.Time{}}, End: nil}, Title: nil, Subtitle: nil, Technologies: nil, Points: []Point{stressedGap, strongGap, stressedWords, strongWords}, Description: nil},
			want:  []Block{Entry{Period: Period{Start: KnownMonth{Start: time.Time{}}, End: nil}, Title: nil, Subtitle: nil, Technologies: nil, Points: []Point{stressedWords, strongWords}, Description: nil}},
		},
		{name: "a section's point only a gap", block: Points{List: []Point{onlyAGap, amongWords}}, want: []Block{Points{List: []Point{amongWords}}}},
		{name: "a section's points left with none", block: Points{List: []Point{onlyAGap}}, want: []Block{}},
		{
			name:  "the experience's line only a gap",
			block: ExperienceHead{Line: gapLine, Technologies: []Technology{usedTechnology}},
			want:  []Block{ExperienceHead{Line: nil, Technologies: []Technology{usedTechnology}}},
		},
		{
			name:  "a technology put forward only a gap",
			block: ExperienceHead{Line: writtenLine, Technologies: []Technology{gapTechnology, usedTechnology}},
			want:  []Block{ExperienceHead{Line: writtenLine, Technologies: []Technology{usedTechnology}}},
		},
		{name: "the experience's head left with none", block: ExperienceHead{Line: gapLine, Technologies: []Technology{gapTechnology}}, want: []Block{}},
		{name: "a paragraph, which is no fact and no point", block: gapParagraph, want: []Block{gapParagraph}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			got := WrittenBlocks([]Block{testCase.block})
			if !reflect.DeepEqual(got, testCase.want) {
				test.Errorf("WrittenBlocks of %+v = %+v, want %+v", testCase.block, got, testCase.want)
			}
		})
	}
}

// The blocks WrittenBlocks is handed keep every fact and point they had, in their order, for the faces that draw them
// whole.
func TestWrittenKeepsTheBlocksHandedWhole(test *testing.T) {
	facts := []Fact{{Label: "looking for", Value: []Span{Gap{Text: "[later]"}}}, {Label: "now", Value: []Span{Text{Text: "here"}}}}
	points := []Point{{Spans: []Span{Gap{Text: "[What was built]"}}}, {Spans: []Span{Text{Text: "Wrote it"}}}}
	role := Entry{Period: Period{Start: KnownMonth{Start: time.Time{}}, End: nil}, Title: nil, Subtitle: nil, Technologies: nil, Points: points, Description: nil}
	handed := Document{Head: Head{}, Blocks: []Block{Facts{Kind: FactsDetails, List: facts}, role}}
	want := Phrases(handed)
	WrittenBlocks(handed.Blocks)
	if got := Phrases(handed); !reflect.DeepEqual(got, want) {
		test.Errorf("after WrittenBlocks the blocks handed hold %q, want %q", got, want)
	}
}
