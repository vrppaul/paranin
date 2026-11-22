package content

import (
	"bytes"
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

// The one character a body may hold that no other value may, inside code alone, and what it is read as there: a tab,
// which gofmt writes and a face expands as it draws the code
const (
	tab   = '\t'
	space = ' '
)

// bodyInspection checks a post's body as goldmark parsed it, node by node, each fault at the file's line of its node.
// It keeps the level of the heading before, which the next may go below by one level alone, and the code it met,
// where a tab may stand.
type bodyInspection struct {
	inspection   *inspection
	source       []byte
	firstLine    int
	headingLevel int
	reference    *regexp.Regexp
	code         []text.Segment
}

// inspectBody finds in a post's body, which starts at a line of its file, what the format refuses: raw HTML, a
// heading of the title's level or one that skips a level, a widget, a footnote no reference cites and a reference to
// no footnote, and characters no value may hold.
func inspectBody(inspection *inspection, body Markdown, firstLine int) {
	checked := &bodyInspection{
		inspection: inspection, source: body.Source, firstLine: firstLine, headingLevel: titleLevel,
		reference: regexp.MustCompile(referencePattern), code: nil,
	}
	// visit returns no error, so the walk ends with none
	_ = ast.Walk(body.Tree, checked.visit)
	checked.inspectCharacters()
}

func (checked *bodyInspection) visit(node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	switch typed := node.(type) {
	case *ast.HTMLBlock:
		checked.addRawHTML(node, typed.Lines().Value(checked.source))
	case *ast.RawHTML:
		checked.addRawHTML(node, typed.Segments.Value(checked.source))
	case *ast.Heading:
		checked.inspectHeading(typed)
	case *ast.FencedCodeBlock:
		if string(typed.Language(checked.source)) == widgetLanguage {
			checked.inspection.addAtLine(checked.lineOf(node), widgetProblem)
		}
		checked.addCodeLines(typed.Lines())
	case *ast.CodeBlock:
		checked.addCodeLines(typed.Lines())
	case *ast.CodeSpan:
		checked.addCodeSpan(typed)
		// a code span's text is code, where a reference is written as it is
		return ast.WalkContinue, nil
	case *extensionast.Footnote:
		if typed.Index < 0 {
			checked.inspection.addAtLine(checked.lineOf(node), uncitedProblem, typed.Ref)
		}
	}
	checked.inspectReferences(node)
	return ast.WalkContinue, nil
}

// addRawHTML writes the fault of raw HTML at its node, quoting its first line, by which the author tells two on a
// line apart.
func (checked *bodyInspection) addRawHTML(node ast.Node, written []byte) {
	firstLine, _, _ := strings.Cut(string(written), string(lineFeed))
	checked.inspection.addAtLine(checked.lineOf(node), rawHTMLProblem, firstLine)
}

// inspectHeading holds a heading below the title's level, and at most one level below the heading before it. The
// level a heading of the title's level stands for is the author's to say, so the heading after it is held to none.
func (checked *bodyInspection) inspectHeading(heading *ast.Heading) {
	switch {
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
		lineStart := bytes.LastIndexByte(checked.source[:line.Start], lineFeed) + 1
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
		joined = append(joined, written.Segment.Value(checked.source)...)
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
	for written := range bytes.Lines(withoutCodeTabs(checked.source, checked.code)) {
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
	return checked.firstLine + bytes.Count(checked.source[:node.Pos()], []byte{lineFeed})
}
