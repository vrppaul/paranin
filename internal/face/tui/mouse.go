package tui

import (
	"slices"

	tea "charm.land/bubbletea/v2"
)

// The lines a step of the wheel scrolls, as Neovim's mousescroll ver:3
const wheelLines = 3

// clicked is the model after a press of a button of the mouse, as Neovim with mouse=nvi and tmux with mouse on take
// it, in normal mode alone and from the left button alone: on the footer what the key its part names does, on a link
// of the blog's list what Enter on it does. With the window of keys open any click closes it, as any key does, and does
// nothing more, since the window hides what lies under it.
func (model Model) clicked(click tea.Mouse) (Model, tea.Cmd) {
	switch model.keys.mode {
	case modeSearch, modeCommand:
		return model, nil
	case modeKeys:
		model.keys.mode = modeNormal
		return model, nil
	case modeNormal:
	}
	if click.Button != tea.MouseLeft {
		return model, nil
	}
	model = model.pendingAndSaidCleared()
	if click.Y == model.frame.footerAt() {
		return model.footerClicked(click.X)
	}
	return model.pageClicked(click.Y-model.frame.pageTop(), click.X-model.frame.left)
}

// footerClicked is the model after a click on a column of the footer: a window chosen, as its digit chooses it; a
// session not shown shown as its reader left it; the window of keys opened, as ? opens it; or nothing.
func (model Model) footerClicked(column int) (Model, tea.Cmd) {
	spots := model.drawn.tabs[model.chosen].footerSpots
	found := slices.IndexFunc(spots, func(placed footerSpot) bool { return placed.columns.holds(column) })
	if found < 0 {
		return model, nil
	}
	switch target := spots[found].leads.(type) {
	case windowTarget:
		return model.choose(target.place), nil
	case sessionTarget:
		return model.switchedTo(target.session), nil
	case keysTarget:
		return model.openKeys()
	case noTarget:
	}
	return model, nil
}

// pageClicked is the model after a click on a row of the page and a column of it: on the blog's list, a link there
// stood on, as j and k stand on it, and followed, as Enter follows it; nothing anywhere else.
func (model Model) pageClicked(row, column int) (Model, tea.Cmd) {
	if row < 0 || row >= model.frame.rows {
		return model, nil
	}
	at := spot{line: model.positions[model.chosen].scroll + row, column: column}
	found := slices.IndexFunc(model.laid[model.chosen].links, func(link laidLink) bool { return link.holds(at) })
	if found < 0 {
		return model, nil
	}
	if found != model.linkStoodOn(model.chosen) {
		model = model.choseLink(found)
	}
	return model.followLink()
}

// wheeled is the model after a step of the wheel, in normal mode alone: the page scrolled three lines down or up, as
// the keys scroll it, which on the blog's list keeps the link the keys stand on in sight. Its left and right, and any
// button reported as a wheel's, do nothing.
func (model Model) wheeled(wheel tea.Mouse) (Model, tea.Cmd) {
	if model.keys.mode != modeNormal {
		return model, nil
	}
	lines := wheelLines
	switch wheel.Button {
	case tea.MouseWheelDown:
	case tea.MouseWheelUp:
		lines = -wheelLines
	case tea.MouseNone, tea.MouseLeft, tea.MouseMiddle, tea.MouseRight, tea.MouseWheelLeft, tea.MouseWheelRight,
		tea.MouseBackward, tea.MouseForward, tea.MouseButton10, tea.MouseButton11:
		return model, nil
	}
	model = model.pendingAndSaidCleared()
	return model.scrolledTo(model.positions[model.chosen].scroll + lines), nil
}
