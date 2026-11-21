package compose

import (
	"strings"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// What parts a command's program from the words after it
const programEnd = " "

// Frame is what a face lays round the documents it draws, on every page alike, in the site's words: the name of the
// theme button, which is pressed while the page is dark, the words of the button that copies a command, the way a
// terminal reaches the site over ssh, what reading a page in a terminal gives, whose command a face writes with the
// page's address, the words of the button that writes to the author, and the page of each project that has one, in
// the profile's order, which a face's command line goes to by slug as the TUI's :e does.
type Frame struct {
	DarkTheme    string
	CommandWords CommandWords
	SSH          Way
	ReadingGives string
	WriteTo      string
	Projects     []document.Page
}

// CommandWords are the site's words for the button that copies a command, which a face lays beside each command it
// draws to copy.
type CommandWords struct {
	words content.Words
}

// Command is a command the reader can run, with the words of the button that copies it: the word on the button, the
// word it turns to once it has, its name, which names the command, since a page may hold more than one, and what the
// toast says once it has copied.
type Command struct {
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
	Command Command
	Gives   string
}

// FrameOf is the frame of a site's pages.
func FrameOf(snapshot *content.Snapshot) Frame {
	words := snapshot.Words
	commandWords := CommandWords{words: words}
	return Frame{
		DarkTheme:    words.Text(content.WordDarkTheme),
		CommandWords: commandWords,
		SSH:          commandWords.Way(snapshot.Profile.Site.SSH, words.Text(content.WordSSHGives)),
		ReadingGives: words.Text(content.WordReadingGives),
		WriteTo:      words.Text(content.WordWriteTo),
		Projects:     projectPages(snapshot.Profile.Projects),
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
func (commandWords CommandWords) Of(text string) Command {
	return Command{
		Text:     text,
		Copy:     commandWords.words.Text(content.WordCopy),
		Copied:   commandWords.words.Text(content.WordCopied),
		CopyName: commandWords.words.Phrase(content.PhraseCopyCommand, text),
		Toast:    commandWords.words.Phrase(content.PhraseCopiedCommand, text),
	}
}

// Way is the way a command reaches the site, giving what the words say: named by its program, the command's first
// word.
func (commandWords CommandWords) Way(command, gives string) Way {
	program, _, _ := strings.Cut(command, programEnd)
	return Way{Program: program, Command: commandWords.Of(command), Gives: gives}
}
