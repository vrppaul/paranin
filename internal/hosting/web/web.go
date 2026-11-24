// Package web serves the site over HTTP (docs/ARCHITECTURE.md, ADR-0003, ADR-0004). It draws every page of a release's
// snapshot once, the first time a request meets that snapshot (release.Derived), as the site stands at the request's
// moment, and serves the drawing until the store holds another or a post dated later comes out, so no request
// composes or draws a page; a draft it serves only where its author runs it to read them. Beside the pages it serves
// every file of the snapshot's build at its path, from memory, for a cache to keep for good, each picture of a post
// served, beside it, for a cache to ask for again, and each picture of the site, which the HTML face draws once and a
// cache asks for again. Each page, and each of its tabs alone, is served at the path routes writes for it in each
// format, and at its own address in the format the reader's request asks for: HTML for a browser, markdown for an
// agent that asks for it, and the manual for `curl` and everyone else; the index at llms.txt, sent as plain text, as
// a name ending in .txt promises. A post served is drawn whole in every format, each of its older versions at its
// revision's address, and its number sends the reader on to its address; so is the blog's list, in each of its
// groupings and filtered by each tag and series, at the address the request's query names (routes.Requested). The
// blog's feeds, its own and each series', are drawn with the pages, in Atom, each post's body as the HTML face draws it
// for a feed reader, and served at the path routes writes for each, for a reader to poll and ask again. Each
// format of a page names in a Link header the address the page is known by, the one where the reader chooses, so a
// search engine counts its markdown and its manual as the page; a filtered list and an older version name none, and
// tell a search engine not to index them. A preview is served at its link alone, under a key
// the server is handed (preview.go), until the link lapses: every face and picture of it where its post would stand,
// the blog's folder swapped for the preview's folder and the link's token, drawn for the request, since each link
// writes its own token into them; once the post is out, the link sends its reader on to it. Whatever is answered in
// the preview's folder is for its reader alone: no search engine indexes it, no cache keeps it and no browser sends its
// address on from it. Any other address is answered with the page of no page, in text or in HTML, whichever the
// reader's request ranks higher: drawn for each request, since it quotes the address asked for, in the chosen format
// alone. Every answer, a refusal net/http writes among them, tells a browser not to guess its type and to load nothing
// but the site's own files, so markup slipped into a page runs no script but the site's own; a post's picture, to load
// and run nothing at all.
package web

