package main

import (
	"errors"
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// A file of tokens with two colours and a blend, its values chosen so each shade is easy to work out
const fixtureTokens = `background:
  about: the page itself
  light: "#000000"
  dark: "#000000"
colours:
  - name: text
    about: what a reader reads
    light: "#767676"
    dark: "#000000"
    basic: own
  - name: accent
    about: what the eye goes to first
    light: "#ffffff"
    dark: "#ff8700"
    basic: 3
blends:
  - name: tint
    about: a result worth seeing first
    over: accent
    light: 0.5
    dark: 0.25
    basic: 3
    indexed_lean: 0.5
mixes:
  - name: soft
    about: the accent softened
    of: accent
    toward: text
    share: 0.2
faces:
  - name: face
    about: the one face
    families: [Mono Face, 3270, Monospace]
    files:
      - weight: 400
        path: fonts/mono-face.woff2
  - name: reading_face
    about: the face words are read in
    families: [Reading Face, serif]
    files:
      - weight: 400
        path: fonts/reading.woff2
      - weight: 400
        italic: true
        path: fonts/reading-italic.woff2
type:
  - name: body
    about: the text read
    size: 1
    line: 1.5
    weight: 400
  - name: reading
    about: words read at length
    face: reading_face
    size: 1.0625
    line: 1.75
    weight: 400
lengths:
  - name: row
    about: the grid's one row
    rem: 1.5
  - name: cell
    about: the grid's one cell
    cells: 1
  - name: tab_stop
    about: how far apart a tab's stops stand
    cells: 4
    terminal: true
  - name: code_in_words
    about: code among words, a share of their size
    em: 0.9375
times:
  - name: blink
    about: a cursor's blink
    milliseconds: 1060
layers:
  - name: row
    about: what sticks to the top
  - name: sheet
    about: what lies over everything
`

// A colour keeps its value whole for a terminal of full colour, falls to the nearest of xterm's 256 for one of
// 256, and to the basic colour the file names for one of sixteen, or to the terminal's own.
func TestEachColourHasAShadeForEachTerminal(test *testing.T) {
	shades := shadesOf(test, fixtureTokens)
	cases := []struct {
		name  string
		got   shade
		want  shade
		theme string
	}{
		{name: "white falls to the cube's white", theme: "light", got: shades["accent"].light, want: shade{trueColour: color.RGBA{R: 255, G: 255, B: 255, A: 255}, indexed: 231, basic: 3}},
		{name: "a colour of the cube is itself", theme: "dark", got: shades["accent"].dark, want: shade{trueColour: color.RGBA{R: 255, G: 135, B: 0, A: 255}, indexed: 208, basic: 3}},
		{name: "a grey falls to the grey ramp", theme: "light", got: shades["text"].light, want: shade{trueColour: color.RGBA{R: 118, G: 118, B: 118, A: 255}, indexed: 243, basic: ownBasic}},
		{name: "black falls to the cube's black", theme: "dark", got: shades["text"].dark, want: shade{trueColour: color.RGBA{R: 0, G: 0, B: 0, A: 255}, indexed: 16, basic: ownBasic}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if testCase.got != testCase.want {
				test.Errorf("the %s shade = %+v, want %+v", testCase.theme, testCase.got, testCase.want)
			}
		})
	}
}

// A blend is its colour laid over the page at its strength in each theme; for 256 colours it leans toward its colour
// first, so a faint blend never rounds to a grey.
func TestABlendIsItsColourLaidOverThePage(test *testing.T) {
	shades := shadesOf(test, fixtureTokens)
	// the fixture with a white page in the dark theme, so what the blend is laid over shows in its shade
	onWhite := shadesOf(test, strings.Replace(fixtureTokens, "  dark: \"#000000\"\ncolours:", "  dark: \"#ffffff\"\ncolours:", 1))
	cases := []struct {
		name string
		got  shade
		want shade
	}{
		{name: "the light theme's strength", got: shades["tint"].light, want: shade{trueColour: color.RGBA{R: 128, G: 128, B: 128, A: 255}, indexed: 250, basic: 3}},
		{name: "the dark theme's strength", got: shades["tint"].dark, want: shade{trueColour: color.RGBA{R: 64, G: 34, B: 0, A: 255}, indexed: 130, basic: 3}},
		{name: "over a page of another colour", got: onWhite["tint"].dark, want: shade{trueColour: color.RGBA{R: 255, G: 225, B: 191, A: 255}, indexed: 215, basic: 3}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if testCase.got != testCase.want {
				test.Errorf("the blend = %+v, want %+v", testCase.got, testCase.want)
			}
		})
	}
}

