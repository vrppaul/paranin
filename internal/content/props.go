package content

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"go.yaml.in/yaml/v3"
)

// What JSON writes round an object's members and an array's items, between two of them, and between a member's name and
// its value
const (
	objectStart   = '{'
	objectEnd     = '}'
	arrayStart    = '['
	arrayEnd      = ']'
	itemSeparator = ','
	nameSeparator = ':'
)

// The tags YAML gives a scalar written with no tag of its own, beside text's (textTag): a whole number, a number with a
// fraction, true or false, null, and a date, which JSON has no type for and writes as text
const (
	wholeNumberTag = "!!int"
	fractionTag    = "!!float"
	truthTag       = "!!bool"
	nullTag        = "!!null"
	timestampTag   = "!!timestamp"
)

// propsJSON is a widget's props as a JSON object, in their written order, every key but its alt text's and its
// caption's, which the page draws round the widget; a value JSON holds none of is a fault at its line.
func (reading *propsReading) propsJSON(props *yaml.Node) json.RawMessage {
	own := []*yaml.Node{}
	for pair := range slices.Chunk(props.Content, keyAndValue) {
		if key := pair[keyOffset].Value; key != altKey && key != captionKey {
			own = append(own, pair...)
		}
	}
	return reading.objectJSON(nil, own)
}

// valueJSON writes a value of a widget's props after what is written: a mapping as an object, a list as an array, a
// scalar as the type YAML reads it as.
func (reading *propsReading) valueJSON(written []byte, node *yaml.Node) []byte {
	switch node.Kind {
	case yaml.MappingNode:
		written = reading.objectJSON(written, node.Content)
	case yaml.SequenceNode:
		written = reading.arrayJSON(written, node.Content)
	case yaml.ScalarNode:
		written = append(written, reading.scalarJSON(node)...)
	case yaml.AliasNode, yaml.DocumentNode:
		// no value of the props is either: the inspection refuses an alias (inspectOwn), and the document holds them all
	}
	return written
}

// objectJSON writes a mapping's keys and values, each key before its value as its node holds them, as a JSON object.
func (reading *propsReading) objectJSON(written []byte, keysAndValues []*yaml.Node) []byte {
	written = append(written, objectStart)
	for start := 0; start < len(keysAndValues); start += keyAndValue {
		if start > 0 {
			written = append(written, itemSeparator)
		}
		written = append(written, reading.scalarJSON(keysAndValues[start+keyOffset])...)
		written = append(written, nameSeparator)
		written = reading.valueJSON(written, keysAndValues[start+valueOffset])
	}
	return append(written, objectEnd)
}

// arrayJSON writes a list's items as a JSON array.
func (reading *propsReading) arrayJSON(written []byte, items []*yaml.Node) []byte {
	written = append(written, arrayStart)
	for index, item := range items {
		if index > 0 {
			written = append(written, itemSeparator)
		}
		written = reading.valueJSON(written, item)
	}
	return append(written, arrayEnd)
}

// scalarJSON is a scalar of a widget's props, a value or a key, as JSON writes the type YAML reads it as: text and a
// date as a string, a number, true or false, and null; or a fault at its line where the page would read no such value
// exactly, a number of no end or a whole number past what JSON.parse holds (pastExactNumbers). The load refuses a
// widget with a fault, so nothing stands in its place.
func (reading *propsReading) scalarJSON(node *yaml.Node) []byte {
	if node.Style == plainStyle && pastExactNumbers(node.Value) {
		reading.add(node.Line, widgetValueProblem, node.Value)
		return nil
	}
	var encoded []byte
	var err error
	switch node.ShortTag() {
	case textTag, timestampTag:
		encoded, err = encodedJSON(node.Value)
	case wholeNumberTag:
		encoded, err = decodedJSON[int64](node)
	case fractionTag:
		encoded, err = decodedJSON[float64](node)
	case truthTag:
		encoded, err = decodedJSON[bool](node)
	case nullTag:
		encoded, err = encodedJSON(nil)
	default:
		// a tag of the value's own, which the inspection refuses already (inspectOwn)
		return nil
	}
	if err != nil {
		reading.add(node.Line, widgetValueProblem, node.Value)
	}
	return encoded
}

// How a scalar is written with no quotes and as no block, the one form YAML reads a number from; the largest whole
// number JSON.parse reads exactly, of a double's 53 bits; and how strconv reads a number's base off its prefix, as YAML
// writes 0x1F
const (
	plainStyle        yaml.Style = 0
	largestExactWhole            = 1<<53 - 1
	baseOfItsPrefix              = 0
)

// pastExactNumbers is whether a plain scalar is a number the page reads no exact value of: a whole number past
// ±(2^53−1), which YAML reads as a fraction once it passes 64 bits, or a number of no end, as 1e400, which YAML
// reads as text.
func pastExactNumbers(written string) bool {
	whole, wholeErr := strconv.ParseInt(written, baseOfItsPrefix, numberBits)
	_, fractionErr := strconv.ParseFloat(written, numberBits)
	pastWhole := wholeErr == nil && (whole > largestExactWhole || whole < -largestExactWhole)
	return pastWhole || errors.Is(wholeErr, strconv.ErrRange) || errors.Is(fractionErr, strconv.ErrRange)
}

// decodedJSON is a scalar decoded into Go's type for it, and that value as JSON; the fault of one the type holds none
// of, or JSON holds none of.
func decodedJSON[Value any](node *yaml.Node) ([]byte, error) {
	var value Value
	if err := node.Decode(&value); err != nil {
		return nil, fmt.Errorf("decode %q: %w", node.Value, err)
	}
	return encodedJSON(value)
}

// encodedJSON is a value as JSON, its text's characters as written: the page escapes what its HTML would read as
// markup as it writes the props, and no other reader of them needs that.
func encodedJSON(value any) ([]byte, error) {
	var written bytes.Buffer
	encoder := json.NewEncoder(&written)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, fmt.Errorf("write %v as JSON: %w", value, err)
	}
	// the encoder ends each value with a line's end, which no member of an object takes
	return bytes.TrimSuffix(written.Bytes(), []byte{lineFeed}), nil
}
