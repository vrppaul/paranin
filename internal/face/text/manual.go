package text

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// How a manual page is laid out, as man(1) lays one out in a terminal of eighty columns
const (
	pageWidth = 78
	// a section's text; what nests under it stands in as far again
	sectionIndent = 7
	// how far a bullet's lines after the first stand in, under its text
	bulletHang = 2
	// the gap after the widest label of a list of labels and values
	labelGap = 3
	// the gap right-aligned text keeps from what stands left of it
	columnGap = 2
	// a centred text has as much room on its left as on its right: the room there is, halved
	sides = 2
)

// The words of the format itself: the section of the manual a page belongs to, its name, and the sections every
// page opens with
const (
	manualSection      = "(1)"
	manualSectionTitle = "User Commands"
	nameSection        = "NAME"
	descriptionSection = "DESCRIPTION"
	bulletMark         = "-"
)

// How a link to a page of the site is written: its text, then its address as a manual between brackets
const (
	addressOpening = "<"
	addressClosing = ">"
)

// The words that stand between the parts of one line of a manual
const (
	dashWord = "—"
	dotWord  = "·"
)

// manualWriter writes one document as a manual page; the site writes its addresses and names the résumé's page.
type manualWriter struct {
	site routes.Site
}

// Manual writes a document as a manual page laid out as text: a title line, NAME and DESCRIPTION from its head
// and the blocks before its first section, then a section for each of its sections, and a closing line.
func Manual(composed document.Document, site routes.Site) string {
	writer := manualWriter{site: site}
	title := strings.ToUpper(writer.pageName(composed.Head.Page)) + manualSection
	lines := []string{columns(title, manualSectionTitle, title)}
	lines = append(lines, writer.head(composed.Head)...)
	lines = append(lines, writer.blocks(composed.Blocks, sectionIndent, true)...)
	lines = append(lines, "", columns(site.Host(), "", title))
	return strings.Join(lines, lineEnd) + lineEnd
}

// pageName is what a page's manual is called: the site's own name for the résumé, its slug for a project's page;
// the index, served in markdown alone, has no manual to name.
func (writer manualWriter) pageName(page document.Page) string {
	switch page.Kind {
	case document.PageResume:
		return writer.site.Host()
	case document.PageProject:
		return page.Slug
	case document.PageNone, document.PageIndex:
	}
	return ""
}

func (writer manualWriter) head(head document.Head) []string {
	name := joinedWords(dashWord, writer.words(head.Title), joinedWords(dotWord, writer.words(head.Subtitle), strings.Fields(head.Place)))
	lines := append([]string{"", nameSection}, paragraph(sectionIndent, 0, name)...)
	lines = append(lines, "", descriptionSection)
	lines = append(lines, paragraph(sectionIndent, 0, writer.words(head.Lead))...)
	if len(head.Ways) > 0 {
		lines = append(append(lines, ""), writer.pairs(sectionIndent, head.Ways)...)
	}
	return lines
}

// blocks writes blocks at an indent, a blank line before each. At the top a section is a section of the page,
// and the blocks before the first one go on with DESCRIPTION; below the top, a section is a title over its blocks.
func (writer manualWriter) blocks(blocks []document.Block, indent int, top bool) []string {
	lines := []string{}
	for _, block := range blocks {
		section, isSection := block.(document.Section)
		if isSection && top {
			lines = append(lines, "", strings.ToUpper(strings.Join(writer.words(section.Title), space)))
			lines = append(lines, trimBlank(writer.blocks(section.Blocks, indent, false))...)
			continue
		}
		lines = append(append(lines, ""), writer.block(block, indent)...)
	}
	return lines
}