// The page reads each colour as a custom property, its value in the light theme and the dark one, which theme
// the page's color-scheme says, the reader's system unless one is chosen; a blend is its colour at its strength,
// laid over whatever the page puts under it.
func TestEachColourIsACustomPropertyOfBothThemes(test *testing.T) {
	read, err := readTokens([]byte(fixtureTokens))
	if err != nil {
		test.Fatalf("readTokens(fixture) = %v, want no fault", err)
	}
	written := string(stylesheet(read))
	cases := []struct {
		name string
		want string
	}{
		{name: "the system's theme until one is chosen", want: "color-scheme: light dark;"},
		{name: "the page itself", want: "--background: light-dark(#000000, #000000);"},
		{name: "what the page is", want: "/* the page itself */"},
		{name: "a colour", want: "--accent: light-dark(#ffffff, #ff8700);"},
		{name: "a blend at its strength", want: "--tint: light-dark(#ffffff80, #ff870040);"},
		{name: "what each colour is for", want: "/* what the eye goes to first */"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !strings.Contains(written, testCase.want) {
				test.Errorf("the stylesheet = %q, want it to hold %q", written, testCase.want)
			}
		})
	}
}

// A length a terminal lays out by as well as the page reaches the faces' Go as a number of cells, under what it is
// for; a length of the page alone reaches none.
func TestALengthOfTheTerminalsIsAConstantOfCells(test *testing.T) {
	read, err := readTokens([]byte(fixtureTokens))
	if err != nil {
		test.Fatalf("readTokens(fixture) = %v, want no fault", err)
	}
	written, err := cellsSource(read)
	if err != nil {
		test.Fatalf("cellsSource(fixture) = %v, want no fault", err)
	}
	cases := []struct {
		name string
		want string
		held bool
	}{
		{name: "the length in cells", want: "TabStop = 4", held: true},
		{name: "what it is for", want: "// TabStop is how far apart a tab's stops stand", held: true},
		{name: "a length of the page alone", want: "Cell =", held: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if held := strings.Contains(string(written), testCase.want); held != testCase.held {
				test.Errorf("the Go = %q, want it to hold %q: %t", written, testCase.want, testCase.held)
			}
		})
	}
}

// The page reads its face, each text style and each length as custom properties of the page's root, the same in
// both themes: the face's families each by name, a text style as its size, its line's height and its weight, each
// size in rem.
func TestThePagesOwnTokensAreCustomPropertiesOfItsRoot(test *testing.T) {
	read, err := readTokens([]byte(fixtureTokens))
	if err != nil {
		test.Fatalf("readTokens(fixture) = %v, want no fault", err)
	}
	written := string(stylesheet(read))
	opening := strings.LastIndex(written, ":root {")
	rule := written[opening : opening+strings.Index(written[opening:], "}")]
	cases := []struct {
		name string
		want string
	}{
		{name: "the face, each name quoted but a generic family in any case", want: `--face: "Mono Face", "3270", Monospace;`},
		{name: "another face, under its own name", want: `--reading-face: "Reading Face", serif;`},
		{name: "a mix, its colour moved its share toward the other in each theme", want: "--soft: light-dark(#e4e4e4, #cc6c00);"},
		{name: "what each mix is for", want: "/* the accent softened */"},
		{name: "a text style's size", want: "--type-body-size: 1rem;"},
		{name: "a text style's line, a fraction of a rem written in full", want: "--type-body-line: 1.5rem;"},
		{name: "a text style's weight", want: "--type-body-weight: 400;"},
		{name: "what each of a text style's properties is for", want: "/* the text read: its size */"},
		{name: "a length", want: "--row: 1.5rem;"},
		{name: "a length in cells, a mono face's ch", want: "--cell: 1ch;"},
		{name: "a length that is a share of its text's size, in em", want: "--code-in-words: 0.9375em;"},
		{name: "what each length is for", want: "/* the grid's one row */"},
		{name: "a time, in milliseconds", want: "--blink: 1060ms;"},
		{name: "a layer, its place among the layers counted from one, named apart from a length", want: "--layer-row: 1;"},
		{name: "a layer after it, over it", want: "--layer-sheet: 2;"},
		{name: "what each layer is for", want: "/* what sticks to the top */"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !strings.Contains(rule, testCase.want) {
				test.Errorf("the root's last rule = %q, want it to hold %q", rule, testCase.want)
			}
		})
	}
}

