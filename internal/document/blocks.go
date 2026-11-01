package document

// Block is one kind of block of a document.
//
//sumtype:decl
type Block interface {
	block()
}

// Section is a titled part of a document.
type Section struct {
	Title  []Span
	Blocks []Block
}

// Paragraph is a run of text.
type Paragraph struct {
	Spans []Span
}

// Table is rows of cells, each column one kind of value.
type Table struct {
	Rows []Row
}

// Row is one row of a table.
type Row struct {
	Cells []Cell
}

// Cell is one cell of a row.
type Cell struct {
	Spans []Span
}

// CodeBlock is text shown as it is, line for line: a terminal's recording.
type CodeBlock struct {
	Text string
}

// Facts is a list of labels and their values.
type Facts struct {
	List []Fact
}

// Entry is a span of time spent somewhere: a role or a school, with what came of it.
type Entry struct {
	Period   Period
	Title    []Span
	Subtitle []Span
	Place    string
	Points   []Point
}

// Point is one thing an entry holds up.
type Point struct {
	Spans []Span
}

// Card is a piece of work: its title, what it is in a line, what it is written in, what it does, its details,
// and its page, PageNone where it has none.
type Card struct {
	Title    []Span
	Subtitle []Span
	Tag      string
	Summary  []Span
	Facts    []Fact
	Page     Page
}

// Contents are pages of the site a document leads to, each by its title and a line on what it is: the site's
// index lists its pages so.
type Contents struct {
	Listings []Listing
}

// Listing is one page a list of contents leads to: the page, its title, and a line on what it is.
type Listing struct {
	Page  Page
	Title []Span
	Note  []Span
}

func (Section) block()   {}
func (Paragraph) block() {}
func (Table) block()     {}
func (CodeBlock) block() {}
func (Facts) block()     {}
func (Entry) block()     {}
func (Card) block()      {}
func (Contents) block()  {}
