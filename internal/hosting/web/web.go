// Package web serves the site over HTTP (docs/ARCHITECTURE.md, ADR-0003, ADR-0004). It draws every page of a release's
// snapshot once, the first time a request meets that snapshot (release.Derived), and serves the drawing until the
// store holds another, so no request composes or draws a page; beside the pages it serves every file of the
// snapshot's build at its path, from memory, for a cache to keep for good, and each picture of the site the HTML face
// draws with them, which a cache asks for again. Each page, and each of its tabs alone, is served at the path routes
// writes for it in each format, and at its own address in the format the reader's request asks for: HTML for a
// browser, markdown for an agent that asks for it, and the manual for `curl` and everyone else; the index at
// llms.txt, sent as plain text, as a name ending in .txt promises. A post served is drawn whole in every format, and
// its number sends the reader on to its address. Each format of a page names in
// a Link header the address the page is known by, the one where the reader chooses, so a search engine counts its
// markdown and its manual as the page. Any other path is answered with the page of no page, in text or in HTML,
// whichever the reader's request ranks higher: drawn for each request, since it quotes the path asked for, in the
// chosen format alone. Every answer, a refusal net/http writes among them, tells a browser not to guess its type and
// to load nothing but the site's own files, so markup slipped into a page runs no script but the site's own.
package web

import (
	"mime"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/vrppaul/paranin/internal/assets"
	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/face/html"
	"github.com/vrppaul/paranin/internal/face/text"
	"github.com/vrppaul/paranin/internal/release"
	"github.com/vrppaul/paranin/internal/routes"
)

// The headers every page is sent with: what it is, text of UTF-8 whatever its media type, and that a cache asks
// again, since a reload can change any page
const (
	utf8Text     = "; charset=utf-8"
	askEveryTime = "no-cache"
	contentType  = "Content-Type"
	cacheControl = "Cache-Control"
)

// How long a cache keeps a built file: a year, which caches read as for good, and with no need to ask again, since
// a file changed is built under another name
const keptForGood = "public, max-age=31536000, immutable"

// The headers every answer is sent with: that it is no other kind than it says, and what a browser may do with it
const (
	noSniffing     = "nosniff"
	typeOptions    = "X-Content-Type-Options"
	securityPolicy = "Content-Security-Policy"
)

// What a browser may do with an answer (docs/WEB.md, "Security headers"): load scripts, styles, images, fonts and
// whatever it fetches from the site's own files alone, so never a script or a style written inline nor a data:
// address; take no other base for the page's relative addresses; send a form to the site alone; and be framed by
// no site at all. The last three are named, since default-src does not reach them
const ownSiteAlone = "default-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'"

// How the header a page names its address in is written: <address>; rel="canonical"
const (
	linkHeader        = "Link"
	linkOpening       = "<"
	canonicalRelation = `>; rel="canonical"`
)

// How a reader's request chooses a format: by the media types its Accept header ranks. A cache keeps one copy
// for each Accept header it meets
const (
	acceptHeader      = "Accept"
	varyHeader        = "Vary"
	markdownMediaType = "text/markdown"
	plainMediaType    = "text/plain"
	htmlMediaType     = "text/html"
	mediaTypeEnd      = ","
	qualityParameter  = "q"
	qualityBits       = 64
	// a media type no range of the header names, and one a range names without saying how much
	unnamedQuality = 0
	fullQuality    = 1
	// what separates a media type's family from its kind, and the range of any family or any kind
	familyEnd = "/"
	anyRange  = "*"
)

// naming is how closely a range of an Accept header names a media type. The closest range a header holds gives
// the type its quality (RFC 9110, 12.5.1).
type naming int

const (
	namedByNone naming = iota
	namedByAny
	namedByFamily
	namedExactly
)

// The pattern every path a reader reads falls under
const pagePattern = "GET /"

// The status an address of no page is answered with, and what stands between a status's code and its reason in
// the line HTTP writes it in
const (
	missingStatus   = http.StatusNotFound
	statusSeparator = " "
)

// Server serves the pages and the built files of the store's current snapshot.
type Server struct {
	// drawn is every page of the store's current snapshot, drawn once for each snapshot
	drawn   *release.Derived[*drawing]
	handler http.Handler
}

// drawing is every page of one snapshot, drawn, by the path it is served at, and every file of its build: in
// each format at its own path, and each file, the same whatever the request asks; in every format at the address
// where the reader chooses; the path a post's number sends the reader on to, by the number's path; and the page of
// an address of no page.
type drawing struct {
	fixed    map[string]answer
	chosen   map[string]choice
	moved    map[string]string
	notFound missingChoice
}

// choice is one page in each format a reader may choose, never none, in the order a tie between two goes: the
// manual, then markdown, then HTML, so `curl` and a header naming none of them get the manual.
type choice []answer

