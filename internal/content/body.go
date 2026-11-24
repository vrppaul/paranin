package content

import (
	"bytes"
	"io/fs"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/yuin/goldmark/ast"
	extensionast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// The level of a post's title, the page's one heading of it, which no heading of the body takes; the deepest level
// markdown writes; and the language of a fenced block that is a widget
const (
	titleLevel     = 1
	deepestLevel   = 6
	widgetLanguage = "widget"
)

// How a reference to a footnote is written, after no backslash, which would escape it; where goldmark finds no
// definition for one it leaves it as text. Of a match's indices the second pair is the reference's
const (
	referencePattern = `(?:^|[^\\])(\[\^[^\]]+\])`
	referenceStart   = 2
	referenceEnd     = 3
)

// What may stand on a table row's line after its last cell: the spaces and the bar that close the row
const rowEdges = " |"

// The one character a body may hold that no other value may, inside code alone, and what it is read as there: a tab,
// which gofmt writes and a face expands as it draws the code
const (
	tab   = '\t'
	space = ' '
)

// writtenBody is a post's body where its file writes it: the body, the file's line it starts on, the post's folder,
// its place in the content's folder, whose pictures it shows, and the site's own address, which a link inside the site
// leaves to its path.
type writtenBody struct {
	body      Markdown
	firstLine int
	folder    fs.FS
	place     string
	site      WebAddress
}

// bodyInspection checks a post's body as goldmark parsed it, node by node, each fault at the file's line of its node,
// reads the pictures it shows from the post's folder, by their names, and keeps its links inside the site. It keeps
// the level of the heading before, which the next may go below by one level alone, and the code it met, where a tab
// may stand.
type bodyInspection struct {
	writtenBody
	inspection   *inspection
	headingLevel int
	reference    *regexp.Regexp
	code         []text.Segment
	pictures     shownPictures
	links        map[*ast.Link]WrittenLink
}

// shownPictures are the pictures a post's body shows: those read, by their names, and every name a picture's address
// gives, read or refused, which the check of the post's folder takes for a part of the post.
type shownPictures struct {
	read  map[string]Picture
	named map[string]bool
}

// add takes the pictures more shows among those shown.
func (shown *shownPictures) add(more shownPictures) {
	maps.Copy(shown.read, more.read)
	maps.Copy(shown.named, more.named)
}

// inspectBody finds in a post's body what the format refuses: a node of a kind no post holds (PostKinds), raw HTML
// among them, a heading of the title's level or one that skips a level, a widget, a pause between no two of the post's
// own blocks, a table's row of more cells than its header, a footnote no reference cites, one that is no one paragraph
// or cites another, a reference to no footnote and one in a heading; what no face draws, a line broken hard, a link's
// title and a picture that is no paragraph of its own; a link to neither the web, nor a path from the site's root, nor
// a mailbox, and one to the site's own address; a picture with no alt text, or none of the post's folder; and
// characters no value may hold, written or referenced. It hands back the pictures the body shows and its links inside
// the site, which only the whole site can hold to a page.
func inspectBody(inspection *inspection, written writtenBody) (shownPictures, map[*ast.Link]WrittenLink) {
	checked := &bodyInspection{
		writtenBody: written, inspection: inspection, headingLevel: titleLevel, reference: regexp.MustCompile(referencePattern),
		code: nil, pictures: shownPictures{read: map[string]Picture{}, named: map[string]bool{}}, links: map[*ast.Link]WrittenLink{},
	}
	// visit returns no error, so the walk ends with none
	_ = ast.Walk(written.body.Tree, checked.visit)
	checked.inspectCharacters()
	return checked.pictures, checked.links
}

func (checked *bodyInspection) visit(node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	if !slices.Contains(PostKinds(), node.Kind()) {
		checked.refuseKind(node)
		return ast.WalkSkipChildren, nil
	}
	// a node whose children the walk passes over holds code, where a reference stands as written
	status := checked.inspectNode(node)
	if status == ast.WalkContinue {
		checked.inspectReferences(node)
	}
	return status, nil
}

// refuseKind writes the fault of a node of a kind no post holds: a kind left out on purpose with its reason, raw HTML
// quoting its first line, by which the author tells two on a line apart; any other kind by its name.
func (checked *bodyInspection) refuseKind(node ast.Node) {
	switch typed := node.(type) {
	case *ast.HTMLBlock:
		checked.addRawHTML(node, typed.Lines().Value(checked.body.Source))
	case *ast.RawHTML:
		checked.addRawHTML(node, typed.Segments.Value(checked.body.Source))
	default:
		checked.inspection.addAtLine(checked.lineOf(node), kindProblem, node.Kind())
	}
}

// inspectNode checks a node of a kind a post holds by what its kind asks, and says whether the walk goes into it.
func (checked *bodyInspection) inspectNode(node ast.Node) ast.WalkStatus {
	switch typed := node.(type) {
	case *ast.Heading:
		checked.inspectHeading(typed)
	case *ast.FencedCodeBlock:
		if checked.body.Language(typed) == widgetLanguage {
			checked.inspection.addAtLine(checked.lineOf(node), widgetProblem)
		}
		checked.addCodeLines(typed.Lines())
	case *ast.CodeBlock:
		checked.addCodeLines(typed.Lines())
	case *ast.ThematicBreak:
		checked.inspectPause(typed)
	case *ast.CodeSpan:
		checked.addCodeSpan(typed)
		// a code span's text is code, where a reference is written as it is
		return ast.WalkSkipChildren
	case *ast.Text:
		checked.inspectText(typed)
	case *ast.Link:
		checked.inspectLink(typed)
	case *ast.Image:
		checked.inspectPicture(typed)
	case *ast.AutoLink:
		checked.inspectAutoLink(typed)
	case *extensionast.TableRow:
		checked.inspectRow(typed)
	case *extensionast.Footnote:
		checked.inspectFootnote(typed)
	case *extensionast.FootnoteLink:
		if inside(typed, extensionast.KindFootnote) {
			checked.inspection.addAtLine(checked.lineOf(node), noteCitingProblem)
		}
		if inside(typed, ast.KindHeading) {
			checked.inspection.addAtLine(checked.lineOf(node), noteHeadingProblem)
		}
	}
	return ast.WalkContinue
}

// inspectPause holds a pause to a place between two of the post's own blocks, the two runs of text it parts: one inside
// a quote or a list, or at the body's start or end, parts none.
func (checked *bodyInspection) inspectPause(pause *ast.ThematicBreak) {
	between := pause.Parent().Kind() == ast.KindDocument &&
		hasBlockToward(pause, ast.Node.PreviousSibling) && hasBlockToward(pause, ast.Node.NextSibling)
	if !between {
		checked.inspection.addAtLine(checked.lineOf(pause), pausePlaceProblem)
	}
}

// placelessKinds are the kinds of block a post's text holds that stand nowhere in it: the footnotes, each drawn where a
// line cites it, and a link's definition, which the links it names carry.
func placelessKinds() []ast.NodeKind {
	return []ast.NodeKind{extensionast.KindFootnoteList, ast.KindLinkReferenceDefinition}
}

// hasBlockToward is whether a block that stands in the text lies beside a node, one way, its siblings taken in turn.
func hasBlockToward(node ast.Node, step func(ast.Node) ast.Node) bool {
	for sibling := step(node); sibling != nil; sibling = step(sibling) {
		if !slices.Contains(placelessKinds(), sibling.Kind()) {
			return true
		}
	}
	return false
}

// inspectRow finds a row of a table with more cells than its header, which GFM drops from the tree: words left on
// the row's line after its last cell.
func (checked *bodyInspection) inspectRow(row *extensionast.TableRow) {
	end := -1
	for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
		if lines := cell.Lines(); lines.Len() > 0 {
			end = lines.At(lines.Len() - 1).Stop
		}
	}
	if end < 0 {
		return
	}
	rest, _, _ := bytes.Cut(checked.body.Source[end:], []byte{lineFeed})
	if strings.Trim(string(rest), rowEdges) != "" {
		checked.inspection.addAtLine(checked.lineOf(row), extraCellsProblem)
	}
}

