package content

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"
)

// How a value's place in a file is written: a key after the place of the mapping that holds it, an item of a list
// after the list's place and in brackets its index
const (
	keySeparator = "."
	itemPlace    = "%s[%d]"
)

// A mapping's node holds its keys and values in one list, each key before its value
const (
	keyAndValue = 2
	keyOffset   = 0
	valueOffset = 1
)

// lineFeed is the one control character a value may hold, which a block of lines, as a recording, is written with.
// A tab is refused: how wide it stands hangs on the terminal, so no face could lay it out.
const lineFeed = '\n'

// inspectCharacters finds every value of a file that holds a character a terminal or a browser acts on rather than
// shows, or bytes that are no text: every face passes a value on as written. It walks the file's document, not the
// types' fields, so a field a type gains is checked as well.
func inspectCharacters(file string, document *yaml.Node) Faults {
	inspection := &inspection{file: file, faults: nil}
	inspectNode(inspection, "", document)
	return inspection.faults
}

// inspectNode checks a node at its place in the file, and every node it holds at theirs. A key needs no check of
// its own: one that no type or catalogue knows is a fault already, and a fault's line writes it escaped.
func inspectNode(inspection *inspection, place string, node *yaml.Node) {
	switch node.Kind {
	case yaml.DocumentNode:
		for _, root := range node.Content {
			inspectNode(inspection, place, root)
		}
	case yaml.SequenceNode:
		for index, item := range node.Content {
			inspectNode(inspection, fmt.Sprintf(itemPlace, place, index), item)
		}
	case yaml.MappingNode:
		for pair := range slices.Chunk(node.Content, keyAndValue) {
			inspectNode(inspection, keyPlace(place, pair[keyOffset].Value), pair[valueOffset])
		}
	case yaml.ScalarNode:
		inspectScalar(inspection, place, node)
	case yaml.AliasNode:
		// an alias's value is checked where its anchor writes it
	}
}

// inspectScalar checks a scalar's text as each reader of it gets it: as written, which a type's own reader takes, and
// as the decoder hands it to a string, which a tag such as !!binary turns into other bytes.
func inspectScalar(inspection *inspection, place string, node *yaml.Node) {
	texts := []string{node.Value}
	var decoded string
	// a scalar no string can take reaches a type only as written
	if node.Decode(&decoded) == nil && decoded != node.Value {
		texts = append(texts, decoded)
	}
	for _, text := range texts {
		if refusedText(text) {
			inspection.add(place, characterProblem, text)
		}
	}
}

// refusedText is whether a text holds a character no value may hold, or bytes that are no text.
func refusedText(text string) bool {
	return !utf8.ValidString(text) || strings.ContainsFunc(text, refusedCharacter)
}

// keyPlace is the place of a key in a mapping at a place; a key of the file's top mapping is its own place.
func keyPlace(mappingPlace, key string) string {
	if mappingPlace == "" {
		return key
	}
	return mappingPlace + keySeparator + key
}

// refusedCharacter is whether a value may not hold a character: one a reader acts on, but the line feed.
func refusedCharacter(character rune) bool {
	return character != lineFeed && actedOn(character)
}

// actedOn is whether a terminal or a browser acts on a character rather than shows it: a control character, which
// moves the cursor, clears the screen or opens a link, or a mark of direction, which turns the text round. Unicode's
// whole set of the marks is taken, since no text here runs right to left.
func actedOn(character rune) bool {
	return unicode.IsControl(character) || unicode.Is(unicode.Bidi_Control, character)
}
