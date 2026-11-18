package html

import (
	"strconv"

	"github.com/vrppaul/paranin/internal/assets"
	"github.com/vrppaul/paranin/internal/compose"
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
	// included is what the page includes to run the site's script, as every page does, so it is seen in the theme
	// the reader chose
	included assets.Included
	// frame is what every page lays round its main part, the top bar, whose site's name and start the page's own
	// link to where the site begins reads
	frame frame
}

// NewNotFound is the page of an address of no page of a site, in a language, titled by the status the host
// answers with, drawn in the build's stylesheet and running its script, in the frame of every page.
func NewNotFound(site routes.Site, status, language string, build *assets.Build, composed compose.Frame) NotFound {
	return NotFound{
		status:   status,
		language: language,
		included: build.Includes(assets.Site),
		frame:    frameOf(site, composed),
	}
}

// For is the page for a path asked for: its status, the path quoted as Go quotes a string, so no control character
// in it reaches the page, then a link to where the site begins.
func (notFound NotFound) For(path string) string {
	return rendered(notFound.page(strconv.Quote(path)))
}
