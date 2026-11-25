package main

import (
	"errors"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

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
// the same sharing one, under a manifest naming each; a card no page shows any more leaves the folder, and so does a
// file a drawing that stopped left beside its place, so the folder holds nothing else.
func TestTheCardsFolderHoldsTheCardsDrawnAlone(test *testing.T) {
	folder := filepath.Join(test.TempDir(), "cards")
	before := []drawnCard{{path: "/", picture: []byte(resumePicture)}, {path: "/blog", picture: []byte(listPicture)}}
	after := []drawnCard{
		{path: "/", picture: []byte(resumePicture)},
		{path: "/p/tidy-tool", picture: []byte(projectPicture)},
		{path: "/blog/series", picture: []byte(projectPicture)},
	}
	writeTestCards(test, folder, before)
	fixtures.WriteIn(test, folder, assets.CardFile([]byte(listPicture))+replacementEnding, listPicture)
	writeTestCards(test, folder, after)
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

// writeTestCards writes cards into a folder, or ends the test where they cannot be written.
func writeTestCards(test *testing.T, folder string, drawn []drawnCard) {
	if err := writeCards(folder, drawn); err != nil {
		test.Fatalf("writeCards(%q) = %v, want no error", folder, err)
	}
}

// A name shaped as a card's, the start of a sha256 in small hex digits and the picture's ending
const cardShapedName = "0123456789abcdef.png"

// A folder that holds anything a drawing of the cards does not write is refused before anything in it changes, so a
// folder named by mistake, as a build's or the repository's, loses nothing: a folder in it, a file of another name, a
// picture not named by its hash, a name of sixteen letters not hex digits or of seventeen hex digits, and a folder or a
// link named as a card.
func TestAFolderHoldingMoreThanCardsIsRefused(test *testing.T) {
	cases := []struct {
		name  string
		place func(test *testing.T, folder string)
	}{
		{name: "a folder in it, as a build's", place: writtenAt("web/manifest.json")},
		{name: "a file of another name", place: writtenAt("notes.txt")},
		{name: "a picture not named by its hash", place: writtenAt("photo.png")},
		{name: "sixteen letters not hex digits", place: writtenAt("ghijklmnopqrstuv.png")},
		{name: "seventeen hex digits", place: writtenAt("0123456789abcdef0.png")},
		{name: "a folder named as a card", place: writtenAt(cardShapedName + "/inside.png")},
		{name: "a link named as a card", place: linkedAsACard},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			folder := test.TempDir()
			testCase.place(test, folder)
			before := folderFiles(test, folder)
			err := writeCards(folder, []drawnCard{{path: "/", picture: []byte(resumePicture)}})
			if files := folderFiles(test, folder); !errors.Is(err, errForeignFile) || !slices.Equal(files, before) {
				test.Errorf("writeCards = %v, the folder holding %q; want %v, the folder holding %q", err, files, errForeignFile, before)
			}
		})
	}
}

// writtenAt places a file at a path of a folder, made with the folders it stands in.
func writtenAt(path string) func(test *testing.T, folder string) {
	return func(test *testing.T, folder string) { fixtures.WriteIn(test, folder, path, resumePicture) }
}

// linkedAsACard places a link named as a card in a folder, leading to a file beside the folder.
func linkedAsACard(test *testing.T, folder string) {
	target := filepath.Join(test.TempDir(), resumeCardFile)
	fixtures.WriteIn(test, filepath.Dir(target), resumeCardFile, resumePicture)
	if err := os.Symlink(target, filepath.Join(folder, cardShapedName)); err != nil {
		test.Fatalf("link %s: %v", cardShapedName, err)
	}
}

// The file a link named as a card leads to
const resumeCardFile = "resume.png"

// A drawing that changes no card writes no file, so a site watching the folder loads nothing again.
func TestADrawingThatChangesNoCardWritesNoFile(test *testing.T) {
	folder := test.TempDir()
	drawn := []drawnCard{{path: "/", picture: []byte(resumePicture)}}
	writeTestCards(test, folder, drawn)
	written := fixtures.Day()
	for _, name := range folderFiles(test, folder) {
		if err := os.Chtimes(filepath.Join(folder, name), written, written); err != nil {
			test.Fatalf("set the time of %s: %v", name, err)
		}
	}
	writeTestCards(test, folder, drawn)
	times := []time.Time{}
	for _, name := range folderFiles(test, folder) {
		info, err := os.Stat(filepath.Join(folder, name))
		if err != nil {
			test.Fatalf("stat %s: %v", name, err)
		}
		times = append(times, info.ModTime().UTC())
	}
	if want := []time.Time{written, written}; !slices.Equal(times, want) {
		test.Errorf("the files written at %v, want %v: untouched", times, want)
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
	folders := release.Folders{Content: fixtures.ContentFolder(test), Build: fixtures.BuildFolder(test), Cards: test.TempDir()}
	drawing := cardDrawing{drawer: typstDrawer(test), content: folders.Content, cards: folders.Cards}
	drawn, err := drawing.draw(test.Context(), fixtures.Day())
	if err != nil {
		test.Fatalf("draw = %v, want the cards drawn", err)
	}
	if _, err := release.Load(folders, releaseChecks()); drawn != len(fixtureCardPaths()) || err != nil {
		test.Errorf("drew %d cards, and a release of them loads with %v; want %d, and no fault", drawn, err, len(fixtureCardPaths()))
	}
}

// A drawing refused leaves the cards' folder as it was: of content the site's own check refuses, its résumé's tab
// titled as the blog takes the blog's address, and of a card that cannot be drawn, its mark in letters the block
// letters do not draw.
func TestARefusedDrawingLeavesTheCardsAsTheyWere(test *testing.T) {
	cases := []struct {
		name        string
		original    string
		replacement string
	}{
		{name: "content the site's check refuses", original: "experience: Experience", replacement: "experience: Blog"},
		{name: "a card that cannot be drawn", original: "blog: blog", replacement: "blog: блог"},
	}
	drawer := typstDrawer(test)
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			contentFolder, cardsFolder := fixtures.Copy(test, fixtures.ContentFolder(test)), test.TempDir()
			fixtures.ReplaceIn(test, contentFolder, wordsFile, testCase.original, testCase.replacement)
			writeTestCards(test, cardsFolder, []drawnCard{{path: "/", picture: []byte(resumePicture)}})
			before := folderFiles(test, cardsFolder)
			drawing := cardDrawing{drawer: drawer, content: contentFolder, cards: cardsFolder}
			if _, err := drawing.draw(test.Context(), fixtures.Day()); err == nil || !slices.Equal(folderFiles(test, cardsFolder), before) {
				test.Errorf("draw = %v, the folder holding %q; want a fault, the folder holding %q", err, folderFiles(test, cardsFolder), before)
			}
		})
	}
}

// The site's words in its own language, in the content's folder
const wordsFile = "i18n/en.yaml"

// typstDrawer draws with the Typst found on the path, in the faces the cards are set in, or ends the test.
func typstDrawer(test *testing.T) card.Drawer {
	drawer, err := card.NewDrawer(test.Context(), typstProgram, fixtures.InModule(test, card.FacesFolder))
	if err != nil {
		test.Fatalf("NewDrawer(%s) = %v, want a drawer", typstProgram, err)
	}
	return drawer
}
