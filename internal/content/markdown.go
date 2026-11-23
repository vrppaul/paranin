package content

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extensionast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"
)

// Markdown is a text in the posts' format, read once: its source, and the tree goldmark parsed it into, whose nodes
// point into the source, so whoever draws the text walks the tree and never parses it again. The tree is read and
// never changed after the load, by no transformer, attribute or id set later: a snapshot's posts may be composed by
// two goroutines at once.
type Markdown struct {
	Source []byte
	Tree   ast.Node
}

// The priorities goldmark's extensions give their parts, the lowest tried first, which place them among CommonMark's
const (
	// a table is read from a paragraph once its links' definitions are
	tableParagraphPriority = 200
	tableTreePriority      = 0
	// a footnote's definition is tried after every other block but a paragraph, its reference before a link
	footnoteBlockPriority  = 999
	footnoteInlinePriority = 101
)

// postFormat is how goldmark reads a post: CommonMark and, of GFM, its tables and footnotes alone.
//
// goldmark's root package would add the extensions itself, but it renders too, which no part under the faces does,
// so the parser is put together from the parts the extensions' own Extend methods add. One is left out: the
// footnotes' transformer, which drops a definition no reference cites, where the check refuses one, and moves the
// definitions and links them to their references, which compose does as it makes the document. A footnote keeps the
// number its first reference gave it; how often each is cited, which the transformer would count, compose counts.
func postFormat() []parser.Option {
	return []parser.Option{
		parser.WithBlockParsers(parser.DefaultBlockParsers()...),
		parser.WithBlockParsers(util.Prioritized(extension.NewFootnoteBlockParser(), footnoteBlockPriority)),
		parser.WithInlineParsers(parser.DefaultInlineParsers()...),
		parser.WithInlineParsers(util.Prioritized(extension.NewFootnoteParser(), footnoteInlinePriority)),
		parser.WithParagraphTransformers(parser.DefaultParagraphTransformers()...),
		parser.WithParagraphTransformers(util.Prioritized(extension.NewTableParagraphTransformer(), tableParagraphPriority)),
		parser.WithASTTransformers(util.Prioritized(extension.NewTableASTTransformer(), tableTreePriority)),
	}
}

// PostKinds are the kinds of node a post's tree may hold, as postFormat parses it, and each has its case in compose's
// walk: the load refuses a node of any other kind, so a kind an extension brings later is refused, never dropped
// unseen. Raw HTML and a picture stand outside it.
func PostKinds() []ast.NodeKind {
	return []ast.NodeKind{
		ast.KindDocument, ast.KindParagraph, ast.KindTextBlock, ast.KindHeading, ast.KindFencedCodeBlock,
		ast.KindCodeBlock, ast.KindBlockquote, ast.KindThematicBreak, ast.KindList, ast.KindListItem,
		ast.KindLinkReferenceDefinition,
		ast.KindText, ast.KindCodeSpan, ast.KindEmphasis, ast.KindLink, ast.KindAutoLink,
		extensionast.KindTable, extensionast.KindTableHeader, extensionast.KindTableRow, extensionast.KindTableCell,
		extensionast.KindFootnoteList, extensionast.KindFootnote, extensionast.KindFootnoteLink,
	}
}

// What escapes the mark after it in a post's text
const escapeMark = '\\'

// Resolved is a run of a post's text as CommonMark reads it, in one pass: a mark escaped by a backslash the mark
// alone, which opens no reference (`\&amp;` reads as written), and a reference to a character, by its number or its
// name, the character. The tree keeps the text as written, which a renderer resolves as it writes.
func Resolved(written []byte) string {
	var builder strings.Builder
	start := 0
	for index := 0; index < len(written)-1; index++ {
		if written[index] == escapeMark && util.IsPunct(written[index+1]) {
			builder.Write(resolvedReferences(written[start:index]))
			builder.WriteByte(written[index+1])
			index++
			start = index + 1
		}
	}
	builder.Write(resolvedReferences(written[start:]))
	return builder.String()
}

// ResolvedAddress is a link's address as goldmark reads it before it writes the address: its escapes, then its
// references to characters.
func ResolvedAddress(written []byte) string {
	return string(resolvedReferences(util.UnescapePunctuations(written)))
}

// resolvedReferences is text with each reference to a character, by its number or its name, the character.
func resolvedReferences(written []byte) []byte {
	return util.ResolveEntityNames(util.ResolveNumericReferences(written))
}

// What ends the language of a fenced block's line of information, and what may stand after it
const languageEnd = " "

// Language is what a fenced block of code is written in, the first word of its line of information; none where it
// has none. goldmark's own Language keeps what it read in the node, a write to a tree two goroutines may read.
func (markdown Markdown) Language(block *ast.FencedCodeBlock) string {
	if block.Info == nil {
		return ""
	}
	language, _, _ := strings.Cut(string(block.Info.Segment.Value(markdown.Source)), languageEnd)
	return language
}

// The line that opens a post's front matter and closes it, the file's line it opens on, and the carriage return a
// file written on Windows ends its lines with
const (
	frontMatterFence = "---"
	openingLine      = 1
	carriageReturn   = '\r'
)

// postParts are the two parts of a post's file: its front matter, from the line that opens it, so YAML counts the
// file's own lines, to the line before the one that closes it; and its body after that, with the line it starts on.
type postParts struct {
	frontMatter []byte
	body        []byte
	bodyLine    int
}

// splitPost parts a post's file into its front matter and its body, or finds no front matter that opens it or none
// that closes.
func splitPost(file string, data []byte) (postParts, Faults) {
	none := postParts{frontMatter: nil, body: nil, bodyLine: 0}
	lines := bytes.SplitAfter(data, []byte{lineFeed})
	if !isFence(lines[0]) {
		return none, Faults{lineFault(file, openingLine, frontMatterProblem)}
	}
	frontMatterEnd := len(lines[0])
	for index, line := range lines[1:] {
		if isFence(line) {
			closingLine := openingLine + 1 + index
			return postParts{frontMatter: data[:frontMatterEnd], body: data[frontMatterEnd+len(line):], bodyLine: closingLine + 1}, nil
		}
		frontMatterEnd += len(line)
	}
	return none, Faults{lineFault(file, openingLine, unclosedProblem)}
}

// isFence is whether a line of a post's file opens or closes its front matter, whichever way the line ends: a
// carriage return is read past here, and refused as a character of the body.
func isFence(line []byte) bool {
	return strings.TrimRight(string(line), string([]rune{carriageReturn, lineFeed})) == frontMatterFence
}
