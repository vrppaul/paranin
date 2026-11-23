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

// Table is rows of cells, each column one kind of value, under a header that names each column where the table has
// one: a post's table may, a skill group's has none. A row's first cell names the row.
type Table struct {
	// Header is the row naming the columns; a row of no cells where the table has none
	Header Row
	Rows   []Row
}

// Row is one row of a table.
type Row struct {
	Cells []Cell
}

// Cell is one cell of a row, of a kind a face draws it by. The kind rides on each cell, not on its column, so a
// face reads a cell's kind off the cell itself; the first cell's look is the Table's own rule.
type Cell struct {
	Kind  CellKind
	Spans []Span
}

// CellKind is what a cell of a table holds.
type CellKind int

const (
	// The zero kind names none
	_ CellKind = iota
	// CellText is words, read as the text around them
	CellText
	// CellDuration is how long something lasted, as a skill's use, which a face sets back from the words beside it
	CellDuration
)

// CodeBlock is text shown as it is, line for line: a terminal's recording, a post's code.
type CodeBlock struct {
	Text string
	// Language is what the code is written in, as its author names it: "go"; none where the author names none
	Language string
}

// List is items one after another, numbered from its start where it is in order: a post's list.
type List struct {
	Ordered bool
	// Start is the first item's number in a list in order; 0 in one that is not
	Start int
	Items []Item
}

// Item is one item of a list, holding blocks as a section does: a paragraph, or more, a list nested in it among them.
type Item struct {
	Blocks []Block
}

// Quote is blocks a post quotes from someone else.
type Quote struct {
	Blocks []Block
}

// Pause is a turn a post takes inside a part, which no heading names: a thematic break.
type Pause struct{}

// Facts is a list of labels and their values, of a kind a face finds them by, not by their place.
type Facts struct {
	Kind FactsKind
	List []Fact
}

// FactsKind is what a list of facts tells.
type FactsKind int

const (
	// The zero kind names none
	_ FactsKind = iota
	// FactsDetails are details of what the document is about: the author's facts, the languages they speak, a
	// project's details
	FactsDetails
	// FactsWays are the ways to reach the author, the mail first, which a face shows where it chooses
	FactsWays
)

// Entry is a span of time spent somewhere, a role or a school as its kind says, with what came of it.
type Entry struct {
	Kind     EntryKind
	Period   Period
	Title    []Span
	Subtitle []Span
	Place    string
	Points   []Point
}

// EntryKind is what an entry's time was spent on.
type EntryKind int

const (
	// The zero kind names none
	_ EntryKind = iota
	// EntryRole is a role: work for an employer or a client
	EntryRole
	// EntrySchool is a school: a course of study
	EntrySchool
)

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

// Checks are a group of checks under the group's name: the hidden health report holds its checks so, as Neovim's
// :checkhealth does.
type Checks struct {
	Title []Span
	List  []Check
}

// Check is one line of a report: how it reads, what it says, and the advice on it, empty where it has none.
type Check struct {
	Level  CheckLevel
	Spans  []Span
	Advice []Span
}

// CheckLevel is how a check reads. The levels rise from a check that passed to one that failed, and a face that
// lists them lists them in that order: a level added takes its place in it.
type CheckLevel int

const (
	// The zero level names none
	_ CheckLevel = iota
	// CheckOK is a check that passed
	CheckOK
	// CheckWarning is a check that passed with something to look at
	CheckWarning
	// CheckError is a check that failed
	CheckError
)

// Skills are every skill of the author in one list, as lazy.nvim lists its plugins: the hidden list of skills holds
// them so.
type Skills struct {
	List []Skill
}

// Skill is one skill of the list: its name, how long, where it was proven, and whether the author is still learning
// it.
type Skill struct {
	Name     []Span
	Duration []Span
	// Months is how long the author used it, where the profile writes a span; 0 where it writes none: for a skill
	// used on the side, one still being learnt, or a gap still to fill
	Months   int
	Sources  []Span
	Learning bool
}

func (Section) block()   {}
func (Paragraph) block() {}
func (Table) block()     {}
func (CodeBlock) block() {}
func (List) block()      {}
func (Quote) block()     {}
func (Pause) block()     {}
func (Facts) block()     {}
func (Entry) block()     {}
func (Card) block()      {}
func (Contents) block()  {}
func (Checks) block()    {}
func (Skills) block()    {}
