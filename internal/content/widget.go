package content

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.yaml.in/yaml/v3"
)

// Widget is an island a post shows as a block of its own (docs/FOUNDATION.md §5.4, Layout of a post): the component of
// web/src/widgets it names, its alt text, which stands for it wherever it does not run, its caption, none where it
// has none, and its props, held to their form as the post loads.
type Widget struct {
	Name    string
	Alt     string
	Caption string
	// Props are a JSON object of every prop the block writes but its alt text and caption, in their written order
	Props json.RawMessage
}

// ShownWidget is a widget where a post's file shows it: the component it names, the file and the line of its block's
// fence.
type ShownWidget struct {
	Name string
	File string
	Line int
}

// KindWidget is the kind of a widget's node in a post's tree.
var KindWidget = ast.NewNodeKind("Widget") //nolint:gochecknoglobals // goldmark names a kind once, by a count it keeps, as its own extensions do

// WidgetNode is a widget where a post's tree holds it, read from its fenced block as the post is parsed: the widget,
// and the faults its reading found, each at its line counted from the block's fence, which the load writes at the
// post's lines.
type WidgetNode struct {
	ast.BaseBlock
	Widget Widget
	faults []readingFault
}

// Kind implements ast.Node.Kind.
func (*WidgetNode) Kind() ast.NodeKind {
	return KindWidget
}

// Dump implements ast.Node.Dump.
func (node *WidgetNode) Dump(source []byte, level int) {
	ast.DumpHelper(node, source, level, nil, nil)
}

// readingFault is a fault a widget's reading found: its line counted from the block's fence, and what is wrong.
type readingFault struct {
	offset  int
	problem string
}

// The language a widget's fenced block names, and how the component named after it is written, as Svelte names one: a
// capital letter, then letters and digits
const (
	widgetLanguage    = "widget"
	widgetNamePattern = `^[A-Z][A-Za-z0-9]*$`
)

// The keys of a widget's alt text and caption among its props, and the tag YAML gives a key that reads as text
const (
	altKey     = "alt"
	captionKey = "caption"
	textTag    = "!!str"
)

// A fault of a widget's block that stands at the line of its fence: of its name, or of its body as a whole
const fenceOffset = 0

// widgetReading reads each fenced block of a post's tree whose language is widget into a widget's node as the post is
// parsed, so nothing changes the tree after the load. It holds the form a component's name is written in, compiled
// once for every post the parser reads.
type widgetReading struct {
	name *regexp.Regexp
}

// Transform implements parser.ASTTransformer: each widget's fenced block, wherever it stands, becomes its node.
func (reading widgetReading) Transform(tree *ast.Document, reader text.Reader, _ parser.Context) {
	written := Markdown{Source: reader.Source(), Tree: tree}
	for _, fence := range widgetFences(written, tree) {
		fence.Parent().ReplaceChild(fence.Parent(), fence, reading.read(written.Source, fence))
	}
}

// widgetFences are the fenced blocks a node holds whose language is widget, at any depth, in reading order.
func widgetFences(written Markdown, parent ast.Node) []*ast.FencedCodeBlock {
	fences := []*ast.FencedCodeBlock{}
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		if fence, isFence := child.(*ast.FencedCodeBlock); isFence && written.Language(fence) == widgetLanguage {
			fences = append(fences, fence)
		}
		fences = append(fences, widgetFences(written, child)...)
	}
	return fences
}

// read reads a widget's fenced block into its node at the block's place: the component it names, and a body that is a
// mapping of its props, a line of alt text and a line of caption among them, written as propsReading holds them, the
// rest kept as JSON.
func (reading widgetReading) read(source []byte, fence *ast.FencedCodeBlock) *WidgetNode {
	info := string(fence.Info.Segment.Value(source))
	props := &propsReading{body: fencedBody(source, fence.Lines()), faults: nil}
	_, name, _ := strings.Cut(info, languageEnd)
	if !reading.name.MatchString(name) {
		props.add(fenceOffset, widgetNameProblem, info)
	}
	widget := Widget{Name: name, Alt: "", Caption: "", Props: nil}
	if mapping, isMapping := props.mapping(); isMapping {
		props.inspect(mapping)
		widget.Alt, widget.Caption, widget.Props = props.altText(mapping), props.caption(mapping), props.propsJSON(mapping)
	}
	node := &WidgetNode{BaseBlock: ast.BaseBlock{}, Widget: widget, faults: props.faults}
	node.SetPos(fence.Pos())
	return node
}

// inspectWidget holds a widget to a block of the post's own, never inside a quote, a list's item or a footnote, writes
// the faults its reading found at the post's lines, and keeps where it is shown.
func (checked *bodyInspection) inspectWidget(widget *WidgetNode) {
	line := checked.lineOf(widget)
	checked.shown.widgets = append(checked.shown.widgets, ShownWidget{Name: widget.Widget.Name, File: checked.inspection.file, Line: line})
	if widget.Parent().Kind() != ast.KindDocument {
		checked.inspection.addAtLine(line, widgetPlaceProblem)
	}
	for _, fault := range widget.faults {
		checked.inspection.addProblemAtLine(line+fault.offset, fault.problem)
	}
}

