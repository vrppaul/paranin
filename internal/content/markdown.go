package content

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
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
