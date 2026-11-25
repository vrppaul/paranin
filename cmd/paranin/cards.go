package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/vrppaul/paranin/internal/assets"
	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/face/card"
	"github.com/vrppaul/paranin/internal/release"
)

// How cards is written after its command: the flag that draws again as the content changes, then the content folder
// and the cards' folder, by their places among the words past the flags
const (
	watchFlag      = "watch"
	watchUsage     = "draw the cards again each time the content changes, until stopped"
	cardsPlace     = 1
	cardsArguments = 2
)

// The program the cards are drawn with, found on the path
const typstProgram = "typst"

// How the cards' manifest is written, for a person to read: each page on a line of its own
const manifestIndent = "  "

// The key of how many cards a drawing drew, which it logs
const drawnKey = "drawn"

// What a cards' folder is refused for: holding a file no drawing of the cards writes, which a drawing would take away
var errForeignFile = errors.New("the cards' folder holds what is no card")

// cardsAsked is what cards is asked to draw: the content folder the cards show, the folder they are written to, and
// whether to draw them again as the content changes.
type cardsAsked struct {
	content string
	cards   string
	watch   bool
}

// cardsAskedOf reads what cards is asked from the words after its command, and whether they are written right. A
// mistake is told of by the usage alone, as for any other command.
func cardsAskedOf(words []string) (cardsAsked, bool) {
	flags := flag.NewFlagSet(cardsCommand, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	watch := flags.Bool(watchFlag, false, watchUsage)
	if err := flags.Parse(words); err != nil || flags.NArg() != cardsArguments {
		return cardsAsked{content: "", cards: "", watch: false}, false
	}
	return cardsAsked{content: flags.Arg(contentPlace), cards: flags.Arg(cardsPlace), watch: *watch}, true
}

// cardDrawing draws the cards of a content folder's pages into the cards' folder, with Typst.
type cardDrawing struct {
	drawer  card.Drawer
	content string
	cards   string
}

// drawCards draws the cards asked for once, or, watching, again each time the content changes, until stopped.
func drawCards(asked cardsAsked) int {
	stopping, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	drawer, err := card.NewDrawer(stopping, typstProgram, card.FacesFolder)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return failed
	}
	drawing := cardDrawing{drawer: drawer, content: asked.content, cards: asked.cards}
	if !asked.watch {
		_, err := drawing.draw(stopping, time.Now())
		return checked(err, asked.content, asked.cards)
	}
	// the watch looks before the first drawing, so a change made while it draws is a change it sees
	changes := release.Watch(stopping, []string{asked.content}, lookInterval)
	drawing.watch(stopping, changes, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	return succeeded
}

// watch draws the cards now, then again at each change of the content until the context ends, logging each drawing; a
// drawing with faults leaves the cards' folder as it was.
func (drawing cardDrawing) watch(stopping context.Context, changes <-chan struct{}, logger *slog.Logger) {
	for {
		drawn, err := drawing.draw(stopping, time.Now())
		if err != nil {
			logger.Error("the cards are not drawn", faultsKey, err.Error())
		} else {
			logger.Info("the cards are drawn", cardsKey, drawing.cards, drawnKey, drawn)
		}
		select {
		case <-stopping.Done():
			return
		case <-changes:
		}
	}
}

// draw draws the cards of every page that shows a preview from a moment on, as the content stands, and writes them to
// the cards' folder whole, and is how many it drew; nothing is written where the content has faults or a card cannot
// be drawn.
func (drawing cardDrawing) draw(ctx context.Context, moment time.Time) (int, error) {
	loaded, err := release.ReadFolder(drawing.content, content.Load)
	if err == nil {
		err = compose.Check(loaded, moment)
	}
	if err != nil {
		return 0, err //nolint:wrapcheck // the content's faults, printed one a line as they come
	}
	drawn := []drawnCard{}
	for _, owned := range settledCards(loaded, moment) {
		picture, err := drawing.drawer.Draw(ctx, owned.Facts)
		if err != nil {
			return 0, fmt.Errorf("draw the card of %q: %w", owned.Path, err)
		}
		drawn = append(drawn, drawnCard{path: owned.Path, picture: picture})
	}
	return len(drawn), writeCards(drawing.cards, drawn)
}

// settledCards are the cards of every page a preview shows from a moment on, now or as a post dated later comes out:
// those of the site the public reads once the last such post is out, which holds every page it reads before (ADR-0005).
// A draft has none.
func settledCards(loaded *content.Snapshot, moment time.Time) []card.Card {
	audience := compose.Audience{At: compose.LastChange(loaded, moment), Drafts: false}
	return card.CardsOf(compose.SiteOf(compose.Prepare(loaded), audience))
}

// drawnCard is a page's card as drawn: the path its page is served at, and the picture.
type drawnCard struct {
	path    string
	picture []byte
}

// writeCards writes cards into their folder, made where it is missing, as writeCardsIn has it.
func writeCards(folder string, drawn []drawnCard) error {
	if err := os.MkdirAll(folder, writtenFolderMode); err != nil {
		return fmt.Errorf("make the cards' folder %q: %w", folder, err)
	}
	root, err := os.OpenRoot(folder)
	if err != nil {
		return fmt.Errorf("open the cards' folder %q: %w", folder, err)
	}
	return errors.Join(writeCardsIn(root, drawn), root.Close())
}

// writeCardsIn writes cards in a root: each card's file by what it holds, then the manifest naming them, each written
// beside its place and renamed into it whole, so a reader meets a manifest whose cards are all there, the one before
// or this one; then it takes away every card the manifest names no more. A root that holds any file a drawing does not
// write is refused before anything is written, so a folder named by mistake, a build's or the repository's, loses
// nothing.
func writeCardsIn(root *os.Root, drawn []drawnCard) error {
	held, err := heldFiles(root)
	if err != nil {
		return err
	}
	names := assets.CardNames{}
	for _, written := range drawn {
		name := assets.CardFile(written.picture)
		names[written.path] = name
		if err := writeChanged(root, name, written.picture); err != nil {
			return err
		}
	}
	manifest, err := json.MarshalIndent(names, "", manifestIndent)
	if err != nil {
		return fmt.Errorf("write the cards' manifest: %w", err)
	}
	if err := writeChanged(root, assets.CardsManifest, manifest); err != nil {
		return err
	}
	return removeStale(root, held, names)
}

// heldFiles are the files a cards' folder holds, each one a drawing writes; a fault where it holds anything else.
func heldFiles(root *os.Root) ([]string, error) {
	entries, err := fs.ReadDir(root.FS(), ".")
	if err != nil {
		return nil, fmt.Errorf("list the cards' folder %q: %w", root.Name(), err)
	}
	held := []string{}
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !isDrawingsFile(entry.Name()) {
			return nil, fmt.Errorf("%w: %q holds %q; name a folder of the cards alone", errForeignFile, root.Name(), entry.Name())
		}
		held = append(held, entry.Name())
	}
	return held, nil
}

