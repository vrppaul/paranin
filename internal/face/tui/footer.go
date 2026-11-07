package tui

import (
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
)

// The marks of the footer: the rim, the panel's ground rising half a row above it, and the rule that stands in for it
// where the terminal shows few colours; the mark before the word of the lines below; and the prompt's, the place it
// stands in, home, as a shell's prompt names it, and the mark it ends in, which the search's slash takes the place of
const (
	rimMark     = "▄"
	rimRuleMark = "─"
	moreMark    = "▾"
	homeMark    = "~"
	promptMark  = "❯"
)

// The mark that stands among the places for those that give way where they do not all fit
const placesLeftMark = "…"

// The words of the footer: Neovim's for the lines below the page, by how many there are; the key that completes a
// place, and what it does; and what Enter does on the prompt
const (
	moreLines       = "more lines"
	moreLine        = "more line"
	completeKeyName = "Tab"
	completesWord   = "completes"
	runHint         = "run"
	findHint        = "find"
	cancelHint      = "cancel"
	enterMark       = "↵"
)

// The room the places stand apart by, and the room after what Tab does before the first of them
const (
	placeGap     = 3
	completesGap = 4
)

// The columns the terminal's cursor takes on the prompt, after what is typed
const cursorWidth = 1

// rimRow is the rim between the page and the panel: bare where the page shows the tab to its end, and where more of
// it lies below, saying in the middle how many lines and the key that scrolls to them.
func (model Model) rimRow() string {
	below := model.linesBelow()
	if below == 0 {
		return model.drawn.rim
	}
	lines := moreLines
	if below == 1 {
		lines = moreLine
	}
	said := model.paint.cut([]piece{
		{text: space + moreMark + space, part: partMoreMark},
		{text: strconv.Itoa(below) + space + lines + moreSeparator, part: partMore},
		{text: downKey, part: partMoreKey},
		{text: space, part: partMore},
	}, model.frame.width)
	left := (model.frame.width - said.width) / sides
	return model.paint.rim(left) + said.text + model.paint.rim(model.frame.width-left-said.width)
}

// promptLine is the prompt as the reader types on it, after the slash of a search or the prompt's own mark, with the
// terminal's cursor after what they typed; or, where they type nothing, what the last key said, alone on the row as
// Neovim's command line says it, so it has the row's whole width, or the prompt as it waits, ending in the site's
// cursor, and no terminal's cursor.
func (model Model) promptLine() (string, *tea.Cursor) {
	switch model.keys.mode {
	case modeSearch:
		return model.typedLine(searchMark)
	case modeCommand:
		return model.typedLine(promptMark + space)
	case modeNormal, modeKeys:
	}
	switch {
	case model.said.failed:
		return model.paint.panelRow([]piece{{text: model.said.text, part: partFailure}}, model.frame), nil
	case model.said.text != "":
		return model.paint.panelRow([]piece{{text: model.said.text, part: partSaid}}, model.frame), nil
	}
	return model.drawn.idlePrompt, nil
}

// promptOpening is what the prompt opens with: the site's host, the place it stands in, and a mark.
func (model Model) promptOpening(mark string) []piece {
	return []piece{
		{text: model.site.Host(), part: partHost},
		{text: space + homeMark + space, part: partHome},
		{text: mark, part: partPromptMark},
	}
}

// typedLine is the prompt as the reader types on it after a mark, and the terminal's own cursor after what they
// typed, held within the screen: a bar the terminal blinks, as the first version's cursor blinked, in the accent
// where the terminal shows many colours; none where the screen is too short to show the prompt.
func (model Model) typedLine(mark string) (string, *tea.Cursor) {
	opening := model.promptOpening(mark)
	shown := model.shownTyping(piecesWidth(opening))
	row := model.paint.panelRow(append(opening, piece{text: shown, part: partTyped}), model.frame)
	line, _ := model.frame.rowOf(footerPrompt)
	if line < 0 {
		return row, nil
	}
	column := min(model.frame.left+piecesWidth(opening)+textWidth(shown), model.frame.width-cursorWidth)
	position := tea.Position{X: max(column, 0), Y: line}
	return row, &tea.Cursor{Position: position, Color: model.paint.cursorColour, Shape: tea.CursorBar, Blink: true}
}

// shownTyping is what the prompt shows of what is typed after its opening: its end, where the line is too narrow
// for all of it and the cursor after it.
func (model Model) shownTyping(openingWidth int) string {
	room := model.frame.width - model.frame.left - openingWidth - cursorWidth
	typed := model.keys.typing
	for textWidth(typed) > max(room, 0) {
		_, first := utf8.DecodeRuneInString(typed)
		typed = typed[first:]
	}
	return typed
}

