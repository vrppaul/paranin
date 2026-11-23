// Package web serves the site over HTTP (docs/ARCHITECTURE.md, ADR-0003, ADR-0004). It draws every page of a release's
// snapshot once, the first time a request meets that snapshot (release.Derived), as the site stands at the request's
// moment, and serves the drawing until the store holds another or a post dated later comes out, so no request
// composes or draws a page; a draft it serves only where its author runs it to read them. Beside the pages it serves
// every file of the snapshot's build at its path, from memory, for a cache to keep for good, each picture of a post
// served, beside it, for a cache to ask for again, and each picture of the site, which the HTML face draws once and a
// cache asks for again. Each page, and each of its tabs alone, is served at the path routes writes for it in each
// format, and at its own address in the format the reader's request asks for: HTML for a browser, markdown for an
// agent that asks for it, and the manual for `curl` and everyone else; the index at llms.txt, sent as plain text, as
// a name ending in .txt promises. A post served is drawn whole in every format, and its number sends the reader on to
// its address; so is the blog's list, in each of its groupings. Each format of a page names in a Link header the
// address the page is known by, the one where the reader chooses, so a search engine counts its markdown and its
// manual as the page. A preview is served at its link alone, under a key the server is handed (preview.go), until the
// link lapses: every face and picture of it where its post would stand, the blog's folder swapped for the preview's
// folder and the link's token, drawn for the request, since each link writes its own token into them; once the post
// is out, the link sends its reader on to it. Whatever is answered in the preview's folder is for its reader alone:
// no search engine indexes it, no cache keeps it and no browser sends its address on from it. Any other path is
// answered with the page of no page, in text or in HTML, whichever the reader's request ranks higher: drawn for each
// request, since it quotes the path asked for, in the chosen format alone. Every answer, a refusal net/http writes
// among them, tells a browser not to guess its type and to load nothing but the site's own files, so markup slipped
// into a page runs no script but the site's own; a post's picture, to load and run nothing at all.
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
	"github.com/vrppaul/paranin/internal/content"
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

// What a browser may do with a post's picture opened by itself, a drawing above all, which could hold a script: load
// nothing and run nothing, as a document of no site; its own styles stay, which a drawing is drawn by
const pictureAlone = "default-src 'none'; style-src 'unsafe-inline'; sandbox"

// What every answer in the preview's folder is sent with, since its address may hold a link's token (docs/WEB.md): that
// no search engine indexes it, that no cache, a shared one least of all, keeps it, and that a browser sends its
// address on to nothing it loads or opens from it
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
// of its posts served; the path a post's number sends the reader on to, by the number's path; the page of an address
// of no page; and each post served at its preview link alone, which a request opening one draws.
type drawing struct {
	served   answers
	moved    map[string]string
	notFound missingChoice
	previews previewing
}

// answers are pages and files by the path each is served at: in each format at its own path, and each file, the same
// whatever the request asks; and each page in every format at the address where the reader chooses.
type answers struct {
	fixed  map[string]answer
	chosen map[string]choice
}

