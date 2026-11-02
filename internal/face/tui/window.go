package tui

import (
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
)

// The marks of a window over the page: the corners and sides of its border, the mark of the item chosen, which is
// the command line's cursor, and the room between an item's name and what it says
const (
	edgeTopLeft     = "╭"
	edgeTopRight    = "╮"
	edgeBottomLeft  = "╰"
	edgeBottomRight = "╯"
	edgeAcross      = "─"
	edgeDown        = "│"
	chosenMark      = cursorMark
	itemGap         = 2
)

// The order the window's layer is drawn in, its Z, above the page's, which is none: lipgloss draws a layer of a
// higher Z over one of a lower
const windowLayerOrder = 1

// paintedWindow is a window painted: its rows, the border's first and last, and the columns it takes.
type paintedWindow struct {
	rows  []string
	width int
}

// withCompletions lays the window of commands over the rows of a screen, its left edge by the page's left crosses,
// where it lists any.
func (model Model) withCompletions(rows []string) []string {
	if len(model.completion.listed) == 0 {
		return rows
	}
	return model.withWindow(rows, model.paint.commandWindow(model.completion.listed, model.completion.chosen), model.frame.left-crossRoom)
}

// withWindow lays a window over the rows of a screen, its bottom edge above the status line and its left edge at a
// column, held within the screen, where it shows.
func (model Model) withWindow(rows []string, window paintedWindow, left int) []string {
	if !model.shows(window) {
		return rows
	}
	top := len(rows) - rowsBelowPage - len(window.rows)
	return laidOver(rows, window.rows, top, min(max(left, 0), model.frame.width-window.width), model.frame.width)
}

// shows is whether a window shows whole on the screen: no wider than it, and no taller than the rows between the
// line of tabs and the status line.
func (model Model) shows(window paintedWindow) bool {
	return window.width <= model.frame.width && len(window.rows) <= model.frame.height-barRows
}

// commandWindow paints commands as a window, each command a row, its name after the colon in a column as wide as
// the widest and what it does after it; the first marked where Tab chose it.
func (paint paint) commandWindow(listed []command, firstChosen bool) paintedWindow {
	nameWidth, doesWidth := 0, 0
	for _, known := range listed {
		nameWidth, doesWidth = max(nameWidth, textWidth(commandMark+known.name)), max(doesWidth, textWidth(known.does))
	}
	inner := textWidth(chosenMark) + nameWidth + itemGap + doesWidth + spaceWidth
	inside := make([]string, 0, len(listed))
	for place, known := range listed {
		mark, name, does := piece{text: space, part: partWindowNote}, partWindowName, partWindowNote
		if firstChosen && place == 0 {
			mark, name, does = piece{text: chosenMark, part: partChosenMark}, partChosenName, partChosenNote
		}
		pieces := []piece{mark, {text: padded(commandMark+known.name, nameWidth+itemGap), part: name}, {text: padded(known.does, doesWidth+spaceWidth), part: does}}
		inside = append(inside, paint.cut(pieces, inner).text)
	}
	return paint.framed(inside, inner)
}

// framed is a window of rows, each painted as wide as its inside, with the border round them.
func (paint paint) framed(inside []string, inner int) paintedWindow {
	edge := paint.parts[partEdge]
	rows := []string{edge.Render(edgeTopLeft + strings.Repeat(edgeAcross, inner) + edgeTopRight)}
	for _, row := range inside {
		rows = append(rows, edge.Render(edgeDown)+row+edge.Render(edgeDown))
	}
	rows = append(rows, edge.Render(edgeBottomLeft+strings.Repeat(edgeAcross, inner)+edgeBottomRight))
	return paintedWindow{rows: rows, width: textWidth(edgeDown) + inner + textWidth(edgeDown)}
}

// laidOver is rows with a window laid over them from a row and a column, as lipgloss composes layers: only the
// rows the window covers are composed again, and they come back with their trailing blanks trimmed, which the
// terminal clears as Bubble Tea draws the screen.
func laidOver(rows, window []string, top, left, width int) []string {
	covered := rows[top : top+len(window)]
	page, over := lipgloss.NewLayer(strings.Join(covered, lineEnd)), lipgloss.NewLayer(strings.Join(window, lineEnd)).X(left).Z(windowLayerOrder)
	layers := lipgloss.NewCompositor(page, over)
	canvas := lipgloss.NewCanvas(width, len(window)).Compose(layers)
	return slices.Replace(slices.Clone(rows), top, top+len(window), strings.Split(canvas.Render(), lineEnd)...)
}
