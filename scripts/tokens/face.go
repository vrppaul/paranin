package main

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
)

// The face's own name, the property every rule of the page's text inherits it from
const faceName = "face"

// How a family is written in the stylesheet: its name quoted, unless it is one of CSS's generic families, a word
// CSS knows, which quoted would name a face of that name; the names parted by commas
const (
	familyQuote     = `"`
	familySeparator = ", "
)

// The weights a face is drawn in: a hundred to nine hundred, in steps of a hundred
const (
	lightestWeight = 100
	heaviestWeight = 900
	weightStep     = 100
)

// A file of the face is a woff2: every browser the page is for reads it, and it is the smallest of a face's formats
const faceFileEnding = ".woff2"

// How the files of the face open in the stylesheet, after the page's own tokens
const faceFilesOpening = `
/* The face's files, which the site serves itself, so every reader sees the copy its cells are counted in */
`

// How a file of the face is written in the stylesheet: a face of its family at its weight, from its path. Its text
// shows in a stand-in until the file has come, then in it (swap): every stand-in is monospace and every line's height
// is in rem, so the swap redraws the letters in rows as tall as before, where block would hide the text while the
// file comes and optional would keep a reader on a slow line in the stand-in
const fontFaceRule = `@font-face {
  font-family: %s;
  src: url("%s") format("woff2");
  font-weight: %d;
  font-display: swap;
}
`

// genericFamilies are CSS's generic families, which the stylesheet writes unquoted.
func genericFamilies() []string {
	return []string{
		"serif", "sans-serif", "monospace", "cursive", "fantasy", "math", "emoji", "fangsong",
		"system-ui", "ui-serif", "ui-sans-serif", "ui-monospace", "ui-rounded",
	}
}

// faceToken is the page's one face as the file writes it: what it is for, the families a browser tries in turn,
// each by its name alone, and the files of the first, which the site serves.
type faceToken struct {
	About    string          `yaml:"about"`
	Families []string        `yaml:"families"`
	Files    []faceFileToken `yaml:"files"`
}

// faceFileToken is a file of the face as the file writes it: the weight it draws, and its path from the tokens'
// folder. A number the file may leave out unseen is a pointer, so its absence is a fault.
type faceFileToken struct {
	Weight *wholeNumber `yaml:"weight"`
	Path   string       `yaml:"path"`
}

// readFace is the face read: what it is for, its families, each a name of letters, digits, hyphens and spaces, and
// the files of the first.
type readFace struct {
	about    string
	families []string
	files    []readFaceFile
}

// readFaceFile is a file of the face read: the weight it draws, its path from the tokens' folder, inside it, and the
// same path from the repository's root and as the stylesheet reaches it.
type readFaceFile struct {
	weight         wholeNumber
	path           string
	fromRoot       string
	fromStylesheet string
}

// readFaceToken refuses a face that says nothing of what it is for, a family written wrong, and a file written wrong
// or of a generic family, which CSS gives no face of its own.
func readFaceToken(token faceToken, named map[string]bool) (readFace, error) {
	if err := checkToken(faceName, token.About, named); err != nil {
		return readFace{}, err
	}
	if err := checkFamilies(token.Families); err != nil {
		return readFace{}, err
	}
	if first := token.Families[0]; len(token.Files) > 0 && isGenericFamily(first) {
		return readFace{}, fmt.Errorf("%w: %q is a generic family, which no file is a face of", errTokens, first)
	}
	files := make([]readFaceFile, 0, len(token.Files))
	for _, written := range token.Files {
		file, err := readFaceFileToken(written)
		if err != nil {
			return readFace{}, err
		}
		files = append(files, file)
	}
	return readFace{about: token.About, families: token.Families, files: files}, nil
}

// checkFamilies refuses a face that names no family, or names one in anything but letters, digits, hyphens and
// spaces, so no name can close the quotes the stylesheet writes it in.
func checkFamilies(families []string) error {
	if len(families) == 0 {
		return fmt.Errorf("%w: the face has no family", errTokens)
	}
	for _, family := range families {
		if family == "" || strings.ContainsFunc(family, isNoFamilyLetter) {
			return fmt.Errorf("%w: the face's family %q is no name of letters, digits, hyphens and spaces", errTokens, family)
		}
	}
	return nil
}

