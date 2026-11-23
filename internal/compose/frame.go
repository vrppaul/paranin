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
// blog's session, which a post belongs to as the résumé's pages belong to the site's, the labels of a post's facts,
// the words of its series and of the panes at its foot, the name of its outline, and the words of the blog's list.
type Frame struct {
	DarkTheme    string
	CopyWords    CopyWords
	SSH          Way
	ReadingGives string
	WriteTo      string
	Projects     []document.Page
	Blog         string
	PostLabels   PostLabels
	SeriesWords  SeriesWords
	PaneWords    PaneWords
	Contents     string
	ListWords    ListWords
}

// ListWords are the site's words a face lays round the blog's list: its window, the way to it beside a post's other
// posts, the word before how it groups its posts, and the name of each grouping, by year and by series.
type ListWords struct {
	Window   string
	AllPosts string
	By       string
	Years    string
	Series   string
}

// PostLabels are the site's words a face labels a post's facts with: the version a reader reads, its series, and its
// tags.
type PostLabels struct {
	Version string
	Series  string
	Tags    string
}

// SeriesWords are the site's words a face writes a post's series in, each with the values the series gives it.
type SeriesWords struct {
	words content.Words
}

// PaneWords are the site's words of the panes at a post's foot: each side's in a series and outside one, the panes'
// heading outside a series, and the word before where the post is discussed.
type PaneWords struct {
	Parts      PaneSides
	Posts      PaneSides
	OtherPosts string
	Discuss    string
}

// PaneSides are the words of the two sides of a post's foot, the posts before it and after it.
type PaneSides struct {
	Before SideWords
	After  SideWords
}

// SideWords are the words of one side of a post's foot: its name, where no part's number names it, and what it says
// where it holds no post.
type SideWords struct {
	Name string
	None string
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
		PostLabels:   PostLabels{Version: words.Text(content.WordVersion), Series: words.Text(content.WordSeries), Tags: words.Text(content.WordTags)},
		SeriesWords:  SeriesWords{words: words},
		PaneWords:    paneWords(words),
		Contents:     words.Text(content.WordContents),
		ListWords: ListWords{
			Window:   words.Text(content.WordListWindow),
			AllPosts: words.Text(content.WordAllPosts),
			By:       words.Text(content.WordBy),
			Years:    words.Text(content.WordYears),
			Series:   words.Text(content.WordSeries),
		},
	}
}

// paneWords are the words of the panes at a post's foot, in a language.
func paneWords(words content.Words) PaneWords {
	return PaneWords{
		Parts: PaneSides{
			Before: SideWords{Name: words.Text(content.WordPrevious), None: words.Text(content.WordFirstPart)},
			After:  SideWords{Name: words.Text(content.WordNext), None: words.Text(content.WordLastPart)},
		},
		Posts: PaneSides{
			Before: SideWords{Name: words.Text(content.WordOlder), None: words.Text(content.WordFirstPost)},
			After:  SideWords{Name: words.Text(content.WordNewer), None: words.Text(content.WordNewestPost)},
		},
		OtherPosts: words.Text(content.WordOtherPosts),
		Discuss:    words.Text(content.WordDiscuss),
	}
}

// PartOf is a part's place in its series, by its number and the count of the series' parts: "1 of 11".
func (seriesWords SeriesWords) PartOf(number, parts int) string {
	return seriesWords.words.Phrase(content.PhrasePartOf, strconv.Itoa(number), strconv.Itoa(parts))
}

// OtherParts is the heading of the panes at the foot of a part of a series, by the series' title.
func (seriesWords SeriesWords) OtherParts(title string) string {
	return seriesWords.words.Phrase(content.PhraseOtherParts, title)
}

// AllParts is the pick that lists every part of a series, by their count.
func (seriesWords SeriesWords) AllParts(parts int) string {
	return seriesWords.words.Phrase(content.PhraseAllParts, strconv.Itoa(parts))
}

// Part is a part of a series by its number: "part 4".
func (seriesWords SeriesWords) Part(number int) string {
	return seriesWords.words.Phrase(content.PhrasePart, strconv.Itoa(number))
}

// State is where a part stands for the reader of the part read, as a face writes it beside the part's title: the part
// read, the part to read next, a part to come, and nothing for any other part out. exhaustive holds the switch to
// every state.
func (seriesWords SeriesWords) State(state document.PartState) string {
	switch state {
	case document.PartReading:
		return seriesWords.words.Text(content.WordReading)
	case document.PartNext:
		return seriesWords.words.Text(content.WordNext)
	case document.PartToCome:
		return seriesWords.words.Text(content.WordSoon)
	case document.PartOut:
	}
	return ""
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
