// Package assets reads what the build makes for the site (docs/ARCHITECTURE.md, ADR-0004, ADR-0005, D22), each from a
// folder on disk, whole: the page's front-end build, so a release holds every file its pages name and every widget's
// code and props' schema its posts need, and the pages' cards, by the manifest their drawing writes (cards.go). It
// knows no face: a face asks it what an entry includes in a page.
package assets

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"iter"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
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

// The widgets' props' schemas beside the manifest (web/plugins/widgetprops.ts), never served, and each one's address
const (
	propsSchemaFile     = "props.schema.json"
	widgetSchemaAddress = "urn:paranin:widget:"
)

// The folder of web/ every widget's component stands in, named for the whole site, and the ending of a component's
// file. The site's entry imports each lazily by a glob of the folder (web/src/entries/site.ts), so the manifest names
// a widget's code by its component's path, a chunk of its own
const (
	widgetsFolder   = "src/widgets"
	componentEnding = ".svelte"
)

// contentTypes are the types the site serves a built file as, by its name's ending: the same on every machine,
// whatever its own tables say. Each comes with the first file of its kind the build writes.
func contentTypes() map[string]string {
	return map[string]string{
		".css":   "text/css; charset=utf-8",
		".js":    "text/javascript; charset=utf-8",
		".woff2": "font/woff2",
	}
}

// The faults a build can hold
var errBuild = errors.New("the build has a fault")

// How a file a manifest names that cannot be read is refused: its folder's kind of fault, the file, and why
const unreadFile = "%w: read %q: %w"

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

// Build is one build, read whole: each file by the path the site serves it at, what each entry the site names
// includes in a page, and the props' schema of every widget it holds the code of, by its component's name.
type Build struct {
	files    map[string]File
	included map[Entry]Included
	widgets  map[string]*jsonschema.Schema
}

// PropsFault is a widget's prop its schema refuses: where it stands in the props, as a JSON Pointer, which is empty for
// the props as a whole, and what the schema says of it.
type PropsFault struct {
	Prop string
	Rule string
}

// chunk is one key of the manifest, as Vite writes it: the file it was built into and the stylesheets it imports,
// each from the build's root, and the chunks it imports, by their keys in the manifest. A file a stylesheet loads has
// a key of its own, its file alone.
type chunk struct {
	File    string   `json:"file"`
	CSS     []string `json:"css"`
	Imports []string `json:"imports"`
}

// Load reads a build from a folder: a build, or every fault that keeps the site from being served from it.
func Load(folder fs.FS) (*Build, error) {
	// the manifest: each chunk by its key, an entry's the source it was built from
	manifest, err := readJSON[map[string]chunk](folder, manifestFile, errBuild)
	if err != nil {
		return nil, err
	}
	files, filesErr := readFiles(folder, manifest)
	widgets, widgetsErr := readWidgets(folder, manifest)
	if err := errors.Join(filesErr, widgetsErr, missingEntries(manifest), missingImports(manifest)); err != nil {
		return nil, err
	}
	included := make(map[Entry]Included, len(named()))
	for _, entry := range named() {
		included[entry] = includedBy(manifest, manifest[string(entry)])
	}
	return &Build{files: files, included: included, widgets: widgets}, nil
}

// WidgetSource is the path from web/ of a widget's component, by its name: the manifest's key of the widget's code.
func WidgetSource(name string) string {
	return path.Join(widgetsFolder, name+componentEnding)
}

// readWidgets compiles the props' schema of every widget a manifest holds the code of, by its component's name, from
// the build's file of them; the faults of a file missing or that does not read, and of a widget whose schema it lacks
// or that does not compile. A schema reads no other schema but JSON Schema's own, which the validator holds: none
// from disk, which would read past the folder the build is read through.
func readWidgets(folder fs.FS, manifest map[string]chunk) (map[string]*jsonschema.Schema, error) {
	schemas, err := readJSON[map[string]json.RawMessage](folder, propsSchemaFile, errBuild)
	if err != nil {
		return nil, err
	}
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(jsonschema.SchemeURLLoader{})
	// a format valibot writes, an email or a date, is a rule the props are held to, as valibot holds them
	compiler.AssertFormat()
	widgets := map[string]*jsonschema.Schema{}
	var faults []error
	for _, name := range widgetNames(manifest) {
		schema, err := compiledSchema(compiler, name, schemas[name])
		if err != nil {
			faults = append(faults, err)
			continue
		}
		widgets[name] = schema
	}
	return widgets, errors.Join(faults...)
}

