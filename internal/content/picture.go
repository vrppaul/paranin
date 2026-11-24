package content

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"maps"
	"math"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/yuin/goldmark/ast"
	extensionast "github.com/yuin/goldmark/extension/ast"
	"golang.org/x/image/webp"
)

// Picture is a picture a post shows, a file of its folder read whole as the post loads, so a reload swaps it with the
// text: the type it is sent as, its size in pixels, which a page keeps its room by before it comes, and what it holds.
type Picture struct {
	MediaType string
	Width     int
	Height    int
	Body      []byte
}

// The most a picture may hold, in bytes: each is held in memory with its snapshot and sent whole to a reader on a
// phone, and a screenshot twice the text's column wide fits in it as WebP, or as a PNG of few colours
const largestPicture = 1 << 20

// pictureType is a type of picture a post may show: what it is sent as, and how its size in pixels is read.
type pictureType struct {
	mediaType string
	readSize  func(io.Reader) (image.Config, error)
}

// pictureTypes are the pictures a post may show, by the ending of the file's name. AVIF is left out: neither Go nor
// its x/image reads its size, and WebP stands in for it.
func pictureTypes() map[string]pictureType {
	// a JPEG goes by either ending
	photograph := pictureType{mediaType: "image/jpeg", readSize: jpeg.DecodeConfig}
	return map[string]pictureType{
		".png":  {mediaType: "image/png", readSize: png.DecodeConfig},
		".jpg":  photograph,
		".jpeg": photograph,
		".webp": {mediaType: "image/webp", readSize: webp.DecodeConfig},
		// a drawing is served as a picture alone, under a policy that runs nothing in it (docs/WEB.md)
		".svg": {mediaType: "image/svg+xml", readSize: drawingSize},
	}
}

// inspectPicture holds a picture to a paragraph of its own among the post's blocks, a list's item's among them, which
// every face draws as a figure; to alt text of words alone, which stands for it wherever it is not seen; to
// a caption of characters a value may hold; and to a file of the post's own folder, which it reads.
func (checked *bodyInspection) inspectPicture(picture *ast.Image) {
	line := checked.lineOf(picture)
	if !standsAlone(picture) {
		checked.inspection.addAtLine(line, pictureAloneProblem)
	}
	checked.inspectAlt(picture, line)
	if caption := Resolved(picture.Title); referencedCharacters(picture.Title, caption) {
		checked.inspection.addAtLine(line, characterProblem, caption)
	}
	name := ResolvedAddress(picture.Destination)
	checked.shown.named[name] = true
	if format, named := checked.inspectPictureName(name, line); named {
		checked.readPicture(name, format, line)
	}
}

// standsAlone is whether a picture is the one thing a paragraph of the post's blocks holds, a tight list's item's
// among them, and no footnote's, whose words stand beside a line.
func standsAlone(picture *ast.Image) bool {
	paragraph := picture.Parent()
	ownParagraph := paragraph.Kind() == ast.KindParagraph || paragraph.Kind() == ast.KindTextBlock
	return ownParagraph && paragraph.ChildCount() == 1 && !inside(picture, extensionast.KindFootnote)
}

// inspectAlt holds a picture's alt text to words, and to words alone, since alt text drops a mark.
func (checked *bodyInspection) inspectAlt(picture *ast.Image, line int) {
	for child := picture.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() != ast.KindText {
			checked.inspection.addAtLine(line, altMarkProblem)
			return
		}
	}
	if blank(checked.body.AltText(picture)) {
		checked.inspection.addAtLine(line, altMissingProblem)
	}
}

// inspectPictureName holds a picture's address to the name of a file of the post's own folder, a slug and the ending of
// a format the site serves a picture as, so no address leads out of the folder, to another site among them, and each
// carries its name as it is; the format it names, where it is one.
func (checked *bodyInspection) inspectPictureName(name string, line int) (pictureType, bool) {
	ending := path.Ext(name)
	types := pictureTypes()
	format, known := types[ending]
	switch {
	case !known:
		checked.inspection.addAtLine(line, pictureTypeProblem, name, strings.Join(slices.Sorted(maps.Keys(types)), kindSeparator))
	case !regexp.MustCompile(slugPattern).MatchString(strings.TrimSuffix(name, ending)):
		checked.inspection.addAtLine(line, pictureNameProblem, name)
	default:
		return format, true
	}
	return format, false
}

// readPicture reads a picture of the post's folder by its name, of a format the site serves, and keeps it; or finds it
// unread, missing among others, past the size a picture may hold, or reading as no picture of its format.
func (checked *bodyInspection) readPicture(name string, format pictureType, line int) {
	body, err := readHeld(checked.folder, path.Join(checked.place, name))
	switch {
	case errors.Is(err, errIrregularFile):
		checked.inspection.addAtLine(line, pictureFileProblem, name)
		return
	case err != nil:
		checked.inspection.addAtLine(line, pictureReadProblem, name, err)
		return
	case len(body) > largestPicture:
		checked.inspection.addAtLine(line, largePictureProblem, name, largestPicture)
		return
	}
	size, err := format.readSize(bytes.NewReader(body))
	if err != nil {
		checked.inspection.addAtLine(line, unreadPictureProblem, name, err)
		return
	}
	checked.shown.read[name] = Picture{MediaType: format.mediaType, Width: size.Width, Height: size.Height, Body: body}
}