// textAnswers are a page in the text formats, each naming the address the page is known by.
type textAnswers struct {
	markdown answer
	manual   answer
}

// missingChoice is the page of no page in each format a reader may choose, never none, in the order a tie between
// them goes: text, then HTML.
type missingChoice []missingFormat

// missingFormat is a format the page of no page is served in: its media type, and how it is drawn for the path
// asked for, which each request quotes in it.
type missingFormat struct {
	mediaType string
	draw      func(path string) string
}

// offer is a format a reader may choose, by the media type an Accept header ranks.
type offer interface {
	servedAs() string
}

func (sent answer) servedAs() string {
	return sent.mediaType
}

func (format missingFormat) servedAs() string {
	return format.mediaType
}

// answer is a page or a built file as it is sent: the media type a page is, which an Accept header ranks, empty for
// a built file, which no reader chooses; the header that says what it is; the Link header naming the address a page
// is known by, empty where it is none a person reads; how long a cache keeps it; and what it holds.
type answer struct {
	mediaType    string
	contentType  string
	link         string
	cacheControl string
	body         []byte
}

// drawnAs is a page's text, drawn, as a media type: text of UTF-8, as every page is, which a cache asks for again.
func drawnAs(mediaType, text string) answer {
	return answer{mediaType: mediaType, contentType: mediaType + utf8Text, link: "", cacheControl: askEveryTime, body: []byte(text)}
}

// builtAs is a file of the build as it is served, which a cache keeps for good: a changed file is built under
// another name.
func builtAs(file assets.File) answer {
	return answer{mediaType: "", contentType: file.ContentType, link: "", cacheControl: keptForGood, body: file.Body}
}

// picturedAs is a picture of the site as it is served, which a cache asks for again: its path keeps its name when the
// author's name changes.
func picturedAs(icon html.Icon) answer {
	return answer{mediaType: "", contentType: icon.MediaType, link: "", cacheControl: askEveryTime, body: icon.Body}
}

// knownAt is a page naming the address it is known by, as a Link header writes it (RFC 8288, RFC 6596).
func (sent answer) knownAt(address string) answer {
	sent.link = linkOpening + address + canonicalRelation
	return sent
}

// New is a server of the site the store holds.
func New(store *release.Store) *Server {
	mux := http.NewServeMux()
	server := &Server{drawn: release.Derive(store, draw), handler: mux}
	mux.HandleFunc(pagePattern, server.servePage)
	return server
}

// ServeHTTP serves a request: a page or a file read, or a refusal. Every answer passes here, a refusal net/http
// writes itself among them, so here it is told what a browser may do with it.
func (server *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	header := writer.Header()
	header.Set(typeOptions, noSniffing)
	header.Set(securityPolicy, ownSiteAlone)
	server.handler.ServeHTTP(writer, request)
}

func (server *Server) servePage(writer http.ResponseWriter, request *http.Request) {
	drawn := server.drawn.Current()
	if found, isFixed := drawn.fixed[request.URL.Path]; isFixed {
		send(writer, http.StatusOK, found)
		return
	}
	if location, isMoved := drawn.moved[request.URL.Path]; isMoved {
		// a post's folder, its slug, may be renamed, so a cache asks again where its number leads
		writer.Header().Set(cacheControl, askEveryTime)
		http.Redirect(writer, request, location, http.StatusMovedPermanently)
		return
	}
	writer.Header().Set(varyHeader, acceptHeader)
	accept := request.Header.Get(acceptHeader)
	if offered, isChosen := drawn.chosen[request.URL.Path]; isChosen {
		send(writer, http.StatusOK, rankedFirst(accept, offered))
		return
	}
	missing := rankedFirst(accept, drawn.notFound)
	send(writer, missingStatus, drawnAs(missing.mediaType, missing.draw(request.URL.Path)))
}

// rankedFirst is the format an Accept header ranks highest, the first of them where it ranks two alike: a browser,
// which ranks HTML above anything else, gets the page it shows. The formats are never none.
func rankedFirst[Offer offer](accept string, formats []Offer) Offer {
	first := formats[0]
	chosen, best := first, quality(accept, first.servedAs())
	for _, format := range formats[1:] {
		if ranked := quality(accept, format.servedAs()); ranked > best {
			chosen, best = format, ranked
		}
	}
	return chosen
}

// quality is the quality an Accept header gives a media type: that of the closest range naming it, nothing where
// no range does.
func quality(accept, mediaType string) float64 {
	best, closest := float64(unnamedQuality), namedByNone
	for part := range strings.SplitSeq(accept, mediaTypeEnd) {
		mediaRange, parameters, err := mime.ParseMediaType(part)
		if err != nil {
			continue
		}
		if named := namingOf(mediaRange, mediaType); named > closest {
			best, closest = qualityOf(parameters), named
		}
	}
	return best
}

