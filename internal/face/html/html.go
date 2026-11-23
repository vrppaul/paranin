// Package html draws documents as the pages a browser shows (docs/ARCHITECTURE.md, ADR-0002, ADR-0003), with
// templ: each kind of block and span is a component of page.templ, and the choice between the kinds is Go's, here,
// where gochecksumtype holds every switch to every kind, as it cannot inside generated code. A page names the
// language of its document, takes its title and description from its head, both as plain text, and names the
// address it is known by, whole, and the path of its markdown; a page with an address carries Open Graph's tags
// for the preview of a shared link, and the record compose makes beside its document, in schema.org's words, as
// typed structs written as JSON-LD (linkeddata.go). Every address is written by routes; a link inside the page
// leads to a page of the site by its path from the root, at the address where the reader's request chooses the
// format, which a browser reads as the page, so a page served from this machine leads to this machine. An address
// of no page reads as NotFound: its status, the path asked for, quoted, and a link to where the site begins under
// its name.
//
// Every page, the page of no page among them, includes what the build it is served with (assets) brings to run the
// site's script: the stylesheet, which styles its markup, what each part is, a heading by its depth, a period's
// months as times; and the script, which runs the theme button in the foot line every page ends with (frame.go). A
// page shows one tab of its document, as the TUI does. A page's title, and each tab's, stands in the design's block
// letters (letters.go), drawn as SVG a screen reader skips, the heading named by its words; a title with a letter the
// font does not draw stays its words.
// The résumé whose first section holds facts opens with its first screen (firstscreen.go): the greeting, the name and
// the lead beside the brief, the mail heading it, then the title, the place and the first facts, read off the
// sections as the TUI reads its brief (document.BriefOf), the ssh command as a chip the script lets copy (frame.go),
// and the ways to the author's profiles; the facts the brief leaves stay in their section, where the brief stood. A
// fact whose value is a command alone, as a project's install line, stands as such a chip too, and every page's foot
// line holds the toast that says what a chip copied; a fact whose value is one link to a mailbox stands as the mail,
// in the size of the text round it.
// A page shows a document as a reader of the finished work sees it (document.WrittenBlocks), a fact or a point that is
// only a gap still to fill left out; a gap among written words, and a month still to fill, stand marked as gaps.
// A page that is not the résumé opens with its head (DESIGN.md, Contents and PageHead), its blocks after it. A post
// is drawn whole, with no tabs (post.go): its head as a first screen, its series' pick among its facts, its outline
// beside its article, its paragraphs numbered and each note beside the line that first cites it, then the panes at its
// foot leading to the posts beside it, in the blog's session of the foot line. The blog's list (list.go) opens with
// the blog's name in block letters, its line, the Switch of its groupings and its tags, then each group under its name,
// its posts as tiles, each a link to its post.
// Every page links the pictures of the site (icon.go), which it draws for its host to serve: a square of the name's
// gradient, as SVG and as pixels.
package html

//go:generate go -C ../../.. run ./scripts/templates -write

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/a-h/templ"

	"github.com/vrppaul/paranin/internal/assets"
	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// How a known month is written: as a reader reads it, and as a time's datetime holds it. Each face writes a month
// its own way (ADR-0003); the text faces write theirs as this one does, by choice, not by one rule
const (
	monthLayout        = "01/2006"
	machineMonthLayout = "2006-01"
)

// The media type of a page's markdown, which its head names beside the link to it
const markdownType = "text/markdown"

// The kind of thing Open Graph says every page is, as the first version said it. The résumé could be a profile,
// which asks for the person's names in fields of its own; until a page needs that, one kind serves every page
const openGraphType = "website"

// The depth of a section's heading under the page's title, h2, and the deepest HTML has, h6; the name of a heading no
// reader links to, which carries none, the attribute a named one carries its name in, and the one a heading the site's
// script finds carries its part in
const (
	firstHeadingDepth   = 2
	deepestHeadingDepth = 6
	unnamed             = ""
	idAttribute         = "id"
	partAttribute       = "data-part"
)

