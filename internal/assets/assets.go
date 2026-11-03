// Package assets reads the page's front-end build (docs/ARCHITECTURE.md, ADR-0004, D22): the files one Vite build
// writes, each with a hash in its name, and the manifest naming them, from a folder on disk. A load reads every
// file the manifest names, each chunk's script and the stylesheets it imports, and checks that the manifest holds
// every entry the site names and every chunk one imports, so a release's files are held whole beside its content
// and a page drawn from it names only files it holds. What an entry includes in a page is worked out once, at the
// load, as Vite's guide to backend integration walks a manifest. It knows no face: a face asks it what an entry
// includes in a page.
package assets

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"iter"
	"maps"
	"path"
	"slices"
)

// Entry is a source the build takes in, by the path from web/ the manifest names it with.
type Entry string

// The entries the site names, each written once here and once among the inputs of web/vite.config.ts
const (
	// Site is the script every page runs, and the stylesheet it imports, which every page is drawn in
	Site Entry = "src/entries/site.ts"
)

// named is every entry the site names, which a build must hold.
func named() []Entry {
	return []Entry{Site}
}

// The manifest, at the build's root, and the root of the site a built file is served under: Vite's base, which it
// writes into each address it puts in a file, as one built file names another
const (
	manifestFile = "manifest.json"
	servedRoot   = "/"
)

// contentTypes are the types the site serves a built file as, by its name's ending: the same on every machine,
// whatever its own tables say. Each comes with the first file of its kind the build writes.
func contentTypes() map[string]string {
	return map[string]string{
		".css": "text/css; charset=utf-8",
		".js":  "text/javascript; charset=utf-8",
	}
}

// The faults a build can hold
var errBuild = errors.New("the build has a fault")

// File is a built file as the site serves it: the type it is sent as, from its name, and what it holds.
type File struct {
	ContentType string
	Body        []byte
}

// Included is what a page includes to run an entry, each by the path the site serves it at, in the order Vite's
// guide lays them in a page's head: the stylesheets of the entry and of every chunk it reaches, the entry's
// script, then the script of every chunk it reaches, which the page preloads so none waits for the one before.
type Included struct {
	Stylesheets []string
	Script      string
	Preloads    []string
}

// Build is one build, read whole: each file by the path the site serves it at, and what each entry the site
// names includes in a page.
type Build struct {
	files    map[string]File
	included map[Entry]Included
}

// chunk is one chunk of the manifest, as Vite writes it: the file it was built into and the stylesheets it
// imports, each from the build's root, and the chunks it imports, by their keys in the manifest.
type chunk struct {
	File    string   `json:"file"`
	CSS     []string `json:"css"`
	Imports []string `json:"imports"`
}

// Load reads a build from a folder: a build, or every fault that keeps the site from being served from it.
func Load(folder fs.FS) (*Build, error) {
	manifest, err := readManifest(folder)
	if err != nil {
		return nil, err
	}
	files, filesErr := readFiles(folder, manifest)
	if err := errors.Join(filesErr, missingEntries(manifest), missingImports(manifest)); err != nil {
		return nil, err
	}
	included := make(map[Entry]Included, len(named()))
	for _, entry := range named() {
		included[entry] = includedBy(manifest, manifest[string(entry)])
	}
	return &Build{files: files, included: included}, nil
}

// readManifest is the manifest of a build: each chunk by its key, an entry's the source it was built from.
func readManifest(folder fs.FS) (map[string]chunk, error) {
	var manifest map[string]chunk
	written, err := fs.ReadFile(folder, manifestFile)
	if err == nil {
		err = json.Unmarshal(written, &manifest)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: read %s: %w", errBuild, manifestFile, err)
	}
	return manifest, nil
}

// readFiles reads every file a manifest names, by the path the site serves it at, and the faults of those that are
// missing or of a type the site serves none of.
func readFiles(folder fs.FS, manifest map[string]chunk) (map[string]File, error) {
	types := contentTypes()
	names := builtFiles(manifest)
	files := make(map[string]File, len(names))
	var faults []error
	for _, name := range names {
		contentType, known := types[path.Ext(name)]
		if !known {
			faults = append(faults, fmt.Errorf("%w: %q is of no type the site serves", errBuild, name))
			continue
		}
		body, err := fs.ReadFile(folder, name)
		if err != nil {
			faults = append(faults, fmt.Errorf("%w: read %q: %w", errBuild, name, err))
			continue
		}
		files[servedPath(name)] = File{ContentType: contentType, Body: body}
	}
	return files, errors.Join(faults...)
}

// builtFiles are the files a manifest names, each once and in order: every chunk's own, and the stylesheets it
// imports, which the manifest names under no key of their own.
func builtFiles(manifest map[string]chunk) []string {
	names := make(map[string]struct{}, len(manifest))
	for _, built := range manifest {
		names[built.File] = struct{}{}
		for _, stylesheet := range built.CSS {
			names[stylesheet] = struct{}{}
		}
	}
	return slices.Sorted(maps.Keys(names))
}

// missingEntries are the faults of the entries the site names that a manifest lacks.
func missingEntries(manifest map[string]chunk) error {
	var faults []error
	for _, entry := range named() {
		if _, found := manifest[string(entry)]; !found {
			faults = append(faults, fmt.Errorf("%w: %s names no entry %q", errBuild, manifestFile, entry))
		}
	}
	return errors.Join(faults...)
}

// missingImports are the faults of the chunks a manifest's chunks import that it lacks, by the importer's key.
func missingImports(manifest map[string]chunk) error {
	var faults []error
	for _, key := range slices.Sorted(maps.Keys(manifest)) {
		for _, imported := range manifest[key].Imports {
			if _, found := manifest[imported]; !found {
				faults = append(faults, fmt.Errorf("%w: %s names no chunk %q, which %q imports", errBuild, manifestFile, imported, key))
			}
		}
	}
	return errors.Join(faults...)
}

// includedBy is what a page includes to run an entry's chunk of a manifest that holds every chunk it imports.
func includedBy(manifest map[string]chunk, entry chunk) Included {
	reached := reachedChunks(manifest, entry.Imports, map[string]bool{})
	stylesheets := servedPaths(entry.CSS)
	preloads := make([]string, 0, len(reached))
	for _, imported := range reached {
		stylesheets = append(stylesheets, servedPaths(imported.CSS)...)
		preloads = append(preloads, servedPath(imported.File))
	}
	return Included{Stylesheets: stylesheets, Script: servedPath(entry.File), Preloads: preloads}
}

// reachedChunks are the chunks a list of imports reaches, each once, every chunk after the chunks it imports, as
// Vite's guide walks them; seen holds the keys met so far, so a chunk two others import is taken once.
func reachedChunks(manifest map[string]chunk, imports []string, seen map[string]bool) []chunk {
	var reached []chunk
	for _, key := range imports {
		if seen[key] {
			continue
		}
		seen[key] = true
		imported := manifest[key]
		reached = append(reached, reachedChunks(manifest, imported.Imports, seen)...)
		reached = append(reached, imported)
	}
	return reached
}

// servedPath is the path the site serves a built file at.
func servedPath(file string) string {
	return servedRoot + file
}

// servedPaths are the paths the site serves built files at, in their order.
func servedPaths(files []string) []string {
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, servedPath(file))
	}
	return paths
}

// Files are every built file, by the path the site serves it at.
func (build *Build) Files() iter.Seq2[string, File] {
	return maps.All(build.files)
}

// Includes is what a page includes to run an entry the site names.
func (build *Build) Includes(entry Entry) Included {
	return build.included[entry]
}