// namingOf is how closely a range names a media type.
func namingOf(mediaRange, mediaType string) naming {
	rangeFamily, rangeKind, _ := strings.Cut(mediaRange, familyEnd)
	family, _, _ := strings.Cut(mediaType, familyEnd)
	switch {
	case mediaRange == mediaType:
		return namedExactly
	case rangeFamily == family && rangeKind == anyRange:
		return namedByFamily
	case rangeFamily == anyRange && rangeKind == anyRange:
		return namedByAny
	default:
		return namedByNone
	}
}

// qualityOf is the quality a range's parameters give, full where they give none or one that does not read.
func qualityOf(parameters map[string]string) float64 {
	value, err := strconv.ParseFloat(parameters[qualityParameter], qualityBits)
	if err != nil {
		return fullQuality
	}
	return value
}

func send(writer http.ResponseWriter, status int, sent answer) {
	header := writer.Header()
	header.Set(contentType, sent.contentType)
	if sent.link != "" {
		header.Set(linkHeader, sent.link)
	}
	header.Set(cacheControl, sent.cacheControl)
	writer.WriteHeader(status)
	// A write fails only when the reader has gone, and nobody is left to tell
	_, _ = writer.Write(sent.body)
}

// statusLine is a status as HTTP writes it in a response's first line: its code, then its reason.
func statusLine(status int) string {
	return strconv.Itoa(status) + statusSeparator + http.StatusText(status)
}

// draw draws every page of a snapshot, at the address it is served at, beside every file of its build, on the day it
// is drawn: the résumé's career line runs to it until the store holds another snapshot.
func draw(snapshot *release.Snapshot) *drawing {
	facts, build := snapshot.Content, snapshot.Build
	site := routes.New(facts.Profile.Site.BaseURL)
	readPages := compose.Pages(facts)
	index := compose.Index(facts)
	frame := compose.FrameOf(facts)
	pageWriter := html.NewPageWriter(site, build, frame, time.Now())
	fixed := map[string]answer{
		routes.Path(index.Head.Page, routes.FormatMarkdown): drawnAs(plainMediaType, text.Markdown(index, site)),
	}
	for path, file := range build.Files() {
		fixed[path] = builtAs(file)
	}
	for _, icon := range html.Icons(facts.Profile.Person.Name) {
		fixed[icon.Path] = picturedAs(icon)
	}
	chosen := map[string]choice{}
	for _, composed := range readPages {
		// a text face writes the page whole at its own address and each tab alone at its own; the HTML face draws the
		// tab an address shows inside its page
		for _, shown := range slices.Concat([]document.Document{composed.Document}, compose.TabDocuments(composed.Document)) {
			page := shown.Head.Page
			texts := drawAndServeText(shown, fixed, site)
			browsed := drawnAs(htmlMediaType, pageWriter.Page(composed.Document, composed.Record, page.Tab)).knownAt(site.Address(page, routes.FormatChosen))
			chosen[routes.Path(page, routes.FormatChosen)] = choice{texts.manual, texts.markdown, browsed}
		}
	}
	moved := map[string]string{}
	for _, post := range compose.Posts(facts) {
		texts := drawAndServeText(post, fixed, site)
		page := post.Head.Page
		browsed := drawnAs(htmlMediaType, pageWriter.Post(post)).knownAt(site.Address(page, routes.FormatChosen))
		chosen[routes.Path(page, routes.FormatChosen)] = choice{texts.manual, texts.markdown, browsed}
		moved[routes.NumberPath(post.Head.Number)] = routes.Path(page, routes.FormatChosen)
	}
	status := statusLine(missingStatus)
	notFound := missingChoice{
		{mediaType: plainMediaType, draw: text.NewNotFound(site, status).For},
		{mediaType: htmlMediaType, draw: html.NewNotFound(site, status, facts.Words.Language(), build, frame).For},
	}
	return &drawing{fixed: fixed, chosen: chosen, moved: moved, notFound: notFound}
}

// drawAndServeText draws a page in the text formats, each naming the address the page is known by, and serves each at
// its own path among the fixed answers.
func drawAndServeText(shown document.Document, fixed map[string]answer, site routes.Site) textAnswers {
	page := shown.Head.Page
	known := site.Address(page, routes.FormatChosen)
	texts := textAnswers{
		markdown: drawnAs(markdownMediaType, text.Markdown(shown, site)).knownAt(known),
		manual:   drawnAs(plainMediaType, text.Manual(shown, site)).knownAt(known),
	}
	fixed[routes.Path(page, routes.FormatMarkdown)] = texts.markdown
	fixed[routes.Path(page, routes.FormatManual)] = texts.manual
	return texts
}
