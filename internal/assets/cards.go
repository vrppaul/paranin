package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"iter"
	"maps"
	"path"
	"slices"
	"strings"
)

// The cards' manifest, at their folder's root, which their drawing writes last and the site never serves
const CardsManifest = manifestFile

// CardNames are the name of each page's card's file, by the path the page is served at, as the cards' manifest writes
// them: two pages whose cards are the same picture name one file.
type CardNames map[string]string

// How a card's file is named: by the start of the sha256 of what it holds, 64 bits in small hex digits, so its name
// changes with what the card shows and no two cards of a site share one, then the ending of the one type a card is
// drawn in
const (
	cardHashLetters = 16
	hexDigits       = "0123456789abcdef"
	cardEnding      = ".png"
)

// What every card is: the type it is drawn and served in, and its size in pixels, the one a link's preview wants
// (docs/SEO.md, A card), which the drawing sets the template to and a page's head names the card by
const (
	CardType   = "image/png"
	CardWidth  = 1200
	CardHeight = 630
)

// The faults the cards' folder can hold
var errCards = errors.New("the cards have a fault")

// Cards are the cards the build drew of the site's pages (ADR-0005), read whole from their folder: each page's card's
// name by the path the page is served at, and every card's file by its name.
type Cards struct {
	names CardNames
	files map[string]File
}

// CardFile is the name of the file a card is written to, by what it holds.
func CardFile(drawn []byte) string {
	sum := sha256.Sum256(drawn)
	return hex.EncodeToString(sum[:])[:cardHashLetters] + cardEnding
}

// IsCardFile is whether a name is one CardFile gives.
func IsCardFile(name string) bool {
	hash, isPicture := strings.CutSuffix(name, cardEnding)
	return isPicture && len(hash) == cardHashLetters && strings.Trim(hash, hexDigits) == ""
}

// LoadCards reads the cards from their folder: the cards, or every fault that keeps the site from serving them: no
// manifest, one that does not read, a file it names that is missing or of no card's type. A file the manifest names
// not is never read, so the site serves nothing of the folder but the cards.
func LoadCards(folder fs.FS) (*Cards, error) {
	names, err := readJSON[CardNames](folder, CardsManifest, errCards)
	if err != nil {
		return nil, err
	}
	files := map[string]File{}
	var faults []error
	for _, name := range slices.Compact(slices.Sorted(maps.Values(names))) {
		file, err := readCard(folder, name)
		if err != nil {
			faults = append(faults, err)
			continue
		}
		files[name] = file
	}
	if err := errors.Join(faults...); err != nil {
		return nil, err
	}
	return &Cards{names: names, files: files}, nil
}

// readCard reads a card's file by its name; a fault where it is missing or of no card's type.
func readCard(folder fs.FS, name string) (File, error) {
	if path.Ext(name) != cardEnding {
		return File{ContentType: "", Body: nil}, fmt.Errorf("%w: %s names %q, which is no %s", errCards, CardsManifest, name, cardEnding)
	}
	body, err := fs.ReadFile(folder, name)
	if err != nil {
		return File{ContentType: "", Body: nil}, fmt.Errorf(unreadFile, errCards, name, err)
	}
	return File{ContentType: CardType, Body: body}, nil
}

// NameOf is the name of the file of a page's card, by the path the page is served at, and whether the page has one.
func (cards *Cards) NameOf(page string) (string, bool) {
	name, found := cards.names[page]
	return name, found
}

// Files are every card's file, by its name.
func (cards *Cards) Files() iter.Seq2[string, File] {
	return maps.All(cards.files)
}
