package tui

import (
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// The title of the window of keys, which the status line names it by too, and the word for what Esc does there
const (
	keysTitle = "keys"
	closeHint = "close"
)

// The count shown before a key in the row that says what a count does, and the mark between the first tab's number
// and the last's
const (
	countExample = "3"
	keyRangeMark = "-"
)

// windowKeys are the keys that open the window of keys and, while it is open, close it.
func windowKeys() []string {
	return []string{keyWindowKey, spaceKey}
}

// keyNotes are the rows of the window of keys, the keys of normal mode in the order a reader learns them; the arrows,
// Home, End, Shift+Tab and Ctrl+C, which only stand in for keys the window names, are left out.
func (model Model) keyNotes() []hint {
	lastTab := len(model.tabs) - 1 + firstTabNumber
	return []hint{
		{key: downKey + space + upKey, does: "line down / up"},
		{key: countExample + downKey, does: "move by a count"},
		{key: topKey + topKey + space + bottomKey, does: "top / bottom"},
		{key: previousKey + space + nextKey, does: "tab before / after"},
		{key: tabKey, does: "next tab"},
		{key: strconv.Itoa(firstTabNumber) + keyRangeMark + strconv.Itoa(lastTab), does: "go to a tab"},
		{key: searchKey, does: "find in the tab"},
		{key: nextMatchKey + space + previousMatchKey, does: "next / prior match"},
		{key: commandKey, does: "command line"},
		{key: escapeKey, does: "end the search"},
		{key: quitKey, does: quitHint},
	}
}

// openKeys opens the window of keys where it shows whole on the screen; where it does not, the keys stay in normal
// mode, so no mode opens unseen.
func (model Model) openKeys(int) (Model, tea.Cmd) {
	if model.shows(model.drawn.keyWindow) {
		model.keys.mode = modeKeys
	}
	return model, nil
}

// keyInWindow is the model after a key while the window of keys is open: the key closes it, and does what it does
// in normal mode, but for Esc and the keys that open the window, which do nothing more.
func (model Model) keyInWindow(key string) (Model, tea.Cmd) {
	model.keys.mode = modeNormal
	if key == escapeKey || slices.Contains(windowKeys(), key) {
		return model, nil
	}
	return model.normalKey(key)
}

// withKeys lays the window of keys over the rows of a screen, its right edge by the page's right crosses.
func (model Model) withKeys(rows []string) []string {
	window := model.drawn.keyWindow
	return model.withWindow(rows, window, model.frame.left+model.frame.inner+crossRoom-window.width)
}

// keyWindow paints the window of keys: its title, and how it closes, at the right; a rule; then each row, its keys
// in a column as wide as the widest and what they do after it.
func (paint paint) keyWindow(notes []hint) paintedWindow {
	keysWidth, doesWidth := 0, 0
	for _, note := range notes {
		keysWidth, doesWidth = max(keysWidth, textWidth(note.key)), max(doesWidth, textWidth(note.does))
	}
	inner := spaceWidth + keysWidth + itemGap + doesWidth + spaceWidth
	closing := escapeKey + space + closeHint + space
	title := []piece{{text: padded(space+keysTitle, inner-textWidth(closing)), part: partWindowNote}, {text: closing, part: partWindowNote}}
	rule := []piece{{text: strings.Repeat(edgeAcross, inner), part: partWindowRule}}
	inside := []string{paint.cut(title, inner).text, paint.cut(rule, inner).text}
	for _, note := range notes {
		pieces := []piece{{text: space + padded(note.key, keysWidth+itemGap), part: partWindowName}, {text: padded(note.does, doesWidth+spaceWidth), part: partWindowNote}}
		inside = append(inside, paint.cut(pieces, inner).text)
	}
	return paint.framed(inside, inner)
}
