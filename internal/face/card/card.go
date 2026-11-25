// Package card draws each page's card, the picture a link's preview shows of it (DESIGN.md, Social cards; ADR-0005).
// It writes a card's facts from a composed page, its title, its line, a post's day and series, the way in, as a face
// writes a document, and draws a card by running Typst on its template (card.typ) with those facts as data, so no
// author's word is ever written into Typst's markup. The template holds the look and fits the words to it, since Typst
// measures what it sets. Which pages have a card is decided where a page's preview is (document.Head's Previewed).
package card

import (
	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// Facts are what a card shows of its page, in one of two looks, so the résumé and the blog read as two sides of the
// site: the résumé's side, the résumé, a project and the blog's list, a title in block letters over a line; and a post.
//
//sumtype:decl
type Facts interface {
	facts()
}

// Resume is the résumé's card: the author's name, which stands in block letters on the lines the profile breaks it
// into with the cursor after it, the author's title under it, and the way in.
type Resume struct {
	Name  string
	Line  string
	WayIn string
}

// Project is a project's card: its name, in block letters, its tagline under it, and the way in.
type Project struct {
	Name  string
	Line  string
	WayIn string
}

// List is the blog's list's card, by year or by series: the cards' mark, in block letters, the list's line under it in
// its language, and the way in.
type List struct {
	Mark  string
	Line  string
	WayIn string
}

// Post is a post's card: the cards' mark, in small block letters at the left of its header, the day the post came out
// as the page writes a day and its series' title at the right, none where its page shows none; its title, its lead
// under it, and the way in.
type Post struct {
	Mark   string
	Day    string
	Series string
	Title  string
	Lead   string
	WayIn  string
}

func (Resume) facts()  {}
func (Project) facts() {}
func (List) facts()    {}
func (Post) facts()    {}

// Writer writes the facts of the cards of a site's pages: the frame of each language, by the translation its pages are
// in, which writes a page's day and the command that reaches it; and the cards' mark, the blog's name in the site's own
// language, which stands in block letters on every language's card, since they draw no other alphabet.
type Writer struct {
	frames map[string]compose.Frame
	mark   string
}

// NewWriter writes the facts of the cards of a site's pages.
func NewWriter(site compose.Site) Writer {
	return Writer{frames: site.Frames(), mark: site.Own.Frame.BlogName}
}

// Card is a page's card to draw: the path its page is served at, which the cards' manifest names it by, and its facts.
type Card struct {
	Path  string
	Facts Facts
}

// CardsOf are the cards of every page of a site that has one of its own (FactsOf), in the order the site serves them.
func CardsOf(site compose.Site) []Card {
	writer := NewWriter(site)
	cards := []Card{}
	for _, served := range site.Served() {
		if facts, isOwn := writer.FactsOf(served); isOwn {
			cards = append(cards, Card{Path: routes.Path(served.Head.Page, routes.FormatChosen), Facts: facts})
		}
	}
	return cards
}

// PageOf is the page whose card a page of a head shows (DESIGN.md, Social cards): its own, a tab its page's; none, the
// page of no kind, on a page no preview shows, an older version, a filtered list, a draft.
func PageOf(head document.Head) document.Page {
	if !head.Previewed() {
		return document.PageOf(document.PageNone)
	}
	whole := head.Page
	whole.Tab = ""
	return whole
}

// FactsOf are the facts of a page's card, where the document is the page that card stands for: none for a tab, which
// shows its page's, or a page no preview shows. exhaustive holds the switch to every kind of page.
func (writer Writer) FactsOf(composed document.Document) (Facts, bool) {
	head := composed.Head
	if PageOf(head) != head.Page {
		return nil, false
	}
	frame := writer.frames[head.Page.Translation]
	wayIn := frame.SSHCommandTo(head)
	switch head.Page.Kind {
	case document.PageResume:
		return Resume{Name: document.PlainText(head.Title), Line: document.PlainText(head.Subtitle), WayIn: wayIn}, true
	case document.PageProject:
		return Project{Name: document.PlainText(head.Title), Line: document.PlainText(head.Subtitle), WayIn: wayIn}, true
	case document.PageBlog, document.PageBlogSeries:
		return List{Mark: writer.mark, Line: document.PlainText(head.Lead), WayIn: wayIn}, true
	case document.PagePost:
		return writer.postFacts(head, frame.Days, wayIn)
	case document.PageNone, document.PageIndex, document.PageHealth, document.PageSkills, document.PageTagFilter:
	}
	return nil, false
}

// postFacts are the facts of a post's card, its current version's, its day written in its language's words: an older
// version and a draft show no preview. gochecksumtype holds the switch to every kind of version.
func (writer Writer) postFacts(head document.Head, days compose.DayWords, wayIn string) (Facts, bool) {
	switch version := head.Version.(type) {
	case document.VersionOut:
		return Post{
			Mark: writer.mark, Day: days.Of(version.Published), Series: seriesTitle(head.Post),
			Title: document.PlainText(head.Title), Lead: document.PlainText(head.Lead), WayIn: wayIn,
		}, true
	case document.VersionOlder, document.VersionDraft:
	}
	return nil, false
}

// seriesTitle is the title of the series a post's page shows, none where it shows none.
func seriesTitle(place *document.PostPlace) string {
	if place == nil || place.Series == nil {
		return ""
	}
	return document.PlainText(place.Series.Title)
}