// inspectText finds a line broken hard, and a reference that reads as a character no value may hold.
func (checked *bodyInspection) inspectText(written *ast.Text) {
	if written.HardLineBreak() {
		checked.inspection.addAtLine(checked.lineOf(written), lineBreakProblem)
	}
	source := written.Segment.Value(checked.body.Source)
	if resolved := Resolved(source); referencedCharacters(source, resolved) {
		checked.inspection.addAtLine(checked.lineOf(written), characterProblem, resolved)
	}
}

// inspectLink keeps a link inside the site and holds any other to a mailbox or another site, with no title.
func (checked *bodyInspection) inspectLink(link *ast.Link) {
	line := checked.lineOf(link)
	if link.Title != nil {
		checked.inspection.addAtLine(line, linkTitleProblem, link.Title)
	}
	address := ResolvedAddress(link.Destination)
	_, isMail := Mailbox(address)
	inside, isInside := siteLinkOf(address)
	web := readWebAddress(address)
	switch {
	case referencedCharacters(link.Destination, address):
		checked.inspection.addAtLine(line, characterProblem, address)
	case isMail:
	case isInside:
		checked.links[link] = WrittenLink{Link: inside, File: checked.inspection.file, Line: line}
	case !web.absolute():
		checked.inspection.addAtLine(line, linkProblem, address)
	case web.sameMachine(checked.site):
		checked.inspection.addAtLine(line, ownAddressProblem, address)
	}
}

