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
// gradient, as SVG and as pixels. A page of the blog names in its head the feeds that follow it, and a post's body is
// drawn once more for a feed reader (feed.go), as plain HTML, every address whole.
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

// The kinds of thing Open Graph says a page is: a post an article, every other page a website. The résumé could be a
// profile, which asks for the person's names in fields of its own; until a page needs that, a website serves it
const (
	articleType = "article"
	websiteType = "website"
)

// The card X shows of a page, which reads the rest of it from Open Graph: a summary, or, where the page names a card of
// its own, the card large
const (
	twitterSummary    = "summary"
	twitterLargeImage = "summary_large_image"
)

// What the robots' line of an indexed page says where it names a card: a search engine may show the card large
const largeImagePreview = "max-image-preview:large"

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
	// a title and its place, a card's tag and its subtitle, a project's tags, the blog's name and a list's grouping in
	// the title the list says of itself; a space that never breaks before the dot, so the dot ends a line and never
	// opens the next
	dot = noBreakSpace + "· "
	// a title and its subtitle; in the title a page says of itself, a tab's title and its page's, the blog's name and
	// the list's filter
	dash = " — "
	// the start and the end of a period
	rangeDash = " – "
	// a listing's title and its line
	labelEnd = ": "
	// a point's mark and its words, so a point copied reads "- " and its words
	pointSpace = " "
	// the words of a command, where alone its lines break, so a copy reads the command as written
	commandSpace = " "
	// an install's note, before the command it says when to take
	noteEnd = ":"
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
// name typed and not; what the line says of a command that failed; the keys it names, Tab's among them, which shows
// while a place fits; and each blog's folder :e opens a post in. A part and an id share a name where the field names
// the part by its id
const (
	vimKeysPart      = "vim-keys"
	vimKeysWords     = "vim keys"
	commandLinePart  = "command-line"
	commandEntryPart = "command-entry"
	commandFieldPart = "command-field"
	commandName      = "command"
	commandNamePart  = "command-name"
	commandMark      = ":"
	placesPart       = "places"
	placesName       = "places"
	placesNamePart   = "places-name"
	placePart        = "place"
	placeTypedPart   = "place-typed"
	placeRestPart    = "place-rest"
	commandSaidPart  = "command-said"
	commandKeysPart  = "command-keys"
	nextKeyPart      = "command-next"
	folderPart       = "folder"
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
	searchNamePart  = "search-name"
	searchKeysPart  = "search-keys"
	searchCountPart = "search-count"
)

// The name of a place on the command line's menu, its number after this, which the field names as the one chosen
const placeIDOpening = "place-"

// What opens the id of every part of the page's own: a mark document.SlugOf never writes and no slug holds, so the
// page's anchors, a heading's and a series' group's (document.Section's Name), are theirs whatever their words
const ownIDMark = "_"

// ownID is the id of a part of the page's own by its name, in the one form no anchor takes (ownIDMark).
func ownID(name string) string {
	return ownIDMark + name
}

// The language the frame's own words are written in on every page, Neovim's and tmux's, and HTTP's status on the page
// of no page, which no translation changes; and what names a title in its page's own language, none other
const (
	frameLanguage = "en"
	pageLanguage  = ""
)

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
// words go and the template each word is drawn from; and the hint's window a chip may stand in, which a copy closes
const (
	commandPart      = "command"
	commandWordsPart = "command-words"
	copyPart         = "copy"
	toastPart        = "toast"
	toastWordsPart   = "toast-words"
	toastWordPart    = "toast-word"
	hintPart         = "hint"
)

// A post's block of code as the site's script finds and runs it (web/src/controllers/copy.ts), copied by its button
// alone, as a command's chip is by a press anywhere on it: the block, and the code it copies
const (
	codePart      = "code"
	codeWordsPart = "code-words"
)

// A post's widget as the site's script mounts it (web/src/islands/widgets.ts): the body of its window its component
// runs in, the mark that it runs and the alt text that stands for it until then
const (
	widgetBodyPart = "widget-body"
	widgetLivePart = "widget-live"
	widgetAltPart  = "widget-alt"
)