// Each file of a face is a face of its first family that the page loads from the repository, at its weight and slant,
// its path from the tokens' folder as the stylesheet reaches it, its text shown in a stand-in until the file has come.
func TestEachFileOfAFaceIsAFontFaceOfItsFirstFamily(test *testing.T) {
	read, err := readTokens([]byte(fixtureTokens))
	if err != nil {
		test.Fatalf("readTokens(fixture) = %v, want no fault", err)
	}
	written := string(stylesheet(read))
	cases := []struct {
		name string
		want string
	}{
		{
			name: "an upright file of the first face",
			want: `@font-face {
  font-family: "Mono Face";
  src: url("../../../design/fonts/mono-face.woff2") format("woff2");
  font-weight: 400;
  font-style: normal;
  font-display: swap;
}
`,
		},
		{
			name: "an italic file of another face",
			want: `@font-face {
  font-family: "Reading Face";
  src: url("../../../design/fonts/reading-italic.woff2") format("woff2");
  font-weight: 400;
  font-style: italic;
  font-display: swap;
}
`,
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !strings.Contains(written, testCase.want) {
				test.Errorf("the stylesheet = %q, want it to hold %q", written, testCase.want)
			}
		})
	}
}

// Each time reaches the page's script as a number of milliseconds, under what it is for, named as a script names a
// property: a name of two words in camel case.
func TestEachTimeReachesThePagesScriptInMilliseconds(test *testing.T) {
	cases := []struct {
		name   string
		tokens string
		want   string
	}{
		{name: "a time, under what it is for", tokens: fixtureTokens, want: "  /** a cursor's blink */\n  blink: 1060,\n"},
		{name: "a time of two words", tokens: strings.Replace(fixtureTokens, "name: blink", "name: slow_blink", 1), want: "  slowBlink: 1060,\n"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			read, err := readTokens([]byte(testCase.tokens))
			if err != nil {
				test.Fatalf("readTokens = %v, want no fault", err)
			}
			if got := string(timesScript(read)); !strings.Contains(got, testCase.want) {
				test.Errorf("the times' script = %q, want it to hold %q", got, testCase.want)
			}
		})
	}
}

// A name of words parted by underscores is a custom property's, parted by hyphens.
func TestATokenOfTwoWordsIsAPropertyOfTwoWords(test *testing.T) {
	read, err := readTokens([]byte(strings.ReplaceAll(fixtureTokens, ": accent\n", ": on_accent\n")))
	if err != nil {
		test.Fatalf("readTokens = %v, want no fault", err)
	}
	if got := string(stylesheet(read)); !strings.Contains(got, "--on-accent: ") {
		test.Errorf("the stylesheet = %q, want it to hold %q", got, "--on-accent: ")
	}
}

// A word of a name may be a number, as a length counted in the grid's steps is named: row_2 is --row-2.
func TestANameMayHoldANumber(test *testing.T) {
	read, err := readTokens([]byte(strings.Replace(fixtureTokens, "name: row", "name: row_2", 1)))
	if err != nil {
		test.Fatalf("readTokens = %v, want no fault", err)
	}
	if got := string(stylesheet(read)); !strings.Contains(got, "--row-2: 1.5rem;") {
		test.Errorf("the stylesheet = %q, want it to hold %q", got, "--row-2: 1.5rem;")
	}
}

