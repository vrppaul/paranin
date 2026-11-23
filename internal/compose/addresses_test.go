package compose

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// How a case writes a post of its own, and the catalogue's word a case retitles a section of the résumé by
const (
	sharedPostText      = "---\nid: 9\ntitle: A post\ndescription: About it.\nstatus: published\npublished: 2026-11-20T09:00+01:00\n---\nIts text.\n"
	wordsFile           = "i18n/en.yaml"
	experienceWord      = "experience: Experience"
	experienceAsTheBlog = "experience: Blog"
)

// No two pages of the site share an address in any format: a post slugged as the address the blog's list by series
// takes, or a tab of the résumé titled as the blog, whose address the list by year takes, is found at every address it
// shares; content whose pages each have their own is found to share none.
func TestNoTwoPagesShareAnAddress(test *testing.T) {
	cases := []struct {
		name   string
		change func(test *testing.T, folder string)
		want   []string
	}{
		{name: "the fixture", change: keepFolder, want: []string{}},
		{name: "a post slugged as the list by series", change: postSluggedSeries, want: []string{"/blog/series", "/blog/series.md", "/blog/series.txt"}},
		{name: "a tab of the résumé titled as the blog", change: tabTitledBlog, want: []string{"/blog", "/blog.md", "/blog.txt"}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			folder := fixtures.Copy(test, fixtures.ContentFolder(test))
			testCase.change(test, folder)
			snapshot, err := content.Load(os.DirFS(folder))
			if err != nil {
				test.Fatalf("load the fixture changed: %v", err)
			}
			var clashes AddressClashes
			errors.As(SharedAddresses(snapshot), &clashes)
			paths := []string{}
			for _, clash := range clashes {
				paths = append(paths, clash.Path)
			}
			if !reflect.DeepEqual(paths, testCase.want) {
				test.Errorf("the addresses shared = %q, want %q", paths, testCase.want)
			}
		})
	}
}

// keepFolder leaves the fixture as it is.
func keepFolder(*testing.T, string) {}

// postSluggedSeries writes a post whose folder bears the name the list by series takes in the blog's folder.
//
//wrapper:a case of the table names the change it writes
func postSluggedSeries(test *testing.T, folder string) {
	fixtures.WriteIn(test, folder, "posts/series/post.en.md", sharedPostText)
}

// tabTitledBlog titles the résumé's experience as the blog, so its tab takes the name the blog's list stands at.
//
//wrapper:a case of the table names the change it writes
func tabTitledBlog(test *testing.T, folder string) {
	fixtures.ReplaceIn(test, folder, wordsFile, experienceWord, experienceAsTheBlog)
}
