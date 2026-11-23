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

// The fixture's published post's file, and its status, which a case turns to a draft's; and the file's line the
// sample post's body opens on, after its front matter
const (
	fixturePostFile    = "posts/reading-slowly/post.en.md"
	publishedStatus    = "status: published"
	draftStatus        = "status: draft"
	samplePostBodyLine = 8
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
			errors.As(Check(snapshot), &clashes)
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

// faultSpot is where a fault of a link stands, as a test reads it: its file and its line.
type faultSpot struct {
	file string
	line int
}

// A post's link inside the site leads to a page the site serves by the path it is served at, and to a section of that
// page where it names one: a path no page takes, a post's number, which only sends the reader on, a page's markdown, a
// draft's page, or an anchor no section of its page takes, is a fault at its file's line. Links to the résumé, a tab, a
// project's page and its tab, the blog's lists, a heading and a series' group lead somewhere.
func TestEveryLinkInsideTheSiteLeadsToAPageAndItsSection(test *testing.T) {
	cases := []struct {
		name   string
		body   string
		change func(test *testing.T, folder string)
		want   []faultSpot
	}{
		{
			name:   "links that lead somewhere",
			body:   "[a](/) [b](/experience) [c](/p/tidy-tool) [d](/p/tidy-tool/why) [e](/blog) [f](/blog/series#reading) [g](/blog/reading-slowly#what-a-reader-keeps) [h](/blog/reading-slowly#in-a-terminal)",
			change: keepFolder, want: []faultSpot{},
		},
		{name: "a path no page takes", body: "[a](/blog/no-such-post)", change: keepFolder, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a post's number", body: "[a](/blog/1)", change: keepFolder, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a page's markdown", body: "[a](/p/tidy-tool.md)", change: keepFolder, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a draft's page", body: "[a](/blog/reading-slowly)", change: fixturePostDrafted, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a heading no post holds", body: "[a](/blog/reading-slowly#no-such-heading)", change: keepFolder, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "an anchor of a page whose sections take none", body: "[a](/experience#experience)", change: keepFolder, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a link on a later line", body: "[a](/), then\n[b](/nowhere)", change: keepFolder, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine + 1}}},
		{
			name: "links on several lines, in their order", body: "[a](/nowhere), then\n[b](/elsewhere), then\n[c](/anywhere)", change: keepFolder,
			want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}, {file: samplePostFile, line: samplePostBodyLine + 1}, {file: samplePostFile, line: samplePostBodyLine + 2}},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			folder := fixtures.Copy(test, fixtures.ContentFolder(test))
			fixtures.WriteIn(test, folder, samplePostFile, samplePostFields+publishedStatus+"\n---\n"+testCase.body+"\n")
			testCase.change(test, folder)
			snapshot, err := content.Load(os.DirFS(folder))
			if err != nil {
				test.Fatalf("load the fixture with the body %q: %v", testCase.body, err)
			}
			var faults content.Faults
			errors.As(Check(snapshot), &faults)
			got := []faultSpot{}
			for _, fault := range faults {
				got = append(got, faultSpot{file: fault.File, line: fault.Line})
			}
			if !reflect.DeepEqual(got, testCase.want) {
				test.Errorf("the faults of the links of %q = %+v, want %+v", testCase.body, got, testCase.want)
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

// fixturePostDrafted turns the fixture's published post to a draft, which the site serves at no address.
//
//wrapper:a case of the table names the change it writes
func fixturePostDrafted(test *testing.T, folder string) {
	fixtures.ReplaceIn(test, folder, fixturePostFile, publishedStatus, draftStatus)
}
