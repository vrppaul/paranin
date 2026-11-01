package content

import (
	"io/fs"
	"sync"
	"sync/atomic"
)

// Store holds the content a running site serves. Each load makes a whole snapshot and swaps it in at once, so a
// reader sees the old one or the new one, never one half loaded; a load that fails leaves the old in place.
type Store struct {
	folder  fs.FS
	current atomic.Pointer[Snapshot]
	// reloading lets one load run at a time, so an older read never lands after a newer one
	reloading sync.Mutex
}

// Open loads the folder for the first time; a site cannot start on content that does not load.
func Open(folder fs.FS) (*Store, error) {
	store := &Store{folder: folder, current: atomic.Pointer[Snapshot]{}, reloading: sync.Mutex{}}
	if err := store.Reload(); err != nil {
		return nil, err
	}
	return store, nil
}

// Current is the snapshot loaded last.
func (store *Store) Current() *Snapshot {
	return store.current.Load()
}

// Reload loads the folder again and swaps the new snapshot in. On a fault it keeps the one it has and hands
// back the faults, for whoever asked to log.
func (store *Store) Reload() error {
	store.reloading.Lock()
	defer store.reloading.Unlock()
	snapshot, err := Load(store.folder)
	if err != nil {
		return err
	}
	store.current.Store(snapshot)
	return nil
}
