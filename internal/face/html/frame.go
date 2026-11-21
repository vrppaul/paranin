package html

import (
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// The number of the foot line's first window, as the TUI's keys choose its tabs from one
const firstWindowNumber = 1

// What parts a command's program from the words after it
const programEnd = " "

// placeKind is what a place of the command line is, which its menu names it by, as the TUI's names it, and which the
// page's script reads: a tab of the page shown, or a project's page, which :e opens too.
type placeKind string

const (
	placeTab     placeKind = "tab"
	placeProject placeKind = "project"
)

// frame is what every page of a site lays round its main part, worked out once for the site (DESIGN.md, FootLine):
// the site's name and where it begins, which the foot line's session leads to; the theme button's name; the words of
// the chip that copies a command; the command that reaches the site over ssh, which the first screen offers as a chip;
// the programs that reach the site and read a page in a terminal, which the foot line names; the words of the first
// screen's button that writes to the author; and the projects' pages, places of every page's command line.
type frame struct {
	siteName       string
	start          string
	darkTheme      string
	commandWords   compose.CommandWords
	ssh            compose.Command
	sshProgram     string
	readingProgram string
	writeTo        string
	projects       []place
}

// footLine is a page's foot line (DESIGN.md, FootLine), as tmux's status line: the frame's session and programs, the
// document's tabs as windows, the path of the page's markdown and its file's name; and the places the command line
// over it goes to.
type footLine struct {
	frame        frame
	windows      []window
	markdown     string
	markdownName string
	places       []place
}

// place is somewhere the command line goes by its name, as the TUI's does: its name, what it is, the address it is
// shown at, and whether it is the page shown.
type place struct {
	name    string
	kind    placeKind
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
		siteName:       site.Host(),
		start:          routes.Reference(resume, routes.FormatChosen),
		darkTheme:      composed.DarkTheme,
		commandWords:   composed.CommandWords,
		ssh:            composed.SSH,
		sshProgram:     programOf(composed.SSH.Text),
		readingProgram: programOf(site.ReadCommand(resume)),
		writeTo:        composed.WriteTo,
		projects:       projectPlaces(composed.Projects),
	}
}

// projectPlaces are the projects' pages as places of the command line, each by its slug.
func projectPlaces(pages []document.Page) []place {
	places := make([]place, 0, len(pages))
	for _, page := range pages {
		places = append(places, place{name: page.Slug, kind: placeProject, address: routes.Reference(page, routes.FormatChosen), current: false})
	}
	return places
}

// programOf is the program a command runs: its first word.
func programOf(command string) string {
	program, _, _ := strings.Cut(command, programEnd)
	return program
}

// footOf is the foot line of a page showing one of its document's tabs: every tab a window, the opening at the page's
// own address, and the markdown of the page whole.
func (writer PageWriter) footOf(shown document.Page, tabs document.Tabs) footLine {
	whole := document.Page{Kind: shown.Kind, Slug: shown.Slug, Tab: ""}
	windows := []window{windowOf(firstWindowNumber, tabs.Opening.Slug, whole, shown)}
	for _, tab := range tabs.Rest {
		page := document.Page{Kind: shown.Kind, Slug: shown.Slug, Tab: tab.Slug}
		windows = append(windows, windowOf(firstWindowNumber+len(windows), tab.Slug, page, shown))
	}
	return footLineOf(writer.frame, windows, whole)
}

// footLineOf is a frame's foot line with windows, offering the markdown of a page.
func footLineOf(around frame, windows []window, page document.Page) footLine {
	markdown := routes.Reference(page, routes.FormatMarkdown)
	return footLine{frame: around, windows: windows, markdown: markdown, markdownName: path.Base(markdown), places: placesOf(windows, around.projects)}
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