// inspectAutoLink holds an address written between angle brackets to an absolute web address of another site.
func (checked *bodyInspection) inspectAutoLink(link *ast.AutoLink) {
	if link.AutoLinkType != ast.AutoLinkURL {
		return
	}
	web := readWebAddress(string(link.URL(checked.body.Source)))
	switch {
	case !web.absolute():
		checked.inspection.addAtLine(checked.lineOf(link), addressProblem, web.Written())
	case web.sameMachine(checked.site):
		checked.inspection.addAtLine(checked.lineOf(link), ownAddressProblem, web.Written())
	}
}

// referencedCharacters is whether a text's references, not its source, read as a character no value may hold once
// resolved: a character the source holds itself the check of the source's lines names.
func referencedCharacters(written []byte, resolved string) bool {
	return refusedText(resolved) && !refusedText(string(written))
}

// inspectFootnote holds a footnote to a reference that cites it, and to one paragraph, which a note set beside the
// line holds.
func (checked *bodyInspection) inspectFootnote(footnote *extensionast.Footnote) {
	if footnote.Index < 0 {
		checked.inspection.addAtLine(checked.lineOf(footnote), uncitedProblem, footnote.Ref)
	}
	if footnote.ChildCount() != 1 || footnote.FirstChild().Kind() != ast.KindParagraph {
		checked.inspection.addAtLine(checked.lineOf(footnote), noteFormProblem, footnote.Ref)
	}
}

// inside is whether a node stands inside one of a kind: a footnote's words, a heading.
func inside(node ast.Node, kind ast.NodeKind) bool {
	for parent := node.Parent(); parent != nil; parent = parent.Parent() {
		if parent.Kind() == kind {
			return true
		}
	}
	return false
}

// addRawHTML writes the fault of raw HTML at its node, quoting its first line, by which the author tells two on a
// line apart.
func (checked *bodyInspection) addRawHTML(node ast.Node, written []byte) {
	firstLine, _, _ := strings.Cut(string(written), string(lineFeed))
	checked.inspection.addAtLine(checked.lineOf(node), rawHTMLProblem, firstLine)
}

