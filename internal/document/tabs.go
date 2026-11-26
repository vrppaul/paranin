package document

import (
	"slices"
	"strings"
	"unicode"
)

// What joins the words of a title in the name a tab is known by
const slugJoint = "-"

// Tabs are a document parted into the screens a reader moves between: its opening, then a tab for each section left.
type Tabs struct {
	Opening Opening
	Rest    []Tab
}

// Tab is a screen of a document: its title, the name it is known by, and the blocks it shows.
type Tab struct {
	Title  []Span
	Slug   string
	Blocks []Block
}

// Opening is a document's first tab, and the brief it opens with, none where it opens with none.
type Opening struct {
	Tab
	Brief *Briefed
}

// TabsOf parts a document by the kind of its page, the one place that decides which pages part into tabs. The résumé
// opens with its first section where that holds details, briefed, under its title; the index and the hidden pages open
// under the head's title with the blocks before their first section; each section left is a tab under its own title.
// A project, a post and the blog's lists are one tab under the head's title holding every block, sections included.
func TabsOf(composed Document) Tabs {
	switch composed.Head.Page.Kind {
	case PageResume:
		return resumeTabs(composed)
	case PageNone, PageIndex, PageHealth, PageSkills:
		return partedTabs(composed)
	case PageProject, PagePost, PageBlog, PageBlogSeries, PageTagFilter:
	}
	return Tabs{Opening: Opening{Tab: tabOf(composed.Head.Title, composed.Blocks), Brief: nil}, Rest: []Tab{}}
}

// resumeTabs parts the résumé: a brief's section, where its first section holds details, is the opening, else the
// document parts as partedTabs.
func resumeTabs(composed Document) Tabs {
	loose, sections := SectionsOf(composed.Blocks)
	briefed := BriefOf(sections)
	if briefed == nil {
		return partedTabs(composed)
	}
	opening := Opening{Tab: tabOf(briefed.Title, slices.Concat(loose, briefed.Before, briefed.After)), Brief: briefed}
	return Tabs{Opening: opening, Rest: sectionTabs(briefed.Rest)}
}

// partedTabs opens a document under the head's title with the blocks before its first section, a tab for each section.
func partedTabs(composed Document) Tabs {
	loose, sections := SectionsOf(composed.Blocks)
	return Tabs{Opening: Opening{Tab: tabOf(composed.Head.Title, loose), Brief: nil}, Rest: sectionTabs(sections)}
}

// sectionTabs are sections as tabs, each under its own title.
func sectionTabs(sections []Section) []Tab {
	tabs := make([]Tab, 0, len(sections))
	for _, section := range sections {
		tabs = append(tabs, tabOf(section.Title, slices.Clip(section.Blocks)))
	}
	return tabs
}

// tabOf is a tab under a title, showing blocks.
func tabOf(title []Span, blocks []Block) Tab {
	return Tab{Title: title, Slug: SlugOf(PlainText(title)), Blocks: blocks}
}

// SlugOf is the name a title is known by: its words in small letters, letters and digits alone, joined by hyphens.
func SlugOf(title string) string {
	words := strings.FieldsFunc(strings.ToLower(title), func(letter rune) bool {
		return !unicode.IsLetter(letter) && !unicode.IsDigit(letter)
	})
	return strings.Join(words, slugJoint)
}
