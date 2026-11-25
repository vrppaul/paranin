package tui

import (
	"slices"
	"strconv"

	tea "charm.land/bubbletea/v2"
)

// The title of the window of keys, and the word for what Esc does there, which the window says of q on any page but
// the résumé
const (
	keysTitle = "keys"
	closeHint = "close"
)

// The name the window of keys goes by on every page's command line, as Neovim's :help opens its own: the command's,
// and its place's, which is how the menu shows the command
const keysPlace = "help"

// The mark between the first tab's number and the last's
const keyRangeMark = "-"

// What the command line says where the window of keys has no room on the screen, as Neovim says it of a window
const notEnoughRoom = "E36: Not enough room"

// windowKeys are the keys that open the window of keys and, while it is open, close it.
func windowKeys() []string {
	return []string{keyWindowKey, spaceKey}
}

// keyNotes are the rows of the window of keys on a tab, the keys of normal mode in the order a reader learns them: on
// the blog's list j and k go from link to link, its ways, its tags and its posts, and Enter follows one; the keys
// between tabs only where more than one tab shows; and what q does on the tab's page. The arrows, Home, End, Shift+Tab
// and Ctrl+C, which only stand in for keys the window names, are left out.
func (model Model) keyNotes(place int) []hint {
	moving := []hint{{key: downKey + space + upKey, does: "line down / up"}}
	if len(model.laid[place].links) > 0 {
		moving = []hint{{key: downKey + space + upKey, does: "link after / before"}, {key: enterMark, does: "follow the link"}}
	}
	notes := append(moving, hint{key: topKey + topKey + space + bottomKey, does: "top / bottom"})
	if lastTab := len(model.tabs) - 1 + firstTabNumber; lastTab > firstTabNumber {
		notes = append(notes,
			hint{key: previousKey + space + nextKey, does: "tab before / after"},
			hint{key: strconv.Itoa(firstTabNumber) + keyRangeMark + strconv.Itoa(lastTab), does: "go to a tab"},
		)
	}
	return append(notes,
		hint{key: commandKey, does: "type a place to go"},
		hint{key: tabKey, does: "complete a place"},
		hint{key: searchKey, does: "find in the tab"},
		hint{key: nextMatchKey + space + previousMatchKey, does: "next / prior match"},
		hint{key: copyKey, does: "copy a found value"},
		hint{key: escapeKey, does: "end the search"},
		hint{key: quitKey, does: closingHint(model.tabs[place].page)},
	)
}

// openKeys opens the window of keys where it shows whole on the screen; where it does not, the keys stay in normal
// mode, so no mode opens unseen, and the command line says there is no room, as Neovim does.
func (model Model) openKeys() (Model, tea.Cmd) {
	if !model.shows(model.drawn.tabs[model.chosen].keyWindow) {
		model.said = commandMessage{text: notEnoughRoom, failed: true}
		return model, nil
	}
	model.keys.mode = modeKeys
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
	window := model.drawn.tabs[model.chosen].keyWindow
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
	rows := []string{}
	for _, note := range notes {
		pieces := []piece{{text: space + padded(note.key, keysWidth+itemGap), part: partWindowName}, {text: padded(note.does, doesWidth+spaceWidth), part: partWindowNote}}
		rows = append(rows, paint.cut(pieces, inner).text)
	}
	return paint.framed(paint.cut(title, inner).text, rows, inner)
}
