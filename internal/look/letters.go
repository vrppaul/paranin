package look

import (
	"strings"
	"unicode/utf8"
)

// How a banner stands in a block font: a column between two glyphs, and how a glyph's rows are written in its table
const (
	glyphGap   = 1
	glyphBreak = "\n"
	glyphSpace = " "
)

// BannerLineGap is the rows of nothing between two lines of a banner, in every face that draws one.
const BannerLineGap = 1

// The half blocks a glyph's cell is made of: the upper half, the lower, both; and the halves a row of cells holds
const (
	upperHalf   = '▀'
	lowerHalf   = '▄'
	fullBlock   = '█'
	halvesInRow = 2
)

// Font is a block font, the design's letters for a name and a title (DESIGN.md, Block letters): how many rows its
// glyphs are tall, and each letter it draws as its rows of half blocks, ▀ the upper half of a cell, ▄ the lower, █
// both. A text with a letter the font has not is drawn as text by whichever face shows it.
type Font struct {
	rows   int
	glyphs map[rune][]string
}

// NameFont is the first version's banner font, in which the author's name stands: each capital, digit, space and
// hyphen three rows tall.
func NameFont() Font {
	glyphs := map[rune]string{
		'A': "█▀▀█\n█▀▀█\n▀  ▀", 'B': "█▀▀▄\n█▀▀█\n▀▀▀ ", 'C': "█▀▀▀\n█   \n▀▀▀▀", 'D': "█▀▀▄\n█  █\n▀▀▀ ",
		'E': "█▀▀▀\n█▀▀ \n▀▀▀▀", 'F': "█▀▀▀\n█▀▀ \n▀   ", 'G': "█▀▀▀\n█ ▀█\n▀▀▀▀", 'H': "█  █\n█▀▀█\n▀  ▀",
		'I': "▀█▀\n █ \n▀▀▀", 'J': "   █\n▄  █\n▀▀▀▀", 'K': "█ ▄▀\n█▀▄ \n▀  ▀", 'L': "█   \n█   \n▀▀▀▀",
		'M': "█▄ ▄█\n█ ▀ █\n▀   ▀", 'N': "█▄ █\n█ ▀█\n▀  ▀", 'O': "█▀▀█\n█  █\n▀▀▀▀", 'P': "█▀▀█\n█▀▀▀\n▀   ",
		'Q': "█▀▀█\n█ ▄█\n▀▀▀▀", 'R': "█▀▀█\n█▀▀▄\n▀  ▀", 'S': "█▀▀▀\n▀▀▀█\n▀▀▀▀", 'T': "▀█▀\n █ \n ▀ ",
		'U': "█  █\n█  █\n▀▀▀▀", 'V': "█  █\n█  █\n ▀▀ ", 'W': "█   █\n█▄▀▄█\n▀   ▀", 'X': "█  █\n▄▀▀▄\n▀  ▀",
		'Y': "█  █\n ██ \n ▀▀ ", 'Z': "▀▀▀█\n▄▀▀ \n▀▀▀▀", ' ': "  \n  \n  ", '-': "   \n▀▀▀\n   ",
		'1': "▀█ \n █ \n▀▀▀", '2': "▀▀▀█\n█▀▀▀\n▀▀▀▀", '3': "▀▀▀█\n ▀▀█\n▀▀▀▀", '4': "█  █\n▀▀▀█\n   ▀",
		'6': "█▀▀▀\n█▀▀█\n▀▀▀▀", '7': "▀▀▀█\n  █ \n  ▀ ", '8': "█▀▀█\n█▀▀█\n▀▀▀▀", '9': "█▀▀█\n▀▀▀█\n▀▀▀▀",
	}
	// the font draws a zero as the O and a five as the S
	glyphs['0'], glyphs['5'] = glyphs['O'], glyphs['S']
	return fontOf(glyphs)
}

