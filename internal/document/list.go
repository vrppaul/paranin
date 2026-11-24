package document

import (
	"slices"
	"time"
)

// Posts are the posts a group of the blog's list holds, the group a section of the list under its name (D25): what
// a series' group is about, none for a year's or the posts in no series; how many it holds, in the site's words; and
// each post, in the list's order.
type Posts struct {
	Line  []Span
	Count []Span
	List  []ListedPost
}

// ListedPost is a post as the blog's list names it: whether it is out or still to come; its title, and its lead, none
// for a part to come; where it stands in its series, the series' title where its group does not name it, and the
// series' name; and its tags.
type ListedPost struct {
	Readiness Readiness
	Title     []Span
	Lead      []Span
	// Part is its place in its series, in the site's words: "part 6"; none outside a series
	Part []Span
	// Series is its series' title where the group it stands in is no series' own; none in its series' group or
	// outside a series
	Series []Span
	// SeriesGroup is its series' group on the list by series (DESIGN.md, Tile); the zero anchor outside a series
	SeriesGroup Anchor
	Tags        []string
}

// Group is a group of the blog's list: its title, what a place reaches it by, a series' name and none for another
// group, and the posts it holds.
type Group struct {
	Title []Span
	Name  string
	Posts Posts
}

// GroupOf is a block of the blog's list as a group, where it is one: a section holding the posts it names alone.
func GroupOf(block Block) (Group, bool) {
	section, isSection := block.(Section)
	if !isSection || len(section.Blocks) != 1 {
		return Group{Title: nil, Name: "", Posts: Posts{Line: nil, Count: nil, List: nil}}, false
	}
	posts, isPosts := section.Blocks[0].(Posts)
	return Group{Title: section.Title, Name: section.Name, Posts: posts}, isPosts
}

// Filter is a way a head offers to filter the blog's list (DESIGN.md, Tag and BlogHead), in the order a face shows
// them: what it filters by, a tag or a series; the name it goes by, a tag or a series' name; its words, a tag as it
// is, a series in the site's words; where it leads, the list it filters, or the whole list where it is the filter
// shown, none where it leads nowhere, as a tag of a post the lists name nowhere; whether it is the filter the list
// shown is filtered by, which a face shows reversed; and then what names it to a screen reader, in the site's words,
// as the way to the whole list, empty elsewhere.
type Filter struct {
	Kind    FilterKind
	Name    string
	Words   []Span
	Page    Page
	Current bool
	WayOut  string
}

// FilterKind is what a filter of the blog's list filters by.
type FilterKind int

const (
	// FilterTag keeps the posts that carry a tag, which a face writes after its hash
	FilterTag FilterKind = iota
	// FilterSeries keeps the parts of a series
	FilterSeries
)

// ListedWords are the words a head lists under its title, as a face that writes no filter as a link writes them: what a
// project's page says the project is made of, or the tags among the filters a post's or the list's head offers.
func (head Head) ListedWords() []string {
	names := slices.Clone(head.MadeOf)
	for _, filter := range head.Filters {
		if filter.MarkedAsTag() {
			names = append(names, filter.Name)
		}
	}
	return names
}

// MarkedAsTag is whether a face writes the filter's words after a tag's mark, as readers know a tag that gathers posts
// (DESIGN.md, Content); a series' words name what it is. exhaustive holds the switch to every kind.
func (filter Filter) MarkedAsTag() bool {
	switch filter.Kind {
	case FilterTag:
		return true
	case FilterSeries:
	}
	return false
}

// InItsSeriesGroup is whether a post stands in its own series' group, which names the series, so its part alone says
// where it stands, set into its tile's top edge.
func (post ListedPost) InItsSeriesGroup() bool {
	return len(post.Part) > 0 && len(post.Series) == 0
}

// Readiness is whether a post the list names is out, to read at its page, or a part of a series still to come.
//
//sumtype:decl
type Readiness interface {
	readiness()
}

// PostOut is a post out: its page, and when it came out, in the zone it was written in.
type PostOut struct {
	Page      Page
	Published time.Time
}

// ComingPart is a part of a series not out yet, which has no page: the site's word saying it comes, where a post out
// stands by its day.
type ComingPart struct {
	Coming []Span
}

func (Posts) block()          {}
func (PostOut) readiness()    {}
func (ComingPart) readiness() {}
