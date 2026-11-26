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
// label: its parts, each whole, the faint mark between two.
func (layout layout) projectAbout(project document.Project) []line {
	about := project.About()
	labels := make([]string, 0, len(about))
	for _, told := range about {
		labels = append(labels, told.Label)
	}
	column, lines := labelColumn(labels), []line{}
	for _, told := range about {
		parts := make([][]word, 0, len(told.Parts))
		for _, part := range told.Parts {
			parts = append(parts, []word{wholeWord(layout.valueWords([]document.Span{part}))})
		}
		lines = append(lines, layout.labelled(labelledWords{label: told.Label, words: faintlyMarked(parts)}, 0, column, kindMuted)...)
	}
	return lines
}

// wholeWord is words laid as one word, a space between each two, which no line breaks inside: a part of what tells
// about a project, a name of more than one word among them.
func wholeWord(words []word) word {
	if len(words) == 0 {
		return word{runs: nil, highlightBefore: false, spaceBefore: kindText}
	}
	return word{runs: wordRuns(words), highlightBefore: words[0].highlightBefore, spaceBefore: words[0].spaceBefore}
}

// projectEnd lays out where a project's reader goes next, after its sections, in one column (DESIGN.md, ProjectEnd):
// where it has others, the heading over them, strong, the way to them all beside it, then the others on the blog's
// grid of tiles; then the site's word for contact, strong, and the author's mail under it, parted as the brief parts
// it. None on a page that is no project's.
func (layout layout) projectEnd(head document.Head) []line {
	project := head.Project
	if project == nil {
		return nil
	}
	mail := document.Link{Target: project.Mail, Spans: []document.Span{document.Text{Text: project.Mail.Address}}}
	contact := []line{{plainRun(project.Labels.Contact, kindStrong)}, mailRuns(wordRuns(layout.valueWords([]document.Span{mail})), project.Mail)}
	if len(project.Others) == 0 {
		return contact
	}
	lead := line{plainRun(project.Labels.Others, kindStrong), gapRun(headingGap)}
	others := slices.Concat(layout.continued(lead, layout.words([]document.Span{project.AllProjects}, kindText), 0), blankLines(), layout.otherTiles(project.Others))
	return stacked(others, contact)
}

// otherTiles are the other projects on the blog's grid of tiles, as many in a row as hold a tile each, sharing their
// edges as the page's panes share their lines; each its name, bold, a link to its page, its tagline muted, and what it
// is made of faint at its foot, each name whole, the faint mark between two.
func (layout layout) otherTiles(others []document.OtherProject) []line {
	grid := gridOf(layout.width)
	inside := layout
	inside.width = grid.inside()
	rows := [][]gridCell{}
	for first := 0; first < len(others); first += grid.columns {
		row := []gridCell{}
		for _, other := range others[first:min(first+grid.columns, len(others))] {
			row = append(row, gridCell{label: nil, words: inside.otherWords(other), leads: true})
		}
		rows = append(rows, row)
	}
	return grid.framed(rows).lines
}

// otherWords are another project's words in its tile: its name, bold, a link to its page, and its tagline muted at
// its top; what it is made of, each name whole, the faint mark between two, at its foot.
func (layout layout) otherWords(other document.OtherProject) tileWords {
	name := layout.words([]document.Span{document.Link{Target: other.Page, Spans: other.Name}}, kindStrong)
	names := make([][]word, 0, len(other.Stack))
	for _, written := range other.Stack {
		names = append(names, []word{wordOf(written, kindFaint)})
	}
	top := slices.Concat(layout.fill(name, 0), layout.fill(layout.words(other.Tagline, kindMuted), 0))
	foot := []line{}
	if len(names) > 0 {
		foot = append(blankLines(), layout.fill(faintlyMarked(names), 0)...)
	}
	return tileWords{top: top, foot: foot}
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
