package compose

import (
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// Frame is what a face lays round the documents it draws, on every page alike, in the site's words: the words of
// the link past it to the page's main part, the titles of the résumé's sections, which it leads to in their order,
// the name of the theme button, which is pressed while the page is dark, the words of the button that copies a
// command, the command that reaches the site over ssh, one of the ways it is served, with those words, and the words
// of the button that writes to the author.
type Frame struct {
	SkipToContent string
	Sections      []string
	DarkTheme     string
	CommandWords  CommandWords
	SSH           Command
	WriteTo       string
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

// FrameOf is the frame of a site's pages, its sections those of the résumé among the pages composed; none where they
// hold no résumé.
func FrameOf(snapshot *content.Snapshot, pages []Composed) Frame {
	words := snapshot.Words
	sections := []string{}
	for _, page := range pages {
		if page.Document.Head.Page.Kind != document.PageResume {
			continue
		}
		for _, block := range page.Document.Blocks {
			if section, isSection := block.(document.Section); isSection {
				sections = append(sections, document.PlainText(section.Title))
			}
		}
	}
	commandWords := CommandWords{words: words}
	return Frame{
		SkipToContent: words.Text(content.WordSkipToContent),
		Sections:      sections,
		DarkTheme:     words.Text(content.WordDarkTheme),
		CommandWords:  commandWords,
		SSH:           commandWords.Of(snapshot.Profile.Site.SSH),
		WriteTo:       words.Phrase(content.PhraseWriteTo, snapshot.Profile.Person.GivenName),
	}
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
