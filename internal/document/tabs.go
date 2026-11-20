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

// TabsOf parts a document: a résumé whose first section holds facts opens with that section, briefed, under its
// title; any other document opens under the head's title; each section left is a tab, under its own.
func TabsOf(composed Document) Tabs {
	loose, sections := SectionsOf(composed.Blocks)
	var opening Opening
	left := sections
	if briefed := BriefOf(sections); composed.Head.Page.Kind == PageResume && briefed != nil {
		opening = Opening{Tab: tabOf(briefed.Title, slices.Concat(loose, briefed.Before, briefed.After)), Brief: briefed}
		left = briefed.Rest
	} else {
		opening = Opening{Tab: tabOf(composed.Head.Title, loose), Brief: nil}
	}
	rest := make([]Tab, 0, len(left))
	for _, section := range left {
		rest = append(rest, tabOf(section.Title, slices.Clip(section.Blocks)))
	}
	return Tabs{Opening: opening, Rest: rest}
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
