package compose

import (
	"slices"
	"strings"

	"github.com/yuin/goldmark/ast"
	extensionast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// The level a post's body and a container's blocks stand at, above every heading's, so each heading opens a section;
// and the level goldmark gives emphasis that is strong
const (
	bodyLevel   = 0
	strongLevel = 2
)

// What a paragraph's line is joined to the next with where its author broke it, and what ends a line of code, which
// the code's last line drops
const (
	lineJoint   = " "
	codeLineEnd = "\n"
)

// Posts are the posts a snapshot serves, each a document, in the snapshot's order: those published, and those
// unlisted, whose address serves them though no list names them; never a draft or a preview.
func Posts(snapshot *content.Snapshot) []document.Document {
	read := blogOf(snapshot)
	documents := []document.Document{}
	for _, post := range snapshot.Posts {
		if served(post.Status) {
			documents = append(documents, postDocument(post, read, snapshot.Words))
		}
	}
	return documents
}

// served is whether a post of a status is served at its address. exhaustive holds the switch to every status.
func served(status content.Status) bool {
	switch status {
	case content.StatusPublished, content.StatusUnlisted:
		return true
	case content.StatusDraft, content.StatusPreview:
	}
	return false
}

// listed is whether a post of a status is named by the blog's lists and its series' parts. exhaustive holds the switch
// to every status.
func listed(status content.Status) bool {
	switch status {
	case content.StatusPublished:
		return true
	case content.StatusDraft, content.StatusPreview, content.StatusUnlisted:
	}
	return false
}

// postDocument is a post as every face draws it, where it stands in the blog: its head, then its body walked from the
// tree the load parsed, which nothing here writes to.
func postDocument(post content.Post, read blog, words content.Words) document.Document {
	body := bodyReader{body: post.Body, notes: map[int][]document.Span{}}
	// readNote returns no error, so the walk ends with none
	_ = ast.Walk(post.Body.Tree, body.readNote)
	return document.Document{Head: postHead(post, read, words), Blocks: body.containedBlocks(post.Body.Tree)}
}

// postHead is a post's head: its page by its slug, its number, the language of the site's words, its title, its
// description as its lead and its page's description, its tags, when it came out, and where it stands in the blog.
func postHead(post content.Post, read blog, words content.Words) document.Head {
	description := textSpans(post.Description)
	return document.Head{
		Page:        postPage(post),
		Number:      post.ID,
		Language:    words.Language(),
		Greeting:    "",
		Title:       textSpans(post.Title),
		Subtitle:    nil,
		Place:       "",
		Tags:        slices.Clone(post.Tags),
		Lead:        description,
		Published:   post.Published.At(),
		Repository:  nil,
		Post:        read.placeOf(post),
		Description: description,
	}
}

// bodyReader reads a post's body into blocks: the body, and the words of each footnote by its number, read first,
// which every citation of it carries.
type bodyReader struct {
	body  content.Markdown
	notes map[int][]document.Span
}

// readNote keeps the words of a footnote the walk meets, by its number; the load holds a footnote to one paragraph.
func (reader bodyReader) readNote(node ast.Node, entering bool) (ast.WalkStatus, error) {
	footnote, isFootnote := node.(*extensionast.Footnote)
	if !isFootnote || !entering {
		return ast.WalkContinue, nil
	}
	reader.notes[footnote.Index] = reader.spans(footnote.FirstChild())
	return ast.WalkSkipChildren, nil
}

// containedBlocks are the blocks a node holds: the body's, a quote's, a list item's.
func (reader bodyReader) containedBlocks(parent ast.Node) []document.Block {
	nodes := []ast.Node{}
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		nodes = append(nodes, child)
	}
	blocks, _ := reader.nested(nodes, bodyLevel)
	return blocks
}

// nested are the blocks of nodes under a heading of a level, and the nodes after them: each heading deeper than the
// level opens a section holding what follows it up to the next heading of its own level or above, where the run
// under the level ends.
func (reader bodyReader) nested(nodes []ast.Node, level int) ([]document.Block, []ast.Node) {
	blocks := []document.Block{}
	for len(nodes) > 0 {
		heading, isHeading := nodes[0].(*ast.Heading)
		switch {
		case isHeading && heading.Level <= level:
			return blocks, nodes
		case isHeading:
			inside, rest := reader.nested(nodes[1:], heading.Level)
			blocks = append(blocks, document.Section{Title: reader.spans(heading), Blocks: inside})
			nodes = rest
		default:
			if block, isBlock := reader.block(nodes[0]); isBlock {
				blocks = append(blocks, block)
			}
			nodes = nodes[1:]
		}
	}
	return blocks, nil
}

// block is the block a node of the body is, of each block kind content.PostKinds names; none for the list of
// footnotes, whose words stand where a line cites them, and a link's definition, which the links it names carry. A
// heading the nesting reads, and the parts of a list or a table their own block.
func (reader bodyReader) block(node ast.Node) (document.Block, bool) {
	switch typed := node.(type) {
	case *ast.Paragraph, *ast.TextBlock:
		return document.Paragraph{Spans: reader.spans(node)}, true
	case *ast.FencedCodeBlock:
		return document.CodeBlock{Text: reader.code(typed.Lines()), Language: reader.body.Language(typed)}, true
	case *ast.CodeBlock:
		return document.CodeBlock{Text: reader.code(typed.Lines()), Language: ""}, true
	case *ast.Blockquote:
		return document.Quote{Blocks: reader.containedBlocks(node)}, true
	case *ast.ThematicBreak:
		return document.Pause{}, true
	case *ast.List:
		return reader.list(typed), true
	case *extensionast.Table:
		return reader.table(typed), true
	case *extensionast.FootnoteList, *ast.LinkReferenceDefinition:
	}
	return nil, false
}

