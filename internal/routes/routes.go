// Package routes writes every address of the site, from a document's link target and a format, or a picture of the
// site, and a post's at its preview link, and reads a request's query into the address of the filtered list it names,
// so no face and no host writes a path of its own (docs/ARCHITECTURE.md, ADR-0003). It reads document alone.
package routes

import (
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/vrppaul/paranin/internal/document"
)

// Format is the form a page is served in.
type Format int

const (
	// The zero format names none, and has no path
	_ Format = iota
	// FormatMarkdown is CommonMark, for agents
	FormatMarkdown
	// FormatManual is a manual page laid out as text, for a terminal
	FormatManual
	// FormatChosen is the page at the address where the reader's request chooses its format
	FormatChosen
)

// Formats are the formats a page is served in, each at an address of its own.
func Formats() []Format {
	return []Format{FormatChosen, FormatMarkdown, FormatManual}
}

// Picture is a picture of the site a browser shows beside its pages.
type Picture int

const (
	// The zero picture names none, and has no path
	_ Picture = iota
	// PictureDrawing is the site's icon drawn to any size
	PictureDrawing
	// PictureHomeScreen is the site's icon a phone keeps on its home screen, where a phone asks for it by name
	PictureHomeScreen
	// PictureAskedByName is the site's icon where a browser asks for it by name, on a page whose head links none
	PictureAskedByName
)

// Where the site's pictures are served: each at the path a browser or a phone asks for it by, where one does
const (
	drawingFile     = "/favicon.svg"
	homeScreenFile  = "/apple-touch-icon.png"
	askedByNameFile = "/favicon.ico"
)

// How the site's paths are written: the résumé at the site's root and in its files, a project's folder, the blog's
// list, by year at the blog's own name and by series in its folder, where a post stands by its slug and by its number,
// the slash before a tab's name in its page's folder, the index where agents look for it (llmstxt.org), and the endings
// of the formats
const (
	resumeRoot     = "/"
	resumeFile     = "/index"
	indexFile      = "/llms.txt"
	projectFolder  = "/p/"
	blogList       = "/blog"
	blogFolder     = blogList + folderEnd
	seriesList     = blogFolder + "series"
	folderEnd      = "/"
	markdownEnding = ".md"
	manualEnding   = ".txt"
	// the folder a post's preview link opens it in, the link's token after it (docs/FOUNDATION.md §5.4, Status)
	previewFolder = "/preview/"
)

// The program a terminal reads the site with, and what parts it from the address in a command
const (
	readingCommand = "curl"
	commandSpace   = " "
)

// The scheme of an address that writes to a mailbox
const mailScheme = "mailto:"

// Site writes the site's addresses whole, from the address it is served at, and knows the site's own name; read at a
// post's preview link, it writes that post's addresses under the link's token.
type Site struct {
	base string
	host string
	// previewed is the post read at its preview link, none where the site is read at its own addresses alone
	previewed *preview
}

// preview is a post read at its preview link: its slug, and the token the link holds.
type preview struct {
	slug  string
	token string
}

// BaseAddress is the address a site is served at, as New's caller hands it: as it is written, ending in no slash,
// and its host. Content's web address is one, read once from the profile and held by its check to the web.
type BaseAddress interface {
	Written() string
	Host() string
}

// New is the site served at a base address, and named by the address's host.
func New(base BaseAddress) Site {
	return Site{base: base.Written(), host: base.Host(), previewed: nil}
}

// Previewing is the site as the reader of a post's preview link reads it (docs/FOUNDATION.md §5.4, Status): every
// address of that post, its page in each format, its sections and its pictures, in the preview's folder under the
// link's token, since its own serve nothing; every other page at its own.
func (site Site) Previewing(slug, token string) Site {
	site.previewed = &preview{slug: slug, token: token}
	return site
}

// ReadCommand is the command a terminal reads a page of the site with: curl and the page's whole address, where the
// site answers a terminal with the manual.
func (site Site) ReadCommand(page document.Page) string {
	return readingCommand + commandSpace + site.Address(page, FormatChosen)
}

