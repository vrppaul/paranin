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

// Derived is a value worked out from the store's snapshots, once for each snapshot a reader meets, and kept until
// the store holds another: what a host draws from the content, which no request or session works out again.
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
