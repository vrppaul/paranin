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

	"github.com/vrppaul/paranin/internal/look"
)

// TypstVersion is the version of Typst the cards are drawn with (ADR-0005): another can move what the template lays
// out, so a card is drawn by this one alone, and a bump of it is looked at on every card.
const TypstVersion = "0.15.1"

// The template and the tokens it reads, embedded, which a drawing lays out in a folder of its own for Typst to read
//
//go:embed card.typ tokens.typ
var templateFiles embed.FS

// How Typst is run: the template it reads, the folder a run lays it out in, the input the facts are handed in as,
// which the template reads as sys.inputs.facts, and the words of its command: how it says its version, the word after
// its name, and what every run of the template names, its root and the repository's faces alone
const (
	templateFile     = "card.typ"
	drawingFolder    = "paranin-card-*"
	factsKey         = "facts"
	inputJoint       = "="
	versionCommand   = "--version"
	versionWordIndex = 1
	rootOption       = "--root"
	fontsOption      = "--font-path"
	inputOption      = "--input"
	noSystemFonts    = "--ignore-system-fonts"
	noEmbeddedFonts  = "--ignore-embedded-fonts"
)

// How Typst draws a card: at 72 ppi, so a point of the card is a pixel, into the PNG it writes to its standard output
const (
	compileCommand = "compile"
	pixelsOption   = "--ppi"
	pixelsPerInch  = "72"
	formatOption   = "--format"
	pictureFormat  = "png"
	standardOutput = "-"
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
var (
	errTypstVersion = errors.New("typst is not the version the cards are drawn with")
	errUndrawnMark  = errors.New("the block letters cannot draw the cards' mark")
)

// Drawer draws cards with Typst: the program, at TypstVersion, and the folder of the faces the cards are set in, which
// alone Typst reads, so no face of the machine's reaches a card.
type Drawer struct {
	program string
	faces   string
}

// NewDrawer draws cards with a Typst program, set in the faces of a folder; it refuses a program at another version
// than TypstVersion.
func NewDrawer(ctx context.Context, program, faces string) (Drawer, error) {
	noDrawer := Drawer{program: "", faces: ""}
	said, err := run(ctx, program, versionCommand)
	if err != nil {
		return noDrawer, fmt.Errorf("ask %s its version: %w", program, err)
	}
	if words := strings.Fields(string(said)); len(words) <= versionWordIndex || words[versionWordIndex] != TypstVersion {
		return noDrawer, fmt.Errorf("%w: %s says %q, want %s", errTypstVersion, program, strings.TrimSpace(string(said)), TypstVersion)
	}
	return Drawer{program: program, faces: faces}, nil
}

// Draw draws a card's facts as a PNG, 1200×630.
func (drawer Drawer) Draw(ctx context.Context, facts Facts) ([]byte, error) {
	drawn, err := drawer.typst(ctx, facts, compileCommand, []string{standardOutput}, pixelsOption, pixelsPerInch, formatOption, pictureFormat)
	if err != nil {
		return nil, fmt.Errorf("draw the card of %+v: %w", facts, err)
	}
	return drawn, nil
}

// typst runs a command of Typst's on the template with a card's facts, the words its command takes after the template
// and then its options, and is what it writes to its standard output.
func (drawer Drawer) typst(ctx context.Context, facts Facts, command string, after []string, options ...string) ([]byte, error) {
	read, err := inputOf(facts)
	if err != nil {
		return nil, err
	}
	input, err := json.Marshal(read)
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
	arguments := slices.Concat([]string{command, filepath.Join(folder, templateFile)}, after, []string{
		rootOption, folder, noSystemFonts, noEmbeddedFonts, fontsOption, drawer.faces, inputOption, factsKey + inputJoint + string(input),
	}, options)
	return run(ctx, drawer.program, arguments...)
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

// sideInput is a card of the résumé's side as the template reads it: its kind, its title, its line and the way in.
type sideInput struct {
	Kind  string     `json:"kind"`
	Title titleInput `json:"title"`
	Line  string     `json:"line"`
	WayIn string     `json:"way_in"`
}

// postInput is a post's card as the template reads it: the cards' mark in block letters, and its words.
type postInput struct {
	Kind   string      `json:"kind"`
	Mark   arrangement `json:"mark"`
	Day    string      `json:"day"`
	Series string      `json:"series"`
	Title  string      `json:"title"`
	Lead   string      `json:"lead"`
	WayIn  string      `json:"way_in"`
}

// titleInput is a title of the résumé's side: its words, and the arrangements it may stand in, in block letters, the
// first on one line, of which the template takes the first that fits; none where the font does not draw every letter
// of a name, which the template then sets in words.
type titleInput struct {
	Words        string        `json:"words"`
	Arrangements []arrangement `json:"arrangements"`
}

// arrangement is block letters on lines as the template places them (look.Drawing): the lines' words, the drawing's
// width, in cells, how far down its ink reaches, in half rows, its runs of filled cells, its cursor, none where none
// follows, and a cell's shape, as wide and as tall as these units.
type arrangement struct {
	Lines        []string    `json:"lines"`
	Width        int         `json:"width"`
	Foot         int         `json:"foot"`
	Rectangles   []rectangle `json:"rectangles"`
	Cursor       *rectangle  `json:"cursor"`
	CellUnits    int         `json:"cell_units"`
	HalfRowUnits int         `json:"half_row_units"`
}

// rectangle is a run of a drawing's filled cells, or its cursor, in cells across and half rows down.
type rectangle struct {
	Left   int `json:"left"`
	Top    int `json:"top"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// inputOf is a card's facts as the template reads them; a fault where the block letters cannot draw the cards' mark,
// which no card sets in words. gochecksumtype holds the switch to every kind.
func inputOf(facts Facts) (any, error) {
	switch typed := facts.(type) {
	case Resume:
		return sideInput{Kind: kindResume, Title: nameTitle(typed.Name), Line: typed.Line, WayIn: typed.WayIn}, nil
	case Project:
		return sideInput{Kind: kindProject, Title: projectTitle(typed.Name), Line: typed.Line, WayIn: typed.WayIn}, nil
	case List:
		mark, err := markArrangement(typed.Mark)
		title := titleInput{Words: typed.Mark, Arrangements: []arrangement{mark}}
		return sideInput{Kind: kindList, Title: title, Line: typed.Line, WayIn: typed.WayIn}, err
	case Post:
		mark, err := markArrangement(typed.Mark)
		return postInput{
			Kind: kindPost, Mark: mark, Day: typed.Day, Series: typed.Series, Title: typed.Title, Lead: typed.Lead, WayIn: typed.WayIn,
		}, err
	}
	return nil, nil
}

// nameTitle is the author's name in block letters with the cursor after it, as the page draws the name.
func nameTitle(name string) titleInput {
	drawing, drawn := look.NameDrawing(name)
	if !drawn {
		return titleInput{Words: name, Arrangements: []arrangement{}}
	}
	return titleInput{Words: name, Arrangements: []arrangement{arrangementOf(drawing)}}
}

// projectTitle is a project's name in block letters: on one line, then broken once after each hyphen in turn, onto two.
func projectTitle(name string) titleInput {
	words := strings.ToUpper(name)
	pieces := slices.DeleteFunc(strings.SplitAfter(words, nameBreak), func(piece string) bool { return piece == "" })
	arrangements := []arrangement{}
	for _, lines := range slices.Concat([][]string{{words}}, breaksOf(pieces)) {
		drawing, drawn := look.NameFont().Draw(lines)
		if !drawn {
			return titleInput{Words: name, Arrangements: []arrangement{}}
		}
		arrangements = append(arrangements, arrangementOf(drawing))
	}
	return titleInput{Words: name, Arrangements: arrangements}
}

// breaksOf are pieces broken once onto two lines, at each place between two of them in turn.
func breaksOf(pieces []string) [][]string {
	breaks := [][]string{}
	for at := 1; at < len(pieces); at++ {
		breaks = append(breaks, []string{strings.Join(pieces[:at], ""), strings.Join(pieces[at:], "")})
	}
	return breaks
}

// markArrangement is the cards' mark in block letters, on one line; a fault where the font does not draw every letter
// of it.
func markArrangement(mark string) (arrangement, error) {
	drawing, drawn := look.NameFont().Draw([]string{strings.ToUpper(mark)})
	if !drawn {
		return arrangementOf(drawing), fmt.Errorf("%w: %q", errUndrawnMark, mark)
	}
	return arrangementOf(drawing), nil
}

// arrangementOf is a drawing of lines of block letters as the template places it.
func arrangementOf(drawing look.Drawing) arrangement {
	placed := arrangement{
		Lines: drawing.Lines, Width: drawing.Width, Foot: drawing.Foot(), Rectangles: make([]rectangle, 0, len(drawing.Rectangles)),
		Cursor: nil, CellUnits: look.CellUnits, HalfRowUnits: look.HalfRowUnits,
	}
	for _, run := range drawing.Rectangles {
		placed.Rectangles = append(placed.Rectangles, rectangleOf(run))
	}
	if drawing.Cursor != nil {
		cursor := rectangleOf(*drawing.Cursor)
		placed.Cursor = &cursor
	}
	return placed
}

// rectangleOf is a rectangle of a drawing as the template places it.
func rectangleOf(drawn look.Rectangle) rectangle {
	return rectangle{Left: drawn.Left, Top: drawn.Top, Width: drawn.Width, Height: drawn.Height}
}
