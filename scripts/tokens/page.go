package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"go.yaml.in/yaml/v3"
)

// The unit every number of pixels in the file is written in, once the stylesheet writes it: a text style's size and
// line, a length
const pixelUnit = "px"

// The weights a face is drawn in: a hundred to nine hundred, in steps of a hundred
const (
	lightestWeight = 100
	heaviestWeight = 900
	weightStep     = 100
)

// The keys a text style's and a length's numbers are written under, as their tags say
const (
	sizeKey   = "size"
	lineKey   = "line"
	weightKey = "weight"
	pixelsKey = "pixels"
)

// The face's own name, the property every rule of the page's text inherits it from
const faceName = "face"

// What a text style's properties are named: the word type, the style's name, then what each sets
const (
	typePrefix   = "type_"
	sizeSuffix   = "_size"
	lineSuffix   = "_line"
	weightSuffix = "_weight"
)

// How a family is written in the stylesheet: its name quoted, unless it is one of CSS's generic families, a word
// CSS knows, which quoted would name a face of that name; the names parted by commas
const (
	familyQuote     = `"`
	familySeparator = ", "
)

// How a number of pixels is written in the stylesheet, and a weight, which takes no unit
const (
	pixelValue  = "%d" + pixelUnit
	weightValue = "%d"
)

// What each of a text style's three properties is for, after what the style is for
const (
	sizeAbout   = "%s: its size"
	lineAbout   = "%s: its line's height"
	weightAbout = "%s: its weight"
)

// The tag YAML gives a number written in digits alone, unquoted, and how a fault in one names its place
const (
	integerTag  = "!!int"
	numberFault = "line %d: %q is no whole number written in digits alone, unquoted"
)

// genericFamilies are CSS's generic families, which the stylesheet writes unquoted.
func genericFamilies() []string {
	return []string{
		"serif", "sans-serif", "monospace", "cursive", "fantasy", "math", "emoji", "fangsong",
		"system-ui", "ui-serif", "ui-sans-serif", "ui-monospace", "ui-rounded",
	}
}

// faceToken is the page's one face as the file writes it: what it is for, and the families a browser tries in turn,
// each by its name alone.
type faceToken struct {
	About    string   `yaml:"about"`
	Families []string `yaml:"families"`
}

// typeToken is a text style as the file writes it: its name, what it is for, its size and the height of its line in
// pixels, and its weight. A number the file may leave out unseen is a pointer, so its absence is a fault.
type typeToken struct {
	Name   string       `yaml:"name"`
	About  string       `yaml:"about"`
	Size   *wholeNumber `yaml:"size"`
	Line   *wholeNumber `yaml:"line"`
	Weight *wholeNumber `yaml:"weight"`
}

// lengthToken is a length the page lays out by, as the file writes it: its name, what it is for, and how many
// pixels long it is.
type lengthToken struct {
	Name   string       `yaml:"name"`
	About  string       `yaml:"about"`
	Pixels *wholeNumber `yaml:"pixels"`
}

// wholeNumber is a number of the file that counts whole things: pixels, or a weight.
type wholeNumber uint

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
		{name: property + sizeSuffix, about: fmt.Sprintf(sizeAbout, style.about), value: fmt.Sprintf(pixelValue, style.size)},
		{name: property + lineSuffix, about: fmt.Sprintf(lineAbout, style.about), value: fmt.Sprintf(pixelValue, style.line)},
		{name: property + weightSuffix, about: fmt.Sprintf(weightAbout, style.about), value: fmt.Sprintf(weightValue, style.weight)},
	}
}

// pageTokens are the tokens the page alone reads, read: its face, its text styles and its lengths.
type pageTokens struct {
	face    readFace
	types   []readType
	lengths []readLength
}

// readFace is the face read: what it is for, and its families, each a name of letters, digits, hyphens and spaces.
type readFace struct {
	about    string
	families []string
}

// readType is a text style read, its numbers each there and each one a face can be drawn at.
type readType struct {
	name   string
	about  string
	size   wholeNumber
	line   wholeNumber
	weight wholeNumber
}

// readLength is a length read, in pixels.
type readLength struct {
	name   string
	about  string
	pixels wholeNumber
}

