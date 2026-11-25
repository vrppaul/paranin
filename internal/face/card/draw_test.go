package card

import (
	"bytes"
	"encoding/json"
	"errors"
	"image/png"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/vrppaul/paranin/internal/fixtures"
)

// The Typst the tests draw with, found on the path as the build finds it, and one of another version
const (
	typstProgram = "typst"
	otherTypst   = "testdata/typst-0.14.0"
)

// How a test asks Typst what the template chose: by its query, of the metadata labelled as a choice, each one's value
const (
	queryCommand = "query"
	choiceLabel  = "<choice>"
	fieldOption  = "--field"
	valueField   = "value"
)

// How big a card is, as Open Graph's image of a page, in pixels
const (
	cardWidth  = 1200
	cardHeight = 630
)

// How the template says whether words were cut and whether block letters shrank
const (
	whole    = false
	cut      = true
	asDrawn  = false
	shrunken = true
)

// choice is what the template states it chose for a part of a card: the words it shows, whether it cut them and how
// many lines they take; or the rows block letters stand on and whether they shrank to fit.
type choice struct {
	Part   string   `json:"part"`
	Words  string   `json:"words"`
	Cut    bool     `json:"cut"`
	Lines  int      `json:"lines"`
	Rows   []string `json:"rows"`
	Shrank bool     `json:"shrank"`
}

// A Typst of another version than the cards are drawn with is refused before it draws a card.
func TestATypstOfAnotherVersionIsRefused(test *testing.T) {
	if _, err := NewDrawer(test.Context(), otherTypst, fixtures.InModule(test, FacesFolder)); !errors.Is(err, errTypstVersion) {
		test.Errorf("NewDrawer(%s) = %v, want %v", otherTypst, err, errTypstVersion)
	}
}

// A folder that lacks a face the cards are set in is refused before a card is drawn, where Typst would set another
// face and the template's measures would fail: none at all, and one of other files.
func TestAFolderLackingTheCardsFacesIsRefused(test *testing.T) {
	otherFiles := test.TempDir()
	fixtures.WriteIn(test, otherFiles, "JetBrainsMono-Regular.woff2", "wOF2")
	cases := []struct {
		name  string
		faces string
	}{
		{name: "no folder", faces: filepath.Join(test.TempDir(), "none")},
		{name: "a folder of other files", faces: otherFiles},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if _, err := NewDrawer(test.Context(), typstProgram, testCase.faces); !errors.Is(err, errNoFaces) {
				test.Errorf("NewDrawer(%s, %q) = %v, want %v", typstProgram, testCase.faces, err, errNoFaces)
			}
		})
	}
}

// A card is drawn as a PNG of Open Graph's size for a page's image.
func TestACardIsDrawnAsAPictureOfItsSize(test *testing.T) {
	drawn, err := typstDrawer(test).Draw(test.Context(), Resume{Name: "Mira Example", Line: "Software Engineer", WayIn: fixtureWayIn})
	if err != nil {
		test.Fatalf("Draw = %v, want a picture", err)
	}
	size, err := png.DecodeConfig(bytes.NewReader(drawn))
	if err != nil || size.Width != cardWidth || size.Height != cardHeight {
		test.Errorf("the card drawn = %d×%d (%v), want a PNG %d×%d", size.Width, size.Height, err, cardWidth, cardHeight)
	}
}

// The cards' mark stands in block letters on every card, so a mark with a letter they cannot draw is refused, on the
// list's card as on a post's, before Typst is run.
func TestAMarkTheBlockLettersCannotDrawIsRefused(test *testing.T) {
	cases := []struct {
		name  string
		facts Facts
	}{
		{name: "the list", facts: List{Mark: "блог", Line: "Что читает автор, и как.", WayIn: fixtureWayIn}},
		{name: "a post", facts: Post{Mark: "блог", Day: "15 нояб. 2026", Series: "", Title: "Медленное чтение", Lead: "", WayIn: fixtureWayIn}},
	}
	drawer := typstDrawer(test)
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if _, err := drawer.Draw(test.Context(), testCase.facts); !errors.Is(err, errUndrawnMark) {
				test.Errorf("Draw(%+v) = %v, want %v", testCase.facts, err, errUndrawnMark)
			}
		})
	}
}