// A length may take a text style's name: the style's properties are named for their kind, so the two never meet.
func TestALengthMayTakeATextStylesName(test *testing.T) {
	read, err := readTokens([]byte(strings.Replace(fixtureTokens, "name: row", "name: body", 1)))
	if err != nil {
		test.Fatalf("readTokens = %v, want no fault", err)
	}
	if got := string(stylesheet(read)); !strings.Contains(got, "--body: 1.5rem;") {
		test.Errorf("the stylesheet = %q, want it to hold %q", got, "--body: 1.5rem;")
	}
}

// A file of tokens with a fault is refused, the fault named: nothing is written from tokens the faces would
// read wrong.
func TestAFaultyFileOfTokensIsRefused(test *testing.T) {
	// the fixture's lines of the length's rem and of the time's milliseconds, which a fault in them names
	remLine := strings.Count(fixtureTokens[:strings.Index(fixtureTokens, "rem: 1.5")], "\n") + 1
	millisecondsLine := strings.Count(fixtureTokens[:strings.Index(fixtureTokens, "milliseconds: 1060")], "\n") + 1
	// the fixture's faces, every line of them, which a test takes away whole
	faces := fixtureTokens[strings.Index(fixtureTokens, "faces:\n"):strings.Index(fixtureTokens, "type:\n")]
	cases := []struct {
		name        string
		original    string
		replacement string
		want        string
	}{
		{name: "a key the file has no place for", original: "    basic: own\n", replacement: "    basic: own\n    weight: 2\n", want: "weight"},
		{name: "a value that is no colour", original: `dark: "#ff8700"`, replacement: `dark: "#ff87"`, want: `"#ff87"`},
		{name: "a basic colour beyond sixteen", original: "    basic: own\n", replacement: "    basic: 16\n", want: "16"},
		{name: "a blend over a colour the file does not name", original: "over: accent", replacement: "over: border", want: `"border"`},
		{name: "a colour named twice", original: "name: text", replacement: "name: accent", want: `"accent" is named twice`},
		{name: "a basic colour left out", original: "    basic: own\n", replacement: "", want: "has no basic"},
		{name: "a colour that says nothing of what it is for", original: "    about: what a reader reads\n", replacement: "", want: "says nothing"},
		{name: "a background that says nothing of what it is for", original: "  about: the page itself\n", replacement: "", want: "background says nothing"},
		{name: "a blend's strength left out", original: "    light: 0.5\n", replacement: "", want: "has no light"},
		{name: "a blend named as a colour is", original: "name: tint", replacement: "name: accent", want: `"accent" is named twice`},
		{name: "a colour with no name", original: "name: text", replacement: `name: ""`, want: "no name of small words"},
		{name: "a name that is no word", original: "name: text", replacement: "name: Text-2", want: `"Text-2"`},
		{name: "a strength beyond the whole", original: "    dark: 0.25\n    basic: 3\n", replacement: "    dark: 1.5\n    basic: 3\n", want: "1.5"},
		{name: "a face with no family", original: "    families: [Mono Face, 3270, Monospace]\n", replacement: "", want: `face "face" has no family`},
		{name: "a page with no face", original: faces, replacement: "faces: []\n", want: "the page has no face"},
		{name: "a face named as a colour is", original: "name: reading_face", replacement: "name: accent", want: `"accent" is named twice`},
		{name: "a text style drawn in no face of the file", original: "face: reading_face\n", replacement: "face: other_face\n", want: `"other_face", which is no face`},
		{
			name: "a text style drawn in a face with no upright file of its weight", original: readingFileLines, replacement: "",
			want: `face "reading_face" has no file of the weight 400`,
		},
		{
			name: "two files of one weight and slant", original: readingFileLines, replacement: readingFileLines + "      - weight: 400\n        italic: true\n        path: fonts/other.woff2\n",
			want: `face "reading_face" has two files of the weight 400, italic true`,
		},
		{name: "a mix of a colour the file does not name", original: "toward: text", replacement: "toward: border", want: `"soft" takes "border", which is no colour`},
		{name: "a mix beyond the whole", original: "share: 0.2", replacement: "share: 1.5", want: `"soft" has the share 1.5`},
		{name: "a mix's share left out", original: "    share: 0.2\n", replacement: "", want: `"soft" has no share`},
		{name: "a mix named as a colour is", original: "name: soft", replacement: "name: text", want: `"text" is named twice`},
		{name: "a family of no name", original: "[Mono Face, 3270, Monospace]", replacement: `["", monospace]`, want: `family "" is no name`},
		{name: "a family that would close its quotes", original: "[Mono Face, 3270, Monospace]", replacement: `["Mono\"; } body {", monospace]`, want: `"Mono\"; } body {"`},
		{name: "a face that says nothing of what it is for", original: "    about: the one face\n", replacement: "", want: `"face" says nothing`},
		{name: "a text style's size left out", original: "    size: 1\n", replacement: "", want: "has no size"},
		{name: "a text style of no size", original: "    size: 1\n", replacement: "    size: 0\n", want: `"body" has the size 0`},
		{name: "a weight between a face's", original: "    weight: 400\n", replacement: "    weight: 450\n", want: "450"},
		{name: "a weight beyond the heaviest", original: "    weight: 400\n", replacement: "    weight: 1000\n", want: "1000"},
		{name: "a weight of none", original: "    weight: 400\n", replacement: "    weight: 0\n", want: "the weight 0"},
		{name: "a length with its unit written, its line named", original: "rem: 1.5", replacement: "rem: 1.5rem", want: fmt.Sprintf(`line %d: "1.5rem" is no number of rem`, remLine)},
		{name: "a length written as a string", original: "rem: 1.5", replacement: `rem: "1.5"`, want: `"1.5" is no number of rem`},
		{name: "a length of no number", original: "rem: 1.5", replacement: "rem: !!float nan", want: `"nan" is no number of rem`},
		{name: "a length of no end", original: "rem: 1.5", replacement: "rem: !!float inf", want: `"inf" is no number of rem`},
		{name: "a length named as a text style's line is", original: "name: row", replacement: "name: type_body_line", want: `"type_body_line" is named twice`},
		{name: "a length named as a text style's weight is", original: "name: row", replacement: "name: type_body_weight", want: `"type_body_weight" is named twice`},
		{name: "a length of no rem", original: "rem: 1.5", replacement: "rem: 0", want: `"row" has the rem 0`},
		{name: "a length less than none", original: "rem: 1.5", replacement: "rem: -1.5", want: `"row" has the rem -1.5`},
		{name: "a length's rem left out", original: "    rem: 1.5\n", replacement: "", want: `"row" has 0 of rem, cells and em`},
		{name: "a length in rem and in cells", original: "    rem: 1.5\n", replacement: "    rem: 1.5\n    cells: 2\n", want: `"row" has 2 of rem, cells and em`},
		{name: "a length in cells and in em", original: "    cells: 1\n", replacement: "    cells: 1\n    em: 1\n", want: `"cell" has 2 of rem, cells and em`},
		{name: "a length of no share of its text", original: "em: 0.9375", replacement: "em: 0", want: `"code_in_words" has the em 0`},
		{name: "a mix under its least contrast", original: "    share: 0.2\n", replacement: "    share: 0.2\n    least_contrast: {ratio: 21, over: [background]}\n", want: `"soft" reads`},
		{name: "a length of no cells", original: "cells: 1", replacement: "cells: 0", want: `"cell" has the cells 0`},
		{name: "a length in rem a terminal lays out by", original: "    rem: 1.5\n", replacement: "    rem: 1.5\n    terminal: true\n", want: `"row" is for a terminal, which lays out in cells alone`},
		{name: "a length of part of a cell", original: "cells: 1", replacement: "cells: 1.5", want: `"1.5" is no whole number`},
		{name: "a length named as a colour is", original: "name: row", replacement: "name: text", want: `"text" is named twice`},
		{name: "a length named as the face's property is", original: "name: row", replacement: "name: face", want: `"face" is named twice`},
		{name: "a length named as a text style's property is", original: "name: row", replacement: "name: type_body_size", want: `"type_body_size" is named twice`},
		{name: "a time that takes none", original: "milliseconds: 1060", replacement: "milliseconds: 0", want: `"blink" has the milliseconds 0`},
		{name: "a time with its unit written, its line named", original: "milliseconds: 1060", replacement: "milliseconds: 1060ms", want: fmt.Sprintf(`line %d: "1060ms" is no whole number`, millisecondsLine)},
		{name: "a time written as a string", original: "milliseconds: 1060", replacement: `milliseconds: "1060"`, want: `"1060" is no whole number`},
		{name: "a time's milliseconds left out", original: "    milliseconds: 1060\n", replacement: "", want: `"blink" has no milliseconds`},
		{name: "a time named as a length is", original: "name: blink", replacement: "name: row", want: `"row" is named twice`},
		{name: "a blend named as the face's property is", original: "name: tint", replacement: "name: face", want: `"face" is named twice`},
		{name: "a weight of part of a hundred", original: "    weight: 400\n", replacement: "    weight: 400.5\n", want: `"400.5" is no whole number`},
		{name: "a blend named as the page's own colour is", original: "name: tint", replacement: "name: background", want: `"background" is named twice`},
		{name: "a layer that says nothing of what it is for", original: "    about: what sticks to the top\n", replacement: "", want: `"row" says nothing`},
		{name: "a layer named twice", original: "name: sheet", replacement: "name: row", want: `"layer_row" is named twice`},
		{name: "a layer whose name is no word", original: "name: sheet", replacement: "name: Sheet-2", want: `"Sheet-2"`},
		{name: "a text style's weight the face has no file of", original: "    weight: 400\n", replacement: "    weight: 700\n", want: "no file of the weight 700"},
		{name: "a file of a weight no text style is drawn in", original: faceFileLines, replacement: faceFileLines + "      - weight: 700\n        path: fonts/mono-face-bold.woff2\n", want: "no text style is drawn in the weight 700"},
		{name: "two files of one weight", original: faceFileLines, replacement: faceFileLines + "      - weight: 400\n        path: fonts/other.woff2\n", want: "two files of the weight 400"},
		{name: "a file's weight between a face's", original: "    - weight: 400\n", replacement: "    - weight: 450\n", want: "450"},
		{name: "a file's weight left out", original: "      - weight: 400\n        path:", replacement: "      - path:", want: "has no weight"},
		{name: "a file of the face that is no woff2", original: "mono-face.woff2", replacement: "mono-face.ttf", want: `"fonts/mono-face.ttf" is no woff2`},
		{name: "a file's path that would close its quotes", original: "fonts/mono-face.woff2", replacement: `'fonts/mono"); } body {.woff2'`, want: `"fonts/mono\"); } body {.woff2"`},
		{name: "a file out of the tokens' folder", original: "fonts/mono-face.woff2", replacement: "../mono-face.woff2", want: `"../mono-face.woff2"`},
		{name: "files of a generic family", original: "[Mono Face, 3270, Monospace]", replacement: "[Monospace, Mono Face]", want: `"Monospace" is a generic family`},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			written := strings.Replace(fixtureTokens, testCase.original, testCase.replacement, 1)
			if written == fixtureTokens {
				test.Fatalf("the fixture holds no %q", testCase.original)
			}
			_, err := readTokens([]byte(written))
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				test.Errorf("readTokens = %v, want a fault naming %q", err, testCase.want)
			}
		})
	}
}

