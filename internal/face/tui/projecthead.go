package tui

import (
	"slices"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// projectHead lays out a project's head as the page lays it, in one column (DESIGN.md, The TUI): the name in the
// titles' font; the tagline, and what the project does; what gets it; then what tells about it.
func (layout layout) projectHead(head document.Head) []line {
	lines := slices.Concat(layout.heading(document.PlainText(head.Title)), layout.fill(layout.words(head.Subtitle, kindStrong), 0), layout.fill(layout.words(head.Lead, kindText), 0))
	if head.Project == nil {
		return lines
	}
	return stacked(stacked(lines, layout.projectGets(*head.Project)), layout.projectAbout(*head.Project))
}

// projectGets are what gets a project: the repository's owner and name, a link to it; its install, a command in a box
// under the site's word; then each other way as the manual lays it, its note on a line of its own and its command
// boxed under it.
func (layout layout) projectGets(project document.Project) []line {
	lines := layout.fill(layout.repositoryWords(project.Repository), 0)
	if install, installed := project.InstallFact(); installed {
		lines = stacked(lines, layout.commandBox(install, 0))
	}
	for _, alternative := range project.Alternatives {
		note := layout.fill(layout.words(alternative.Note, kindText), 0)
		box := layout.commandBox(document.CommandFact(project.Labels.Install, alternative.Command), 0)
		lines = stacked(lines, append(note, box...))
	}
	return lines
}

// projectAbout are what tells about a project, each under the site's word, the words in a column after the widest
// label, as far as the project gives each: what it is made of, each name whole; a link to its newest release and one
// to their history; its licence. The faint mark stands between two names and between the two links.
func (layout layout) projectAbout(project document.Project) []line {
	told := []labelledWords{}
	if len(project.Stack) > 0 {
		names := make([][]word, 0, len(project.Stack))
		for _, name := range project.Stack {
			names = append(names, []word{wordOf(name, kindText)})
		}
		told = append(told, labelledWords{label: project.Labels.Stack, words: faintlyMarked(names)})
	}
	if release := project.Release; release != nil {
		links := [][]word{layout.valueWords([]document.Span{release.Latest}), layout.valueWords([]document.Span{release.Changelog})}
		told = append(told, labelledWords{label: project.Labels.Release, words: faintlyMarked(links)})
	}
	if project.Licence != "" {
		told = append(told, labelledWords{label: project.Labels.Licence, words: textWords(project.Licence, kindText)})
	}
	labels := make([]string, 0, len(told))
	for _, labelled := range told {
		labels = append(labels, labelled.label)
	}
	column, lines := labelColumn(labels), []line{}
	for _, labelled := range told {
		lines = append(lines, layout.labelled(labelled, 0, column, kindMuted)...)
	}
	return lines
}

// faintlyMarked are groups of words in a row, the place mark in faint between each two, as the page parts a project's
// names and links.
func faintlyMarked(groups [][]word) []word {
	words := []word{}
	for index, group := range groups {
		if index > 0 {
			words = append(words, wordOf(placeMark, kindFaint))
		}
		words = append(words, group...)
	}
	return words
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
