package html

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"slices"
	"strconv"
	"unicode/utf8"

	"github.com/vrppaul/paranin/internal/look"
	"github.com/vrppaul/paranin/internal/routes"
)

// How the site's icon stands in its square (DESIGN.md, Favicon), in the square's own units, 32 a side, so a unit is
// half a pixel at 16 and a pixel at 32: the first letter of the author's name in the name's font, its cells square, as
// many units a side as lets the letter and the cursor fill the square but for a unit's margin round them; the cursor,
// a beam two units after the letter, three wide and as tall as the letter's last three half rows. The two stand in the
// middle, a unit left over standing before them. Where the font draws no such letter, the cursor stands alone in the
// middle, as tall as beside a letter of four cells
const (
	iconSide         = 32
	iconMargin       = 1
	iconCursorGap    = 2
	iconCursorWidth  = 3
	iconCursorHalves = 3
	// a margin stands on both sides, and what a length leaves of the square is shared between both, the side before
	// taking a unit left over
	bothSides      = 2
	leftOverBefore = 1
	// the width the letter and the cursor may take across the square, and the letter a lone cursor is as tall as beside
	iconAcross       = iconSide - bothSides*iconMargin - iconCursorGap - iconCursorWidth
	loneLetterCells  = 4
	loneCursorHeight = iconCursorHalves * (iconAcross / loneLetterCells)
)

