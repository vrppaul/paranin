package compose

import (
	"time"

	"github.com/vrppaul/paranin/internal/document"
)

// IndexedPage is a page a search engine may index, as a sitemap names it: the page, and the day its text last
// changed, read by its date, the zero time where it has none.
type IndexedPage struct {
	Page    document.Page
	Changed time.Time
}

// IndexedPages are the pages the site serves, in every language, that a search engine may index (document.Head's
// Indexed), each by the day its text last changed (docs/SEO.md, Crawling).
func (site Site) IndexedPages() []IndexedPage {
	changed := site.changedDays()
	indexed := []IndexedPage{}
	for _, served := range site.served {
		if page := served.Head.Page; served.Head.Indexed() {
			indexed = append(indexed, IndexedPage{Page: page, Changed: changed[page]})
		}
	}
	return indexed
}

// changedDays are the days the text of the blogs' pages last changed, by page: a post's, its current version's, as its
// record has it; a list's, the latest of those among the posts it names. The résumé, a project's page and their tabs
// have none, since a wrong day is worse than none.
func (site Site) changedDays() map[document.Page]time.Time {
	days := map[document.Page]time.Time{}
	for _, blog := range site.Blogs() {
		for _, list := range blog.Lists {
			days[list.Head.Page] = blog.listsChanged
		}
		for _, post := range blog.Posts {
			days[post.Document.Head.Page] = changedOn(post.Record)
		}
	}
	return days
}

// changedOn is when the text of a page with a record last changed: when a post's current version came out; none for
// the résumé's and a project's, which no day of the content dates. gochecksumtype holds the switch to every kind.
func changedOn(record Record) time.Time {
	switch typed := record.(type) {
	case PostRecord:
		return typed.VersionOut
	case PersonRecord, WorkRecord:
	}
	return time.Time{}
}