// Address is where a target leads, written whole: a page's or a section's of one in a format, an address outside the
// site as it is, or a mailbox's under the scheme that writes to it. A target with no path leads nowhere, and its
// address is empty.
func (site Site) Address(target document.Target, format Format) string {
	switch target.(type) {
	case document.Address, document.Mail:
		return site.Reference(target, format)
	case document.Page, document.Anchor, document.File:
		if path := site.Reference(target, format); path != "" {
			return site.base + path
		}
	}
	return ""
}

// Reference is where a link inside one of the site's own pages leads: a page of the site, or a section of one, by its
// path from the root, which a browser reads against the address the page came from, so a page served from this
// machine or a staging host leads there and not to the live site; an address outside the site as it is; a mailbox
// under the scheme that writes to it. A target with no path leads nowhere, and its reference is empty.
func (site Site) Reference(target document.Target, format Format) string {
	switch typed := target.(type) {
	case document.Address:
		return typed.URL
	case document.Mail:
		return mailScheme + typed.Address
	case document.Page:
		return site.path(typed, format)
	case document.Anchor:
		return site.anchorPath(typed, format)
	case document.File:
		return site.filePath(typed)
	}
	return ""
}

// What parts a section's anchor from its page's path
const anchorMark = "#"

// anchorPath is where a section of a page is, in a format: its page's path, then the section's anchor; empty where the
// page is served in none.
func (site Site) anchorPath(anchor document.Anchor, format Format) string {
	path := site.path(anchor.Page, format)
	if path == "" {
		return ""
	}
	return path + anchorMark + anchor.Name
}

// path is where the site's reader reaches a page in a format: the post previewed in the preview's folder, every other
// page at its own path.
func (site Site) path(page document.Page, format Format) string {
	if site.previews(page) {
		return site.previewed.paths().inFormat(format)
	}
	return Path(page, format)
}

// filePath is where the site's reader reaches a file beside its page: a picture of the post previewed in the preview's
// folder, every other file at its own path.
func (site Site) filePath(file document.File) string {
	if site.previews(file.Page) {
		return fileIn(site.previewed.paths(), file.Name)
	}
	return FilePath(file)
}

// previews is whether a page is the post the site's reader reads at its preview link.
func (site Site) previews(page document.Page) bool {
	return site.previewed != nil && page.Kind == document.PagePost && page.Slug == site.previewed.slug
}

// paths are where a post read at its preview link is served: in the preview's folder under the link's token, by its
// slug, as it would be in the blog's folder, so its files and sections stand beside it as they would there.
func (previewed preview) paths() paths {
	path := previewFolder + previewed.token + folderEnd + previewed.slug
	return paths{chosen: path, files: path}
}

// InPreviewFolder is whether a path is in the preview's folder, by its opening alone, whether or not it names a token
// and a post: every answer there is sent as a preview link's is, since what follows may hold a token.
//
//wrapper:names the folder whose every answer goes out privately, which a host asks of a path no link is read from
func InPreviewFolder(path string) bool {
	return strings.HasPrefix(path, previewFolder)
}

// PreviewLink is a path in the preview's folder as PreviewPath reads it: the token it holds, the slug of the post it
// names, and the path the same face or file of that post has in the blog's folder.
type PreviewLink struct {
	Token    string
	Slug     string
	PostPath string
}

// What ends a post's slug in a path: its folder's end, before a file of it, or the dot a format's ending opens with
const slugEnds = folderEnd + "."

