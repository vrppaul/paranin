package tui

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/x/ansi"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// What parts words and lines
const (
	space      = " "
	spaceWidth = 1
	lineEnd    = "\n"
)

// kind is what a piece of text is to its reader; the paint decides how each kind looks.
type kind int

const (
	// kindText is what a reader reads
	kindText kind = iota
	// kindMuted stands beside the text and says less
	kindMuted
	// kindFaint may be passed over: a mark, a gap still to fill
	kindFaint
	// kindStrong is a title
	kindStrong
	// kindCode is a command a reader types or copies
	kindCode
)

// run is a piece of text of one kind: highlighted where it is a result worth seeing first, and leading to an
// address where it is a link.
type run struct {
	text        string
	kind        kind
	highlighted bool
	address     string
}

// line is one line of a tab, the runs it is made of; every line a layout makes opens with its indent.
type line []run

// word is text no line breaks inside: its runs, and whether the space before it lay in a highlight, which the
// space between it and the word before keeps.
type word struct {
	runs            []run
	highlightBefore bool
}

// words are the words of spans read as a kind of text: marks become runs of their own kind, a link carries its
// address, and a command stays one word, its spaces and all, so it never breaks.
func (layout layout) words(spans []document.Span, base kind) []word {
	reader := wordReader{site: layout.site, words: []word{}, current: word{runs: nil, highlightBefore: false}}
	reader.read(spans, run{text: "", kind: base, highlighted: false, address: ""})
	return reader.done()
}

// textWords are the words of a text with no marks, all of one kind: a place, a tag, a period.
func textWords(text string, base kind) []word {
	words := []word{}
	for written := range strings.FieldsSeq(text) {
		words = append(words, word{runs: []run{{text: written, kind: base, highlighted: false, address: ""}}, highlightBefore: false})
	}
	return words
}

// wordReader gathers words from spans read one by one: a word goes on until a space ends it.
type wordReader struct {
	site    routes.Site
	words   []word
	current word
}

// read takes spans in the form a run gives them, each kind of span changing a copy of it as the span stands for.
func (reader *wordReader) read(spans []document.Span, form run) {
	for _, span := range spans {
		switch typed := span.(type) {
		case document.Text:
			reader.letters(typed.Text, form)
		case document.Code:
			code := form
			code.kind = kindCode
			reader.keep(typed.Text, code)
		case document.Gap:
			gap := form
			gap.kind = kindFaint
			reader.letters(typed.Text, gap)
		case document.Highlight:
			highlighted := form
			highlighted.highlighted = true
			reader.read(typed.Spans, highlighted)
		case document.Link:
			linked := form
			linked.address = reader.site.Address(typed.Target, routes.FormatChosen)
			reader.read(typed.Spans, linked)
		}
	}
}

// letters takes text letter by letter, ending a word at each space.
func (reader *wordReader) letters(text string, form run) {
	for _, letter := range text {
		if unicode.IsSpace(letter) {
			reader.end()
			reader.current.highlightBefore = form.highlighted
			continue
		}
		reader.keep(string(letter), form)
	}
}

// keep takes text into the word in hand whole, in a run of its own where its form differs from the last one.
func (reader *wordReader) keep(text string, form run) {
	runs := reader.current.runs
	if last := len(runs) - 1; last >= 0 && sameForm(runs[last], form) {
		runs[last].text += text
		return
	}
	form.text = text
	reader.current.runs = append(runs, form)
}

func (reader *wordReader) end() {
	if len(reader.current.runs) > 0 {
		reader.words = append(reader.words, reader.current)
	}
	reader.current = word{runs: nil, highlightBefore: false}
}

func (reader *wordReader) done() []word {
	reader.end()
	return reader.words
}

func sameForm(first, second run) bool {
	return first.kind == second.kind && first.highlighted == second.highlighted && first.address == second.address
}

// fill lays words out in lines of the layout's width, each from an indent. A word wider than a line breaks where
// the line ends, as a terminal breaks it, so none of it is lost.
func (layout layout) fill(words []word, indent int) []line {
	lines := []line{}
	current, used := line{indentRun(indent)}, 0
	for _, next := range brokenWords(words, max(layout.width-indent, 1)) {
		width := wordsWidth([]word{next})
		switch {
		case used == 0:
			current, used = append(current, next.runs...), width
		case indent+used+spaceWidth+width <= layout.width:
			current = append(append(current, spaceRun(next)), next.runs...)
			used += spaceWidth + width
		default:
			lines = append(lines, current)
			current, used = append(line{indentRun(indent)}, next.runs...), width
		}
	}
	if used > 0 {
		lines = append(lines, current)
	}
	return lines
}

// brokenWords are words with each one wider than the room broken into pieces as wide as the room.
func brokenWords(words []word, room int) []word {
	broken := make([]word, 0, len(words))
	for _, next := range words {
		for wordsWidth([]word{next}) > room {
			piece, rest := cutWord(next, room)
			broken = append(broken, piece)
			next = rest
		}
		broken = append(broken, next)
	}
	return broken
}

// cutWord parts a word after a width: the piece that fills it, and the rest, with no space before it.
func cutWord(whole word, width int) (word, word) {
	piece := word{runs: nil, highlightBefore: whole.highlightBefore}
	rest := word{runs: nil, highlightBefore: false}
	used := 0
	for _, part := range whole.runs {
		partWidth := textWidth(part.text)
		switch {
		case used >= width:
			rest.runs = append(rest.runs, part)
		case used+partWidth <= width:
			piece.runs = append(piece.runs, part)
		default:
			head, tail := part, part
			head.text, tail.text = ansi.Cut(part.text, 0, width-used), ansi.Cut(part.text, width-used, partWidth)
			piece.runs, rest.runs = append(piece.runs, head), append(rest.runs, tail)
		}
		used += partWidth
	}
	return piece, rest
}

// spaceRun is the space before a word, highlighted where the space was.
func spaceRun(next word) run {
	return run{text: space, kind: kindText, highlighted: next.highlightBefore, address: ""}
}

// wordRuns are words laid in one line, a space between each two.
func wordRuns(words []word) line {
	runs := line{}
	for index, next := range words {
		if index > 0 {
			runs = append(runs, spaceRun(next))
		}
		runs = append(runs, next.runs...)
	}
	return runs
}

// wordsWidth is how many columns words take on one line, a space between each two.
func wordsWidth(words []word) int {
	width := 0
	for _, piece := range wordRuns(words) {
		width += textWidth(piece.text)
	}
	return width
}

func indentRun(indent int) run {
	return gapRun(indent)
}

func gapRun(width int) run {
	return run{text: strings.Repeat(space, max(width, 0)), kind: kindText, highlighted: false, address: ""}
}

// padded is a text with spaces after it to a width.
func padded(text string, width int) string {
	return text + strings.Repeat(space, max(width-textWidth(text), 0))
}

// textWidth is how many columns of a terminal a text takes.
func textWidth(text string) int {
	return ansi.StringWidth(text)
}
