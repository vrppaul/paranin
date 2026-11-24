package feed

import (
	"strings"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// The lines of robots.txt (RFC 9309, docs/SEO.md, Crawling): the index for a language model in a comment, since no page
// links it and no rule names it; every crawler let in, those that train a model among them, and nothing disallowed, so
// a crawler reaches the noindex of every page that has one; a blank line closing the group; and the sitemap
const (
	indexComment = "# The site for a language model: "
	everyCrawler = "User-agent: *"
	everyPath    = "Allow: /"
	groupEnd     = ""
	sitemapLine  = "Sitemap: "
)

// Robots writes robots.txt, naming the index and the sitemap by their whole addresses as the site writes them.
func Robots(site routes.Site) string {
	lines := []string{
		indexComment + site.Address(document.PageOf(document.PageIndex), routes.FormatMarkdown),
		everyCrawler,
		everyPath,
		groupEnd,
		sitemapLine + site.CrawlerAddress(routes.CrawlerSitemap),
	}
	return strings.Join(lines, lineEnd) + lineEnd
}
