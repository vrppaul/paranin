package compose

import (
	"maps"
	"slices"
	"time"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// Site is every page a snapshot's site serves an audience, composed once, so a post's links read the pages its hosts
// draw, with the frame round them and until when it holds.
type Site struct {
	// Index is the list of the site's pages an agent reads first
	Index document.Document
	// Pages are the résumé and each project's page that has one, each with its record
	Pages []Composed
	// Lists are the blog's list by year and by series
	Lists []document.Document
	// Filtered are the blog's list by year filtered by each tag its head names, then by each series with a part out:
	// views of the list, which a reader opens from its head and no index lists
	Filtered []document.Document
	// Posts are the posts the site serves, in the snapshot's order, each link inside the site read into its page
	Posts []document.Document
	// Versions are the older versions of the posts the site serves, each a page of its own, which no list names and no
	// search engine indexes, in the posts' order and each post's newest first
	Versions []document.Document
	// Pictures are the pictures of the posts the site serves, each at its file beside its post, in no order
	Pictures []PostPicture
	// Previews are the posts the site serves at their preview links alone, in the snapshot's order, each with its
	// pictures: no list, page or session names them
	Previews []Preview
	// Feeds are the blog's feeds, its own and then each series', each holding its newest posts the lists name; none
	// where the lists name none
	Feeds []Feed
	// Frame is what a face lays round the pages, the pages :e opens among it
	Frame Frame
	// Until is when the site changes as the next post that comes later comes out, which a host keeps what it drew
	// until; the zero time where none comes later, so the site holds as long as its snapshot
	Until time.Time
	// paths are the pages served, and the feeds, by the path a link inside the site reaches each by
	paths map[string]document.Target
	// served are the documents of every page the site serves, each tab of a page alone and each older version of a
	// post too
	served []document.Document
	// postsServed are the posts the site serves as the content holds them, whose links Check reads
	postsServed []content.Post
}

// SiteOf composes a snapshot's site, as Prepare read it, as an audience reads it: its pages first, then its posts,
// their older versions and its previews, whose links lead among them and the feeds by their paths, and the feeds of
// its posts.
func SiteOf(prepared *Prepared, audience Audience) Site {
	snapshot := prepared.snapshot
	read := blogOf(snapshot, audience)
	site := Site{
		Index: indexOf(snapshot, read), Pages: pagesOf(snapshot, read), Lists: blogLists(read, snapshot.Words),
		Filtered: filteredLists(read, snapshot.Words), Posts: nil, Versions: nil, Pictures: picturesOf(read), Previews: nil, Feeds: nil,
		Frame: frameOf(snapshot, read), Until: audience.nextChange(snapshot.Posts), paths: nil, served: nil, postsServed: read.served,
	}
	pages := site.pagesButPosts()
	site.paths = pagesAt(append(headsOf(pages), postHeads(read, snapshot.Words)...))
	maps.Copy(site.paths, feedsAt(read.feeds))
	site.Posts = composedPosts(prepared, read, site.paths)
	site.Feeds = read.withPosts(site.Posts)
	site.Versions = olderDocuments(prepared, read, site.paths)
	site.Previews = previewsOf(prepared, read, site.paths)
	site.served = slices.Concat(pages, site.Posts, site.Versions)
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
// project's page with each of their tabs alone, and the blog's lists, filtered or not.
func (site Site) pagesButPosts() []document.Document {
	pages := []document.Document{site.Index}
	for _, page := range site.Pages {
		pages = append(append(pages, page.Document), page.Tabs...)
	}
	return slices.Concat(pages, site.Lists, site.Filtered)
}

// pagesAt are the pages of heads by the path a link inside the site reaches each by, where the reader chooses its
// format; a page served in none, as the index, has none.
func pagesAt(heads []document.Head) map[string]document.Target {
	pages := make(map[string]document.Target, len(heads))
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