// A space a line never breaks at, U+00A0, written as its bytes in UTF-8 so the code shows it
const noBreakSpace = "\xc2\xa0"

// What stands between the parts of one line, the page's own choice as each face's is: the text faces write the same
// marks, with a plain space before the dot
const (
	// a title and its place, a card's tag and its subtitle, a project's tags; a space that never breaks before the dot,
	// so the dot ends a line and never opens the next
	dot = noBreakSpace + "· "
	// a title and its subtitle
	dash = " — "
	// the start and the end of a period
	rangeDash = " – "
	// a listing's title and its line
	labelEnd = ": "
	// a point's mark and its words, so a point copied reads "- " and its words
	pointSpace = " "
)

// The theme button as the site's script finds and runs it (web/src/controllers/theme.ts): the part it stands in,
// and the id of the words that name it and show as its tooltip; and the part a tooltip is, which Esc hides
// (web/src/controllers/tooltip.ts)
const (
	themePart   = "theme"
	themeNameID = "theme-name"
	tooltipPart = "tooltip"
)

// The vim keys as the site's script finds and runs them (web/src/controllers/vimswitch.ts and commandline.ts): the
// switch that turns them on and its words; the command line they open, the part a reader types in and its field, the
// field's name to a screen reader and the mark before it; the menu of its places, each place and the letters of its
// name typed and not; what the line says of a command that failed; and the keys it names, Tab's among them, which
// shows while a place fits. A part and an id share a name where the field names the part by its id
const (
	vimKeysPart      = "vim-keys"
	vimKeysWords     = "vim keys"
	commandLinePart  = "command-line"
	commandEntryPart = "command-entry"
	commandFieldPart = "command-field"
	commandName      = "command"
	commandMark      = ":"
	placesPart       = "places"
	placesName       = "places"
	placePart        = "place"
	placeTypedPart   = "place-typed"
	placeRestPart    = "place-rest"
	commandSaidPart  = "command-said"
	commandKeysPart  = "command-keys"
	nextKeyPart      = "command-next"
)

// The search the vim keys' slash opens on the same line (web/src/controllers/commandline.ts): the colon's mark and the
// slash's, which the script shows by the key that opened the line, the field's name while a search is typed, and the
// keys a search names, the count of its matches first; and the foot line, which a match brought into sight stands
// clear of (web/src/controllers/search.ts)
const (
	footPart        = "foot"
	commandMarkPart = "command-mark"
	searchMarkPart  = "search-mark"
	searchMark      = "/"
	searchName      = "search"
	searchKeysPart  = "search-keys"
	searchCountPart = "search-count"
)

// The id of a place on the command line's menu, its number after this, which the field names as the one chosen
const placeIDOpening = "place-"

// The keys the command line names, each with what it does, as the TUI's names them; Esc closes a hint's window too
const (
	tabKey     = "tab"
	nextHint   = "next"
	enterMark  = "↵"
	goHint     = "go"
	findHint   = "find"
	escapeKey  = "esc"
	cancelHint = "cancel"
	closeHint  = "close"
)

// How the foot line writes tmux's parts: the session between its brackets, and the colon between a window's number
// and its name
const (
	sessionOpening = "["
	sessionClosing = "]"
	windowColon    = ":"
)

// The box of sixteen units every drawn mark stands in (DESIGN.md, Iconography): the theme button's and the corners'
// crosses
const markViewBox = "0 0 16 16"

// A command's chip as the site's script finds and runs it (web/src/controllers/copy.ts): the chip, the words it
// copies, and its button, hidden until the script shows it; the toast that says what was copied, with the place its
// words go; and the hint's window a chip may stand in, which a copy closes
const (
	commandPart      = "command"
	commandWordsPart = "command-words"
	copyPart         = "copy"
	toastPart        = "toast"
	toastWordsPart   = "toast-words"
	hintPart         = "hint"
)