// TitleFont is the font each section of the résumé stands under, smaller than the name's: each capital, digit, space
// and hyphen two rows tall, so four half blocks high.
func TitleFont() Font {
	// X stands on three cells, not five as M and W: on five its halves stood apart and read as )(
	glyphs := map[rune]string{
		'A': "▄▀▄\n█▀█", 'B': "██▄\n█▄█", 'C': "█▀▀\n█▄▄", 'D': "█▀▄\n█▄▀", 'E': "██▀\n█▄▄", 'F': "██▀\n█  ",
		'G': "█▀▀\n█▄█", 'H': "█▄█\n█ █", 'I': "▀█▀\n▄█▄", 'J': "  █\n█▄█", 'K': "█▄▀\n█ █", 'L': "█  \n█▄▄",
		'M': "█▄ ▄█\n█ ▀ █", 'N': "█▄ █\n█ ▀█", 'O': "█▀█\n█▄█", 'P': "█▀█\n█▀▀", 'Q': "█▀█\n▀▀█", 'R': "█▀█\n█▀▄",
		'S': "▄▀▀\n▄▄▀", 'T': "▀█▀\n █ ", 'U': "█ █\n█▄█", 'V': "█ █\n▀▄▀", 'W': "█ ▄ █\n▀▄▀▄▀", 'X': "▀▄▀\n▄▀▄",
		'Y': "█ █\n █ ", 'Z': "▀▀█\n█▄▄", ' ': "  \n  ", '-': "▄▄▄\n   ",
		'1': "▄█ \n▄█▄", '3': "▀██\n▄▄█", '4': "█ █\n▀▀█", '5': "█▀▀\n▄▄█", '6': "█▄▄\n█▄█", '7': "▀▀█\n █ ",
		'8': "█▀█\n███",
	}
	// four half blocks give a zero no other shape than the O, a two than the Z, a nine than the Q
	glyphs['0'], glyphs['2'], glyphs['9'] = glyphs['O'], glyphs['Z'], glyphs['Q']
	return fontOf(glyphs)
}

// fontOf is a font of glyphs, each glyph's rows written as one text, a break between each two; its glyphs are as
// tall as the tallest of them.
func fontOf(glyphs map[rune]string) Font {
	built := Font{rows: 0, glyphs: make(map[rune][]string, len(glyphs))}
	for letter, glyph := range glyphs {
		rows := strings.Split(glyph, glyphBreak)
		built.glyphs[letter], built.rows = rows, max(built.rows, len(rows))
	}
	return built
}

// Banner is texts in the font, each text its rows of glyphs, a column between each two; nothing where a text has a
// letter the font does not draw.
func (font Font) Banner(texts []string) [][]string {
	rows := make([][]string, 0, len(texts))
	for _, text := range texts {
		for _, letter := range text {
			if _, drawn := font.glyphs[letter]; !drawn {
				return nil
			}
		}
		rows = append(rows, font.blockRows(text))
	}
	return rows
}

// blockRows are a text's rows in the font, a column between each two glyphs; every letter of it is one the font draws.
func (font Font) blockRows(text string) []string {
	rows := make([]string, font.rows)
	for index, letter := range text {
		for row, glyphRow := range font.glyphs[letter] {
			if index > 0 {
				rows[row] += strings.Repeat(glyphSpace, glyphGap)
			}
			rows[row] += glyphRow
		}
	}
	return rows
}

// Drawing is lines of block letters as a face that draws shapes lays them out (DESIGN.md, Block letters): the lines'
// words, the runs of each half row's filled cells, the cursor after the last line where one follows, and how wide and
// tall the drawing is, in cells and half rows.
type Drawing struct {
	Lines      []string
	Rectangles []Rectangle
	Cursor     *Rectangle
	Width      int
	Height     int
}

// A cell of a drawing's shape, in units of its own: a mono face's cell three fifths as wide as its row is tall, so a
// cell is CellUnits wide and a half row HalfRowUnits tall
const (
	CellUnits    = 6
	HalfRowUnits = 5
)

// Rectangle is a run of a drawing's filled cells on one half row, or its cursor: its left, in cells, its top, in half
// rows, and its width and height in the same.
type Rectangle struct {
	Left   int
	Top    int
	Width  int
	Height int
}

// How tall a run of filled cells is, on one half row, and the cursor after the name, a half row, an underscore on the
// letters' foot, a glyph's gap after them
const (
	runHalves    = 1
	cursorHalves = 1
)

