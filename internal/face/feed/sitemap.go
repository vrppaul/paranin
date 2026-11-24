package feed

import (
	"encoding/xml"
	"time"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/routes"
)

// How a sitemap writes the day a page changed: the date alone, which is all the content knows (W3C Datetime)
const dayLayout = time.DateOnly

// urlSet is a sitemap as sitemaps.org writes it: in its namespace, a url for each page. It names no priority or change
// frequency, which Google reads not at all, and no hreflang, which the pages carry (docs/SEO.md, Crawling).
type urlSet struct {
	XMLName xml.Name     `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	Pages   []sitemapURL `xml:"url"`
}

// sitemapURL is a page as a sitemap names it: its whole address, and the day its text last changed, where it has one.
type sitemapURL struct {
	Address string `xml:"loc"`
	Changed string `xml:"lastmod,omitempty"`
}

// Sitemap writes the sitemap of the pages a search engine may index, each by its whole address as the site writes it.
func Sitemap(pages []compose.IndexedPage, site routes.Site) string {
	named := make([]sitemapURL, 0, len(pages))
	for _, page := range pages {
		named = append(named, sitemapURL{Address: site.Address(page.Page, routes.FormatChosen), Changed: writtenDay(page.Changed)})
	}
	// the field's tag names the element, in the sitemap's namespace
	return writtenXML(urlSet{XMLName: xml.Name{Space: "", Local: ""}, Pages: named})
}

// writtenDay is a day as a sitemap writes it, empty for the zero time, which no day of the content is.
func writtenDay(day time.Time) string {
	if day.IsZero() {
		return ""
	}
	return day.Format(dayLayout)
}
