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
// title, a subtitle and the place beside it, a post's tags, a lead, the day a post came out, and what gets a project
// and tells about it, which every face shows; where a post stands in the blog; the page's description and the feeds
// that follow it; and, for a face that has a use for it, a greeting before the title. The ways to reach the author
// are facts of their own kind, which a face places where it chooses.
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
	// Filters are the tags a post's head and the blog's list's offer to filter the list by; none on any other page
	Filters []Filter
	Lead    []Span
	// Version is the version of a post its reader reads; none on any other page
	Version Version
	// Project is what gets a project and what tells about it; none on any page but a project's
	Project *Project
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
	// Languages are the languages a post's Switch offers, the one read and each the lists name it in; none for one alone
	Languages []Language
}

// Previewed is whether a link's preview shows the page, its address and its card (DESIGN.md, Social cards): every page
// with an address of its own but a draft, which only its author reads.
func (head Head) Previewed() bool {
	_, draft := head.Version.(VersionDraft)
	return head.Page.OwnAddress() && !draft
}

// CardPage is the page whose card a page of a head shows (DESIGN.md, Social cards): its own, a tab its page's; none,
// the page of no kind, on a page no preview shows, an older version, a filtered list, a draft.
func (head Head) CardPage() Page {
	if !head.Previewed() {
		return PageOf(PageNone)
	}
	whole := head.Page
	whole.Tab = ""
	return whole
}

// Indexed is whether a search engine may index the page, which then names the address it is known by (docs/SEO.md,
// What is indexed): every page a preview shows, a post only where the blog's lists name it. Not indexed: the index and
// the hidden pages, which no browser reads as a page of the site; a filtered list, a view of the list, and an older
// version of a post, which a newer stands over, neither with an address of its own (docs/FOUNDATION.md §5.4); an
// unlisted post, which its author hands to whom they choose; and a draft its author reads. A page not indexed names no
// address it is known by, since a search engine reads noindex beside one as two signs that disagree.
func (head Head) Indexed() bool {
	return head.Previewed() && (head.Post == nil || head.Post.Listed)
}

// Language is a language a post is read in (DESIGN.md, Switch): its tag, which every face names it by, its locale, the
// post's page in it, and whether it is the language read.
type Language struct {
	Tag string
	// Locale is the locale Open Graph marks a page in it in, language_TERRITORY
	Locale string
	Page   Page
	Read   bool
}

// Spans are a language as a face with no Switch writes it: the one read strong, each other a link to the post in it.
func (language Language) Spans() []Span {
	name := []Span{Text{Text: language.Tag}}
	if language.Read {
		return []Span{Strong{Spans: name}}
	}
	return []Span{Link{Target: language.Page, Spans: name}}
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

// Project is what a project's page says of it beside its name and lines (DESIGN.md, ProjectHead): what gets the
// project, where its code lives and how it is installed, and what tells about it.
type Project struct {
	// Repository is where the project's code lives; none where the profile names no place
	Repository *Repository
	// InstallLabel is the site's word for installing, over the first install where a face labels each
	InstallLabel string
	// Installs are the ways to install the project: the first, then each with a note on when to take it instead
	Installs []Install
	// About are the facts that tell about the project, each under the site's word, as far as the profile gives them
	About []Fact
}

// Install is a way to install a project: a note on when to take it, none on the first, and its command.
type Install struct {
	Note    []Span
	Command string
}

// InstallFacts are the installs as a face that labels each lists them among facts (InstallFact).
func (project Project) InstallFacts() []Fact {
	facts := make([]Fact, 0, len(project.Installs))
	for _, install := range project.Installs {
		facts = append(facts, project.InstallFact(install))
	}
	return facts
}

// InstallFact is an install as a face that labels it lists it among facts, its command as code: under its note, or
// under the site's word where it has none, as the first.
func (project Project) InstallFact(install Install) Fact {
	label := project.InstallLabel
	if len(install.Note) > 0 {
		label = PlainText(install.Note)
	}
	return Fact{Label: label, Value: []Span{Code{Text: install.Command}}}
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
