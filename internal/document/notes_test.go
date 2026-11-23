package document

import (
	"reflect"
	"testing"
)

// The notes a face gathers at the end are each note once, in the order of their first citation, wherever in the
// blocks a line cites them: in a list's item, a quote, a table's header and its cells, under a mark.
func TestTheNotesGatherEachNoteOnceInTheOrderFirstCited(test *testing.T) {
	first := Note{Number: 1, Spans: []Span{Text{Text: "A source."}}}
	second := Note{Number: 2, Spans: []Span{Text{Text: "Another."}}}
	third := Note{Number: 3, Spans: []Span{Text{Text: "A third."}}}
	fourth := Note{Number: 4, Spans: []Span{Text{Text: "A fourth."}}}
	blocks := []Block{
		Paragraph{Spans: []Span{Text{Text: "Cited"}, first}},
		Section{Title: []Span{Text{Text: "Part"}}, Blocks: []Block{
			List{Ordered: false, Start: 0, Items: []Item{{Blocks: []Block{Paragraph{Spans: []Span{Strong{Spans: []Span{second}}}}}}}},
			Quote{Blocks: []Block{Paragraph{Spans: []Span{first}}}},
			Table{
				Header: Row{Cells: []Cell{{Kind: CellText, Spans: []Span{third}}}},
				Rows:   []Row{{Cells: []Cell{{Kind: CellText, Spans: []Span{fourth, second}}}}},
			},
		}},
	}
	if got, want := NotesOf(blocks), []Note{first, second, third, fourth}; !reflect.DeepEqual(got, want) {
		test.Errorf("NotesOf = %+v, want %+v", got, want)
	}
}

// A title read as plain text, as a tab's name and a browser's title bar read it, holds the words of its marks and
// none of a note it cites, whose words stand apart from it.
func TestPlainTextLeavesANotesWordsOut(test *testing.T) {
	title := []Span{
		Emphasis{Spans: []Span{Text{Text: "Reading"}}},
		Text{Text: " "},
		Strong{Spans: []Span{Text{Text: "slowly"}}},
		Note{Number: 1, Spans: []Span{Text{Text: "A source."}}},
	}
	if got, want := PlainText(title), "Reading slowly"; got != want {
		test.Errorf("PlainText = %q, want %q", got, want)
	}
}
