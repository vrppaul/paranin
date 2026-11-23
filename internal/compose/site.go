package compose

import (
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// Site is every page a snapshot's site serves, composed once, so a post's links read the pages its hosts draw.
type Site struct {
	// Index is the list of the site's pages an agent reads first
	Index document.Document
	// Pages are the résumé and each project's page that has one, each with its record
	Pages []Composed
	// Lists are the blog's list by year and by series
	Lists []document.Document
	// Posts are the posts the site serves, in the snapshot's order, each link inside the site read into its page
	Posts []document.Document
	// paths are the pages served by the path a link inside the site reaches each by
	paths map[string]document.Page
	// served are the documents of every page the site serves, each tab of a page alone too
	served []document.Document
}

// SiteOf composes a snapshot's site: its pages first, then its posts, whose links lead among them by their paths.
func SiteOf(snapshot *content.Snapshot) Site {
	site := Site{Index: Index(snapshot), Pages: Pages(snapshot), Lists: Lists(snapshot), Posts: nil, paths: nil, served: nil}
	pages := site.pagesButPosts()
	site.paths = pagesAt(append(headsOf(pages), postHeads(snapshot)...))
	site.Posts = composedPosts(snapshot, site.paths)
	site.served = append(pages, site.Posts...)
	return site
}

// Documents are the documents of the pages a person reads, in their order, for a face that draws no record.
func (site Site) Documents() []document.Document {
	documents := make([]document.Document, 0, len(site.Pages))
	for _, page := range site.Pages {
		documents = append(documents, page.Document)
	}
	return documents
}

// pagesButPosts are the documents of every page the site serves but its posts: the index, the résumé and each
// project's page with each of their tabs alone, and the blog's lists.
func (site Site) pagesButPosts() []document.Document {
	pages := []document.Document{site.Index}
	for _, page := range site.Pages {
		pages = append(append(pages, page.Document), page.Tabs...)
	}
	return append(pages, site.Lists...)
}

// pagesAt are the pages of heads by the path a link inside the site reaches each by, where the reader chooses its
// format; a page served in none, as the index, has none.
func pagesAt(heads []document.Head) map[string]document.Page {
	pages := make(map[string]document.Page, len(heads))
	for _, head := range heads {
		if path := routes.Path(head.Page, routes.FormatChosen); path != "" {
			pages[path] = head.Page
		}
	}
	return pages
}

// headsOf are the heads of documents, in their order.
func headsOf(documents []document.Document) []document.Head {
	heads := make([]document.Head, 0, len(documents))
	for _, composed := range documents {
		heads = append(heads, composed.Head)
	}
	return heads
}