// readFaceFileToken refuses a file of no weight a face is drawn in, one that is no woff2, and one whose path leaves
// the tokens' folder or holds anything but letters, digits, hyphens, underscores, dots and slashes, so no path can
// close the quotes the stylesheet writes it in.
func readFaceFileToken(token faceFileToken) (readFaceFile, error) {
	weight, err := required(token.Weight, token.Path, weightKey)
	if err != nil {
		return readFaceFile{}, err
	}
	if err := checkWeight(weight, token.Path); err != nil {
		return readFaceFile{}, err
	}
	if !fs.ValidPath(token.Path) || strings.ContainsFunc(token.Path, isNoPathLetter) {
		return readFaceFile{}, fmt.Errorf("%w: the face's file %q is no path inside the tokens' folder of letters, digits, "+
			"hyphens, underscores, dots and slashes", errTokens, token.Path)
	}
	if path.Ext(token.Path) != faceFileEnding {
		return readFaceFile{}, fmt.Errorf("%w: the face's file %q is no woff2", errTokens, token.Path)
	}
	fromRoot := path.Join(path.Dir(tokensPath), token.Path)
	fromStylesheet, err := filepath.Rel(path.Dir(stylesheetPath), fromRoot)
	if err != nil {
		return readFaceFile{}, fmt.Errorf("%w: the face's file %q from %s: %w", errTokens, token.Path, stylesheetPath, err)
	}
	return readFaceFile{weight: weight, path: token.Path, fromRoot: fromRoot, fromStylesheet: filepath.ToSlash(fromStylesheet)}, nil
}

// isNoFamilyLetter is whether a letter has no place in a family's name.
func isNoFamilyLetter(letter rune) bool {
	return !unicode.IsLetter(letter) && !unicode.IsDigit(letter) && letter != '-' && letter != ' '
}

// isNoPathLetter is whether a letter has no place in the path of a file of the face.
func isNoPathLetter(letter rune) bool {
	return !unicode.IsLetter(letter) && !unicode.IsDigit(letter) && !strings.ContainsRune("-_./", letter)
}

// isGenericFamily is whether a family is one of CSS's generic families, which CSS knows by its bare word in any case.
func isGenericFamily(family string) bool {
	return slices.Contains(genericFamilies(), strings.ToLower(family))
}

// checkWeight refuses a weight no face is drawn in, the fault naming what it is the weight of.
func checkWeight(weight wholeNumber, owner string) error {
	if weight < lightestWeight || weight > heaviestWeight || weight%weightStep != 0 {
		return fmt.Errorf("%w: %q has the weight %d, no weight of %d to %d in steps of %d",
			errTokens, owner, weight, lightestWeight, heaviestWeight, weightStep)
	}
	return nil
}

// checkFaceWeights refuses a face that has not one file for each weight its text styles are drawn in, and none
// besides: a browser draws a weight with no file in the nearest it has, or thickens one, and never fetches a file of a
// weight no style is drawn in.
func checkFaceWeights(files []readFaceFile, types []readType) error {
	filed := make(map[wholeNumber]bool, len(files))
	for _, file := range files {
		if filed[file.weight] {
			return fmt.Errorf("%w: the face has two files of the weight %d", errTokens, file.weight)
		}
		filed[file.weight] = true
	}
	drawn := make(map[wholeNumber]bool, len(types))
	for _, style := range types {
		if !filed[style.weight] {
			return fmt.Errorf("%w: the face has no file of the weight %d, which %q is drawn in", errTokens, style.weight, style.name)
		}
		drawn[style.weight] = true
	}
	for _, file := range files {
		if !drawn[file.weight] {
			return fmt.Errorf("%w: no text style is drawn in the weight %d, which the face's file %q draws", errTokens, file.weight, file.path)
		}
	}
	return nil
}

// checkFaceFiles refuses a face whose file the repository lacks: the page would ask for a file its build never holds.
func checkFaceFiles(folder fs.FS, face readFace) error {
	for _, file := range face.files {
		if _, err := fs.Stat(folder, file.fromRoot); err != nil {
			return fmt.Errorf("%w: %s names the face's file %q: %w", errTokens, tokensPath, file.path, err)
		}
	}
	return nil
}

// familyList is the face's families as CSS writes them: each name quoted, so the browser reads it as a face's name
// whatever it is made of, but a generic family, which CSS knows by its bare word in any case.
func familyList(families []string) string {
	written := make([]string, 0, len(families))
	for _, family := range families {
		if !isGenericFamily(family) {
			family = familyQuote + family + familyQuote
		}
		written = append(written, family)
	}
	return strings.Join(written, familySeparator)
}

// writeFaceFiles writes each file of the face as a face of its first family, which the face's reader holds is no
// generic family.
func writeFaceFiles(builder *strings.Builder, face readFace) {
	builder.WriteString(faceFilesOpening)
	family := familyQuote + face.families[0] + familyQuote
	for _, file := range face.files {
		fmt.Fprintf(builder, fontFaceRule, family, file.fromStylesheet, file.weight)
	}
}
