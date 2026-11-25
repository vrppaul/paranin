package card

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/vrppaul/paranin/internal/look"
)

// TypstVersion is the version of Typst the cards are drawn with (ADR-0005): another can move what the template lays
// out, so a card is drawn by this one alone, and a bump of it is looked at on every card.
const TypstVersion = "0.15.1"

// The template and the tokens it reads, embedded, which a drawing lays out in a folder of its own for Typst to read
//
//go:embed card.typ tokens.typ
var templateFiles embed.FS

// How Typst is run: the template it compiles, the folder a drawing lays it out in, the input the facts are handed in
// as, which the template reads as sys.inputs.facts, and the words of its command: how it says its version, the word
// after its name, and how it draws, at 72 ppi so a point of the card is a pixel, with the repository's faces alone,
// into the PNG it writes to its standard output
const (
	templateFile     = "card.typ"
	drawingFolder    = "paranin-card-*"
	factsKey         = "facts"
	inputJoint       = "="
	versionCommand   = "--version"
	compileCommand   = "compile"
	rootOption       = "--root"
	fontsOption      = "--font-path"
	pixelsOption     = "--ppi"
	pixelsPerInch    = "72"
	formatOption     = "--format"
	pictureFormat    = "png"
	inputOption      = "--input"
	noSystemFonts    = "--ignore-system-fonts"
	noEmbeddedFonts  = "--ignore-embedded-fonts"
	standardOutput   = "-"
	versionWordIndex = 1
)

// The kinds of card as the template tells them apart (card.typ)
const (
	kindResume  = "resume"
	kindProject = "project"
	kindList    = "list"
	kindPost    = "post"
)

// What a project's name breaks after where it does not fit on one line
const nameBreak = "-"

// The faults of drawing a card
var errTypstVersion = errors.New("typst is not the version the cards are drawn with")

// Drawer draws cards with Typst: the program, at TypstVersion, and the folder of the faces the cards are set in, which
// alone Typst reads, so no face of the machine's reaches a card.
type Drawer struct {
	program string
	faces   string
}

// NewDrawer draws cards with a Typst program, set in the faces of a folder; it refuses a program at another version
// than TypstVersion.
func NewDrawer(ctx context.Context, program, faces string) (Drawer, error) {
	said, err := run(ctx, program, versionCommand)
	if err != nil {
		return Drawer{program: "", faces: ""}, fmt.Errorf("ask %s its version: %w", program, err)
	}
	if words := strings.Fields(string(said)); len(words) <= versionWordIndex || words[versionWordIndex] != TypstVersion {
		return Drawer{program: "", faces: ""}, fmt.Errorf("%w: %s says %q, want %s", errTypstVersion, program, strings.TrimSpace(string(said)), TypstVersion)
	}
	return Drawer{program: program, faces: faces}, nil
}

// Draw draws a card's facts as a PNG, 1200×630.
func (drawer Drawer) Draw(ctx context.Context, facts Facts) ([]byte, error) {
	input, err := json.Marshal(inputOf(facts))
	if err != nil {
		return nil, fmt.Errorf("write a card's facts: %w", err)
	}
	folder, err := os.MkdirTemp("", drawingFolder)
	if err != nil {
		return nil, fmt.Errorf("make a folder to draw a card in: %w", err)
	}
	defer func() { _ = os.RemoveAll(folder) }()
	if err := os.CopyFS(folder, templateFiles); err != nil {
		return nil, fmt.Errorf("lay the cards' template out in %s: %w", folder, err)
	}
	drawn, err := run(ctx, drawer.program, compileCommand, rootOption, folder, noSystemFonts, noEmbeddedFonts, fontsOption, drawer.faces,
		pixelsOption, pixelsPerInch, formatOption, pictureFormat, inputOption, factsKey+inputJoint+string(input), filepath.Join(folder, templateFile), standardOutput)
	if err != nil {
		return nil, fmt.Errorf("draw the card of %+v: %w", facts, err)
	}
	return drawn, nil
}

// run runs a program with arguments, and is what it writes to its standard output; a fault says what it wrote to its
// standard error.
func run(ctx context.Context, program string, arguments ...string) ([]byte, error) {
	var written, said bytes.Buffer
	command := exec.CommandContext(ctx, program, arguments...) //nolint:gosec // Typst, as its caller names it, run with no shell
	command.Stdout, command.Stderr = &written, &said
	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(said.String()))
	}
	return written.Bytes(), nil
}

