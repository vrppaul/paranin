// Package routes writes every address of the site, from a document's link target and a format, so no face and
// no host writes a path of its own (docs/ARCHITECTURE.md, ADR-0003). It reads document alone.
package routes

import (
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

// How the site's paths are written: the résumé at the site's root and in its files, a project's folder, the index
// where agents look for it (llmstxt.org), and the endings of the formats
const (
	resumeRoot     = "/"
	resumeFile     = "/index"
	indexFile      = "/llms.txt"
	projectFolder  = "/p/"
	markdownEnding = ".md"
	manualEnding   = ".txt"
	// the mark before a place's id, where a link leads to a place on a page
	placeMark = "#"
)

// The program a terminal reads the site with, and what parts it from the address in a command
const (
	readingCommand = "curl"
	commandSpace   = " "
)

// The scheme of an address that writes to a mailbox
const mailScheme = "mailto:"

// Site writes the site's addresses whole, from the address it is served at, and knows the site's own name.
type Site struct {
	base string
	host string
}

// BaseAddress is the address a site is served at, as New's caller hands it: as it is written, ending in no slash,
// and its host. Content's web address is one, read once from the profile and held by its check to the web.
type BaseAddress interface {
	Written() string
	Host() string
}

// New is the site served at a base address, and named by the address's host.
func New(base BaseAddress) Site {
	return Site{base: base.Written(), host: base.Host()}
}

// ReadCommand is the command a terminal reads a page of the site with: curl and the page's whole address, where the
// site answers a terminal with the manual.
func (site Site) ReadCommand(page document.Page) string {
	return readingCommand + commandSpace + site.Address(page, FormatChosen)
}

// Address is where a target leads, written whole: a page's in a format, an address outside the site as it is, or a
// mailbox's under the scheme that writes to it. A target with no path leads nowhere, and its address is empty.
func (site Site) Address(target document.Target, format Format) string {
	switch typed := target.(type) {
	case document.Address:
		return typed.URL
	case document.Mail:
		return mailScheme + typed.Address
	case document.Page:
		if path := Path(typed, format); path != "" {
			return site.base + path
		}
	}
	return ""
}

// Reference is where a link inside one of the site's own pages leads: a page of the site by its path from the
// root, which a browser reads against the address the page came from, so a page served from this machine or a
// staging host leads there and not to the live site; an address outside the site as it is; a mailbox under the
// scheme that writes to it. A target with no path leads nowhere, and its reference is empty.
func Reference(target document.Target, format Format) string {
	switch typed := target.(type) {
	case document.Address:
		return typed.URL
	case document.Mail:
		return mailScheme + typed.Address
	case document.Page:
		return Path(typed, format)
	}
	return ""
}

// PlaceReference is where a link inside one of the site's pages leads to a place on a page of the site, by the
// place's id: the page's reference, then the place.
func PlaceReference(page document.Page, format Format, id string) string {
	return Reference(page, format) + Place(id)
}

// Place is where a link leads to a place on the page it stands on, by the place's id.
func Place(id string) string {
	return placeMark + id
}

// Path is where a page is served in a format, empty where it is served in none: PageNone, the hidden pages (the
// health report and the list of skills), the index in any format but markdown, or no format.
func Path(page document.Page, format Format) string {
	switch page.Kind {
	case document.PageNone, document.PageHealth, document.PageSkills:
		return ""
	case document.PageResume:
		return resumePath(format)
	case document.PageProject:
		return projectPath(page.Slug, format)
	case document.PageIndex:
		return indexPath(format)
	}
	return ""
}

func resumePath(format Format) string {
	switch format {
	case FormatMarkdown:
		return resumeFile + markdownEnding
	case FormatManual:
		return resumeFile + manualEnding
	case FormatChosen:
		return resumeRoot
	}
	return ""
}

func projectPath(slug string, format Format) string {
	switch format {
	case FormatMarkdown:
		return projectFolder + slug + markdownEnding
	case FormatManual:
		return projectFolder + slug + manualEnding
	case FormatChosen:
		return projectFolder + slug
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

// Host is the site's own name: the host of its base address.
func (site Site) Host() string {
	return site.host
}
