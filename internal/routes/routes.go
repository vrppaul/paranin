// Package routes writes every address of the site, from a document's link target and a format, so no face and
// no host writes a path of its own (docs/ARCHITECTURE.md, ADR-0003). It reads document alone.
package routes

import "github.com/vrppaul/paranin/internal/document"

// Format is the form a page is served in.
type Format int

const (
	// The zero format names none
	_ Format = iota
	// FormatMarkdown is CommonMark, for agents and terminals
	FormatMarkdown
)

// How the site's paths are written: the résumé in markdown, a project's folder, and markdown's ending
const (
	resumeMarkdown = "/index.md"
	projectFolder  = "/p/"
	markdownEnding = ".md"
)

// Site writes the site's addresses whole, from the address it is served at.
type Site struct {
	base string
}

// New is the site served at a base address, which ends in no slash.
func New(base string) Site {
	return Site{base: base}
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

// Path is where a page is served in a format, empty where it is served in none: PageNone, or a format the page
// has not.
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
		return resumeMarkdown
	}
	return ""
}

func projectPath(slug string, format Format) string {
	switch format {
	case FormatMarkdown:
		return projectFolder + slug + markdownEnding
	}
	return ""
}
