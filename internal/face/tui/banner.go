package tui

import (
	"slices"
	"strings"
	"unicode"
)

// How a banner stands in the block font: how many rows a glyph is tall, a column between two glyphs, and a row of
// nothing between two of its lines; and how many rows of a tab a page still shows under a banner or a title at its
// top
const (
	glyphRows    = 3
	glyphGap     = 1
	blockLineGap = 1
	rowsUnderTop = 4
)

// blockFont is the first version's banner font: each capital, digit, space and hyphen three rows tall, of half blocks,
// each glyph's rows written as one text, a line's end between each two. A text with a letter it has not is drawn as
// text.
func blockFont() map[rune][]string {
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
	font := make(map[rune][]string, len(glyphs))
	for letter, glyph := range glyphs {
		font[letter] = strings.Split(glyph, lineEnd)
	}
	return font
}

// bannerRows are texts in the block font, each text its rows of glyphs; nothing where a text has a letter the font
// does not draw.
func bannerRows(texts []string) [][]string {
	font := blockFont()
	rows := make([][]string, 0, len(texts))
	for _, text := range texts {
		for _, letter := range text {
			if _, drawn := font[letter]; !drawn {
				return nil
			}
		}
		rows = append(rows, blockRows(font, text))
	}
	return rows
}

// blockRows are a text's rows in the block font, a column between each two glyphs; every letter of it is one the
// font draws.
func blockRows(font map[rune][]string, text string) []string {
	rows := make([]string, glyphRows)
	for index, letter := range text {
		glyph := font[letter]
		for row := range rows {
			if index > 0 {
				rows[row] += strings.Repeat(space, glyphGap)
			}
			rows[row] += glyph[row]
		}
	}
	return rows
}

// rowsWidth is the widest of rows of text.
func rowsWidth(rows []string) int {
	width := 0
	for _, row := range rows {
		width = max(width, textWidth(row))
	}
	return width
}

// tonedBanner lays texts' rows of glyphs out as lines of a kind, a row of nothing between two texts, each cell in the
// tone its column stands at along a width, so the colours of every line of the banner run alike.
func tonedBanner(rows [][]string, shown kind, width int) []line {
	lines := []line{}
	for index, written := range rows {
		if index > 0 {
			lines = slices.Concat(lines, slices.Repeat(blankLines(), blockLineGap))
		}
		lines = slices.Concat(lines, tonedRows(written, shown, width))
	}
	return lines
}

// tonedRows are rows of the block font as lines of a kind, each cell in the tone its column stands at along a width,
// cells of one tone side by side one run.
func tonedRows(rows []string, shown kind, width int) []line {
	lines := make([]line, 0, len(rows))
	for _, row := range rows {
		toned := line{}
		for column, letter := range []rune(row) {
			tone := toneAt(column, width)
			if last := len(toned) - 1; last >= 0 && toned[last].tone == tone {
				toned[last].text += string(letter)
				continue
			}
			toned = append(toned, run{text: string(letter), kind: shown, highlighted: false, address: "", copied: "", tone: tone})
		}
		lines = append(lines, toned)
	}
	return lines
}

// gradientWords are the words of a text of a kind, each letter in the tone its place along the whole text stands at,
// so the colours run on across the words wherever a line breaks them.
func gradientWords(text string, shown kind) []word {
	letters := []rune(text)
	words, current := []word{}, word{runs: nil, highlightBefore: false}
	for index, letter := range letters {
		if unicode.IsSpace(letter) {
			words, current = appendWord(words, current), word{runs: nil, highlightBefore: false}
			continue
		}
		piece := run{text: string(letter), kind: shown, highlighted: false, address: "", copied: "", tone: toneAt(index, len(letters))}
		current.runs = append(current.runs, piece)
	}
	return appendWord(words, current)
}

// appendWord is words with one more after them, where that one holds any text.
func appendWord(words []word, next word) []word {
	if len(next.runs) == 0 {
		return words
	}
	return append(words, next)
}

// toneAt is the tone of a column along a width: the first at its start, the last at its end.
func toneAt(column, width int) int {
	return column * (gradientTones - 1) / max(width-1, 1)
}

// fits is whether lines at the top of a tab fit its page: no wider than it, and with as many rows of the tab still
// showing under them as rowsUnderTop keeps.
func (layout layout) fits(lines []line) bool {
	return linesWidth(lines) <= layout.width && len(lines)+rowsUnderTop <= layout.pageRows
}
