package main

import (
	"fmt"
	"math"
	"strconv"

	"go.yaml.in/yaml/v3"
)

// How a size of the file is written in the stylesheet, a text style's size and line and a length: in rem, the root's
// text size, which is the one its reader set, so the whole page grows with it; in the fewest decimals that write it
// exactly, a float of 64 bits. No size is written in pixels
const (
	remUnit          = "rem"
	shortestDecimals = -1
	sizeFloatBits    = 64
)

// The keys a text style's, a length's and a time's numbers are written under, as their tags say
const (
	sizeKey         = "size"
	lineKey         = "line"
	weightKey       = "weight"
	remKey          = "rem"
	cellsKey        = "cells"
	millisecondsKey = "milliseconds"
)

// How a length in cells is written in the stylesheet: in ch, a cell of the page's mono face, so a space inside a line
// stands on the face's own grid
const cellsValue = "%dch"

// How a time is written in the stylesheet: its milliseconds
const (
	millisecondUnit  = "ms"
	millisecondValue = "%d" + millisecondUnit
)

// What a text style's properties are named: the word type, the style's name, then what each sets
const (
	typePrefix   = "type_"
	sizeSuffix   = "_size"
	lineSuffix   = "_line"
	weightSuffix = "_weight"
)

// How a weight and a layer are written in the stylesheet: a number, with no unit
const numberValue = "%d"

// What a layer's property is named: the word layer, then the layer's name, so a layer may share its name with the
// length of the part it lifts; and the place of the lowest layer, over the page's own
const (
	layerPrefix = "layer_"
	lowestPlace = 1
)

// What each of a text style's three properties is for, after what the style is for
const (
	sizeAbout   = "%s: its size"
	lineAbout   = "%s: its line's height"
	weightAbout = "%s: its weight"
)

// The tags YAML gives a number written unquoted, in digits alone or with a fraction, and how a fault in a whole
// number and in a size names its place
const (
	integerTag  = "!!int"
	floatTag    = "!!float"
	numberFault = "line %d: %q is no whole number written in digits alone, unquoted"
	sizeFault   = "line %d: %q is no number of rem written in digits, unquoted and with no unit"
)

// typeToken is a text style as the file writes it: its name, what it is for, its size and the height of its line in
// rem, and its weight. A number the file may leave out unseen is a pointer, so its absence is a fault.
type typeToken struct {
	Name   string       `yaml:"name"`
	About  string       `yaml:"about"`
	Size   *remSize     `yaml:"size"`
	Line   *remSize     `yaml:"line"`
	Weight *wholeNumber `yaml:"weight"`
}

// lengthToken is a length the page lays out by, as the file writes it: its name, what it is for, and how long it is,
// in rem or in the face's cells, never both.
type lengthToken struct {
	Name  string       `yaml:"name"`
	About string       `yaml:"about"`
	Rem   *remSize     `yaml:"rem"`
	Cells *wholeNumber `yaml:"cells"`
}

// layerToken is a layer of what stands over the page, as the file writes it: its name and what stands on it. Its place
// among the layers, lowest first, is its height over the page.
type layerToken struct {
	Name  string `yaml:"name"`
	About string `yaml:"about"`
}

// timeToken is a time the page waits or moves by, as the file writes it: its name, what it is for, and how many
// milliseconds long it is.
type timeToken struct {
	Name         string       `yaml:"name"`
	About        string       `yaml:"about"`
	Milliseconds *wholeNumber `yaml:"milliseconds"`
}

// wholeNumber is a number of the file that counts whole things: cells, milliseconds, or a weight.
type wholeNumber uint

// remSize is a size of the file in rem, a fraction among them: a text style's size or line, or a length.
type remSize float64

