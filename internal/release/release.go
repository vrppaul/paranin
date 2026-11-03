// Package release holds what the site serves: a release's content and its front-end build, loaded together and held
// as one snapshot that nothing changes after, which a store swaps whole (docs/ARCHITECTURE.md, ADR-0003, ADR-0004).
// A page names the files of the build it was drawn with, so the two parts are never swapped apart: a load that
// fails in either keeps both as they were. What a host draws from a snapshot is worked out once for each
// (Derived).
package release

import (
	"errors"
	"io/fs"
	"os"
	"sync"
	"sync/atomic"

	"github.com/vrppaul/paranin/internal/assets"
	"github.com/vrppaul/paranin/internal/content"
)

// Folders are where a release is read from: its content, and its front-end build.
type Folders struct {
	Content fs.FS
	Build   fs.FS
}

// FoldersAt are the folders of a release at their paths on disk.
func FoldersAt(contentPath, buildPath string) Folders {
	return Folders{Content: os.DirFS(contentPath), Build: os.DirFS(buildPath)}
}

// Snapshot is a release as one load read it.
type Snapshot struct {
	Content *content.Snapshot
	Build   *assets.Build
}

// Load reads both parts of a release and checks each whole: a snapshot, or every fault of either part.
func Load(folders Folders) (*Snapshot, error) {
	loaded, contentErr := content.Load(folders.Content)
	build, buildErr := assets.Load(folders.Build)
	if err := errors.Join(contentErr, buildErr); err != nil {
		return nil, err
	}
	return &Snapshot{Content: loaded, Build: build}, nil
}

// Store holds the release a running site serves. Each load makes a whole snapshot and swaps it in at once, so a
// reader sees the old one or the new one, never one half loaded; a load that fails leaves the old in place.
type Store struct {
	folders Folders
	current atomic.Pointer[Snapshot]
	// reloading lets one load run at a time, so an older read never lands after a newer one
	reloading sync.Mutex
}

// Open loads the folders for the first time; a site cannot start on a release that does not load.
func Open(folders Folders) (*Store, error) {
	store := &Store{folders: folders, current: atomic.Pointer[Snapshot]{}, reloading: sync.Mutex{}}
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
	snapshot, err := Load(store.folders)
	if err != nil {
		return err
	}
	store.current.Store(snapshot)
	return nil
}

// Derived is a value worked out from the store's snapshots, once for each snapshot a reader meets, and kept until
// the store holds another: what a host draws from a release, which no request or session works out again.
type Derived[Value any] struct {
	store *Store
	work  func(*Snapshot) Value
	// last is the value of the last snapshot a reader met
	last atomic.Pointer[derivation[Value]]
}

// derivation is a value and the snapshot it was worked out from.
type derivation[Value any] struct {
	snapshot *Snapshot
	value    Value
}

// Derive is a value the store's snapshots make, by the work that makes it.
func Derive[Value any](store *Store, work func(*Snapshot) Value) *Derived[Value] {
	return &Derived[Value]{store: store, work: work, last: atomic.Pointer[derivation[Value]]{}}
}

// Current is the value of the store's current snapshot. Two readers that meet a new snapshot at once may each work
// it out, and a value is kept only over the one its reader found. The kept value is read before the snapshot, and
// a store's snapshots only move forward, so the snapshot worked out is never older than the value it replaces.
func (derived *Derived[Value]) Current() Value {
	found := derived.last.Load()
	snapshot := derived.store.Current()
	if found != nil && found.snapshot == snapshot {
		return found.value
	}
	worked := &derivation[Value]{snapshot: snapshot, value: derived.work(snapshot)}
	derived.last.CompareAndSwap(found, worked)
	return worked.value
}
