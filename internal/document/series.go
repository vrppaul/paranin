package document

import "time"

// PostPlace is where a post stands in the blog, which only a post has: the series it is a part of, none outside a
// series; the posts its foot leads to; and where it is discussed, each a link by its site's name, standing in for
// comments (D13).
type PostPlace struct {
	Series      *Series
	Neighbours  Neighbours
	Discussions []Link
}

// Series is a run of posts read in order (D25), as a post that is a part of it shows it: its title, and every part,
// those out and those to come, in their order.
type Series struct {
	Title []Span
	Parts []Part
}

// Part is a part of a series: its number in it, counted from one, its title, where it stands for the reader of the
// post that shows it, and its page, none for a part still to come.
type Part struct {
	Number int
	Title  []Span
	State  PartState
	Page   Page
}

// PartState is where a part of a series stands for the reader of one of its parts.
type PartState int

const (
	// The zero state names none
	_ PartState = iota
	// PartReading is the part the reader reads
	PartReading
	// PartNext is the part out after it, the one to read next
	PartNext
	// PartOut is any other part out
	PartOut
	// PartToCome is a part not out yet, which has no page
	PartToCome
)

// Neighbours are the posts a post's foot leads to: in a series the parts before and after it, outside one the posts
// the blog lists before and after it by date; none on a side with nothing there.
type Neighbours struct {
	Before *Neighbour
	After  *Neighbour
}

// Neighbour is a post beside the one read: its number as a part of its series, 0 outside one; its title and its
// lead, none for a part to come; its page, none for a part to come; and when it came out, where the blog orders the
// neighbours by it, outside a series, the zero time inside one.
type Neighbour struct {
	Part      int
	Title     []Span
	Lead      []Span
	Page      Page
	Published time.Time
}
