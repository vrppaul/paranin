// Package document is the model every face draws (docs/ARCHITECTURE.md, ADR-0003): a document is a head and its
// blocks, and text inside a block is a tree of spans. Blocks and spans are closed sets, so a face that switches
// over every kind, as gochecksumtype holds it to, draws every kind there is. The model knows no content and no
// face: compose fills it, and each face decides how each kind looks.
package document

// Document is one page as every face draws it.
type Document struct {
	Head   Head
	Blocks []Block
}

// Head is what every face shows first: the page the document is, its title, a subtitle and the place beside it,
// a lead, and the ways to reach the author.
type Head struct {
	Page     Page
	Title    []Span
	Subtitle []Span
	Place    string
	Lead     []Span
	Ways     []Fact
}

// Fact is a label and its value: a fact of the résumé, a way to reach its author, a detail of a project.
type Fact struct {
	Label string
	Value []Span
}
