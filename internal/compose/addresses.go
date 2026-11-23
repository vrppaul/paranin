package compose

import (
	"cmp"
	"errors"
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

// What is said of a post's link inside the site that leads nowhere: to no page the site serves, by its path; to a
// section of a page that has none, by the anchor and the path; or to none of its sections, naming those it has
const (
	noPageProblem     = "%q leads to no page the site serves: a link inside it is a page's path from the root, as /blog/<slug>"
	noSectionsProblem = "%q names a section of %q, which has none: link to the page alone"
	noSectionProblem  = "%q names no section of %q, whose sections are %s"
	anchorJoint       = ", "
)

// Check holds a snapshot to what only its whole site knows: no two pages at one address, no link inside it to nowhere.
func Check(snapshot *content.Snapshot) error {
	site := SiteOf(snapshot)
	return errors.Join(sharedAddresses(headsOf(site.served)), linksLeadingNowhere(servedPosts(snapshot), site.paths, anchorsOf(site.served)))
}

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

// sharedAddresses are the addresses more than one of the served pages' heads takes, as AddressClashes; none where
// every page has addresses of its own. It reads them through the addresses routes writes, so a tab of the résumé
// titled as the blog, or a post slugged as its list by series, is found as any other clash.
func sharedAddresses(heads []document.Head) error {
	titlesAt := map[string][]string{}
	for _, head := range heads {
		for _, format := range servedFormats() {
			if path := routes.Path(head.Page, format); path != "" {
				titlesAt[path] = append(titlesAt[path], document.PlainText(head.Title))
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

// linksLeadingNowhere are the faults of the served posts' links inside the site, in the order of their files' lines.
func linksLeadingNowhere(posts []content.Post, pages map[string]document.Page, anchors map[document.Page][]string) error {
	faults := content.Faults{}
	for _, post := range posts {
		for _, written := range post.SiteLinks {
			if problem, nowhere := leadsNowhere(written.Link, pages, anchors); nowhere {
				faults = append(faults, content.Fault{File: written.File, Field: "", Line: written.Line, Problem: problem})
			}
		}
	}
	if len(faults) == 0 {
		return nil
	}
	slices.SortFunc(faults, byPlace)
	return faults
}

// leadsNowhere is what is wrong with a link inside the site, where it leads to no page served or no section of one.
func leadsNowhere(link content.SiteLink, pages map[string]document.Page, anchors map[document.Page][]string) (string, bool) {
	page, found := pages[link.Path]
	taken := anchors[page]
	switch {
	case !found:
		return fmt.Sprintf(noPageProblem, link.Path), true
	case link.Anchor == "" || slices.Contains(taken, link.Anchor):
		return "", false
	case len(taken) == 0:
		return fmt.Sprintf(noSectionsProblem, link.Anchor, link.Path), true
	}
	return fmt.Sprintf(noSectionProblem, link.Anchor, link.Path, strings.Join(taken, anchorJoint)), true
}

// byPlace orders faults by their files, then their lines, then their words, as a reader goes through them.
func byPlace(first, second content.Fault) int {
	return cmp.Or(cmp.Compare(first.File, second.File), cmp.Compare(first.Line, second.Line), cmp.Compare(first.Problem, second.Problem))
}

// anchorsOf are the anchors of each document's sections, in reading order, by its page.
func anchorsOf(documents []document.Document) map[document.Page][]string {
	anchors := make(map[document.Page][]string, len(documents))
	for _, composed := range documents {
		anchors[composed.Head.Page] = sectionAnchors(composed.Blocks)
	}
	return anchors
}

// sectionAnchors are the anchors of the sections among blocks that have one, each followed by those inside it.
func sectionAnchors(blocks []document.Block) []string {
	anchors := []string{}
	for _, block := range blocks {
		if section, isSection := block.(document.Section); isSection {
			if section.Name != "" {
				anchors = append(anchors, section.Name)
			}
			anchors = append(anchors, sectionAnchors(section.Blocks)...)
		}
	}
	return anchors
}

// servedFormats are the formats a page is served in, each at an address of its own.
func servedFormats() []routes.Format {
	return []routes.Format{routes.FormatChosen, routes.FormatMarkdown, routes.FormatManual}
}
