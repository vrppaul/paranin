// Command commitmessage holds a commit's message to .claude/rules/prose.md: a short subject that tells nothing of
// what the commit replaces, and a few lines of why under it. Taken from stint's scripts/check_commit_message.py,
// which refuses history in the body too, where prose.md has a commit say what changed.
//
// The commit-msg hook runs it on the message's file:
//
//	go run ./scripts/commitmessage .git/COMMIT_EDITMSG
//
// It prints each fault as path:line: what is wrong, and exits 0 when the message holds, 1 when it has a fault or
// cannot be read, and 2 when the command is written wrong.
package main

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/vrppaul/paranin/scripts/history"
)

// What the program exits with
const (
	succeeded     = 0
	failed        = 1
	usageMistaken = 2
)

// How the command is written: the one argument, the message's file
const usage = "usage: go run ./scripts/commitmessage <message file>"

// The longest a subject may be, in letters, and the most lines of why under it
const (
	longestSubject = 72
	mostBodyLines  = 5
)

// What git writes beside the author's words: its own lines, and the diff `git commit -v` shows under the scissors
const (
	gitCommentMark = "#"
	scissors       = "# ------------------------ >8 ------------------------"
)

// A trailer git reads in a message's last paragraph, as `Co-Authored-By: …` and `Claude-Session: …`
const trailerPattern = `^[A-Za-z][A-Za-z-]*: \S`

// Lines are counted from one, as an editor counts them, and part at a line break
const (
	firstLine = 1
	lineBreak = "\n"
)

// kind is what is wrong with a message.
type kind int

const (
	// kindLongSubject is a subject longer than a line of the log may hold
	kindLongSubject kind = iota
	// kindLongBody is more lines of why than a reader keeps in mind
	kindLongBody
	// kindHistory is a phrase in the subject that tells what the code was
	kindHistory
)

// fault is what is wrong with a message, and the line it stands on.
type fault struct {
	kind kind
	line int
	says string
}

// writtenLine is a line the author wrote, and its number in the file.
type writtenLine struct {
	number int
	text   string
}

func main() {
	os.Exit(run(os.Args[1:]))
}

// run reads the message's file the arguments name and prints each fault of it.
func run(arguments []string) int {
	if len(arguments) != 1 {
		fmt.Fprintln(os.Stderr, usage)
		return usageMistaken
	}
	path := arguments[0]
	message, err := os.ReadFile(path) //nolint:gosec // the message's file git hands the commit-msg hook
	if err != nil {
		fmt.Fprintf(os.Stderr, "read message %q: %v\n", path, err)
		return failed
	}
	found := faultsOf(string(message))
	for _, each := range found {
		fmt.Fprintf(os.Stderr, "%s:%d: %s\n", path, each.line, each.says)
	}
	if len(found) > 0 {
		return failed
	}
	return succeeded
}

// faultsOf are a message's faults: its subject's length and history, its body's length.
func faultsOf(message string) []fault {
	written := writtenLines(message)
	if len(written) == 0 {
		return nil
	}
	subject := written[0]
	found := []fault{}
	if letters := utf8.RuneCountInString(subject.text); letters > longestSubject {
		says := fmt.Sprintf("the subject is %d letters where %d are the most: say what the commit does", letters, longestSubject)
		found = append(found, fault{kind: kindLongSubject, line: subject.number, says: says})
	}
	if body := bodyOf(written[1:]); len(body) > mostBodyLines {
		says := fmt.Sprintf("the body is %d lines where %d are the most: the why alone, no list of files, no tally", len(body), mostBodyLines)
		found = append(found, fault{kind: kindLongBody, line: body[0].number, says: says})
	}
	for _, phrase := range history.NewReader().Within(subject.text) {
		says := fmt.Sprintf("%q tells what the code was: say what the commit does", phrase.Phrase)
		found = append(found, fault{kind: kindHistory, line: subject.number, says: says})
	}
	return found
}

// writtenLines are the lines the author wrote, numbered as in the file.
func writtenLines(message string) []writtenLine {
	aboveTheDiff, _, _ := strings.Cut(message, scissors)
	written := []writtenLine{}
	for index, text := range strings.Split(aboveTheDiff, lineBreak) {
		if !strings.HasPrefix(text, gitCommentMark) {
			written = append(written, writtenLine{number: index + firstLine, text: text})
		}
	}
	return written
}

// bodyOf are the lines under the subject that say why, the blank ones and the trailers left out.
func bodyOf(underTheSubject []writtenLine) []writtenLine {
	paragraphs := paragraphsOf(underTheSubject)
	trailer := regexp.MustCompile(trailerPattern)
	if last := len(paragraphs) - 1; last >= 0 && allTrailers(paragraphs[last], trailer) {
		paragraphs = paragraphs[:last]
	}
	return slices.Concat(paragraphs...)
}

// paragraphsOf are lines in their paragraphs, the blank lines between them left out.
func paragraphsOf(lines []writtenLine) [][]writtenLine {
	paragraphs := [][]writtenLine{}
	current := []writtenLine{}
	for _, line := range lines {
		if strings.TrimSpace(line.text) != "" {
			current = append(current, line)
			continue
		}
		if len(current) > 0 {
			paragraphs = append(paragraphs, current)
			current = []writtenLine{}
		}
	}
	if len(current) > 0 {
		paragraphs = append(paragraphs, current)
	}
	return paragraphs
}

// allTrailers tells whether every line of a paragraph is a trailer git reads.
func allTrailers(paragraph []writtenLine, trailer *regexp.Regexp) bool {
	return !slices.ContainsFunc(paragraph, func(line writtenLine) bool { return !trailer.MatchString(line.text) })
}
