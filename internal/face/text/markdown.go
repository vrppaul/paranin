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
	"strconv"
	"strings"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// How markdown is written here
const (
	headingMark   = "#"
	listMark      = "- "
	orderedMark   = ". "
	quoteMark     = ">"
	breakMark     = "***"
	highlightMark = "**"
	strongMark    = "**"
	emphasisMark  = "*"
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

// How a table with a header is written, as GFM writes one: its cells between bars, the header's under a rule
const (
	cellBar  = "|"
	cellRule = "---"
)

// How a note is written, as GFM writes a footnote: [^1] where a line cites it, [^1]: and its words at the end
const (
	noteOpening    = "[^"
	noteClosing    = "]"
	definitionMark = ":"
)

// How a link is written: [text](address), the address between brackets where it holds a space or a parenthesis; and a
// picture as a link after its mark, its title, the caption, after the address between quotes, which an escape keeps
// from closing it and from opening a reference to a character
const (
	linkTextOpening        = "["
	linkTextClosing        = "]"
	linkAddressOpening     = "("
	linkAddressClosing     = ")"
	enclosedOpening        = "<"
	enclosedClosing        = ">"
	addressBreakingLetters = " ()"
	pictureMark            = "!"
	titleQuote             = `"`
	titleMarkup            = `\"&`
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
	// a title and its place, a tag and a subtitle, and each two names of a project's stack, in markdown and the manual
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

// markdownWriter writes one document: the site writes its addresses, and the escapers keep text, gaps and a picture's
// title from reading as markup.
type markdownWriter struct {
	site        routes.Site
	textEscape  *strings.Replacer
	gapEscape   *strings.Replacer
	titleEscape *strings.Replacer
}

// Markdown writes a document as CommonMark: its head, then each block, a blank line between them, a project's end as
// more sections, and the notes its blocks cite at the end.
func Markdown(composed document.Document, site routes.Site) string {
	writer := markdownWriter{
		site: site, textEscape: escaperOf(inlineMarkup), gapEscape: escaperOf(gapMarkup), titleEscape: escaperOf(titleMarkup),
	}
	chunks := writer.head(composed.Head)
	chunks = writer.blocks(chunks, composed.Blocks, sectionDepth)
	if project := composed.Head.Project; project != nil {
		for _, section := range project.EndSections() {
			chunks = writer.block(chunks, section, sectionDepth)
		}
	}
	chunks = appendChunk(chunks, writer.notes(document.NotesOf(composed.Blocks)))
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

// head writes the title as a heading, then the subtitle and the version a post's reader reads on a line, an older
// version's with the newer one that stands over it, a translation older than its original's with the way to the
// original, where the author is on the next, as the page stands it apart, the languages a post is read in on the next,
// the one read strong and each other a link to the post in it, a post's tags on the next, the lead as a quote, and what
// gets a project and tells about it.
func (writer markdownWriter) head(head document.Head) []string {
	chunks := []string{heading(titleDepth, writer.spans(head.Title))}
	version := joined(dot, writer.text(versionText(head)), writer.spans(versionNote(head)))
	chunks = appendLine(chunks, joined(dot, writer.spans(head.Subtitle), version))
	place := make([]string, 0, len(head.Place.Parts()))
	for _, part := range head.Place.Parts() {
		place = append(place, writer.text(part))
	}
	chunks = appendLine(chunks, joined(dot, place...))
	languages := make([]string, 0, len(head.Languages))
	for _, language := range head.Languages {
		languages = append(languages, writer.spans(language.Spans()))
	}
	chunks = appendLine(chunks, joined(dot, languages...))
	named := head.Tags()
	tags := make([]string, 0, len(named))
	for _, tag := range named {
		tags = append(tags, writer.text(tag))
	}
	chunks = appendLine(chunks, joined(dot, tags...))
	chunks = appendChunk(chunks, quote(lineStart(writer.spans(head.Lead))))
	return writer.project(chunks, head.Project)
}

// project writes what gets a project, each a line of its own: where its code lives, then its install under the site's
// word, then each other way under its note; then what tells about it as a list of facts. What gets it stands in
// lines, not items, since markdown would read two lists in a row as one.
func (writer markdownWriter) project(chunks []string, project *document.Project) []string {
	if project == nil {
		return chunks
	}
	for _, fact := range project.Repository.Facts() {
		chunks = appendLine(chunks, writer.fact(fact))
	}
	if install, installed := project.InstallFact(); installed {
		chunks = appendLine(chunks, writer.fact(install))
	}
	for _, alternative := range project.Alternatives {
		chunks = appendLine(chunks, writer.spans(alternative.Note)+labelEnd+writer.spans([]document.Span{document.Code{Text: alternative.Command}}))
	}
	return appendChunk(chunks, writer.facts(project.AboutFacts(dot)))
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
		return appendChunk(chunks, codeBlock(typed.Text(), typed.Language))
	case document.List:
		return appendChunk(chunks, writer.list(typed, depth))
	case document.Quote:
		return appendChunk(chunks, quote(strings.Join(writer.blocks(nil, typed.Blocks, depth), blankLine)))
	case document.Pause:
		return appendChunk(chunks, breakMark)
	case document.Figure:
		return appendChunk(chunks, writer.figure(typed))
	case document.Widget:
		return appendChunk(chunks, writer.widget(typed))
	case document.Facts:
		return writer.factsBlock(chunks, typed)
	case document.ExperienceHead:
		return writer.blocks(chunks, typed.Blocks(), depth)
	case document.Entry:
		return writer.entry(chunks, typed, depth)
	case document.Points:
		return appendChunk(chunks, writer.points(typed.List))
	case document.Side:
		return writer.side(chunks, typed, depth)
	case document.Clip:
		// a recording as a picture stands, a link by what it shows, its marks kept, which a reader of markdown follows to
		// play it
		return appendLine(chunks, writer.spans([]document.Span{document.Link{Target: typed.Recording, Spans: typed.Shows}}))
	case document.Reference:
		// the details set apart from the story by a thematic break, as a turn a post takes
		return writer.blocks(appendChunk(chunks, breakMark), typed.Blocks(), depth)
	case document.ProjectList:
		return writer.projectList(chunks, typed, depth)
	case document.Contents:
		return appendChunk(chunks, writer.contents(typed))
	case document.Posts:
		chunks = appendLine(appendLine(chunks, writer.spans(typed.Line)), writer.spans(typed.Count))
		for _, post := range typed.List {
			chunks = writer.listedPost(chunks, post, depth)
		}
		return chunks
	case document.Checks, document.Skills:
		// only the hidden pages hold checks and the list of skills, and no host hands them to a text face: a page
		// served with them would fail TestEveryPhraseReachesEveryFormat
	}
	return chunks
}

// listedPost writes a post of the blog's list as an entry is written: under its title, a link to it where it is out;
// the day it came out, or the word saying it comes, its part and its series in a line; its tags on the next; then its
// lead.
func (writer markdownWriter) listedPost(chunks []string, post document.ListedPost, depth int) []string {
	title, when := post.Title, ""
	switch typed := post.Readiness.(type) {
	case document.PostOut:
		title, when = linkedTo(typed.Page, post.Title), dayText(typed.Published)
	case document.ComingPart:
		when = writer.spans(typed.Coming)
	}
	chunks = appendChunk(chunks, heading(depth, writer.spans(title)))
	chunks = appendLine(chunks, joined(dot, when, writer.spans(post.Part), writer.spans(post.Series)))
	tags := make([]string, 0, len(post.Tags))
	for _, tag := range post.Tags {
		tags = append(tags, writer.text(tag))
	}
	chunks = appendLine(chunks, joined(dot, tags...))
	return appendLine(chunks, writer.spans(post.Lead))
}

// linkedTo are spans as a link to a page, where it is one: a listed post's title.
func linkedTo(page document.Page, spans []document.Span) []document.Span {
	if page.Kind == document.PageNone {
		return spans
	}
	return []document.Span{document.Link{Target: page, Spans: spans}}
}

// table writes a table under a header as GFM writes one, the posts' own format; one with no header, as a skill
// group's, each row an item, its cells in a line.
func (writer markdownWriter) table(table document.Table) string {
	if len(table.Header.Cells) > 0 {
		return writer.headedTable(table)
	}
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

// headedTable writes a table as GFM writes one: its header's row, a rule under it, then its rows.
func (writer markdownWriter) headedTable(table document.Table) string {
	rule := make([]string, 0, len(table.Header.Cells))
	for range table.Header.Cells {
		rule = append(rule, cellRule)
	}
	lines := []string{writer.tableRow(table.Header), barred(rule)}
	for _, row := range table.Rows {
		lines = append(lines, writer.tableRow(row))
	}
	return strings.Join(lines, lineEnd)
}

// tableRow writes a row of a table's cells, a bar in a cell escaped, since GFM would end the cell at it.
func (writer markdownWriter) tableRow(row document.Row) string {
	cells := make([]string, 0, len(row.Cells))
	for _, cell := range row.Cells {
		cells = append(cells, strings.ReplaceAll(writer.spans(cell.Spans), cellBar, escapeMark+cellBar))
	}
	return barred(cells)
}

// barred writes cells as a row of a table, a bar before, between and after them.
func barred(cells []string) string {
	return cellBar + space + strings.Join(cells, space+cellBar+space) + space + cellBar
}

// list writes a list's items, each its blocks after its mark, the lines after the first standing in under the mark's
// text. A list nested in an item follows the block before it on the next line, and the list stays tight, its items
// one a line; any other block after another stands a blank line below it, and the list is loose, a blank line
// between each two items, as CommonMark reads them.
func (writer markdownWriter) list(list document.List, depth int) string {
	items := make([]string, 0, len(list.Items))
	separator := lineEnd
	for index, item := range list.Items {
		written := ""
		for place, block := range item.Blocks {
			chunk := strings.Join(writer.block(nil, block, depth), blankLine)
			_, isList := block.(document.List)
			switch {
			case place == 0:
				written = chunk
			case isList:
				written += lineEnd + chunk
			default:
				written += blankLine + chunk
				separator = blankLine
			}
		}
		items = append(items, hung(itemMark(list, index), written))
	}
	return strings.Join(items, separator)
}

// itemMark is the mark an item of a list opens with: a dash, or in a list in order its number.
func itemMark(list document.List, index int) string {
	if !list.Ordered {
		return listMark
	}
	return strconv.Itoa(list.Start+index) + orderedMark
}

// hung writes text after a mark, its lines after the first standing in as far as the mark, so markdown reads them as
// the item's; a blank line stays blank.
func hung(mark, text string) string {
	lines := strings.Split(text, lineEnd)
	indent := strings.Repeat(space, len(mark))
	for index, line := range lines[1:] {
		if line != "" {
			lines[index+1] = indent + line
		}
	}
	return mark + strings.Join(lines, lineEnd)
}

// notes writes the notes a document's blocks cite, as GFM writes footnotes at the end: each a line, by its number.
func (writer markdownWriter) notes(notes []document.Note) string {
	lines := make([]string, 0, len(notes))
	for _, note := range notes {
		lines = append(lines, noteReference(note.Number)+definitionMark+space+writer.spans(note.Spans))
	}
	return strings.Join(lines, lineEnd)
}

// noteReference is how a line cites a note by its number, and how the note's words open at the end.
func noteReference(number int) string {
	return noteOpening + strconv.Itoa(number) + noteClosing
}

// factsBlock writes a block of facts by its kind: details as a list; the ways to reach the author each a paragraph of
// its own, since they follow the author's details and markdown would read two lists in a row as one.
func (writer markdownWriter) factsBlock(chunks []string, facts document.Facts) []string {
	switch facts.Kind {
	case document.FactsWays:
		for _, fact := range facts.List {
			chunks = appendLine(chunks, writer.fact(fact))
		}
		return chunks
	case document.FactsDetails:
	}
	return appendChunk(chunks, writer.facts(facts.List))
}

func (writer markdownWriter) facts(facts []document.Fact) string {
	items := make([]string, 0, len(facts))
	for _, fact := range facts {
		items = append(items, writer.fact(fact))
	}
	return list(items)
}

// fact writes a fact as its label and its value.
func (writer markdownWriter) fact(fact document.Fact) string {
	return writer.text(fact.Label) + labelEnd + writer.spans(fact.Value)
}

// entry writes an entry under its title and subtitle, its period in a line, its technologies in another, then its
// points, and its description, its paragraphs one after another.
func (writer markdownWriter) entry(chunks []string, entry document.Entry, depth int) []string {
	chunks = appendChunk(chunks, heading(depth, joined(dash, writer.spans(entry.Title), writer.spans(entry.Subtitle))))
	chunks = appendLine(chunks, writer.text(periodText(entry.Period)))
	chunks = appendChunk(appendLine(chunks, writer.names(entry.Technologies)), writer.points(entry.Points))
	for _, passage := range entry.Description {
		chunks = appendLine(chunks, writer.spans(passage.Spans))
	}
	return chunks
}

// points writes points, an entry's or a project's section's, as a list.
func (writer markdownWriter) points(points []document.Point) string {
	items := make([]string, 0, len(points))
	for _, point := range points {
		items = append(items, writer.spans(point.Spans))
	}
	return list(items)
}

// names writes names in a line, as written, a dot between two: a role's technologies, what a project is made of.
func (writer markdownWriter) names(names []string) string {
	written := make([]string, 0, len(names))
	for _, name := range names {
		written = append(written, writer.text(name))
	}
	return joined(dot, written...)
}

// projectList writes the projects' line, then each project at its level under its name as a heading: a featured one's
// name a link to its page, its line, its description and what it is made of; an auxiliary one's name a link to its
// repository, and its line; then, after a thematic break, as the page sets them under a rule, each one told in words by
// its name, its line and its description.
func (writer markdownWriter) projectList(chunks []string, list document.ProjectList, depth int) []string {
	chunks = appendLine(chunks, writer.spans(list.Line))
	for _, project := range list.Featured {
		chunks = appendChunk(chunks, heading(depth, writer.spans([]document.Span{document.Link{Target: project.Page, Spans: project.Name}})))
		chunks = appendLine(appendLine(chunks, writer.spans(project.Tagline)), writer.spans(project.Description))
		chunks = appendLine(chunks, writer.names(project.Stack))
	}
	for _, project := range list.Auxiliary {
		chunks = appendChunk(chunks, heading(depth, writer.spans([]document.Span{document.Link{Target: project.Repository, Spans: project.Name}})))
		chunks = appendLine(chunks, writer.spans(project.Tagline))
	}
	if len(list.Told) > 0 {
		chunks = appendChunk(chunks, breakMark)
	}
	for _, project := range list.Told {
		chunks = appendChunk(chunks, heading(depth, writer.spans(project.Name)))
		chunks = appendLine(appendLine(chunks, writer.spans(project.Tagline)), writer.spans(project.Description))
	}
	return chunks
}

// contents writes each listed place as an item of a list.
func (writer markdownWriter) contents(contents document.Contents) string {
	items := make([]string, 0, len(contents.Listings))
	for _, listing := range contents.Listings {
		items = append(items, writer.listing(listing))
	}
	return list(items)
}

// listing writes a listed place: a link to it in markdown, then the line on what it is.
func (writer markdownWriter) listing(listing document.Listing) string {
	title := writer.spans([]document.Span{document.Link{Target: listing.Target, Spans: listing.Title}})
	return joined(labelEnd, title, writer.spans(listing.Note))
}

// side writes what adds to a section after its words: each of its links a paragraph of its own, as the ways to reach
// the author, since they may follow the section's points and markdown would read two lists in a row as one.
func (writer markdownWriter) side(chunks []string, side document.Side, depth int) []string {
	for _, block := range side.Blocks {
		contents, isContents := block.(document.Contents)
		if !isContents {
			chunks = writer.block(chunks, block, depth)
			continue
		}
		for _, listing := range contents.Listings {
			chunks = appendLine(chunks, writer.listing(listing))
		}
	}
	return chunks
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
		case document.Emphasis:
			builder.WriteString(emphasisMark + writer.spans(typed.Spans) + emphasisMark)
		case document.Strong:
			builder.WriteString(strongMark + writer.spans(typed.Spans) + strongMark)
		case document.Link:
			builder.WriteString(link(writer.spans(typed.Spans), linkAddress(writer.site.Address(typed.Target, routes.FormatMarkdown))))
		case document.Note:
			builder.WriteString(noteReference(typed.Number))
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

// link writes a link: its words, then its target, the address and what follows it, between parentheses.
func link(text, target string) string {
	return linkTextOpening + text + linkTextClosing + linkAddressOpening + target + linkAddressClosing
}

// figure writes a picture as markdown shows one, the picture's mark before its captioned link.
func (writer markdownWriter) figure(figure document.Figure) string {
	return pictureMark + writer.captionedLink(figure.Alt, writer.site.Address(figure.Picture, routes.FormatMarkdown), figure.Caption)
}

// widget writes a widget as a picture stands, a captioned link, to the live one on the post's page, which runs it.
func (writer markdownWriter) widget(widget document.Widget) string {
	return writer.captionedLink(widget.Alt, writer.site.Address(widget.Live, routes.FormatChosen), widget.Caption)
}

// captionedLink writes what stands for a picture or a widget: its alt text as the link's words, its address whole, and
// its caption, where it has one, as its title.
func (writer markdownWriter) captionedLink(alt, address, caption string) string {
	target := linkAddress(address)
	if caption != "" {
		target += space + titleQuote + writer.titleEscape.Replace(caption) + titleQuote
	}
	return link(writer.text(alt), target)
}

// linkAddress is an address as a link or a picture writes it: between brackets where a space or a parenthesis would
// end it.
func linkAddress(address string) string {
	if strings.ContainsAny(address, addressBreakingLetters) {
		return enclosedOpening + address + enclosedClosing
	}
	return address
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

// codeBlock fences text with more backticks than any run inside it, and three at least, its language after the
// opening fence.
func codeBlock(code, language string) string {
	fence := strings.Repeat(codeMark, max(shortestFence, longestRun(code, codeMark)+1))
	return fence + language + lineEnd + code + lineEnd + fence
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

// quote writes markdown as a quote, where there is any: every line of it marked, a blank one included, so a quote
// of two paragraphs stays one quote. Each line reads after its mark as it read before it, so a list or code inside
// the quote stays one, and a text that opens no block opens none there.
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
	return quoteMark + space + line
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
