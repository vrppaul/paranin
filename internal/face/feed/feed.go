// Package feed writes what a machine polls rather than a person reads: the blog's feeds, for a feed reader to learn
// what came out (RFC 4287, docs/FOUNDATION.md §5.4, Feeds), and robots.txt and the sitemap, for a crawler to learn what
// to read of the site (docs/SEO.md, Crawling). Each is a format of its own that no face a person reads draws.
package feed

import (
	"encoding/xml"
	"time"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// What a feed is written as: the relation of a feed's link to itself, and of a link to the page a person reads what it
// holds on; the kind of text a post's content is, HTML a reader draws; and how a moment is written
const (
	selfRelation      = "self"
	alternateRelation = "alternate"
	htmlContent       = "html"
	momentLayout      = time.RFC3339
)

// How the package's files are laid out: each element of an XML document indented, so a feed and the sitemap read in a
// diff, and the end of a line, which ends each line of robots.txt and each XML document
const (
	indent  = "  "
	lineEnd = "\n"
)

// writtenXML is a feed or the sitemap written as an XML document: its header, each element indented, and a line's end.
// Each is a struct of strings encoding/xml writes, so no text reaches the XML unescaped. Ways weighed: a template,
// whose text could slip past the escaping; a library, which brings a dependency for what encoding/xml writes.
func writtenXML[Written atomFeed | urlSet](written Written) string {
	// a struct of strings never fails to marshal
	marshalled, _ := xml.MarshalIndent(written, "", indent)
	return xml.Header + string(marshalled) + lineEnd
}

// BodyWriter writes a post's body as HTML a feed reader shows, every address whole: the html face's work.
type BodyWriter interface {
	Body(post document.Document) string
}

// atomFeed is a feed as Atom writes it: in its namespace and its language, named for good, its title and line, when it
// last changed, its link to itself and to the page a person reads what it holds on, its author, and its entries.
type atomFeed struct {
	XMLName  xml.Name    `xml:"http://www.w3.org/2005/Atom feed"`
	Language string      `xml:"http://www.w3.org/XML/1998/namespace lang,attr"`
	ID       string      `xml:"id"`
	Title    string      `xml:"title"`
	Subtitle string      `xml:"subtitle,omitempty"`
	Updated  string      `xml:"updated"`
	Links    []atomLink  `xml:"link"`
	Author   atomAuthor  `xml:"author"`
	Entries  []atomEntry `xml:"entry"`
}

// atomLink is a link of a feed or an entry: how it relates to them, the media type it leads to, and its address.
type atomLink struct {
	Relation  string `xml:"rel,attr"`
	MediaType string `xml:"type,attr"`
	Address   string `xml:"href,attr"`
}

// atomAuthor is who writes a feed: their name, and the address of the page that is theirs.
type atomAuthor struct {
	Name    string `xml:"name"`
	Address string `xml:"uri"`
}

// atomEntry is a post as Atom writes it: named for good, its title, its link to the page a person reads it on, when it
// came out and when its current version did, its lead, and its body.
type atomEntry struct {
	ID        string      `xml:"id"`
	Title     string      `xml:"title"`
	Link      atomLink    `xml:"link"`
	Published string      `xml:"published"`
	Updated   string      `xml:"updated"`
	Summary   string      `xml:"summary"`
	Content   atomContent `xml:"content"`
}

// atomContent is a post's body in a feed: HTML, which encoding/xml escapes as text, for a reader to draw.
type atomContent struct {
	Kind string `xml:"type,attr"`
	Text string `xml:",chardata"`
}

// Atom writes a feed in Atom, its addresses whole and its names for good as the site writes them, each post's body as
// bodies write it.
func Atom(feed compose.Feed, site routes.Site, bodies BodyWriter) string {
	entries := make([]atomEntry, 0, len(feed.Posts))
	for _, post := range feed.Posts {
		head := post.Document.Head
		entries = append(entries, atomEntry{
			ID:        site.TagURI(routes.NumberPath(head.Page.Translation, head.Number)),
			Title:     document.PlainText(head.Title),
			Link:      atomLink{Relation: alternateRelation, MediaType: routes.PageType, Address: site.Address(head.Page, routes.FormatChosen)},
			Published: post.Published.Format(momentLayout),
			Updated:   post.Updated.Format(momentLayout),
			Summary:   document.PlainText(head.Description),
			Content:   atomContent{Kind: htmlContent, Text: bodies.Body(post.Document)},
		})
	}
	return writtenXML(atomFeed{
		// the field's tag names the element, in Atom's namespace
		XMLName:  xml.Name{Space: "", Local: ""},
		Language: feed.Language,
		ID:       site.TagURI(routes.FeedPath(feed.Target)),
		Title:    document.PlainText(feed.Title),
		Subtitle: document.PlainText(feed.Subtitle),
		Updated:  feed.Updated.Format(momentLayout),
		Links: []atomLink{
			{Relation: selfRelation, MediaType: routes.FeedType, Address: site.Address(feed.Target, routes.FormatChosen)},
			{Relation: alternateRelation, MediaType: routes.PageType, Address: site.Address(feed.Shown, routes.FormatChosen)},
		},
		Author:  atomAuthor{Name: feed.Author, Address: site.Address(feed.AuthorPage, routes.FormatChosen)},
		Entries: entries,
	})
}