// A post's block of code as the site's script finds and runs it (web/src/controllers/copy.ts), copied by its button
// alone, as a command's chip is by a press anywhere on it: the block, and the code it copies
const (
	codePart      = "code"
	codeWordsPart = "code-words"
)

// PageWriter writes the pages of one snapshot, from what is the same on each, worked out once: the site, which
// writes their addresses, what each includes to run the site's script, the frame round each, and the day they are
// drawn on, which the résumé's career line runs to.
type PageWriter struct {
	site     routes.Site
	included assets.Included
	frame    frame
	today    time.Time
}

// pageMetadata is what a page says of itself to whoever reads it without showing it: its title and description
// as plain text, the address it is known by, whole, and the path of its markdown; each empty where it has none.
type pageMetadata struct {
	title       string
	description string
	canonical   string
	markdown    string
}

// NewPageWriter writes the pages of a site, drawn in a build's stylesheet and running its script, with the frame
// round each, on a day the host reads off its clock.
func NewPageWriter(site routes.Site, build *assets.Build, composed compose.Frame, today time.Time) PageWriter {
	return PageWriter{site: site, included: build.Includes(assets.Site), frame: frameOf(site, composed), today: today}
}

// Page writes a document as a page of HTML showing one of its tabs (document.TabsOf), parted as every face parts it, by
// the tab's name, its opening where the name is none of theirs: what the page says of itself, its record as linked
// data where it has one, the tab as its main part, as a reader of the finished work sees it (document.WrittenBlocks),
// then the foot line, every tab a window there.
func (writer PageWriter) Page(composed document.Document, record compose.Record, shown string) string {
	head := composed.Head
	tabs := document.TabsOf(composed)
	page, title := head.Page, head.Title
	var main templ.Component
	if tab, isTab := tabNamed(tabs.Rest, shown); isTab {
		page, title = document.Page{Kind: head.Page.Kind, Slug: head.Page.Slug, Tab: tab.Slug}, tab.Title
		main = writer.tabPart(head.Page.Kind, tab.Title, document.WrittenBlocks(tab.Blocks))
	} else {
		main = writer.openingPart(composed, tabs.Opening)
	}
	return rendered(writer.page(head.Language, writer.metadataOf(head, page, title), record, main, writer.footOf(page, tabs)))
}

// tabNamed is the tab of a name among tabs, where one has it.
func tabNamed(tabs []document.Tab, name string) (document.Tab, bool) {
	index := slices.IndexFunc(tabs, func(tab document.Tab) bool { return tab.Slug == name })
	if index < 0 {
		return document.Tab{Title: nil, Slug: "", Blocks: nil}, false
	}
	return tabs[index], true
}

// rendered is the HTML a component writes.
func rendered(component templ.Component) string {
	var builder strings.Builder
	// A builder takes every write, no component here returns an error of its own, and the linked data is structs
	// of strings, which JSON always encodes
	_ = component.Render(context.Background(), &builder)
	return builder.String()
}

// metadataOf is what a page of a head says of itself, the page shown under a title: its document's whole or a tab.
func (writer PageWriter) metadataOf(head document.Head, shown document.Page, title []document.Span) pageMetadata {
	return pageMetadata{
		title:       document.PlainText(title),
		description: document.PlainText(head.Description),
		canonical:   writer.site.Address(shown, routes.FormatChosen),
		markdown:    routes.Reference(shown, routes.FormatMarkdown),
	}
}