import (
	"mime"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vrppaul/paranin/internal/assets"
	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/face/feed"
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

// What a browser may do with a post's picture opened by itself, a drawing above all, which could hold a script: load
// nothing and run nothing, as a document of no site; its own styles stay, which a drawing is drawn by
const pictureAlone = "default-src 'none'; style-src 'unsafe-inline'; sandbox"

// What every answer in the preview's folder is sent with, since its address may hold a link's token (docs/WEB.md): that
// no search engine indexes it, as none indexes a filtered list either, that no cache, a shared one least of all, keeps
// it, and that a browser sends its address on to nothing it loads or opens from it
const (
	robotsHeader   = "X-Robots-Tag"
	noIndex        = "noindex"
	keptByNone     = "private, no-store"
	referrerHeader = "Referrer-Policy"
	noReferrer     = "no-referrer"
)

// How a post's preview link sends its reader on once the post is out: to the same face at its own address, for a
// while alone, since a browser keeps a permanent move even were the post taken back to a preview
const (
	locationHeader  = "Location"
	sentOnForAWhile = http.StatusFound
)

// How the header a page names its address in is written: <address>; rel="canonical"
const (
	linkHeader        = "Link"
	linkOpening       = "<"
	canonicalRelation = `>; rel="canonical"`
)

// How a reader's request chooses a format: by the media types its Accept header ranks, which routes names. A cache
// keeps one copy for each Accept header it meets
const (
	acceptHeader     = "Accept"
	varyHeader       = "Vary"
	mediaTypeEnd     = ","
	qualityParameter = "q"
	qualityBits      = 64
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

// Server serves the pages and the built files of the store's current snapshot, and the pictures of the site.
type Server struct {
	// drawn is every page of the store's current snapshot, drawn once for each snapshot and again as a post dated
	// later comes out
	drawn *release.Derived[*drawing]
	// pictures are the pictures of the site, drawn once, since no snapshot changes them
	pictures map[string]answer
	// previewKey signs the preview links the server opens, none where it opens none
	previewKey *PreviewKey
	handler    http.Handler
}

// drawing is every page of one snapshot, drawn, by the path it is served at, every file of its build and every picture
// of its posts served; where a path sends the reader on to, by the path: a post's number, and its current version's
// day; the page of an address of no page in each language; each post served at its preview link alone, which a request
// opening one draws; and the site's addresses, which read a request's path.
type drawing struct {
	served answers
	moved  map[string]move
	// notFound are the page of no page in each language, by the language a page is translated into, which names its
	// folder (compose.Frame.Translation)
	notFound map[string]missingChoice
	previews previewing
	site     routes.Site
}

// notFoundAt is the page of no page for a path asked for: in the language whose folder opens it, the site's own where
// no language's does.
func (drawn *drawing) notFoundAt(path string) missingChoice {
	translation, _ := drawn.site.LanguageFolderOf(path)
	return drawn.notFound[translation]
}

// move is where a path sends its reader on to, and for how long: a post's number for good, as its slug is the post's
// address, which a cache asks again for, since a folder may be renamed; its current version's day for a while, since
// that day names an older version once the post is updated.
type move struct {
	location string
	status   int
}

// answers are pages and files by the path each is served at: in each format at its own path, and each file, the same
// whatever the request asks; and each page in every format at the address where the reader chooses.
type answers struct {
	fixed  map[string]answer
	chosen map[string]choice
	// fixedLater and chosenLater are the pages drawn only once a request first asks for them, as fixed and chosen hold
	// the others (laterPage)
	fixedLater  map[string]lazyFormat
	chosenLater map[string][]lazyFormat
}

// answersOf are answers holding nothing yet.
func answersOf() answers {
	return answers{fixed: map[string]answer{}, chosen: map[string]choice{}, fixedLater: map[string]lazyFormat{}, chosenLater: map[string][]lazyFormat{}}
}

// previewing is what a preview's faces are drawn from as a request opens its link, which holds a token of its own, so
// each link's faces lead to its own address: each post served at its preview link alone, by its page; the build each
// page is drawn in, and the day it is drawn on.
type previewing struct {
	posts map[document.Page]framedPreview
	build *assets.Build
	today time.Time
}

// framedPreview is a post served at its preview link, and the frame round it, in the post's language.
type framedPreview struct {
	preview compose.Preview
	frame   compose.Frame
}

// choice is one page in each format a reader may choose, never none, in the order a tie between two goes (inTieOrder).
type choice []answer

// inTieOrder are a page's formats a reader may choose, in the order a tie between two goes: the manual, then
// markdown, then HTML, so `curl` and a header naming none of them get the manual.
func inTieOrder[Format any](manual, markdown, browsed Format) []Format {
	return []Format{manual, markdown, browsed}
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

// answer is a page or a file as it is sent: the media type a page is, which an Accept header ranks, empty for a file,
// which no reader chooses; the header that says what it is; the Link header naming the address a page is known by,
// empty where it is none a person reads; what it tells a search engine, empty where it leaves it free to index it; how
// long a cache keeps it; the policy a browser holds it to, empty where it is the site's own; where it sends the reader
// on to, empty where it sends them nowhere; and what it holds.
type answer struct {
	mediaType    string
	contentType  string
	link         string
	robots       string
	cacheControl string
	policy       string
	location     string
	body         []byte
}

// drawnAs is a page's text, drawn, as a media type: text of UTF-8, as every page is, which a cache asks for again.
func drawnAs(mediaType, text string) answer {
	return answer{
		mediaType: mediaType, contentType: mediaType + utf8Text, link: "", robots: "", cacheControl: askEveryTime, policy: "", location: "",
		body: []byte(text),
	}
}

// builtAs is a file of the build as it is served, which a cache keeps for good: a changed file is built under
// another name.
func builtAs(file assets.File) answer {
	return answer{
		mediaType: "", contentType: file.ContentType, link: "", robots: "", cacheControl: keptForGood, policy: "", location: "", body: file.Body,
	}
}

// picturedAs is a picture of the site as it is served, which a cache asks for again: its path keeps its name when a
// new build of the server draws it anew.
func picturedAs(icon html.Icon) answer {
	return answer{
		mediaType: "", contentType: icon.MediaType, link: "", robots: "", cacheControl: askEveryTime, policy: "", location: "", body: icon.Body,
	}
}

// postPicturedAs is a post's picture as it is served, which a cache asks for again, since a reload may change it
// under the same name, and which a browser opening it by itself runs nothing in.
func postPicturedAs(picture content.Picture) answer {
	return answer{
		mediaType: "", contentType: picture.MediaType, link: "", robots: "", cacheControl: askEveryTime, policy: pictureAlone, location: "",
		body: picture.Body,
	}
}

// sentOnTo is an answer that sends the reader on to a path, saying nothing of its own.
func sentOnTo(path string) answer {
	sent := drawnAs(routes.PlainType, "")
	sent.location = path
	return sent
}

// privately is an answer as it is sent in the preview's folder, whose address may hold a link's token: kept by no
// cache, as ServeHTTP has said already of whatever is answered there, and known at no address, since the post's own
// serves nothing yet.
func (sent answer) privately() answer {
	sent.cacheControl, sent.link = keptByNone, ""
	return sent
}

// picturesOf are the pictures of the site, by the path each is served at.
func picturesOf(icons []html.Icon) map[string]answer {
	pictures := make(map[string]answer, len(icons))
	for _, icon := range icons {
		pictures[icon.Path] = picturedAs(icon)
	}
	return pictures
}

// knownAs is a page's answer naming the address the page is known by, as a Link header writes it (RFC 8288, RFC 6596);
// a page no search engine indexes, a filtered list or an older version of a post, names none and tells a search engine
// not to index it (document.Page's Indexed, docs/FOUNDATION.md §5.4).
func (sent answer) knownAs(page document.Page, site routes.Site) answer {
	if !page.Indexed() {
		sent.robots = noIndex
		return sent
	}
	sent.link = linkOpening + site.Address(page, routes.FormatChosen) + canonicalRelation
	return sent
}

// New is a server of the site the store holds, serving its drafts too where drafts says, to their author alone, and
// each preview at the link a key signs, none where it is handed no key.
func New(store *release.Store, drafts bool, previewKey *PreviewKey) *Server {
	mux := http.NewServeMux()
	prepared := release.Derive(store, func(snapshot *release.Snapshot, _ time.Time) (*compose.Prepared, time.Time) {
		// the zero time: it lapses with its snapshot alone (release.Work)
		return compose.Prepare(snapshot.Content), time.Time{}
	})
	drawn := release.Derive(store, func(snapshot *release.Snapshot, moment time.Time) (*drawing, time.Time) {
		return draw(snapshot, prepared.Of(snapshot, moment), compose.Audience{At: moment, Drafts: drafts})
	})
	server := &Server{drawn: drawn, pictures: picturesOf(html.Icons()), previewKey: previewKey, handler: mux}
	mux.HandleFunc(pagePattern, server.servePage)
	return server
}

// ServeHTTP serves a request: a page or a file read, or a refusal. Every answer passes here, a refusal net/http
// writes itself among them, so here it is told what a browser may do with it, and, in the preview's folder, whose
// paths may hold a link's token, that no search engine indexes it, no cache keeps it and no browser sends its address
// on: a redirect to a path's clean form and a refusal of a method carry it too.
func (server *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	header := writer.Header()
	header.Set(typeOptions, noSniffing)
	header.Set(securityPolicy, ownSiteAlone)
	if routes.InPreviewFolder(request.URL.Path) {
		header.Set(robotsHeader, noIndex)
		header.Set(referrerHeader, noReferrer)
		header.Set(cacheControl, keptByNone)
	}
	server.handler.ServeHTTP(writer, request)
}

func (server *Server) servePage(writer http.ResponseWriter, request *http.Request) {
	path := request.URL.Path
	if picture, isPicture := server.pictures[path]; isPicture {
		send(writer, http.StatusOK, picture)
		return
	}
	moment := time.Now()
	drawn := server.drawn.Current(moment)
	if routes.InPreviewFolder(path) {
		status, sent := server.atLink(drawn, writer.Header(), request, moment)
		send(writer, status, sent.privately())
		return
	}
	if moved, isMoved := drawn.moved[path]; isMoved {
		writer.Header().Set(cacheControl, askEveryTime)
		http.Redirect(writer, request, moved.location, moved.status)
		return
	}
	status, sent := drawn.served.answerTo(writer.Header(), request, drawn.site.Requested(path, request.URL.Query()), drawn.notFoundAt(path))
	send(writer, status, sent)
}

// answerTo is the status and the answer to a request for a path among the answers: its file or page in the one format
// served there, the page in the format the request ranks first where the reader chooses, each drawn first where it is
// drawn only once asked for, or else the page of no page, quoting the address the request asked for, as it was written.
// Where the reader chooses, the header says the answer hangs on Accept.
func (served answers) answerTo(header http.Header, request *http.Request, path string, notFound missingChoice) (int, answer) {
	if found, isFixed := served.fixed[path]; isFixed {
		return http.StatusOK, found
	}
	if waiting, isLater := served.fixedLater[path]; isLater {
		return http.StatusOK, waiting.draw()
	}
	header.Set(varyHeader, acceptHeader)
	accept := request.Header.Get(acceptHeader)
	if offered, isChosen := served.chosen[path]; isChosen {
		return http.StatusOK, rankedFirst(accept, offered)
	}
	if offered, isLater := served.chosenLater[path]; isLater {
		return http.StatusOK, rankedFirst(accept, offered).draw()
	}
	missing := rankedFirst(accept, notFound)
	return missingStatus, drawnAs(missing.mediaType, missing.draw(routes.Asked(request.URL.Path, request.URL.RawQuery)))
}

// atLink is the status and the answer to a request in a preview's folder at a moment: where its path is a preview link
// whose token the key signed for its post and has not lapsed, the post's face or picture there while it is a preview,
// drawn for that link, or the reader sent on to the same at its own address once the post is out; else, a path of no
// post among them, as no page, so a link tells nothing of which posts there are.
func (server *Server) atLink(drawn *drawing, header http.Header, request *http.Request, moment time.Time) (int, answer) {
	link, isLink := drawn.site.PreviewPath(request.URL.Path)
	notFound := drawn.notFoundAt(request.URL.Path)
	if !isLink || !server.previewKey.admits(link.Token, link.Post, moment) {
		return answersOf().answerTo(header, request, link.PostPath, notFound)
	}
	if previewed, isPreview := drawn.previews.posts[link.Post]; isPreview {
		return drawn.previews.at(previewed, drawn.site.Previewing(link.Post, link.Token)).answerTo(header, request, link.PostPath, notFound)
	}
	if drawn.served.holds(link.PostPath) {
		return sentOnForAWhile, sentOnTo(link.PostPath)
	}
	return answersOf().answerTo(header, request, link.PostPath, notFound)
}

// holds is whether the answers hold a path, in one format or where the reader chooses, drawn or still to draw.
func (served answers) holds(path string) bool {
	_, isFixed := served.fixed[path]
	_, isChosen := served.chosen[path]
	_, isFixedLater := served.fixedLater[path]
	_, isChosenLater := served.chosenLater[path]
	return isFixed || isChosen || isFixedLater || isChosenLater
}

// at are a preview's answers as a site read at its link writes them, as its post would be served in its blog's folder:
// its faces, each drawn once asked for, leading to the others at that link, and its pictures.
func (previews previewing) at(previewed framedPreview, site routes.Site) answers {
	post := previewed.preview.Post
	writer := html.NewPageWriter(site, previews.build, previewed.frame, previews.today)
	served := answersOf()
	served.serveLater(laterPageOf(post, textWriter{site: site, manual: previewed.frame.Manual}, func() string { return writer.Preview(post) }))
	served.servePictures(previewed.preview.Pictures)
	return served
}

// lazyFormat is a format a reader may choose, drawn only once chosen: its media type, which an Accept header ranks, and
// how it is drawn.
type lazyFormat struct {
	mediaType string
	draw      func() answer
}

func (format lazyFormat) servedAs() string {
	return format.mediaType
}

// laterPage is a page in each format a reader may choose, its manual, its markdown and its HTML, each drawn the first
// time a request asks for it and kept for those after, once however many ask at the same moment: a filtered list,
// of which there is one for each tag and series, and an older version of a post, one for each revision, few of either
// read; a preview, whose faces each link draws anew.
type laterPage struct {
	page     document.Page
	manual   lazyFormat
	markdown lazyFormat
	browsed  lazyFormat
}

// laterPageOf is a page drawn when asked for: its manual and its markdown as the text faces write them, and its HTML
// as browse draws it, each naming the address it is known by.
func laterPageOf(shown document.Document, writer textWriter, browse func() string) laterPage {
	page := shown.Head.Page
	return laterPage{
		page:     page,
		manual:   lazyFormat{mediaType: routes.PlainType, draw: sync.OnceValue(func() answer { return writer.manualOf(shown) })},
		markdown: lazyFormat{mediaType: routes.MarkdownType, draw: sync.OnceValue(func() answer { return writer.markdownOf(shown) })},
		browsed: lazyFormat{
			mediaType: routes.PageType, draw: sync.OnceValue(func() answer { return drawnAs(routes.PageType, browse()).knownAs(page, writer.site) }),
		},
	}
}

// serveLater serves a page drawn when asked for, as serveEveryFormat serves one drawn already: each text format at its
// own path, and every format at the address where the reader chooses.
func (served answers) serveLater(later laterPage) {
	served.fixedLater[routes.Path(later.page, routes.FormatMarkdown)] = later.markdown
	served.fixedLater[routes.Path(later.page, routes.FormatManual)] = later.manual
	served.chosenLater[routes.Path(later.page, routes.FormatChosen)] = inTieOrder(later.manual, later.markdown, later.browsed)
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
	if sent.robots != "" {
		header.Set(robotsHeader, sent.robots)
	}
	if sent.policy != "" {
		header.Set(securityPolicy, sent.policy)
	}
	if sent.location != "" {
		header.Set(locationHeader, sent.location)
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

// draw draws every page of a snapshot, as Prepare read it, as an audience reads it, at the address it is served at,
// beside every file of its build, every picture of its posts served and each feed, and keeps each preview to draw for
// its link; and says until when it holds: until the next post dated later comes out, the résumé's career line running
// to the audience's day until then, or the store holds another snapshot.
func draw(snapshot *release.Snapshot, prepared *compose.Prepared, audience compose.Audience) (*drawing, time.Time) {
	facts, build := snapshot.Content, snapshot.Build
	addresses := routes.New(facts.Profile.Site.BaseURL, facts.Profile.Site.TagAuthority, facts.TranslatedInto())
	composition := compose.SiteOf(prepared, audience)
	frame := composition.Own.Frame
	pageWriter := html.NewPageWriter(addresses, build, frame, audience.At)
	served := answersOf()
	served.fixed[routes.Path(composition.Index.Head.Page, routes.FormatMarkdown)] = drawnAs(routes.PlainType, text.Markdown(composition.Index, addresses))
	for path, file := range build.Files() {
		served.fixed[path] = builtAs(file)
	}
	for _, composed := range composition.Pages {
		// a text face writes the page whole at its own address and each tab alone at its own; the HTML face draws the
		// tab an address shows inside its page
		for _, shown := range slices.Concat([]document.Document{composed.Document}, composed.Tabs) {
			served.serveEveryFormat(shown, pageWriter.Page(composed.Document, composed.Record, shown.Head.Page.Tab), textWriter{site: addresses, manual: frame.Manual})
		}
	}
	drawn := blogDrawing{served: served, moved: map[string]move{}, previews: previewing{
		posts: map[document.Page]framedPreview{}, build: build, today: audience.At,
	}}
	missing, notFound := missingDrawing{site: addresses, status: statusLine(missingStatus), build: build}, map[string]missingChoice{}
	for _, blog := range composition.Blogs() {
		drawn.draw(blog, html.NewPageWriter(addresses, build, blog.Frame, audience.At), textWriter{site: addresses, manual: blog.Frame.Manual})
		notFound[blog.Frame.Translation] = missing.choice(blog.Frame)
	}
	return &drawing{served: served, moved: drawn.moved, notFound: notFound, previews: drawn.previews, site: addresses}, composition.Until
}

// missingDrawing is what the page of no page is drawn from in any language: the site, the status the host answers
// with, and the build it is drawn in.
type missingDrawing struct {
	site   routes.Site
	status string
	build  *assets.Build
}

// choice is the page of no page in a frame's language: as text, which no language changes, and as HTML in the frame.
func (missing missingDrawing) choice(frame compose.Frame) missingChoice {
	return missingChoice{
		{mediaType: routes.PlainType, draw: text.NewNotFound(missing.site, missing.status).For},
		{mediaType: routes.PageType, draw: html.NewNotFound(missing.site, missing.status, missing.build, frame).For},
	}
}

// blogDrawing is every blog as draw serves it: the answers it adds its pages to, where its paths send the reader on to,
// and its previews, to draw for their links.
type blogDrawing struct {
	served   answers
	moved    map[string]move
	previews previewing
}

// draw serves a blog in its language's frame: each post, its number sending the reader on to it, and its pictures; the
// lists, the filtered ones and the older versions drawn when first asked for; the feeds; and keeps each preview.
func (drawn blogDrawing) draw(blog compose.Blog, pageWriter html.PageWriter, writer textWriter) {
	served := drawn.served
	served.servePictures(blog.Pictures)
	for _, post := range blog.Posts {
		page := post.Head.Page
		served.serveEveryFormat(post, pageWriter.Post(post), writer)
		drawn.moved[routes.NumberPath(page.Translation, post.Head.Number)] = move{location: routes.Path(page, routes.FormatChosen), status: http.StatusMovedPermanently}
		currentDayMoves(drawn.moved, post)
	}
	for _, list := range blog.Lists {
		served.serveEveryFormat(list, pageWriter.List(list), writer)
	}
	// a filtered list and an older version of a post are drawn as a request first asks for them: there is one for each
	// tag and series and each revision, and few are read
	for _, list := range blog.Filtered {
		served.serveLater(laterPageOf(list, writer, func() string { return pageWriter.List(list) }))
	}
	for _, older := range blog.Versions {
		served.serveLater(laterPageOf(older, writer, func() string { return pageWriter.Post(older) }))
	}
	bodies := html.NewFeedWriter(writer.site)
	for _, composedFeed := range blog.Feeds {
		served.fixed[routes.FeedPath(composedFeed.Target)] = drawnAs(routes.FeedType, feed.Atom(composedFeed, writer.site, bodies))
	}
	for _, previewed := range blog.Previews {
		drawn.previews.posts[previewed.Post.Head.Page] = framedPreview{preview: previewed, frame: blog.Frame}
	}
}

// currentDayMoves sends the reader of a post's current version by its day, the name its versions list it by, on to the
// same face at the post's own address, for a while.
func currentDayMoves(moved map[string]move, post document.Document) {
	page := post.Head.Page
	for _, version := range document.VersionsOf(post.Head.Version) {
		if version.Current {
			for _, format := range routes.Formats() {
				moved[routes.Path(page.OlderVersion(version.Name), format)] = move{location: routes.Path(page, format), status: sentOnForAWhile}
			}
		}
	}
}

// servePictures serves posts' pictures, each at its file beside its post.
func (served answers) servePictures(pictures []compose.PostPicture) {
	for _, shown := range pictures {
		served.fixed[routes.FilePath(shown.File)] = postPicturedAs(shown.Picture)
	}
}

// serveEveryFormat draws a page in the text formats and serves each at its own path among the fixed answers, and serves
// them with the page's HTML, drawn already, at the address where the reader chooses; each names the address the page
// is known by, as the site writes its addresses.
func (served answers) serveEveryFormat(shown document.Document, browsed string, writer textWriter) {
	page := shown.Head.Page
	markdown, manual := writer.markdownOf(shown), writer.manualOf(shown)
	served.fixed[routes.Path(page, routes.FormatMarkdown)] = markdown
	served.fixed[routes.Path(page, routes.FormatManual)] = manual
	served.chosen[routes.Path(page, routes.FormatChosen)] = inTieOrder(manual, markdown, drawnAs(routes.PageType, browsed).knownAs(page, writer.site))
}

// textWriter is what the text faces write a page with: the site, which writes its addresses, and the manual's own
// words in the page's language.
type textWriter struct {
	site   routes.Site
	manual compose.ManualWords
}

// markdownOf is a page in markdown, as the site writes its addresses, naming the address the page is known by.
func (writer textWriter) markdownOf(shown document.Document) answer {
	return drawnAs(routes.MarkdownType, text.Markdown(shown, writer.site)).knownAs(shown.Head.Page, writer.site)
}

// manualOf is a page as a manual, as the site writes its addresses, naming the address the page is known by.
func (writer textWriter) manualOf(shown document.Document) answer {
	return drawnAs(routes.PlainType, text.Manual(shown, writer.site, writer.manual)).knownAs(shown.Head.Page, writer.site)
}
