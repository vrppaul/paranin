package document

import "time"

// Posts are the posts a group of the blog's list holds, the group a section of the list under its name (D25): what
// a series' group is about, none for a year's or the posts in no series; how many it holds, in the site's words; and
// each post, in the list's order.
type Posts struct {
	Line  []Span
	Count []Span
	List  []ListedPost
}

// ListedPost is a post as the blog's list names it: whether it is out or still to come; its title, and its lead, none
// for a part to come; where it stands in its series, and the series' title where its group does not name it; and its
// tags.
type ListedPost struct {
	Readiness Readiness
	Title     []Span
	Lead      []Span
	// Part is its place in its series, in the site's words: "part 6"; none outside a series
	Part []Span
	// Series is its series' title where the group it stands in is no series' own; none in its series' group or
	// outside a series
	Series []Span
	Tags   []string
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
