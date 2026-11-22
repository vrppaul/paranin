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
	// partPanel is the ground of the panel, the footer's and the command line's
	partPanel
	// partHost is the session the footer opens with, named by the site's host; partTabNumber a tab's number, and the
	// mark of tabs given way; partTabName its name; partTabShown the tab shown, reversed
	partHost
	partTabNumber
	partTabName
	partTabShown
	// partRuler is the ruler, at the footer's right
	partRuler
	// partCommandMark is the mark the command line opens with; partTyped what the reader types on it, partSaid what
	// it says after a key, partFailure what it says went wrong; partCount the count of the matches of a search typed
	partCommandMark
	partTyped
	partSaid
	partFailure
	partCount
	// partPanelKey and partPanelNote are a key the panel names and what it says the key does
	partPanelKey
	partPanelNote
	// partMenu is an item of the menu over the command line, on the menu's ground; partMenuTyped the letters of its
	// name the reader typed; partMenuKind its kind; partMenuChosen the item Tab chose; partMenuTrack and partMenuThumb
	// the menu's scrollbar
	partMenu
	partMenuTyped
	partMenuKind
	partMenuChosen
	partMenuTrack
	partMenuThumb
	// partEdge is the border of a window over the page and the rule inside it, in one colour as tmux draws a border;
	// partWindowName and partWindowNote are a row it holds, its keys and what they do
	partEdge
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
	// the text of the page's last two rows, where more of the tab lies below, takes these colours: the one above the
	// less faded, the last the most
	fadedLess color.Color
	fadedMost color.Color
	// highlight is a result worth seeing first: the tint behind it, and with few colours the text on it darkened,
	// with none underlined
	highlightGround color.Color
	highlightText   color.Color
	// the match of the search the count names stands on the accent, its text laid on it, and with few colours or none
	// reversed; the others are underlined in the accent, or in the terminal's own colour where it shows few or none
	matchGround    color.Color
	matchText      color.Color
	matchUnderline color.Color
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
	// with no colours the menu has no ground to part it from the page, so it is reversed
	menu := plain.Background(colour(look.ColourQuiet)).Reverse(noColours)
	return paint{
		kinds: map[kind]lipgloss.Style{
			kindText:   plain.Foreground(colour(look.ColourText)),
			kindMuted:  plain.Foreground(colour(look.ColourMuted)),
			kindFaint:  plain.Foreground(colour(look.ColourFaint)).Faint(noColours),
			kindStrong: plain.Foreground(colour(look.ColourText)).Bold(true),
			// code in the text's colour on the panel's ground, as the page sets it, the accent reading too faint as small
			// text in the light theme; reversed where the terminal shows sixteen colours or none, whose panel is the
			// terminal's own ground
			kindCode:    panel.Foreground(colour(look.ColourText)).Reverse(fewColours),
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
			// a repository's owner and name in the colours of the mail's two parts
			kindOwner:      plain.Foreground(colour(look.ColourAccent)).Bold(true),
			kindRepository: plain.Foreground(colour(look.ColourHost)).Bold(true),
			kindPrompt:     plain.Foreground(colour(look.ColourPrompt)).Bold(true),
			// a bar's cells and the mark of no measured time differ in shape, so they need no colour to tell them apart
			kindBar: plain.Foreground(colour(look.ColourTimeBar)).Bold(true),
		},
		gradient: gradientOf(colour(look.ColourAccent), colour(look.ColourHost), fewColours),
		swatches: swatchesOf(colour),
		parts: map[part]lipgloss.Style{
			partCross:       plain.Foreground(colour(look.ColourFaint)),
			partTrack:       plain.Foreground(colour(look.ColourQuiet)),
			partThumb:       plain.Foreground(colour(look.ColourFaint)),
			partPanel:       panel,
			partHost:        panel.Foreground(colour(look.ColourHost)).Bold(true),
			partTabNumber:   panel.Foreground(colour(look.ColourFaint)),
			partTabName:     panel.Foreground(colour(look.ColourText)),
			partTabShown:    plain.Foreground(colour(look.ColourText)).Bold(true).Reverse(true),
			partRuler:       panel.Foreground(colour(look.ColourFaint)),
			partCommandMark: panel.Foreground(colour(look.ColourPrompt)).Bold(true),
			partTyped:       panel.Foreground(colour(look.ColourText)),
			partSaid:        panel.Foreground(colour(look.ColourText)),
			partFailure:     panel.Foreground(colour(look.ColourError)).Bold(true),
			partCount:       panel.Foreground(colour(look.ColourMuted)),
			partPanelKey:    panel.Foreground(colour(look.ColourText)).Bold(true),
			partPanelNote:   panel.Foreground(colour(look.ColourFaint)),
			partMenu:        menu.Foreground(colour(look.ColourText)),
			partMenuTyped:   menu.Foreground(colour(look.ColourAccent)).Bold(true),
			partMenuKind:    menu.Foreground(colour(look.ColourMuted)),
			// the item chosen reversed, as DESIGN.md marks what is chosen, and as the tab shown is; where the terminal shows
			// no colours, whose menu is reversed to stand out from the page, it stands upright and underlined instead
			partMenuChosen: plain.Foreground(colour(look.ColourText)).Bold(true).Reverse(!noColours).Underline(noColours),
			// the thumb a shade brighter than the track: with sixteen colours the track, the ground's colour, vanishes,
			// and the thumb stays in sight
			partMenuTrack:  menu.Foreground(colour(look.ColourFaint)),
			partMenuThumb:  menu.Foreground(colour(look.ColourMuted)),
			partEdge:       plain.Foreground(colour(look.ColourEdge)),
			partWindowName: panel.Foreground(colour(look.ColourText)).Bold(true),
			partWindowNote: panel.Foreground(colour(look.ColourMuted)),
		},
		fadedLess:       colour(look.ColourFaint),
		fadedMost:       colour(look.ColourEdge),
		highlightGround: colour(look.ColourTint),
		highlightText:   colour(look.ColourOnAccent),
		matchGround:     colour(look.ColourAccent),
		matchText:       colour(look.ColourOnAccent),
		matchUnderline:  accentWhereColourful(colour(look.ColourAccent), fewColours),
		cursorColour:    accentWhereColourful(colour(look.ColourAccent), fewColours),
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
		kindMailbox, kindHost, kindCurrent, kindPast, kindOwner, kindRepository, kindPrompt, kindBar:
	}
	return nil
}

