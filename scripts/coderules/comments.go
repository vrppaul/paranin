package main

import (
	"fmt"
	"go/ast"
	"strings"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
)

// The widest a line of a comment may be, in letters, a tab one of them as golangci-lint's lll counts it; what
// opens a line comment and a block comment; what ends a line; and what may stand before a comment that opens one
const (
	widestComment  = 120
	commentOpening = "/"
	lineEnd        = "\n"
	indentation    = " \t"
)

const commentMessage = "a comment's line is %d columns wide, over the %d a line of prose may take: wrap it"

const commentWidthDescription = `refuse a comment's line wider than prose may run

A line that a comment opens is at most 120 columns wide, a tab counted as one, as golangci-lint's lll counts
it: prose wrapped where the rest of the code's prose is, so a diff reads it and an editor shows it whole. A
line of code may run wider, as a struct literal with every field written does; a comment after it is the
code's. A block comment is held by its first line. Test files are held too; generated files are passed over.`

func commentWidthAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "commentwidth",
		Doc:  commentWidthDescription,
		Run:  runCommentWidth,
	}
}

// runCommentWidth reports each line of a written file that a comment opens and that runs wider than prose may.
func runCommentWidth(pass *analysis.Pass) (any, error) {
	for _, file := range writtenFiles(pass) {
		name := pass.Fset.File(file.Pos()).Name()
		source, err := pass.ReadFile(name)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		lines := strings.Split(string(source), lineEnd)
		for _, group := range file.Comments {
			for _, comment := range group.List {
				reportWideComment(pass, lines, comment)
			}
		}
	}
	return nil, nil
}

// reportWideComment reports a comment whose line it opens and which runs wider than prose may.
func reportWideComment(pass *analysis.Pass, lines []string, comment *ast.Comment) {
	line := lines[pass.Fset.Position(comment.Pos()).Line-1]
	if !strings.HasPrefix(strings.TrimLeft(line, indentation), commentOpening) {
		return
	}
	if width := utf8.RuneCountInString(line); width > widestComment {
		pass.Reportf(comment.Pos(), commentMessage, width, widestComment)
	}
}
