package main

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A file of tokens with two colours, a blend over the page and one over a colour, its values chosen so each shade
// is easy to work out
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
  - name: chosen
    about: an item chosen
    over: text
    under: accent
    light: 0.5
    dark: 0.5
    basic: own
    indexed_lean: 0
face:
  about: the one face
  families: [Mono Face, 3270, Monospace]
type:
  - name: body
    about: the text read
    size: 16
    line: 24
    weight: 400
lengths:
  - name: row
    about: the grid's one row
    pixels: 24
times:
  - name: blink
    about: a cursor's blink
    milliseconds: 1060
`

// A colour keeps its value whole for a terminal of full colour, falls to the nearest of xterm's 256 for one of
// 256, and to the basic colour the file names for one of sixteen, or to the terminal's own.
func TestEachColourHasAShadeForEachTerminal(test *testing.T) {
	shades := fixtureShades(test)
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

// A blend is its colour laid over the page at its strength in each theme, or over the colour it names to lie
// under it; for 256 colours it leans toward its colour first, so a faint blend never rounds to a grey.
func TestABlendIsItsColourLaidOverThePage(test *testing.T) {
	shades := fixtureShades(test)
	cases := []struct {
		name string
		got  shade
		want shade
	}{
		{name: "the light theme's strength", got: shades["tint"].light, want: shade{trueColour: color.RGBA{R: 128, G: 128, B: 128, A: 255}, indexed: 250, basic: 3}},
		{name: "the dark theme's strength", got: shades["tint"].dark, want: shade{trueColour: color.RGBA{R: 64, G: 34, B: 0, A: 255}, indexed: 130, basic: 3}},
		{name: "over a colour in the light theme", got: shades["chosen"].light, want: shade{trueColour: color.RGBA{R: 187, G: 187, B: 187, A: 255}, indexed: 250, basic: ownBasic}},
		{name: "over a colour in the dark theme", got: shades["chosen"].dark, want: shade{trueColour: color.RGBA{R: 128, G: 68, B: 0, A: 255}, indexed: 94, basic: ownBasic}},
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
		{name: "a blend laid over the page, whatever it lies over in a terminal", want: "--chosen: light-dark(#76767680, #00000080);"},
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

// The page reads its face, each text style and each length as custom properties of the page's root, the same in
// both themes: the face's families each by name, a text style as its size, its line's height and its weight.
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
		{name: "a text style's size", want: "--type-body-size: 16px;"},
		{name: "a text style's line", want: "--type-body-line: 24px;"},
		{name: "a text style's weight", want: "--type-body-weight: 400;"},
		{name: "what each of a text style's properties is for", want: "/* the text read: its size */"},
		{name: "a length", want: "--row: 24px;"},
		{name: "what each length is for", want: "/* the grid's one row */"},
		{name: "a time, in milliseconds", want: "--blink: 1060ms;"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !strings.Contains(rule, testCase.want) {
				test.Errorf("the root's last rule = %q, want it to hold %q", rule, testCase.want)
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

// A length may take a text style's name: the style's properties are named for their kind, so the two never meet.
func TestALengthMayTakeATextStylesName(test *testing.T) {
	read, err := readTokens([]byte(strings.Replace(fixtureTokens, "name: row", "name: body", 1)))
	if err != nil {
		test.Fatalf("readTokens = %v, want no fault", err)
	}
	if got := string(stylesheet(read)); !strings.Contains(got, "--body: 24px;") {
		test.Errorf("the stylesheet = %q, want it to hold %q", got, "--body: 24px;")
	}
}

// A file of tokens with a fault is refused, the fault named: nothing is written from tokens the faces would
// read wrong.
func TestAFaultyFileOfTokensIsRefused(test *testing.T) {
	// the fixture's line of the length's pixels, which a fault in them names
	pixelsLine := strings.Count(fixtureTokens[:strings.Index(fixtureTokens, "pixels: 24")], "\n") + 1
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
		{name: "a blend under a colour the file does not name", original: "under: accent", replacement: "under: border", want: `"border"`},
		{name: "a colour named twice", original: "name: text", replacement: "name: accent", want: `"accent" is named twice`},
		{name: "a basic colour left out", original: "    basic: own\n", replacement: "", want: "has no basic"},
		{name: "a colour that says nothing of what it is for", original: "    about: what a reader reads\n", replacement: "", want: "says nothing"},
		{name: "a background that says nothing of what it is for", original: "  about: the page itself\n", replacement: "", want: "background says nothing"},
		{name: "a blend's strength left out", original: "    light: 0.5\n", replacement: "", want: "has no light"},
		{name: "a blend named as a colour is", original: "name: tint", replacement: "name: accent", want: `"accent" is named twice`},
		{name: "a colour with no name", original: "name: text", replacement: `name: ""`, want: "no name of small words"},
		{name: "a name that is no word", original: "name: text", replacement: "name: Text-2", want: `"Text-2"`},
		{name: "a strength beyond the whole", original: "    dark: 0.25\n    basic: 3\n", replacement: "    dark: 1.5\n    basic: 3\n", want: "1.5"},
		{name: "a face with no family", original: "  families: [Mono Face, 3270, Monospace]\n", replacement: "", want: "face has no family"},
		{name: "a family of no name", original: "[Mono Face, 3270, Monospace]", replacement: `["", monospace]`, want: `family "" is no name`},
		{name: "a family that would close its quotes", original: "[Mono Face, 3270, Monospace]", replacement: `["Mono\"; } body {", monospace]`, want: `"Mono\"; } body {"`},
		{name: "a face that says nothing of what it is for", original: "  about: the one face\n", replacement: "", want: `"face" says nothing`},
		{name: "a text style's size left out", original: "    size: 16\n", replacement: "", want: "has no size"},
		{name: "a text style of no size", original: "    size: 16\n", replacement: "    size: 0\n", want: `"body" has the size 0`},
		{name: "a weight between a face's", original: "    weight: 400\n", replacement: "    weight: 450\n", want: "450"},
		{name: "a weight beyond the heaviest", original: "    weight: 400\n", replacement: "    weight: 1000\n", want: "1000"},
		{name: "a weight of none", original: "    weight: 400\n", replacement: "    weight: 0\n", want: "the weight 0"},
		{name: "a length of part of a pixel, its line named", original: "pixels: 24", replacement: "pixels: 1.5", want: fmt.Sprintf(`line %d: "1.5" is no whole number`, pixelsLine)},
		{name: "a length with its unit written", original: "pixels: 24", replacement: "pixels: 24px", want: `"24px" is no whole number`},
		{name: "a length written as a string", original: "pixels: 24", replacement: `pixels: "24"`, want: `"24" is no whole number`},
		{name: "a length named as a text style's line is", original: "name: row", replacement: "name: type_body_line", want: `"type_body_line" is named twice`},
		{name: "a length named as a text style's weight is", original: "name: row", replacement: "name: type_body_weight", want: `"type_body_weight" is named twice`},
		{name: "a length of no pixels", original: "pixels: 24", replacement: "pixels: 0", want: `"row" has the pixels 0`},
		{name: "a length's pixels left out", original: "    pixels: 24\n", replacement: "", want: `"row" has no pixels`},
		{name: "a length named as a colour is", original: "name: row", replacement: "name: text", want: `"text" is named twice`},
		{name: "a length named as the face's property is", original: "name: row", replacement: "name: face", want: `"face" is named twice`},
		{name: "a length named as a text style's property is", original: "name: row", replacement: "name: type_body_size", want: `"type_body_size" is named twice`},
		{name: "a time that takes none", original: "milliseconds: 1060", replacement: "milliseconds: 0", want: `"blink" has the milliseconds 0`},
		{name: "a time's milliseconds left out", original: "    milliseconds: 1060\n", replacement: "", want: `"blink" has no milliseconds`},
		{name: "a time named as a length is", original: "name: blink", replacement: "name: row", want: `"row" is named twice`},
		{name: "a blend named as the face's property is", original: "name: chosen", replacement: "name: face", want: `"face" is named twice`},
		{name: "a weight of part of a hundred", original: "    weight: 400\n", replacement: "    weight: 400.5\n", want: `"400.5" is no whole number`},
		{name: "a blend named as the page's own colour is", original: "name: tint", replacement: "name: background", want: `"background" is named twice`},
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

// The check finds a generated file out of step with its tokens, the palette or the stylesheet, and passes the
// files the tokens make.
func TestTheCheckFindsAGeneratedFileOutOfStep(test *testing.T) {
	root := test.TempDir()
	writeFixture(test, filepath.Join(root, tokensPath), fixtureTokens)
	for _, folder := range []string{filepath.Dir(palettePath), filepath.Dir(stylesheetPath)} {
		if err := os.MkdirAll(filepath.Join(root, folder), 0o750); err != nil {
			test.Fatalf("make the folder %s: %v", folder, err)
		}
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

// fixtureShades are the shades of the fixture's colours and blends, by name.
func fixtureShades(test *testing.T) map[string]themedShade {
	read, err := readTokens([]byte(fixtureTokens))
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
