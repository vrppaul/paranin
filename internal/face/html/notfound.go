package html

import (
	"strconv"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// NotFound is the page of an address the site has no page at, for a browser: drawn for each path asked for, from
// the parts worked out once for the site.
type NotFound struct {
	// status is the status the host answers with, as HTTP writes it: 404 Not Found
	status string
	// language is the tag of the language of the site's words. The page's own words are HTTP's English status and
	// the site's name, so the two agree only while the site is in English alone (D12)
	language string
	// start is the path where the site begins, and siteName the name a link to it reads as
	start    string
	siteName string
}

// NewNotFound is the page of an address of no page of a site, in a language, titled by the status the host
// answers with.
func NewNotFound(site routes.Site, status, language string) NotFound {
	start := document.Page{Kind: document.PageResume, Slug: ""}
	return NotFound{status: status, language: language, start: routes.Reference(start, routes.FormatChosen), siteName: site.Host()}
}

// For is the page for a path asked for: its status, the path quoted as Go quotes a string, so no control character
// in it reaches the page, then a link to where the site begins.
func (notFound NotFound) For(path string) string {
	return rendered(notFound.page(strconv.Quote(path)))
}
