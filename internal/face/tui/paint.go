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
	// partCommand is what the command line says, partFailure what it says went wrong, and partPrompt the slash or
	// the colon the reader types after
	partCommand
	partFailure
	partPrompt
	// partEdge is the border of a window over the page, and partWindowRule a rule inside it; partWindowName and
	// partWindowNote are an item it lists, its name and what it says; partChosenMark, partChosenName and
	// partChosenNote are the item chosen, its mark among them
	partEdge
	partWindowRule
	partWindowName
	partWindowNote
	partChosenMark
	partChosenName
	partChosenNote
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
	// a match of the search stands on the accent, its text laid on it, and with few colours or none reversed
	matchGround color.Color
	matchText   color.Color
	// the terminal's cursor is the accent, or the terminal's own colour where it shows few colours or none
	cursorColour color.Color
	fewColours   bool
	noColours    bool
}

// paintFor is the paint of a theme in a terminal that shows a profile of colours.
func paintFor(theme look.Theme, profile colorprofile.Profile) paint {
	colour := colours{shades: look.Palette(theme), profile: profile}.of
	// colorprofile orders its profiles from none to full colour
	fewColours, noColours := profile < colorprofile.ANSI256, profile < colorprofile.ANSI
	plain := lipgloss.NewStyle()
	bar := plain.Background(colour(look.ColourSunk))
	chosen := plain.Background(colour(look.ColourChosen))
	return paint{
		kinds: map[kind]lipgloss.Style{
			kindText:   plain.Foreground(colour(look.ColourText)),
			kindMuted:  plain.Foreground(colour(look.ColourMuted)),
			kindFaint:  plain.Foreground(colour(look.ColourFaint)).Faint(noColours),
			kindStrong: plain.Foreground(colour(look.ColourText)).Bold(true),
			kindCode:   plain.Foreground(colour(look.ColourAccent)),
			kindCursor: plain.Foreground(colour(look.ColourAccent)),
			kindPassed: plain.Foreground(colour(look.ColourPassed)),
			kindWarned: plain.Foreground(colour(look.ColourWarned)),
			kindFailed: plain.Foreground(colour(look.ColourError)),
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
			partCommand:      plain.Foreground(colour(look.ColourText)),
			partFailure:      plain.Foreground(colour(look.ColourError)).Bold(true),
			partPrompt:       plain.Foreground(colour(look.ColourAccent)).Bold(true),
			partEdge:         plain.Foreground(colour(look.ColourEdge)),
			partWindowRule:   bar.Foreground(colour(look.ColourQuiet)),
			partWindowName:   bar.Foreground(colour(look.ColourText)).Bold(true),
			partWindowNote:   bar.Foreground(colour(look.ColourMuted)),
			partChosenMark:   chosen.Foreground(colour(look.ColourAccent)),
			partChosenName:   chosen.Foreground(colour(look.ColourText)).Bold(true),
			partChosenNote:   chosen.Foreground(colour(look.ColourText)),
		},
		highlightGround: colour(look.ColourTint),
		highlightText:   colour(look.ColourOnAccent),
		matchGround:     colour(look.ColourAccent),
		matchText:       colour(look.ColourOnAccent),
		cursorColour:    cursorColourOf(colour(look.ColourAccent), fewColours),
		fewColours:      fewColours,
		noColours:       noColours,
	}
}

// cursorColourOf is the colour the terminal's cursor takes: the accent, or none where the terminal shows few
// colours. A terminal takes the cursor's colour as red, green and blue: with sixteen colours those would be the
// standard ones, not the reader's theme's, and with none lipgloss's absent colour would read as black.
func cursorColourOf(accent color.Color, fewColours bool) color.Color {
	if fewColours {
		return nil
	}
	return accent
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

// tab paints a tab's lines in the frame, each match of the search marked: each at the page's left edge, cut at its
// right one, the screen's width filled.
func (paint paint) tab(lines []markedLine, frame frame) []string {
	drawn := make([]string, 0, len(lines))
	gutter := strings.Repeat(space, frame.left)
	for _, shown := range lines {
		painted, width := paint.line(shown.runs, frame.inner)
		drawn = append(drawn, gutter+painted+strings.Repeat(space, max(frame.width-frame.left-width, 0)))
	}
	return drawn
}

// line paints a line cut at a width, and says how many columns it took.
func (paint paint) line(shown []matchedRun, width int) (string, int) {
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

// style is the look of a run: its kind's, highlighted where it is a result, marked where it matches the search,
// carrying its address where it links.
func (paint paint) style(piece matchedRun) lipgloss.Style {
	style := paint.kinds[piece.kind]
	if piece.highlighted {
		style = style.Background(paint.highlightGround).Underline(paint.noColours)
		if paint.fewColours {
			style = style.Foreground(paint.highlightText)
		}
	}
	// a match stands out over a highlight too, and with few colours or none it is reversed, as a highlight is not
	if piece.matched {
		style = style.Background(paint.matchGround).Foreground(paint.matchText).Reverse(paint.fewColours)
	}
	if piece.address != "" {
		style = style.Hyperlink(piece.address)
	}
	return style
}
