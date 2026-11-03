// Package text draws documents as text an agent or a terminal reads (docs/ARCHITECTURE.md, ADR-0003): CommonMark
// for an agent, and a manual page for a terminal, laid out in 78 columns as man(1) lays one out, a section of the
// page for each section of the document, a command one word that never breaks. Each draws every kind of block
// and span a document has, and writes every address through routes. The site's index is markdown too: a
// document's lead is a quote, so the index reads as llmstxt.org lays out llms.txt, a title, a quote of what the
// site is, and its pages listed under their sections. An address of no page reads as NotFound.
//
// In markdown, text reads as the text it is: what CommonMark would take for markup is escaped, and so is what
// would open a block at the start of any line or close a heading. The escapes are part of the text a `curl`
// reader sees (search\_code), the price of one rule for every text over a rule for each place in a line.
package text

import (
	"slices"
	"strings"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// How markdown is written here
const (
	headingMark   = "#"
	listMark      = "- "
	quoteMark     = ">"
	highlightMark = "**"
	codeMark      = "`"
	escapeMark    = `\`
	blankLine     = "\n\n"
	lineEnd       = "\n"
	space         = " "
	// the document's title, and a heading under it, one level down
	titleDepth   = 1
	sectionDepth = 2
	// the fewest backticks a block of code is fenced with
	shortestFence = 3
)

// How a link is written: [text](address), the address between brackets where it holds a space or a parenthesis
const (
	linkTextOpening        = "["
	linkTextClosing        = "]"
	linkAddressOpening     = "("
	linkAddressClosing     = ")"
	enclosedOpening        = "<"
	enclosedClosing        = ">"
	addressBreakingLetters = " ()"
)

// What CommonMark reads as markup inside a line: an escape, emphasis, code, a link, HTML, an entity. A gap is
// written with its brackets, so they stay
const (
	inlineMarkup = `\*_` + "`" + "[]<&"
	gapMarkup    = `\*_` + "`" + "<&"
)

// What opens a block when a line starts with it: a heading, a quote, a list, a rule, a fence, a table; and a
// number ended by one of its ends, an ordered list. Spaces opening a line would make it code, so they go
const (
	blockOpeners    = "#>-+=|~"
	digits          = "0123456789"
	orderedListEnds = ".)"
	lineIndent      = " \t"
)

// What stands between the parts of one line
const (
	// a title and its place, a tag and a subtitle
	dot = " · "
	// a title and its subtitle, the cells of a row
	dash = " — "
	// the start and the end of a period
	rangeDash = " – "
	// a label and its value
	labelEnd = ": "
)

// What an escaper replaces each letter with: the letter, and the letter after its escape
const replacementWidth = 2

// markdownWriter writes one document: the site writes its addresses, and the escapers keep text and gaps from
// reading as markup.
type markdownWriter struct {
	site       routes.Site
	textEscape *strings.Replacer
	gapEscape  *strings.Replacer
}

// Markdown writes a document as CommonMark: its head, then each block, a blank line between them.
func Markdown(composed document.Document, site routes.Site) string {
	writer := markdownWriter{site: site, textEscape: escaperOf(inlineMarkup), gapEscape: escaperOf(gapMarkup)}
	chunks := writer.head(composed.Head)
	chunks = writer.blocks(chunks, composed.Blocks, sectionDepth)
	return strings.Join(chunks, blankLine) + lineEnd
}

// escaperOf puts an escape before each of the letters.
func escaperOf(letters string) *strings.Replacer {
	pairs := make([]string, 0, len(letters)*replacementWidth)
	for _, letter := range letters {
		pairs = append(pairs, string(letter), escapeMark+string(letter))
	}
	return strings.NewReplacer(pairs...)
}

func (writer markdownWriter) head(head document.Head) []string {
	chunks := []string{heading(titleDepth, writer.spans(head.Title))}
	chunks = appendLine(chunks, joined(dot, writer.spans(head.Subtitle), writer.text(head.Place)))
	chunks = appendChunk(chunks, quote(writer.spans(head.Lead)))
	return appendChunk(chunks, writer.facts(head.Ways))
}

func (writer markdownWriter) blocks(chunks []string, blocks []document.Block, depth int) []string {
	for _, block := range blocks {
		chunks = writer.block(chunks, block, depth)
	}
	return chunks
}

func (writer markdownWriter) block(chunks []string, block document.Block, depth int) []string {
	switch typed := block.(type) {
	case document.Section:
		chunks = appendChunk(chunks, heading(depth, writer.spans(typed.Title)))
		return writer.blocks(chunks, typed.Blocks, depth+1)
	case document.Paragraph:
		return appendLine(chunks, writer.spans(typed.Spans))
	case document.Table:
		return appendChunk(chunks, writer.table(typed))
	case document.CodeBlock:
		return appendChunk(chunks, codeBlock(typed.Text))
	case document.Facts:
		return appendChunk(chunks, writer.facts(typed.List))
	case document.Entry:
		return writer.entry(chunks, typed, depth)
	case document.Card:
		return writer.card(chunks, typed, depth)
	case document.Contents:
		return appendChunk(chunks, writer.contents(typed))
	case document.Checks, document.Skills:
		// only the hidden pages hold checks and the list of skills, and no host hands them to a text face: a page
		// served with them would fail TestEveryPhraseReachesEveryFormat
	}
	return chunks
}

// table writes each row as an item, its cells in a line.
func (writer markdownWriter) table(table document.Table) string {
	items := make([]string, 0, len(table.Rows))
	for _, row := range table.Rows {
		cells := make([]string, 0, len(row.Cells))
		for _, cell := range row.Cells {
			cells = append(cells, writer.spans(cell.Spans))
		}
		items = append(items, joined(dash, cells...))
	}
	return list(items)
}

func (writer markdownWriter) facts(facts []document.Fact) string {
	items := make([]string, 0, len(facts))
	for _, fact := range facts {
		items = append(items, writer.text(fact.Label)+labelEnd+writer.spans(fact.Value))
	}
	return list(items)
}

// entry writes an entry under its title and subtitle, its period and place in a line, then its points.
func (writer markdownWriter) entry(chunks []string, entry document.Entry, depth int) []string {
	chunks = appendChunk(chunks, heading(depth, joined(dash, writer.spans(entry.Title), writer.spans(entry.Subtitle))))
	chunks = appendLine(chunks, joined(dot, writer.period(entry.Period), writer.text(entry.Place)))
	points := make([]string, 0, len(entry.Points))
	for _, point := range entry.Points {
		points = append(points, writer.spans(point.Spans))
	}
	return appendChunk(chunks, list(points))
}

// card writes a card under its title, a link to its page where it has one; its tag and subtitle in a line, its
// summary, then its facts.
func (writer markdownWriter) card(chunks []string, card document.Card, depth int) []string {
	title := card.Title
	if card.Page.Kind != document.PageNone {
		title = []document.Span{document.Link{Target: card.Page, Spans: card.Title}}
	}
	chunks = appendChunk(chunks, heading(depth, writer.spans(title)))
	chunks = appendLine(chunks, joined(dot, writer.text(card.Tag), writer.spans(card.Subtitle)))
	chunks = appendLine(chunks, writer.spans(card.Summary))
	return appendChunk(chunks, writer.facts(card.Facts))
}

// contents writes each listed page as an item: a link to the page in markdown, then the line on what it is.
func (writer markdownWriter) contents(contents document.Contents) string {
	items := make([]string, 0, len(contents.Listings))
	for _, listing := range contents.Listings {
		title := writer.spans([]document.Span{document.Link{Target: listing.Page, Spans: listing.Title}})
		items = append(items, joined(labelEnd, title, writer.spans(listing.Note)))
	}
	return list(items)
}

// period writes a period escaped as a gap is: its gaps keep their brackets.
func (writer markdownWriter) period(period document.Period) string {
	return writer.gapEscape.Replace(periodText(period))
}

func (writer markdownWriter) spans(spans []document.Span) string {
	var builder strings.Builder
	for _, span := range spans {
		switch typed := span.(type) {
		case document.Text:
			builder.WriteString(writer.text(typed.Text))
		case document.Code:
			builder.WriteString(codeSpan(typed.Text))
		case document.Gap:
			builder.WriteString(writer.gapEscape.Replace(typed.Text))
		case document.Highlight:
			builder.WriteString(highlightMark + writer.spans(typed.Spans) + highlightMark)
		case document.Link:
			builder.WriteString(link(writer.spans(typed.Spans), writer.site.Address(typed.Target, routes.FormatMarkdown)))
		}
	}
	return builder.String()
}

// text escapes a text, each line after a break of its own included.
func (writer markdownWriter) text(text string) string {
	lines := strings.Split(writer.textEscape.Replace(text), lineEnd)
	for index, line := range lines[1:] {
		lines[index+1] = lineStart(line)
	}
	return strings.Join(lines, lineEnd)
}

// heading writes a title at a depth; a heading's mark closing the title is escaped, as CommonMark would drop it.
func heading(depth int, title string) string {
	if closed, found := strings.CutSuffix(title, headingMark); found {
		title = closed + escapeMark + headingMark
	}
	return strings.Repeat(headingMark, depth) + space + title
}

// link writes a link, its address between brackets where a space or a parenthesis would end it.
func link(text, address string) string {
	if strings.ContainsAny(address, addressBreakingLetters) {
		address = enclosedOpening + address + enclosedClosing
	}
	return linkTextOpening + text + linkTextClosing + linkAddressOpening + address + linkAddressClosing
}

// codeSpan writes code between runs of backticks one longer than any run inside it, so none of its own closes
// it. Code that opens or closes with a backtick, or opens and closes with a space, gets a space inside each
// run, which CommonMark takes away again.
func codeSpan(code string) string {
	fence := strings.Repeat(codeMark, longestRun(code, codeMark)+1)
	padded := strings.HasPrefix(code, codeMark) || strings.HasSuffix(code, codeMark) ||
		(strings.HasPrefix(code, space) && strings.HasSuffix(code, space) && strings.TrimSpace(code) != "")
	if padded {
		code = space + code + space
	}
	return fence + code + fence
}

// codeBlock fences text with more backticks than any run inside it, and three at least.
func codeBlock(code string) string {
	fence := strings.Repeat(codeMark, max(shortestFence, longestRun(code, codeMark)+1))
	return fence + lineEnd + code + lineEnd + fence
}

// longestRun is the length of the longest run of a mark in a text.
func longestRun(text, mark string) int {
	longest, run := 0, 0
	for _, letter := range text {
		if string(letter) != mark {
			run = 0
			continue
		}
		run++
		longest = max(longest, run)
	}
	return longest
}

// lineStart escapes what a line opening with this text would read as a block, its opening spaces gone.
func lineStart(line string) string {
	line = strings.TrimLeft(line, lineIndent)
	number := len(line) - len(strings.TrimLeft(line, digits))
	if number > 0 && strings.ContainsAny(line[number:min(number+1, len(line))], orderedListEnds) {
		return line[:number] + escapeMark + line[number:]
	}
	if number == 0 && line != "" && strings.ContainsAny(line[:1], blockOpeners) {
		return escapeMark + line
	}
	return line
}

// quote writes text as a quote, where there is one: every line of it marked, a blank one included, so a quote
// of two paragraphs stays one quote; the text after each mark opens no block of its own.
func quote(text string) string {
	if text == "" {
		return ""
	}
	lines := strings.Split(text, lineEnd)
	for index, line := range lines {
		lines[index] = quoteLine(line)
	}
	return strings.Join(lines, lineEnd)
}

// quoteLine is one line of a quote: its mark alone where the line is blank.
func quoteLine(line string) string {
	if line == "" {
		return quoteMark
	}
	return quoteMark + space + lineStart(line)
}

// list writes items as a markdown list, one a line; an item reads as text, never as a block of its own.
func list(items []string) string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, listMark+lineStart(item))
	}
	return strings.Join(lines, lineEnd)
}

// joined writes the parts that are there, with a separator between them.
func joined(separator string, parts ...string) string {
	present := slices.DeleteFunc(slices.Clone(parts), func(part string) bool { return part == "" })
	return strings.Join(present, separator)
}

// appendLine adds a line of text, where there is one, escaped so it opens no block.
func appendLine(chunks []string, line string) []string {
	return appendChunk(chunks, lineStart(line))
}

// appendChunk adds a chunk of the document, where there is one.
func appendChunk(chunks []string, chunk string) []string {
	if chunk == "" {
		return chunks
	}
	return append(chunks, chunk)
}