// PreviewPath reads a path in the preview's folder as the preview link it is, the one rule paths keeps the other way
// round, and whether it is one, naming a token and a post.
func PreviewPath(path string) (PreviewLink, bool) {
	inFolder, isInFolder := strings.CutPrefix(path, previewFolder)
	token, inPost, namesPost := strings.Cut(inFolder, folderEnd)
	slug := inPost
	if end := strings.IndexAny(inPost, slugEnds); end >= 0 {
		slug = inPost[:end]
	}
	if !isInFolder || !namesPost || token == "" || slug == "" {
		return PreviewLink{Token: "", Slug: "", PostPath: ""}, false
	}
	return PreviewLink{Token: token, Slug: slug, PostPath: blogFolder + inPost}, true
}

// FilePath is where a file is served beside its page, whatever the format: a post's picture in the post's own folder,
// by its name; empty for a page of any other kind, which serves none.
func FilePath(file document.File) string {
	switch file.Page.Kind {
	case document.PagePost:
		return fileIn(postPaths(file.Page.Slug), file.Name)
	case document.PageNone, document.PageResume, document.PageProject, document.PageIndex, document.PageHealth,
		document.PageSkills, document.PageBlog, document.PageBlogSeries, document.PageTagFilter, document.PageSeriesFilter:
	}
	return ""
}

// Path is where a page, or a tab of it shown alone, is served in a format, empty where it is served in none:
// PageNone, the hidden pages (the health report and the list of skills), the index in any format but markdown, or no
// format.
func Path(page document.Page, format Format) string {
	switch page.Kind {
	case document.PageNone, document.PageHealth, document.PageSkills:
		return ""
	case document.PageResume:
		return resumePaths(page.Tab).inFormat(format)
	case document.PageProject:
		return projectPaths(page.Slug, page.Tab).inFormat(format)
	case document.PageIndex:
		return indexPath(format)
	case document.PagePost:
		return versionPaths(page.Slug, page.Revision).inFormat(format)
	case document.PageBlog:
		return listPaths().inFormat(format)
	case document.PageBlogSeries:
		return paths{chosen: seriesList, files: seriesList}.inFormat(format)
	case document.PageTagFilter:
		return filtered(listPaths().inFormat(format), tagKey, page.Slug)
	case document.PageSeriesFilter:
		return filtered(listPaths().inFormat(format), seriesKey, page.Slug)
	}
	return ""
}

// The keys of the query a filter of the blog's list is written in, /blog?tag=payments and /blog?series=stint, the
// site's one address with a query (docs/FOUNDATION.md §5.4, Series, tags and lists), and the mark that opens a query
const (
	tagKey     = "tag"
	seriesKey  = "series"
	queryStart = "?"
)

// filterKeys are the keys of the query a filter of the blog's list is written in.
func filterKeys() []string {
	return []string{tagKey, seriesKey}
}

// listPaths are where the blog's list by year is served, and every filter of it beside it, with its query.
func listPaths() paths {
	return paths{chosen: blogList, files: blogList}
}

// filtered is the path of the blog's list in a format, filtered by a key's value in its query; none where the list has
// no path in the format.
func filtered(listPath, key, value string) string {
	if listPath == "" {
		return ""
	}
	return listPath + queryStart + url.Values{key: {value}}.Encode()
}

// Asked is the address a request asked for as its reader wrote it, which the page of no page quotes back: its path as
// the request decoded it, then its query as it was written, where it has one.
func Asked(path, rawQuery string) string {
	if rawQuery == "" {
		return path
	}
	return path + queryStart + rawQuery
}

// Requested is the address a request for a path with a query reaches its page by, as Path writes it: on the blog's
// list by year, in any format, the list filtered by the one filter its query names, by tag or by series; every other
// key, and the query of any other path, passed over, as a link's tracking words are. A query that names a filter
// twice, or both, is read with its filters alone, an address Path never writes, so it reaches no page.
func Requested(path string, query url.Values) string {
	named := url.Values{}
	for _, key := range filterKeys() {
		if values, isNamed := query[key]; isNamed {
			named[key] = values
		}
	}
	if len(named) == 0 || !slices.ContainsFunc(Formats(), func(format Format) bool { return Path(document.PageOf(document.PageBlog), format) == path }) {
		return path
	}
	return path + queryStart + named.Encode()
}

