package tui

import (
	"image/color"
	"slices"
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
	// partHost is the session the page shown belongs to, of those the footer opens with, the site's by its host or
	// the blog's; partSession another session; partTabNumber a tab's number, and the mark of tabs given way;
	// partTabName its name; partTabShown the tab shown, reversed
	partHost
	partSession
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
	// the tile the keys stand on stands on the panel's ground, its title underlined in the accent, or in the
	// terminal's own colour where it shows few colours or none, whose panel is the terminal's own ground
	pointedGround    color.Color
	pointedUnderline color.Color
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
	accentMark := accentWhereColourful(colour(look.ColourAccent), fewColours)
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
			kindBar:      plain.Foreground(colour(look.ColourTimeBar)).Bold(true),
			kindEmphasis: plain.Foreground(colour(look.ColourText)).Italic(true),
			// the way chosen reversed and bold, as the tab shown is
			kindChosen:    plain.Foreground(colour(look.ColourText)).Bold(true).Reverse(true),
			kindTileTitle: plain.Foreground(colour(look.ColourText)).Bold(true),
			kindQuietLine: plain.Foreground(colour(look.ColourQuiet)),
			kindOff:       plain.Foreground(colour(look.ColourQuietHalf)),
			kindQuoteBar:  plain.Foreground(colour(look.ColourEdge)),
			kindLookMark:  plain.Foreground(colour(look.ColourAccent)),
			// the dot that says a widget runs in the colour of a thing passed, the word beside it telling it with none
			kindLiveMark: plain.Foreground(colour(look.ColourPassed)),
			// each kind of word in code in its own hue (DESIGN.md, Article)
			kindCodeKeyword:  plain.Foreground(colour(look.ColourCodeKeyword)),
			kindCodeString:   plain.Foreground(colour(look.ColourCodeString)),
			kindCodeNumber:   plain.Foreground(colour(look.ColourCodeNumber)),
			kindCodeFunction: plain.Foreground(colour(look.ColourCodeFunction)),
			kindCodeType:     plain.Foreground(colour(look.ColourCodeType)),
			kindCodeBuiltin:  plain.Foreground(colour(look.ColourCodeBuiltin)),
			kindCodeOperator: plain.Foreground(colour(look.ColourCodeOperator)),
			kindCodeVariable: plain.Foreground(colour(look.ColourCodeVariable)),
		},
		gradient: gradientOf(colour(look.ColourAccent), colour(look.ColourHost), fewColours),
		swatches: swatchesOf(colour),
		parts: map[part]lipgloss.Style{
			partCross:       plain.Foreground(colour(look.ColourFaint)),
			partTrack:       plain.Foreground(colour(look.ColourQuiet)),
			partThumb:       plain.Foreground(colour(look.ColourFaint)),
			partPanel:       panel,
			partHost:        panel.Foreground(colour(look.ColourHost)).Bold(true),
			partSession:     panel.Foreground(colour(look.ColourMuted)),
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
			// the letters typed bold and underlined in the accent, as the page marks them, the accent reading too faint as
			// small text in the light theme
			partMenuTyped: menu.Foreground(colour(look.ColourText)).Bold(true).Underline(true).UnderlineColor(accentMark),
			partMenuKind:  menu.Foreground(colour(look.ColourMuted)),
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
		fadedLess:        colour(look.ColourFaint),
		fadedMost:        colour(look.ColourEdge),
		highlightGround:  colour(look.ColourTint),
		highlightText:    colour(look.ColourOnAccent),
		pointedGround:    colour(look.ColourSunk),
		pointedUnderline: accentMark,
		matchGround:      colour(look.ColourAccent),
		matchText:        colour(look.ColourOnAccent),
		matchUnderline:   accentMark,
		cursorColour:     accentMark,
		fewColours:       fewColours,
		noColours:        noColours,
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
		kindMailbox, kindHost, kindCurrent, kindPast, kindOwner, kindRepository, kindPrompt, kindBar, kindEmphasis,
		kindChosen, kindTileTitle, kindQuietLine, kindOff, kindQuoteBar, kindLookMark, kindLiveMark, kindCodeKeyword, kindCodeString, kindCodeNumber,
		kindCodeFunction, kindCodeType, kindCodeBuiltin, kindCodeOperator, kindCodeVariable:
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

// lineMarks are what sets runs of a line apart beside their kinds: which match of the line the search's count names,
// noMatch where none; the columns the tile the keys stand on takes on it, none where the line crosses no such tile;
// and the column its part right of what stays in sight over its left part starts at, the groups' names or a post's
// Outline, 0 on a tab with neither.
type lineMarks struct {
	named   int
	pointed columnSpan
	split   int
}

// columnSpan is the columns of a line from a first to the one past its last; the zero span holds none.
type columnSpan struct {
	start int
	end   int
}

// holds is whether a column lies in the span.
func (span columnSpan) holds(column int) bool {
	return column >= span.start && column < span.end
}

// pageMarks are what sets runs of a tab's lines apart beside their kinds: the match the search's count names; the
// tile the keys stand on, the zero tile where the tab holds none; the column the lines' parts right of what stays in
// sight over their left part start at; and the lines it may stand over, the only ones parted there.
type pageMarks struct {
	named     namedMatch
	pointed   laidTile
	split     int
	overLines []lineSpan
}

// lineSpan is a tab's lines from a first to the one past its last; the zero span holds none.
type lineSpan struct {
	first int
	end   int
}

// holds is whether a line lies in the span.
func (span lineSpan) holds(line int) bool {
	return line >= span.first && line < span.end
}

// on are the marks of one line of the tab: parted at the split where what stays in sight may stand over it.
func (marks pageMarks) on(line int) lineMarks {
	split := 0
	if slices.ContainsFunc(marks.overLines, func(span lineSpan) bool { return span.holds(line) }) {
		split = marks.split
	}
	return lineMarks{named: marks.named.occurrenceOn(line), pointed: marks.pointed.groundOn(line), split: split}
}

// tab paints a tab's lines in the frame, as the tab's marks set their runs apart.
func (paint paint) tab(lines []markedLine, frame frame, marks pageMarks) []paintedLine {
	drawn := make([]paintedLine, 0, len(lines))
	for place, shown := range lines {
		drawn = append(drawn, paint.pageLine(shown, frame, nil, marks.on(place)))
	}
	return drawn
}

// pageLine paints a line of a tab from the screen's left edge, cut at the page's right one, its text in a faded colour
// where one is given, its runs set apart as its marks say: on a list with a names' column or a post with its
// Outline, its parts before and after the column its marks split it at apart, which what stays in sight over its
// left part stands beside; any other line whole.
func (paint paint) pageLine(shown markedLine, frame frame, faded color.Color, marks lineMarks) paintedLine {
	margin := strings.Repeat(space, frame.left)
	if marks.split == 0 {
		painted, width := paint.line(shown.runs, 0, frame.inner, faded, marks)
		return paintedLine{left: paintedBar{text: margin + painted, width: frame.left + width}, right: paintedBar{text: "", width: 0}}
	}
	left, right := partedAt(shown.runs, marks.split)
	before, beforeWidth := paint.line(left, 0, frame.inner, faded, marks)
	after, afterWidth := paint.line(right, marks.split, frame.inner, faded, marks)
	if afterWidth > 0 {
		before += strings.Repeat(space, marks.split-beforeWidth)
		beforeWidth = marks.split
	}
	return paintedLine{left: paintedBar{text: margin + before, width: frame.left + beforeWidth}, right: paintedBar{text: after, width: afterWidth}}
}

// partedAt are runs parted at a column: those before it, and those from it on, a run that spans it cut in two.
func partedAt(runs []matchedRun, column int) ([]matchedRun, []matchedRun) {
	used := 0
	for index, piece := range runs {
		width := textWidth(piece.text)
		if used+width <= column {
			used += width
			continue
		}
		head, tail := piece, piece
		head.text, tail.text = ansi.Cut(piece.text, 0, column-used), ansi.Cut(piece.text, column-used, width)
		return append(slices.Clone(runs[:index]), head), append([]matchedRun{tail}, runs[index+1:]...)
	}
	return runs, nil
}

// line paints runs from a column of the page, cut at a width, its text in a faded colour where one is given, its runs
// set apart as its marks say, and says how many columns it took.
func (paint paint) line(shown []matchedRun, from, width int, faded color.Color, marks lineMarks) (string, int) {
	var builder strings.Builder
	used := from
	for _, piece := range shown {
		text := ansi.Truncate(piece.text, width-used, "")
		if text == "" {
			continue
		}
		named := piece.occurrence != noMatch && piece.occurrence == marks.named
		builder.WriteString(paint.style(piece, faded, named, marks.pointed.holds(used)).Render(text))
		used += textWidth(text)
	}
	return builder.String(), used - from
}

// style is the look of a run: its kind's, its text in a faded colour where one is given, highlighted where it is a
// result, on the panel's ground where it lies in the tile the keys stand on, its title then underlined, marked where
// it matches the search, apart where it is the match the count names, carrying its address where it links.
func (paint paint) style(piece matchedRun, faded color.Color, named, pointed bool) lipgloss.Style {
	style := paint.kinds[piece.kind]
	if tones := paint.tonesOf(piece.kind); len(tones) > 0 {
		style = style.Foreground(tones[piece.tone%len(tones)])
	}
	if faded != nil {
		style = style.Foreground(faded)
	}
	if pointed {
		style = style.Background(paint.pointedGround)
	}
	if pointed && piece.kind == kindTileTitle {
		style = style.Underline(true).UnderlineColor(paint.pointedUnderline)
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
