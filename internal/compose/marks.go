package compose

import (
	"strings"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// markedSpans reads a text of the profile into spans: a result between two marks (==40%==) is a highlight, and
// a gap ([N]) a gap, which may sit inside a highlight. A mark never closed stays text.
func markedSpans(text string) []document.Span {
	spans := []document.Span{}
	rest := text
	for {
		opening := strings.Index(rest, content.HighlightMark)
		if opening < 0 {
			break
		}
		insideStart := opening + len(content.HighlightMark)
		insideLength := strings.Index(rest[insideStart:], content.HighlightMark)
		if insideLength <= 0 {
			break
		}
		spans = append(spans, gapSpans(rest[:opening])...)
		spans = append(spans, document.Highlight{Spans: gapSpans(rest[insideStart : insideStart+insideLength])})
		rest = rest[insideStart+insideLength+len(content.HighlightMark):]
	}
	return append(spans, gapSpans(rest)...)
}

// gapSpans reads a text with no highlight into text and gaps. A gap opens with a bracket and closes with the next
// one, with something between them and no bracket opened inside; any other bracket is text.
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