// The fixture's file of its first face, at its weight, after which a test writes another; and the upright file of its
// face for reading
const (
	faceFileLines    = "      - weight: 400\n        path: fonts/mono-face.woff2\n"
	readingFileLines = "      - weight: 400\n        path: fonts/reading.woff2\n"
)

// The fixture's files of its faces, from the repository's root
func faceFilePaths() []string {
	return []string{"design/fonts/mono-face.woff2", "design/fonts/reading.woff2", "design/fonts/reading-italic.woff2"}
}

// The fixture's last lines of its colours text and accent, after which a test writes the least contrast each needs
const (
	textLast   = "    basic: own\n"
	accentLast = "    dark: \"#ff8700\"\n    basic: 3\n"
)

// needing is the fixture with the least contrast its colours text and accent each need written under them.
func needing(test *testing.T, text, accent string) []byte {
	test.Helper()
	return []byte(writtenAfter(test, writtenAfter(test, fixtureTokens, textLast, text), accentLast, accent))
}

// writtenAfter is a file of tokens with lines written after the lines that end a colour.
func writtenAfter(test *testing.T, tokens, last, lines string) string {
	test.Helper()
	if !strings.Contains(tokens, last) {
		test.Fatalf("the fixture holds no %q", last)
	}
	return strings.Replace(tokens, last, last+lines, 1)
}

