package assets_test

import (
	"maps"
	"os"
	"testing"

	"github.com/vrppaul/paranin/internal/assets"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// The cards' fixture's cards a test reads and changes, from its root, beside its manifest
const (
	resumeCard = "resume-fixture.png"
	pageCard   = "page-fixture.png"
)

// The cards are every file their manifest names, each once, by its name, as a PNG, and each page's card is the file the
// manifest names for the path the page is served at; a file the manifest names not is none of them.
func TestTheCardsAreTheFilesTheirManifestNames(test *testing.T) {
	cards := fixtures.LoadedCards(test)
	wantFiles := map[string]assets.File{
		resumeCard: {ContentType: assets.CardType, Body: []byte("PNG résumé fixture\n")},
		pageCard:   {ContentType: assets.CardType, Body: []byte("PNG page fixture\n")},
	}
	if got := maps.Collect(cards.Files()); !maps.EqualFunc(got, wantFiles, sameFile) {
		test.Errorf("Files() = %q, want %q", got, wantFiles)
	}
	cases := []struct {
		page      string
		wantName  string
		wantFound bool
	}{
		{page: "/", wantName: resumeCard, wantFound: true},
		{page: "/ru/blog/reading-slowly", wantName: pageCard, wantFound: true},
		{page: "/experience", wantName: "", wantFound: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.page, func(test *testing.T) {
			if name, found := cards.NameOf(testCase.page); name != testCase.wantName || found != testCase.wantFound {
				test.Errorf("NameOf(%q) = %q, %t; want %q, %t", testCase.page, name, found, testCase.wantName, testCase.wantFound)
			}
		})
	}
}

// Cards the site cannot serve are refused: no manifest, one that does not read, a file it names missing or of a type
// no card is drawn in.
func TestCardsNoSiteServesAreRefused(test *testing.T) {
	cases := []struct {
		name   string
		change func(test *testing.T, folder string)
	}{
		{name: "no manifest", change: func(test *testing.T, folder string) { removeFile(test, folder, manifestPath) }},
		{name: "a manifest that does not read", change: func(test *testing.T, folder string) { fixtures.ReplaceIn(test, folder, manifestPath, "{", "[") }},
		{name: "a card missing", change: func(test *testing.T, folder string) { removeFile(test, folder, pageCard) }},
		{name: "a card of another type", change: func(test *testing.T, folder string) { renameNamed(test, folder, resumeCard, "resume-fixture.svg") }},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			folder := fixtures.Copy(test, fixtures.CardsFolder(test))
			testCase.change(test, folder)
			if cards, err := assets.LoadCards(os.DirFS(folder)); cards != nil || err == nil {
				test.Errorf("LoadCards = %v, %v; want no cards and a fault", cards, err)
			}
		})
	}
}
