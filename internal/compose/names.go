package compose

import "strconv"

// What joins a name asked for again to its count, as a slug joins its words, and the count it takes the second time
const (
	countJoint  = "-"
	secondCount = 2
)

// names are the names given so far on one page, which its places go by, so no two share one.
type names struct {
	taken map[string]bool
}

// namesHeld are a page's names with those its other places go by taken already.
func namesHeld(held []string) names {
	taken := make(map[string]bool, len(held))
	for _, name := range held {
		taken[name] = true
	}
	return names{taken: taken}
}

// given is the name a place asking for one is given, and takes: the one it asks for, or that with its count after
// it, from the second, where it is taken already.
func (held names) given(asked string) string {
	name := asked
	for count := secondCount; held.taken[name]; count++ {
		name = asked + countJoint + strconv.Itoa(count)
	}
	held.taken[name] = true
	return name
}