// previewing is what a preview's faces are drawn from as a request opens its link, which holds a token of its own, so
// each link's faces lead to its own address: each post served at its preview link alone, by its slug; the site's
// addresses, the build each page is drawn in, the frame round it, and the day it is drawn on.
type previewing struct {
	posts map[string]compose.Preview
	site  routes.Site
	build *assets.Build
	frame compose.Frame
	today time.Time
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
// empty where it is none a person reads; how long a cache keeps it; the policy a browser holds it to, empty where it
// is the site's own; where it sends the reader on to, empty where it sends them nowhere; and what it holds.
type answer struct {
	mediaType    string
	contentType  string
	link         string
	cacheControl string
	policy       string
	location     string
	body         []byte
}

// drawnAs is a page's text, drawn, as a media type: text of UTF-8, as every page is, which a cache asks for again.
func drawnAs(mediaType, text string) answer {
	return answer{
		mediaType: mediaType, contentType: mediaType + utf8Text, link: "", cacheControl: askEveryTime, policy: "", location: "",
		body: []byte(text),
	}
}

// builtAs is a file of the build as it is served, which a cache keeps for good: a changed file is built under
// another name.
func builtAs(file assets.File) answer {
	return answer{mediaType: "", contentType: file.ContentType, link: "", cacheControl: keptForGood, policy: "", location: "", body: file.Body}
}

// picturedAs is a picture of the site as it is served, which a cache asks for again: its path keeps its name when a
// new build of the server draws it anew.
func picturedAs(icon html.Icon) answer {
	return answer{mediaType: "", contentType: icon.MediaType, link: "", cacheControl: askEveryTime, policy: "", location: "", body: icon.Body}
}

// postPicturedAs is a post's picture as it is served, which a cache asks for again, since a reload may change it
// under the same name, and which a browser opening it by itself runs nothing in.
func postPicturedAs(picture content.Picture) answer {
	return answer{
		mediaType: "", contentType: picture.MediaType, link: "", cacheControl: askEveryTime, policy: pictureAlone, location: "", body: picture.Body,
	}
}

// sentOnTo is an answer that sends the reader on to a path, saying nothing of its own.
func sentOnTo(path string) answer {
	sent := drawnAs(plainMediaType, "")
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

// knownAt is a page naming the address it is known by, as a Link header writes it (RFC 8288, RFC 6596).
func (sent answer) knownAt(address string) answer {
	sent.link = linkOpening + address + canonicalRelation
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
	if location, isMoved := drawn.moved[path]; isMoved {
		// a post's folder, its slug, may be renamed, so a cache asks again where its number leads
		writer.Header().Set(cacheControl, askEveryTime)
		http.Redirect(writer, request, location, http.StatusMovedPermanently)
		return
	}
	status, sent := drawn.served.answerTo(writer.Header(), request, path, drawn.notFound)
	send(writer, status, sent)
}

// answerTo is the status and the answer to a request for a path among the answers: its file or page in the one format
// served there, the page in the format the request ranks first where the reader chooses, or else the page of no page,
// quoting the path the request asked for. Where the reader chooses, the header says the answer hangs on Accept.
func (served answers) answerTo(header http.Header, request *http.Request, path string, notFound missingChoice) (int, answer) {
	if found, isFixed := served.fixed[path]; isFixed {
		return http.StatusOK, found
	}
	header.Set(varyHeader, acceptHeader)
	accept := request.Header.Get(acceptHeader)
	if offered, isChosen := served.chosen[path]; isChosen {
		return http.StatusOK, rankedFirst(accept, offered)
	}
	missing := rankedFirst(accept, notFound)
	return missingStatus, drawnAs(missing.mediaType, missing.draw(request.URL.Path))
}

// atLink is the status and the answer to a request in the preview's folder at a moment: where its path is a preview
// link whose token the key signed for its slug and has not lapsed, the post's face or picture there while it is a
// preview, drawn for that link, or the reader sent on to the same at its own address once the post is out; else, a
// path of no post among them, as no page, so a link tells nothing of which posts there are.
func (server *Server) atLink(drawn *drawing, header http.Header, request *http.Request, moment time.Time) (int, answer) {
	none := answers{fixed: nil, chosen: nil}
	link, isLink := routes.PreviewPath(request.URL.Path)
	if !isLink || !server.previewKey.admits(link.Token, link.Slug, moment) {
		return none.answerTo(header, request, link.PostPath, drawn.notFound)
	}
	if previewed, isPreview := drawn.previews.posts[link.Slug]; isPreview {
		return drawn.previews.faceAt(previewed, link, header, request, drawn.notFound)
	}
	if drawn.served.holds(link.PostPath) {
		return sentOnForAWhile, sentOnTo(link.PostPath)
	}
	return none.answerTo(header, request, link.PostPath, drawn.notFound)
}

// holds is whether the answers hold a path, in one format or where the reader chooses.
func (served answers) holds(path string) bool {
	_, isFixed := served.fixed[path]
	_, isChosen := served.chosen[path]
	return isFixed || isChosen
}

// faceAt is the status and the answer at a preview's link: its face or picture at the link's path, drawn alone, leading
// to the others at that link, its markdown or its manual at its own path, the format the reader's request ranks first
// at the post's own address, or a picture; the page of no page where the post has none there.
func (previews previewing) faceAt(
	previewed compose.Preview, link routes.PreviewLink, header http.Header, request *http.Request, notFound missingChoice,
) (int, answer) {
	post := previewed.Post
	page := post.Head.Page
	site := previews.site.Previewing(page.Slug, link.Token)
	manual := lazyFormat{mediaType: plainMediaType, draw: func() answer { return manualOf(post, site) }}
	markdown := lazyFormat{mediaType: markdownMediaType, draw: func() answer { return markdownOf(post, site) }}
	writer := html.NewPageWriter(site, previews.build, previews.frame, previews.today)
	browsed := lazyFormat{mediaType: htmlMediaType, draw: func() answer { return drawnAs(htmlMediaType, writer.Preview(post)) }}
	switch link.PostPath {
	case routes.Path(page, routes.FormatManual):
		return http.StatusOK, manual.draw()
	case routes.Path(page, routes.FormatMarkdown):
		return http.StatusOK, markdown.draw()
	case routes.Path(page, routes.FormatChosen):
		header.Set(varyHeader, acceptHeader)
		return http.StatusOK, rankedFirst(request.Header.Get(acceptHeader), inTieOrder(manual, markdown, browsed)).draw()
	}
	for _, shown := range previewed.Pictures {
		if routes.FilePath(shown.File) == link.PostPath {
			return http.StatusOK, postPicturedAs(shown.Picture)
		}
	}
	return answers{fixed: nil, chosen: nil}.answerTo(header, request, link.PostPath, notFound)
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
// beside every file of its build and every picture of its posts served, and keeps each preview to draw for its link;
// and says until when it holds: until the next post dated later comes out, the résumé's career line running to the
// audience's day until then, or the store holds another snapshot.
func draw(snapshot *release.Snapshot, prepared *compose.Prepared, audience compose.Audience) (*drawing, time.Time) {
	facts, build := snapshot.Content, snapshot.Build
	addresses := routes.New(facts.Profile.Site.BaseURL)
	composition := compose.SiteOf(prepared, audience)
	frame := composition.Frame
	pageWriter := html.NewPageWriter(addresses, build, frame, audience.At)
	served := answers{
		fixed: map[string]answer{
			routes.Path(composition.Index.Head.Page, routes.FormatMarkdown): drawnAs(plainMediaType, text.Markdown(composition.Index, addresses)),
		},
		chosen: map[string]choice{},
	}
	for path, file := range build.Files() {
		served.fixed[path] = builtAs(file)
	}
	served.servePictures(composition.Pictures)
	for _, composed := range composition.Pages {
		// a text face writes the page whole at its own address and each tab alone at its own; the HTML face draws the
		// tab an address shows inside its page
		for _, shown := range slices.Concat([]document.Document{composed.Document}, composed.Tabs) {
			served.serveEveryFormat(shown, pageWriter.Page(composed.Document, composed.Record, shown.Head.Page.Tab), addresses)
		}
	}
	moved := map[string]string{}
	for _, post := range composition.Posts {
		served.serveEveryFormat(post, pageWriter.Post(post), addresses)
		moved[routes.NumberPath(post.Head.Number)] = routes.Path(post.Head.Page, routes.FormatChosen)
	}
	for _, list := range composition.Lists {
		served.serveEveryFormat(list, pageWriter.List(list), addresses)
	}
	status := statusLine(missingStatus)
	notFound := missingChoice{
		{mediaType: plainMediaType, draw: text.NewNotFound(addresses, status).For},
		{mediaType: htmlMediaType, draw: html.NewNotFound(addresses, status, facts.Words.Language(), build, frame).For},
	}
	previews := previewing{posts: map[string]compose.Preview{}, site: addresses, build: build, frame: frame, today: audience.At}
	for _, previewed := range composition.Previews {
		previews.posts[previewed.Post.Head.Page.Slug] = previewed
	}
	return &drawing{served: served, moved: moved, notFound: notFound, previews: previews}, composition.Until
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
func (served answers) serveEveryFormat(shown document.Document, browsed string, site routes.Site) {
	page := shown.Head.Page
	markdown, manual := markdownOf(shown, site), manualOf(shown, site)
	served.fixed[routes.Path(page, routes.FormatMarkdown)] = markdown
	served.fixed[routes.Path(page, routes.FormatManual)] = manual
	served.chosen[routes.Path(page, routes.FormatChosen)] = inTieOrder(
		manual, markdown, drawnAs(htmlMediaType, browsed).knownAt(site.Address(page, routes.FormatChosen)),
	)
}

// markdownOf is a page in markdown, as the site writes its addresses, naming the address the page is known by.
func markdownOf(shown document.Document, site routes.Site) answer {
	return drawnAs(markdownMediaType, text.Markdown(shown, site)).knownAt(site.Address(shown.Head.Page, routes.FormatChosen))
}

// manualOf is a page as a manual, as the site writes its addresses, naming the address the page is known by.
func manualOf(shown document.Document, site routes.Site) answer {
	return drawnAs(plainMediaType, text.Manual(shown, site)).knownAt(site.Address(shown.Head.Page, routes.FormatChosen))
}
