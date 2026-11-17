package tui

import (
	"slices"
	"unicode"

	"github.com/vrppaul/paranin/internal/look"
)

// How many rows of a tab a page still shows under a banner or a title at its top
const rowsUnderTop = 4

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
			lines = slices.Concat(lines, slices.Repeat(blankLines(), look.BannerLineGap))
		}
		lines = slices.Concat(lines, tonedRows(written, shown, width))
	}
	return lines
}

// tonedRows are rows of a block font as lines of a kind, each cell in the tone its column stands at along a width,
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
