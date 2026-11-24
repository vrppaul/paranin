package html

import (
	"path"
	"slices"
	"strconv"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// The number of the foot line's first window, as the TUI's keys choose its tabs from one; and of the command line's
// first place, as a reader counts them
const (
	firstWindowNumber = 1
	firstPlaceNumber  = 1
)

// How a hint is known: its window's name, the way's program after the opening, and its title's, the window's name
// before the ending; each an id of the page's own (ownID)
const (
	hintIDOpening   = "hint-"
	hintTitleEnding = "-title"
)

// frame is what every page of a site lays round its main part, worked out once for the site (DESIGN.md, FootLine):
// the site, which writes every address a page leads to and names the site, and where it begins, which the page of no
// address leads to; the theme button's name; the locale Open Graph marks a page in; the words of
// the button that copies a command or a block of code; the way a terminal reaches the site over ssh, whose command the
// first screen offers as a chip too; what reading a page in a terminal gives; the words of the first screen's button
// that writes to the author; the projects' pages, places of every page's command line, the pages its :e opens, and the
// blog's folder, by the name a terminal writes and by its address, where :e opens a post a page does not list;
// the site's sessions, which the foot line names; the labels of a post's facts, the words of its versions, of its
// series and of the panes at its foot, the name of its outline, and the words round the blog's list. Each is in the
// language of the pages it frames.
type frame struct {
	site          routes.Site
	start         string
	darkTheme     string
	locale        string
	copyWords     compose.CopyWords
	ssh           compose.Way
	readingGives  string
	writeTo       string
	projects      []document.Place
	pages         []compose.OpenedPage
	folder        string
	folderAddress string
	sessions      []compose.Session
	postLabels    compose.PostLabels
	versionWords  compose.VersionWords
	seriesWords   compose.SeriesWords
	paneWords     compose.PaneWords
	contents      string
	listWords     compose.ListWords
}

// footLine is a page's foot line (DESIGN.md, FootLine), as tmux's status line: the sessions, the document's tabs as
// windows, the hints of the ways a terminal reaches the page, the path of the page's markdown and its file's name; and
// the places the command line over it goes to.
type footLine struct {
	frame        frame
	sessions     []session
	windows      []window
	hints        []hint
	markdown     string
	markdownName string
	places       []place
}

// hint is a way a terminal reaches the page, as the foot line offers it (DESIGN.md, Hint): its chip opens a window that
// holds its command, which the chip and the window's close button name by its id, and whose title names it.
type hint struct {
	way     compose.Way
	id      string
	titleID string
}

// place is somewhere the command line goes, as the TUI's does: its name, what it is, the address it leads to, whether
// it is the page shown, and whether :e opens it, where a name alone does not.
type place struct {
	name    string
	kind    document.PlaceKind
	address string
	current bool
	edited  bool
}

// session is a part of the site as the foot line shows it, as tmux shows a session: its name, the address it leads to,
// and whether the page shown belongs to it.
type session struct {
	name    string
	address string
	current bool
}

// window is a tab of the document as the foot line shows it, as tmux shows a window: its number, its name, the
// address it is shown at, and whether it is the one shown.
type window struct {
	number  string
	name    string
	address string
	current bool
}

// frameOf is the frame of a site's pages.
func frameOf(site routes.Site, composed compose.Frame) frame {
	return frame{
		site:          site,
		start:         site.Reference(document.PageOf(document.PageResume), routes.FormatChosen),
		darkTheme:     composed.DarkTheme,
		locale:        composed.Locale,
		copyWords:     composed.CopyWords,
		ssh:           composed.SSH,
		readingGives:  composed.ReadingGives,
		writeTo:       composed.WriteTo,
		projects:      composed.Projects,
		pages:         composed.Pages,
		folder:        routes.TerminalName(routes.BlogFolder()),
		folderAddress: routes.BlogFolder(),
		sessions:      composed.Sessions,
		postLabels:    composed.PostLabels,
		versionWords:  composed.VersionWords,
		seriesWords:   composed.SeriesWords,
		paneWords:     composed.PaneWords,
		contents:      composed.Contents,
		listWords:     composed.ListWords,
	}
}

// sessionsOf are the foot line's sessions on a page, each leading to where it begins, the one the page belongs to
// current.
func (around frame) sessionsOf(page document.Page) []session {
	sessions := make([]session, 0, len(around.sessions))
	for _, held := range around.sessions {
		sessions = append(sessions, session{name: held.Name, address: around.site.Reference(held.Start, routes.FormatChosen), current: held.Holds(page)})
	}
	return sessions
}

// footOf is the foot line of a page showing one of its document's tabs: every tab a window, the opening at the page's
// own address, the page whole to read in a terminal and as markdown, and its command line going to the page's own
// places too.
func (writer PageWriter) footOf(shown document.Page, tabs document.Tabs, own []document.Place) footLine {
	whole := document.Page{Kind: shown.Kind, Slug: shown.Slug, Tab: "", Revision: "", Translation: shown.Translation}
	around := writer.frame
	windows := []window{around.windowOf(firstWindowNumber, tabs.Opening.Slug, whole, shown)}
	for _, tab := range tabs.Rest {
		page := document.Page{Kind: shown.Kind, Slug: shown.Slug, Tab: tab.Slug, Revision: "", Translation: shown.Translation}
		windows = append(windows, around.windowOf(firstWindowNumber+len(windows), tab.Slug, page, shown))
	}
	return footLineOf(around, around.sessionsOf(whole), windows, whole, own)
}

// blogWindowsOf are the blog session's windows as the foot line shows them, numbered in their order, the one showing
// the page shown current.
func (around frame) blogWindowsOf(windows []compose.Window, shown document.Page) []window {
	shownWindows := make([]window, 0, len(windows))
	for _, held := range windows {
		shownWindows = append(shownWindows, around.windowOf(firstWindowNumber+len(shownWindows), held.Name, held.Page, shown))
	}
	return shownWindows
}

// footLineOf is a frame's foot line with sessions and windows, offering a page of the site to read: over ssh, by curl
// at its address, and as markdown; its command line going to the page's own places too.
func footLineOf(around frame, sessions []session, windows []window, page document.Page, own []document.Place) footLine {
	markdown := around.site.Reference(page, routes.FormatMarkdown)
	reading := around.copyWords.Way(around.site.ReadCommand(page), around.readingGives)
	return footLine{
		frame:        around,
		sessions:     sessions,
		windows:      windows,
		hints:        []hint{hintOf(around.ssh), hintOf(reading)},
		markdown:     markdown,
		markdownName: path.Base(markdown),
		places:       placesOf(windows, own, around, page),
	}
}

// placeID is the id of the place of the command line's menu at an index, counted from one, which the field names as
// the one chosen.
func placeID(index int) string {
	return ownID(placeIDOpening + strconv.Itoa(index+firstPlaceNumber))
}

// hintOf is a way's hint, known by the program it runs.
func hintOf(way compose.Way) hint {
	id := ownID(hintIDOpening + way.Program)
	return hint{way: way, id: id, titleID: id + hintTitleEnding}
}

// placesOf are the places of a page's command line: each window, then each of the page's own places, then each
// project's page, each by its name; then each page Tab offers after :e on the page shown (pageOffered). A project's
// page is a place even where a window takes its name, since :e opens it by its slug, as the TUI's does; the page's
// script leaves a name to the first place that has it.
func placesOf(windows []window, own []document.Place, around frame, shown document.Page) []place {
	places := make([]place, 0, len(windows)+len(own)+len(around.projects)+len(around.pages))
	for _, window := range windows {
		places = append(places, place{name: window.name, kind: document.PlaceTab, address: window.address, current: window.current, edited: false})
	}
	for _, named := range slices.Concat(own, around.projects) {
		places = append(places, around.placeOf(named, windows, false))
	}
	for _, opened := range around.pages {
		if pageOffered(opened, shown) {
			places = append(places, around.placeOf(opened.Place, windows, true))
		}
	}
	return places
}

// pageOffered is whether Tab offers a page :e opens on the page shown: every page it offers anywhere, but a post on the
// blog's own pages alone, so the blog's posts weigh on no other page, where :e opens them by the blog's folder (the
// command line's data-folder).
func pageOffered(opened compose.OpenedPage, shown document.Page) bool {
	return opened.Offered && (opened.Place.Kind != document.PlacePost || shown.InBlog())
}

// placeOf is a place of a page as the command line offers it, by its name or by :e: at its address, inside its page
// at its section's anchor where it names one; the page shown where a window at that address is.
func (around frame) placeOf(offered document.Place, windows []window, edited bool) place {
	address := around.site.Reference(offered.Page, routes.FormatChosen)
	if offered.Section != "" {
		address += fragmentMark + offered.Section
	}
	current := slices.ContainsFunc(windows, func(shown window) bool { return shown.current && shown.address == address })
	return place{name: offered.Name, kind: offered.Kind, address: address, current: current, edited: edited}
}

// windowOf is a tab as a window by its number and name, leading to the page that shows it, current where that page is
// the one shown.
func (around frame) windowOf(number int, name string, page, shown document.Page) window {
	return window{
		number:  strconv.Itoa(number),
		name:    name,
		address: around.site.Reference(page, routes.FormatChosen),
		current: page == shown,
	}
}
