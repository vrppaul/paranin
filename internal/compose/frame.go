package compose

import (
	"slices"
	"strconv"
	"strings"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// What parts a command's program from the words after it
const programEnd = " "

// KeysPlace is the name the TUI's window of keys goes by on every page's command line, as Neovim's :help opens its
// own, which no heading's anchor takes
const KeysPlace = "help"

// OpenedPage is a page :e opens by a name, as a place of the command line, and whether Tab offers it.
type OpenedPage struct {
	Place   document.Place
	Offered bool
}

// Frame is what a face lays round the documents it draws, on every page alike, in the site's words: the name of the
// theme button, which is pressed while the page is dark, the words of the button that copies a command or a block of
// code, the way a terminal reaches the site over ssh, what reading a page in a terminal gives, whose command a face
// writes with the page's address, the words of the button that writes to the author, the page of each project that
// has one, in the profile's order, a place of every page's command line by its slug, and every page :e opens, by each
// name it reads, the projects' among them, the site's sessions, the labels of a post's facts, the words of its series
// and of the panes at its foot, the name of its outline, and the words of the blog's list.
type Frame struct {
	DarkTheme    string
	CopyWords    CopyWords
	SSH          Way
	ReadingGives string
	WriteTo      string
	Projects     []document.Place
	Pages        []OpenedPage
	Sessions     []Session
	PostLabels   PostLabels
	SeriesWords  SeriesWords
	PaneWords    PaneWords
	Contents     string
	ListWords    ListWords
}

// Session is one of the site's sessions, as a face's foot line names them after tmux's (DESIGN.md, FootLine): its
// name, and the page it begins at, which its name leads to.
type Session struct {
	Name  string
	Start document.Page
}

// Holds is whether a page belongs to the session: a page of the blog to the blog's, any other to the site's.
func (session Session) Holds(page document.Page) bool {
	return session.Start.InBlog() == page.InBlog()
}

// Grouping is a way the blog's list groups its posts, as its head's Switch offers it: its word, and the list grouped
// so.
type Grouping struct {
	Words string
	List  document.Page
}

// Window is a window of the blog's session, as a face's foot line names it after tmux's: its name, and the page it
// shows.
type Window struct {
	Name string
	Page document.Page
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

// CopyButton is the words of a button that copies text whole, a command or a post's block of code: the word on the
// button, the word it turns to once it has, its name, which names what it copies, since a page may hold more than
// one, and what the toast says once it has copied.
type CopyButton struct {
	Copy     string
	Copied   string
	CopyName string
	Toast    string
}

// Copyable is a command the reader can run, with the button that copies it.
type Copyable struct {
	Text   string
	Button CopyButton
}

// Way is a way a terminal reaches the site: the program it runs, which names it, its command whole to copy, and what
// a reader gets by it.
type Way struct {
	Program string
	Command Copyable
	Gives   string
}

// frameOf is the frame of a site's pages, with the blog's pages it serves its audience.
func frameOf(snapshot *content.Snapshot, read blog) Frame {
	words := snapshot.Words
	copyWords := CopyWords{words: words}
	projects := projectPlaces(snapshot.Profile.Projects)
	return Frame{
		DarkTheme:    words.Text(content.WordDarkTheme),
		CopyWords:    copyWords,
		SSH:          copyWords.Way(snapshot.Profile.Site.SSH, words.Text(content.WordSSHGives)),
		ReadingGives: words.Text(content.WordReadingGives),
		WriteTo:      words.Text(content.WordWriteTo),
		Projects:     projects,
		Pages:        slices.Concat(projectPages(projects), blogPages(read, words)),
		Sessions: []Session{
			{Name: snapshot.Profile.Site.BaseURL.Host(), Start: document.PageOf(document.PageResume)},
			{Name: words.Text(content.WordBlog), Start: document.PageOf(document.PageBlog)},
		},
		PostLabels:  PostLabels{Version: words.Text(content.WordVersion), Series: words.Text(content.WordSeries), Tags: words.Text(content.WordTags)},
		SeriesWords: SeriesWords{words: words},
		PaneWords:   paneWords(words),
		Contents:    words.Text(content.WordContents),
		ListWords:   listWordsOf(words),
	}
}

// listWordsOf are the site's words of the blog's list.
func listWordsOf(words content.Words) ListWords {
	return ListWords{
		Window:   words.Text(content.WordListWindow),
		AllPosts: words.Text(content.WordAllPosts),
		By:       words.Text(content.WordBy),
		Years:    words.Text(content.WordYears),
		Series:   words.Text(content.WordSeries),
	}
}

// Groupings are the ways the blog's list groups its posts: by year, then by series.
func (listWords ListWords) Groupings() []Grouping {
	return []Grouping{{Words: listWords.Years, List: document.PageOf(document.PageBlog)}, {Words: listWords.Series, List: document.PageOf(document.PageBlogSeries)}}
}

// BlogWindows are the windows of the blog's session showing a list, and a post read beside it, PageNone where none
// is: the list first, by the site's word for its window, then the post by its slug.
func (listWords ListWords) BlogWindows(list, post document.Page) []Window {
	windows := []Window{{Name: listWords.Window, Page: list}}
	if post.Kind == document.PageNone {
		return windows
	}
	return append(windows, Window{Name: post.Slug, Page: post})
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

// PaneName is the name set into the top edge of a side of a post's foot holding a neighbour: the neighbour's part,
// where it is one of a series, "part 4"; else the side's own name, "older".
func (seriesWords SeriesWords) PaneName(side SideWords, neighbour *document.Neighbour) string {
	if neighbour != nil && neighbour.Part > 0 {
		return seriesWords.Part(neighbour.Part)
	}
	return side.Name
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

// projectPlaces are the pages of the projects that have one, each a place by its slug.
func projectPlaces(projects []content.Project) []document.Place {
	places := []document.Place{}
	for _, project := range projectsWithPages(projects) {
		places = append(places, document.Place{Name: project.Slug, Kind: document.PlaceProject, Page: projectPage(project), Section: ""})
	}
	return places
}

// projectPages are the projects' places as pages :e opens by their slugs, each one Tab offers.
func projectPages(projects []document.Place) []OpenedPage {
	pages := make([]OpenedPage, 0, len(projects))
	for _, project := range projects {
		pages = append(pages, OpenedPage{Place: project, Offered: true})
	}
	return pages
}

// blogPages are the blog's pages :e opens by the names a terminal reaches them by (D26): its list in each grouping and
// each post the list names, newest first, which Tab offers; then each post served and no list names, by its slug, every
// post served by its number, and the list by each of its filters, which :e opens though Tab offers none.
func blogPages(read blog, words content.Words) []OpenedPage {
	pages := []OpenedPage{}
	for _, grouping := range listWordsOf(words).Groupings() {
		pages = append(pages, openedPage(grouping.List, routes.Path(grouping.List, routes.FormatChosen), document.PlaceList, true))
	}
	for _, post := range read.newest {
		pages = append(pages, openedPage(postPage(post), routes.Path(postPage(post), routes.FormatChosen), document.PlacePost, true))
	}
	for _, post := range read.served {
		if !read.audience.lists(post) {
			pages = append(pages, openedPage(postPage(post), routes.Path(postPage(post), routes.FormatChosen), document.PlacePost, false))
		}
	}
	for _, post := range read.served {
		pages = append(pages, openedPage(postPage(post), routes.NumberPath(post.ID), document.PlacePost, false))
	}
	for _, filter := range read.filters {
		pages = append(pages, openedPage(filter.page, routes.Path(filter.page, routes.FormatChosen), document.PlaceList, false))
	}
	return pages
}

// openedPage is a page :e opens as a place of a kind, by the name a terminal reaches its path by, and whether Tab
// offers it.
func openedPage(page document.Page, path string, kind document.PlaceKind, offered bool) OpenedPage {
	return OpenedPage{Place: document.Place{Name: routes.TerminalName(path), Kind: kind, Page: page, Section: ""}, Offered: offered}
}

// Of is a command with the words of the button that copies it: the word on it and the word it turns to, then its name
// and what the toast says, which both name the command.
func (copyWords CopyWords) Of(command string) Copyable {
	return Copyable{
		Text: command,
		Button: CopyButton{
			Copy:     copyWords.words.Text(content.WordCopy),
			Copied:   copyWords.words.Text(content.WordCopied),
			CopyName: copyWords.words.Phrase(content.PhraseCopyCommand, command),
			Toast:    copyWords.words.Phrase(content.PhraseCopiedCommand, command),
		},
	}
}

// CodeButton is the button that copies a block of code: the word on it and the word it turns to, as a command's; its
// name, which names the block by its number among a page's, counted from one, so a screen reader tells two apart; and
// what the toast says, which names the code by what it is, not by its lines.
func (copyWords CopyWords) CodeButton(number int) CopyButton {
	return CopyButton{
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