// isDrawingsFile is whether a file's name is one a drawing of the cards writes: the manifest or a card, or either left
// beside its place by a drawing that stopped before renaming it.
func isDrawingsFile(name string) bool {
	written, _ := strings.CutSuffix(name, replacementEnding)
	return written == assets.CardsManifest || assets.IsCardFile(written)
}

// writeChanged writes a file of a root whole where it does not hold those bytes already, so a drawing that changes no
// card touches no file, and a site watching the folder loads nothing again.
func writeChanged(root *os.Root, name string, text []byte) error {
	if held, err := root.ReadFile(name); err == nil && bytes.Equal(held, text) {
		return nil
	}
	return replaceIn(root, name, text, writtenFileMode)
}

// removeStale takes away the files a cards' folder held that its manifest names no more: cards no page shows, and any
// left beside its place by a drawing that stopped.
func removeStale(root *os.Root, held []string, names assets.CardNames) error {
	kept := map[string]bool{assets.CardsManifest: true}
	for _, name := range names {
		kept[name] = true
	}
	var faults []error
	for _, name := range slices.DeleteFunc(held, func(name string) bool { return kept[name] }) {
		// a file left beside its place may have been written again and renamed into it since
		if err := root.Remove(name); err != nil && !errors.Is(err, fs.ErrNotExist) {
			faults = append(faults, err)
		}
	}
	return errors.Join(faults...)
}

// releaseChecks are what a release is held to beyond each part's own load: its content to the site's own check, and
// its cards to one for every page a preview shows, now or as a post dated later comes out, each from the moment it is
// loaded on.
func releaseChecks() release.Checks {
	return release.Checks{Content: checkedNow, Cards: cardsCheckedNow}
}

// cardsCheckedNow holds a release's cards to one for every page a preview shows from the moment they are loaded on.
//
//wrapper:reads the clock as each load is checked, which the check takes as a value
func cardsCheckedNow(loaded *content.Snapshot, cards *assets.Cards) error {
	return missingCardsOf(loaded, cards, time.Now())
}

// missingCards are the pages a preview shows, by the path each is served at, whose card a release lacks.
type missingCards []string

// How a page with no card is refused, the page quoted, and the faults' separator
const (
	missingCardProblem = "%q has no card: paranin cards draws them"
	missingCardJoint   = "\n"
)

func (missing missingCards) Error() string {
	lines := make([]string, 0, len(missing))
	for _, path := range missing {
		lines = append(lines, fmt.Sprintf(missingCardProblem, path))
	}
	return strings.Join(lines, missingCardJoint)
}

// missingCardsOf are the pages a preview shows from a moment on whose card a release's cards lack; none where they hold
// every one.
func missingCardsOf(loaded *content.Snapshot, cards *assets.Cards, moment time.Time) error {
	missing := missingCards{}
	for _, owned := range settledCards(loaded, moment) {
		if _, found := cards.NameOf(owned.Path); !found {
			missing = append(missing, owned.Path)
		}
	}
	if len(missing) > 0 {
		return missing
	}
	return nil
}
