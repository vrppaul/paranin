package main

import (
	"errors"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/vrppaul/paranin/internal/assets"
	"github.com/vrppaul/paranin/internal/face/card"
	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/release"
)

// A post published after the tests' day, which comes out on the clock with no build: its file in the content's folder,
// its text, and the path it is served at
const (
	laterPostFile = "posts/coming-soon/post.en.md"
	laterPostText = "---\nid: 3\ntitle: Coming soon\ndescription: A post to come.\nstatus: published\npublished: 2026-12-01T09:00+00:00\n---\nIts text.\n"
	laterPostPath = "/blog/coming-soon"
)

// The pages of the fixture content a preview shows, by the path each is served at: the résumé, the project with a page,
// and the blog's lists and its post, in English and in Russian
func fixtureCardPaths() []string {
	return []string{
		"/", "/blog", "/blog/reading-slowly", "/blog/series", "/p/tidy-tool", "/ru/blog", "/ru/blog/reading-slowly", "/ru/blog/series",
	}
}

// The cards drawn are those of every page a preview shows from the moment they are drawn on, a post dated later among
// them, which comes out with no build; a tab shares its page's, and a draft, which no preview shows, has none.
func TestEveryPageThatWillShowAPreviewHasACard(test *testing.T) {
	cases := []struct {
		name   string
		change func(test *testing.T, folder string)
		want   []string
	}{
		{name: "the fixture, with a draft", change: fixtures.KeepFolder, want: fixtureCardPaths()},
		{name: "a post dated later", change: writeLaterPost, want: slices.Sorted(slices.Values(append(fixtureCardPaths(), laterPostPath)))},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			folder := fixtures.Copy(test, fixtures.ContentFolder(test))
			testCase.change(test, folder)
			got := []string{}
			for _, owned := range settledCards(fixtures.LoadedFrom(test, folder), fixtures.Day()) {
				got = append(got, owned.Path)
			}
			slices.Sort(got)
			if !slices.Equal(got, testCase.want) {
				test.Errorf("the pages with cards = %q, want %q", got, testCase.want)
			}
		})
	}
}

// writeLaterPost writes a post dated after the tests' day beside the fixture's.
//
//wrapper:a case of the table names the change it writes
func writeLaterPost(test *testing.T, folder string) {
	fixtures.WriteIn(test, folder, laterPostFile, laterPostText)
}

// The pictures a test draws by hand, each a card's as the folder holds it
const (
	resumePicture  = "the résumé's card"
	listPicture    = "the list's card"
	projectPicture = "the project's card"
)

// Cards are written into their folder whole, each page's by a name that changes with what it shows, two pages that show
// the same sharing one, under a manifest naming each; a card no page shows any more leaves the folder, which holds
// nothing else.
func TestTheCardsFolderHoldsTheCardsDrawnAlone(test *testing.T) {
	folder := filepath.Join(test.TempDir(), "cards")
	before := []drawnCard{{path: "/", picture: []byte(resumePicture)}, {path: "/blog", picture: []byte(listPicture)}}
	after := []drawnCard{
		{path: "/", picture: []byte(resumePicture)},
		{path: "/p/tidy-tool", picture: []byte(projectPicture)},
		{path: "/blog/series", picture: []byte(projectPicture)},
	}
	for _, drawn := range [][]drawnCard{before, after} {
		if err := writeCards(folder, drawn); err != nil {
			test.Fatalf("writeCards(%q) = %v, want no error", folder, err)
		}
	}
	cards, err := assets.LoadCards(os.DirFS(folder))
	if err != nil {
		test.Fatalf("LoadCards(%q) = %v, want the cards", folder, err)
	}
	pictures := map[string]string{}
	for name, file := range cards.Files() {
		pictures[name] = string(file.Body)
	}
	shown := map[string]string{}
	for _, page := range []string{"/", "/p/tidy-tool", "/blog/series"} {
		name, _ := cards.NameOf(page)
		shown[page] = pictures[name]
	}
	wantShown := map[string]string{"/": resumePicture, "/p/tidy-tool": projectPicture, "/blog/series": projectPicture}
	_, listKept := cards.NameOf("/blog")
	wantFiles := []string{assets.CardFile([]byte(resumePicture)), assets.CardFile([]byte(projectPicture)), assets.CardsManifest}
	slices.Sort(wantFiles)
	if files := folderFiles(test, folder); !maps.Equal(shown, wantShown) || listKept || !slices.Equal(files, wantFiles) {
		test.Errorf("the cards shown = %q, the list's kept: %t, the folder holding %q; want %q, none, %q", shown, listKept, files,
			wantShown, wantFiles)
	}
}

