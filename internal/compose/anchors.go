package compose

import (
	"cmp"
	"strconv"

	"github.com/vrppaul/paranin/internal/document"
)

// What joins an anchor asked for again to its count, as a slug joins its words, and the count it takes the second time
const (
	countJoint  = "-"
	secondCount = 2
)

// The name of a heading whose title holds no letter or digit, which document.SlugOf makes none of
const unnamedHeading = "heading"

// anchors are the anchors given so far on one page, to the headings and widgets of one version of a post or to the
// sections of a project's page, so no two share one.
type anchors map[string]bool

// given is the anchor a heading or a widget asking for one is given, and takes: the one it asks for, or that with its
// count after it, from the second, where an earlier one took it already.
func (taken anchors) given(asked string) string {
	anchor := asked
	for count := secondCount; taken[anchor]; count++ {
		anchor = asked + countJoint + strconv.Itoa(count)
	}
	taken[anchor] = true
	return anchor
}

// headingAnchor is the anchor a heading of a title is given: its title's slug, or the word for a heading where the
// title holds no letter or digit.
func (taken anchors) headingAnchor(title []document.Span) string {
	return taken.given(cmp.Or(document.SlugOf(document.PlainText(title)), unnamedHeading))
}
