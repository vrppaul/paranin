package html

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"strconv"

	"github.com/vrppaul/paranin/internal/look"
	"github.com/vrppaul/paranin/internal/routes"
)

// The media type each picture is served as, and the side of each drawn in pixels: the PNG a phone keeps on its home
// screen, and the one an icon file holds where a browser asks for the icon by name
const (
	drawingType     = "image/svg+xml"
	homeScreenType  = "image/png"
	homeScreenSide  = 180
	askedByNameType = "image/vnd.microsoft.icon"
	askedByNameSide = 32
)

// How a page's head links a picture of the site, and how it names the size of the one of a size, its side by its side
const (
	iconRelation       = "icon"
	homeScreenRelation = "apple-touch-icon"
	sizesBetween       = "x"
)

// The theme whose gradient the icon takes for every reader: its accent and host's colour stand apart from a light
// tab and a dark one alike, where the dark theme's paler pair fades into a light tab
const iconTheme = look.ThemeLight

// The drawing's square, a unit a side; the name the gradient goes by; and how the drawing writes a colour
const (
	drawingSide     = 1
	iconGradientID  = "name-gradient"
	hexColourLayout = "#%02x%02x%02x"
)

// The gradient's line across the square, in shares of its side, as SVG reads a gradient's line on the box it fills:
// from the top-left corner to the bottom-right, the way the name reads
const (
	gradientStartAcross = 0
	gradientStartDown   = 0
	gradientEndAcross   = 1
	gradientEndDown     = 1
)

// Where a pixel's middle stands from its edge
const halfPixel = 0.5

// How an icon file holding one PNG opens (Microsoft's ICO): its kind, an icon, and one picture; then the picture's
// entry: no palette, one plane of 32 bits a pixel
const (
	icoKind         = 1
	icoCount        = 1
	icoPlanes       = 1
	icoBitsPerPixel = 32
)

// Icon is a picture of the site as the site serves it: at its path, as its media type.
type Icon struct {
	Path      string
	MediaType string
	Body      []byte
}

// iconKind is a picture of the site a page's head links: how it links it, the size it names, if any, the picture it
// is, the media type it is served as, and how it is drawn from the icon's gradient.
type iconKind struct {
	relation  string
	sizes     string
	picture   routes.Picture
	mediaType string
	draw      func(gradient iconGradient) []byte
}

// iconKinds are the pictures of the site, in the order a page's head links them: the icon file first, naming its size,
// so a browser that takes the drawing prefers the drawing to it.
func iconKinds() []iconKind {
	askedByNameSides := strconv.Itoa(askedByNameSide)
	return []iconKind{
		{
			relation: iconRelation, sizes: askedByNameSides + sizesBetween + askedByNameSides,
			picture: routes.PictureAskedByName, mediaType: askedByNameType, draw: askedByNameIcon,
		},
		{relation: iconRelation, sizes: "", picture: routes.PictureDrawing, mediaType: drawingType, draw: drawingIcon},
		{
			relation: homeScreenRelation, sizes: "", picture: routes.PictureHomeScreen, mediaType: homeScreenType,
			draw: homeScreenIcon,
		},
	}
}

// iconGradient is the site's icon (DESIGN.md, Favicon): the colour at the start of the gradient's line and the one
// at its end, which fill the square edge to edge, mixed along that line.
type iconGradient struct {
	start color.RGBA
	end   color.RGBA
}

// icoOpening is how an icon file opens: nothing, its kind, and how many pictures it holds.
type icoOpening struct {
	Reserved uint16
	Kind     uint16
	Count    uint16
}

// icoEntry is an icon file's entry for a picture: its width and height, the colours of its palette, nothing, its
// planes and bits a pixel, its length and where it starts in the file.
type icoEntry struct {
	Width        uint8
	Height       uint8
	Colours      uint8
	Reserved     uint8
	Planes       uint16
	BitsPerPixel uint16
	Size         uint32
	Start        uint32
}

// Icons are the pictures of the site, each drawn from the name's gradient, at the path routes writes for it.
func Icons() []Icon {
	gradient := nameGradient()
	icons := make([]Icon, 0, len(iconKinds()))
	for _, kind := range iconKinds() {
		icons = append(icons, Icon{Path: routes.PicturePath(kind.picture), MediaType: kind.mediaType, Body: kind.draw(gradient)})
	}
	return icons
}

