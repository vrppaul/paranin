package assets_test

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/vrppaul/paranin/internal/assets"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// The fixture's files a test changes, from its root
const (
	manifestPath   = "manifest.json"
	stylesheetPath = "assets/site-fixture.css"
)

// The types the fixture's files are served as
const (
	styleType  = "text/css; charset=utf-8"
	scriptType = "text/javascript; charset=utf-8"
	fontType   = "font/woff2"
)

// Every file the manifest names, a chunk's script, the stylesheets it imports and a file a stylesheet loads, as the
// font, is served at its path from the site's root, as the type its name says, a widget's code, which the site's
// script loads by itself, among them; the manifest itself is no file of the site.
func TestABuildServesEveryFileItsManifestNames(test *testing.T) {
	build := fixtures.LoadedBuild(test)
	want := map[string]assets.File{
		"/assets/ReadingPace-fixture.js":  {ContentType: scriptType, Body: []byte("import \"./base-fixture.js\";\nexport default {};\n")},
		"/assets/ReadingPace-fixture.css": {ContentType: styleType, Body: []byte(".pace{margin:0}\n")},
		"/assets/site-fixture.js":         {ContentType: scriptType, Body: []byte("import \"./shared-fixture.js\";\nimport \"./base-fixture.js\";\n")},
		"/assets/shared-fixture.js":       {ContentType: scriptType, Body: []byte("import \"./base-fixture.js\";\n")},
		"/assets/base-fixture.js":         {ContentType: scriptType, Body: []byte("export {};\n")},
		"/" + stylesheetPath:              {ContentType: styleType, Body: []byte("body{margin:0}\n")},
		"/assets/shared-fixture.css":      {ContentType: styleType, Body: []byte("p{margin:0}\n")},
		"/assets/base-fixture.css":        {ContentType: styleType, Body: []byte("h1{margin:0}\n")},
		"/assets/font-fixture.woff2":      {ContentType: fontType, Body: []byte("wOF2\n")},
	}
	if got := maps.Collect(build.Files()); !maps.EqualFunc(got, want, sameFile) {
		test.Errorf("Files() = %q, want %q", got, want)
	}
}

// A page that runs the site's script includes what Vite's guide to backend integration lays in a page for it: the
// stylesheets of the entry and of every chunk it reaches, the entry's script, and every chunk it reaches to
// preload, each once, the deepest first.
func TestTheSiteEntryIncludesEveryChunkReached(test *testing.T) {
	build := fixtures.LoadedBuild(test)
	want := assets.Included{
		Stylesheets: []string{"/" + stylesheetPath, "/assets/base-fixture.css", "/assets/shared-fixture.css"},
		Script:      "/assets/site-fixture.js",
		Preloads:    []string{"/assets/base-fixture.js", "/assets/shared-fixture.js"},
	}
	if got := build.Includes(assets.Site); !sameIncluded(got, want) {
		test.Errorf("Includes(%q) = %q, want %q", assets.Site, got, want)
	}
}

// A build the site cannot be served from is refused, the fault named: no manifest, one that does not read, one
// lacking an entry the site names or a chunk one imports, a file it names missing or of a type the site serves
// none of.
func TestABuildNoSiteServesFromIsRefused(test *testing.T) {
	cases := []struct {
		name   string
		change func(test *testing.T, folder string)
		want   string
	}{
		{name: "no manifest", change: func(test *testing.T, folder string) { removeFile(test, folder, manifestPath) }, want: manifestPath},
		{
			name:   "a manifest that does not read",
			change: func(test *testing.T, folder string) { fixtures.ReplaceIn(test, folder, manifestPath, "{", "[") },
			want:   manifestPath,
		},
		{
			name: "an entry the site names missing",
			change: func(test *testing.T, folder string) {
				fixtures.ReplaceIn(test, folder, manifestPath, `"src/entries/site.ts": {`, `"src/entries/other.ts": {`)
			},
			want: `"src/entries/site.ts"`,
		},
		{
			name: "a chunk an entry imports missing",
			change: func(test *testing.T, folder string) {
				fixtures.ReplaceIn(test, folder, manifestPath, `"_base-fixture.js": {`, `"_other-fixture.js": {`)
			},
			want: `"_base-fixture.js"`,
		},
		{name: "a file missing", change: func(test *testing.T, folder string) { removeFile(test, folder, stylesheetPath) }, want: stylesheetPath},
		{
			name: "a file of no type the site serves",
			change: func(test *testing.T, folder string) {
				renameBuilt(test, folder, stylesheetPath, "assets/site-fixture.png")
			},
			want: `"assets/site-fixture.png"`,
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			folder := fixtures.Copy(test, fixtures.BuildFolder(test))
			testCase.change(test, folder)
			if _, err := assets.Load(os.DirFS(folder)); err == nil || !strings.Contains(err.Error(), testCase.want) {
				test.Errorf("Load = %v, want a fault naming %q", err, testCase.want)
			}
		})
	}
}

func sameFile(first, second assets.File) bool {
	return first.ContentType == second.ContentType && string(first.Body) == string(second.Body)
}

func sameIncluded(first, second assets.Included) bool {
	return slices.Equal(first.Stylesheets, second.Stylesheets) && first.Script == second.Script &&
		slices.Equal(first.Preloads, second.Preloads)
}

func removeFile(test *testing.T, folder, file string) {
	if err := os.Remove(filepath.Join(folder, file)); err != nil {
		test.Fatalf("remove %s: %v", file, err)
	}
}

// renameBuilt moves a built file to another name, and the manifest's name for it with it.
func renameBuilt(test *testing.T, folder, from, to string) {
	if err := os.Rename(filepath.Join(folder, from), filepath.Join(folder, to)); err != nil {
		test.Fatalf("rename %s: %v", from, err)
	}
	fixtures.ReplaceIn(test, folder, manifestPath, from, to)
}