// How much of the way from the gradient's start to its end a colour is, at most and at least
const (
	noShare    = 0
	wholeShare = 1
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

// How far in from each edge a phone's picture is drawn: a tenth of its side, since a phone cuts the picture's corners
// round, about 40 of its 180 pixels, and would cut the letter's corner and the cursor's foot where they stand a unit in
const (
	homeScreenInsetParts = 10
	homeScreenInset      = homeScreenSide / homeScreenInsetParts
)

// How a page's head links a picture of the site, and how it names the size of the one of a size, its side by its side
const (
	iconRelation       = "icon"
	homeScreenRelation = "apple-touch-icon"
	sizesBetween       = "x"
)

// The theme a picture of one theme is drawn in, a phone's home screen and the icon asked for by name: the dark one, a
// terminal's, where the site began. The drawing has both and shows the reader's
const pictureTheme = look.ThemeDark

// How the drawing names the letter it draws once and lays in each theme's part, how a reference leads to a name, and
// how it writes a colour
const (
	iconLetterID    = "icon-letter"
	referenceMark   = "#"
	hexColourLayout = "#%02x%02x%02x"
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
// is, the media type it is served as, and how it is drawn from the icon's shapes.
type iconKind struct {
	relation  string
	sizes     string
	picture   routes.Picture
	mediaType string
	draw      func(picture iconPicture) []byte
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

// iconPicture is the site's icon in its square's units: the letter's rectangles, which the gradient fills from the
// letter's left edge to its right, and the cursor's.
type iconPicture struct {
	letter      []letterRectangle
	letterLeft  int
	letterRight int
	cursor      letterRectangle
}

// iconColours are the icon's colours in a theme: the ground, the letter's gradient from its start to its end, and the
// cursor.
type iconColours struct {
	ground color.RGBA
	start  color.RGBA
	end    color.RGBA
	cursor color.RGBA
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

// Icons are the pictures of the site, each drawn from the first letter of the author's name, at the path routes
// writes for it.
func Icons(author string) []Icon {
	picture := iconPictureOf(author)
	icons := make([]Icon, 0, len(iconKinds()))
	for _, kind := range iconKinds() {
		icons = append(icons, Icon{Path: routes.PicturePath(kind.picture), MediaType: kind.mediaType, Body: kind.draw(picture)})
	}
	return icons
}

// iconPictureOf is the icon of an author: the first letter of the name, as the name's font stands it, and the
// cursor after it; the cursor alone where the font does not draw the letter.
func iconPictureOf(author string) iconPicture {
	var halves [][]bool
	if lines := look.NameLines(author); len(lines) > 0 {
		initial, _ := utf8.DecodeRuneInString(lines[0])
		halves = look.HalfRows(look.NameFont().Banner([]string{string(initial)}))
	}
	inked := inkedHalves(halves)
	if inked == 0 {
		return loneCursor()
	}
	return letterWithCursor(halves[:inked])
}

// inkedHalves is how many of a glyph's half rows reach down to its ink's foot: none for a glyph with no ink.
func inkedHalves(halves [][]bool) int {
	inked := 0
	for halfRow, filled := range halves {
		if slices.Contains(filled, true) {
			inked = halfRow + 1
		}
	}
	return inked
}

// letterWithCursor is a glyph's inked half rows, every one as wide, on square cells as large as the square lets
// them, and the cursor after them on their foot, the two in the middle of the square.
func letterWithCursor(halves [][]bool) iconPicture {
	cells, inked := len(halves[0]), len(halves)
	cellSide := min(iconAcross/cells, (iconSide-bothSides*iconMargin)/inked)
	grid := letterGrid{
		cellWidth:     cellSide,
		halfRowHeight: cellSide,
		left:          centredStart(cells*cellSide + iconCursorGap + iconCursorWidth),
		top:           centredStart(inked * cellSide),
	}
	letterRight := grid.left + cells*cellSide
	picture := iconPicture{letter: nil, letterLeft: grid.left, letterRight: letterRight, cursor: letterRectangle{
		left: letterRight + iconCursorGap, top: grid.top + (inked-iconCursorHalves)*cellSide,
		width: iconCursorWidth, height: iconCursorHalves * cellSide,
	}}
	for halfRow, filled := range halves {
		picture.letter = append(picture.letter, halfRowRectangles(filled, halfRow, grid)...)
	}
	return picture
}

// loneCursor is the icon of an author whose name's first letter the font does not draw: the cursor alone, in the
// middle of the square.
func loneCursor() iconPicture {
	cursor := letterRectangle{
		left: centredStart(iconCursorWidth), top: centredStart(loneCursorHeight), width: iconCursorWidth, height: loneCursorHeight,
	}
	return iconPicture{letter: nil, letterLeft: cursor.left, letterRight: cursor.left, cursor: cursor}
}

// centredStart is where a length starts in the middle of the square, a unit left over standing before it.
func centredStart(length int) int {
	return (iconSide - length + leftOverBefore) / bothSides
}

// drawingIcon is the icon as SVG, in the light theme and the dark one, the reader's shown.
func drawingIcon(picture iconPicture) []byte {
	return []byte(rendered(iconDrawing(picture, iconColoursOf(look.ThemeLight), iconColoursOf(look.ThemeDark))))
}

// iconColoursOf are the icon's colours in a theme: the page's ground, the name's gradient from the accent to the
// host's colour and the cursor in the accent, as web/src/ui/ui.css colours the name's block letters on the page.
func iconColoursOf(theme look.Theme) iconColours {
	palette := look.Palette(theme)
	return iconColours{
		ground: look.Background(theme),
		start:  palette[look.ColourAccent].TrueColour,
		end:    palette[look.ColourHost].TrueColour,
		cursor: palette[look.ColourAccent].TrueColour,
	}
}

// hexColour is a colour as SVG writes it, #rrggbb.
//
//wrapper:names how the icon writes each of its colours
func hexColour(colour color.RGBA) string {
	return fmt.Sprintf(hexColourLayout, colour.R, colour.G, colour.B)
}

// homeScreenIcon is the icon a phone keeps on its home screen: a PNG of 180 pixels, the ground a whole square, which
// the phone rounds itself, the letter and the cursor drawn in from its edges.
func homeScreenIcon(picture iconPicture) []byte {
	return encodedPNG(rasterIcon(picture, homeScreenSide, homeScreenInset))
}

// askedByNameIcon is the icon a browser asks for by name: an icon file holding a PNG of 32 pixels.
func askedByNameIcon(picture iconPicture) []byte {
	encoded := encodedPNG(rasterIcon(picture, askedByNameSide, 0))
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

// rasterIcon is the icon in pixels, a square of a side, in the theme a picture of one theme is drawn in: the ground
// whole, the letter and the cursor drawn an inset in from each edge; each edge at the pixel edge nearest it, so none is
// blurred, as the drawing's crisp edges are; the letter's gradient worked out at each column's middle.
func rasterIcon(picture iconPicture, side, inset int) *image.RGBA {
	colours := iconColoursOf(pictureTheme)
	canvas := image.NewRGBA(image.Rect(0, 0, side, side))
	place := iconPlace{scale: float64(side-bothSides*inset) / iconSide, inset: float64(inset)}
	draw.Draw(canvas, canvas.Bounds(), image.NewUniform(colours.ground), image.Point{}, draw.Src)
	for _, rectangle := range picture.letter {
		bounds := place.pixelBounds(rectangle)
		for column := bounds.Min.X; column < bounds.Max.X; column++ {
			colour := gradientAt(picture, colours, place.units(float64(column)+halfPixel))
			strip := image.Rect(column, bounds.Min.Y, column+1, bounds.Max.Y)
			draw.Draw(canvas, strip, image.NewUniform(colour), image.Point{}, draw.Src)
		}
	}
	draw.Draw(canvas, place.pixelBounds(picture.cursor), image.NewUniform(colours.cursor), image.Point{}, draw.Src)
	return canvas
}

// iconPlace is where the square's units stand in a picture's pixels: how many pixels a unit is, and how far in from
// the picture's edge the square starts.
type iconPlace struct {
	scale float64
	inset float64
}

// pixel is a length across the square, in its units, as the picture's pixels from its edge, at the pixel edge
// nearest it.
func (place iconPlace) pixel(units int) int {
	return int(math.Round(place.inset + float64(units)*place.scale))
}

// units is a place in the picture's pixels as the square's units from its edge.
func (place iconPlace) units(pixels float64) float64 {
	return (pixels - place.inset) / place.scale
}

// pixelBounds is a rectangle of the square's units in the picture's pixels.
func (place iconPlace) pixelBounds(rectangle letterRectangle) image.Rectangle {
	return image.Rect(
		place.pixel(rectangle.left), place.pixel(rectangle.top),
		place.pixel(rectangle.left+rectangle.width), place.pixel(rectangle.top+rectangle.height),
	)
}

// gradientAt is the letter's gradient at a place across the square, in its units: the start at the letter's left
// edge, the end at its right, as SVG mixes them, each channel on its own.
func gradientAt(picture iconPicture, colours iconColours, across float64) color.RGBA {
	start, end := colours.start, colours.end
	share := (across - float64(picture.letterLeft)) / float64(picture.letterRight-picture.letterLeft)
	share = min(max(share, noShare), wholeShare)
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