// sideInput is a card of the résumé's side as the template reads it: its kind, its title in block letters, its line
// and the way in.
type sideInput struct {
	Kind  string  `json:"kind"`
	Title letters `json:"title"`
	Line  string  `json:"line"`
	WayIn string  `json:"way_in"`
}

// postInput is a post's card as the template reads it.
type postInput struct {
	Kind   string  `json:"kind"`
	Mark   letters `json:"mark"`
	Day    string  `json:"day"`
	Series string  `json:"series"`
	Title  string  `json:"title"`
	Lead   string  `json:"lead"`
	WayIn  string  `json:"way_in"`
}

// letters are words in the name's block letters (look.NameFont) as the template draws them: the words, which it sets as
// text where the font does not draw every letter of them and so gives no lines; the lines they stand on, each of the
// pieces it may break between, as a project's name breaks after a hyphen; the cells of the cursor after them, none
// where none follows; and the font's gaps, of columns between two pieces and of half rows between two lines.
type letters struct {
	Words           string       `json:"words"`
	Lines           []letterLine `json:"lines"`
	CursorCells     int          `json:"cursor_cells"`
	GapCells        int          `json:"gap_cells"`
	LineGapHalfRows int          `json:"line_gap_half_rows"`
}

// letterLine is a line of block letters: the pieces it is made of, side by side.
type letterLine struct {
	Pieces []letterPiece `json:"pieces"`
}

// letterPiece is a piece of block letters as its half rows, top to bottom, each a column's half filled or empty.
type letterPiece struct {
	HalfRows [][]bool `json:"half_rows"`
}

// inputOf is a card's facts as the template reads them. gochecksumtype holds the switch to every kind.
func inputOf(facts Facts) any {
	switch typed := facts.(type) {
	case Resume:
		return sideInput{Kind: kindResume, Title: nameLetters(typed.Name), Line: typed.Line, WayIn: typed.WayIn}
	case Project:
		pieces := slices.DeleteFunc(strings.SplitAfter(strings.ToUpper(typed.Name), nameBreak), func(piece string) bool { return piece == "" })
		return sideInput{Kind: kindProject, Title: lettersOf(typed.Name, [][]string{pieces}), Line: typed.Line, WayIn: typed.WayIn}
	case List:
		return sideInput{Kind: kindList, Title: lettersOf(typed.Title, [][]string{{strings.ToUpper(typed.Title)}}), Line: typed.Line, WayIn: typed.WayIn}
	case Post:
		return postInput{
			Kind: kindPost, Mark: lettersOf(typed.Blog, [][]string{{strings.ToUpper(typed.Blog)}}), Day: typed.Day, Series: typed.Series,
			Title: typed.Title, Lead: typed.Lead, WayIn: typed.WayIn,
		}
	}
	return nil
}

// lettersOf are words in the name's font, on lines, each of its pieces: no lines where the font does not draw every
// letter of them.
func lettersOf(words string, lines [][]string) letters {
	drawn := letters{
		Words: words, Lines: make([]letterLine, 0, len(lines)), CursorCells: 0,
		GapCells: look.GlyphGap, LineGapHalfRows: look.HalvesInRow * look.BannerLineGap,
	}
	font := look.NameFont()
	for _, line := range lines {
		pieces := make([]letterPiece, 0, len(line))
		for _, piece := range line {
			banner := font.Banner([]string{piece})
			if len(banner) == 0 {
				drawn.Lines = []letterLine{}
				return drawn
			}
			pieces = append(pieces, letterPiece{HalfRows: look.HalfRows(banner)})
		}
		drawn.Lines = append(drawn.Lines, letterLine{Pieces: pieces})
	}
	return drawn
}

// nameLetters are the author's name in block letters on the lines the profile breaks it into, each whole, the cursor
// after it as wide as its last letter, as the page draws the name; no cursor where it is set as text.
func nameLetters(name string) letters {
	lines := look.NameLines(name)
	pieces := make([][]string, 0, len(lines))
	for _, line := range lines {
		pieces = append(pieces, []string{line})
	}
	drawn := lettersOf(name, pieces)
	if len(drawn.Lines) > 0 {
		last, _ := utf8.DecodeLastRuneInString(lines[len(lines)-1])
		drawn.CursorCells = look.NameFont().LetterWidth(last)
	}
	return drawn
}
