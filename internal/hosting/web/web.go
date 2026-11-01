// Package web serves the site over HTTP (docs/ARCHITECTURE.md, ADR-0003). It draws every page of a snapshot
// once, the first time a request meets that snapshot, and serves the drawing until the store holds another, so
// no request composes or draws a page. Each page is served at the path routes writes for it; any other path is
// not found.
package web

import (
	"net/http"
	"sync/atomic"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/face/text"
	"github.com/vrppaul/paranin/internal/routes"
)

// The headers every page is sent with: what it is, that it is no other kind, and that a cache asks again, since
// a reload can change any page
const (
	markdownType = "text/markdown; charset=utf-8"
	noSniffing   = "nosniff"
	askEveryTime = "no-cache"
	contentType  = "Content-Type"
	typeOptions  = "X-Content-Type-Options"
	cacheControl = "Cache-Control"
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

// drawing is every page of one snapshot, drawn, by the path it is served at.
type drawing struct {
	snapshot *content.Snapshot
	pages    map[string]drawnPage
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
	found, isPage := server.current().pages[request.URL.Path]
	if !isPage {
		http.NotFound(writer, request)
		return
	}
	header := writer.Header()
	header.Set(contentType, found.contentType)
	header.Set(typeOptions, noSniffing)
	header.Set(cacheControl, askEveryTime)
	// A write fails only when the reader has gone, and nobody is left to tell
	_, _ = writer.Write(found.body)
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
	for _, composed := range documents {
		path := routes.Path(composed.Head.Page, routes.FormatMarkdown)
		pages[path] = drawnPage{contentType: markdownType, body: []byte(text.Markdown(composed, site))}
	}
	return &drawing{snapshot: snapshot, pages: pages}
}