// UnmarshalYAML reads a whole number written in decimal digits alone, unquoted: the YAML library would cut a
// fraction off unseen, and the stylesheet would draw a number the file never wrote. Its fault names the line;
// readTokens marks it as the tokens' own.
func (number *wholeNumber) UnmarshalYAML(node *yaml.Node) error {
	parsed, err := strconv.ParseUint(node.Value, decimal, strconv.IntSize)
	if node.Tag != integerTag || err != nil {
		return fmt.Errorf(numberFault, node.Line, node.Value)
	}
	*number = wholeNumber(parsed)
	return nil
}

// UnmarshalYAML reads a size written as a number, unquoted, its unit left to the stylesheet: a string would let a
// unit in, and a number with no end or none at all draws nothing, which YAML lets through only under a tag written
// out, `!!float inf`. Its fault names the line; readTokens marks it as the tokens' own.
func (size *remSize) UnmarshalYAML(node *yaml.Node) error {
	parsed, err := strconv.ParseFloat(node.Value, sizeFloatBits)
	if (node.Tag != integerTag && node.Tag != floatTag) || err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return fmt.Errorf(sizeFault, node.Line, node.Value)
	}
	*size = remSize(parsed)
	return nil
}

// stylesheetValue is a size as the stylesheet writes it: in rem, in the fewest decimals that write it exactly.
func (size remSize) stylesheetValue() string {
	return strconv.FormatFloat(float64(size), 'f', shortestDecimals, sizeFloatBits) + remUnit
}

// styleProperty is one of the properties a text style is written as: its token's name, what it is for, and its
// value as the stylesheet writes it.
type styleProperty struct {
	name  string
	about string
	value string
}

// properties are the three properties a text style is written as, its size, its line's height and its weight: the
// reader claims their names and the writer writes them, from this one list.
func (style readType) properties() []styleProperty {
	property := typePrefix + style.name
	return []styleProperty{
		{name: property + sizeSuffix, about: fmt.Sprintf(sizeAbout, style.about), value: style.size.stylesheetValue()},
		{name: property + lineSuffix, about: fmt.Sprintf(lineAbout, style.about), value: style.line.stylesheetValue()},
		{name: property + weightSuffix, about: fmt.Sprintf(weightAbout, style.about), value: fmt.Sprintf(numberValue, style.weight)},
	}
}

// pageTokens are the tokens the page alone reads, read: its face, its text styles, its lengths, its times and its
// layers.
type pageTokens struct {
	face    readFace
	types   []readType
	lengths []readLength
	times   []readTime
	layers  []readLayer
}

// readType is a text style read, its numbers each there and each one a face can be drawn at.
type readType struct {
	name   string
	about  string
	size   remSize
	line   remSize
	weight wholeNumber
}

// readLength is a length read, in rem or in cells.
type readLength struct {
	name  string
	about string
	value lengthValue
}

// lengthValue is a length in one of the units the file writes one in, each written in the stylesheet its own way.
//
//sumtype:decl
type lengthValue interface {
	stylesheetValue() string
}

// cellCount is a length in whole cells of the page's mono face.
type cellCount uint

// stylesheetValue is a length in cells as the stylesheet writes it: in ch.
func (cells cellCount) stylesheetValue() string {
	return fmt.Sprintf(cellsValue, cells)
}

// readTime is a time read, in milliseconds.
type readTime struct {
	name         string
	about        string
	milliseconds wholeNumber
}

// readLayer is a layer read: its property's name, what stands on it, and its place among the layers, the lowest
// first, over the page's own.
type readLayer struct {
	name  string
	about string
	place int
}

// readPageTokens reads the face, the text styles, the lengths, the times and the layers of the file, each property a
// name of its own beside every property named already.
func readPageTokens(written tokensFile, named map[string]bool) (pageTokens, error) {
	face, err := readFaceToken(written.Face, named)
	if err != nil {
		return pageTokens{}, err
	}
	types, err := readEach(written.Type, named, readTypeToken)
	if err != nil {
		return pageTokens{}, err
	}
	if err := checkFaceWeights(face.files, types); err != nil {
		return pageTokens{}, err
	}
	lengths, err := readEach(written.Lengths, named, readLengthToken)
	if err != nil {
		return pageTokens{}, err
	}
	times, err := readEach(written.Times, named, readTimeToken)
	if err != nil {
		return pageTokens{}, err
	}
	layers, err := readLayerTokens(written.Layers, named)
	if err != nil {
		return pageTokens{}, err
	}
	return pageTokens{face: face, types: types, lengths: lengths, times: times, layers: layers}, nil
}