// PageWriter writes the pages of one snapshot, from what is the same on each, worked out once: how it writes text
// inside a block, each link by its path from the root; what each includes to run the site's script, the frame round
// each, whose site writes their addresses, the cards the pages name, and the day they are drawn on, which the résumé's
// career line runs to.
type PageWriter struct {
	spanWriter
	included assets.Included
	cards    *assets.Cards
	frame    frame
	today    time.Time
}

// pageMetadata is what a page says of itself to whoever reads it without showing it: its title and description
// as plain text and the path of its markdown, each empty where it has none; the address it names; what Open Graph says
// it is; its record as linked data; the feeds that follow it, which a feed reader finds it by; and the page in each
// language it is read in, which a search engine reads beside the address it is known by.
type pageMetadata struct {
	title         string
	description   string
	address       pageAddress
	markdown      string
	openGraphType string
	record        templ.Component
	// card is the card the page shows; none where the page shows none or the cards' manifest names none for it
	card *socialCard
	// post is a post's record as its Open Graph and its linked data both write it; none where the page's record is no
	// post's
	post       *writtenPost
	feeds      []headFeed
	alternates []alternate
	// otherLocales are the locales of the languages a post is read in but the one read, which Open Graph names too
	otherLocales []string
}

// socialCard is the card a page names: its address, whole, and the words that stand for it where it is not seen.
type socialCard struct {
	url string
	alt string
}

// pageAddress is the address a page names in its head, by whether it has one of its own and whether a search engine
// may index it (document.Head's Indexed).
//
//sumtype:decl
type pageAddress interface {
	pageAddress()
}

// noAddress is no address of the page's own: no search engine indexes it and no preview shows it.
type noAddress struct{}

// shownAddress is the page's own address, whole, which a link's preview shows and no search engine indexes.
type shownAddress struct {
	url string
}

// knownAddress is the address, whole, a search engine knows the page by, which a link's preview shows too.
type knownAddress struct {
	url string
}

func (noAddress) pageAddress()    {}
func (shownAddress) pageAddress() {}
func (knownAddress) pageAddress() {}

// headFeed is a feed a page names in its head: its title, as plain text, and its path from the root.
type headFeed struct {
	title string
	path  string
}

// alternate is the page in a language for a search engine: the language's tag, or defaultLanguage, and its address.
type alternate struct {
	language string
	address  string
}

// What a head names the page in the site's own language as, for a reader of a language the site offers no page in
const defaultLanguage = "x-default"

// alternatesOf are the page in each language a post is read in, then the site's own language's again as the default.
func alternatesOf(languages []document.Language, site routes.Site) []alternate {
	alternates := make([]alternate, 0, len(languages))
	for _, language := range languages {
		alternates = append(alternates, alternate{language: language.Tag, address: site.Address(language.Page, routes.FormatChosen)})
	}
	for _, language := range languages {
		if language.Page.Translation == document.Untranslated {
			alternates = append(alternates, alternate{language: defaultLanguage, address: site.Address(language.Page, routes.FormatChosen)})
		}
	}
	return alternates
}

// otherLocalesOf are the locales of the languages a post is read in but the one read.
func otherLocalesOf(languages []document.Language) []string {
	locales := []string{}
	for _, language := range languages {
		if !language.Read {
			locales = append(locales, language.Locale)
		}
	}
	return locales
}

// NewPageWriter writes the pages of a site, drawn in a build's stylesheet and running its script, naming each page's
// card as the cards' manifest has it, with the frame round each, on a day the host reads off its clock.
func NewPageWriter(site routes.Site, build *assets.Build, cards *assets.Cards, composed compose.Frame, today time.Time) PageWriter {
	return PageWriter{
		spanWriter: spanWriter{addressOf: routes.Reference}, included: build.Includes(assets.Site), cards: cards, frame: frameOf(site, composed),
		today: today,
	}
}

