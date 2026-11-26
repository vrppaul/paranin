package document

import "strings"

// Target is where a link leads: an address outside the site, a page of it or a section of one, a file served beside a
// page, a feed of the blog, or a mailbox, whose address routes writes in each face's format.
//
//sumtype:decl
type Target interface {
	target()
}

// Address is an address outside the site, written whole.
type Address struct {
	URL string
}

// Page is a page of the site, by its kind and slug, and the tab of it shown alone, or the older version of a post, in
// the language it is translated into; the zero page is none.
type Page struct {
	Kind PageKind
	Slug string
	// Tab is the name of the tab shown alone (Tab.Slug); none where the page is shown whole
	Tab string
	// Revision is the older version of a post shown, by the day it came out as its revision's name writes it,
	// YYYY-MM-DD; none for the current version, and on any other page
	Revision string
	// Translation is the language the page is translated into, by its tag, as ru; Untranslated in English, the site's
	// own
	Translation string
}

// Untranslated is the translation of a page in the site's own language: none.
const Untranslated = ""

// PageKind is what a page of the site is.
type PageKind int

const (
	// PageNone is no page: a page left unset is none
	PageNone PageKind = iota
	// PageResume is the résumé, the site's first page; it has no slug
	PageResume
	// PageProject is a project's own page, by the project's slug
	PageProject
	// PageIndex is the site's index, the list of its pages an agent reads first; it has no slug
	PageIndex
	// PageHealth is the hidden health report, after Neovim's :checkhealth; it has no slug
	PageHealth
	// PageSkills is the hidden list of every skill, after lazy.nvim's :Lazy; it has no slug
	PageSkills
	// PagePost is a post of the blog, by its slug; it has no tabs
	PagePost
	// PageBlog is the blog's list, its posts by year; it has no slug and no tabs
	PageBlog
	// PageBlogSeries is the blog's list, its posts by series; it has no slug and no tabs
	PageBlogSeries
	// PageTagFilter is the blog's list by year filtered to the posts of a tag, by the tag; it has no tabs
	PageTagFilter
)

// PageOf is the page of a kind that has no slug and is shown whole: the résumé, the index, a hidden page, the blog's
// list by year or by series; PageNone's, no page.
func PageOf(kind PageKind) Page {
	return Page{Kind: kind, Slug: "", Tab: "", Revision: "", Translation: Untranslated}
}

// PostPage is a post's page, its current version's, in the site's own language, by its slug.
func PostPage(slug string) Page {
	return Page{Kind: PagePost, Slug: slug, Tab: "", Revision: "", Translation: Untranslated}
}

// TagFilter is the blog's list filtered by a tag (docs/FOUNDATION.md §5.4, Series, tags and lists).
func TagFilter(tag string) Page {
	return Page{Kind: PageTagFilter, Slug: tag, Tab: "", Revision: "", Translation: Untranslated}
}

// OlderVersion is the older version of the post a page is, in its language, by its revision's name (docs/FOUNDATION.md
// §5.4, Revisions).
func (page Page) OlderVersion(revision string) Page {
	page.Revision = revision
	return page
}

// InLanguage is the page in a language, by its tag: translated into it, or in the site's own where the tag is none.
func (page Page) InLanguage(translation string) Page {
	page.Translation = translation
	return page
}

// InBlog is whether a page is the blog's, a post or the list, by year, by series or filtered, which a face shows in
// the blog's session, as tmux keeps a session apart. exhaustive holds the switch to every kind.
func (page Page) InBlog() bool {
	switch page.Kind {
	case PagePost, PageBlog, PageBlogSeries, PageTagFilter:
		return true
	case PageNone, PageResume, PageProject, PageIndex, PageHealth, PageSkills:
	}
	return false
}

// OwnAddress is whether a page has an address of its own, which a link's preview shows (Head's Previewed and Indexed
// name the pages that have none): a page a browser reads as a page of the site, but an older version of a post. No
// page, the index and the hidden pages are read by no browser as one, and a filtered list is a view of the list.
// exhaustive holds the switch to every kind.
func (page Page) OwnAddress() bool {
	switch page.Kind {
	case PageResume, PageProject, PagePost, PageBlog, PageBlogSeries:
		return page.Revision == ""
	case PageNone, PageIndex, PageHealth, PageSkills, PageTagFilter:
	}
	return false
}

// Unfiltered is the list a filtered list filters, the blog's list by year in its language, and whether the page is a
// filtered list. exhaustive holds the switch to every kind.
func (page Page) Unfiltered() (Page, bool) {
	switch page.Kind {
	case PageTagFilter:
		return PageOf(PageBlog).InLanguage(page.Translation), true
	case PageNone, PageResume, PageProject, PageIndex, PageHealth, PageSkills, PagePost, PageBlog, PageBlogSeries:
	}
	return page, false
}

// Anchor is a section of a page of the site, by the page and the section's anchor (Section.Name): a heading of a post,
// a group of the blog's list.
type Anchor struct {
	Page Page
	Name string
}

// File is a file the site serves beside a page, by the page and the file's name: a post's picture, a project's clip's
// recording and still.
type File struct {
	Page Page
	Name string
}

// Feed is a feed of the blog a reader subscribes to (docs/FOUNDATION.md §5.4, Feeds), by the series it follows, a
// series' parts out, by the series' name, or, where it names none, every post the blog lists.
type Feed struct {
	Series string
	// Translation is the language of the blog the feed follows, as a page names it
	Translation string
}

// Mail is a mailbox a link writes to, by its address as the content writes it.
type Mail struct {
	Address string
}

// MailAt is the at a mail address parts at, between its mailbox and its host, which a face draws between the two
const MailAt = "@"

// Parts are the mail's mailbox and its host, parted at its last at, since a host holds none; not parted where it
// holds no at.
//
//wrapper:names a mail's mailbox and host, which a face draws apart
func (mail Mail) Parts() (box, host string, parted bool) {
	return cutLast(mail.Address, MailAt)
}

// cutLast is a text cut at the last of a mark: what stands before it and what after; not cut where it holds none.
func cutLast(text, mark string) (before, after string, cut bool) {
	at := strings.LastIndex(text, mark)
	if at < 0 {
		return text, "", false
	}
	return text[:at], text[at+len(mark):], true
}

// LinkOf is the link spans are, where they are one link alone, with no words beside it.
func LinkOf(spans []Span) (Link, bool) {
	if len(spans) != 1 {
		return Link{Target: nil, Spans: nil}, false
	}
	link, isLink := spans[0].(Link)
	return link, isLink
}

// MailOf is the mail spans write to, where they are one link to a mailbox.
func MailOf(spans []Span) (Mail, bool) {
	link, isLink := LinkOf(spans)
	mail, isMail := link.Target.(Mail)
	return mail, isLink && isMail
}

func (Address) target() {}
func (Page) target()    {}
func (Anchor) target()  {}
func (File) target()    {}
func (Feed) target()    {}
func (Mail) target()    {}
