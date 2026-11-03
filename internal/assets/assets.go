// Package assets reads the page's front-end build (docs/ARCHITECTURE.md, ADR-0004, D22): the files one Vite build
// writes, each with a hash in its name, and the manifest naming them, from a folder on disk. A load reads the file
// each chunk of the manifest was built into, a CSS entry's all it brings, and checks that the manifest holds every
// entry the site names, so a release's files are held whole beside its content and a page drawn from it names
// only files it holds. It knows no face: a face asks it what an entry brings to a page.
package assets

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"iter"
	"maps"
	"path"
)

// Entry is a source the build takes in, by the path from web/ the manifest names it with.
type Entry string

// The entries the site names, each written once here and once among the inputs of web/vite.config.ts
const (
	// Stylesheet is the page's stylesheet, which every page includes
	Stylesheet Entry = "src/ui/ui.css"
)

// named is every entry the site names, which a build must hold.
func named() []Entry {
	return []Entry{Stylesheet}
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
	}
}

// The faults a build can hold
var errBuild = errors.New("the build has a fault")

// File is a built file as the site serves it: the type it is sent as, from its name, and what it holds.
type File struct {
	ContentType string
	Body        []byte
}

// Build is one build, read whole: each file by the path the site serves it at, and what each entry the site
// names brings to a page.
type Build struct {
	files       map[string]File
	stylesheets map[Entry][]string
}

// chunk is one chunk of the manifest, as Vite writes it: the file it was built into, from the build's root.
type chunk struct {
	File string `json:"file"`
}

// Load reads a build from a folder: a build, or every fault that keeps the site from being served from it.
func Load(folder fs.FS) (*Build, error) {
	manifest, err := readManifest(folder)
	if err != nil {
		return nil, err
	}
	files, filesErr := readFiles(folder, manifest)
	if err := errors.Join(filesErr, missingEntries(manifest)); err != nil {
		return nil, err
	}
	stylesheets := make(map[Entry][]string, len(named()))
	for _, entry := range named() {
		stylesheets[entry] = []string{servedPath(manifest[string(entry)].File)}
	}
	return &Build{files: files, stylesheets: stylesheets}, nil
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

// readFiles reads the file of every chunk of a manifest, by the path the site serves it at, and the faults of
// those that are missing or of a type the site serves none of.
func readFiles(folder fs.FS, manifest map[string]chunk) (map[string]File, error) {
	types := contentTypes()
	files := make(map[string]File, len(manifest))
	var faults []error
	for _, built := range manifest {
		contentType, known := types[path.Ext(built.File)]
		if !known {
			faults = append(faults, fmt.Errorf("%w: %q is of no type the site serves", errBuild, built.File))
			continue
		}
		body, err := fs.ReadFile(folder, built.File)
		if err != nil {
			faults = append(faults, fmt.Errorf("%w: read %q: %w", errBuild, built.File, err))
			continue
		}
		files[servedPath(built.File)] = File{ContentType: contentType, Body: body}
	}
	return files, errors.Join(faults...)
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

// servedPath is the path the site serves a built file at.
func servedPath(file string) string {
	return servedRoot + file
}

// Files are every built file, by the path the site serves it at.
func (build *Build) Files() iter.Seq2[string, File] {
	return maps.All(build.files)
}

// Stylesheets are the paths of the stylesheets an entry the site names brings to a page.
func (build *Build) Stylesheets(entry Entry) []string {
	return build.stylesheets[entry]
}
