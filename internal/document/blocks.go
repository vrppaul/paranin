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

func (Section) block()   {}
func (Paragraph) block() {}
func (Table) block()     {}
func (Facts) block()     {}
func (Entry) block()     {}
func (Card) block()      {}