// folderFiles are the names of the files of a folder, in order.
func folderFiles(test *testing.T, folder string) []string {
	entries, err := os.ReadDir(folder)
	if err != nil {
		test.Fatalf("read %s: %v", folder, err)
	}
	names := []string{}
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

// The page of the fixture's post, by the path it is served at, and its line in the cards' fixture's manifest
const (
	fixturePostPath = "/blog/reading-slowly"
	fixturePostCard = `  "/blog/reading-slowly": "page-fixture.png",` + "\n"
)

// A release is refused where a page a preview shows from the moment it loads on lacks its card, each such page named:
// a post's, and a post's dated later, which comes out with no build.
func TestAReleaseLackingACardIsRefused(test *testing.T) {
	cases := []struct {
		name         string
		changeCards  func(test *testing.T, folder string)
		changeOthers func(test *testing.T, folder string)
		want         missingCards
	}{
		{name: "every card there", changeCards: fixtures.KeepFolder, changeOthers: fixtures.KeepFolder, want: nil},
		{name: "a post's card left out", changeCards: leavePostCardOut, changeOthers: fixtures.KeepFolder, want: missingCards{fixturePostPath}},
		{name: "a post dated later with no card", changeCards: fixtures.KeepFolder, changeOthers: writeLaterPost, want: missingCards{laterPostPath}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			contentFolder, cardsFolder := fixtures.Copy(test, fixtures.ContentFolder(test)), fixtures.Copy(test, fixtures.CardsFolder(test))
			testCase.changeOthers(test, contentFolder)
			testCase.changeCards(test, cardsFolder)
			cards, err := assets.LoadCards(os.DirFS(cardsFolder))
			if err != nil {
				test.Fatalf("LoadCards = %v, want the cards", err)
			}
			var missing missingCards
			errors.As(missingCardsOf(fixtures.LoadedFrom(test, contentFolder), cards, fixtures.Day()), &missing)
			if !slices.Equal(missing, testCase.want) {
				test.Errorf("the pages lacking a card = %q, want %q", missing, testCase.want)
			}
		})
	}
}

// leavePostCardOut takes the fixture post's card out of the cards' manifest.
//
//wrapper:a case of the table names the change it writes
func leavePostCardOut(test *testing.T, folder string) {
	fixtures.ReplaceIn(test, folder, assets.CardsManifest, fixturePostCard, "")
}

// The cards drawn over the fixture content, by the Typst found on the path, are the cards a release of it takes.
func TestTheCardsDrawnAreTheCardsAReleaseTakes(test *testing.T) {
	drawer, err := card.NewDrawer(test.Context(), typstProgram, fixtures.CardFaces(test))
	if err != nil {
		test.Fatalf("NewDrawer(%s) = %v, want a drawer", typstProgram, err)
	}
	folders := release.Folders{Content: fixtures.ContentFolder(test), Build: fixtures.BuildFolder(test), Cards: test.TempDir()}
	drawing := cardDrawing{drawer: drawer, content: folders.Content, cards: folders.Cards}
	drawn, err := drawing.draw(test.Context(), fixtures.Day())
	if err != nil {
		test.Fatalf("draw = %v, want the cards drawn", err)
	}
	if _, err := release.Load(folders, releaseCheckedNow); drawn != len(fixtureCardPaths()) || err != nil {
		test.Errorf("drew %d cards, and a release of them loads with %v; want %d, and no fault", drawn, err, len(fixtureCardPaths()))
	}
}
