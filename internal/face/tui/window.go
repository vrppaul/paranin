package tui

import (
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
)

// The marks of every frame the TUI draws, a window over the page and the box a command stands in: square, as tmux
// draws its panes, its corners, its edges across and down, and where a rule inside it meets the left and right edges
const (
	frameTopLeft     = "┌"
	frameTopRight    = "┐"
	frameBottomLeft  = "└"
	frameBottomRight = "┘"
	frameAcross      = "─"
	frameDown        = "│"
	frameRuleLeft    = "├"
	frameRuleRight   = "┤"
)

// The room between a row's keys and what they do, in a window over the page
const itemGap = 2

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
	top := len(rows) - footerRows - len(window.rows)
	return laidOver(rows, window.rows, top, min(max(left, 0), model.frame.width-window.width), model.frame.width)
}

// shows is whether a window shows whole on the screen: no wider than it, and no taller than the rows the frame
// leaves the page.
func (model Model) shows(window paintedWindow) bool {
	return window.width <= model.frame.width && len(window.rows) <= model.frame.height-model.frame.above-footerRows
}

// framed is a window of a heading over rows, each painted as wide as its inside, with the border round them and a
// rule between the heading and the rows that meets the border at both sides.
func (paint paint) framed(heading string, body []string, inner int) paintedWindow {
	edge := paint.parts[partEdge]
	rows := []string{
		edge.Render(frameTopLeft + strings.Repeat(frameAcross, inner) + frameTopRight),
		edge.Render(frameDown) + heading + edge.Render(frameDown),
		edge.Render(frameRuleLeft + strings.Repeat(frameAcross, inner) + frameRuleRight),
	}
	for _, row := range body {
		rows = append(rows, edge.Render(frameDown)+row+edge.Render(frameDown))
	}
	rows = append(rows, edge.Render(frameBottomLeft+strings.Repeat(frameAcross, inner)+frameBottomRight))
	return paintedWindow{rows: rows, width: textWidth(frameDown) + inner + textWidth(frameDown)}
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
