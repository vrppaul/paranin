package html

import (
	"strings"
	"unicode"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// The id of the page's main part, where the link past the top bar leads, and what joins the words of a section's id
const (
	mainID           = "main"
	sectionIDJoining = "-"
)

// frame is what every page of a site lays round its main part, worked out once for the site (DESIGN.md, TopBar,
// SkipLink and Footer): the link past the top bar, its words and where it leads; the site's name and where it begins;
// a link to each of the résumé's sections; the theme button's name; and the footer's ways the site is served that
// are the same on every page, the command that reaches it over ssh and the index agents read.
type frame struct {
	skipToContent string
	toMain        string
	siteName      string
	start         string
	sections      []sectionLink
	darkTheme     string
	ssh           string
	agentsIndex   string
}

// sectionLink is the top bar's link to a section of the résumé: its title, and its place on the résumé.
type sectionLink struct {
	title   string
	address string
}

// frameOf is the frame of a site's pages: its sections led to on the résumé, by its path from the site's root.
func frameOf(site routes.Site, composed compose.Frame) frame {
	resume := document.Page{Kind: document.PageResume, Slug: ""}
	sections := make([]sectionLink, 0, len(composed.Sections))
	for _, title := range composed.Sections {
		sections = append(sections, sectionLink{title: title, address: routes.PlaceReference(resume, routes.FormatChosen, sectionID(title))})
	}
	return frame{
		skipToContent: composed.SkipToContent,
		toMain:        routes.Place(mainID),
		siteName:      site.Host(),
		start:         routes.Reference(resume, routes.FormatChosen),
		sections:      sections,
		darkTheme:     composed.DarkTheme,
		ssh:           composed.SSH,
		agentsIndex:   routes.Reference(document.Page{Kind: document.PageIndex, Slug: ""}, routes.FormatMarkdown),
	}
}

// sectionID is the id a section of the résumé stands at, which the top bar leads to: its title's words in small
// letters, joined by hyphens, with nothing but letters and digits in them.
func sectionID(title string) string {
	words := strings.FieldsFunc(strings.ToLower(title), func(letter rune) bool { return !unicode.IsLetter(letter) && !unicode.IsDigit(letter) })
	return strings.Join(words, sectionIDJoining)
}
