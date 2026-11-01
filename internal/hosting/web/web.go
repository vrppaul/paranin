// Package web serves the site over HTTP (docs/ARCHITECTURE.md, ADR-0003). It draws every page of a snapshot
// once, the first time a request meets that snapshot, and serves the drawing until the store holds another, so
// no request composes or draws a page. Each page is served at the path routes writes for it in each format, and
// at its own address in the format the reader's request asks for; any other path is not found.
package web

import (
	"mime"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/face/text"
	"github.com/vrppaul/paranin/internal/routes"
)

// The headers every page is sent with: what it is, that it is no other kind, and that a cache asks again, since
// a reload can change any page
const (
	markdownType = markdownMediaType + utf8Text
	manualType   = manualMediaType + utf8Text
	utf8Text     = "; charset=utf-8"
	noSniffing   = "nosniff"
	askEveryTime = "no-cache"
	contentType  = "Content-Type"
	typeOptions  = "X-Content-Type-Options"
	cacheControl = "Cache-Control"
)

// How a reader's request chooses a format: by the media types its Accept header ranks. A cache keeps one copy
// for each Accept header it meets
const (
	acceptHeader      = "Accept"
	varyHeader        = "Vary"
	markdownMediaType = "text/markdown"
	manualMediaType   = "text/plain"
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

// Server serves the pages of the store's current snapshot.
type Server struct {
	store *content.Store
	// drawn is the drawing of the last snapshot a request met
	drawn   atomic.Pointer[drawing]
	handler http.Handler
}

// drawing is every page of one snapshot, drawn, by the path it is served at: in each format at its own path, and
// both ways at the address where the reader chooses.
type drawing struct {
	snapshot *content.Snapshot
	pages    map[string]drawnPage
	chosen   map[string]choice
}

// choice is one page in each format a reader may choose.
type choice struct {
	markdown drawnPage
	manual   drawnPage
}

// drawnPage is one page as it is sent.
type drawnPage struct {
	contentType string
	body        []byte
}

// New is a server of the site the store holds.
func New(store *content.Store) *Server {
	mux := http.NewServeMux()
	server := &Server{store: store, drawn: atomic.Pointer[drawing]{}, handler: mux}
	mux.HandleFunc(pagePattern, server.servePage)
	return server
}

// ServeHTTP serves a request: a page read, or a refusal.
func (server *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	server.handler.ServeHTTP(writer, request)
}

func (server *Server) servePage(writer http.ResponseWriter, request *http.Request) {
	drawn := server.current()
	if found, isPage := drawn.pages[request.URL.Path]; isPage {
		send(writer, found)
		return
	}
	if offered, isChosen := drawn.chosen[request.URL.Path]; isChosen {
		writer.Header().Set(varyHeader, acceptHeader)
		send(writer, offered.chosenBy(request.Header.Get(acceptHeader)))
		return
	}
	http.NotFound(writer, request)
}

// chosenBy is the page in the format an Accept header ranks higher: markdown where the header ranks it above
// plain text, the manual otherwise, `curl` and a header naming neither among them.
func (offered choice) chosenBy(accept string) drawnPage {
	if quality(accept, markdownMediaType) > quality(accept, manualMediaType) {
		return offered.markdown
	}
	return offered.manual
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

func send(writer http.ResponseWriter, page drawnPage) {
	header := writer.Header()
	header.Set(contentType, page.contentType)
	header.Set(typeOptions, noSniffing)
	header.Set(cacheControl, askEveryTime)
	// A write fails only when the reader has gone, and nobody is left to tell
	_, _ = writer.Write(page.body)
}

// current is the drawing of the store's current snapshot. Two requests that meet a new snapshot at once may each
// draw it, and a drawing is kept only over the one its request found. The drawing is read before the snapshot,
// and a store's snapshots only move forward, so the snapshot a request draws is never older than the drawing it
// replaces.
func (server *Server) current() *drawing {
	found := server.drawn.Load()
	snapshot := server.store.Current()
	if found != nil && found.snapshot == snapshot {
		return found
	}
	drawn := draw(snapshot)
	server.drawn.CompareAndSwap(found, drawn)
	return drawn
}

// draw draws every page of a snapshot, at the address it is served at.
func draw(snapshot *content.Snapshot) *drawing {
	site := routes.New(snapshot.Profile.Site.BaseURL)
	documents := compose.Documents(snapshot)
	pages := make(map[string]drawnPage, len(documents))
	chosen := make(map[string]choice, len(documents))
	for _, composed := range documents {
		page := composed.Head.Page
		offered := choice{
			markdown: drawnPage{contentType: markdownType, body: []byte(text.Markdown(composed, site))},
			manual:   drawnPage{contentType: manualType, body: []byte(text.Manual(composed, site))},
		}
		pages[routes.Path(page, routes.FormatMarkdown)] = offered.markdown
		pages[routes.Path(page, routes.FormatManual)] = offered.manual
		chosen[routes.Path(page, routes.FormatChosen)] = offered
	}
	return &drawing{snapshot: snapshot, pages: pages, chosen: chosen}
}
