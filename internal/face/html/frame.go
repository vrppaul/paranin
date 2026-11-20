package html

import (
	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// The id of the page's main part, where the link past the top bar leads
const mainID = "main"

// frame is what every page of a site lays round its main part, worked out once for the site (DESIGN.md, TopBar,
// SkipLink and Footer): the link past the top bar, its words and where it leads; the site's name and where it begins;
// a link to each of the résumé's sections; the theme button's name; the words of the chip that copies a command; the
// footer's ways the site is served that are the same on every page, the command that reaches it over ssh, which the
// first screen offers as a chip, and the index agents read; and the words of the first screen's button that writes to
// the author.
type frame struct {
	skipToContent string
	toMain        string
	siteName      string
	start         string
	sections      []sectionLink
	darkTheme     string
	commandWords  compose.CommandWords
	ssh           compose.Command
	agentsIndex   string
	writeTo       string
}

// sectionLink is the top bar's link to a section of the résumé: its title, and its place on the résumé.
type sectionLink struct {
	title   string
	address string
}

// frameOf is the frame of a site's pages: its sections led to on the résumé, by its path from the site's root.
func frameOf(site routes.Site, composed compose.Frame) frame {
	resume := document.Page{Kind: document.PageResume, Slug: "", Tab: ""}
	sections := make([]sectionLink, 0, len(composed.Sections))
	for _, title := range composed.Sections {
		sections = append(sections, sectionLink{title: title, address: routes.PlaceReference(resume, routes.FormatChosen, document.SlugOf(title))})
	}
	return frame{
		skipToContent: composed.SkipToContent,
		toMain:        routes.Place(mainID),
		siteName:      site.Host(),
		start:         routes.Reference(resume, routes.FormatChosen),
		sections:      sections,
		darkTheme:     composed.DarkTheme,
		commandWords:  composed.CommandWords,
		ssh:           composed.SSH,
		agentsIndex:   routes.Reference(document.Page{Kind: document.PageIndex, Slug: "", Tab: ""}, routes.FormatMarkdown),
		writeTo:       composed.WriteTo,
	}
}
