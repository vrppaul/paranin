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
	// partCross is a corner cross around the page
	partCross part = iota
	// partTrack and partThumb are the scrollbar's track and its thumb
	partTrack
	partThumb
	// partRim is the rim that parts the footer from the page; partMoreMark, partMore and partMoreKey are its word of
	// the lines below: the mark, what it says, and the key that scrolls
	partRim
	partMoreMark
	partMore
	partMoreKey
	// partPanel is the ground of the panel that holds the prompt and the places
	partPanel
	// partHost, partHome and partPromptMark open the prompt: the site's host, the place it stands in, and its mark
	partHost
	partHome
	partPromptMark
	// partTyped is what the reader types on the prompt, partSaid what it says after a key, partFailure what it says
	// went wrong, and partSiteCursor the site's cursor, where nothing is typed or said
	partTyped
	partSaid
	partFailure
	partSiteCursor
	// partPanelKey and partPanelNote are a key the panel names and what it says the key does
	partPanelKey
	partPanelNote
	// partPlace is a place listed; partPlaceTyped and partPlaceRest the start of one the reader typed and the rest of
	// it; partPlaceChosen the place Tab chose
	partPlace
	partPlaceTyped
	partPlaceRest
	partPlaceChosen
	// partStatus is the ground of the status line, partTitle the tab's title that opens it, as Neovim's opens with the
	// mode, partTally the tab's count of checks, and partRuler the ruler
	partStatus
	partTitle
	partTally
	partRuler
	// partEdge is the border of a window over the page, and partWindowRule a rule inside it; partWindowName and
	// partWindowNote are a row it holds, its keys and what they do
	partEdge
	partWindowRule
	partWindowName
	partWindowNote
)

// paint is how each kind of text and each part of the frame looks in one terminal: its theme, and the colours it
// shows. With sixteen colours or none, what colour alone would tell is told by reversing or underlining too.
type paint struct {
	kinds map[kind]lipgloss.Style
	parts map[part]lipgloss.Style
	// gradient is the colours from the accent to the host's the name and a tab's title go through, and swatches the
	// design's colours the swatches go round: each laid over its kind's look
	gradient []color.Color
	swatches []color.Color
	// the text of the two rows of the page above the rim, where more of the tab lies below, takes these colours: the
	// one above the less faded, the last the most
	fadedLess color.Color
	fadedMost color.Color
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
	panel := plain.Background(colour(look.ColourSunk))
	deep := plain.Background(colour(look.ColourDeep))
	rim := plain.Foreground(colour(look.ColourSunk))
	if fewColours {
		rim = plain.Foreground(colour(look.ColourFaint))
	}
	return paint{
		kinds: map[kind]lipgloss.Style{
			kindText:    plain.Foreground(colour(look.ColourText)),
			kindMuted:   plain.Foreground(colour(look.ColourMuted)),
			kindFaint:   plain.Foreground(colour(look.ColourFaint)).Faint(noColours),
			kindStrong:  plain.Foreground(colour(look.ColourText)).Bold(true),
			kindCode:    plain.Foreground(colour(look.ColourAccent)),
			kindCursor:  plain.Foreground(colour(look.ColourAccent)),
			kindPassed:  plain.Foreground(colour(look.ColourPassed)),
			kindWarned:  plain.Foreground(colour(look.ColourWarned)),
			kindFailed:  plain.Foreground(colour(look.ColourError)),
			kindName:    plain.Bold(true),
			kindHeading: plain.Bold(true),
			kindLabel:   plain.Foreground(colour(look.ColourLabel)).Bold(true),
			kindMailbox: plain.Foreground(colour(look.ColourAccent)),
			kindHost:    plain.Foreground(colour(look.ColourHost)).Bold(true),
			kindCurrent: plain.Foreground(colour(look.ColourAccent)),
			kindPast:    plain.Foreground(colour(look.ColourEdge)).Faint(noColours),
			kindSwatch:  plain,
		},
		gradient: gradientOf(colour(look.ColourAccent), colour(look.ColourHost), fewColours),
		swatches: swatchesOf(colour),
		parts: map[part]lipgloss.Style{
			partCross:       plain.Foreground(colour(look.ColourFaint)),
			partTrack:       plain.Foreground(colour(look.ColourQuiet)),
			partThumb:       plain.Foreground(colour(look.ColourFaint)),
			partRim:         rim,
			partMoreMark:    panel.Foreground(colour(look.ColourAccent)).Bold(true),
			partMore:        panel.Foreground(colour(look.ColourMuted)),
			partMoreKey:     panel.Foreground(colour(look.ColourText)).Bold(true),
			partPanel:       panel,
			partHost:        panel.Foreground(colour(look.ColourHost)).Bold(true),
			partHome:        panel.Foreground(colour(look.ColourLabel)).Bold(true),
			partPromptMark:  panel.Foreground(colour(look.ColourPrompt)).Bold(true),
			partTyped:       panel.Foreground(colour(look.ColourText)),
			partSaid:        panel.Foreground(colour(look.ColourText)),
			partFailure:     panel.Foreground(colour(look.ColourError)).Bold(true),
			partSiteCursor:  panel.Foreground(colour(look.ColourAccent)),
			partPanelKey:    panel.Foreground(colour(look.ColourText)).Bold(true),
			partPanelNote:   panel.Foreground(colour(look.ColourFaint)),
			partPlace:       panel.Foreground(colour(look.ColourMuted)),
			partPlaceTyped:  panel.Foreground(colour(look.ColourAccent)).Bold(true),
			partPlaceRest:   panel.Foreground(colour(look.ColourText)).Bold(true),
			partPlaceChosen: plain.Background(colour(look.ColourChosen)).Foreground(colour(look.ColourText)).Bold(true).Reverse(fewColours),
			partStatus:      deep,
			partTitle:       plain.Foreground(colour(look.ColourOnAccent)).Background(colour(look.ColourAccent)).Bold(true).Reverse(noColours),
			partTally:       deep.Foreground(colour(look.ColourFaint)),
			partRuler:       deep.Foreground(colour(look.ColourFaint)),
			partEdge:        plain.Foreground(colour(look.ColourEdge)),
			partWindowRule:  panel.Foreground(colour(look.ColourQuiet)),
			partWindowName:  panel.Foreground(colour(look.ColourText)).Bold(true),
			partWindowNote:  panel.Foreground(colour(look.ColourMuted)),
		},
		fadedLess:       colour(look.ColourFaint),
		fadedMost:       colour(look.ColourEdge),
		highlightGround: colour(look.ColourTint),
		highlightText:   colour(look.ColourOnAccent),
		matchGround:     colour(look.ColourAccent),
		matchText:       colour(look.ColourOnAccent),
		cursorColour:    cursorColourOf(colour(look.ColourAccent), fewColours),
		fewColours:      fewColours,
		noColours:       noColours,
	}
}