// list is a list of the body, its items each holding blocks, numbered from its start where it is in order.
func (reader bodyReader) list(list *ast.List) document.List {
	items := []document.Item{}
	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		items = append(items, document.Item{Blocks: reader.containedBlocks(item)})
	}
	return document.List{Ordered: list.IsOrdered(), Start: list.Start, Items: items}
}

// table is a table of the body: its header, then its rows.
func (reader bodyReader) table(table *extensionast.Table) document.Table {
	header, rows := document.Row{Cells: nil}, []document.Row{}
	for row := table.FirstChild(); row != nil; row = row.NextSibling() {
		if row.Kind() == extensionast.KindTableHeader {
			header = reader.row(row)
			continue
		}
		rows = append(rows, reader.row(row))
	}
	return document.Table{Header: header, Rows: rows}
}

func (reader bodyReader) row(row ast.Node) document.Row {
	cells := []document.Cell{}
	for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
		cells = append(cells, document.Cell{Kind: document.CellText, Spans: reader.spans(cell)})
	}
	return document.Row{Cells: cells}
}

// code is the text of a block of code, its lines as written, the last one's ending dropped.
func (reader bodyReader) code(lines *text.Segments) string {
	var builder strings.Builder
	for index := range lines.Len() {
		line := lines.At(index)
		builder.Write(line.Value(reader.body.Source))
	}
	return strings.TrimSuffix(builder.String(), codeLineEnd)
}

// spans are the spans of the inline nodes a node holds, a run of text one span.
func (reader bodyReader) spans(parent ast.Node) []document.Span {
	spans := []document.Span{}
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		spans = reader.appendSpan(spans, child)
	}
	return spans
}

// appendSpan adds the span an inline node is, of each inline kind content.PostKinds names: text as a reader reads
// it, code, emphasis or strong text, a link and its target, and a citation of a footnote as the note itself.
func (reader bodyReader) appendSpan(spans []document.Span, node ast.Node) []document.Span {
	switch typed := node.(type) {
	case *ast.Text:
		return appendJoined(spans, reader.text(typed))
	case *ast.CodeSpan:
		return append(spans, document.Code{Text: reader.codeSpan(typed)})
	case *ast.Emphasis:
		if typed.Level == strongLevel {
			return append(spans, document.Strong{Spans: reader.spans(typed)})
		}
		return append(spans, document.Emphasis{Spans: reader.spans(typed)})
	case *ast.Link:
		return append(spans, document.Link{Target: addressTarget(content.ResolvedAddress(typed.Destination)), Spans: reader.spans(typed)})
	case *ast.AutoLink:
		label := string(typed.Label(reader.body.Source))
		return append(spans, document.Link{Target: autoLinkTarget(typed.AutoLinkType, label), Spans: textSpans(label)})
	case *extensionast.FootnoteLink:
		return append(spans, document.Note{Number: typed.Index, Spans: reader.notes[typed.Index]})
	}
	return spans
}

// text is a run of text as a reader reads it, a line its author broke joined to the next.
func (reader bodyReader) text(written *ast.Text) string {
	resolved := content.Resolved(written.Segment.Value(reader.body.Source))
	if written.SoftLineBreak() {
		resolved += lineJoint
	}
	return resolved
}

// codeSpan is the text of a code span as written, a line it is broken at joined to the next, as CommonMark reads it.
func (reader bodyReader) codeSpan(span *ast.CodeSpan) string {
	var builder strings.Builder
	for child := span.FirstChild(); child != nil; child = child.NextSibling() {
		if written, isText := child.(*ast.Text); isText {
			value := string(written.Segment.Value(reader.body.Source))
			builder.WriteString(strings.Replace(value, codeLineEnd, lineJoint, 1))
		}
	}
	return builder.String()
}

// appendJoined adds text to the spans, joined to the text that ends them, where they end in text.
func appendJoined(spans []document.Span, text string) []document.Span {
	if last := len(spans) - 1; last >= 0 {
		if before, isText := spans[last].(document.Text); isText {
			spans[last] = document.Text{Text: before.Text + text}
			return spans
		}
	}
	return append(spans, document.Text{Text: text})
}

// addressTarget is where a link's address leads: a mailbox where it writes to one, an address of the web otherwise,
// which the load holds a link's address to.
func addressTarget(address string) document.Target {
	if mailbox, isMail := content.Mailbox(address); isMail {
		return document.Mail{Address: mailbox}
	}
	return document.Address{URL: address}
}

// autoLinkTarget is where an address written between angle brackets leads: a mailbox, or an address of the web.
func autoLinkTarget(kind ast.AutoLinkType, written string) document.Target {
	if kind == ast.AutoLinkEmail {
		return document.Mail{Address: written}
	}
	return document.Address{URL: written}
}
