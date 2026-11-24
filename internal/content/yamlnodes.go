package content

import (
	"fmt"
	"iter"
	"slices"

	"go.yaml.in/yaml/v3"
)

// valueNodes are the nodes of YAML that hold values, from a node at a place in its file: the node, then every mapping,
// list and scalar it holds, each at its place, a mapping's keys never among them, and an alias where it stands, never
// the value it names.
func valueNodes(place string, node *yaml.Node) iter.Seq2[string, *yaml.Node] {
	return func(yield func(string, *yaml.Node) bool) { walkValues(place, node, yield) }
}

// walkValues hands a node at its place to yield, then each node of a value it holds at theirs, as valueNodes says;
// whether yield asked for more.
func walkValues(place string, node *yaml.Node, yield func(string, *yaml.Node) bool) bool {
	if !yield(place, node) {
		return false
	}
	for _, held := range valuesHeld(place, node) {
		if !walkValues(held.place, held.node, yield) {
			return false
		}
	}
	return true
}

// placedNode is a node of YAML at its place in its file.
type placedNode struct {
	place string
	node  *yaml.Node
}

// valuesHeld are the nodes of the values a node at a place holds itself, each at its place: a document's root, a list's
// items, a mapping's values; none of a scalar, and none of an alias, whose value stands where its anchor writes it.
func valuesHeld(place string, node *yaml.Node) []placedNode {
	held := []placedNode{}
	switch node.Kind {
	case yaml.DocumentNode:
		for _, root := range node.Content {
			held = append(held, placedNode{place: place, node: root})
		}
	case yaml.SequenceNode:
		for index, item := range node.Content {
			held = append(held, placedNode{place: fmt.Sprintf(itemPlace, place, index), node: item})
		}
	case yaml.MappingNode:
		for pair := range slices.Chunk(node.Content, keyAndValue) {
			held = append(held, placedNode{place: keyPlace(place, pair[keyOffset].Value), node: pair[valueOffset]})
		}
	case yaml.ScalarNode, yaml.AliasNode:
	}
	return held
}

// valueOf is the value a mapping writes under a key, and whether it writes one.
func valueOf(mapping *yaml.Node, key string) (*yaml.Node, bool) {
	for pair := range slices.Chunk(mapping.Content, keyAndValue) {
		if pair[keyOffset].Value == key {
			return pair[valueOffset], true
		}
	}
	return nil, false
}