// How many tones the gradient goes through from the accent to the host's colour, along the author's name and a tab's
// title
const gradientTones = 16

// swatchColours are the design's colours the brief's swatches show, one swatch each, in the order they stand; the
// layout lays out as many swatches as there are.
func swatchColours() []look.Colour {
	return []look.Colour{look.ColourError, look.ColourAccent, look.ColourWarned, look.ColourPassed, look.ColourLabel, look.ColourHost}
}

// swatchesOf are the swatches' colours as a terminal shows them.
func swatchesOf(colour func(look.Colour) color.Color) []color.Color {
	shown := []color.Color{}
	for _, name := range swatchColours() {
		shown = append(shown, colour(name))
	}
	return shown
}

// gradientOf are the gradient's colours: those between the accent and the host's, blended; where the terminal shows
// few colours, the accent alone, every tone going round to it, which blending would only round to a few of its
// sixteen.
func gradientOf(accent, host color.Color, fewColours bool) []color.Color {
	if fewColours {
		return []color.Color{accent}
	}
	return lipgloss.Blend1D(gradientTones, accent, host)
}

// tonesOf are the colours a kind painted in several takes, each run the one its tone names, round them where its tone
// runs past them; none for a kind of one colour.
func (paint paint) tonesOf(shown kind) []color.Color {
	switch shown {
	case kindName, kindHeading:
		return paint.gradient
	case kindSwatch:
		return paint.swatches
	case kindText, kindMuted, kindFaint, kindStrong, kindCode, kindCursor, kindPassed, kindWarned, kindFailed, kindLabel,
		kindMailbox, kindHost, kindCurrent, kindPast:
	}
	return nil
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
// right one.
func (paint paint) tab(lines []markedLine, frame frame) []paintedLine {
	drawn := make([]paintedLine, 0, len(lines))
	gutter := strings.Repeat(space, frame.left)
	for _, shown := range lines {
		painted, width := paint.line(shown.runs, frame.inner, nil)
		drawn = append(drawn, paintedLine{text: gutter + painted, width: frame.left + width})
	}
	return drawn
}

// faded paints a line of the page as a row above the rim fades, its text in a faded colour, its highlight, its marks
// and its link kept.
func (paint paint) faded(shown markedLine, frame frame, colour color.Color) string {
	painted, _ := paint.line(shown.runs, frame.inner, colour)
	return strings.Repeat(space, frame.left) + painted
}

// line paints a line cut at a width, its text in a faded colour where one is given, and says how many columns it
// took.
func (paint paint) line(shown []matchedRun, width int, faded color.Color) (string, int) {
	var builder strings.Builder
	used := 0
	for _, piece := range shown {
		text := ansi.Truncate(piece.text, width-used, "")
		if text == "" {
			continue
		}
		builder.WriteString(paint.style(piece, faded).Render(text))
		used += textWidth(text)
	}
	return builder.String(), used
}

// style is the look of a run: its kind's, its text in a faded colour where one is given, highlighted where it is a
// result, marked where it matches the search, carrying its address where it links.
func (paint paint) style(piece matchedRun, faded color.Color) lipgloss.Style {
	style := paint.kinds[piece.kind]
	if tones := paint.tonesOf(piece.kind); len(tones) > 0 {
		style = style.Foreground(tones[piece.tone%len(tones)])
	}
	if faded != nil {
		style = style.Foreground(faded)
	}
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