// inspectHeading holds a heading to the post's own blocks, never inside a quote or a list's item, where it opens no
// part of the post; below the title's level, and at most one level below the heading before it. The level a heading
// of the title's level stands for is the author's to say, so the heading after it is held to none.
func (checked *bodyInspection) inspectHeading(heading *ast.Heading) {
	switch {
	case heading.Parent().Kind() != ast.KindDocument:
		checked.inspection.addAtLine(checked.lineOf(heading), nestedHeadingProblem)
		return
	case heading.Level == titleLevel:
		checked.inspection.addAtLine(checked.lineOf(heading), titleLevelProblem)
		checked.headingLevel = deepestLevel
		return
	case heading.Level > checked.headingLevel+1:
		checked.inspection.addAtLine(checked.lineOf(heading), skippedLevelProblem, heading.Level, checked.headingLevel)
	}
	checked.headingLevel = heading.Level
}

// addCodeLines keeps a code block's lines as code, each from the start of its line in the file, so the indent of an
// indented block counts as its code too.
func (checked *bodyInspection) addCodeLines(lines *text.Segments) {
	for index := range lines.Len() {
		line := lines.At(index)
		lineStart := bytes.LastIndexByte(checked.body.Source[:line.Start], lineFeed) + 1
		checked.code = append(checked.code, text.NewSegment(lineStart, line.Stop))
	}
}

// addCodeSpan keeps a code span's text as code.
func (checked *bodyInspection) addCodeSpan(span *ast.CodeSpan) {
	for child := span.FirstChild(); child != nil; child = child.NextSibling() {
		if written, isText := child.(*ast.Text); isText {
			checked.code = append(checked.code, written.Segment)
		}
	}
}

// inspectReferences finds, in each run of a node's children that are text, a reference to a footnote that no
// definition answers.
func (checked *bodyInspection) inspectReferences(parent ast.Node) {
	run := []*ast.Text{}
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		if written, isText := child.(*ast.Text); isText {
			run = append(run, written)
			continue
		}
		checked.inspectRun(run)
		run = []*ast.Text{}
	}
	checked.inspectRun(run)
}

// inspectRun finds a reference no definition answers in a run of text, read whole, since goldmark parts a run where
// it tried a link, at the reference's bracket among others. The fault stands at the line of the text it starts in.
func (checked *bodyInspection) inspectRun(run []*ast.Text) {
	joined := []byte{}
	starts := make([]int, 0, len(run))
	for _, written := range run {
		starts = append(starts, len(joined))
		joined = append(joined, written.Segment.Value(checked.body.Source)...)
	}
	for _, match := range checked.reference.FindAllSubmatchIndex(joined, -1) {
		after, _ := slices.BinarySearch(starts, match[referenceStart]+1)
		checked.inspection.addAtLine(checked.lineOf(run[after-1]), unansweredProblem, joined[match[referenceStart]:match[referenceEnd]])
	}
}

// inspectCharacters finds every line of the body that holds a character no value of the content may hold, but a
// tab inside code, and quotes it with that tab read as a space.
func (checked *bodyInspection) inspectCharacters() {
	line := checked.firstLine
	for written := range bytes.Lines(withoutCodeTabs(checked.body.Source, checked.code)) {
		if text := strings.TrimSuffix(string(written), string(lineFeed)); refusedText(text) {
			checked.inspection.addAtLine(line, characterProblem, text)
		}
		line++
	}
}

// withoutCodeTabs is a copy of a source with every tab inside code read as a space.
func withoutCodeTabs(source []byte, code []text.Segment) []byte {
	readable := bytes.Clone(source)
	for _, segment := range code {
		for index := segment.Start; index < segment.Stop; index++ {
			if readable[index] == tab {
				readable[index] = space
			}
		}
	}
	return readable
}

// lineOf is the file's line a node starts on. Every block the parser opens and every inline it reads has its
// position, and those are the nodes a check names.
func (checked *bodyInspection) lineOf(node ast.Node) int {
	return checked.firstLine + bytes.Count(checked.body.Source[:node.Pos()], []byte{lineFeed})
}
