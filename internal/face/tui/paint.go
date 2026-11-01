package tui

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"github.com/vrppaul/paranin/internal/look"
)

// part is a part of the frame around the page, which the paint gives a look of its own.
type part int

const (
	// partBar is the ground of the line of tabs and of the status line
	partBar part = iota
	// partTabNumber and partTabTitle are a tab not chosen: its number and its title
	partTabNumber
	partTabTitle
	// partChosenNumber and partChosenTitle are the chosen tab, which stands on the page's own ground
	partChosenNumber
	partChosenTitle
	// partHost is the site's name at the end of the line of tabs
	partHost
	// partMode is the block that opens the status line
	partMode
	// partStatus is what the status line says, and partStatusRule the rule between its parts
	partStatus
	partStatusRule
	// partHintKey is a key the status line names
	partHintKey
	// partCross is a corner cross around the page
	partCross
)

// paint is how each kind of text and each part of the frame looks in one terminal: its theme, and the colours it
// shows. With sixteen colours or none, what colour alone would tell is told by reversing or underlining too.
type paint struct {
	kinds map[kind]lipgloss.Style
	parts map[part]lipgloss.Style
	// highlight is a result worth seeing first: the tint behind it, and with few colours the text on it darkened,
	// with none underlined
	highlightGround color.Color
	highlightText   color.Color
	fewColours      bool
	noColours       bool
}

// paintFor is the paint of a theme in a terminal that shows a profile of colours.
func paintFor(theme look.Theme, profile colorprofile.Profile) paint {
	colour := colours{shades: look.Palette(theme), profile: profile}.of
	// colorprofile orders its profiles from none to full colour
	fewColours, noColours := profile < colorprofile.ANSI256, profile < colorprofile.ANSI
	plain := lipgloss.NewStyle()
	bar := plain.Background(colour(look.ColourSunk))
	return paint{
		kinds: map[kind]lipgloss.Style{
			kindText:   plain.Foreground(colour(look.ColourText)),
			kindMuted:  plain.Foreground(colour(look.ColourMuted)),
			kindFaint:  plain.Foreground(colour(look.ColourFaint)).Faint(noColours),
			kindStrong: plain.Foreground(colour(look.ColourText)).Bold(true),
			kindCode:   plain.Foreground(colour(look.ColourAccent)),
		},
		parts: map[part]lipgloss.Style{
			partBar:          bar,
			partTabNumber:    bar.Foreground(colour(look.ColourFaint)),
			partTabTitle:     bar.Foreground(colour(look.ColourMuted)),
			partChosenNumber: plain.Foreground(colour(look.ColourFaint)).Reverse(fewColours),
			partChosenTitle:  plain.Foreground(colour(look.ColourText)).Bold(true).Reverse(fewColours),
			partHost:         bar.Foreground(colour(look.ColourFaint)),
			partMode:         plain.Foreground(colour(look.ColourOnAccent)).Background(colour(look.ColourAccent)).Bold(true).Reverse(noColours),
			partStatus:       bar.Foreground(colour(look.ColourMuted)),
			partStatusRule:   bar.Foreground(colour(look.ColourQuiet)),
			partHintKey:      bar.Foreground(colour(look.ColourText)).Bold(true),
			partCross:        plain.Foreground(colour(look.ColourFaint)),
		},
		highlightGround: colour(look.ColourTint),
		highlightText:   colour(look.ColourOnAccent),
		fewColours:      fewColours,
		noColours:       noColours,
	}
}

// colours are the design's colours in one theme, as a terminal of a profile shows them.
type colours struct {
	shades  map[look.Colour]look.Shade
	profile colorprofile.Profile
}

func (palette colours) of(name look.Colour) color.Color {
	return colourIn(palette.shades[name], palette.profile)
}

// colourIn is a shade as a terminal of a profile shows it: whole, one of 256, one of sixteen, or none.
func colourIn(shade look.Shade, profile colorprofile.Profile) color.Color {
	switch profile {
	case colorprofile.TrueColor:
		return shade.TrueColour
	case colorprofile.ANSI256:
		return ansi.IndexedColor(shade.Indexed)
	case colorprofile.ANSI:
		if shade.Basic == look.BasicOwn {
			return lipgloss.NoColor{}
		}
		return ansi.BasicColor(uint8(shade.Basic))
	case colorprofile.ASCII, colorprofile.NoTTY, colorprofile.Unknown:
	}
	return lipgloss.NoColor{}
}

// tab paints a tab's lines in the frame: each at the page's left edge, cut at its right one, the screen's width
// filled.
func (paint paint) tab(lines []line, frame frame) []string {
	drawn := make([]string, 0, len(lines))
	gutter := strings.Repeat(space, frame.left)
	for _, shown := range lines {
		painted, width := paint.line(shown, frame.inner)
		drawn = append(drawn, gutter+painted+strings.Repeat(space, max(frame.width-frame.left-width, 0)))
	}
	return drawn
}

// line paints a line cut at a width, and says how many columns it took.
func (paint paint) line(shown line, width int) (string, int) {
	var builder strings.Builder
	used := 0
	for _, piece := range shown {
		text := ansi.Truncate(piece.text, width-used, "")
		if text == "" {
			continue
		}
		builder.WriteString(paint.style(piece).Render(text))
		used += textWidth(text)
	}
	return builder.String(), used
}

// style is the look of a run: its kind's, highlighted where it is a result, carrying its address where it links.
func (paint paint) style(piece run) lipgloss.Style {
	style := paint.kinds[piece.kind]
	if piece.highlighted {
		style = style.Background(paint.highlightGround).Underline(paint.noColours)
		if paint.fewColours {
			style = style.Foreground(paint.highlightText)
		}
	}
	if piece.address != "" {
		style = style.Hyperlink(piece.address)
	}
	return style
}
