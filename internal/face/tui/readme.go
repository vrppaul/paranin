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

// readme lays out a project's head as its forge shows its README, as glow draws one: the name in the titles' font,
// the repository's owner and name under it, a link to it; what it is made of; the tagline, and what the project does.
func (layout layout) readme(head document.Head) []line {
	lines := slices.Concat(layout.heading(document.PlainText(head.Title)), layout.fill(layout.repositoryWords(head.Repository), 0))
	lines = stacked(lines, layout.fill(madeOfWords(head.Tags), 0))
	return stacked(lines, slices.Concat(layout.fill(layout.words(head.Subtitle, kindStrong), 0), layout.fill(layout.words(head.Lead, kindText), 0)))
}

// repositoryWords are a repository's owner and name, as the parts of its address's path give them, the path's mark
// between each two, each a piece of the link to the repository that copies its address; the address as written where
// its path gives none; none where there is no repository.
func (layout layout) repositoryWords(repository *document.Repository) []word {
	if repository == nil {
		return nil
	}
	written := repository.Address.URL
	link := run{text: written, kind: kindText, highlighted: false, address: layout.site.Address(repository.Address, routes.FormatChosen), copied: written, tone: 0, cites: noCitation}
	parts := pathParts(written)
	if len(parts) == 0 {
		return []word{{runs: []run{link}, highlightBefore: false}}
	}
	words := []word{}
	for index, part := range parts {
		if index > 0 {
			words = append(words, word{runs: []run{retold(link, pathMark, kindFaint)}, highlightBefore: false})
		}
		partKind := kindOwner
		if index == len(parts)-1 {
			partKind = kindRepository
		}
		words = append(words, word{runs: []run{retold(link, part, partKind)}, highlightBefore: false})
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

// madeOfWords are what a project is made of, its tags, as the page's head writes them: words that say less than the
// text, the place mark between each two tags; not chips, which would read as something to press.
func madeOfWords(tags []string) []word {
	words := []word{}
	for _, tag := range tags {
		words = markedWords(words, textWords(tag, kindMuted), placeMark)
	}
	return words
}
