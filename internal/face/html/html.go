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
// months as times; and the script, which runs the theme button in the top bar every page opens with (frame.go). A
// page's title, and on the résumé each section's, stands in the design's block letters (letters.go), drawn as SVG a
// screen reader skips, the heading named by its words; a title with a letter the font does not draw stays its words.
// The résumé whose first section holds facts opens with its first screen (firstscreen.go): the greeting, the name and
// the lead beside the brief, the mail heading it, then the title, the place and the first facts, read off the
// sections as the TUI reads its brief (document.BriefOf), and the ssh command as a chip the script lets copy
// (frame.go); the facts the brief leaves stay in their section, where the brief stood. A fact whose value is a
// command alone, as a project's install line, stands as such a chip too, and every document's page holds the toast
// that says what a chip copied.
// A page shows a document as a reader of the finished work sees it (document.Written), a fact or a point that is only
// a gap still to fill left out; a gap among written words, and a month still to fill, stand marked as gaps.
// The classes of the page's other parts arrive with its design (TODO, "The look").
package html

//go:generate go -C ../../.. run ./scripts/templates -write

import (
	"context"
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

// The depth of a section's heading under the page's title, h2, and the deepest HTML has, h6
const (
	firstHeadingDepth   = 2
	deepestHeadingDepth = 6
)

// What stands between the parts of one line, the page's own choice as each face's is: the text faces' are the
// same today, and the page's stylesheet may part them
const (
	// a title and its place, a card's tag and its subtitle
	dot = " · "
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
// and the id of the words that name it and show as its tooltip; the part a tooltip is, which Esc hides
// (web/src/controllers/tooltip.ts); and the top bar's links to the résumé's sections and each section they lead to,
// which mark the section the reader is in (web/src/controllers/sections.ts)
const (
	themePart    = "theme"
	themeNameID  = "theme-name"
	tooltipPart  = "tooltip"
	sectionsPart = "sections"
	sectionPart  = "section"
)

// The box of sixteen units every drawn mark stands in (DESIGN.md, Iconography): the theme button's and the corners'
// crosses
const markViewBox = "0 0 16 16"

// A command's chip as the site's script finds and runs it (web/src/controllers/command.ts): the chip, the words it
// copies, and its button, hidden until the script shows it; and the toast that says what was copied, with the place
// its words go
const (
	commandPart      = "command"
	commandWordsPart = "command-words"
	copyPart         = "copy"
	toastPart        = "toast"
	toastWordsPart   = "toast-words"
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

// Page writes a document as a page of HTML: what it says of itself, its record as linked data where it has one,
// then the top bar, and its main part: its head, then each block, but for a fact or a point that is only a gap.
func (writer PageWriter) Page(composed document.Document, record compose.Record) string {
	return rendered(writer.page(document.Written(composed), writer.metadataOf(composed.Head), record))
}

// rendered is the HTML a component writes.
func rendered(component templ.Component) string {
	var builder strings.Builder
	// A builder takes every write, no component here returns an error of its own, and the linked data is structs
	// of strings, which JSON always encodes
	_ = component.Render(context.Background(), &builder)
	return builder.String()
}

// metadataOf is what a page of a head says of itself.
func (writer PageWriter) metadataOf(head document.Head) pageMetadata {
	return pageMetadata{
		title:       document.PlainText(head.Title),
		description: document.PlainText(head.Description),
		canonical:   writer.site.Address(head.Page, routes.FormatChosen),
		markdown:    routes.Reference(head.Page, routes.FormatMarkdown),
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
	}
	return templ.NopComponent
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
	}
	return templ.NopComponent
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

// heading is a title at a depth, from a section's under the page's title; the depths past HTML's deepest heading
// are written at it.
func heading(depth int, title templ.Component) templ.Component {
	levels := headingLevels()
	return levels[min(depth, deepestHeadingDepth)-firstHeadingDepth](title)
}

// headingLevels are HTML's headings below the page's title, from the shallowest.
func headingLevels() []func(templ.Component) templ.Component {
	return []func(templ.Component) templ.Component{headingTwo, headingThree, headingFour, headingFive, headingSix}
}

// linked are spans as a link to a page, where the page has an address: a card's title, a listing's.
func linked(page document.Page, spans []document.Span) []document.Span {
	if page.Kind == document.PageNone {
		return spans
	}
	return []document.Span{document.Link{Target: page, Spans: spans}}
}

func monthText(start time.Time) string {
	return start.Format(monthLayout)
}

func machineMonth(start time.Time) string {
	return start.Format(machineMonthLayout)
}
