// Package routes writes every address of the site, from a document's link target and a format, or a picture of the
// site, and a post's at its preview link, and reads a request's query into the address of the filtered list it names,
// so no face and no host writes a path of its own (docs/ARCHITECTURE.md, ADR-0003). It reads document alone.
package routes

import (
	"net/url"
	"path"
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

// The media types the site writes its pages and feeds in, each named once for every face and host: markdown, plain
// text, the manual's and the index's, the page a browser shows, and a feed
const (
	MarkdownType = "text/markdown"
	PlainType    = "text/plain"
	PageType     = "text/html"
	FeedType     = "application/atom+xml"
)

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
	// the folder a post's preview link opens it in, its language's folder and the link's token after it
	// (docs/FOUNDATION.md §5.4, Status)
	previewRoot   = "/preview"
	previewFolder = previewRoot + folderEnd
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
	// tagAuthority is what the site's feeds and posts are named under for good, wherever the site moves (RFC 4151)
	tagAuthority string
	// translations are the languages the site is translated into, by their tags, each a folder of its own
	translations []string
	// previewed is the post read at its preview link, none where the site is read at its own addresses alone
	previewed *preview
}

// preview is a post read at its preview link: its page, in the language it is read in, and the token the link holds.
type preview struct {
	post  document.Page
	token string
}

// inLanguage is a page's path in the site's own language as its translation has it, in the language's folder.
func inLanguage(translation, path string) string {
	if translation == document.Untranslated || path == "" {
		return path
	}
	return pathStart + translation + path
}

// languageFolderOf parts a path into the language the site is translated into whose folder opens it, and the path
// after; none and the path whole where no such folder does.
func (site Site) languageFolderOf(path string) (string, string) {
	inside, _ := strings.CutPrefix(path, pathStart)
	language, rest, parted := strings.Cut(inside, folderEnd)
	if !parted || !slices.Contains(site.translations, language) {
		return document.Untranslated, path
	}
	return language, folderEnd + rest
}

// BaseAddress is the address a site is served at, as New's caller hands it: as it is written, ending in no slash,
// and its host. Content's web address is one, read once from the profile and held by its check to the web.
type BaseAddress interface {
	Written() string
	Host() string
}

// New is the site served at a base address, and named by the address's host; what it serves is named for good under a
// tag authority, as the profile gives it; it is translated into languages, by their tags, each read from a folder.
func New(base BaseAddress, tagAuthority string, translations []string) Site {
	return Site{base: base.Written(), host: base.Host(), tagAuthority: tagAuthority, translations: translations, previewed: nil}
}

// How a path of the site is named for good (RFC 4151, and RFC 4287's atom:id): the scheme, the site's tag authority,
// then a colon and the path with no opening slash
const (
	tagScheme    = "tag:"
	tagSeparator = ":"
)

// TagURI is the name a path of the site keeps for good under its tag authority, wherever the site moves: a feed's by
// its path, a post's by the path its number reaches it at, which never changes (D26).
func (site Site) TagURI(path string) string {
	return tagScheme + site.tagAuthority + tagSeparator + strings.TrimPrefix(path, pathStart)
}

// Previewing is the site as the reader of a post's preview link reads it (docs/FOUNDATION.md §5.4, Status): every
// address of that post, in the language it is read in, in the preview's folder under the link's token; every other
// page at its own.
func (site Site) Previewing(post document.Page, token string) Site {
	site.previewed = &preview{post: post, token: token}
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
	case document.Page, document.Anchor, document.File, document.Feed:
		if path := site.Reference(target, format); path != "" {
			return site.base + path
		}
	}
	return ""
}

// Reference is where a link inside one of the site's own pages leads: a page of the site, a section of one, a file or a
// feed, by its path from the root, which a browser reads against the address the page came from, so a page served
// from this machine or a staging host leads there and not to the live site; an address outside the site as it is; a
// mailbox under the scheme that writes to it. A target with no path leads nowhere, and its reference is empty.
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
	case document.Feed:
		return FeedPath(typed)
	}
	return ""
}

// The ending of a feed's path, which no format's, picture's or day's is, so a feed never shares a page's or a file's
// path
const feedEnding = ".atom"