// Page writes a document as a page of HTML showing one of its tabs (document.TabsOf), parted as every face parts it, by
// the tab's name, its opening where the name is none of theirs: what the page says of itself, a tab by its title and
// its page's, since two pages may hold tabs of one title; its record as linked data where it has one, the tab as its
// main part, as a reader of the finished work sees it (document.WrittenBlocks), then the foot line, every tab a window
// there.
func (writer PageWriter) Page(composed document.Document, record compose.Record, shown string) string {
	head := composed.Head
	tabs := document.TabsOf(composed)
	page, title := head.Page, head.Title
	var main templ.Component
	if tab, isTab := tabNamed(tabs.Rest, shown); isTab {
		page = document.Page{Kind: head.Page.Kind, Slug: head.Page.Slug, Tab: tab.Slug, Revision: "", Translation: head.Page.Translation}
		title = joinedTitle(tab.Title, dash, head.Title)
		main = writer.tabPart(head.Page.Kind, tab.Title, document.WrittenBlocks(tab.Blocks))
	} else {
		main = writer.openingPart(composed, tabs.Opening)
	}
	return rendered(writer.page(head.Language, writer.metadataOf(head, page, title, record), main, writer.footOf(head, page, tabs, document.PlacesOf(composed))))
}

// joinedTitle is a title a page says of itself made of two, a mark between them.
func joinedTitle(first []document.Span, mark string, second []document.Span) []document.Span {
	return slices.Concat(first, []document.Span{document.Text{Text: mark}}, second)
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

// metadataOf is what a page of a head says of itself, the page shown under a title: its document's whole or a tab,
// with its record.
func (writer PageWriter) metadataOf(head document.Head, shown document.Page, title []document.Span, record compose.Record) pageMetadata {
	site := writer.frame.site
	url := site.Address(shown, routes.FormatChosen)
	words := document.PlainText(title)
	card := writer.cardOf(head, words)
	linked, post := writer.recordOf(record, shown, url, card)
	metadata := pageMetadata{
		title:         words,
		description:   document.PlainText(head.Description),
		address:       addressOf(head, url),
		markdown:      routes.Reference(shown, routes.FormatMarkdown),
		openGraphType: openGraphTypeOf(shown.Kind),
		record:        linked,
		card:          card,
		post:          post,
		feeds:         make([]headFeed, 0, len(head.Feeds)),
		alternates:    alternatesOf(head.Languages, site),
		otherLocales:  otherLocalesOf(head.Languages),
	}
	for _, feed := range head.Feeds {
		metadata.feeds = append(metadata.feeds, headFeed{title: document.PlainText(feed.Spans), path: routes.Reference(feed.Target, routes.FormatChosen)})
	}
	return metadata
}

// cardOf is the card a page of a head shows, by the name the cards' manifest gives its page's, with the page's title as
// its alt; none where the page shows no card or the manifest names none.
func (writer PageWriter) cardOf(head document.Head, alt string) *socialCard {
	name, named := writer.cards.NameOf(routes.Path(head.CardPage(), routes.FormatChosen))
	if !named {
		return nil
	}
	return &socialCard{url: writer.frame.site.CardAddress(name), alt: alt}
}

// addressOf is the address a page of a head names, the page shown at a URL: its document's whole or a tab, which
// shares its document's address of its own (document.Head's Previewed and Indexed).
func addressOf(head document.Head, url string) pageAddress {
	if !head.Previewed() {
		return noAddress{}
	}
	if !head.Indexed() {
		return shownAddress{url: url}
	}
	return knownAddress{url: url}
}

// openGraphTypeOf is what Open Graph says a page of a kind is: a post an article, any other a website.
func openGraphTypeOf(kind document.PageKind) string {
	switch kind {
	case document.PagePost:
		return articleType
	case document.PageNone, document.PageResume, document.PageProject, document.PageIndex, document.PageHealth, document.PageSkills,
		document.PageBlog, document.PageBlogSeries, document.PageTagFilter:
	}
	return websiteType
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
		return codeBlock(typed.Runs)
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
	case document.List, document.Quote, document.Pause, document.Figure, document.Widget:
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

// factValue is the component of a fact's value by its shape: a command alone as a chip that copies it (DESIGN.md,
// Command), one link to a mailbox as the Mail in the size of the text round it, anything else as its spans.
func (writer PageWriter) factValue(fact document.Fact) templ.Component {
	if text, isCommand := fact.Command(); isCommand {
		return command(writer.frame.copyWords.Of(text))
	}
	if mail, isMail := document.MailOf(fact.Value); isMail {
		return writer.mailLink(mail, false)
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
