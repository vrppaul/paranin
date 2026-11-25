package release

import (
	"context"
	"io/fs"
	"path"
	"slices"
	"strings"
	"time"
)

// What starts a hidden name, which an editor gives its swap and lock files beside the file it edits: no load reads
// one, so a look passes over it
const hiddenMark = "."

// stamp is a file of a folder as a look sees it: its path there, and when it was last written. A file added, taken
// away or renamed changes the paths a look finds; one written again, its time. So a write is seen as long as the
// file system's times are finer than the interval between two looks: on one that keeps a time to the second, a
// write within the second of one already taken waits for the next change.
type stamp struct {
	path    string
	written int64
}

// sight is what a look sees of folders: the stamps of each folder's files, apart, in the folders' order.
type sight [][]stamp

// Watch sends on the channel it returns each time folders change and then hold still for one look, looking at them at
// every interval until the context ends (ADR-0003): a release's, for a site that loads it again, or the content alone,
// for the cards drawn from it. It takes its first look before it returns, so a change made after that is seen. A change
// still going on at a look waits for the next: an editor's save or a build that writes many files makes one change,
// taken at most two looks after its last write. Changes that pile up before the receiver takes one are sent as one.
func Watch(ctx context.Context, folderPaths []string, interval time.Duration) <-chan struct{} {
	changes := make(chan struct{}, 1)
	go watch(ctx, folderPaths, interval, look(folderPaths), changes)
	return changes
}

// watch looks at the folders at every interval until the context ends, and sends a change once a look sees what
// the look before saw, and that is not what was last taken.
func watch(ctx context.Context, folderPaths []string, interval time.Duration, taken sight, changes chan<- struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	previous := taken
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		current := look(folderPaths)
		if slices.EqualFunc(current, previous, slices.Equal[[]stamp]) && !slices.EqualFunc(current, taken, slices.Equal[[]stamp]) {
			taken = current
			SendUnlessWaiting(changes)
		}
		previous = current
	}
}

// SendUnlessWaiting wakes whoever reads a channel of one place, never waiting on them: where a wake already waits to
// be taken, it stands for this one too, so sends that pile up are taken as one.
func SendUnlessWaiting(wakes chan<- struct{}) {
	select {
	case wakes <- struct{}{}:
	default:
	}
}

// look is what every folder holds, each file stamped, in the order a walk finds them.
func look(folderPaths []string) sight {
	seen := make(sight, 0, len(folderPaths))
	for _, folderPath := range folderPaths {
		seen = append(seen, stamped(folderPath))
	}
	return seen
}

// stamped is every file of a folder on disk but the hidden ones, stamped, read as a load reads it (ReadFolder). A file
// that goes while the walk passes it is left out, and a folder that cannot be opened looks empty: the look after sees
// the folder as it is then.
func stamped(folderPath string) []stamp {
	// the walk reports no error, so the read hands back none but the folder's, which the look leaves to the next
	stamps, _ := ReadFolder(folderPath, walkedStamps)
	return stamps
}

// walkedStamps are the stamps of every file a walk of a folder finds. A walk that cannot read an entry passes over it,
// so it has no error to report.
func walkedStamps(folder fs.FS) ([]stamp, error) {
	gathered := &collector{stamps: nil}
	_ = fs.WalkDir(folder, ".", gathered.visit)
	return gathered.stamps, nil
}

// collector gathers the stamps of the files a walk finds.
type collector struct {
	stamps []stamp
}

// visit stamps an entry a walk found where it is a file that can be read and its name is not hidden.
func (gathered *collector) visit(name string, entry fs.DirEntry, err error) error {
	if err != nil || !entry.Type().IsRegular() || strings.HasPrefix(path.Base(name), hiddenMark) {
		return nil
	}
	info, err := entry.Info()
	if err != nil {
		return nil
	}
	gathered.stamps = append(gathered.stamps, stamp{path: name, written: info.ModTime().UnixNano()})
	return nil
}
