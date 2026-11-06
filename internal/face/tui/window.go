package tui

import (
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
)

// The marks of a window over the page: the corners and sides of its border, and the room between a row's keys and
// what they do
const (
	edgeTopLeft     = "╭"
	edgeTopRight    = "╮"
	edgeBottomLeft  = "╰"
	edgeBottomRight = "╯"
	edgeAcross      = "─"
	edgeDown        = "│"
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

// withWindow lays a window over the rows of a screen, its bottom edge above the footer and its left edge at a
// column, held within the screen, where it shows.
func (model Model) withWindow(rows []string, window paintedWindow, left int) []string {
	if !model.shows(window) {
		return rows
	}
	top := len(rows) - len(model.frame.footer) - len(window.rows)
	return laidOver(rows, window.rows, top, min(max(left, 0), model.frame.width-window.width), model.frame.width)
}

// shows is whether a window shows whole on the screen: no wider than it, and no taller than the rows the frame
// leaves the page.
func (model Model) shows(window paintedWindow) bool {
	return window.width <= model.frame.width && len(window.rows) <= model.frame.height-model.frame.above-len(model.frame.footer)
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
