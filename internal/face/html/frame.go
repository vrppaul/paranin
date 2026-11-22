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

// How a hint is known: its window's id, the way's program after the opening, and its title's, the window's id before
// the ending
const (
	hintIDOpening   = "hint-"
	hintTitleEnding = "-title"
)

// placeKind is what a place of the command line is, which its menu names it by, as the TUI's names it, and which the
// page's script reads: a tab of the page shown, or a project's page, which :e opens too.
type placeKind string

const (
	placeTab     placeKind = "tab"
	placeProject placeKind = "project"
)

// frame is what every page of a site lays round its main part, worked out once for the site (DESIGN.md, FootLine):
// the site's name and where it begins, which the foot line's session leads to; the theme button's name; the words of
// the button that copies a command or a block of code; the way a terminal reaches the site over ssh, whose command
// the first screen offers as a chip too; what reading a page in a terminal gives; the words of the first screen's
// button that writes to the author; the projects' pages, places of every page's command line; the name of the blog's
// session; and the labels of a post's facts.
type frame struct {
	siteName     string
	start        string
	darkTheme    string
	copyWords    compose.CopyWords
	ssh          compose.Way
	readingGives string
	writeTo      string
	projects     []place
	blog         string
	postLabels   compose.PostLabels
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

// place is somewhere the command line goes by its name, as the TUI's does: its name, what it is, the address it is
// shown at, and whether it is the page shown.
type place struct {
	name    string
	kind    placeKind
	address string
	current bool
}

// session is a part of the site as the foot line shows it, as tmux shows a session: its name, the address it leads to,
// none where its page is still to come, and whether the page shown belongs to it.
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
	resume := document.Page{Kind: document.PageResume, Slug: "", Tab: ""}
	return frame{
		siteName:     site.Host(),
		start:        routes.Reference(resume, routes.FormatChosen),
		darkTheme:    composed.DarkTheme,
		copyWords:    composed.CopyWords,
		ssh:          composed.SSH,
		readingGives: composed.ReadingGives,
		writeTo:      composed.WriteTo,
		projects:     projectPlaces(composed.Projects),
		blog:         composed.Blog,
		postLabels:   composed.PostLabels,
	}
}

// siteSession is the site's session alone, which the résumé's pages and the page of no page belong to. The blog's
// session leads to the blog's list, which is still to come, so until it is these pages leave it out.
func (around frame) siteSession() []session {
	return []session{{name: around.siteName, address: around.start, current: true}}
}

// projectPlaces are the projects' pages as places of the command line, each by its slug.
func projectPlaces(pages []document.Page) []place {
	places := make([]place, 0, len(pages))
	for _, page := range pages {
		places = append(places, place{name: page.Slug, kind: placeProject, address: routes.Reference(page, routes.FormatChosen), current: false})
	}
	return places
}

// footOf is the foot line of a page showing one of its document's tabs: every tab a window, the opening at the page's
// own address, and the page whole to read in a terminal and as markdown.
func (writer PageWriter) footOf(shown document.Page, tabs document.Tabs) footLine {
	whole := document.Page{Kind: shown.Kind, Slug: shown.Slug, Tab: ""}
	windows := []window{windowOf(firstWindowNumber, tabs.Opening.Slug, whole, shown)}
	for _, tab := range tabs.Rest {
		page := document.Page{Kind: shown.Kind, Slug: shown.Slug, Tab: tab.Slug}
		windows = append(windows, windowOf(firstWindowNumber+len(windows), tab.Slug, page, shown))
	}
	return footLineOf(writer.frame, writer.site, writer.frame.siteSession(), windows, whole)
}

// footLineOf is a frame's foot line with sessions and windows, offering a page of the site to read: over ssh, by curl
// at its address, and as markdown.
func footLineOf(around frame, site routes.Site, sessions []session, windows []window, page document.Page) footLine {
	markdown := routes.Reference(page, routes.FormatMarkdown)
	reading := around.copyWords.Way(site.ReadCommand(page), around.readingGives)
	return footLine{
		frame:        around,
		sessions:     sessions,
		windows:      windows,
		hints:        []hint{hintOf(around.ssh), hintOf(reading)},
		markdown:     markdown,
		markdownName: path.Base(markdown),
		places:       placesOf(windows, around.projects),
	}
}

// ids are the names the foot line gives its parts, which no other part of its page may take: the command line's
// field, its menu and each place on it, the keys it names, each hint and its title, and the theme button's words.
func (foot footLine) ids() []string {
	ids := []string{commandFieldPart, placesPart, commandKeysPart, searchKeysPart, themeNameID}
	for _, offered := range foot.hints {
		ids = append(ids, offered.id, offered.titleID)
	}
	for index := range foot.places {
		ids = append(ids, placeID(index))
	}
	return ids
}

// placeID is the name of the place of the command line's menu at an index, counted from one, which the field names
// as the one chosen.
func placeID(index int) string {
	return placeIDOpening + strconv.Itoa(index+firstPlaceNumber)
}

// hintOf is a way's hint, known by the program it runs.
func hintOf(way compose.Way) hint {
	id := hintIDOpening + way.Program
	return hint{way: way, id: id, titleID: id + hintTitleEnding}
}

// placesOf are the places of a page's command line: each window, then each project's page, the page shown where a
// window at its address is. A project's page is a place even where a window takes its name, since :e opens it by its
// slug, as the TUI's does; the page's script leaves a name to the first place that has it.
func placesOf(windows []window, projects []place) []place {
	places := make([]place, 0, len(windows)+len(projects))
	for _, shown := range windows {
		places = append(places, place{name: shown.name, kind: placeTab, address: shown.address, current: shown.current})
	}
	for _, project := range projects {
		project.current = slices.ContainsFunc(windows, func(shown window) bool { return shown.current && shown.address == project.address })
		places = append(places, project)
	}
	return places
}

// windowOf is a tab as a window by its number and name, leading to the page that shows it, current where that page is
// the one shown.
func windowOf(number int, name string, page, shown document.Page) window {
	return window{
		number:  strconv.Itoa(number),
		name:    name,
		address: routes.Reference(page, routes.FormatChosen),
		current: page == shown,
	}
}