// FeedPath is where a feed is served, in Atom alone, whatever the format a link names (docs/FOUNDATION.md §5.4,
// Feeds): the blog's beside its list by year, /blog.atom, and a series' in the list by series' folder by the series'
// name, /blog/series/stint.atom, where no post stands, since the check refuses a post slugged as that list; in its
// language's folder, as that language's list is.
func FeedPath(feed document.Feed) string {
	if feed.Series == "" {
		return inLanguage(feed.Translation, blogList+feedEnding)
	}
	return inLanguage(feed.Translation, seriesList+folderEnd+feed.Series+feedEnding)
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

// previews is whether a page is the post the site's reader reads at its preview link, in the language they read it in.
func (site Site) previews(page document.Page) bool {
	return site.previewed != nil && page.Kind == document.PagePost && page.Slug == site.previewed.post.Slug &&
		page.Translation == site.previewed.post.Translation
}

// paths are where a post read at its preview link is served: in the preview's folder, its language's folder in it, as
// in the blog's folder.
func (previewed preview) paths() paths {
	path := previewRoot + inLanguage(previewed.post.Translation, folderEnd+previewed.token+folderEnd+previewed.post.Slug)
	return paths{chosen: path, files: path}
}

// InPreviewFolder is whether a path is in the preview's folder by the opening of its clean form, whether or not it
// names a language, a token and a post: every answer there is sent as a preview link's is, since what follows may hold
// a token, the redirect net/http's ServeMux sends a path not in clean form on to its clean form among them. No page
// stands there, since the check refuses one whose folder would (PreviewFolderTaken).
func InPreviewFolder(written string) bool {
	return strings.HasPrefix(cleanPath(written), previewFolder)
}

// cleanPath is a path in the clean form net/http's ServeMux sends a request on to: rooted, with no empty segment, no
// dot and no segment leading out, and its final slash kept.
func cleanPath(written string) string {
	rooted := pathStart + strings.TrimPrefix(written, pathStart)
	clean := path.Clean(rooted)
	if strings.HasSuffix(rooted, folderEnd) && clean != pathStart {
		clean += folderEnd
	}
	return clean
}

// PreviewFolderTaken is the folder of a page, where its files and older versions stand, and whether it is the
// preview's folder, as a tab of the résumé titled preview would take: the check refuses such a page, so
// InPreviewFolder never names one.
func PreviewFolderTaken(page document.Page) (string, bool) {
	path := Path(page, FormatChosen)
	folder := path + folderEnd
	return folder, path != "" && InPreviewFolder(folder)
}

// previewFolderOf parts a path in the preview's folder into the language the site is translated into whose folder
// opens what follows, and the path after it, from its slash; and whether the path is in the preview's folder at all.
func (site Site) previewFolderOf(path string) (string, string, bool) {
	inFolder, isInFolder := strings.CutPrefix(path, previewFolder)
	if !isInFolder {
		return document.Untranslated, path, false
	}
	language, rest := site.languageFolderOf(folderEnd + inFolder)
	return language, rest, true
}

// LanguageOf is the language a path is read in: the one whose folder opens it, or, in the preview's folder, the one
// whose folder opens what follows; the site's own where no language's folder does.
func (site Site) LanguageOf(path string) string {
	if language, _, inPreview := site.previewFolderOf(path); inPreview {
		return language
	}
	language, _ := site.languageFolderOf(path)
	return language
}

// PreviewLink is a path in the preview's folder as PreviewPath reads it: its token, its post, and that post's path.
type PreviewLink struct {
	Token    string
	Post     document.Page
	PostPath string
}

// What ends a post's slug in a path: its folder's end, before a file of it, or the dot a format's ending opens with
const slugEnds = folderEnd + "."

// PreviewPath reads a path in the preview's folder as the preview link it is, the one rule paths keeps the other way
// round, and whether it is one, naming a token and a post.
func (site Site) PreviewPath(path string) (PreviewLink, bool) {
	language, inFolder, isInFolder := site.previewFolderOf(path)
	token, inPost, namesPost := strings.Cut(strings.TrimPrefix(inFolder, folderEnd), folderEnd)
	slug := inPost
	if end := strings.IndexAny(inPost, slugEnds); end >= 0 {
		slug = inPost[:end]
	}
	if !isInFolder || !namesPost || token == "" || slug == "" {
		return PreviewLink{Token: "", Post: document.PageOf(document.PageNone), PostPath: ""}, false
	}
	return PreviewLink{Token: token, Post: document.PostPage(slug).InLanguage(language), PostPath: inLanguage(language, blogFolder+inPost)}, true
}

// FilePath is where a file is served beside its page, whatever the format: a post's picture in the post's own folder,
// by its name, in its language's folder; empty for a page of any other kind, which serves none.
func FilePath(file document.File) string {
	switch file.Page.Kind {
	case document.PagePost:
		return inLanguage(file.Page.Translation, fileIn(postPaths(file.Page.Slug), file.Name))
	case document.PageNone, document.PageResume, document.PageProject, document.PageIndex, document.PageHealth,
		document.PageSkills, document.PageBlog, document.PageBlogSeries, document.PageTagFilter:
	}
	return ""
}

// Path is where a page, or a tab of it shown alone, is served in a format, in its language's folder where it is
// translated (docs/FOUNDATION.md §5.4, Translations); empty where it is served in none: PageNone, the hidden pages (the
// health report and the list of skills), the index in any format but markdown, or no format.
func Path(page document.Page, format Format) string {
	return inLanguage(page.Translation, ownPath(page, format))
}

// ownPath is where a page is served in a format in the site's own language, as Path has it.
func ownPath(page document.Page, format Format) string {
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
		return filtered(listPaths().inFormat(format), page.Slug)
	}
	return ""
}

