package html

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/a-h/templ"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/look"
)

// How block letters stand in their SVG (DESIGN.md, Block letters): in the drawing's own units, a cell look.CellUnits
// wide and half a row look.HalfRowUnits tall. The SVG is as many font sizes wide as it has cells, so the stylesheet
// sets a cell's width by the SVG's font size, which a token holds
const (
	fontSizeUnit = "em"
	// how a size in font sizes is written: as many decimals as it takes for a width, four for a height, enough that
	// no row of a block falls between pixels; a float of 64 bits
	shortestDecimals = -1
	heightDecimals   = 4
	floatBits        = 64
)

// How a drawing's gradient is named on the page, before the words it draws, which name it apart from the page's other
// drawings; and the classes the stylesheet sizes a drawing by, a cell of the name's font or of the titles'
const (
	gradientPrefix = "letters-"
	nameLetters    = "letters letters-name"
	titleLetters   = "letters letters-title"
)

// letterDrawing is lines of block letters as an SVG draws them: the words they show, the class of their font, their
// rectangles, a cursor where the drawing ends in one, and their size, each in the drawing's units.
type letterDrawing struct {
	words      string
	class      string
	rectangles []letterRectangle
	cursor     *letterRectangle
	width      int
	height     int
}

// letterRectangle is a run of filled cells of one half row, in the drawing's units.
type letterRectangle struct {
	left   int
	top    int
	width  int
	height int
}

// drawnTitle is a page's title as its h1: the résumé's in the name's font, the cursor after it, another page's in the
// titles'; its words where the font does not draw every letter of them.
func (writer PageWriter) drawnTitle(head document.Head) templ.Component {
	words := document.PlainText(head.Title)
	if head.Page.Kind == document.PageResume {
		if drawing, drawn := look.NameDrawing(words); drawn {
			return letteredHeadingOne(drawingOf(drawing, words, nameLetters))
		}
		return headingOne(writer.spans(head.Title))
	}
	return titleHeading(words, writer.spans(head.Title))
}

// titleHeading is words as a page's h1 in the titles' block letters, named by the words; the title as it is where the
// font does not draw every letter of them.
func titleHeading(words string, title templ.Component) templ.Component {
	if drawing, drawn := titleDrawing(words); drawn {
		return letteredHeadingOne(drawing)
	}
	return headingOne(title)
}

// titleDrawing is words in the titles' block letters, which have capitals alone: a page's title or a section's of the
// résumé; not drawn where the font does not draw every letter of them.
func titleDrawing(words string) (letterDrawing, bool) {
	drawing, drawn := look.TitleFont().Draw([]string{strings.ToUpper(words)})
	return drawingOf(drawing, words, titleLetters), drawn
}

// drawingOf is a drawing of block letters in the SVG's units, the words it shows, and the class of its font.
func drawingOf(drawing look.Drawing, words, class string) letterDrawing {
	drawn := letterDrawing{
		words: words, class: class, rectangles: make([]letterRectangle, 0, len(drawing.Rectangles)), cursor: nil,
		width: drawing.Width * look.CellUnits, height: drawing.Height * look.HalfRowUnits,
	}
	for _, rectangle := range drawing.Rectangles {
		drawn.rectangles = append(drawn.rectangles, unitsOf(rectangle))
	}
	if drawing.Cursor != nil {
		cursor := unitsOf(*drawing.Cursor)
		drawn.cursor = &cursor
	}
	return drawn
}

// unitsOf is a rectangle of a drawing in the SVG's units.
func unitsOf(rectangle look.Rectangle) letterRectangle {
	return letterRectangle{
		left: rectangle.Left * look.CellUnits, top: rectangle.Top * look.HalfRowUnits,
		width: rectangle.Width * look.CellUnits, height: rectangle.Height * look.HalfRowUnits,
	}
}

// viewBoxOf is a drawing's own units, a width and a height from its corner, as an SVG's viewBox writes them.
func viewBoxOf(width, height int) string {
	return "0 0 " + strconv.Itoa(width) + " " + strconv.Itoa(height)
}

// svgWidth is the drawing's width in font sizes, a cell one, as its SVG's width.
func (drawing letterDrawing) svgWidth() string {
	return strconv.FormatFloat(float64(drawing.width)/look.CellUnits, 'f', shortestDecimals, floatBits) + fontSizeUnit
}

// svgHeight is the drawing's height in font sizes, a cell's width one, as its SVG's height.
func (drawing letterDrawing) svgHeight() string {
	return strconv.FormatFloat(float64(drawing.height)/look.CellUnits, 'f', heightDecimals, floatBits) + fontSizeUnit
}

// gradientID is the id of the drawing's gradient on the page, one of the page's own: its words, in small letters and
// digits alone. Two drawings of the same words draw the same gradient, so a second of one name changes nothing it
// draws.
func (drawing letterDrawing) gradientID() string {
	return ownID(gradientPrefix + strings.Map(keptInName, strings.ToLower(drawing.words)))
}

// keptInName is a letter as a gradient's name keeps it: a letter or a digit as it is, anything else left out.
func keptInName(letter rune) rune {
	if unicode.IsLetter(letter) || unicode.IsDigit(letter) {
		return letter
	}
	return -1
}