// The fault of a picture that is no regular file, which no read opens
var errIrregularFile = errors.New("no regular file")

// readHeld is what a regular file of a folder holds, opened once and read no further than one byte past the most a
// picture may hold, however large it is or grows while it is read; anything else, as a named pipe, whose read would
// wait for a writer, is never opened. A fault is its cause alone, since whoever asked names the file.
func readHeld(folder fs.FS, file string) ([]byte, error) {
	info, err := fs.Stat(folder, file)
	if err != nil {
		return nil, causeOf(err)
	}
	if !info.Mode().IsRegular() {
		return nil, errIrregularFile
	}
	opened, err := folder.Open(file)
	if err != nil {
		return nil, causeOf(err)
	}
	body, err := io.ReadAll(io.LimitReader(opened, largestPicture+1))
	if err := errors.Join(err, opened.Close()); err != nil {
		return nil, causeOf(err)
	}
	return body, nil
}

// What joins two lines of alt text its author broke apart
const wordSpace = " "

// AltText is a picture's alt text as a reader reads it, its words joined across the lines its author broke them at.
func (markdown Markdown) AltText(picture *ast.Image) string {
	var builder strings.Builder
	for child := picture.FirstChild(); child != nil; child = child.NextSibling() {
		if written, isText := child.(*ast.Text); isText {
			builder.WriteString(Resolved(written.Segment.Value(markdown.Source)))
			if written.SoftLineBreak() {
				builder.WriteString(wordSpace)
			}
		}
	}
	return builder.String()
}

// drawingRoot is what the load reads of a drawing: its root, an SVG's, and the size it names in its width and its
// height, or in its view box, the last two of four numbers.
type drawingRoot struct {
	XMLName xml.Name `xml:"http://www.w3.org/2000/svg svg"`
	Width   string   `xml:"width,attr"`
	Height  string   `xml:"height,attr"`
	ViewBox string   `xml:"viewBox,attr"`
}

// How a drawing writes its size: a length in pixels, a number with or without px after it; and its view box, four
// numbers parted by spaces or commas, of which the last two are its width and its height
const (
	pixelUnit       = "px"
	viewBoxParts    = 4
	viewBoxWidth    = 2
	viewBoxHeight   = 3
	viewBoxSeparate = " ,"
	numberBits      = 64
)

// drawingSize is the size of an SVG in pixels: its width and its height where both are written in pixels; where one
// is, that one, and the other by its view box's shape; its view box's where neither is; none where it names no size,
// so no page could keep its room.
func drawingSize(reader io.Reader) (image.Config, error) {
	var root drawingRoot
	if err := xml.NewDecoder(reader).Decode(&root); err != nil {
		return image.Config{}, fmt.Errorf("read the drawing's root: %w", err)
	}
	width, widthKnown := pixels(root.Width)
	height, heightKnown := pixels(root.Height)
	boxWidth, boxHeight, boxKnown := viewBoxSize(root.ViewBox)
	switch {
	case widthKnown && heightKnown:
	case boxKnown && widthKnown:
		height = width * boxHeight / boxWidth
	case boxKnown && heightKnown:
		width = height * boxWidth / boxHeight
	case boxKnown:
		width, height = boxWidth, boxHeight
	default:
		return image.Config{}, fmt.Errorf("width %q, height %q and view box %q name no size in pixels", root.Width, root.Height, root.ViewBox)
	}
	return image.Config{ColorModel: nil, Width: wholePixels(width), Height: wholePixels(height)}, nil
}

// viewBoxSize is the width and the height a drawing's view box names, where it names four numbers, those two above
// nothing.
func viewBoxSize(viewBox string) (float64, float64, bool) {
	box := strings.FieldsFunc(viewBox, func(letter rune) bool { return strings.ContainsRune(viewBoxSeparate, letter) })
	if len(box) != viewBoxParts {
		return 0, 0, false
	}
	width, widthKnown := pixels(box[viewBoxWidth])
	height, heightKnown := pixels(box[viewBoxHeight])
	return width, height, widthKnown && heightKnown
}

// pixels is a length of a drawing in pixels, where it is written as a number of them above nothing, and no more than an
// int holds on any machine.
func pixels(length string) (float64, bool) {
	number, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(length), pixelUnit), numberBits)
	if err != nil || math.IsNaN(number) || number <= 0 || number > math.MaxInt32 {
		return 0, false
	}
	return number, true
}

// wholePixels is a length in whole pixels, at least one, and no more than an int holds on any machine, which a length
// worked out from a view box's shape may pass.
func wholePixels(length float64) int {
	return max(int(math.Round(min(length, math.MaxInt32))), 1)
}