// NumberPath is where a post's number reaches it (D26), which answers by sending the reader on to the post's own
// address: in the blog's folder, as its slug is, since a slug is never digits alone.
func NumberPath(number int) string {
	return blogFolder + strconv.Itoa(number)
}

// The slash every path opens with, which a name a terminal writes for a path leaves out
const pathStart = "/"

// TerminalName is the name a path reaches its page by in a terminal, which has no root to begin a path at (D26): an SSH
// user name and the TUI's :e write blog/12 for /blog/12.
//
//wrapper:names the address a terminal reads, which an SSH user name and :e both write, once beside the paths
func TerminalName(path string) string {
	return strings.TrimPrefix(path, pathStart)
}

// BlogFolder is the blog's folder, where the site answers for every post by its slug and by its number, sending a
// number on to its slug.
func BlogFolder() string {
	return blogFolder
}

// IsBlogName is whether a name a terminal writes is the blog's: its list's, with a query or none, as blog?tag=go is,
// or one in its folder, as blog/99 is, whether or not a page goes by it.
func IsBlogName(name string) bool {
	path, _, _ := strings.Cut(pathStart+name, queryStart)
	return path == blogList || strings.HasPrefix(path, blogFolder)
}

// postPaths are where a post is served: in the blog's folder by its slug, its files under the same name. A post has
// no tabs.
func postPaths(slug string) paths {
	path := blogFolder + slug
	return paths{chosen: path, files: path}
}

// versionPaths are where a version of a post is served: the current at the post's own paths, an older one in the post's
// folder by its revision's name, a day, YYYY-MM-DD, its files under the same name (docs/FOUNDATION.md §5.4, Revisions).
// The folder holds the post's pictures too, each by a name that ends in its type's ending, which no day and no format's
// ending is, so a version and a picture never share a path.
func versionPaths(slug, revision string) paths {
	if revision == "" {
		return postPaths(slug)
	}
	path := fileIn(postPaths(slug), revision)
	return paths{chosen: path, files: path}
}

// paths are where a page is served: its own address, where the reader's request chooses the format, and the name
// its files are written under, before each format's ending.
type paths struct {
	chosen string
	files  string
}

// fileIn is where a file beside a page is served: in the page's own folder, by its name.
func fileIn(served paths, name string) string {
	return served.chosen + folderEnd + name
}

// resumePaths are where the résumé is served: whole at the site's root, its files under the name an index takes; a
// tab of it at the root, by its name.
func resumePaths(tab string) paths {
	if tab == "" {
		return paths{chosen: resumeRoot, files: resumeFile}
	}
	return paths{chosen: resumeRoot + tab, files: resumeRoot + tab}
}

// projectPaths are where a project's page is served: in the projects' folder by its slug, a tab of it in the page's
// own folder, by the tab's name.
func projectPaths(slug, tab string) paths {
	path := projectFolder + slug
	if tab != "" {
		path += folderEnd + tab
	}
	return paths{chosen: path, files: path}
}

// inFormat is the path a page is served at in a format.
func (served paths) inFormat(format Format) string {
	switch format {
	case FormatMarkdown:
		return served.files + markdownEnding
	case FormatManual:
		return served.files + manualEnding
	case FormatChosen:
		return served.chosen
	}
	return ""
}

// indexPath is where the index is served: in markdown alone, which llms.txt is.
func indexPath(format Format) string {
	switch format {
	case FormatMarkdown:
		return indexFile
	case FormatManual, FormatChosen:
	}
	return ""
}

// PicturePath is where a picture of the site is served, from the site's root; the zero picture has none.
func PicturePath(picture Picture) string {
	switch picture {
	case PictureDrawing:
		return drawingFile
	case PictureHomeScreen:
		return homeScreenFile
	case PictureAskedByName:
		return askedByNameFile
	}
	return ""
}

// Host is the site's own name: the host of its base address.
func (site Site) Host() string {
	return site.host
}
