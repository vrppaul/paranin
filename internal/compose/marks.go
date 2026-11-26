package compose

import (
	"strings"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// markedSpans reads a text of the profile into spans: a run between two code marks (`go test`) is code, inside which
// no mark is read; a result between two highlight marks (==40%==) is a highlight, which may hold code and gaps; and a
// gap ([N]) a gap. A mark never closed stays text.
func markedSpans(text string) []document.Span {
	spans := []document.Span{}
	rest := text
	for {
		opening := proseIndex(rest, content.HighlightMark)
		if opening < 0 {
			break
		}
		insideStart := opening + len(content.HighlightMark)
		insideLength := proseIndex(rest[insideStart:], content.HighlightMark)
		if insideLength <= 0 {
			break
		}
		spans = append(spans, codeSpans(rest[:opening])...)
		spans = append(spans, document.Highlight{Spans: codeSpans(rest[insideStart : insideStart+insideLength])})
		rest = rest[insideStart+insideLength+len(content.HighlightMark):]
	}
	return append(spans, codeSpans(rest)...)
}

// How many code marks hold a run of code, one at each side
const codeMarksRound = 2

// markedPiece is a run of a text of the profile between its code marks: code, read as written, or prose, whose other
// marks are read.
type markedPiece struct {
	text string
	code bool
}

// codePieces parts a text at its code: a code mark opens a run of code and the next one closes it, with something
// between them; a code mark with none after it is prose.
func codePieces(text string) []markedPiece {
	pieces := []markedPiece{}
	rest := text
	for {
		opening := strings.Index(rest, content.CodeMark)
		if opening < 0 {
			break
		}
		insideStart := opening + len(content.CodeMark)
		insideLength := strings.Index(rest[insideStart:], content.CodeMark)
		if insideLength <= 0 {
			break
		}
		pieces = append(pieces, markedPiece{text: rest[:opening], code: false}, markedPiece{text: rest[insideStart : insideStart+insideLength], code: true})
		rest = rest[insideStart+insideLength+len(content.CodeMark):]
	}
	return append(pieces, markedPiece{text: rest, code: false})
}

// proseIndex is where a mark first stands in a text outside its code; -1 where it stands nowhere outside it.
func proseIndex(text, mark string) int {
	pieceStart := 0
	for _, piece := range codePieces(text) {
		if at := strings.Index(piece.text, mark); !piece.code && at >= 0 {
			return pieceStart + at
		}
		pieceStart += len(piece.text)
		if piece.code {
			pieceStart += codeMarksRound * len(content.CodeMark)
		}
	}
	return -1
}

// codeSpans reads a text with no highlight into code, as written, and the prose round it, read for gaps.
func codeSpans(text string) []document.Span {
	spans := []document.Span{}
	for _, piece := range codePieces(text) {
		if piece.code {
			spans = append(spans, document.Code{Text: piece.text})
		} else {
			spans = append(spans, gapSpans(piece.text)...)
		}
	}
	return spans
}

// gapSpans reads prose with no highlight and no code into text and gaps. A gap opens with a bracket and closes with
// the next one, with something between them and no bracket opened inside; any other bracket is text.
func gapSpans(text string) []document.Span {
	spans := []document.Span{}
	textStart, cursor := 0, 0
	for {
		opening := strings.Index(text[cursor:], content.GapOpening)
		if opening < 0 {
			break
		}
		opening += cursor
		insideStart := opening + len(content.GapOpening)
		bracket := strings.IndexAny(text[insideStart:], content.GapOpening+content.GapClosing)
		if bracket < 0 {
			break
		}
		cursor = insideStart + bracket
		if bracket == 0 || !strings.HasPrefix(text[cursor:], content.GapClosing) {
			continue
		}
		gapEnd := cursor + len(content.GapClosing)
		spans = appendText(spans, text[textStart:opening])
		spans = append(spans, document.Gap{Text: text[opening:gapEnd]})
		textStart, cursor = gapEnd, gapEnd
	}
	return appendText(spans, text[textStart:])
}

// appendText adds a run of text, where there is one.
func appendText(spans []document.Span, text string) []document.Span {
	if text == "" {
		return spans
	}
	return append(spans, document.Text{Text: text})
}