// propsReading holds a widget's body to the YAML its props are written in: the body as written, and the faults found,
// each at its line counted from the block's fence, which the body's first line follows.
type propsReading struct {
	body   []byte
	faults []readingFault
}

// fencedBody is a fenced block's body as its text writes it, its lines one after another, each with its line's end.
func fencedBody(source []byte, segments *text.Segments) []byte {
	body := []byte{}
	for index := range segments.Len() {
		line := segments.At(index)
		body = append(body, line.Value(source)...)
	}
	return body
}

// mapping is the mapping a widget's body is, one document of YAML; or a fault where the body is empty, reads as no
// YAML, holds a second document, whose props would be dropped, or is no mapping.
func (reading *propsReading) mapping() (*yaml.Node, bool) {
	decoder := yaml.NewDecoder(bytes.NewReader(reading.body))
	var body, more yaml.Node
	err := decoder.Decode(&body)
	switch {
	case errors.Is(err, io.EOF):
		reading.add(fenceOffset, widgetMappingProblem)
		return nil, false
	case err != nil:
		reading.add(fenceOffset, widgetYAMLProblem, err)
		return nil, false
	case !errors.Is(decoder.Decode(&more), io.EOF):
		reading.add(fenceOffset, widgetMappingProblem)
		return nil, false
	}
	// a document decoded holds its one root
	props := body.Content[0]
	if props.Kind != yaml.MappingNode {
		reading.add(props.Line, widgetMappingProblem)
		return nil, false
	}
	return props, true
}

// inspect holds every value of a widget's props, and each mapping's keys, to the YAML props are written in, as
// inspectOwn holds each; and each mapping's keys to text, none twice.
func (reading *propsReading) inspect(props *yaml.Node) {
	for _, node := range valueNodes("", props) {
		reading.inspectOwn(node)
		if node.Kind == yaml.MappingNode {
			reading.inspectKeys(node)
		}
	}
}

// inspectKeys holds a mapping's keys to text, none twice, each as inspectOwn holds a node.
func (reading *propsReading) inspectKeys(mapping *yaml.Node) {
	seen := map[string]bool{}
	for pair := range slices.Chunk(mapping.Content, keyAndValue) {
		key := pair[keyOffset]
		switch {
		case key.Kind != yaml.ScalarNode || key.ShortTag() != textTag:
			reading.add(key.Line, widgetKeyProblem, key.Value)
		case seen[key.Value]:
			reading.add(key.Line, widgetTwiceProblem, key.Value)
		}
		seen[key.Value] = true
		reading.inspectOwn(key)
	}
}

// inspectOwn holds one node of a widget's body, a value or a key, to the YAML props are written in: no anchor, alias
// or tag of its own, by which a value would stand for another or read as a type of its own; and no text its escapes
// turn into a character no value may hold, which the check of the body's lines does not see.
func (reading *propsReading) inspectOwn(node *yaml.Node) {
	switch {
	case node.Anchor != "", node.Kind == yaml.AliasNode:
		reading.add(node.Line, widgetAnchorProblem)
	case node.Style&yaml.TaggedStyle != 0:
		reading.add(node.Line, widgetTagProblem, node.Tag)
	// a refused character the body holds as written faults its line already, wherever it stands
	case node.Kind == yaml.ScalarNode && referencedCharacters(reading.body, node.Value):
		reading.add(node.Line, characterProblem, node.Value)
	}
}

// altText is a widget's alt text, held to a line of words: written, one line, not blank.
func (reading *propsReading) altText(props *yaml.Node) string {
	alt, written := valueOf(props, altKey)
	if !written {
		reading.add(fenceOffset, widgetAltProblem)
		return ""
	}
	return reading.lineOfText(alt, altKey)
}

// caption is a widget's caption, held to a line of words; none where its body writes none.
func (reading *propsReading) caption(props *yaml.Node) string {
	caption, written := valueOf(props, captionKey)
	if !written {
		return ""
	}
	return reading.lineOfText(caption, captionKey)
}

// lineOfText is a widget's alt text or caption, by its key, held to a line of words: a value a string takes, which a
// list or a mapping is not, of one line, not blank, as a null is.
func (reading *propsReading) lineOfText(value *yaml.Node, key string) string {
	var line string
	if value.Decode(&line) != nil || strings.ContainsRune(line, lineFeed) || blank(line) {
		reading.add(value.Line, widgetLineProblem, key)
		return ""
	}
	return line
}

// add writes a fault at a line counted from the block's fence, the line YAML counts a node of the body at, from its
// problem's format and the values it names.
func (reading *propsReading) add(offset int, format string, values ...any) {
	reading.faults = append(reading.faults, readingFault{offset: offset, problem: fmt.Sprintf(format, values...)})
}