// widgetNames are the names of the widgets a manifest holds the code of, in their order: each a component of the
// widgets' folder, whose code is a chunk of its own.
func widgetNames(manifest map[string]chunk) []string {
	names := []string{}
	for _, key := range slices.Sorted(maps.Keys(manifest)) {
		if path.Dir(key) == widgetsFolder && path.Ext(key) == componentEnding {
			names = append(names, strings.TrimSuffix(path.Base(key), componentEnding))
		}
	}
	return names
}

// compiledSchema compiles a widget's props' schema, as the build's file of them writes it; the fault of a schema the
// file lacks, written as none, or of one that does not compile.
func compiledSchema(compiler *jsonschema.Compiler, name string, written json.RawMessage) (*jsonschema.Schema, error) {
	// the decoder would refuse a schema the file lacks as an end of input, which names no fault the author can mend
	if written == nil {
		return nil, fmt.Errorf("%w: %s holds no schema of the props of the widget %q", errBuild, propsSchemaFile, name)
	}
	address := widgetSchemaAddress + name
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(written))
	if err == nil {
		err = compiler.AddResource(address, document)
	}
	var schema *jsonschema.Schema
	if err == nil {
		schema, err = compiler.Compile(address)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %s holds a schema of the props of the widget %q that does not compile: %w", errBuild, propsSchemaFile, name, err)
	}
	return schema, nil
}

// readJSON is a file of JSON at a folder's root, a manifest or the widgets' props' schemas, read into its type; a
// fault of the folder's kind where it is missing or does not read.
func readJSON[Value any](folder fs.FS, file string, folderFault error) (Value, error) {
	var value Value
	written, err := fs.ReadFile(folder, file)
	if err == nil {
		err = json.Unmarshal(written, &value)
	}
	if err != nil {
		return value, fmt.Errorf("%w: read %s: %w", folderFault, file, err)
	}
	return value, nil
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
			faults = append(faults, fmt.Errorf(unreadFile, errBuild, name, err))
			continue
		}
		files[servedPath(name)] = File{ContentType: contentType, Body: body}
	}
	return files, errors.Join(faults...)
}

// builtFiles are the files a manifest names, each once and in order: every key's own, a chunk's or a file a
// stylesheet loads, and the stylesheets a chunk imports, which the manifest names under no key of their own.
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

// HoldsWidget is whether the build holds the code of a widget, by its component's name.
func (build *Build) HoldsWidget(name string) bool {
	_, held := build.widgets[name]
	return held
}

// RefusedProps are the faults a widget's schema finds in its props, a JSON object, each where the schema refuses a
// prop: none where it takes them all, or where the build holds no such widget, which HoldsWidget says.
func (build *Build) RefusedProps(name string, props json.RawMessage) []PropsFault {
	schema, held := build.widgets[name]
	if !held {
		return nil
	}
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(props))
	if err == nil {
		err = schema.Validate(instance)
	}
	var refused *jsonschema.ValidationError
	switch {
	case err == nil:
		return nil
	case errors.As(err, &refused):
		return refusedLeaves(*refused.DetailedOutput())
	default:
		return []PropsFault{{Prop: "", Rule: err.Error()}}
	}
}

// refusedLeaves are the faults a unit of the validator's output holds: its own where it is a leaf, the rule a prop
// breaks; or else those of each unit under it, a part of the schema the prop's value was held to.
func refusedLeaves(unit jsonschema.OutputUnit) []PropsFault {
	if unit.Error != nil {
		return []PropsFault{{Prop: unit.InstanceLocation, Rule: unit.Error.String()}}
	}
	faults := []PropsFault{}
	for _, cause := range unit.Errors {
		faults = append(faults, refusedLeaves(cause)...)
	}
	return faults
}
