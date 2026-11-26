package text

import (
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/look"
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

// The marks of the format itself: the section of the manual a page belongs to, whose name and those of a page's own
// sections are the site's words in the page's language, and a bullet
const (
	manualSection = "(1)"
	bulletMark    = "-"
)

// The marks of a post's blocks: the bar a quote's lines stand behind, and how far its text stands in from where the
// bar does; a note's number, between brackets, where a line cites it and before its words; and the rule under a
// table's header
const (
	quoteBar    = "│"
	quoteIndent = 2
	noteOpens   = "["
	noteCloses  = "]"
	headerRule  = "─"
)

// A tab in a line of code
const tabLetter = '\t'

// How a link to a page of the site is written: its text, then its address as a manual between brackets; the slash a
// path from the site's root opens with and parts its folders by; and what joins those parts in a manual's name, as
// git-log(1) joins a command's
const (
	addressOpening = "<"
	addressClosing = ">"
	pathRoot       = "/"
	nameJoint      = "-"
)

// The words that stand between the parts of one line of a manual: a title and its subtitle, and a subtitle and its
// place or each two of a post's tags
const (
	dashWord = "—"
	dotWord  = "·"
)

// manualWriter writes one document as a manual page; the site writes its addresses and names the résumé's page, and the
// manual's own words name its sections in the page's language.
type manualWriter struct {
	site routes.Site
	own  compose.ManualWords
}

// Manual writes a document as a manual page laid out as text, in the manual's own words of its language: a title line
// naming the manual's section, NAME from its head, DESCRIPTION from its head and the blocks that are no section, then a
// section for each of its sections and of a project's end, NOTES where its blocks cite any, each section's
// name in capitals, and a closing line, the version a post's reader reads in its middle, its day or a draft's word,
// where man(1) writes a page's date.
func Manual(composed document.Document, site routes.Site, own compose.ManualWords) string {
	writer := manualWriter{site: site, own: own}
	title := strings.ToUpper(writer.pageName(composed.Head.Page)) + manualSection
	others, sections := document.SectionsOf(composed.Blocks)
	lines := []string{columns(title, own.Section, title)}
	lines = append(lines, writer.name(composed.Head)...)
	lines = append(lines, writer.description(composed.Head, others)...)
	for _, section := range sections {
		lines = append(lines, writer.pageSection(section)...)
	}
	if project := composed.Head.Project; project != nil {
		for _, section := range project.EndSections() {
			lines = append(lines, writer.pageSection(section)...)
		}
	}
	lines = append(lines, writer.notes(document.NotesOf(composed.Blocks))...)
	lines = append(lines, "", columns(site.Host(), versionText(composed.Head), title))
	return strings.Join(lines, lineEnd) + lineEnd
}

// pageName is what a page's manual is called: the site's own name for the résumé, its slug for a project's page and
// a post, its path from the root for the blog's list, its folders joined by a hyphen (blog-series), and a filtered
// list's the list's it filters, as man(1) names a page by its name and never by how it was asked for; the index,
// served in markdown alone, and the hidden pages, served in neither, have no manual to name.
func (writer manualWriter) pageName(page document.Page) string {
	switch page.Kind {
	case document.PageResume:
		return writer.site.Host()
	case document.PageProject, document.PagePost:
		return page.Slug
	case document.PageBlog, document.PageBlogSeries:
		return strings.ReplaceAll(strings.TrimPrefix(routes.Path(page, routes.FormatChosen), pathRoot), pathRoot, nameJoint)
	case document.PageTagFilter:
		list, _ := page.Unfiltered()
		return writer.pageName(list)
	case document.PageNone, document.PageIndex, document.PageHealth, document.PageSkills:
	}
	return ""
}

// name writes the name section: the title with its subtitle and place.
func (writer manualWriter) name(head document.Head) []string {
	name := joinedWords(dashWord, writer.words(head.Title), joinedWords(dotWord, writer.words(head.Subtitle), strings.Fields(head.Place)))
	return append([]string{"", strings.ToUpper(writer.own.Name)}, paragraph(sectionIndent, 0, name)...)
}

// description writes the description section: what the version read says under its day, that a newer version stands
// or that the original changed since a translation, first, then the lead, the languages a post is read in, each other
// than the one read with the address of its manual, a post's or the list's tags, what gets a project and tells about
// it, then the blocks that are no section; none where there is nothing to describe.
func (writer manualWriter) description(head document.Head, others []document.Block) []string {
	lines := paragraph(sectionIndent, 0, writer.words(head.Lead))
	if note := versionNote(head); len(note) > 0 {
		lines = append(append(paragraph(sectionIndent, 0, writer.words(note)), ""), lines...)
	}
	if len(head.Languages) > 0 {
		languages := make([][]string, 0, len(head.Languages))
		for _, language := range head.Languages {
			languages = append(languages, writer.words(language.Spans()))
		}
		lines = append(append(lines, ""), paragraph(sectionIndent, 0, joinedWords(dotWord, languages...))...)
	}
	if named := head.Tags(); len(named) > 0 {
		tags := make([][]string, 0, len(named))
		for _, tag := range named {
			tags = append(tags, strings.Fields(tag))
		}
		lines = append(append(lines, ""), paragraph(sectionIndent, 0, joinedWords(dotWord, tags...))...)
	}
	lines = append(lines, writer.project(head.Project)...)
	lines = trimBlank(append(lines, writer.blocks(others, sectionIndent)...))
	if len(lines) == 0 {
		return nil
	}
	return append([]string{"", strings.ToUpper(writer.own.Description)}, lines...)
}

// project writes what gets a project and what tells about it, a blank line before each part: where its code lives
// and its install as labelled pairs; each other way under its note, its command standing in under the note, since a
// note is too long a label to set the pairs' column by; then the facts that tell about it. None on a page that is no
// project's.
func (writer manualWriter) project(project *document.Project) []string {
	if project == nil {
		return nil
	}
	gets := project.Repository.Facts()
	if install, installed := project.InstallFact(); installed {
		gets = append(gets, install)
	}
	lines := append([]string{""}, writer.pairs(sectionIndent, gets)...)
	for _, alternative := range project.Alternatives {
		lines = append(append(lines, ""), paragraph(sectionIndent, 0, writer.words(alternative.Note))...)
		lines = append(lines, codeLines(sectionIndent+bulletHang, alternative.Command)...)
	}
	if about := project.AboutFacts(dot); len(about) > 0 {
		lines = append(append(lines, ""), writer.pairs(sectionIndent, about)...)
	}
	return lines
}

// pageSection writes a section of the page: its title in capitals over its blocks.
func (writer manualWriter) pageSection(section document.Section) []string {
	lines := []string{"", strings.ToUpper(strings.Join(writer.words(section.Title), space))}
	return append(lines, trimBlank(writer.blocks(section.Blocks, sectionIndent))...)
}

// blocks writes blocks at an indent, a blank line before each; a section among them is a title over its blocks.
func (writer manualWriter) blocks(blocks []document.Block, indent int) []string {
	lines := []string{}
	for _, block := range blocks {
		lines = append(append(lines, ""), writer.block(block, indent)...)
	}
	return lines
}

func (writer manualWriter) block(block document.Block, indent int) []string {
	switch typed := block.(type) {
	case document.Section:
		title := paragraph(indent, 0, writer.words(typed.Title))
		return append(title, trimBlank(writer.blocks(typed.Blocks, indent+sectionIndent))...)
	case document.Paragraph:
		return paragraph(indent, 0, writer.words(typed.Spans))
	case document.Table:
		return writer.table(indent, typed)
	case document.CodeBlock:
		return codeLines(indent, typed.Text())
	case document.List:
		return writer.list(indent, typed)
	case document.Quote:
		return writer.quote(indent, typed)
	case document.Pause:
		return pause(indent)
	case document.Figure:
		return standIn(indent, typed.Alt, writer.site.Address(typed.Picture, routes.FormatManual), typed.Caption)
	case document.Widget:
		// the live widget is the post's page, which runs it
		return standIn(indent, typed.Alt, writer.site.Address(typed.Live, routes.FormatChosen), typed.Caption)
	case document.Facts:
		return writer.pairs(indent, typed.List)
	case document.ExperienceHead:
		return trimBlank(writer.blocks(typed.Blocks(), indent))
	case document.Entry:
		return writer.entry(indent, typed)
	case document.Points:
		return writer.bullets(indent, typed.List)
	case document.Card:
		return writer.card(indent, typed)
	case document.Contents:
		return writer.contents(indent, typed)
	case document.Posts:
		return writer.posts(indent, typed)
	case document.Checks, document.Skills:
		// only the hidden pages hold checks and the list of skills, and no host hands them to a text face: a page
		// served with them would fail TestEveryPhraseReachesEveryFormat
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

// posts writes a group of the blog's list: a series' line and how many posts it holds, then each post, a blank line
// before each.
func (writer manualWriter) posts(indent int, posts document.Posts) []string {
	lines := paragraph(indent, 0, writer.words(posts.Line))
	lines = append(lines, paragraph(indent, 0, writer.words(posts.Count))...)
	for _, post := range posts.List {
		lines = append(append(lines, ""), writer.listedPost(indent, post)...)
	}
	return lines
}

// listedPost writes a post of the blog's list as a card is written: its title with the day it came out at the right,
// or the word saying it comes; then under them its lead, its part and its series, its tags, and its page's address as
// a manual, where it is out.
func (writer manualWriter) listedPost(indent int, post document.ListedPost) []string {
	when, address := []string{}, ""
	switch typed := post.Readiness.(type) {
	case document.PostOut:
		when, address = strings.Fields(dayText(typed.Published)), writer.site.Address(typed.Page, routes.FormatManual)
	case document.ComingPart:
		when = writer.words(typed.Coming)
	}
	lines := titleLine(indent, writer.words(post.Title), when)
	inner := indent + sectionIndent
	lines = append(lines, paragraph(inner, 0, writer.words(post.Lead))...)
	lines = append(lines, paragraph(inner, 0, joinedWords(dotWord, writer.words(post.Part), writer.words(post.Series)))...)
	tags := make([][]string, 0, len(post.Tags))
	for _, tag := range post.Tags {
		tags = append(tags, strings.Fields(tag))
	}
	lines = append(lines, paragraph(inner, 0, joinedWords(dotWord, tags...))...)
	return append(lines, paragraph(inner, 0, strings.Fields(address))...)
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

// entry writes an entry's title and subtitle with its period at the right, its technologies in a line under them, then
// its points, and its description, each of its paragraphs a row of nothing below what stands before it.
func (writer manualWriter) entry(indent int, entry document.Entry) []string {
	left := joinedWords(dashWord, writer.words(entry.Title), writer.words(entry.Subtitle))
	lines := titleLine(indent, left, strings.Fields(periodText(entry.Period)))
	technologies := make([][]string, 0, len(entry.Technologies))
	for _, technology := range entry.Technologies {
		technologies = append(technologies, strings.Fields(technology))
	}
	lines = append(lines, paragraph(indent+sectionIndent, 0, joinedWords(dotWord, technologies...))...)
	lines = append(lines, writer.bullets(indent+sectionIndent, entry.Points)...)
	for _, passage := range entry.Description {
		lines = append(append(lines, ""), paragraph(indent+sectionIndent, 0, writer.words(passage.Spans))...)
	}
	return lines
}

// table writes a table under a header in columns, as tbl(1) sets one, where its cells fit the page each on a line;
// a table with no header, or one too wide, as a bullet a row, its cells in a line, the header's first.
func (writer manualWriter) table(indent int, table document.Table) []string {
	rows := table.Rows
	if len(table.Header.Cells) > 0 {
		rows = append([]document.Row{table.Header}, rows...)
	}
	cellWords := make([][][]string, 0, len(rows))
	cellTexts := make([][]string, 0, len(rows))
	for _, row := range rows {
		words, texts := [][]string{}, []string{}
		for _, cell := range row.Cells {
			words = append(words, writer.words(cell.Spans))
			texts = append(texts, strings.Join(words[len(words)-1], space))
		}
		cellWords, cellTexts = append(cellWords, words), append(cellTexts, texts)
	}
	widths := cellWidths(cellTexts)
	if len(table.Header.Cells) > 0 && indent+widthOfColumns(widths) <= pageWidth {
		return columnedRows(indent, cellTexts, widths)
	}
	lines := []string{}
	for _, words := range cellWords {
		lines = append(lines, bullet(indent, joinedWords(dashWord, words...))...)
	}
	return lines
}

// columnedRows lays rows of cells out in columns of widths at an indent, a rule under the first row, the header.
func columnedRows(indent int, cells [][]string, widths []int) []string {
	rule := make([]string, 0, len(widths))
	for _, width := range widths {
		rule = append(rule, strings.Repeat(headerRule, width))
	}
	lines := []string{}
	for index, row := range cells {
		lines = append(lines, columnedLine(indent, row, widths))
		if index == 0 {
			lines = append(lines, columnedLine(indent, rule, widths))
		}
	}
	return lines
}

// columnedLine is one row of cells, each but the last padded to its column's width, a gap between each two.
func columnedLine(indent int, row []string, widths []int) string {
	var line strings.Builder
	line.WriteString(strings.Repeat(space, indent))
	for place, text := range row {
		line.WriteString(text)
		if place < len(row)-1 {
			line.WriteString(strings.Repeat(space, widths[place]-textWidth(text)+columnGap))
		}
	}
	return line.String()
}

// cellWidths are the widths of the rows' columns, each its widest cell's.
func cellWidths(cells [][]string) []int {
	widths := []int{}
	for _, row := range cells {
		for place, text := range row {
			if place == len(widths) {
				widths = append(widths, 0)
			}
			widths[place] = max(widths[place], textWidth(text))
		}
	}
	return widths
}

// widthOfColumns is how wide columns of widths stand together, a gap between each two.
func widthOfColumns(widths []int) int {
	total := 0
	for _, width := range widths {
		total += width
	}
	return total + max(len(widths)-1, 0)*columnGap
}

// list writes a list's items, each its blocks after its mark, a dash or its number, the lines after the first
// standing in under the item's text, as far as the widest mark takes: a list nested in an item right under the block
// before it, as markdown writes it, any other block a blank line below.
func (writer manualWriter) list(indent int, list document.List) []string {
	marks := listMarks(list)
	hang := bulletHang
	for _, mark := range marks {
		hang = max(hang, textWidth(mark)+1)
	}
	lines := []string{}
	for index, item := range list.Items {
		inner := writer.itemBlocks(item.Blocks, indent+hang)
		first := strings.Repeat(space, indent) + marks[index]
		if len(inner) > 0 {
			first += strings.Repeat(space, hang-textWidth(marks[index])) + strings.TrimPrefix(inner[0], strings.Repeat(space, indent+hang))
			inner = inner[1:]
		}
		lines = append(append(lines, first), inner...)
	}
	return lines
}

// listMarks are the marks a list's items open with, in their order: a dash each, or each its number.
func listMarks(list document.List) []string {
	marks := make([]string, 0, len(list.Items))
	for index := range list.Items {
		mark := bulletMark
		if list.Ordered {
			mark = list.Number(index)
		}
		marks = append(marks, mark)
	}
	return marks
}

// itemBlocks writes an item's blocks at an indent: a list nested in it right under the block before, any other block
// a blank line below.
func (writer manualWriter) itemBlocks(blocks []document.Block, indent int) []string {
	lines := []string{}
	for place, block := range blocks {
		if _, isList := block.(document.List); place > 0 && !isList {
			lines = append(lines, "")
		}
		lines = append(lines, writer.block(block, indent)...)
	}
	return lines
}

// quote writes a quote's blocks behind a bar, its text standing in from it.
func (writer manualWriter) quote(indent int, quote document.Quote) []string {
	inner := trimBlank(writer.blocks(quote.Blocks, indent+quoteIndent))
	bar := strings.Repeat(space, indent) + quoteBar
	lines := make([]string, 0, len(inner))
	for _, line := range inner {
		lines = append(lines, bar+strings.TrimPrefix(line, strings.Repeat(space, indent+1)))
	}
	return lines
}

// pause writes a pause as the page draws it: the three stars in the middle of the text's column, from the indent to
// the page's edge, a row of air above and below them.
func pause(indent int) []string {
	room := pageWidth - indent - textWidth(look.PauseMark)
	return []string{"", strings.Repeat(space, indent+room/sides) + look.PauseMark, ""}
}

// standIn writes what stands for a picture or a widget, which a terminal shows neither of, each part from a line of its
// own: its alt text, its address as a link to a page of the site writes one, and its caption.
func standIn(indent int, alt, address, caption string) []string {
	enclosed := addressOpening + address + addressClosing
	lines := slices.Concat(paragraph(indent, 0, strings.Fields(alt)), paragraph(indent, 0, []string{enclosed}))
	return append(lines, paragraph(indent, 0, strings.Fields(caption))...)
}

// notes writes the notes the page's blocks cite, in a section of their own at its end, each its number and then
// its words, the lines after the first standing in under them; none where the page cites none.
func (writer manualWriter) notes(notes []document.Note) []string {
	if len(notes) == 0 {
		return nil
	}
	lines := []string{"", strings.ToUpper(writer.own.Notes)}
	for _, note := range notes {
		mark := noteMark(note.Number)
		lines = append(lines, paragraph(sectionIndent, textWidth(mark)+1, append([]string{mark}, writer.words(note.Spans)...))...)
	}
	return lines
}

// noteMark is a note's number as a line cites it and as its words open.
func noteMark(number int) string {
	return noteOpens + strconv.Itoa(number) + noteCloses
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

// bullets lays points out at an indent, an entry's or a project's section's, each a bullet.
func (writer manualWriter) bullets(indent int, points []document.Point) []string {
	lines := []string{}
	for _, point := range points {
		lines = append(lines, bullet(indent, writer.words(point.Spans))...)
	}
	return lines
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

// codeLines writes text line for line at an indent, never wrapped, a tab as the spaces to its next stop.
func codeLines(indent int, code string) []string {
	lines := []string{}
	for line := range strings.SplitSeq(code, lineEnd) {
		lines = append(lines, strings.Repeat(space, indent)+expandedTabs(line))
	}
	return lines
}

// expandedTabs is a line of code with each tab the spaces to the next tab stop, counted from the line's start, the
// stops as far apart as the page's: a tab reaches no terminal, whose stops the page's indent would move.
func expandedTabs(line string) string {
	var builder strings.Builder
	column := 0
	for _, letter := range line {
		if letter != tabLetter {
			builder.WriteRune(letter)
			column++
			continue
		}
		stop := look.TabStop - column%look.TabStop
		builder.WriteString(strings.Repeat(space, stop))
		column += stop
	}
	return builder.String()
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

// siteAddress is the manual a link inside the site leads to, a section's its page's, since curl and man read no anchor.
func (writer manualWriter) siteAddress(target document.Target) string {
	switch typed := target.(type) {
	case document.Page:
		return writer.site.Address(typed, routes.FormatManual)
	case document.Anchor:
		return writer.site.Address(typed.Page, routes.FormatManual)
	case document.Address, document.Mail, document.File, document.Feed:
	}
	return ""
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
		case document.Emphasis:
			writer.readSpans(reader, typed.Spans)
		case document.Strong:
			writer.readSpans(reader, typed.Spans)
		case document.Note:
			reader.keep(noteMark(typed.Number))
		case document.Link:
			writer.readSpans(reader, typed.Spans)
			if address := writer.siteAddress(typed.Target); address != "" {
				reader.read(space + addressOpening + address + addressClosing)
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

// textWidth is how many columns of a terminal a text of the manual takes.
//
//wrapper:the one place the manual says how it measures a text, which every line it lays out reads
func textWidth(text string) int {
	return utf8.RuneCountInString(text)
}
