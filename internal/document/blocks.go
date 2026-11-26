package document

import (
	"encoding/json"
	"strconv"
	"strings"
)

// Block is one kind of block of a document.
//
//sumtype:decl
type Block interface {
	block()
}

// Section is a titled part of a document.
type Section struct {
	Title []Span
	// Name is the section's anchor on every face, which a link and the command line reach it by inside its page: a
	// post's heading's, unique on its page, and a series' group's, the series' name; none where nothing reaches it
	Name   string
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

// Cell is one cell of a row, words read as the text around them; the first cell's look is the Table's own rule.
type Cell struct {
	Spans []Span
}

// CodeBlock is text shown as it is, line for line: a post's code. Its runs hold the text whole by kind of word, so a
// face colours a kind and never reads a language.
type CodeBlock struct {
	// Language is what the code is written in, as its author names it: "go"; none where the author names none
	Language string
	Runs     []CodeRun
}

// CodeRun is a run of a code block's text, every word of it one kind.
type CodeRun struct {
	Kind CodeKind
	Text string
}

// CodeKind is the kind of word a run of code holds, as a reader of code tells words apart.
type CodeKind int

const (
	// The zero kind names none
	_ CodeKind = iota
	// CodePlain is a word of no kind below, a name a program gives among them, and the space between words
	CodePlain
	// CodeKeyword is a word of the language itself: func, return
	CodeKeyword
	// CodeString is text written into the code, in its quotes
	CodeString
	// CodeNumber is a number, or a constant the language names: 2, true, nil
	CodeNumber
	// CodeFunction is a function's name, or a key of data, as YAML's, which names what follows it as a function does
	CodeFunction
	// CodeType is the name of a type: int, string
	CodeType
	// CodeBuiltin is a function the language comes with: len
	CodeBuiltin
	// CodeOperator is a sign that works on values: *, :=
	CodeOperator
	// CodeVariable is a variable a language marks as one, as a shell's $HOME
	CodeVariable
	// CodeComment is words for the reader that the program skips
	CodeComment
	// CodePunctuation is a sign that holds the code's parts apart or together: brackets, a comma
	CodePunctuation
)

// Text is the code whole, its runs joined: what a reader copies, and what a face that colours nothing writes.
func (block CodeBlock) Text() string {
	var builder strings.Builder
	for _, run := range block.Runs {
		builder.WriteString(run.Text)
	}
	return builder.String()
}

// List is items one after another, numbered from its start where it is in order: a post's list.
type List struct {
	Ordered bool
	// Start is the first item's number in a list in order; 0 in one that is not
	Start int
	Items []Item
}

// What ends an item's number in a list in order, as a face writes it: 3.
const numberEnd = "."

// Number is an item's number as a face writes it in a list in order, by the item's place, counted from the list's
// start: 3.
func (list List) Number(index int) string {
	return strconv.Itoa(list.Start+index) + numberEnd
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

// Figure is a picture a post shows as a block of its own: its file, served beside the post; its alt text, which stands
// for it wherever it is not seen; its caption, none where it has none; and its size in pixels, by which a page keeps
// its room before it comes.
type Figure struct {
	Picture File
	Alt     string
	Caption string
	Width   int
	Height  int
}

// Widget is an island a post shows as a block of its own: the component it names; its alt text, which stands for it
// wherever it does not run; its caption, none where it has none; its props, which the island hands the component;
// and the live one, the post's page at the widget's anchor, which every face that runs no island leads to.
type Widget struct {
	Name    string
	Alt     string
	Caption string
	// Props are a JSON object of the props the post writes, in their written order, but its alt text and caption: as
	// JSON, since each widget's own types are the page's (docs/FOUNDATION.md §5.4, Layout of a post)
	Props json.RawMessage
	Live  Anchor
}

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
	// FactsDetails are details of what the document is about: the author's facts, the languages they speak
	FactsDetails
	// FactsWays are the ways to reach the author, the mail first, which a face shows where it chooses
	FactsWays
)

// ExperienceHead is what the author's experience opens with (DESIGN.md, ExperienceHead): its line, the one thing a
// reader should keep of it, none where the profile writes none; and the technologies it puts forward.
type ExperienceHead struct {
	Line         []Span
	Technologies []Technology
}

// Technology is a technology the experience puts forward: its name, and how long it was used, as the skills' list
// writes a duration.
type Technology struct {
	Name string
	Used []Span
}

// Blocks are the head as a face draws it until the head has a look of its own: its line as a paragraph, and its
// technologies as details, each labelled by its name, its value how long it was used; each where the head has it.
func (head ExperienceHead) Blocks() []Block {
	blocks := []Block{}
	if len(head.Line) > 0 {
		blocks = append(blocks, Paragraph{Spans: head.Line})
	}
	if len(head.Technologies) > 0 {
		facts := make([]Fact, 0, len(head.Technologies))
		for _, technology := range head.Technologies {
			facts = append(facts, Fact{Label: technology.Name, Value: technology.Used})
		}
		blocks = append(blocks, Facts{Kind: FactsDetails, List: facts})
	}
	return blocks
}

// Entry is a role, work for an employer or a client, over a span of time, with what came of it: the technologies it
// was done with, by name, its points, and its description, paragraphs that tell it at length.
type Entry struct {
	Period       Period
	Title        []Span
	Subtitle     []Span
	Technologies []string
	Points       []Point
	Description  []Paragraph
}

// Point is one thing an entry or a project's section holds up.
type Point struct {
	Spans []Span
}

// Points are a project's section's points, after its paragraphs, each drawn as an entry's.
type Points struct {
	List []Point
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

func (Section) block()        {}
func (Paragraph) block()      {}
func (Table) block()          {}
func (CodeBlock) block()      {}
func (List) block()           {}
func (Quote) block()          {}
func (Pause) block()          {}
func (Figure) block()         {}
func (Widget) block()         {}
func (Facts) block()          {}
func (ExperienceHead) block() {}
func (Entry) block()          {}
func (Points) block()         {}
func (Card) block()           {}
func (Contents) block()       {}
func (Checks) block()         {}
func (Skills) block()         {}