// Draw is texts in the font as a drawing, a row of nothing between two; not drawn where the font does not draw every
// letter of them.
func (font Font) Draw(texts []string) (Drawing, bool) {
	banner := font.Banner(texts)
	if len(banner) == 0 {
		return Drawing{Lines: nil, Rectangles: nil, Cursor: nil, Width: 0, Height: 0}, false
	}
	halves := halfRows(banner)
	drawing := Drawing{Lines: texts, Rectangles: nil, Cursor: nil, Width: 0, Height: len(halves)}
	for halfRow, filled := range halves {
		drawing.Width = max(drawing.Width, len(filled))
		drawing.Rectangles = append(drawing.Rectangles, halfRowRectangles(filled, halfRow)...)
	}
	return drawing, true
}

// Foot is how far down the drawing's ink reaches, in half rows from its top: a font may leave its last half row empty.
func (drawing Drawing) Foot() int {
	foot := 0
	for _, rectangle := range drawing.Rectangles {
		foot = max(foot, rectangle.Top+rectangle.Height)
	}
	return foot
}

// NameDrawing is the author's name in the name's font, on the lines the profile breaks it into (NameLines), with the
// cursor after it as wide as its last letter; not drawn where the font does not draw every letter of it.
func NameDrawing(name string) (Drawing, bool) {
	lines := NameLines(name)
	font := NameFont()
	drawing, drawn := font.Draw(lines)
	if !drawn {
		return drawing, false
	}
	last, _ := utf8.DecodeLastRuneInString(lines[len(lines)-1])
	return drawing.withCursor(font.LetterWidth(last)), true
}

// withCursor is the drawing with the cursor after its last line, a glyph's gap apart, as wide as a letter and a half
// row tall on the foot of the letters' ink.
func (drawing Drawing) withCursor(letterCells int) Drawing {
	foot, lastLineRight := drawing.Foot(), 0
	for _, rectangle := range drawing.Rectangles {
		if rectangle.Top+rectangle.Height == foot {
			lastLineRight = max(lastLineRight, rectangle.Left+rectangle.Width)
		}
	}
	cursor := Rectangle{Left: lastLineRight + glyphGap, Top: foot - cursorHalves, Width: letterCells, Height: cursorHalves}
	drawing.Cursor = &cursor
	drawing.Width = max(drawing.Width, cursor.Left+cursor.Width)
	return drawing
}

// halfRowRectangles are the runs of a half row's filled cells, as rectangles of the half row they stand on.
func halfRowRectangles(filled []bool, halfRow int) []Rectangle {
	rectangles := []Rectangle{}
	start := -1
	for column := 0; column <= len(filled); column++ {
		on := column < len(filled) && filled[column]
		if on && start < 0 {
			start = column
		}
		if !on && start >= 0 {
			rectangles = append(rectangles, Rectangle{Left: start, Top: halfRow, Width: column - start, Height: runHalves})
			start = -1
		}
	}
	return rectangles
}

// halfRows are a banner's cells as half rows, top to bottom, each a column's half filled or empty, two empty half rows
// for each row of nothing between two of its lines: the glyphs' ▀, ▄ and █ read as what they fill.
func halfRows(banner [][]string) [][]bool {
	halves := [][]bool{}
	for index, rows := range banner {
		if index > 0 {
			for range halvesInRow * BannerLineGap {
				halves = append(halves, nil)
			}
		}
		for _, row := range rows {
			upper, lower := rowHalves(row)
			halves = append(halves, upper, lower)
		}
	}
	return halves
}

// rowHalves are a row of glyphs' cells as its upper half row and its lower, each a cell's half filled or empty.
func rowHalves(row string) (upper, lower []bool) {
	for _, cell := range row {
		upper = append(upper, cell == upperHalf || cell == fullBlock)
		lower = append(lower, cell == lowerHalf || cell == fullBlock)
	}
	return upper, lower
}

// LetterWidth is how many cells wide a letter's glyph is in the font; none where the font does not draw it.
func (font Font) LetterWidth(letter rune) int {
	glyph := font.glyphs[letter]
	if len(glyph) == 0 {
		return 0
	}
	return utf8.RuneCountInString(glyph[0])
}

// NameLines are a name as the name's font stands it: its first word above the rest, in capitals; nothing where the
// name has no words.
func NameLines(name string) []string {
	words := strings.Fields(strings.ToUpper(name))
	if len(words) == 0 {
		return nil
	}
	lines := []string{words[0]}
	if len(words) > 1 {
		lines = append(lines, strings.Join(words[1:], glyphSpace))
	}
	return lines
}
