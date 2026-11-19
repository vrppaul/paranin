package html

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/a-h/templ"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/look"
)

// How block letters stand in their SVG (DESIGN.md, Block letters): a cell is cellUnits wide and half a row
// halfRowUnits tall in the drawing's own units, a mono face's cell three fifths as wide as its row is tall. The SVG is
// as many font sizes wide as it has cells, so the stylesheet sets a cell's width by the SVG's font size, which a
// token holds
const (
	cellUnits    = 6
	halfRowUnits = 5
	fontSizeUnit = "em"
	// how a size in font sizes is written: as many decimals as it takes for a width, four for a height, enough that
	// no row of a block falls between pixels; a float of 64 bits
	shortestDecimals = -1
	heightDecimals   = 4
	floatBits        = 64
)

// The column of nothing between the last letter and the cursor, as between two letters, and how tall the cursor is, a
// half row on the letters' foot: an underscore
const (
	cursorGapCells = 1
	cursorHalves   = 1
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

// letterGrid is how a drawing lays a font's cells, in its own units: a cell's width, a half row's height, and where
// the first cell stands.
type letterGrid struct {
	cellWidth     int
	halfRowHeight int
	left          int
	top           int
}

// headingGrid is how a heading's block letters lay their cells: a cell three fifths as wide as its row is tall, as a
// mono face's, from the drawing's corner.
func headingGrid() letterGrid {
	return letterGrid{cellWidth: cellUnits, halfRowHeight: halfRowUnits, left: 0, top: 0}
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
		lines := look.NameLines(words)
		font := look.NameFont()
		if drawing, drawn := drawLetters(font, lines, words, nameLetters); drawn {
			last := []rune(lines[len(lines)-1])
			return letteredHeadingOne(withCursor(drawing, font.LetterWidth(last[len(last)-1])))
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
	return drawLetters(look.TitleFont(), []string{strings.ToUpper(words)}, words, titleLetters)
}

// resumeBlock is a block at the top of the résumé's main part: a section stands at the place the top bar leads to,
// under its title in block letters where the titles' font draws every letter of it, as words where it does not; any
// other block as its kind is.
func (writer PageWriter) resumeBlock(block document.Block) templ.Component {
	section, isSection := block.(document.Section)
	if !isSection {
		return writer.block(block, firstHeadingDepth)
	}
	words := document.PlainText(section.Title)
	if drawing, drawn := titleDrawing(words); drawn {
		return writer.resumeSection(section, letteredHeadingTwo(drawing))
	}
	return writer.resumeSection(section, heading(firstHeadingDepth, writer.spans(section.Title)))
}

// drawLetters is texts in a font as a drawing of the words they show, of the font's class, a row of nothing between
// two texts; not drawn where the font does not draw every letter of them.
func drawLetters(font look.Font, texts []string, words, class string) (letterDrawing, bool) {
	banner := font.Banner(texts)
	if len(banner) == 0 {
		return letterDrawing{words: "", class: "", rectangles: nil, cursor: nil, width: 0, height: 0}, false
	}
	drawing := letterDrawing{words: words, class: class, rectangles: nil, cursor: nil, width: 0, height: 0}
	halves := look.HalfRows(banner)
	for halfRow, filled := range halves {
		drawing.width = max(drawing.width, len(filled)*cellUnits)
		drawing.rectangles = append(drawing.rectangles, halfRowRectangles(filled, halfRow, headingGrid())...)
	}
	drawing.height = len(halves) * halfRowUnits
	return drawing, true
}

// halfRowRectangles are the runs of a half row's filled cells, as rectangles of the half row they stand on, laid on
// a grid.
func halfRowRectangles(filled []bool, halfRow int, grid letterGrid) []letterRectangle {
	rectangles := []letterRectangle{}
	start := -1
	for column := 0; column <= len(filled); column++ {
		on := column < len(filled) && filled[column]
		if on && start < 0 {
			start = column
		}
		if !on && start >= 0 {
			rectangles = append(rectangles, letterRectangle{
				left: grid.left + start*grid.cellWidth, top: grid.top + halfRow*grid.halfRowHeight,
				width: (column - start) * grid.cellWidth, height: grid.halfRowHeight,
			})
			start = -1
		}
	}
	return rectangles
}

// withCursor is a drawing with the cursor after its last line, a column apart, as wide as a letter and a half row
// tall on the foot of the letters' ink, which their last half row may leave empty.
func withCursor(drawing letterDrawing, letterCells int) letterDrawing {
	foot := 0
	lastLineRight := 0
	for _, rectangle := range drawing.rectangles {
		foot = max(foot, rectangle.top+rectangle.height)
	}
	for _, rectangle := range drawing.rectangles {
		if rectangle.top+rectangle.height == foot {
			lastLineRight = max(lastLineRight, rectangle.left+rectangle.width)
		}
	}
	cursor := letterRectangle{
		left: lastLineRight + cursorGapCells*cellUnits, top: foot - cursorHalves*halfRowUnits,
		width: letterCells * cellUnits, height: cursorHalves * halfRowUnits,
	}
	drawing.cursor = &cursor
	drawing.width = max(drawing.width, cursor.left+cursor.width)
	return drawing
}

// viewBox is the drawing's own units, as the SVG's viewBox writes them.
func (drawing letterDrawing) viewBox() string {
	return viewBoxOf(drawing.width, drawing.height)
}

// viewBoxOf is a drawing's own units, a width and a height from its corner, as an SVG's viewBox writes them.
func viewBoxOf(width, height int) string {
	return "0 0 " + strconv.Itoa(width) + " " + strconv.Itoa(height)
}

// svgWidth is the drawing's width in font sizes, a cell one, as its SVG's width.
func (drawing letterDrawing) svgWidth() string {
	return strconv.FormatFloat(float64(drawing.width)/cellUnits, 'f', shortestDecimals, floatBits) + fontSizeUnit
}

// svgHeight is the drawing's height in font sizes, a cell's width one, as its SVG's height.
func (drawing letterDrawing) svgHeight() string {
	return strconv.FormatFloat(float64(drawing.height)/cellUnits, 'f', heightDecimals, floatBits) + fontSizeUnit
}

// gradientID is the name of the drawing's gradient on the page: its words, in small letters and digits alone. Two
// drawings of the same words draw the same gradient, so a second of one name changes nothing it draws.
func (drawing letterDrawing) gradientID() string {
	return gradientPrefix + strings.Map(keptInName, strings.ToLower(drawing.words))
}

// keptInName is a letter as a gradient's name keeps it: a letter or a digit as it is, anything else left out.
func keptInName(letter rune) rune {
	if unicode.IsLetter(letter) || unicode.IsDigit(letter) {
		return letter
	}
	return -1
}
