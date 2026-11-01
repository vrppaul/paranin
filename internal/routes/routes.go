// Package routes writes every address of the site, from a document's link target and a format, so no face and
// no host writes a path of its own (docs/ARCHITECTURE.md, ADR-0003). It reads document alone.
package routes

import (
	"net/url"

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

// How the site's paths are written: the résumé at the site's root and in its files, a project's folder, and the
// endings of the formats
const (
	resumeRoot     = "/"
	resumeFile     = "/index"
	projectFolder  = "/p/"
	markdownEnding = ".md"
	manualEnding   = ".txt"
)

// Site writes the site's addresses whole, from the address it is served at, and knows the site's own name.
type Site struct {
	base string
	host string
}

// New is the site served at a base address, which ends in no slash; content's check holds it to a web address.
func New(base string) Site {
	// content's check has parsed this address already and refused one that does not parse; the second parse is
	// a debt (TODO.md), and its failure branch keeps the whole address as the name
	host := base
	if parsed, err := url.Parse(base); err == nil {
		host = parsed.Host
	}
	return Site{base: base, host: host}
}

// Address is where a target leads, written whole: a page's in a format, or an address outside the site as it is.
// A target with no path leads nowhere, and its address is empty.
func (site Site) Address(target document.Target, format Format) string {
	switch typed := target.(type) {
	case document.Address:
		return typed.URL
	case document.Page:
		if path := Path(typed, format); path != "" {
			return site.base + path
		}
	}
	return ""
}

// Path is where a page is served in a format, empty where it is served in none: PageNone, or no format.
func Path(page document.Page, format Format) string {
	switch page.Kind {
	case document.PageNone:
		return ""
	case document.PageResume:
		return resumePath(format)
	case document.PageProject:
		return projectPath(page.Slug, format)
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

// Host is the site's own name: the host of its base address.
func (site Site) Host() string {
	return site.host
}
