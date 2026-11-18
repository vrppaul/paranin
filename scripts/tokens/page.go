package main

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode"

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

// The weights a face is drawn in: a hundred to nine hundred, in steps of a hundred
const (
	lightestWeight = 100
	heaviestWeight = 900
	weightStep     = 100
)

// The keys a text style's, a length's and a time's numbers are written under, as their tags say
const (
	sizeKey         = "size"
	lineKey         = "line"
	weightKey       = "weight"
	remKey          = "rem"
	millisecondsKey = "milliseconds"
)

// How a time is written in the stylesheet: its milliseconds
const (
	millisecondUnit  = "ms"
	millisecondValue = "%d" + millisecondUnit
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

// How a weight is written in the stylesheet: its number, with no unit
const weightValue = "%d"

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
// rem, and its weight. A number the file may leave out unseen is a pointer, so its absence is a fault.
type typeToken struct {
	Name   string       `yaml:"name"`
	About  string       `yaml:"about"`
	Size   *remSize     `yaml:"size"`
	Line   *remSize     `yaml:"line"`
	Weight *wholeNumber `yaml:"weight"`
}

// lengthToken is a length the page lays out by, as the file writes it: its name, what it is for, and how many rem
// long it is.
type lengthToken struct {
	Name  string   `yaml:"name"`
	About string   `yaml:"about"`
	Rem   *remSize `yaml:"rem"`
}

// timeToken is a time the page waits or moves by, as the file writes it: its name, what it is for, and how many
// milliseconds long it is.
type timeToken struct {
	Name         string       `yaml:"name"`
	About        string       `yaml:"about"`
	Milliseconds *wholeNumber `yaml:"milliseconds"`
}

// wholeNumber is a number of the file that counts whole things: milliseconds, or a weight.
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
		{name: property + weightSuffix, about: fmt.Sprintf(weightAbout, style.about), value: fmt.Sprintf(weightValue, style.weight)},
	}
}

// pageTokens are the tokens the page alone reads, read: its face, its text styles, its lengths and its times.
type pageTokens struct {
	face    readFace
	types   []readType
	lengths []readLength
	times   []readTime
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
	size   remSize
	line   remSize
	weight wholeNumber
}

// readLength is a length read, in rem.
type readLength struct {
	name  string
	about string
	rem   remSize
}

// readTime is a time read, in milliseconds.
type readTime struct {
	name         string
	about        string
	milliseconds wholeNumber
}

// readPageTokens reads the face, the text styles, the lengths and the times of the file, each property a name of its
// own beside every property named already.
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
	times := make([]readTime, 0, len(written.Times))
	for _, token := range written.Times {
		timed, err := readTimeToken(token, named)
		if err != nil {
			return pageTokens{}, err
		}
		times = append(times, timed)
	}
	return pageTokens{face: face, types: types, lengths: lengths, times: times}, nil
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

// readLengthToken reads a length, more than no rem long.
func readLengthToken(token lengthToken, named map[string]bool) (readLength, error) {
	if err := checkToken(token.Name, token.About, named); err != nil {
		return readLength{}, err
	}
	length, err := counted(token.Rem, token.Name, remKey)
	if err != nil {
		return readLength{}, err
	}
	return readLength{name: token.Name, about: token.About, rem: length}, nil
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
