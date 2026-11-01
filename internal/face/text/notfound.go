package text

import (
	"strconv"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// The program a terminal reads the site with
const readingCommand = "curl"

// NotFound is the text of an address the site has no page at, for a terminal, drawn once for a site: each path
// asked for is all that differs.
type NotFound struct {
	// status is the status the host answers with, as HTTP writes it: 404 Not Found
	status string
	// startCommand is the command that reads the site where it begins
	startCommand string
}

// NewNotFound is the text of an address of no page of a site, opening with the status the host answers with.
func NewNotFound(site routes.Site, status string) NotFound {
	start := document.Page{Kind: document.PageResume, Slug: ""}
	return NotFound{status: status, startCommand: readingCommand + space + site.Address(start, routes.FormatChosen)}
}

// For is the text for a path asked for: the status, the path quoted as Go quotes a string, so no control
// character in it reaches the reader's terminal, then where the site begins.
func (notFound NotFound) For(path string) string {
	return notFound.status + labelEnd + strconv.Quote(path) + blankLine + notFound.startCommand + lineEnd
}