// nameGradient is the gradient the name's block letters take, from the accent to the host's colour, as
// web/src/ui/ui.css colours them, in the icon's theme.
func nameGradient() iconGradient {
	palette := look.Palette(iconTheme)
	return iconGradient{start: palette[look.ColourAccent].TrueColour, end: palette[look.ColourHost].TrueColour}
}

// drawingIcon is the icon as SVG, which a browser draws to any size.
func drawingIcon(gradient iconGradient) []byte {
	return []byte(rendered(iconDrawing(gradient)))
}

// hexColour is a colour as SVG writes it, #rrggbb.
//
//wrapper:names how the icon writes each of its colours
func hexColour(colour color.RGBA) string {
	return fmt.Sprintf(hexColourLayout, colour.R, colour.G, colour.B)
}

// homeScreenIcon is the icon a phone keeps on its home screen: a PNG of 180 pixels, whose corners the phone rounds.
func homeScreenIcon(gradient iconGradient) []byte {
	return encodedPNG(rasterIcon(gradient, homeScreenSide))
}

// askedByNameIcon is the icon a browser asks for by name: an icon file holding a PNG of 32 pixels.
func askedByNameIcon(gradient iconGradient) []byte {
	encoded := encodedPNG(rasterIcon(gradient, askedByNameSide))
	opening := icoOpening{Reserved: 0, Kind: icoKind, Count: icoCount}
	// an entry is as long whatever it holds, so the PNG starts after the opening and an empty entry
	var emptyEntry icoEntry
	start := binary.Size(opening) + binary.Size(emptyEntry)
	entry := icoEntry{
		Width: askedByNameSide, Height: askedByNameSide, Colours: 0, Reserved: 0, Planes: icoPlanes,
		BitsPerPixel: icoBitsPerPixel,
		Size:         uint32(len(encoded)), //nolint:gosec // a PNG of 32 pixels a side, far below 4GB
		Start:        uint32(start),        //nolint:gosec // two records of a few bytes
	}
	var file bytes.Buffer
	// A buffer takes every write, and the opening and the entry are of fixed sizes
	_ = binary.Write(&file, binary.LittleEndian, opening)
	_ = binary.Write(&file, binary.LittleEndian, entry)
	file.Write(encoded)
	return file.Bytes()
}

// rasterIcon is the icon in pixels, a square of a side, each pixel the gradient's colour at its middle, as SVG
// paints it there.
func rasterIcon(gradient iconGradient, side int) *image.RGBA {
	canvas := image.NewRGBA(image.Rect(0, 0, side, side))
	for row := range side {
		down := (float64(row) + halfPixel) / float64(side)
		for column := range side {
			across := (float64(column) + halfPixel) / float64(side)
			canvas.SetRGBA(column, row, gradient.at(gradientShare(across, down)))
		}
	}
	return canvas
}

// gradientShare is how far along the gradient's line a point of the square stands, each of its coordinates a share
// of the side, as SVG mixes a linear gradient: the point's projection on the line, as a share of the line's length.
func gradientShare(across, down float64) float64 {
	lineAcross := float64(gradientEndAcross - gradientStartAcross)
	lineDown := float64(gradientEndDown - gradientStartDown)
	projected := (across-gradientStartAcross)*lineAcross + (down-gradientStartDown)*lineDown
	return projected / (lineAcross*lineAcross + lineDown*lineDown)
}

// at is the gradient's colour a share of the way along its line, as SVG mixes it, each channel on its own.
func (gradient iconGradient) at(share float64) color.RGBA {
	start, end := gradient.start, gradient.end
	return color.RGBA{
		R: mixedChannel(start.R, end.R, share), G: mixedChannel(start.G, end.G, share),
		B: mixedChannel(start.B, end.B, share), A: math.MaxUint8,
	}
}

// mixedChannel is a channel a share of the way from one value to another.
func mixedChannel(from, to uint8, share float64) uint8 {
	return uint8(math.Round(float64(from) + (float64(to)-float64(from))*share))
}

// encodedPNG is a picture as a PNG.
func encodedPNG(picture image.Image) []byte {
	var encoded bytes.Buffer
	// A buffer takes every write, and the picture's size is one a PNG holds
	_ = png.Encode(&encoded, picture)
	return encoded.Bytes()
}
