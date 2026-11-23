// Package release holds what the site serves: a release's content and its front-end build, loaded together and held
// as one snapshot that nothing changes after, which a store swaps whole (docs/ARCHITECTURE.md, ADR-0003, ADR-0004).
// A page names the files of the build it was drawn with, so the two parts are never swapped apart: a load that
// fails in either keeps both as they were. The content is held, beside its own load, to a check the composition root
// hands in from the parts over a release, which it knows nothing of. What a host draws from a snapshot is worked out
// once for each, and again once what it drew lapses (Derived). Watch looks at the folders in a goroutine of its own,
// until its context ends, and says when they changed and held still, for whoever reloads (watch.go). A folder on disk
// is read so no read leaves it (ReadFolder).
package release

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vrppaul/paranin/internal/assets"
	"github.com/vrppaul/paranin/internal/content"
)

// Folders are where a release is read from on disk, by their paths: its content, and its front-end build.
type Folders struct {
	Content string
	Build   string
}

// ReadFolder reads a folder on disk through one root (os.Root), opened for the read and closed after it: no read
// leaves the folder, by a link inside it among others, and every file of the read comes from the folder its path
// named as the read began, so a release switched in by a link while one is read is read whole at the next.
func ReadFolder[Value any](folderPath string, read func(fs.FS) (Value, error)) (Value, error) {
	root, err := os.OpenRoot(folderPath)
	if err != nil {
		var none Value
		return none, fmt.Errorf("open the folder %q: %w", folderPath, err)
	}
	value, err := read(root.FS())
	return value, errors.Join(err, root.Close())
}

// Snapshot is a release as one load read it.
type Snapshot struct {
	Content *content.Snapshot
	Build   *assets.Build
}

// Check is what the parts over a release hold its content to beyond its own load, which the composition root hands
// in: that no two of the pages it makes share an address, and that each link of a post inside the site leads to a page
// and, where it names one, a section of it. It hands back the faults it finds, none where it finds none.
type Check func(*content.Snapshot) error

// Load reads both parts of a release and checks each whole, the content by the check too: a snapshot, or every fault
// of either part.
func Load(folders Folders, check Check) (*Snapshot, error) {
	loaded, contentErr := ReadFolder(folders.Content, content.Load)
	if contentErr == nil {
		contentErr = check(loaded)
	}
	build, buildErr := ReadFolder(folders.Build, assets.Load)
	if err := errors.Join(contentErr, buildErr); err != nil {
		return nil, err
	}
	return &Snapshot{Content: loaded, Build: build}, nil
}

// Store holds the release a running site serves. Each load makes a whole snapshot and swaps it in at once, so a
// reader sees the old one or the new one, never one half loaded; a load that fails leaves the old in place.
type Store struct {
	folders Folders
	check   Check
	current atomic.Pointer[Snapshot]
	// reloading lets one load run at a time, so an older read never lands after a newer one
	reloading sync.Mutex
}

// Open loads the folders for the first time, its content held to the check at each load; a site cannot start on a
// release that does not load.
func Open(folders Folders, check Check) (*Store, error) {
	store := &Store{folders: folders, check: check, current: atomic.Pointer[Snapshot]{}, reloading: sync.Mutex{}}
	if err := store.Reload(); err != nil {
		return nil, err
	}
	return store, nil
}

// Current is the snapshot loaded last.
func (store *Store) Current() *Snapshot {
	return store.current.Load()
}

// Reload loads the folders again and swaps the new snapshot in. On a fault it keeps the one it has and hands back
// the faults, for whoever asked to log.
func (store *Store) Reload() error {
	store.reloading.Lock()
	defer store.reloading.Unlock()
	snapshot, err := Load(store.folders, store.check)
	if err != nil {
		return err
	}
	store.current.Store(snapshot)
	return nil
}

// Derived is a value worked out from the store's snapshots at a moment, once for each snapshot a reader meets, and
// kept until the store holds another or the moment its work says it lapses at comes: what a host draws from a release,
// which no request or session works out again.
type Derived[Value any] struct {
	store *Store
	work  Work[Value]
	// last is the value of the last snapshot a reader met
	last atomic.Pointer[derivation[Value]]
}

// Work works a value out from a snapshot at a moment, and says when it lapses: the zero time where it holds as long as
// its snapshot.
type Work[Value any] func(snapshot *Snapshot, moment time.Time) (Value, time.Time)

// derivation is a value, the snapshot it was worked out from and when it lapses, the zero time where it never does.
type derivation[Value any] struct {
	snapshot *Snapshot
	value    Value
	lapses   time.Time
}

// Derive is a value the store's snapshots make, by the work that makes it.
func Derive[Value any](store *Store, work Work[Value]) *Derived[Value] {
	return &Derived[Value]{store: store, work: work, last: atomic.Pointer[derivation[Value]]{}}
}

// Current is the value of the store's current snapshot at a moment.
//
//wrapper:names the snapshot a reader meets, the store's current one, which Of is handed
func (derived *Derived[Value]) Current(moment time.Time) Value {
	return derived.Of(derived.store.Current(), moment)
}

// Of is the value of a snapshot at a moment, the one kept where it holds: a value worked out from a snapshot another
// value is worked out from in turn reads the same snapshot, though the store swaps in another meanwhile. The value of
// the snapshot asked for replaces the one kept, an older snapshot's too, so readers of two snapshots at once may work a
// value out again; none is ever handed a value of another snapshot.
func (derived *Derived[Value]) Of(snapshot *Snapshot, moment time.Time) Value {
	found := derived.last.Load()
	if found.holds(snapshot, moment) {
		return found.value
	}
	value, lapses := derived.work(snapshot, moment)
	worked := &derivation[Value]{snapshot: snapshot, value: value, lapses: lapses}
	derived.last.CompareAndSwap(found, worked)
	return worked.value
}

// holds is whether a derivation, none where nothing was worked out yet, is the value of a snapshot at a moment: worked
// out from it, and not lapsed by then.
func (found *derivation[Value]) holds(snapshot *Snapshot, moment time.Time) bool {
	return found != nil && found.snapshot == snapshot && (found.lapses.IsZero() || moment.Before(found.lapses))
}
