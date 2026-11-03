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
	stylesheetPath = "assets/ui-fixture.css"
)

// Every file the manifest names is served at its path from the site's root, as the type its name says; the
// manifest itself is no file of the site.
func TestABuildServesEveryFileItsManifestNames(test *testing.T) {
	build := fixtures.LoadedBuild(test)
	want := map[string]assets.File{"/" + stylesheetPath: {ContentType: "text/css; charset=utf-8", Body: []byte("body{margin:0}\n")}}
	if got := maps.Collect(build.Files()); !maps.EqualFunc(got, want, sameFile) {
		test.Errorf("Files() = %q, want %q", got, want)
	}
}

// A page that includes the stylesheet links the file it was built into.
func TestTheStylesheetLinksItsBuiltFile(test *testing.T) {
	build := fixtures.LoadedBuild(test)
	if got, want := build.Stylesheets(assets.Stylesheet), []string{"/" + stylesheetPath}; !slices.Equal(got, want) {
		test.Errorf("Stylesheets(%q) = %q, want %q", assets.Stylesheet, got, want)
	}
}

// A build the site cannot be served from is refused, the fault named: no manifest, one that does not read, one
// lacking an entry the site names, a file it names missing or of a type the site serves none of.
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
				fixtures.ReplaceIn(test, folder, manifestPath, `"src/ui/ui.css": {`, `"src/ui/other.css": {`)
			},
			want: `"src/ui/ui.css"`,
		},
		{name: "a file missing", change: func(test *testing.T, folder string) { removeFile(test, folder, stylesheetPath) }, want: stylesheetPath},
		{
			name: "a file of no type the site serves",
			change: func(test *testing.T, folder string) {
				renameBuilt(test, folder, stylesheetPath, "assets/ui-fixture.png")
			},
			want: `"assets/ui-fixture.png"`,
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
