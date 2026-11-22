package main

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
)

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

// A file of a face is a woff2: every browser the page is for reads it, and it is the smallest of a face's formats
const faceFileEnding = ".woff2"

// How a file of a face draws its letters, as CSS's font-style names it: upright, or slanted as italic letters are
const (
	uprightStyle = "normal"
	italicStyle  = "italic"
)

// How the files of the faces open in the stylesheet, after the page's own tokens
const faceFilesOpening = `
/* The faces' files, which the site serves itself, so every reader sees the faces the page is drawn and counted in */
`

// How a file of a face is written in the stylesheet: a face of its family at its weight and slant, from its path. Its
// text shows in a stand-in until the file has come, then in it (swap): every line's height is in rem, so the swap
// redraws the letters in rows as tall as before, where block would hide the text while the file comes and optional
// would keep a reader on a slow line in the stand-in
const fontFaceRule = `@font-face {
  font-family: %s;
  src: url("%s") format("woff2");
  font-weight: %d;
  font-style: %s;
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

// faceToken is a face of the page as the file writes it: its name, what it is for, the families a browser tries in
// turn, each by its name alone, and the files of the first, which the site serves.
type faceToken struct {
	Name     string          `yaml:"name"`
	About    string          `yaml:"about"`
	Families []string        `yaml:"families"`
	Files    []faceFileToken `yaml:"files"`
}

// faceFileToken is a file of a face as the file writes it: the weight it draws, whether its letters are italic, and
// its path from the tokens' folder. A number the file may leave out unseen is a pointer, so its absence is a fault; a
// file left upright says nothing of it.
type faceFileToken struct {
	Weight *wholeNumber `yaml:"weight"`
	Italic bool         `yaml:"italic"`
	Path   string       `yaml:"path"`
}

// readFace is a face read: its name, what it is for, its families, each a name of letters, digits, hyphens and spaces,
// and the files of the first.
type readFace struct {
	name     string
	about    string
	families []string
	files    []readFaceFile
}

// readFaceFile is a file of a face read: the weight it draws, whether it draws italic letters, its path from the
// tokens' folder, inside it, and the same path from the repository's root and as the stylesheet reaches it.
type readFaceFile struct {
	weight         wholeNumber
	italic         bool
	path           string
	fromRoot       string
	fromStylesheet string
}

// cut is a weight and a slant of a face, which one file draws.
type cut struct {
	weight wholeNumber
	italic bool
}

// readFaces reads the page's faces, the first the one its every text is written in unless its style names another;
// a page with no face is refused.
func readFaces(written []faceToken, named map[string]bool) ([]readFace, error) {
	faces, err := readEach(written, named, readFaceToken)
	if err != nil {
		return nil, err
	}
	if len(faces) == 0 {
		return nil, fmt.Errorf("%w: the page has no face", errTokens)
	}
	return faces, nil
}

// readFaceToken refuses a face that says nothing of what it is for, a family written wrong, and a file written wrong
// or of a generic family, which CSS gives no face of its own.
func readFaceToken(token faceToken, named map[string]bool) (readFace, error) {
	if err := checkToken(token.Name, token.About, named); err != nil {
		return readFace{}, err
	}
	if err := checkFamilies(token.Name, token.Families); err != nil {
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
	return readFace{name: token.Name, about: token.About, families: token.Families, files: files}, nil
}

// checkFamilies refuses a face that names no family, or names one in anything but letters, digits, hyphens and
// spaces, so no name can close the quotes the stylesheet writes it in.
func checkFamilies(face string, families []string) error {
	if len(families) == 0 {
		return fmt.Errorf("%w: the face %q has no family", errTokens, face)
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
	return readFaceFile{
		weight: weight, italic: token.Italic, path: token.Path, fromRoot: fromRoot, fromStylesheet: filepath.ToSlash(fromStylesheet),
	}, nil
}

// isNoFamilyLetter is whether a letter has no place in a family's name.
func isNoFamilyLetter(letter rune) bool {
	return !unicode.IsLetter(letter) && !unicode.IsDigit(letter) && letter != '-' && letter != ' '
}

// isNoPathLetter is whether a letter has no place in the path of a file of a face.
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

// faceOf is the face a text style is drawn in: the one it names, or the page's first where it names none; a fault
// where it names no face of the file.
func faceOf(faces []readFace, name, style string) (readFace, error) {
	if name == "" {
		return faces[0], nil
	}
	found := slices.IndexFunc(faces, func(face readFace) bool { return face.name == name })
	if found < 0 {
		return readFace{}, fmt.Errorf("%w: the text style %q is drawn in %q, which is no face of the file", errTokens, style, name)
	}
	return faces[found], nil
}

// checkFaceCuts refuses a face that has not one upright file for each weight a text style drawn in it is drawn in, and
// a file of a weight no text style of the page is drawn in, or a second of one weight and slant: a browser draws a
// weight with no file in the nearest it has, or thickens one, and never fetches a file of a weight no style is drawn
// in. A face's text may be made strong or italic, in a weight of another style, as an article's words are.
func checkFaceCuts(faces []readFace, types []readType) error {
	drawnAnywhere := make(map[wholeNumber]bool, len(types))
	for _, style := range types {
		drawnAnywhere[style.weight] = true
	}
	faults := make([]error, 0, len(faces))
	for _, face := range faces {
		faults = append(faults, checkCutsOf(face, types, drawnAnywhere))
	}
	return errors.Join(faults...)
}

// checkCutsOf refuses a face whose files leave out an upright weight a style drawn in it needs, draw a weight no style
// is drawn in, or draw one cut twice.
func checkCutsOf(face readFace, types []readType, drawnAnywhere map[wholeNumber]bool) error {
	filed := make(map[cut]bool, len(face.files))
	for _, file := range face.files {
		drawn := cut{weight: file.weight, italic: file.italic}
		if filed[drawn] {
			return fmt.Errorf("%w: the face %q has two files of the weight %d, italic %t", errTokens, face.name, file.weight, file.italic)
		}
		filed[drawn] = true
		if !drawnAnywhere[file.weight] {
			return fmt.Errorf("%w: no text style is drawn in the weight %d, which the face's file %q draws", errTokens, file.weight, file.path)
		}
	}
	for _, style := range types {
		if style.face == face.name && !filed[cut{weight: style.weight, italic: false}] {
			return fmt.Errorf("%w: the face %q has no file of the weight %d, which %q is drawn in", errTokens, face.name, style.weight, style.name)
		}
	}
	return nil
}

// checkFaceFiles refuses a face whose file the repository lacks: the page would ask for a file its build never holds.
func checkFaceFiles(folder fs.FS, faces []readFace) error {
	for _, face := range faces {
		for _, file := range face.files {
			if _, err := fs.Stat(folder, file.fromRoot); err != nil {
				return fmt.Errorf("%w: %s names the face's file %q: %w", errTokens, tokensPath, file.path, err)
			}
		}
	}
	return nil
}

// familyList is a face's families as CSS writes them: each name quoted, so the browser reads it as a face's name
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

// writeFaceFiles writes each file of each face as a face of its first family, which the face's reader holds is no
// generic family.
func writeFaceFiles(builder *strings.Builder, faces []readFace) {
	builder.WriteString(faceFilesOpening)
	for _, face := range faces {
		family := familyQuote + face.families[0] + familyQuote
		for _, file := range face.files {
			fmt.Fprintf(builder, fontFaceRule, family, file.fromStylesheet, file.weight, slantOf(file.italic))
		}
	}
}

// slantOf is how CSS names a file's slant.
func slantOf(italic bool) string {
	if italic {
		return italicStyle
	}
	return uprightStyle
}