// leastContrastLine is a least contrast as the file writes it under a colour.
func leastContrastLine(written string) string {
	return "    least_contrast: " + written + "\n"
}

// refusal is what a fault of a colour under its least contrast names: the colour, the ground and the theme.
type refusal struct {
	colour string
	ground string
	theme  string
}

// A colour under its least contrast over a ground it names, in either theme, is refused, the fault naming the
// colour, the ground and the theme; every colour that needs one is checked, and one that reads its least passes.
func TestAColourUnderItsLeastContrastIsRefused(test *testing.T) {
	cases := []struct {
		name   string
		text   string
		accent string
		want   refusal
	}{
		{
			name: "a grey over black in the dark theme", text: leastContrastLine("{ratio: 3, over: [background]}"),
			want: refusal{colour: "text", ground: "background", theme: darkKey},
		},
		{
			name: "a grey over black in the light theme", text: leastContrastLine("{ratio: 4.7, over: [background]}"),
			want: refusal{colour: "text", ground: "background", theme: lightKey},
		},
		{
			name: "a colour that needs one after a colour that needs none", accent: leastContrastLine("{ratio: 21, over: [background]}"),
			want: refusal{colour: "accent", ground: "background", theme: darkKey},
		},
		{
			name: "the second colour that needs one, the first passing, a ratio of 21:1 taken",
			text: leastContrastLine("{ratio: 1.5, over: [accent]}"), accent: leastContrastLine("{ratio: 21, over: [background]}"),
			want: refusal{colour: "accent", ground: "background", theme: darkKey},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			_, err := readTokens(needing(test, testCase.text, testCase.accent))
			var fault underContrast
			if !errors.As(err, &fault) {
				test.Fatalf("readTokens = %v, want %q under its least contrast", err, testCase.want.colour)
			}
			if got := (refusal{colour: fault.colour, ground: fault.ground, theme: fault.theme}); got != testCase.want {
				test.Errorf("readTokens refused %+v, want %+v", got, testCase.want)
			}
		})
	}
}

