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

// Head is what a face shows first: the page the document is, the language it is written in, its title, a subtitle
// and the place beside it, what a project is made of, a lead and where a project's code lives, which every face
// shows; the page's description; and, for a face that has a use for it, a greeting before the title. The ways to
// reach the author are facts of their own kind, which a face places where it chooses.
type Head struct {
	Page Page
	// Language is a tag of BCP 47, as HTML's lang and Typst's text take it: "en"
	Language string
	// Greeting is what a face may say to the reader before the author's name, in the site's words; none on a page
	// that greets nobody
	Greeting string
	Title    []Span
	Subtitle []Span
	Place    string
	// Tags are what a project's page says the project is made of, by name: the language it is written in, then what
	// it is built on; none on another page
	Tags []string
	Lead []Span
	// Repository is where the code of a project's page lives; none on another page or for a project that names none
	Repository *Repository
	// Description is what the page is about, for its metadata: a search engine's snippet, a link's preview where
	// it is shared. A face writes it as plain text; none where no browser reads the page (the index, the hidden
	// pages) or the content gives none
	Description []Span
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
