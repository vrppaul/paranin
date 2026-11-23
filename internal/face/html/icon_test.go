package html

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"image/png"
	"io"
	"testing"
)

// What a picture the site serves holds that a browser cannot read
var (
	errNotADrawing     = errors.New("the drawing is no SVG")
	errMisnamedPicture = errors.New("the icon file's entry names no picture it holds")
)

// The namespace an SVG drawing's root stands in, which a browser needs to draw it as SVG, and the root's own name
const (
	drawingNamespace = "http://www.w3.org/2000/svg"
	drawingElement   = "svg"
)

// Where an icon file's fields stand, in bytes from its start (Microsoft's ICO): its kind and how many pictures it
// holds; then the first picture's entry: its width, its height, its length and where it starts; and where the opening
// and that entry end
const (
	iconFileKindAt     = 2
	iconFileCountAt    = 4
	entryWidthAt       = 6
	entryHeightAt      = 7
	entryLengthAt      = 14
	entryStartAt       = 18
	iconFileOpeningEnd = 22
)

// The kind an icon file names, an icon where 2 is a cursor, and the one picture the site's holds
const (
	iconFileKind     = 1
	iconFilePictures = 1
)

// pictureReaders are how a browser reads each picture of the site, by the media type it is served as.
func pictureReaders() map[string]func(body []byte) error {
	return map[string]func(body []byte) error{
		"image/svg+xml":            readDrawing,
		"image/png":                readPNG,
		"image/vnd.microsoft.icon": readIconFile,
	}
}

// readDrawing reads a drawing as XML whose root is SVG's svg element.
func readDrawing(body []byte) error {
	decoder := xml.NewDecoder(bytes.NewReader(body))
	opened := false
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read the drawing: %w", err)
		}
		if element, isElement := token.(xml.StartElement); isElement && !opened {
			if element.Name != (xml.Name{Space: drawingNamespace, Local: drawingElement}) {
				return fmt.Errorf("%w: its root is %+v", errNotADrawing, element.Name)
			}
			opened = true
		}
	}
}

// readPNG reads a picture as a PNG.
func readPNG(body []byte) error {
	if _, err := png.Decode(bytes.NewReader(body)); err != nil {
		return fmt.Errorf("read the PNG: %w", err)
	}
	return nil
}

// readIconFile reads an icon file as a browser does: its opening, then the entry of its one picture, which must name
// the bytes and the size of the PNG it holds.
func readIconFile(body []byte) error {
	if len(body) < iconFileOpeningEnd {
		return fmt.Errorf("%w: the file is %d bytes, shorter than its opening", errMisnamedPicture, len(body))
	}
	kind := binary.LittleEndian.Uint16(body[iconFileKindAt:])
	count := binary.LittleEndian.Uint16(body[iconFileCountAt:])
	width, height := int(body[entryWidthAt]), int(body[entryHeightAt])
	start := int(binary.LittleEndian.Uint32(body[entryStartAt:]))
	pictureEnd := start + int(binary.LittleEndian.Uint32(body[entryLengthAt:]))
	if kind != iconFileKind || count != iconFilePictures || start != iconFileOpeningEnd || pictureEnd != len(body) {
		return fmt.Errorf("%w: kind %d, %d pictures, bytes %d to %d of %d", errMisnamedPicture, kind, count, start,
			pictureEnd, len(body))
	}
	picture, err := png.DecodeConfig(bytes.NewReader(body[start:pictureEnd]))
	if err != nil {
		return fmt.Errorf("read the icon file's PNG: %w", err)
	}
	if picture.Width != width || picture.Height != height {
		return fmt.Errorf("%w: the entry says %d by %d, the PNG is %d by %d", errMisnamedPicture, width, height,
			picture.Width, picture.Height)
	}
	return readPNG(body[start:pictureEnd])
}

// A browser reads every picture of the site as the media type it is served as: the drawing as SVG, the home
// screen's as a PNG, and the icon file by its one entry, which names the PNG it holds.
func TestABrowserReadsEveryPictureOfTheSite(test *testing.T) {
	readers := pictureReaders()
	for _, icon := range Icons() {
		test.Run(icon.Path, func(test *testing.T) {
			read, known := readers[icon.MediaType]
			if !known {
				test.Fatalf("the picture at %q is served as %q, which no browser here reads", icon.Path, icon.MediaType)
			}
			if err := read(icon.Body); err != nil {
				test.Errorf("reading the picture at %q as %q = %v, want no fault", icon.Path, icon.MediaType, err)
			}
		})
	}
}
