package card

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/golden"
)

// The Typst the tests draw with, found on the path as the build finds it, and one of another version
const (
	typstProgram = "typst"
	otherTypst   = "testdata/typst-0.14.0"
)

// A Typst of another version than the cards are drawn with is refused before it draws a card.
func TestATypstOfAnotherVersionIsRefused(test *testing.T) {
	if _, err := NewDrawer(test.Context(), otherTypst, fixtures.CardFaces(test)); !errors.Is(err, errTypstVersion) {
		test.Errorf("NewDrawer(%s) = %v, want %v", otherTypst, err, errTypstVersion)
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

// Each kind of card draws as its golden picture, as the template fits its words: the résumé's name on the lines the
// profile breaks it into, the cursor after it; a project's name on one line where it fits, else broken after a hyphen;
// the blog's list; a post's title on one line and two lead lines under it, or on two, cut after its last whole word
// that fits, and one lead line under it; a translated post under the cards' mark in Latin; and a project's name the
// block letters cannot draw, set in words.
func TestEachKindOfCardDrawsAsItsGoldenPicture(test *testing.T) {
	cases := []struct {
		file  string
		facts Facts
	}{
		{file: "resume.png", facts: Resume{Name: "Mira Example", Line: "Software Engineer", WayIn: fixtureWayIn}},
		{file: "project.png", facts: Project{Name: "tidy-tool", Line: "Tidies files.", WayIn: fixtureWayIn}},
		{
			file:  "project-broken.png",
			facts: Project{Name: "tidy-tool-for-every-folder", Line: "Tidies every file of every folder it is shown, and then their names.", WayIn: fixtureWayIn},
		},
		{file: "project-in-words.png", facts: Project{Name: "tidy.tool", Line: "Tidies files.", WayIn: fixtureWayIn}},
		{file: "list.png", facts: List{Mark: fixtureMark, Line: "What the writer reads, and how.", WayIn: listWay}},
		{
			file: "post.png",
			facts: Post{
				Mark: fixtureMark, Day: "10 Nov 2026", Series: "Reading", Title: "Reading slowly", Lead: "Why a page read twice is read once.",
				WayIn: postWay,
			},
		},
		{
			file: "post-long.png",
			facts: Post{
				Mark: fixtureMark, Day: "10 Nov 2026", Series: "", Title: "Why a page read twice is read once, if the first reading was slow",
				Lead: "A page read twice is read once, if the first reading was slow, and the margins kept what it said.", WayIn: postWay,
			},
		},
		{
			file: "post-translated.png",
			facts: Post{
				Mark: fixtureMark, Day: "15 нояб. 2026", Series: "Чтение", Title: "Медленное чтение",
				Lead: "Почему страницу, прочитанную дважды, читают один раз.", WayIn: translatedWay,
			},
		},
	}
	drawer := typstDrawer(test)
	for _, testCase := range cases {
		test.Run(testCase.file, func(test *testing.T) {
			drawn, err := drawer.Draw(test.Context(), testCase.facts)
			if err != nil {
				test.Fatalf("Draw(%+v) = %v, want a picture", testCase.facts, err)
			}
			golden.MatchPicture(test, testCase.file, drawn)
		})
	}
}

// typstDrawer draws with the Typst on the path, in the cards' faces; the test ends where there is none, or another
// version.
func typstDrawer(test *testing.T) Drawer {
	test.Helper()
	program, err := exec.LookPath(typstProgram)
	if err != nil {
		test.Fatalf("find %s: %v; install it with go run ./scripts/typst", typstProgram, err)
	}
	drawer, err := NewDrawer(test.Context(), program, fixtures.CardFaces(test))
	if err != nil {
		test.Fatalf("NewDrawer(%s) = %v", program, err)
	}
	return drawer
}