// The key of the query the blog's list filtered by a tag is written in, /blog?tag=payments, the site's one address
// with a query (docs/FOUNDATION.md §5.4, Series, tags and lists), and the mark that opens a query
const (
	tagKey     = "tag"
	queryStart = "?"
)

// listPaths are where the blog's list by year is served, and every filter of it beside it, with its query.
func listPaths() paths {
	return paths{chosen: blogList, files: blogList}
}

// filtered is the path of the blog's list in a format, filtered by a tag in its query; none where the list has no
// path in the format.
func filtered(listPath, tag string) string {
	if listPath == "" {
		return ""
	}
	return listPath + queryStart + url.Values{tagKey: {tag}}.Encode()
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
// list by year, in any language and format, the list filtered by the tag its query names; every other key, and the
// query of any other path, passed over, as a link's tracking words are. A query that names a tag twice is read with
// its tags alone, an address Path never writes, so it reaches no page.
func (site Site) Requested(path string, query url.Values) string {
	tags, isNamed := query[tagKey]
	_, own := site.languageFolderOf(path)
	if !isNamed || !slices.ContainsFunc(Formats(), func(format Format) bool { return ownPath(document.PageOf(document.PageBlog), format) == own }) {
		return path
	}
	return path + queryStart + url.Values{tagKey: tags}.Encode()
}

// NumberPath is where a post's number reaches it in a language (D26), which answers by sending the reader on to the
// post's own address there: in the blog's folder, as its slug is, since a slug is never digits alone.
func NumberPath(translation string, number int) string {
	return inLanguage(translation, blogFolder+strconv.Itoa(number))
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

// BlogFolders are the blog's folder in the site's own language, then in each it is translated into, where the site
// answers for every post by its slug and by its number, sending a number on to its slug.
func (site Site) BlogFolders() []string {
	folders := []string{blogFolder}
	for _, translation := range site.translations {
		folders = append(folders, inLanguage(translation, blogFolder))
	}
	return folders
}

// BlogNamed is the language whose blog a name a terminal writes is, by its tag, none for the site's own, and whether it
// is a blog's: its list's, with a query or none, as blog?tag=go is, or one in its folder, as blog/99 is, whether or not
// a page goes by it, in the site's own language or in the folder of one the site is translated into, as ru/blog is.
func (site Site) BlogNamed(name string) (string, bool) {
	translation, rest := site.languageFolderOf(pathStart + name)
	path, _, _ := strings.Cut(rest, queryStart)
	if path != blogList && !strings.HasPrefix(path, blogFolder) {
		return document.Untranslated, false
	}
	return translation, true
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