func (writer manualWriter) block(block document.Block, indent int) []string {
	switch typed := block.(type) {
	case document.Section:
		title := paragraph(indent, 0, writer.words(typed.Title))
		return append(title, trimBlank(writer.blocks(typed.Blocks, indent+sectionIndent, false))...)
	case document.Paragraph:
		return paragraph(indent, 0, writer.words(typed.Spans))
	case document.Table:
		return writer.table(indent, typed)
	case document.CodeBlock:
		return codeLines(indent, typed.Text)
	case document.Facts:
		return writer.pairs(indent, typed.List)
	case document.Entry:
		return writer.entry(indent, typed)
	case document.Card:
		return writer.card(indent, typed)
	case document.Contents:
		return writer.contents(indent, typed)
	}
	return nil
}

// card writes a card's title and subtitle with its tag at the right, then under them its summary, its page's
// address as a manual, where it has one, and its facts.
func (writer manualWriter) card(indent int, card document.Card) []string {
	lines := titleLine(indent, joinedWords(dashWord, writer.words(card.Title), writer.words(card.Subtitle)), strings.Fields(card.Tag))
	inner := indent + sectionIndent
	lines = append(lines, paragraph(inner, 0, writer.words(card.Summary))...)
	if address := writer.site.Address(card.Page, routes.FormatManual); address != "" {
		lines = append(lines, paragraph(inner, 0, []string{address})...)
	}
	return append(lines, writer.pairs(inner, card.Facts)...)
}

// contents writes each listed page as a bullet: its title and its address as a manual, as a link to it reads, then
// the line on what it is.
func (writer manualWriter) contents(indent int, contents document.Contents) []string {
	lines := []string{}
	for _, listing := range contents.Listings {
		title := writer.words([]document.Span{document.Link{Target: listing.Page, Spans: listing.Title}})
		lines = append(lines, bullet(indent, joinedWords(dashWord, title, writer.words(listing.Note)))...)
	}
	return lines
}

// entry writes an entry's title and subtitle with its period and place at the right, then its points under them.
func (writer manualWriter) entry(indent int, entry document.Entry) []string {
	left := joinedWords(dashWord, writer.words(entry.Title), writer.words(entry.Subtitle))
	right := joinedWords(dotWord, strings.Fields(periodText(entry.Period)), strings.Fields(entry.Place))
	lines := titleLine(indent, left, right)
	for _, point := range entry.Points {
		lines = append(lines, bullet(indent+sectionIndent, writer.words(point.Spans))...)
	}
	return lines
}

// table writes each row as a bullet, its cells in a line.
func (writer manualWriter) table(indent int, table document.Table) []string {
	lines := []string{}
	for _, row := range table.Rows {
		cells := [][]string{}
		for _, cell := range row.Cells {
			cells = append(cells, writer.words(cell.Spans))
		}
		lines = append(lines, bullet(indent, joinedWords(dashWord, cells...))...)
	}
	return lines
}

// pairs writes labels and their values, each value in a column after the widest label.
func (writer manualWriter) pairs(indent int, facts []document.Fact) []string {
	labelWidth := 0
	for _, fact := range facts {
		labelWidth = max(labelWidth, textWidth(fact.Label))
	}
	column := labelWidth + labelGap
	lines := []string{}
	for _, fact := range facts {
		for index, line := range wrap(writer.words(fact.Value), pageWidth-indent-column) {
			lead := strings.Repeat(space, column)
			if index == 0 {
				lead = fact.Label + strings.Repeat(space, column-textWidth(fact.Label))
			}
			lines = append(lines, strings.Repeat(space, indent)+lead+line)
		}
	}
	return lines
}

// titleLine writes left words with right words right-aligned on the same line, or each on lines of their own
// where both do not fit.
func titleLine(indent int, left, right []string) []string {
	leftText, rightText := strings.Join(left, space), strings.Join(right, space)
	gap := pageWidth - indent - textWidth(leftText) - textWidth(rightText)
	if rightText == "" || gap < columnGap {
		return append(paragraph(indent, 0, left), paragraph(indent, 0, right)...)
	}
	return []string{strings.Repeat(space, indent) + leftText + strings.Repeat(space, gap) + rightText}
}

