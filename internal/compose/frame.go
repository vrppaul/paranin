package compose

import (
	"strconv"
	"strings"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// What parts a command's program from the words after it
const programEnd = " "

// Frame is what a face lays round the documents it draws, on every page alike, in the site's words: the name of the
// theme button, which is pressed while the page is dark, the words of the button that copies a command or a block of
// code, the way a terminal reaches the site over ssh, what reading a page in a terminal gives, whose command a face
// writes with the page's address, the words of the button that writes to the author, the page of each project that
// has one, in the profile's order, which a face's command line goes to by slug as the TUI's :e does, the name of the
// blog's session, which a post belongs to as the résumé's pages belong to the site's, and the labels of a post's
// facts.
type Frame struct {
	DarkTheme    string
	CopyWords    CopyWords
	SSH          Way
	ReadingGives string
	WriteTo      string
	Projects     []document.Page
	Blog         string
	PostLabels   PostLabels
}

// PostLabels are the site's words a face labels a post's facts with: the version a reader reads, and its tags.
type PostLabels struct {
	Version string
	Tags    string
}

// CopyWords are the site's words for the button that copies a command or a block of code, which a face lays beside
// each it draws to copy.
type CopyWords struct {
	words content.Words
}

// Copyable is text the reader copies whole, a command they can run or a post's block of code, with the words of the
// button that copies it: the word on the button, the word it turns to once it has, its name, which names what it
// copies, since a page may hold more than one, and what the toast says once it has copied.
type Copyable struct {
	Text     string
	Copy     string
	Copied   string
	CopyName string
	Toast    string
}

// Way is a way a terminal reaches the site: the program it runs, which names it, its command whole to copy, and what
// a reader gets by it.
type Way struct {
	Program string
	Command Copyable
	Gives   string
}

// FrameOf is the frame of a site's pages.
func FrameOf(snapshot *content.Snapshot) Frame {
	words := snapshot.Words
	copyWords := CopyWords{words: words}
	return Frame{
		DarkTheme:    words.Text(content.WordDarkTheme),
		CopyWords:    copyWords,
		SSH:          copyWords.Way(snapshot.Profile.Site.SSH, words.Text(content.WordSSHGives)),
		ReadingGives: words.Text(content.WordReadingGives),
		WriteTo:      words.Text(content.WordWriteTo),
		Projects:     projectPages(snapshot.Profile.Projects),
		Blog:         words.Text(content.WordBlog),
		PostLabels:   PostLabels{Version: words.Text(content.WordVersion), Tags: words.Text(content.WordTags)},
	}
}

// projectPages are the pages of the projects that have one.
func projectPages(projects []content.Project) []document.Page {
	pages := []document.Page{}
	for _, project := range projectsWithPages(projects) {
		pages = append(pages, projectPage(project))
	}
	return pages
}

// Of is a command with the words of the button that copies it: the word on it and the word it turns to, then its name
// and what the toast says, which both name the command.
func (copyWords CopyWords) Of(command string) Copyable {
	return Copyable{
		Text:     command,
		Copy:     copyWords.words.Text(content.WordCopy),
		Copied:   copyWords.words.Text(content.WordCopied),
		CopyName: copyWords.words.Phrase(content.PhraseCopyCommand, command),
		Toast:    copyWords.words.Phrase(content.PhraseCopiedCommand, command),
	}
}

// CodeOf is a block of code with the words of the button that copies it: the word on it and the word it turns to, as
// a command's; its name, which names the block by its number among a page's, counted from one, so a screen reader
// tells two apart; and what the toast says, which names the code by what it is, not by its lines.
func (copyWords CopyWords) CodeOf(code string, number int) Copyable {
	return Copyable{
		Text:     code,
		Copy:     copyWords.words.Text(content.WordCopy),
		Copied:   copyWords.words.Text(content.WordCopied),
		CopyName: copyWords.words.Phrase(content.PhraseCopyCode, strconv.Itoa(number)),
		Toast:    copyWords.words.Text(content.WordCopiedCode),
	}
}

// Way is the way a command reaches the site, giving what the words say: named by its program, the command's first
// word.
func (copyWords CopyWords) Way(command, gives string) Way {
	program, _, _ := strings.Cut(command, programEnd)
	return Way{Program: program, Command: copyWords.Of(command), Gives: gives}
}
