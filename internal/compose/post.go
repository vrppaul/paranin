package compose

import (
	"cmp"
	"maps"
	"strings"

	"github.com/yuin/goldmark/ast"
	extensionast "github.com/yuin/goldmark/extension/ast"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
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

// The name of a heading whose title holds no letter or digit, which document.SlugOf makes none of
const unnamedHeading = "heading"

// composedPosts are the posts the blog serves its audience, each a document of its current version, in the snapshot's
// order; each link inside the site read into the page or the feed at its path among paths.
func composedPosts(prepared *Prepared, read blog, paths map[string]document.Target) []document.Document {
	documents := []document.Document{}
	for _, post := range read.served {
		documents = append(documents, postDocument(currentOf(post), read, prepared, paths))
	}
	return documents
}

// PostPicture is a picture of a post the site serves: its file beside the post, and the picture.
type PostPicture struct {
	File    document.File
	Picture content.Picture
}

// picturesOf are the pictures of the posts the blog serves its audience, each at its file beside its post, in no
// order.
func picturesOf(read blog) []PostPicture {
	pictures := []PostPicture{}
	for _, post := range read.served {
		pictures = append(pictures, postPictures(post)...)
	}
	return pictures
}

// postPictures are a post's pictures, each at its file beside it, in no order.
func postPictures(post content.Post) []PostPicture {
	pictures := make([]PostPicture, 0, len(post.Pictures))
	for name, picture := range post.Pictures {
		pictures = append(pictures, PostPicture{File: document.File{Page: postPage(post), Name: name}, Picture: picture})
	}
	return pictures
}

// Preview is a post the site serves its audience at its preview link alone (docs/FOUNDATION.md §5.4, Status): the post
// as its link's reader reads it, and its pictures, each at its file beside it; and the post as the content holds it,
// whose links Check reads.
type Preview struct {
	Post     document.Document
	Pictures []PostPicture
	written  content.Post
}

// previewsOf are the posts the blog serves its audience at their preview links, in the snapshot's order, each composed
// as composedPosts composes a post, its links read among the pages served and its own page.
func previewsOf(prepared *Prepared, read blog, paths map[string]document.Target) []Preview {
	previews := []Preview{}
	for _, post := range read.previewed {
		previews = append(previews, Preview{
			Post: postDocument(currentOf(post), read, prepared, withOwnPage(paths, postPage(post))), Pictures: postPictures(post), written: post,
		})
	}
	return previews
}

// withOwnPage are the pages served and the feeds beside the page of a post read at its preview link: its link to a
// section of its own leads there, though its own address serves nothing yet.
func withOwnPage(paths map[string]document.Target, page document.Page) map[string]document.Target {
	own := maps.Clone(paths)
	own[routes.Path(page, routes.FormatChosen)] = page
	return own
}

// postHeads are the heads of the posts composedPosts makes and of their older versions olderDocuments makes, their
// bodies left unread: where the site serves each version of each post.
func postHeads(read blog) []document.Head {
	heads := []document.Head{}
	for _, post := range read.served {
		heads = append(heads, postHead(currentOf(post), read))
		for _, version := range read.olderVersionsOf(post) {
			heads = append(heads, postHead(version, read))
		}
	}
	return heads
}

// postDocument is a version of a post as every face draws it, where the post stands in the blog: its head, then its
// body walked from the tree the load parsed, which nothing here writes to, its code as Prepare split it, its pictures
// the post's, each heading named clear of the page's other places and each link inside the site read into the page or
// the feed at its path among paths.
func postDocument(version postVersion, read blog, prepared *Prepared, paths map[string]document.Target) document.Document {
	snapshot, text := prepared.snapshot, version.text
	head := postHead(version, read)
	body := bodyReader{
		body: text.Body, page: head.Page, pictures: version.post.Pictures, links: text.SiteLinks, paths: paths, codes: prepared.codes,
		notes: map[int][]document.Span{}, names: namesHeld(namesBeside(version.post, read, snapshot)),
	}
	// readNote returns no error, so the walk ends with none
	_ = ast.Walk(text.Body.Tree, body.readNote)
	return document.Document{Head: head, Blocks: body.containedBlocks(text.Body.Tree)}
}

// namesBeside are the names a post's page offers its places by but its headings, which theirs keep clear of on every
// face, so each name on the page leads to one place: its windows', each project's page's, the window of keys', each
// part's its series names, out or not, so a heading's anchor holds as a part comes out, and each of its versions'.
func namesBeside(post content.Post, read blog, snapshot *content.Snapshot) []string {
	taken := []string{KeysPlace}
	list := document.PageOf(document.PageBlog).InLanguage(read.words.Translation())
	for _, window := range read.listWords().BlogWindows(list, postPage(post)) {
		taken = append(taken, window.Name)
	}
	for _, project := range projectsWithPages(snapshot.Profile.Projects) {
		taken = append(taken, project.Slug)
	}
	for _, version := range read.versionsOf(post, postPage(post)) {
		taken = append(taken, version.Name)
	}
	return append(taken, read.seriesOf[post.Slug].Posts...)
}

// postHead is the head of a version of a post: its page, the post's own or its revision's, its number, the language of
// the blog's words, the version's title, its description as its lead and its page's description, the post's tags as
// the filters of the list they are, the version read, where the post stands in the blog, the feeds that follow it, and
// the languages it is read in.
func postHead(version postVersion, read blog) document.Head {
	post, description, words := version.post, textSpans(version.text.Description), read.words
	languages := read.languagesOf(version)
	return document.Head{
		Page:        version.page,
		Number:      post.ID,
		Language:    words.Language(),
		Greeting:    "",
		Title:       textSpans(version.text.Title),
		Subtitle:    nil,
		Place:       "",
		Filters:     read.postFilters(post),
		MadeOf:      nil,
		Lead:        description,
		Version:     read.versionOf(version, words, languages),
		Repository:  nil,
		Post:        read.placeOf(post),
		Description: description,
		Feeds:       read.feedLinks(read.seriesShownOn(post)),
		Languages:   languages,
	}
}

// bodyReader reads a post's body into blocks: the body; the post's page and its pictures, by their names, which its
// figures stand beside and are sized by; its links inside the site, as the load read them, by their nodes, and the
// site's pages and feeds by the paths they name; its code as Prepare split it, by its node; the words of each footnote
// by its number, read first, which every citation of it carries; and the names taken on the post's page, which each
// heading's is kept clear of.
type bodyReader struct {
	body     content.Markdown
	page     document.Page
	pictures map[string]content.Picture
	links    map[*ast.Link]content.WrittenLink
	paths    map[string]document.Target
	codes    map[ast.Node]document.CodeBlock
	notes    map[int][]document.Span
	names    names
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
			title := reader.spans(heading)
			name := reader.names.given(cmp.Or(document.SlugOf(document.PlainText(title)), unnamedHeading))
			inside, rest := reader.nested(nodes[1:], heading.Level)
			blocks = append(blocks, document.Section{Title: title, Name: name, Blocks: inside})
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
// heading the nesting reads, and the parts of a list or a table their own block. A paragraph that opens with a picture
// is a figure: the load holds a picture to a paragraph of its own.
func (reader bodyReader) block(node ast.Node) (document.Block, bool) {
	switch typed := node.(type) {
	case *ast.Paragraph, *ast.TextBlock:
		if picture, isPicture := node.FirstChild().(*ast.Image); isPicture {
			return reader.figure(picture), true
		}
		return document.Paragraph{Spans: reader.spans(node)}, true
	case *ast.FencedCodeBlock, *ast.CodeBlock:
		return reader.codes[node], true
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

// figure is a picture of the body: its file beside the post, by the name its address gives it, its alt text and its
// caption as a reader reads them, and the size of the picture the load read.
func (reader bodyReader) figure(picture *ast.Image) document.Figure {
	name := content.ResolvedAddress(picture.Destination)
	shown := reader.pictures[name]
	return document.Figure{
		Picture: document.File{Page: reader.page, Name: name},
		Alt:     reader.body.AltText(picture),
		Caption: content.Resolved(picture.Title),
		Width:   shown.Width,
		Height:  shown.Height,
	}
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
		return append(spans, document.Link{Target: reader.linkTarget(typed), Spans: reader.spans(typed)})
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

// linkTarget is where a link leads: inside the site, its page, a section of it or a feed; else its address.
func (reader bodyReader) linkTarget(link *ast.Link) document.Target {
	written, isInside := reader.links[link]
	if !isInside {
		return addressTarget(content.ResolvedAddress(link.Destination))
	}
	target, isServed := reader.paths[written.Link.Path]
	switch {
	case !isServed:
		// a path nothing is served at reads as no page, which Check refuses
		return document.PageOf(document.PageNone)
	case written.Link.Anchor == "":
		return target
	}
	// a feed has no sections, so its section is one of no page, which Check refuses
	page, _ := target.(document.Page)
	return document.Anchor{Page: page, Name: written.Link.Anchor}
}

// addressTarget is where a link's address outside the site leads: a mailbox where it writes to one, an address of the
// web otherwise, which the load holds such an address to.
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
