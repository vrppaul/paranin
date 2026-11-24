// Package document is the model every face draws (docs/ARCHITECTURE.md, ADR-0003): a document is a head and its
// blocks, and text inside a block is a tree of spans. Blocks and spans are closed sets, so a face that switches
// over every kind, as gochecksumtype holds it to, draws every kind there is. The model knows no content and no
// face: compose fills it, and each face decides how each kind looks. What a face reads off a document by the model's
// own terms is read here once: its plain text, its phrases, the brief a résumé opens with (brief.go), the document as
// a reader of the finished work sees it, with nothing that is only a gap still to fill (written.go), and a link to a
// mailbox, its mailbox and its host (target.go).
package document

// Document is one page as every face draws it.
type Document struct {
	Head   Head
	Blocks []Block
}

// Head is what a face shows first: the page the document is, a post's number, the language it is written in, its
// title, a subtitle and the place beside it, what a project is made of or a post's tags, a lead, the day a post came
// out and where a project's code lives, which every face shows; where a post stands in the blog; the page's
// description and the feeds that follow it; and, for a face that has a use for it, a greeting before the title. The
// ways to reach the author are facts of their own kind, which a face places where it chooses.
type Head struct {
	Page Page
	// Number is a post's number, which blog/<number> reaches as its slug does (D26); 0 on any other page
	Number int
	// Language is a tag of BCP 47, as HTML's lang and Typst's text take it: "en"
	Language string
	// Greeting is what a face may say to the reader before the author's name, in the site's words; none on a page
	// that greets nobody
	Greeting string
	Title    []Span
	// Subtitle is what stands beside the title: the author's title on the résumé, a project's tagline on its page,
	// the filter in the site's words on a filtered list
	Subtitle []Span
	Place    string
	// Filters are the ways a post's head and the blog's list's offer to filter the list: on a list filtered by a series
	// that series, then each tag; none on any other page
	Filters []Filter
	// MadeOf is what a project's page says the project is made of, by name: the language it is written in, then what
	// it is built on; none on another page, a post's and the list's tags standing among their Filters
	MadeOf []string
	Lead   []Span
	// Version is the version of a post its reader reads; none on any other page
	Version Version
	// Repository is where the code of a project's page lives; none on another page or for a project that names none
	Repository *Repository
	// Post is where a post stands in the blog and where it is discussed; none on any other page
	Post *PostPlace
	// Description is what the page is about, for its metadata: a search engine's snippet, a link's preview where
	// it is shared. A face writes it as plain text; none where no browser reads the page (the index, the hidden
	// pages) or the content gives none
	Description []Span
	// Feeds are the feeds that follow a page of the blog, each a link to a Feed by its title: a series' first, on a
	// page that shows the series, then the blog's, which the résumé names too; none on any other page, or where no
	// feed holds a post
	Feeds []Link
}

// Fact is a label and its value: a fact of the résumé, a way to reach its author, a detail of a project.
type Fact struct {
	Label string
	Value []Span
}

// Command is the command a fact's value is, where it is code alone: a command a reader types or copies whole, as a
// project's install line or the ssh way to reach the author; none where the value holds more than the code.
func (fact Fact) Command() (string, bool) {
	if len(fact.Value) != 1 {
		return "", false
	}
	code, isCode := fact.Value[0].(Code)
	return code.Text, isCode
}

// Repository is where a project's code lives: the site's word that names it, and its address.
type Repository struct {
	Label   string
	Address Address
}

// Facts are the repository as a face lists it among facts: its label, and its address as a link reading as it is
// written; none where there is no repository.
func (repository *Repository) Facts() []Fact {
	if repository == nil {
		return nil
	}
	return []Fact{{Label: repository.Label, Value: []Span{Link{Target: repository.Address, Spans: []Span{Text{Text: repository.Address.URL}}}}}}
}
