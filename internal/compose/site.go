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
// draw, in every language, and until when it holds.
type Site struct {
	// Index is the list of the site's pages an agent reads first
	Index document.Document
	// Pages are the résumé and each project's page that has one, each with its record
	Pages []Composed
	// Own is the blog in the site's own language, English
	Own Blog
	// Translations are the blog in each other language the site has words for, in their tags' order
	Translations []Blog
	// Until is when the site changes as the next post that comes later comes out, in any language, which a host keeps
	// what it drew until; the zero time where none comes later, so the site holds as long as its snapshot
	Until time.Time
	// Addresses write the site's addresses whole, from where its profile says it is served
	Addresses routes.Site
	// paths are the pages served, and the feeds, in every language, by the path a link inside the site reaches each by
	paths map[string]document.Target
	// served are the documents of every page the site serves, in every language, each tab and older version among them
	served []document.Document
}

// Blog is the blog in one language, its pages in its words with the frame round them.
type Blog struct {
	// Lists are the blog's list by year and by series
	Lists []document.Document
	// Filtered are the blog's list by year filtered by each tag its head names: views of the list, which a reader opens
	// from its head and no index lists
	Filtered []document.Document
	// Posts are the posts the blog serves, in the snapshot's order, each with its record, each link inside the site read
	// into its page
	Posts []Composed
	// Versions are the older versions of the posts the blog serves, each a page of its own, which no list names and no
	// search engine indexes, in the posts' order and each post's newest first
	Versions []document.Document
	// Pictures are the pictures of the posts the blog serves, each at its file beside its post, in no order
	Pictures []PostPicture
	// Feeds are the blog's feeds, its own and then each series', each holding its newest posts the lists name; none
	// where the lists name none
	Feeds []Feed
	// Frame is what a face lays round the pages in the blog's language, the pages :e opens among it
	Frame Frame
	// served are the documents of every page the blog serves, each older version of a post too
	served []document.Document
	// postsServed are the posts the blog serves as the content holds them, whose links Check reads
	postsServed []content.Post
	// listsChanged is when the lists' text last changed, as the latest current version among the posts they name came
	// out; the zero time where they name none
	listsChanged time.Time
}

// SiteOf composes a snapshot's site, as Prepare read it, as an audience reads it: its pages and each blog's but their
// posts first, then the posts and their older versions, whose links lead among them all.
func SiteOf(prepared *Prepared, audience Audience) Site {
	snapshot := prepared.snapshot
	own, translations := blogsOf(snapshot, audience)
	edited := editedPages(snapshot, append([]blog{own}, translations...))
	site := Site{
		Index: indexOf(snapshot, own), Pages: pagesOf(snapshot, own), Own: blogPagesOf(snapshot, own, edited), Translations: nil,
		Until:     audience.nextChange(allPosts(snapshot)),
		Addresses: routes.New(snapshot.Profile.Site.BaseURL, snapshot.Profile.Site.TagAuthority, snapshot.TranslatedInto()),
		paths:     map[string]document.Target{}, served: nil,
	}
	for _, read := range translations {
		site.Translations = append(site.Translations, blogPagesOf(snapshot, read, edited))
	}
	composing := []blogComposing{{composed: &site.Own, read: own}}
	for index, read := range translations {
		composing = append(composing, blogComposing{composed: &site.Translations[index], read: read})
	}
	site.served = site.ownPages()
	heads := headsOf(site.served)
	for _, blog := range composing {
		heads = slices.Concat(heads, headsOf(blog.composed.lists()), postHeads(blog.read))
		maps.Copy(site.paths, feedsAt(blog.read.feeds))
	}
	maps.Copy(site.paths, pagesAt(heads))
	for _, blog := range composing {
		blog.composed.composePosts(prepared, blog.read, site.paths)
		site.served = append(site.served, blog.composed.served...)
	}
	return site
}

// blogComposing is a blog of the site being composed, and what it is read from.
type blogComposing struct {
	composed *Blog
	read     blog
}

// Blogs are the blog in every language, English first.
func (site Site) Blogs() []Blog {
	return append([]Blog{site.Own}, site.Translations...)
}

// blogPagesOf is a blog as a snapshot reads it in a language, but its posts: its lists, filtered or not, its posts'
// pictures and its frame, with the pages :e opens in every language.
func blogPagesOf(snapshot *content.Snapshot, read blog, edited []OpenedPage) Blog {
	return Blog{
		Lists: blogLists(read), Filtered: filteredLists(read), Posts: nil, Versions: nil, Pictures: picturesOf(read), Feeds: nil,
		Frame: frameOf(snapshot, read, edited), served: nil, postsServed: read.served, listsChanged: latestVersionOut(read.listed),
	}
}

// Served are the documents of every page the blog serves, its lists, filtered or not, its posts and each older version
// of them, for a face that draws them all, as the TUI does.
func (composed Blog) Served() []document.Document {
	return composed.served
}

// composePosts composes a blog's posts, their feeds and older versions, their links read among paths.
func (composed *Blog) composePosts(prepared *Prepared, read blog, paths map[string]document.Target) {
	composed.Posts = composedPosts(prepared, read, paths)
	posts := DocumentsOf(composed.Posts)
	composed.Feeds = read.withPosts(posts)
	composed.Versions = olderDocuments(prepared, read, paths)
	composed.served = slices.Concat(composed.lists(), posts, composed.Versions)
}

// DocumentsOf are the documents of composed pages, in their order, for a face that draws no record.
func DocumentsOf(pages []Composed) []document.Document {
	documents := make([]document.Document, 0, len(pages))
	for _, page := range pages {
		documents = append(documents, page.Document)
	}
	return documents
}

// ownPages are the documents of the index, the résumé and each project's page, with each of their tabs alone.
func (site Site) ownPages() []document.Document {
	pages := []document.Document{site.Index}
	for _, page := range site.Pages {
		pages = append(pages, page.PageAndTabs()...)
	}
	return pages
}

// lists are the blog's lists, filtered or not.
//
//wrapper:names the pages of a blog that are its list, which the addresses and the posts' places read together
func (composed Blog) lists() []document.Document {
	return slices.Concat(composed.Lists, composed.Filtered)
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
