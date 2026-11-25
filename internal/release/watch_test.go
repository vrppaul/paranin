package release

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/synctest"
	"time"

	"github.com/vrppaul/paranin/internal/fixtures"
)

// How often a test's watch looks at the folders, the files of the content a test renames and hides beside, and the
// card it draws
const (
	lookInterval = time.Second
	wordsFile    = "i18n/en.yaml"
	renamedWords = "i18n/en.renamed.yaml"
	swapFile     = ".profile.yaml.swp"
	drawnCard    = "0123456789abcdef.png"
)

// A change in the folders is taken once it has held still for one look, and once only; folders that do not change
// send nothing, and neither does a hidden file, as an editor's swap file beside the file it edits.
func TestAChangeIsTakenOnceHeldStill(test *testing.T) {
	cases := []struct {
		name   string
		change func(test *testing.T, folders Folders)
		want   int
	}{
		{name: "a file written", change: writeAnotherAuthor, want: 1},
		{name: "a file renamed, its time kept", change: renameWords, want: 1},
		{name: "a hidden file written", change: writeSwap, want: 0},
		{name: "a card written", change: writeCard, want: 1},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			synctest.Test(test, func(test *testing.T) {
				folders := fixtureCopies(test)
				watching, stop := context.WithCancel(context.Background())
				changes := Watch(watching, folders.Paths(), lookInterval)
				time.Sleep(2 * lookInterval)
				synctest.Wait()
				unchanged := len(changes)
				testCase.change(test, folders)
				time.Sleep(lookInterval)
				synctest.Wait()
				stillMoving := len(changes)
				time.Sleep(lookInterval)
				synctest.Wait()
				settled := len(changes)
				for range settled {
					<-changes
				}
				time.Sleep(2 * lookInterval)
				synctest.Wait()
				again := len(changes)
				stop()
				synctest.Wait()
				if unchanged != 0 || stillMoving != 0 || settled != testCase.want || again != 0 {
					test.Errorf("changes sent: %d with nothing changed, %d a look after the change, %d once it held still, %d "+
						"after it was taken; want 0, 0, %d, 0", unchanged, stillMoving, settled, again, testCase.want)
				}
			})
		})
	}
}

// A change made as soon as the watch begins, before it has looked once more, is taken: its first look came before
// it began.
func TestAChangeMadeAsTheWatchBeginsIsTaken(test *testing.T) {
	synctest.Test(test, func(test *testing.T) {
		folders := fixtureCopies(test)
		watching, stop := context.WithCancel(context.Background())
		changes := Watch(watching, folders.Paths(), lookInterval)
		writeAnotherAuthor(test, folders)
		time.Sleep(2 * lookInterval)
		synctest.Wait()
		taken := len(changes)
		stop()
		synctest.Wait()
		if taken != 1 {
			test.Errorf("changes sent two looks after a change made as the watch began: %d, want 1", taken)
		}
	})
}

// A change still going on at every look, as a build writing its files, is taken once, after its last write.
func TestAChangeStillGoingOnIsTakenAfterItsLastWrite(test *testing.T) {
	synctest.Test(test, func(test *testing.T) {
		folders := fixtureCopies(test)
		watching, stop := context.WithCancel(context.Background())
		changes := Watch(watching, folders.Paths(), lookInterval)
		for _, write := range []struct{ original, replacement string }{
			{original: authorName, replacement: "name: Mira Other"},
			{original: "name: Mira Other", replacement: "name: Mira Third"},
			{original: "name: Mira Third", replacement: "name: Mira Fourth"},
		} {
			fixtures.ReplaceIn(test, folders.Content, profileFile, write.original, write.replacement)
			time.Sleep(lookInterval)
			synctest.Wait()
		}
		whileWriting := len(changes)
		time.Sleep(lookInterval)
		synctest.Wait()
		afterwards := len(changes)
		stop()
		synctest.Wait()
		if whileWriting != 0 || afterwards != 1 {
			test.Errorf("changes sent: %d while every look saw a write, %d a look after the last; want 0, 1", whileWriting, afterwards)
		}
	})
}

// writeAnotherAuthor writes the author's name otherwise.
//
//wrapper:a case of the table names the change it writes
func writeAnotherAuthor(test *testing.T, folders Folders) {
	fixtures.ReplaceIn(test, folders.Content, profileFile, authorName, "name: Mira Other")
}

// renameWords moves the word catalogue to another name, keeping the time it was written.
func renameWords(test *testing.T, folders Folders) {
	if err := os.Rename(filepath.Join(folders.Content, wordsFile), filepath.Join(folders.Content, renamedWords)); err != nil {
		test.Fatalf("rename %s: %v", wordsFile, err)
	}
}

// writeCard writes a card in the cards' folder, as their drawing does.
//
//wrapper:a case of the table names the change it writes
func writeCard(test *testing.T, folders Folders) {
	fixtures.WriteIn(test, folders.Cards, drawnCard, authorName)
}

// writeSwap writes a swap file beside the profile, as an editor keeps one.
//
//wrapper:a case of the table names the change it writes
func writeSwap(test *testing.T, folders Folders) {
	fixtures.WriteIn(test, folders.Content, swapFile, authorName)
}
