package html

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"image/color"
	"image/png"
	"io"
	"math"
	"strconv"
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

// How far a pixel's colour may stand from the colour at the place it covers, in steps of a channel: the fixture's
// channels part by less than a step a pixel, so the half pixel between a corner and its pixel's middle moves a
// channel by under half a step, and rounding by at most half another
const colourStepsAllowed = 1

// gradientFixture is a gradient the raster's test draws, and the colour halfway along it.
type gradientFixture struct {
	gradient iconGradient
	midpoint color.RGBA
}

// rasterPlace is a pixel of the raster by where it stands, its column and its row, and the colour it should hold.
type rasterPlace struct {
	name   string
	column int
	row    int
	want   color.RGBA
}

// drawnIcon is the icon's drawing as SVG writes it, read for the line its gradient runs along.
type drawnIcon struct {
	Gradient drawnGradientLine `xml:"linearGradient"`
}

// drawnGradientLine is the line a drawing's gradient runs along, from the point it starts at to the one it ends at,
// each across and down, as the SVG writes them.
type drawnGradientLine struct {
	StartAcross string `xml:"x1,attr"`
	StartDown   string `xml:"y1,attr"`
	EndAcross   string `xml:"x2,attr"`
	EndDown     string `xml:"y2,attr"`
}

// diagonalFixture is a gradient whose channels part by 30 at most, under a step a pixel at either side the site
// draws, and its midpoint.
func diagonalFixture() gradientFixture {
	return gradientFixture{
		gradient: iconGradient{
			start: color.RGBA{R: 64, G: 96, B: 128, A: math.MaxUint8},
			end:   color.RGBA{R: 94, G: 126, B: 98, A: math.MaxUint8},
		},
		midpoint: color.RGBA{R: 79, G: 111, B: 113, A: math.MaxUint8},
	}
}

// withinColourSteps is whether every channel of one colour stands within the steps allowed of the other's.
func withinColourSteps(got, want color.RGBA) bool {
	return channelGap(got.R, want.R) <= colourStepsAllowed && channelGap(got.G, want.G) <= colourStepsAllowed &&
		channelGap(got.B, want.B) <= colourStepsAllowed && got.A == want.A
}

// channelGap is how many steps part two values of a channel.
func channelGap(first, second uint8) int {
	return max(int(first)-int(second), int(second)-int(first))
}

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

// The icon in pixels runs its gradient on the diagonal, at every side the site draws it: its top-left corner holds
// the gradient's start, its bottom-right the end, and the two other corners and its middle the midpoint.
func TestTheRasterIconRunsItsGradientOnTheDiagonal(test *testing.T) {
	fixture := diagonalFixture()
	for _, side := range []int{homeScreenSide, askedByNameSide} {
		test.Run(strconv.Itoa(side), func(test *testing.T) {
			canvas := rasterIcon(fixture.gradient, side)
			first, middle, last := 0, side/2, side-1
			places := []rasterPlace{
				{name: "top-left corner", column: first, row: first, want: fixture.gradient.start},
				{name: "top-right corner", column: last, row: first, want: fixture.midpoint},
				{name: "middle", column: middle, row: middle, want: fixture.midpoint},
				{name: "bottom-left corner", column: first, row: last, want: fixture.midpoint},
				{name: "bottom-right corner", column: last, row: last, want: fixture.gradient.end},
			}
			for _, place := range places {
				got := canvas.RGBAAt(place.column, place.row)
				if !withinColourSteps(got, place.want) {
					test.Errorf("the %d-pixel icon's %s, pixel (%d, %d) = %v, want %v within %d steps", side,
						place.name, place.column, place.row, got, place.want, colourStepsAllowed)
				}
			}
		})
	}
}

// The icon's drawing names its gradient's line from the square's top-left corner to its bottom-right, in shares of
// the square, so a browser draws it on the diagonal as the raster does.
func TestTheDrawingRunsItsGradientOnTheDiagonal(test *testing.T) {
	var drawing drawnIcon
	if err := xml.Unmarshal(drawingIcon(diagonalFixture().gradient), &drawing); err != nil {
		test.Fatalf("reading the drawing = %v, want no fault", err)
	}
	want := drawnGradientLine{StartAcross: "0", StartDown: "0", EndAcross: "1", EndDown: "1"}
	if drawing.Gradient != want {
		test.Errorf("the drawing's gradient runs along %+v, want %+v", drawing.Gradient, want)
	}
}