// bullet lays words out as one item of a list at an indent: its mark, then the words, its lines after the first
// standing in under its text.
func bullet(indent int, words []string) []string {
	return paragraph(indent, bulletHang, append([]string{bulletMark}, words...))
}

// paragraph lays words out at an indent; the lines after the first stand in by the hang.
func paragraph(indent, hang int, words []string) []string {
	lines := []string{}
	for index, line := range wrap(words, pageWidth-indent-hang) {
		lead := indent
		if index > 0 {
			lead += hang
		}
		lines = append(lines, strings.Repeat(space, lead)+line)
	}
	return lines
}

// codeLines writes text line for line at an indent, never wrapped.
func codeLines(indent int, code string) []string {
	lines := []string{}
	for line := range strings.SplitSeq(code, lineEnd) {
		lines = append(lines, strings.Repeat(space, indent)+line)
	}
	return lines
}

// columns writes text at the left, the centre and the right of a line of the page's width.
func columns(left, centre, right string) string {
	line := left
	start := max((pageWidth-textWidth(centre))/sides, textWidth(left)+1)
	line += strings.Repeat(space, start-textWidth(line)) + centre
	return line + strings.Repeat(space, max(pageWidth-textWidth(line)-textWidth(right), 1)) + right
}

// wrap lays words out in lines of a width at most; a word longer than the width stands on a line of its own.
func wrap(words []string, width int) []string {
	lines := []string{}
	line := ""
	for _, word := range words {
		switch {
		case line == "":
			line = word
		case textWidth(line)+1+textWidth(word) <= width:
			line += space + word
		default:
			lines = append(lines, line)
			line = word
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// words are the words of spans as a terminal shows them, a line breaking only between two: marks gone, a link
// by its text and, where it leads to a page of the site, its address as a manual; and a command one word, its
// spaces and all, so it is never broken.
func (writer manualWriter) words(spans []document.Span) []string {
	reader := wordReader{words: []string{}, word: ""}
	writer.readSpans(&reader, spans)
	return reader.done()
}

func (writer manualWriter) readSpans(reader *wordReader, spans []document.Span) {
	for _, span := range spans {
		switch typed := span.(type) {
		case document.Text:
			reader.read(typed.Text)
		case document.Code:
			reader.keep(typed.Text)
		case document.Gap:
			reader.read(typed.Text)
		case document.Highlight:
			writer.readSpans(reader, typed.Spans)
		case document.Link:
			writer.readSpans(reader, typed.Spans)
			if page, isPage := typed.Target.(document.Page); isPage {
				reader.read(space + addressOpening + writer.site.Address(page, routes.FormatManual) + addressClosing)
			}
		}
	}
}

// wordReader gathers words from text read piece by piece: a word goes on until a space ends it.
type wordReader struct {
	words []string
	word  string
}

// read takes text, ending a word at each space.
func (reader *wordReader) read(text string) {
	for _, letter := range text {
		if unicode.IsSpace(letter) {
			reader.end()
			continue
		}
		reader.word += string(letter)
	}
}

// keep takes text into the word in hand whole, its spaces with it.
func (reader *wordReader) keep(text string) {
	reader.word += text
}

func (reader *wordReader) end() {
	if reader.word != "" {
		reader.words = append(reader.words, reader.word)
	}
	reader.word = ""
}

func (reader *wordReader) done() []string {
	reader.end()
	return reader.words
}

// joinedWords are the parts that hold words, a word between each two.
func joinedWords(separator string, parts ...[]string) []string {
	words := []string{}
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		if len(words) > 0 {
			words = append(words, separator)
		}
		words = append(words, part...)
	}
	return words
}

// trimBlank drops the blank line a list of blocks opens with, where a title stands right above it.
func trimBlank(lines []string) []string {
	if len(lines) > 0 && lines[0] == "" {
		return lines[1:]
	}
	return lines
}

func textWidth(text string) int {
	return utf8.RuneCountInString(text)
}
