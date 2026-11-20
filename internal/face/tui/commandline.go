package tui

import (
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
)

// The command line's words for the keys it names: what Enter does on a place or a command, and on a search; what Tab
// does; what Esc does
const (
	enterMark  = "↵"
	goHint     = "go"
	findHint   = "find"
	nextHint   = "next"
	cancelHint = "cancel"
)

// The room between two keys the command line names, and between the count of the matches and the first of them
const (
	hintGap  = 3
	countGap = 4
)

// The columns the terminal's cursor takes on the command line, after what is typed
const cursorWidth = 1

// commandLine is the command line as the keys leave it, over the footer as Neovim's stands with no row of its own:
// the row painted, nothing where it does not show; and where the reader types on it, the column what is typed ends
// at, which the menu over it stands by, and the terminal's cursor.
type commandLine struct {
	row      string
	typedEnd int
	cursor   *tea.Cursor
}

// commandLine is the command line while the reader types after a slash or a colon; or, where they type nothing,
// what the last key said, alone on the row, so it has the row's whole width; or nothing.
func (model Model) commandLine() commandLine {
	switch model.keys.mode {
	case modeSearch:
		return model.typedLine(searchMark, model.searchHints())
	case modeCommand:
		return model.typedLine(commandKey, model.commandHints())
	case modeNormal, modeKeys:
	}
	said := piece{text: model.said.text, part: partSaid}
	if model.said.failed {
		said.part = partFailure
	}
	if said.text == "" {
		return commandLine{row: "", typedEnd: 0, cursor: nil}
	}
	return commandLine{row: model.paint.panelLine([]piece{said}, nil, model.drawn.footer.place.margin, model.frame.width), typedEnd: 0, cursor: nil}
}

// searchHints are what the command line names at its right while the reader types a search: which of how many
// matches the page shows, where what is typed matches, as the chosen tab marks it, then what Enter and Esc do.
func (model Model) searchHints() []piece {
	hints := hintPieces([]hint{{key: enterMark, does: findHint}, {key: escapeKey, does: cancelHint}})
	matches := model.drawn.tabs[model.chosen].matches
	if len(matches) == 0 {
		return hints
	}
	count := matchCount(matchAtOrAfter(matches, model.search.begunAt), len(matches))
	return append([]piece{{text: count, part: partCount}, {text: strings.Repeat(space, countGap), part: partPanel}}, hints...)
}

// commandHints are what the command line names at its right while the reader types a place or a command: Tab where
// there are places to go round, then what Enter and Esc do.
func (model Model) commandHints() []piece {
	hints := []hint{{key: enterMark, does: goHint}, {key: escapeKey, does: cancelHint}}
	if len(model.completion.listed) > 0 {
		hints = append([]hint{{key: tabKey, does: nextHint}}, hints...)
	}
	return hintPieces(hints)
}

// typedLine is the command line as the reader types on it after a mark, the keys it names at its right where they
// fit beside what is typed, and the terminal's own cursor after what they typed, held within the screen: a bar the
// terminal blinks, so the program writes nothing to blink it and an idle session's timeout still runs out, in the
// accent where the terminal shows many colours; none where the screen has no row for it.
func (model Model) typedLine(mark string, hints []piece) commandLine {
	margin := model.drawn.footer.place.margin
	opening := []piece{{text: mark, part: partCommandMark}}
	room := model.frame.width - sides*margin - piecesWidth(opening) - cursorWidth
	if textWidth(model.keys.typing)+hintGap+piecesWidth(hints) > room {
		hints = nil
	}
	typed := shownTyping(model.keys.typing, room)
	row := model.paint.panelLine(append(opening, piece{text: typed, part: partTyped}), hints, margin, model.frame.width)
	end := margin + piecesWidth(opening) + textWidth(typed)
	line, shown := model.frame.commandRow()
	if !shown {
		return commandLine{row: row, typedEnd: end, cursor: nil}
	}
	position := tea.Position{X: max(min(end, model.frame.width-cursorWidth), 0), Y: line}
	return commandLine{row: row, typedEnd: end, cursor: &tea.Cursor{Position: position, Color: model.paint.cursorColour, Shape: tea.CursorBar, Blink: true}}
}

// shownTyping is what the command line shows of what is typed: its end, where the room is too narrow for all of it.
func shownTyping(typed string, room int) string {
	for textWidth(typed) > max(room, 0) {
		_, first := utf8.DecodeRuneInString(typed)
		typed = typed[first:]
	}
	return typed
}

// hintPieces are keys and what each does, a gap between each two.
func hintPieces(hints []hint) []piece {
	gap := piece{text: strings.Repeat(space, hintGap), part: partPanel}
	pieces := []piece{}
	for index, named := range hints {
		if index > 0 {
			pieces = append(pieces, gap)
		}
		pieces = append(pieces, piece{text: named.key, part: partPanelKey}, piece{text: space + named.does, part: partPanelNote})
	}
	return pieces
}

// panelLine paints a row of the panel of a width: pieces from a margin in at its left, which have the row to its
// edge where none stand at its right, as what a key says has; pieces ending a margin in from its right, cut at the
// room left; and the panel's ground between and around them.
func (paint paint) panelLine(left, right []piece, margin, width int) string {
	ground := paint.parts[partPanel]
	opening := paint.cut(left, width-margin)
	ending := paint.cut(right, width-sides*margin-opening.width)
	endMargin := 0
	if ending.width > 0 {
		endMargin = margin
	}
	fill := strings.Repeat(space, max(width-margin-opening.width-ending.width-endMargin, 0))
	return ground.Render(strings.Repeat(space, margin)) + opening.text + ground.Render(fill) + ending.text + ground.Render(strings.Repeat(space, endMargin))
}
