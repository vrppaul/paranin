package compose

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

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

// What is said of a post's link inside the site that leads nowhere: to a post that comes out after the post linking
// it, by its path and the moment it comes out, as front matter writes it; to no page the site serves, by its path; to
// a section of a page that has none, by the anchor and the path; or to none of its sections, naming those it has
const (
	laterProblem      = "%q leads to a post that comes out only at %s, after the post linking it"
	noPageProblem     = "%q leads to no page the site serves: a link inside it is a page's path from the root, as /blog/<slug>"
	noSectionsProblem = "%q names a section of %q, which has none: link to the page alone"
	noSectionProblem  = "%q names no section of %q, whose sections are %s"
	anchorJoint       = ", "
)

// Check holds a snapshot to what only its whole site knows, at each moment from now on it changes: no two pages at one
// address, a preview's among them, and no link inside it to nowhere, a draft's as its author reads the site, and a
// preview's as its link's reader does once every post that comes later is out, so a preview may link to a post dated
// later than now.
func Check(snapshot *content.Snapshot, now time.Time) error {
	public := Audience{At: now, Drafts: false}
	changes, later := public.changes(snapshot.Posts), laterPosts(snapshot.Posts, public)
	faults := content.Faults{}
	prepared := Prepare(snapshot)
	// the public's site once every post that comes later is out, the last the loop composes
	var latest Site
	for _, moment := range changes {
		latest = SiteOf(prepared, Audience{At: moment, Drafts: false})
		faults = append(faults, destinationsOf(latest, later).linksLeadingNowhere(latest.postsServed)...)
	}
	for _, previewed := range latest.Previews {
		faults = append(faults, previewDestinations(latest, previewed, later).linksLeadingNowhere([]content.Post{previewed.written})...)
	}
	// the author's site once every post that comes later is out holds the most pages
	authorsSite := SiteOf(prepared, Audience{At: changes[len(changes)-1], Drafts: true})
	drafts := slices.DeleteFunc(slices.Clone(authorsSite.postsServed), func(post content.Post) bool { return post.Status != content.StatusDraft })
	faults = append(faults, destinationsOf(authorsSite, later).linksLeadingNowhere(drafts)...)
	var linkFaults error
	if len(faults) > 0 {
		slices.SortFunc(faults, byPlace)
		linkFaults = slices.Compact(faults)
	}
	return errors.Join(sharedAddresses(append(headsOf(authorsSite.served), previewHeads(authorsSite.Previews)...)), linkFaults)
}

// previewHeads are the heads of the posts previewed, which take their addresses once out.
func previewHeads(previews []Preview) []document.Head {
	heads := make([]document.Head, 0, len(previews))
	for _, previewed := range previews {
		heads = append(heads, previewed.Post.Head)
	}
	return heads
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
		for _, format := range routes.Formats() {
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

// destinations are where a site's links inside it may lead: the pages it serves by their paths, the anchors of each
// page's sections, and the posts that come later by their paths, each by the moment it comes out.
type destinations struct {
	pages   map[string]document.Page
	anchors map[document.Page][]string
	later   map[string]content.Moment
}

// destinationsOf are where a site's links may lead, beside the posts that come later.
func destinationsOf(site Site, later map[string]content.Moment) destinations {
	return destinations{pages: site.paths, anchors: anchorsOf(site.served), later: later}
}

// previewDestinations are where a preview's links may lead, beside the posts that come later: where the site's do, and
// to its own page and its sections.
func previewDestinations(site Site, previewed Preview, later map[string]content.Moment) destinations {
	return destinations{
		pages:   withOwnPage(site.paths, previewed.Post.Head.Page),
		anchors: anchorsOf(append(slices.Clone(site.served), previewed.Post)),
		later:   later,
	}
}

// laterPosts are the posts that come later than an audience's moment, by the path each is served at once out, each by
// the moment it comes out.
func laterPosts(posts []content.Post, audience Audience) map[string]content.Moment {
	later := map[string]content.Moment{}
	for _, post := range posts {
		if audience.comesLater(post) {
			later[routes.Path(postPage(post), routes.FormatChosen)] = post.Published
		}
	}
	return later
}

// linksLeadingNowhere are the faults of posts' links inside the site, in no order.
func (reached destinations) linksLeadingNowhere(posts []content.Post) content.Faults {
	faults := content.Faults{}
	for _, post := range posts {
		for _, written := range post.SiteLinks {
			if problem, nowhere := reached.leadsNowhere(written.Link); nowhere {
				faults = append(faults, content.Fault{File: written.File, Field: "", Line: written.Line, Problem: problem})
			}
		}
	}
	return faults
}

// leadsNowhere is what is wrong with a link inside the site, where it leads to a post not out yet, to no page served
// or to no section of one.
func (reached destinations) leadsNowhere(link content.SiteLink) (string, bool) {
	page, found := reached.pages[link.Path]
	taken := reached.anchors[page]
	comes, isLater := reached.later[link.Path]
	switch {
	case !found && isLater:
		return fmt.Sprintf(laterProblem, link.Path, comes.Written()), true
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