// The template fits each card's words and letters to it, and says what it chose: the résumé's name on the lines the
// profile breaks it into; a project's name on one line where it fits, else broken after a hyphen onto two, shrunk only
// where two would not fit, or set in words where the block letters cannot draw it; the cards' mark; a post's title on
// one line, or on two, cut after its last whole word that fits, and its lead on the lines left; a word wider than the
// card cut inside it; and no words where there are none.
func TestTheTemplateFitsEachCardsWordsAndStatesTheChoices(test *testing.T) {
	cases := []struct {
		name  string
		facts Facts
		want  []choice
	}{
		{
			name:  "the résumé",
			facts: Resume{Name: "Mira Example", Line: "Software Engineer", WayIn: fixtureWayIn},
			want:  []choice{lettersChosen("title", asDrawn, "MIRA", "EXAMPLE"), wordsChosen("line", "Software Engineer", whole, 1)},
		},
		{
			name:  "a project's name that fits on one line",
			facts: Project{Name: "tidy-tool", Line: "Tidies files.", WayIn: fixtureWayIn},
			want:  []choice{lettersChosen("title", asDrawn, "TIDY-TOOL"), wordsChosen("line", "Tidies files.", whole, 1)},
		},
		{
			name:  "a project's name broken after a hyphen",
			facts: Project{Name: "tidy-tool-for-files", Line: "Tidies files.", WayIn: fixtureWayIn},
			want:  []choice{lettersChosen("title", asDrawn, "TIDY-TOOL-", "FOR-FILES"), wordsChosen("line", "Tidies files.", whole, 1)},
		},
		{
			name:  "a project's name too long for two lines at its cell",
			facts: Project{Name: "tidy-tool-for-every-folder", Line: "Tidies files.", WayIn: fixtureWayIn},
			want:  []choice{lettersChosen("title", shrunken, "TIDY-TOOL-FOR-", "EVERY-FOLDER"), wordsChosen("line", "Tidies files.", whole, 1)},
		},
		{
			name:  "a project's name the block letters cannot draw",
			facts: Project{Name: "tidy.tool", Line: "Tidies files.", WayIn: fixtureWayIn},
			want:  []choice{wordsChosen("title", "tidy.tool", whole, 1), wordsChosen("line", "Tidies files.", whole, 1)},
		},
		{
			name:  "a project with no tagline",
			facts: Project{Name: "tidy-tool", Line: "", WayIn: fixtureWayIn},
			want:  []choice{lettersChosen("title", asDrawn, "TIDY-TOOL"), wordsChosen("line", "", whole, 0)},
		},
		{
			name:  "the list",
			facts: List{Mark: fixtureMark, Line: "What the writer reads, and how.", WayIn: listWay},
			want:  []choice{lettersChosen("title", asDrawn, "BLOG"), wordsChosen("line", "What the writer reads, and how.", whole, 1)},
		},
		{
			name: "a post whose title stands on one line",
			facts: Post{
				Mark: fixtureMark, Day: "10 Nov 2026", Series: "Reading", Title: "Reading slowly", Lead: "Why a page read twice is read once.",
				WayIn: postWay,
			},
			want: []choice{
				lettersChosen("mark", asDrawn, "BLOG"), wordsChosen("title", "Reading slowly", whole, 1),
				wordsChosen("lead", "Why a page read twice is read once.", whole, 1),
			},
		},
		{
			name: "a post whose title and lead run past their lines",
			facts: Post{
				Mark: fixtureMark, Day: "10 Nov 2026", Series: "", Title: "Why a page read twice is read once, if the first reading was slow",
				Lead: "A page read twice is read once, if the first reading was slow, and the margins kept what it said.", WayIn: postWay,
			},
			want: []choice{
				lettersChosen("mark", asDrawn, "BLOG"), wordsChosen("title", "Why a page read twice is read…", cut, 2),
				wordsChosen("lead", "A page read twice is read once, if the…", cut, 1),
			},
		},
		{
			name: "a post whose title is one word wider than the card",
			facts: Post{
				Mark: fixtureMark, Day: "10 Nov 2026", Series: "", Title: "Internationalization", Lead: "Why a page read twice is read once.",
				WayIn: postWay,
			},
			want: []choice{
				lettersChosen("mark", asDrawn, "BLOG"), wordsChosen("title", "Internationalizat…", cut, 1),
				wordsChosen("lead", "Why a page read twice is read once.", whole, 1),
			},
		},
	}
	drawer := typstDrawer(test)
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := choicesOf(test, drawer, testCase.facts); !reflect.DeepEqual(got, testCase.want) {
				test.Errorf("the choices for %+v = %+v, want %+v", testCase.facts, got, testCase.want)
			}
		})
	}
}

// wordsChosen is the choice of the words shown for a part of a card, whether they were cut and the lines they take.
func wordsChosen(part, words string, wasCut bool, lines int) choice {
	return choice{Part: part, Words: words, Cut: wasCut, Lines: lines, Rows: nil, Shrank: false}
}

// lettersChosen is the choice of the rows block letters stand on for a part of a card, and whether they shrank.
func lettersChosen(part string, shrank bool, rows ...string) choice {
	return choice{Part: part, Words: "", Cut: false, Lines: 0, Rows: rows, Shrank: shrank}
}

// choicesOf are the choices the template states for a card's facts, as Typst's query of them says.
func choicesOf(test *testing.T, drawer Drawer, facts Facts) []choice {
	test.Helper()
	said, err := drawer.typst(test.Context(), facts, queryCommand, []string{choiceLabel}, fieldOption, valueField)
	if err != nil {
		test.Fatalf("query the choices for %+v: %v", facts, err)
	}
	choices := []choice{}
	if err := json.Unmarshal(said, &choices); err != nil {
		test.Fatalf("read the choices %s: %v", said, err)
	}
	return choices
}

// typstDrawer draws with the Typst on the path, in the cards' faces; the test ends where there is none, or another
// version.
func typstDrawer(test *testing.T) Drawer {
	test.Helper()
	program, err := exec.LookPath(typstProgram)
	if err != nil {
		test.Fatalf("find %s: %v; install it with go run ./scripts/typst", typstProgram, err)
	}
	drawer, err := NewDrawer(test.Context(), program, fixtures.InModule(test, FacesFolder))
	if err != nil {
		test.Fatalf("NewDrawer(%s) = %v", program, err)
	}
	return drawer
}