// readLayerTokens reads the layers, each in its place, the lowest first; a layer's property is named by the word layer
// and its name, which no other property may hold.
func readLayerTokens(written []layerToken, named map[string]bool) ([]readLayer, error) {
	layers := make([]readLayer, 0, len(written))
	for index, token := range written {
		if err := checkName(token.Name, token.About); err != nil {
			return nil, err
		}
		layer := readLayer{name: layerPrefix + token.Name, about: token.About, place: lowestPlace + index}
		if err := claim(layer.name, named); err != nil {
			return nil, err
		}
		layers = append(layers, layer)
	}
	return layers, nil
}

// readTypeToken reads a text style, each of its three properties a name of its own.
func readTypeToken(token typeToken, named map[string]bool) (readType, error) {
	if err := checkName(token.Name, token.About); err != nil {
		return readType{}, err
	}
	size, err := counted(token.Size, token.Name, sizeKey)
	if err != nil {
		return readType{}, err
	}
	line, err := counted(token.Line, token.Name, lineKey)
	if err != nil {
		return readType{}, err
	}
	weight, err := required(token.Weight, token.Name, weightKey)
	if err != nil {
		return readType{}, err
	}
	if err := checkWeight(weight, token.Name); err != nil {
		return readType{}, err
	}
	style := readType{name: token.Name, about: token.About, size: size, line: line, weight: weight}
	for _, property := range style.properties() {
		if err := claim(property.name, named); err != nil {
			return readType{}, err
		}
	}
	return style, nil
}

// readLengthToken reads a length, more than nothing long, in rem or in cells: one of the two, since a length written
// in both would leave the stylesheet to guess which the file means.
func readLengthToken(token lengthToken, named map[string]bool) (readLength, error) {
	if err := checkToken(token.Name, token.About, named); err != nil {
		return readLength{}, err
	}
	switch {
	case token.Rem != nil && token.Cells != nil:
		return readLength{}, fmt.Errorf("%w: %q has both %s and %s", errTokens, token.Name, remKey, cellsKey)
	case token.Rem != nil:
		rem, err := counted(token.Rem, token.Name, remKey)
		if err != nil {
			return readLength{}, err
		}
		return readLength{name: token.Name, about: token.About, value: rem}, nil
	case token.Cells != nil:
		cells, err := counted(token.Cells, token.Name, cellsKey)
		if err != nil {
			return readLength{}, err
		}
		return readLength{name: token.Name, about: token.About, value: cellCount(cells)}, nil
	default:
		return readLength{}, fmt.Errorf("%w: %q has neither %s nor %s", errTokens, token.Name, remKey, cellsKey)
	}
}

// readTimeToken reads a time, more than no milliseconds long.
func readTimeToken(token timeToken, named map[string]bool) (readTime, error) {
	if err := checkToken(token.Name, token.About, named); err != nil {
		return readTime{}, err
	}
	milliseconds, err := counted(token.Milliseconds, token.Name, millisecondsKey)
	if err != nil {
		return readTime{}, err
	}
	return readTime{name: token.Name, about: token.About, milliseconds: milliseconds}, nil
}

// counted is a number the file must write, a size or milliseconds, more than none.
func counted[Number remSize | wholeNumber](written *Number, token, key string) (Number, error) {
	value, err := required(written, token, key)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, fmt.Errorf("%w: %q has the %s %v, which is not more than nothing", errTokens, token, key, value)
	}
	return value, nil
}
