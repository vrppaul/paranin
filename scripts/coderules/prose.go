package main

import (
	"go/ast"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/vrppaul/paranin/scripts/history"
)

const noHistoryDescription = `refuse a comment that tells what the code was, or dates itself

A comment says what stands; what changed, and when, a commit says (.claude/rules/prose.md). A comment that
holds a phrase of scripts/history, "no longer" or "renamed from", or a date, is refused. The comments of a
group are read as one passage, so a phrase broken over two lines is found all the same; it is reported at the
comment that holds its last word, the line a test can name it on. Test files are held too, and so is the Go
templ writes from a template, whose comments are the template's own. Taken from claude-review's
scripts/check_prose.py, which holds no length either: a longer comment already written stays until a change
rewrites it.`

const (
	historyMessage = "%q tells what the code was: say what stands, history is git's"
	dateMessage    = "%q dates the comment: a date is git's"
)

// A day as ISO 8601 writes it, in this century
const datePattern = `\b20\d\d-\d\d-\d\d\b`

// What the comments of a group are joined with, so a phrase broken over two lines is read whole
const passageJoint = " "

// The marks a comment's words stand between
const (
	lineCommentMark  = "//"
	blockCommentOpen = "/*"
	blockCommentEnd  = "*/"
)

func noHistoryAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{Name: "nohistory", Doc: noHistoryDescription, Run: runNoHistory}
}

// passagePart is a comment of a group, and where its words start in the passage the group is read as.
type passagePart struct {
	comment *ast.Comment
	start   int
}

func runNoHistory(pass *analysis.Pass) (any, error) {
	reader, dates := history.NewReader(), regexp.MustCompile(datePattern)
	for _, file := range slices.Concat(writtenFiles(pass), templateFiles(pass)) {
		for _, group := range file.Comments {
			passage, parts := passageOf(group)
			for _, found := range reader.Within(passage) {
				pass.Reportf(commentAt(parts, found.End-1).Pos(), historyMessage, found.Phrase)
			}
			for _, span := range dates.FindAllStringIndex(passage, -1) {
				pass.Reportf(commentAt(parts, span[1]-1).Pos(), dateMessage, passage[span[0]:span[1]])
			}
		}
	}
	return nil, nil
}

// passageOf is a group's comments as one passage of words, and where each comment starts in it.
func passageOf(group *ast.CommentGroup) (string, []passagePart) {
	var builder strings.Builder
	parts := make([]passagePart, 0, len(group.List))
	for _, comment := range group.List {
		parts = append(parts, passagePart{comment: comment, start: builder.Len()})
		builder.WriteString(strings.Join(strings.Fields(commentWords(comment.Text)), passageJoint))
		builder.WriteString(passageJoint)
	}
	return builder.String(), parts
}

// commentWords is a comment's text with the marks that open and close it taken off.
func commentWords(text string) string {
	if words, isLine := strings.CutPrefix(text, lineCommentMark); isLine {
		return words
	}
	return strings.TrimSuffix(strings.TrimPrefix(text, blockCommentOpen), blockCommentEnd)
}

// commentAt is the comment of a passage that holds the byte at an offset.
func commentAt(parts []passagePart, offset int) *ast.Comment {
	holding := parts[0].comment
	for _, part := range parts {
		if part.start <= offset {
			holding = part.comment
		}
	}
	return holding
}