// A colour that reads its least contrast over each ground it names, each theme's over that theme's ground, passes:
// the fixture's accent over text, which reads in the dark theme only over the dark theme's black.
func TestAColourAtItsLeastContrastPasses(test *testing.T) {
	if _, err := readTokens(needing(test, "", leastContrastLine("{ratio: 4.5, over: [background, text]}"))); err != nil {
		test.Errorf("readTokens = %v, want no fault", err)
	}
}

// A least contrast written wrong is refused: over a ground the file does not name, over none, and a ratio no two
// colours have, left out among them.
func TestALeastContrastWrittenWrongIsRefused(test *testing.T) {
	cases := []struct {
		name string
		text string
		want error
	}{
		{name: "a ground the file does not name", text: "{ratio: 3, over: [border]}", want: errUnknownGround},
		{name: "no ground", text: "{ratio: 3, over: []}", want: errNoGround},
		{name: "a ratio of none", text: "{ratio: 1, over: [accent]}", want: errRatioBeyond},
		{name: "a ratio beyond black on white", text: "{ratio: 22, over: [accent]}", want: errRatioBeyond},
		{name: "a ratio left out", text: "{over: [accent]}", want: errMissing},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			_, err := readTokens(needing(test, leastContrastLine(testCase.text), ""))
			if !errors.Is(err, testCase.want) {
				test.Errorf("readTokens = %v, want %v", err, testCase.want)
			}
		})
	}
}

// Contrast is read as WCAG 2 reads it, whichever colour is first: on the line of sRGB near black, on its curve above.
func TestContrastIsWCAGs(test *testing.T) {
	// the contrast is read to the hundredth, as WCAG's examples write it
	const hundredths = 100
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	cases := []struct {
		name   string
		first  color.RGBA
		second color.RGBA
		want   float64
	}{
		{name: "white on black", first: white, second: black, want: 21},
		{name: "black on white", first: black, second: white, want: 21},
		{name: "the grey that reads 4.54:1 on white", first: color.RGBA{R: 118, G: 118, B: 118, A: 255}, second: white, want: 4.54},
		{name: "a dark grey on black, on the curve", first: color.RGBA{R: 51, G: 51, B: 51, A: 255}, second: black, want: 1.66},
		{name: "a near black on black, on the line", first: color.RGBA{R: 10, G: 10, B: 10, A: 255}, second: black, want: 1.06},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := math.Round(contrast(testCase.first, testCase.second)*hundredths) / hundredths; got != testCase.want {
				test.Errorf("contrast(%v, %v) = %v, want %v", testCase.first, testCase.second, got, testCase.want)
			}
		})
	}
}

