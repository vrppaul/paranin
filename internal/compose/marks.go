package compose

import (
	"slices"
	"strings"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// markedSpans reads a text of the profile into spans: a run between two code marks (`go test`) is code, inside which
// no mark is read; a result between two highlight marks (==40%==) is a highlight, which may hold code and gaps; and a
// gap ([N]) a gap. A mark never closed stays text.
func markedSpans(text string) []document.Span {
	marked := markedText{text: text, code: codeRunsOf(text)}
	spans, cursor := []document.Span{}, 0
	for {
		opening := marked.proseIndex(content.HighlightMark, cursor)
		if opening < 0 {
			break
		}
		insideStart := opening + len(content.HighlightMark)
		closing := marked.proseIndex(content.HighlightMark, insideStart)
		if closing <= insideStart {
			break
		}
		spans = append(spans, marked.codeSpans(cursor, opening)...)
		spans = append(spans, document.Highlight{Spans: marked.codeSpans(insideStart, closing)})
		cursor = closing + len(content.HighlightMark)
	}
	return append(spans, marked.codeSpans(cursor, len(text))...)
}

// markedText is a text of the profile with its runs of code, read once, so its other marks are read round them.
type markedText struct {
	text string
	code []codeRun
}

// codeRun is a run of code in a text, by where its opening code mark and its closing one stand.
type codeRun struct {
	opening int
	closing int
}

// end is where the text after a run of code starts, its closing mark passed.
//
//wrapper:names where the prose after code starts, which both reading the highlights and the spans need
func (run codeRun) end() int {
	return run.closing + len(content.CodeMark)
}

// codeRunsOf are a text's runs of code: a code mark opens one and the next closes it, with something between them; a
// pair with nothing between them, and a code mark with none after it, are prose.
func codeRunsOf(text string) []codeRun {
	runs, cursor := []codeRun{}, 0
	for {
		opening := strings.Index(text[cursor:], content.CodeMark)
		if opening < 0 {
			return runs
		}
		opening += cursor
		insideStart := opening + len(content.CodeMark)
		closing := strings.Index(text[insideStart:], content.CodeMark)
		if closing < 0 {
			return runs
		}
		run := codeRun{opening: opening, closing: insideStart + closing}
		if closing > 0 {
			runs = append(runs, run)
		}
		cursor = run.end()
	}
}

// proseIndex is where a mark first stands at or after a place of the text outside its code; -1 where it stands nowhere
// outside it.
func (marked markedText) proseIndex(mark string, from int) int {
	cursor := from
	for {
		found := strings.Index(marked.text[cursor:], mark)
		if found < 0 {
			return -1
		}
		found += cursor
		holding := slices.IndexFunc(marked.code, func(run codeRun) bool { return run.opening <= found && found < run.end() })
		if holding < 0 {
			return found
		}
		cursor = marked.code[holding].end()
	}
}

// codeSpans reads the text from one place to another, which no highlight mark stands between, into code, as written,
// and the prose round it, read for gaps.
func (marked markedText) codeSpans(from, to int) []document.Span {
	spans, proseStart := []document.Span{}, from
	for _, run := range marked.code {
		if run.opening < from || run.opening >= to {
			continue
		}
		spans = append(spans, gapSpans(marked.text[proseStart:run.opening])...)
		spans = append(spans, document.Code{Text: marked.text[run.opening+len(content.CodeMark) : run.closing]})
		proseStart = run.end()
	}
	return append(spans, gapSpans(marked.text[proseStart:to])...)
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
