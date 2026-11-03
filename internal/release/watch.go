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

// sight is what a look sees of a release: the stamps of each folder's files, apart.
type sight struct {
	content []stamp
	build   []stamp
}

// same is whether two looks saw the same.
func (seen sight) same(other sight) bool {
	return slices.Equal(seen.content, other.content) && slices.Equal(seen.build, other.build)
}

// Watch sends on the channel it returns each time the release's folders change and then hold still for one look,
// looking at them at every interval until the context ends (ADR-0003). It takes its first look before it returns,
// so a change made after that is seen. A change still going on at a look waits for the next: an editor's save or a
// build that writes many files makes one change, taken at most two looks after its last write. Changes that pile
// up before the receiver takes one are sent as one.
func Watch(ctx context.Context, folders Folders, interval time.Duration) <-chan struct{} {
	changes := make(chan struct{}, 1)
	go watch(ctx, folders, interval, look(folders), changes)
	return changes
}

// watch looks at the folders at every interval until the context ends, and sends a change once a look sees what
// the look before saw, and that is not what was last taken.
func watch(ctx context.Context, folders Folders, interval time.Duration, taken sight, changes chan<- struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	previous := taken
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		current := look(folders)
		if current.same(previous) && !current.same(taken) {
			taken = current
			sendOnce(changes)
		}
		previous = current
	}
}

// sendOnce sends a change unless one is already waiting to be taken.
func sendOnce(changes chan<- struct{}) {
	select {
	case changes <- struct{}{}:
	default:
	}
}

// look is what both folders hold, each file stamped, in the order a walk finds them.
func look(folders Folders) sight {
	return sight{content: stamped(folders.Content), build: stamped(folders.Build)}
}

// stamped is every file of a folder but the hidden ones, stamped. A file that goes while the walk passes it is left
// out: the look after sees the folder as it is then.
func stamped(folder fs.FS) []stamp {
	gathered := &collector{stamps: nil}
	// A walk that cannot read an entry passes over it, so its error has nothing to report
	_ = fs.WalkDir(folder, ".", gathered.visit)
	return gathered.stamps
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