// The check finds a generated file out of step with its tokens, the palette or the stylesheet, and passes the
// files the tokens make.
func TestTheCheckFindsAGeneratedFileOutOfStep(test *testing.T) {
	root := fixtureRoot(test)
	for _, path := range faceFilePaths() {
		writeFixture(test, filepath.Join(root, path), "")
	}
	if code := run([]string{writeFlag}, root); code != succeeded {
		test.Fatalf("writing = exit %d, want %d", code, succeeded)
	}
	cases := []struct {
		name   string
		path   string
		change func(generated string) string
		want   int
	}{
		{name: "the files the tokens make", path: palettePath, change: func(generated string) string { return generated }, want: succeeded},
		{name: "a palette edited by hand", path: palettePath, change: func(generated string) string { return generated + "\n// edited\n" }, want: failed},
		{name: "a stylesheet edited by hand", path: stylesheetPath, change: func(generated string) string { return generated + "\n/* edited */\n" }, want: failed},
		{name: "the terminal's lengths edited by hand", path: cellsPath, change: func(generated string) string { return generated + "\n// edited\n" }, want: failed},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			generatedPath := filepath.Join(root, testCase.path)
			generated, err := os.ReadFile(generatedPath)
			if err != nil {
				test.Fatalf("read %s: %v", generatedPath, err)
			}
			writeFixture(test, generatedPath, testCase.change(string(generated)))
			if code := run(nil, root); code != testCase.want {
				test.Errorf("the check = exit %d, want %d", code, testCase.want)
			}
			writeFixture(test, generatedPath, string(generated))
		})
	}
}

// A file of a face the tokens name and the repository lacks is refused, any face's, and nothing is written: the page
// would ask for a file the build never holds, and show a stand-in face.
func TestAFileOfAFaceMissingIsRefused(test *testing.T) {
	for _, missing := range faceFilePaths() {
		test.Run(missing, func(test *testing.T) {
			root := fixtureRoot(test)
			for _, path := range slices.DeleteFunc(faceFilePaths(), func(path string) bool { return path == missing }) {
				writeFixture(test, filepath.Join(root, path), "")
			}
			if code := run([]string{writeFlag}, root); code != failed {
				test.Errorf("writing with no %s = exit %d, want %d", missing, code, failed)
			}
			if _, err := os.Stat(filepath.Join(root, stylesheetPath)); !errors.Is(err, os.ErrNotExist) {
				test.Errorf("stat %s = %v, want it not written", stylesheetPath, err)
			}
		})
	}
}

// fixtureRoot is a repository holding the fixture's tokens and the folders the generated files go to, and no file of
// a face.
func fixtureRoot(test *testing.T) string {
	test.Helper()
	root := test.TempDir()
	writeFixture(test, filepath.Join(root, tokensPath), fixtureTokens)
	for _, folder := range []string{filepath.Dir(palettePath), filepath.Dir(stylesheetPath)} {
		if err := os.MkdirAll(filepath.Join(root, folder), 0o750); err != nil {
			test.Fatalf("make the folder %s: %v", folder, err)
		}
	}
	return root
}

// shadesOf are the shades of a file's colours and blends, by name.
func shadesOf(test *testing.T, tokens string) map[string]themedShade {
	read, err := readTokens([]byte(tokens))
	if err != nil {
		test.Fatalf("readTokens(fixture) = %v, want no fault", err)
	}
	shades := map[string]themedShade{}
	for _, worked := range paletteOf(read) {
		shades[worked.name] = worked.shades
	}
	return shades
}

func writeFixture(test *testing.T, path, text string) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		test.Fatalf("make the folder of %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		test.Fatalf("write %s: %v", path, err)
	}
}
