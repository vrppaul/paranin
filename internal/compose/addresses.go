package compose

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// How a clash of addresses is written: the address, then the title of each page that takes it; one clash a line
const (
	clashFormat = "%s is the address of more than one page: %s"
	titleJoint  = " and "
	clashJoint  = "\n"
)

// AddressClash is an address more than one page of the site takes, in any format, as routes writes it, and the title
// of each of those pages.
type AddressClash struct {
	Path   string
	Titles []string
}

// AddressClashes are the addresses more than one page of a snapshot takes, in their order: content a release refuses
// as it refuses any fault, since the host would serve one page and hide the other.
type AddressClashes []AddressClash

func (clashes AddressClashes) Error() string {
	lines := make([]string, 0, len(clashes))
	for _, clash := range clashes {
		titles := make([]string, 0, len(clash.Titles))
		for _, title := range clash.Titles {
			titles = append(titles, strconv.Quote(title))
		}
		lines = append(lines, fmt.Sprintf(clashFormat, clash.Path, strings.Join(titles, titleJoint)))
	}
	return strings.Join(lines, clashJoint)
}

// SharedAddresses are the addresses more than one page of a snapshot takes, as AddressClashes; none where every page
// has addresses of its own. It reads the pages compose makes through the addresses routes writes, so a tab of the
// résumé titled as the blog, or a post slugged as its list by series, is found as any other clash.
func SharedAddresses(snapshot *content.Snapshot) error {
	titlesAt := map[string][]string{}
	for _, served := range servedPages(snapshot) {
		for _, format := range servedFormats() {
			if path := routes.Path(served.Head.Page, format); path != "" {
				titlesAt[path] = append(titlesAt[path], document.PlainText(served.Head.Title))
			}
		}
	}
	clashes := AddressClashes{}
	for _, path := range slices.Sorted(maps.Keys(titlesAt)) {
		if titles := titlesAt[path]; len(titles) > 1 {
			clashes = append(clashes, AddressClash{Path: path, Titles: titles})
		}
	}
	if len(clashes) == 0 {
		return nil
	}
	return clashes
}

// servedPages are every page a snapshot's site serves: the index, the résumé and each project's page with each of their
// tabs alone, each post, and the blog's list in each grouping.
func servedPages(snapshot *content.Snapshot) []document.Document {
	pages := []document.Document{Index(snapshot)}
	for _, page := range Documents(snapshot) {
		pages = append(append(pages, page), TabDocuments(page)...)
	}
	return slices.Concat(pages, Posts(snapshot), Lists(snapshot))
}

// servedFormats are the formats a page is served in, each at an address of its own.
func servedFormats() []routes.Format {
	return []routes.Format{routes.FormatChosen, routes.FormatMarkdown, routes.FormatManual}
}