// block is the component of a block, at the depth of its headings.
func (writer PageWriter) block(block document.Block, depth int) templ.Component {
	switch typed := block.(type) {
	case document.Section:
		return writer.section(typed, depth)
	case document.Paragraph:
		return writer.paragraph(typed)
	case document.Table:
		return writer.table(typed)
	case document.CodeBlock:
		return codeBlock(typed.Text)
	case document.Facts:
		return writer.facts(typed.List)
	case document.Entry:
		return writer.entry(typed, depth)
	case document.Card:
		return writer.card(typed, depth)
	case document.Contents:
		return writer.contents(typed)
	case document.Checks, document.Skills:
		// only the hidden pages hold checks and the list of skills, and no host hands them to a page: one served
		// with them would fail TestEveryPhraseReachesThePage
	case document.List, document.Quote:
		// a post's alone, which its article draws (post.go)
	case document.Posts:
		return writer.tiles(typed)
	}
	return templ.NopComponent
}

// setBack is whether a table's cell of a kind stands back from the words beside it, as a duration does.
func setBack(kind document.CellKind) bool {
	switch kind {
	case document.CellText:
		return false
	case document.CellDuration:
		return true
	}
	// the zero kind, which compose never builds, reads as words
	return false
}

// span is the component of a span; a link whose target has no address reads as its text.
func (writer PageWriter) span(span document.Span) templ.Component {
	switch typed := span.(type) {
	case document.Text:
		return text(typed.Text)
	case document.Code:
		return code(typed.Text)
	case document.Gap:
		return gap(typed.Text)
	case document.Highlight:
		return writer.highlight(typed.Spans)
	case document.Link:
		if reference := routes.Reference(typed.Target, routes.FormatChosen); reference != "" {
			return writer.link(reference, typed.Spans)
		}
		return writer.spans(typed.Spans)
	case document.Emphasis:
		return writer.emphasis(typed.Spans)
	case document.Strong:
		return writer.strong(typed.Spans)
	case document.Note:
		// a note's words stand beside the line or at the article's end (notesBeside); where a line cites it, its number
		return citation(typed.Number)
	}
	return templ.NopComponent
}

// factValue is the component of a fact's value by its shape: a command alone as a chip that copies it (DESIGN.md,
// Command), one link to a mailbox as the Mail in the size of the text round it, anything else as its spans.
func (writer PageWriter) factValue(fact document.Fact) templ.Component {
	if text, isCommand := fact.Command(); isCommand {
		return command(writer.frame.copyWords.Of(text))
	}
	if mail, isMail := document.MailOf(fact.Value); isMail {
		return mailLink(mail, false)
	}
	return writer.spans(fact.Value)
}

// month is the component of a month: a known one as a time, a gap as a gap, what still goes on as its text.
func month(month document.Month) templ.Component {
	switch typed := month.(type) {
	case document.KnownMonth:
		return knownMonth(typed.Start)
	case document.GapMonth:
		return gap(typed.Text)
	case document.CurrentMonth:
		return text(typed.Text)
	}
	return templ.NopComponent
}

// heading is a title at a depth, from a section's under the page's title, under the name a reader links to it by,
// none where it is empty; the depths past HTML's deepest heading are written at it.
func heading(depth int, id string, title templ.Component) templ.Component {
	named := templ.Attributes{}
	if id != unnamed {
		named[idAttribute] = id
	}
	return headingWith(depth, named, title)
}

// headingWith is a title at a depth, as heading is, with the attributes it is named and found by.
func headingWith(depth int, named templ.Attributes, title templ.Component) templ.Component {
	levels := headingLevels()
	return levels[min(depth, deepestHeadingDepth)-firstHeadingDepth](named, title)
}

// headingLevels are HTML's headings below the page's title, from the shallowest, each with the attributes it is
// named by.
func headingLevels() []func(templ.Attributes, templ.Component) templ.Component {
	return []func(templ.Attributes, templ.Component) templ.Component{headingTwo, headingThree, headingFour, headingFive, headingSix}
}

// linked are spans as a link to a page, where the page has an address: a card's title, a listing's.
func linked(page document.Page, spans []document.Span) []document.Span {
	if page.Kind == document.PageNone {
		return spans
	}
	return []document.Span{document.Link{Target: page, Spans: spans}}
}
