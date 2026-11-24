// Package look is the design's tokens as the faces read them (docs/FOUNDATION.md §5.5). palette.go, which
// scripts/tokens writes from design/tokens.yaml, names each colour and gives its shade in each theme, cells.go each
// length a terminal lays out by, in cells, and shares.go each share of the screen's height it measures by, in percent;
// this file holds the types the palette fills, and how a day reads. No colour of the design is written anywhere else
// in Go. Its generate writes the page's tokens.css, times.ts and shares.ts too, under web/src/tokens, from the same
// tokens. letters.go holds the design's block fonts, in which the TUI and the page both draw the author's name and the
// titles, each its own way; glyphs.go the terminal's glyphs both faces draw.
package look

import (
	"image/color"
	"math"
)

//go:generate go -C ../.. run ./scripts/tokens -write

// Theme is the light theme or the dark one, chosen by the reader's background.
type Theme int

const (
	// ThemeLight is for a light background: Catppuccin Latte
	ThemeLight Theme = iota
	// ThemeDark is for a dark background: Catppuccin Macchiato
	ThemeDark
)

// Colour is one of the design's colours, named for what it does; palette.go declares them.
type Colour int

// Shade is one colour as each kind of terminal shows it: whole, as the nearest of xterm's 256 colours, and as
// one of the sixteen a basic terminal names, or the terminal's own.
type Shade struct {
	TrueColour color.RGBA
	Indexed    uint8
	Basic      Basic
}

// Basic is one of a terminal's sixteen colours, 0 to 15, or BasicOwn.
type Basic uint8

// BasicOwn is the terminal's own colour, where no basic colour serves: a number no colour of sixteen has
const BasicOwn Basic = math.MaxUint8
