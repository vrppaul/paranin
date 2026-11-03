package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"image/color"
	"math"
	"slices"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

// How a colour is written in the file: a hash, then two hexadecimal digits for each of red, green and blue
const (
	colourMark   = "#"
	channelCount = 3
	redChannel   = 0
	greenChannel = 1
	blueChannel  = 2
	opaque       = math.MaxUint8
)

// The strength of a blend that lays its colour over the page whole
const strengthOfWhole = 1

// The basic colours a terminal of sixteen names, and the word for the terminal's own colour
const (
	basicColours = 16
	ownWord      = "own"
	// a basic colour is written in decimal, and read into a byte
	decimal   = 10
	basicBits = 8
	// the keys a basic colour, a blend's strength in each theme and its lean are written under, as their tags say
	basicKey = "basic"
	lightKey = "light"
	darkKey  = "dark"
	leanKey  = "indexed_lean"
	// ownBasic is the terminal's own colour: a number no colour of sixteen has; the generated file names it
	ownBasic = math.MaxUint8
)

// The faults a file of tokens can hold
var errTokens = errors.New("the tokens have a fault")

// tokensFile is the file as it is written.
type tokensFile struct {
	Background backgroundToken `yaml:"background"`
	Colours    []colourToken   `yaml:"colours"`
	Blends     []blendToken    `yaml:"blends"`
}

// themedValue is one value in each theme.
type themedValue[Value any] struct {
	Light Value `yaml:"light"`
	Dark  Value `yaml:"dark"`
}

// backgroundToken is the page's own colour as the file writes it: what it is for, and its value in each theme.
type backgroundToken struct {
	About string `yaml:"about"`
	Light string `yaml:"light"`
	Dark  string `yaml:"dark"`
}

// colourToken is a colour as the file writes it: its name, what it is for, its value in each theme, and the
// basic colour it falls back to. A value the file may leave out unseen is a pointer, so its absence is a fault.
type colourToken struct {
	Name  string      `yaml:"name"`
	About string      `yaml:"about"`
	Light string      `yaml:"light"`
	Dark  string      `yaml:"dark"`
	Basic *basicValue `yaml:"basic"`
}

// blendToken is a blend as the file writes it: the colour it lays, the colour it lies over, or the page where it
// names none, its strength in each theme, the basic colour it falls back to, and how far it leans toward its colour
// for 256 colours.
type blendToken struct {
	Name        string      `yaml:"name"`
	About       string      `yaml:"about"`
	Over        string      `yaml:"over"`
	Under       string      `yaml:"under"`
	Light       *float64    `yaml:"light"`
	Dark        *float64    `yaml:"dark"`
	Basic       *basicValue `yaml:"basic"`
	IndexedLean *float64    `yaml:"indexed_lean"`
}

// basicValue is one of a terminal's sixteen colours, or ownBasic.
type basicValue uint8

// UnmarshalYAML reads a basic colour: a number below sixteen, or the word for the terminal's own.
func (basic *basicValue) UnmarshalYAML(node *yaml.Node) error {
	if node.Value == ownWord {
		*basic = ownBasic
		return nil
	}
	number, err := strconv.ParseUint(node.Value, decimal, basicBits)
	if err != nil || number >= basicColours {
		return fmt.Errorf("%w: basic %q is neither %s nor a colour of 0 to %d", errTokens, node.Value, ownWord, basicColours-1)
	}
	*basic = basicValue(number)
	return nil
}

// tokens are the colours and blends of the file read: every value there, each a colour or a strength, every name
// a word of its own, every blend over a colour the file names.
type tokens struct {
	background      themedValue[color.RGBA]
	backgroundAbout string
	colours         []readColour
	blends          []readBlend
}

// readColour is a colour of the file with its values read.
type readColour struct {
	name   string
	about  string
	values themedValue[color.RGBA]
	basic  basicValue
}

// readBlend is a blend of the file with its colours found, the one it lies over the page's where it names none, and
// its strengths read.
type readBlend struct {
	name        string
	about       string
	over        themedValue[color.RGBA]
	under       themedValue[color.RGBA]
	strengths   themedValue[float64]
	basic       basicValue
	indexedLean float64
}

// readTokens reads the file strictly: a key it has no place for or one it leaves out, a value that is no colour,
// a strength beyond the whole, a name that is no word or written twice, and a blend over a colour the file does
// not name are each a fault.
func readTokens(data []byte) (tokens, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	var written tokensFile
	if err := decoder.Decode(&written); err != nil {
		return tokens{}, fmt.Errorf("%w: %w", errTokens, err)
	}
	background, err := themedColour(written.Background.Light, written.Background.Dark)
	if err != nil {
		return tokens{}, fmt.Errorf("background: %w", err)
	}
	if written.Background.About == "" {
		return tokens{}, fmt.Errorf("%w: the background says nothing of what it is for", errTokens)
	}
	named := map[string]bool{}
	colours := make([]readColour, 0, len(written.Colours))
	for _, token := range written.Colours {
		colour, err := readColourToken(token, named)
		if err != nil {
			return tokens{}, err
		}
		colours = append(colours, colour)
	}
	blends := make([]readBlend, 0, len(written.Blends))
	for _, token := range written.Blends {
		blend, err := readBlendToken(token, background, colours, named)
		if err != nil {
			return tokens{}, err
		}
		blends = append(blends, blend)
	}
	return tokens{background: background, backgroundAbout: written.Background.About, colours: colours, blends: blends}, nil
}