// accentWhereColourful is the accent, or none, the terminal's own colour, where the terminal shows few colours: the
// colour of its cursor and of an underline. A terminal takes the cursor's colour as red, green and blue, with sixteen
// colours the standard ones, not the reader's theme's, and with none lipgloss's absent colour would read as black;
// and a coloured underline is one more thing a terminal of few colours may not draw.
func accentWhereColourful(accent color.Color, fewColours bool) color.Color {
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

// tab paints a tab's lines in the frame, each match of the search marked, the one its count names apart.
func (paint paint) tab(lines []markedLine, frame frame, named namedMatch) []paintedLine {
	drawn := make([]paintedLine, 0, len(lines))
	for place, shown := range lines {
		drawn = append(drawn, paint.pageLine(shown, frame, named.occurrenceOn(place)))
	}
	return drawn
}

// pageLine paints a line of a tab at the page's left edge, cut at its right one, its matches marked, the one of them
// the count names apart, where the count names one of this line.
func (paint paint) pageLine(shown markedLine, frame frame, named int) paintedLine {
	painted, width := paint.line(shown.runs, frame.inner, nil, named)
	return paintedLine{text: strings.Repeat(space, frame.left) + painted, width: frame.left + width}
}

// faded paints a line of the page as a row at its foot fades, its text in a faded colour, its highlight, its marks
// and its link kept.
func (paint paint) faded(shown markedLine, frame frame, colour color.Color, named int) string {
	painted, _ := paint.line(shown.runs, frame.inner, colour, named)
	return strings.Repeat(space, frame.left) + painted
}

// line paints a line cut at a width, its text in a faded colour where one is given, its match the count names
// apart, and says how many columns it took.
func (paint paint) line(shown []matchedRun, width int, faded color.Color, named int) (string, int) {
	var builder strings.Builder
	used := 0
	for _, piece := range shown {
		text := ansi.Truncate(piece.text, width-used, "")
		if text == "" {
			continue
		}
		builder.WriteString(paint.style(piece, faded, piece.occurrence != noMatch && piece.occurrence == named).Render(text))
		used += textWidth(text)
	}
	return builder.String(), used
}

// style is the look of a run: its kind's, its text in a faded colour where one is given, highlighted where it is a
// result, marked where it matches the search, apart where it is the match the count names, carrying its address
// where it links.
func (paint paint) style(piece matchedRun, faded color.Color, named bool) lipgloss.Style {
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
	// the match the count names stands out over a highlight too, and with few colours or none it is reversed, as a
	// highlight is not; the others are underlined, and with no colours bold as well, since a highlight is underlined
	// there
	switch {
	case named:
		style = style.Background(paint.matchGround).Foreground(paint.matchText).Reverse(paint.fewColours)
	case piece.occurrence != noMatch:
		style = style.Underline(true).UnderlineColor(paint.matchUnderline).Bold(paint.noColours)
	}
	if piece.address != "" {
		style = style.Hyperlink(piece.address)
	}
	return style
}
