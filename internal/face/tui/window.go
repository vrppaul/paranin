package tui

import (
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
)

// The marks of the window of commands: the corners and sides of its border, the mark of the command chosen, which
// is the command line's cursor, and the room between a command's name and what it does
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

// withCompletions lays the window of commands over the rows of a screen: its bottom edge above the status line,
// its left edge by the page's left crosses, where it fits between the line of tabs and the status line.
func (model Model) withCompletions(rows []string) []string {
	listed := model.completion.listed
	if len(listed) == 0 {
		return rows
	}
	window := model.paint.commandWindow(listed, model.completion.chosen)
	width := model.frame.width
	windowWidth, bottom := textWidth(window[0]), len(rows)-rowsBelowPage
	top := bottom - len(window)
	if top < rowsAbovePage || windowWidth > width {
		return rows
	}
	return laidOver(rows, window, top, min(max(model.frame.left-crossRoom, 0), width-windowWidth), width)
}

// commandWindow paints commands as a window with a border round it, each command a row, its name after the colon
// in a column as wide as the widest and what it does after it; the first marked where Tab chose it.
func (paint paint) commandWindow(listed []command, firstChosen bool) []string {
	nameWidth, doesWidth := 0, 0
	for _, known := range listed {
		nameWidth, doesWidth = max(nameWidth, textWidth(commandMark+known.name)), max(doesWidth, textWidth(known.does))
	}
	inner := textWidth(chosenMark) + nameWidth + itemGap + doesWidth + spaceWidth
	edge := paint.parts[partEdge]
	rows := []string{edge.Render(edgeTopLeft + strings.Repeat(edgeAcross, inner) + edgeTopRight)}
	for place, known := range listed {
		mark, name, does := piece{text: space, part: partWindowNote}, partWindowName, partWindowNote
		if firstChosen && place == 0 {
			mark, name, does = piece{text: chosenMark, part: partChosenMark}, partChosenName, partChosenNote
		}
		pieces := []piece{mark, {text: padded(commandMark+known.name, nameWidth+itemGap), part: name}, {text: padded(known.does, doesWidth+spaceWidth), part: does}}
		rows = append(rows, edge.Render(edgeDown)+paint.cut(pieces, inner).text+edge.Render(edgeDown))
	}
	return append(rows, edge.Render(edgeBottomLeft+strings.Repeat(edgeAcross, inner)+edgeBottomRight))
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
