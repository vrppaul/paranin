package look

import (
	"strings"
	"unicode/utf8"
)

// How a glyph's rows are written in a font's table
const (
	glyphBreak = "\n"
	glyphSpace = " "
)

// How a banner stands in a block font, in every face that draws one: the columns of nothing between two glyphs, and the
// rows of nothing between two lines
const (
	GlyphGap      = 1
	BannerLineGap = 1
)

// HalvesInRow are the half rows a row of cells holds, each a cell's upper half or its lower.
const HalvesInRow = 2

// The half blocks a glyph's cell is made of: the upper half, the lower, both
const (
	upperHalf = '▀'
	lowerHalf = '▄'
	fullBlock = '█'
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
				rows[row] += strings.Repeat(glyphSpace, GlyphGap)
			}
			rows[row] += glyphRow
		}
	}
	return rows
}

// HalfRows are a banner's cells as half rows, top to bottom, each a column's half filled or empty, two empty half rows
// for each row of nothing between two of its lines: the glyphs' ▀, ▄ and █ read as what they fill.
func HalfRows(banner [][]string) [][]bool {
	halves := [][]bool{}
	for index, rows := range banner {
		if index > 0 {
			for range HalvesInRow * BannerLineGap {
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