// placesRow is the row under the prompt: in normal mode the places the tab chosen may go to; while the reader types
// a place or a command, those that start with what is typed, the one Tab chose marked, or, where none does, what
// Enter and Esc do; while they type a search, what Enter and Esc do.
func (model Model) placesRow(drawn drawnTab) string {
	switch model.keys.mode {
	case modeSearch:
		return model.paint.panelRow(hintPieces([]hint{{key: enterMark, does: findHint}, {key: escapeKey, does: cancelHint}}), model.frame)
	case modeCommand:
		if len(model.completion.listed) == 0 {
			return model.paint.panelRow(hintPieces([]hint{{key: enterMark, does: runHint}, {key: escapeKey, does: cancelHint}}), model.frame)
		}
		name, argument, argued := commandParts(model.completion.typed)
		if argued {
			name = argument
		}
		return model.paint.panelRow(model.placePieces(model.completion.listed, model.completion.chosen, name), model.frame)
	case modeNormal, modeKeys:
	}
	return drawn.places
}

// placePieces are what Tab does, then the places listed, the start the reader typed of each marked, and the one Tab
// chose. Where they would not all fit, the first ones give way until the chosen one does; where none is chosen, the
// last one stays in sight, help among the places listed idle, and those before it give way from its side, a mark
// standing for them.
func (model Model) placePieces(listed []completion, chosen int, typed string) []piece {
	opening := []piece{
		{text: completeKeyName, part: partPanelKey},
		{text: space + completesWord + strings.Repeat(space, completesGap), part: partPanelNote},
	}
	room := model.frame.width - model.frame.left - piecesWidth(opening)
	items := placeItems(listed, chosen, typed)
	shown := items
	for first := 1; piecesWidth(joinedItems(shown)) > room && first <= chosen; first++ {
		shown = items[first:]
	}
	last := len(items) - 1
	for kept := last - 1; piecesWidth(joinedItems(shown)) > room && chosen == noneChosen && kept >= 0; kept-- {
		shown = slices.Concat(items[:kept], [][]piece{{{text: placesLeftMark, part: partPlace}}}, items[last:])
	}
	return append(opening, joinedItems(shown)...)
}

// placeItems are the pieces of each place listed: the start the reader typed of it marked, or all of it where Tab
// chose it.
func placeItems(listed []completion, chosen int, typed string) [][]piece {
	items := make([][]piece, 0, len(listed))
	for index, shown := range listed {
		switch {
		case index == chosen:
			items = append(items, []piece{{text: shown.shown, part: partPlaceChosen}})
		case typed != "" && strings.HasPrefix(shown.shown, typed):
			items = append(items, []piece{{text: typed, part: partPlaceTyped}, {text: shown.shown[len(typed):], part: partPlaceRest}})
		default:
			items = append(items, []piece{{text: shown.shown, part: partPlace}})
		}
	}
	return items
}

// joinedItems are the pieces of the items of a row of the panel, places or keys, a gap between each two.
func joinedItems(items [][]piece) []piece {
	gap := piece{text: strings.Repeat(space, placeGap), part: partPanel}
	pieces := []piece{}
	for index, item := range items {
		if index > 0 {
			pieces = append(pieces, gap)
		}
		pieces = append(pieces, item...)
	}
	return pieces
}

// hintPieces are keys and what each does, a gap between each two, as between places.
func hintPieces(hints []hint) []piece {
	items := make([][]piece, 0, len(hints))
	for _, named := range hints {
		items = append(items, []piece{{text: named.key, part: partPanelKey}, {text: space + named.does, part: partPanelNote}})
	}
	return joinedItems(items)
}

// panelRow paints pieces as a row of the panel, standing on the page's left edge, cut at the screen's and filled with
// the panel's ground.
func (paint paint) panelRow(pieces []piece, frame frame) string {
	gutter := paint.parts[partPanel].Render(strings.Repeat(space, frame.left))
	painted := paint.cut(pieces, frame.width-frame.left)
	return gutter + painted.text + paint.parts[partPanel].Render(strings.Repeat(space, max(frame.width-frame.left-painted.width, 0)))
}

// rim paints the rim across a width: the panel's ground rising half a row into it, or with few colours, where the
// panel has no ground of its own, a rule.
func (paint paint) rim(width int) string {
	mark := rimMark
	if paint.fewColours {
		mark = rimRuleMark
	}
	return paint.parts[partRim].Render(strings.Repeat(mark, max(width, 0)))
}
