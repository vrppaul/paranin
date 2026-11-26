package tui

import (
	"slices"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// What stands between two names of a project's stack, as the page writes them
const stackJoint = space + placeMark + space

// readme lays out a project's head as the page lays it, in one column (DESIGN.md, The TUI): the name in the titles'
// font; the tagline, and what the project does; what gets it, the repository's owner and name, a link to it, then each
// way to install it, a command in a box under its label; then the facts that tell about it.
func (layout layout) readme(head document.Head) []line {
	lines := slices.Concat(layout.heading(document.PlainText(head.Title)), layout.fill(layout.words(head.Subtitle, kindStrong), 0), layout.fill(layout.words(head.Lead, kindText), 0))
	if head.Project == nil {
		return lines
	}
	gets := stacked(layout.fill(layout.repositoryWords(head.Project.Repository), 0), layout.factBlock(head.Project.InstallFacts(), 0))
	return stacked(stacked(lines, gets), layout.factBlock(head.Project.AboutFacts(stackJoint), 0))
}

// repositoryWords are a repository's owner and name, as its address names them, the path's mark between the two, each
// a piece of the link to the repository that copies its address; the address as written where it names none; none
// where there is no repository.
func (layout layout) repositoryWords(repository *document.Repository) []word {
	if repository == nil {
		return nil
	}
	written := repository.Address.URL
	link := run{text: written, kind: kindText, highlighted: false, address: layout.site.Address(repository.Address, routes.FormatChosen), copied: written, tone: 0, cites: noCitation, opens: document.PageOf(document.PageNone)}
	_, owner, name, parted := repository.Parts()
	if !parted {
		return []word{{runs: []run{link}, highlightBefore: false, spaceBefore: kindText}}
	}
	return []word{
		{runs: []run{retold(link, owner, kindOwner)}, highlightBefore: false, spaceBefore: kindText},
		{runs: []run{retold(link, document.RepositoryPathMark, kindFaint)}, highlightBefore: false, spaceBefore: kindText},
		{runs: []run{retold(link, name, kindRepository)}, highlightBefore: false, spaceBefore: kindText},
	}
}
