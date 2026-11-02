package tui

import (
	tea "charm.land/bubbletea/v2"
)

// What the command line says after y: the value copied, written after these words for a terminal that takes no
// OSC 52, as the first version wrote it; or that no line found holds a value, around the key that finds one
const (
	copiedOpening        = "copied: "
	nothingToCopyOpening = "nothing to copy: find a value with "
	nothingToCopyClosing = " first"
)

// copiedValue copies the value on the line the search found, as Neovim's yy yanks a line: the match a search or n
// showed stands in for the cursor this TUI has not. The value goes to the reader's clipboard over OSC 52, which the
// terminal at their end takes.
func (model Model) copiedValue(int) (Model, tea.Cmd) {
	value, found := model.valueFound()
	if !found {
		model.said = commandMessage{text: nothingToCopyOpening + searchKey + nothingToCopyClosing, failed: true}
		return model, nil
	}
	model.said = commandMessage{text: copiedOpening + value, failed: false}
	return model, tea.SetClipboard(value)
}

// valueFound is the value of the fact on the line of the match standing on the chosen tab, and whether that line
// holds one.
func (model Model) valueFound() (string, bool) {
	shown, standing := model.standingMatch()
	if !standing {
		return "", false
	}
	found := model.laid[model.chosen][model.drawn.tabs[model.chosen].matches[shown.place]]
	for _, piece := range found {
		if piece.copied != "" {
			return piece.copied, true
		}
	}
	return "", false
}