func readColourToken(token colourToken, named map[string]bool) (readColour, error) {
	if err := checkToken(token.Name, token.About, named); err != nil {
		return readColour{}, err
	}
	values, err := themedColour(token.Light, token.Dark)
	if err != nil {
		return readColour{}, fmt.Errorf("colour %q: %w", token.Name, err)
	}
	basic, err := required(token.Basic, token.Name, basicKey)
	if err != nil {
		return readColour{}, err
	}
	return readColour{name: token.Name, about: token.About, values: values, basic: basic}, nil
}

func readBlendToken(token blendToken, page themedValue[color.RGBA], colours []readColour, named map[string]bool) (readBlend, error) {
	if err := checkToken(token.Name, token.About, named); err != nil {
		return readBlend{}, err
	}
	over, overErr := namedColour(colours, token.Over, token.Name, "lays")
	under := page
	var underErr error
	if token.Under != "" {
		under, underErr = namedColour(colours, token.Under, token.Name, "lies over")
	}
	light, lightErr := strength(token.Light, token.Name, lightKey)
	dark, darkErr := strength(token.Dark, token.Name, darkKey)
	lean, leanErr := strength(token.IndexedLean, token.Name, leanKey)
	basic, basicErr := required(token.Basic, token.Name, basicKey)
	if err := errors.Join(overErr, underErr, lightErr, darkErr, leanErr, basicErr); err != nil {
		return readBlend{}, err
	}
	return readBlend{
		name:        token.Name,
		about:       token.About,
		over:        over,
		under:       under,
		strengths:   themedValue[float64]{Light: light, Dark: dark},
		basic:       basic,
		indexedLean: lean,
	}, nil
}

// namedColour is the values of the colour of a name a blend names, as it says what it does with it, or a fault
// where the file has no colour of that name.
func namedColour(colours []readColour, name, blend, does string) (themedValue[color.RGBA], error) {
	found := slices.IndexFunc(colours, func(known readColour) bool { return known.name == name })
	if found < 0 {
		return themedValue[color.RGBA]{}, fmt.Errorf("%w: the blend %q %s %q, which is no colour of the file", errTokens, blend, does, name)
	}
	return colours[found].values, nil
}

// checkToken refuses a token whose name is no word, or words parted by underscores, or is one the file has named
// already, and one that says nothing of what it is for; it marks the name as given.
func checkToken(name, about string, named map[string]bool) error {
	if !isTokenName(name) {
		return fmt.Errorf("%w: %q is no name of small words parted by underscores", errTokens, name)
	}
	if named[name] {
		return fmt.Errorf("%w: %q is named twice", errTokens, name)
	}
	if about == "" {
		return fmt.Errorf("%w: %q says nothing of what it is for", errTokens, name)
	}
	named[name] = true
	return nil
}

// isTokenName is whether a name is small words parted by underscores: on_accent.
func isTokenName(name string) bool {
	for word := range strings.SplitSeq(name, wordSeparator) {
		if word == "" || strings.ContainsFunc(word, func(letter rune) bool { return letter < 'a' || letter > 'z' }) {
			return false
		}
	}
	return true
}

// required is a value the file must write, or a fault naming the token and the key it left out.
func required[Value any](written *Value, token, key string) (Value, error) {
	if written == nil {
		var none Value
		return none, fmt.Errorf("%w: %q has no %s", errTokens, token, key)
	}
	return *written, nil
}

// strength is a strength the file must write, from nothing to the whole.
func strength(written *float64, token, key string) (float64, error) {
	value, err := required(written, token, key)
	if err != nil {
		return 0, err
	}
	if value < 0 || value > strengthOfWhole {
		return 0, fmt.Errorf("%w: %q has the %s %v, beyond 0 to 1", errTokens, token, key, value)
	}
	return value, nil
}

func themedColour(light, dark string) (themedValue[color.RGBA], error) {
	lightColour, err := parseColour(light)
	if err != nil {
		return themedValue[color.RGBA]{}, err
	}
	darkColour, err := parseColour(dark)
	if err != nil {
		return themedValue[color.RGBA]{}, err
	}
	return themedValue[color.RGBA]{Light: lightColour, Dark: darkColour}, nil
}

// parseColour reads a colour written #rrggbb.
func parseColour(written string) (color.RGBA, error) {
	digits, marked := strings.CutPrefix(written, colourMark)
	channels, err := hex.DecodeString(digits)
	if !marked || len(channels) != channelCount || err != nil {
		return color.RGBA{}, fmt.Errorf("%w: %q is no colour written #rrggbb", errTokens, written)
	}
	return color.RGBA{R: channels[redChannel], G: channels[greenChannel], B: channels[blueChannel], A: opaque}, nil
}
