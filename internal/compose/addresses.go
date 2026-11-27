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

	"github.com/yuin/goldmark/ast"

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
// a section of a page that has none, by the anchor and the path; or to none of its sections and widgets, naming the
// anchors it has
const (
	laterProblem      = "%q leads to a post that comes out only at %s, after the post linking it"
	noPageProblem     = "%q leads to no page the site serves: a link inside it is a page's path from the root, as /blog/<slug>"
	noSectionsProblem = "%q names a section of %q, which has none: link to the page alone"
	noSectionProblem  = "%q names no section or widget of %q, whose anchors are %s"
	anchorJoint       = ", "
)

// Check holds a snapshot to what only its whole site knows, in every language, at each moment from now on it changes:
// no two pages at one address, and no link inside it to nowhere, a draft's as its author reads the site once every
// post that comes later is out.
func Check(snapshot *content.Snapshot, now time.Time) error {
	public := Audience{At: now, Drafts: false}
	posts := allPosts(snapshot)
	changes, later := public.changes(posts), laterPosts(posts, public)
	faults := content.Faults{}
	prepared := Prepare(snapshot)
	for _, moment := range changes {
		site := SiteOf(prepared, Audience{At: moment, Drafts: false})
		reached := destinationsOf(site, later)
		for _, blog := range site.Blogs() {
			faults = append(faults, reached.linksLeadingNowhere(blog.postsServed)...)
		}
	}
	// the author's site once every post that comes later is out holds the most pages
	authorsSite := SiteOf(prepared, Audience{At: changes[len(changes)-1], Drafts: true})
	heads, authorReaches := headsOf(authorsSite.served), destinationsOf(authorsSite, later)
	for _, blog := range authorsSite.Blogs() {
		drafts := slices.DeleteFunc(slices.Clone(blog.postsServed), func(post content.Post) bool { return post.Status != content.StatusDraft })
		faults = append(faults, authorReaches.linksLeadingNowhere(drafts)...)
	}
	var linkFaults error
	if len(faults) > 0 {
		slices.SortFunc(faults, byPlace)
		linkFaults = slices.Compact(faults)
	}
	return errors.Join(sharedAddresses(heads), linkFaults)
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
		title := document.PlainText(head.Title)
		for _, path := range addressesOf(head.Page) {
			titlesAt[path] = append(titlesAt[path], title)
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

// addressesOf are the paths a page is served at, one for each format it is served in.
func addressesOf(page document.Page) []string {
	paths := []string{}
	for _, format := range routes.Formats() {
		if path := routes.Path(page, format); path != "" {
			paths = append(paths, path)
		}
	}
	return paths
}

// destinations are where a site's links inside it may lead: the pages it serves and its feeds by their paths, the
// anchors of each page's sections, and the posts that come later by their paths, each by the moment it comes out.
type destinations struct {
	paths   map[string]document.Target
	anchors map[document.Page][]string
	later   map[string]content.Moment
}

// destinationsOf are where a site's links may lead, in every language, beside the posts that come later.
func destinationsOf(site Site, later map[string]content.Moment) destinations {
	return destinations{paths: site.paths, anchors: anchorsOf(site.served), later: later}
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

// linksLeadingNowhere are the faults of posts' links inside the site, and their revisions', in no order.
func (reached destinations) linksLeadingNowhere(posts []content.Post) content.Faults {
	faults := content.Faults{}
	for _, post := range posts {
		for _, version := range append([]content.Revision{post.Current()}, post.Revisions...) {
			faults = append(faults, reached.faultsOf(version.SiteLinks)...)
		}
	}
	return faults
}

// faultsOf are the faults of links inside the site, those that lead nowhere, in no order.
func (reached destinations) faultsOf(links map[*ast.Link]content.WrittenLink) content.Faults {
	faults := content.Faults{}
	for _, written := range links {
		if problem, nowhere := reached.leadsNowhere(written.Link); nowhere {
			faults = append(faults, content.Fault{File: written.File, Field: "", Line: written.Line, Problem: problem})
		}
	}
	return faults
}

// leadsNowhere is what is wrong with a link inside the site, where it leads to a post not out yet, to no page served
// or to no section of one.
func (reached destinations) leadsNowhere(link content.SiteLink) (string, bool) {
	target, found := reached.paths[link.Path]
	// a feed has no sections, and no page's are taken
	page, _ := target.(document.Page)
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

// anchorsOf are the anchors of each document's sections and widgets, in reading order, by its page.
func anchorsOf(documents []document.Document) map[document.Page][]string {
	anchors := make(map[document.Page][]string, len(documents))
	for _, composed := range documents {
		anchors[composed.Head.Page] = anchorsAmong(composed.Blocks)
	}
	return anchors
}

// anchorsAmong are the anchors among blocks a link may name, in reading order: each section's that has one, followed
// by those inside it, and each widget's.
func anchorsAmong(blocks []document.Block) []string {
	anchors := []string{}
	for _, block := range blocks {
		switch typed := block.(type) {
		case document.Section:
			if typed.Name != "" {
				anchors = append(anchors, typed.Name)
			}
			anchors = append(anchors, anchorsAmong(typed.Blocks)...)
		case document.Widget:
			anchors = append(anchors, typed.Live.Name)
		case document.Reference:
			anchors = append(anchors, anchorsAmong(typed.Blocks())...)
		case document.Paragraph, document.Table, document.CodeBlock, document.List, document.Quote, document.Pause, document.Figure,
			document.Facts, document.ExperienceHead, document.Entry, document.Points, document.Side, document.Clip, document.ProjectList,
			document.Contents,
			document.Checks, document.Skills, document.Posts:
			// a widget stands as a block of the post's own, never inside a quote or a list's item, and no other block
			// takes an anchor
		}
	}
	return anchors
}
