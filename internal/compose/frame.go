package compose

import (
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// Frame is what a face lays round the documents it draws, on every page alike, in the site's words: the name of the
// theme button, which is pressed while the page is dark, the words of the button that copies a command, the command
// that reaches the site over ssh, one of the ways it is served, with those words, the words of the button that
// writes to the author, and the page of each project that has one, in the profile's order, which a face's command
// line goes to by slug as the TUI's :e does.
type Frame struct {
	DarkTheme    string
	CommandWords CommandWords
	SSH          Command
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

// FrameOf is the frame of a site's pages.
func FrameOf(snapshot *content.Snapshot) Frame {
	words := snapshot.Words
	commandWords := CommandWords{words: words}
	return Frame{
		DarkTheme:    words.Text(content.WordDarkTheme),
		CommandWords: commandWords,
		SSH:          commandWords.Of(snapshot.Profile.Site.SSH),
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
