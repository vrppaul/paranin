package main

import (
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A file of tokens with one colour and one blend over it, its values chosen so each shade is easy to work out
const fixtureTokens = `background:
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

// A blend is its colour laid over the page at its strength in each theme; for 256 colours it leans toward its
// colour first, so a faint blend never rounds to a grey.
func TestABlendIsItsColourLaidOverThePage(test *testing.T) {
	tint := fixtureShades(test)["tint"]
	cases := []struct {
		name string
		got  shade
		want shade
	}{
		{name: "the light theme's strength", got: tint.light, want: shade{trueColour: color.RGBA{R: 128, G: 128, B: 128, A: 255}, indexed: 250, basic: 3}},
		{name: "the dark theme's strength", got: tint.dark, want: shade{trueColour: color.RGBA{R: 64, G: 34, B: 0, A: 255}, indexed: 130, basic: 3}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if testCase.got != testCase.want {
				test.Errorf("the tint = %+v, want %+v", testCase.got, testCase.want)
			}
		})
	}
}

// A file of tokens with a fault is refused, the fault named: nothing is written from tokens the faces would
// read wrong.
func TestAFaultyFileOfTokensIsRefused(test *testing.T) {
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
		{name: "a blend's strength left out", original: "    light: 0.5\n", replacement: "", want: "has no light"},
		{name: "a blend named as a colour is", original: "name: tint", replacement: "name: accent", want: `"accent" is named twice`},
		{name: "a colour with no name", original: "name: text", replacement: `name: ""`, want: "no name of small words"},
		{name: "a name that is no word", original: "name: text", replacement: "name: Text-2", want: `"Text-2"`},
		{name: "a strength beyond the whole", original: "    dark: 0.25\n    basic: 3\n", replacement: "    dark: 1.5\n    basic: 3\n", want: "1.5"},
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

// The check finds a generated file out of step with its tokens, and passes the file the tokens make.
func TestTheCheckFindsAGeneratedFileOutOfStep(test *testing.T) {
	root := test.TempDir()
	writeFixture(test, filepath.Join(root, tokensPath), fixtureTokens)
	if err := os.MkdirAll(filepath.Join(root, filepath.Dir(palettePath)), 0o750); err != nil {
		test.Fatalf("make the folder of the palette: %v", err)
	}
	if code := run([]string{writeFlag}, root); code != succeeded {
		test.Fatalf("writing = exit %d, want %d", code, succeeded)
	}
	cases := []struct {
		name   string
		change func(generated string) string
		want   int
	}{
		{name: "the file the tokens make", change: func(generated string) string { return generated }, want: succeeded},
		{name: "a file edited by hand", change: func(generated string) string { return generated + "\n// edited\n" }, want: failed},
	}
	generatedPath := filepath.Join(root, palettePath)
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
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
