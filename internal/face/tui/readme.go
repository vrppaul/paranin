package tui

import (
	"net/url"
	"slices"
	"strings"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// The mark between two parts of an address's path, and so between a repository's owner and its name as the page
// writes them
const pathMark = "/"

// readme lays out a project's head as the page lays it, in one column (DESIGN.md, The TUI): the name in the titles'
// font; the tagline, and what the project does; what gets it, the repository's owner and name, a link to it, then each
// way to install it, a command in a box under its label; then the facts that tell about it.
func (layout layout) readme(head document.Head) []line {
	lines := slices.Concat(layout.heading(document.PlainText(head.Title)), layout.fill(layout.words(head.Subtitle, kindStrong), 0), layout.fill(layout.words(head.Lead, kindText), 0))
	if head.Project == nil {
		return lines
	}
	gets := stacked(layout.fill(layout.repositoryWords(head.Project.Repository), 0), layout.factBlock(head.Project.InstallFacts(), 0))
	return stacked(stacked(lines, gets), layout.factBlock(head.Project.About, 0))
}

// repositoryWords are a repository's owner and name, as the parts of its address's path give them, the path's mark
// between each two, each a piece of the link to the repository that copies its address; the address as written where
// its path gives none; none where there is no repository.
func (layout layout) repositoryWords(repository *document.Repository) []word {
	if repository == nil {
		return nil
	}
	written := repository.Address.URL
	link := run{text: written, kind: kindText, highlighted: false, address: layout.site.Address(repository.Address, routes.FormatChosen), copied: written, tone: 0, cites: noCitation, opens: document.PageOf(document.PageNone)}
	parts := pathParts(written)
	if len(parts) == 0 {
		return []word{{runs: []run{link}, highlightBefore: false, spaceBefore: kindText}}
	}
	words := []word{}
	for index, part := range parts {
		if index > 0 {
			words = append(words, word{runs: []run{retold(link, pathMark, kindFaint)}, highlightBefore: false, spaceBefore: kindText})
		}
		partKind := kindOwner
		if index == len(parts)-1 {
			partKind = kindRepository
		}
		words = append(words, word{runs: []run{retold(link, part, partKind)}, highlightBefore: false, spaceBefore: kindText})
	}
	return words
}

// pathParts are the parts of an address's path, by which a forge's address names a repository's owner and name; none
// where the address has no path or reads as no address.
func pathParts(address string) []string {
	parsed, err := url.Parse(address)
	if err != nil {
		return nil
	}
	path := strings.Trim(parsed.Path, pathMark)
	if path == "" {
		return nil
	}
	return strings.Split(path, pathMark)
}