// readPageTokens reads the face, the text styles and the lengths of the file, each property a name of its own
// beside every property named already.
func readPageTokens(written tokensFile, named map[string]bool) (pageTokens, error) {
	face, err := readFaceToken(written.Face, named)
	if err != nil {
		return pageTokens{}, err
	}
	types := make([]readType, 0, len(written.Type))
	for _, token := range written.Type {
		style, err := readTypeToken(token, named)
		if err != nil {
			return pageTokens{}, err
		}
		types = append(types, style)
	}
	lengths := make([]readLength, 0, len(written.Lengths))
	for _, token := range written.Lengths {
		length, err := readLengthToken(token, named)
		if err != nil {
			return pageTokens{}, err
		}
		lengths = append(lengths, length)
	}
	return pageTokens{face: face, types: types, lengths: lengths}, nil
}

// readFaceToken refuses a face that says nothing of what it is for, names no family, or names one in anything but
// letters, digits, hyphens and spaces, so no name can close the quotes the stylesheet writes it in.
func readFaceToken(token faceToken, named map[string]bool) (readFace, error) {
	if err := checkToken(faceName, token.About, named); err != nil {
		return readFace{}, err
	}
	if len(token.Families) == 0 {
		return readFace{}, fmt.Errorf("%w: the face has no family", errTokens)
	}
	for _, family := range token.Families {
		if family == "" || strings.ContainsFunc(family, isNoFamilyLetter) {
			return readFace{}, fmt.Errorf("%w: the face's family %q is no name of letters, digits, hyphens and spaces", errTokens, family)
		}
	}
	return readFace{about: token.About, families: token.Families}, nil
}

// isNoFamilyLetter is whether a letter has no place in a family's name.
func isNoFamilyLetter(letter rune) bool {
	return !unicode.IsLetter(letter) && !unicode.IsDigit(letter) && letter != '-' && letter != ' '
}

// readTypeToken reads a text style, each of its three properties a name of its own.
func readTypeToken(token typeToken, named map[string]bool) (readType, error) {
	if err := checkName(token.Name, token.About); err != nil {
		return readType{}, err
	}
	size, err := pixels(token.Size, token.Name, sizeKey)
	if err != nil {
		return readType{}, err
	}
	line, err := pixels(token.Line, token.Name, lineKey)
	if err != nil {
		return readType{}, err
	}
	weight, err := required(token.Weight, token.Name, weightKey)
	if err != nil {
		return readType{}, err
	}
	if weight < lightestWeight || weight > heaviestWeight || weight%weightStep != 0 {
		return readType{}, fmt.Errorf("%w: %q has the weight %d, no weight of %d to %d in steps of %d",
			errTokens, token.Name, weight, lightestWeight, heaviestWeight, weightStep)
	}
	style := readType{name: token.Name, about: token.About, size: size, line: line, weight: weight}
	for _, property := range style.properties() {
		if err := claim(property.name, named); err != nil {
			return readType{}, err
		}
	}
	return style, nil
}

// readLengthToken reads a length, more than no pixels long.
func readLengthToken(token lengthToken, named map[string]bool) (readLength, error) {
	if err := checkToken(token.Name, token.About, named); err != nil {
		return readLength{}, err
	}
	length, err := pixels(token.Pixels, token.Name, pixelsKey)
	if err != nil {
		return readLength{}, err
	}
	return readLength{name: token.Name, about: token.About, pixels: length}, nil
}

// pixels is a number of pixels the file must write, more than none.
func pixels(written *wholeNumber, token, key string) (wholeNumber, error) {
	value, err := required(written, token, key)
	if err != nil {
		return 0, err
	}
	if value == 0 {
		return 0, fmt.Errorf("%w: %q has the %s 0, which draws nothing", errTokens, token, key)
	}
	return value, nil
}

// familyList is the face's families as CSS writes them: each name quoted, so the browser reads it as a face's name
// whatever it is made of, but a generic family, which CSS knows by its bare word in any case.
func familyList(families []string) string {
	written := make([]string, 0, len(families))
	for _, family := range families {
		if !slices.Contains(genericFamilies(), strings.ToLower(family)) {
			family = familyQuote + family + familyQuote
		}
		written = append(written, family)
	}
	return strings.Join(written, familySeparator)
}
