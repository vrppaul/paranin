package compose

import "strconv"

// What joins a heading's words asked for again to its count, as a slug joins its words, and the count it takes the
// second time
const (
	countJoint  = "-"
	secondCount = 2
)

// anchors are the anchors given so far to the headings of one version of a post, so no two share one.
type anchors map[string]bool

// given is the anchor a heading asking for one is given, and takes: the one it asks for, or that with its count after
// it, from the second, where an earlier heading took it already.
func (taken anchors) given(asked string) string {
	anchor := asked
	for count := secondCount; taken[anchor]; count++ {
		anchor = asked + countJoint + strconv.Itoa(count)
	}
	taken[anchor] = true
	return anchor
}
