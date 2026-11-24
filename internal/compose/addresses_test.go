package compose

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// How a case writes a post of its own, its file and the moment it is dated, and the catalogue's word a case retitles a
// section of the résumé by
const (
	sharedPostFile      = "posts/series/post.en.md"
	sharedPostMoment    = "2026-11-20T09:00+01:00"
	sharedPostText      = "---\nid: 9\ntitle: A post\ndescription: About it.\nstatus: published\npublished: " + sharedPostMoment + "\n---\nIts text.\n"
	wordsFile           = "i18n/en.yaml"
	experienceWord      = "experience: Experience"
	experienceAsTheBlog = "experience: Blog"
	// the experience retitled as the preview's folder is named
	experienceAsThePreview = "experience: Preview"
)

// The fixture's published post's file, its revision, and its status, which a case turns to a draft's; the field a case
// dates a post by, and the day the fixture's post was updated, which a case dating it later takes away with its
// revision; and the file's line the sample post's body opens on, after its front matter
const (
	fixturePostFile     = "posts/reading-slowly/post.en.md"
	fixtureRevisionFile = "posts/reading-slowly/revisions/2026-11-10.en.md"
	publishedStatus     = "status: published"
	draftStatus         = "status: draft"
	previewStatus       = "status: preview"
	publishedField      = "published: "
	fixturePostUpdated  = "\nupdated: 2026-11-12"
	samplePostBodyLine  = 8
)

// A translation a case writes, the fixture's and one beside a post of its own; the line its body opens on, after its
// front matter; the source it names, of a source's form and of no text of its original, so it loads stale; and a path
// no page takes
const (
	sharedTranslationFile = "posts/series/post.ru.md"
	translationBodyLine   = 8
	unknownSource         = "en@000000000000"
	nowhere               = "/nowhere"
)

// No two pages of the site share an address in any format: a post slugged as the address the blog's list by series
// takes, a draft, a preview or one out later among them, a tab of the résumé titled as the blog, whose address the
// list by year takes, or one titled as the preview's folder, is found at every address it shares; content whose pages
// each have their own is found to share none.
func TestNoTwoPagesShareAnAddress(test *testing.T) {
	cases := []struct {
		name   string
		change func(test *testing.T, folder string)
		want   []string
	}{
		{name: "the fixture", change: fixtures.KeepFolder, want: []string{}},
		{name: "a post slugged as the list by series", change: postSluggedSeries, want: []string{"/blog/series", "/blog/series.md", "/blog/series.txt"}},
		{name: "a draft slugged as the list by series", change: draftSluggedSeries, want: []string{"/blog/series", "/blog/series.md", "/blog/series.txt"}},
		{
			name: "a translation slugged as the list by series", change: translationSluggedSeries,
			want: []string{"/blog/series", "/blog/series.md", "/blog/series.txt", "/ru/blog/series", "/ru/blog/series.md", "/ru/blog/series.txt"},
		},
		{
			name: "a translation previewed, slugged as the list by series", change: translationPreviewSluggedSeries,
			want: []string{"/blog/series", "/blog/series.md", "/blog/series.txt", "/ru/blog/series", "/ru/blog/series.md", "/ru/blog/series.txt"},
		},
		{name: "a post out later slugged as the list by series", change: laterPostSluggedSeries, want: []string{"/blog/series", "/blog/series.md", "/blog/series.txt"}},
		{name: "a preview slugged as the list by series", change: previewSluggedSeries, want: []string{"/blog/series", "/blog/series.md", "/blog/series.txt"}},
		{name: "a tab of the résumé titled as the blog", change: tabTitledBlog, want: []string{"/blog", "/blog.md", "/blog.txt"}},
		{name: "a tab of the résumé titled as the preview's folder", change: tabTitledPreview, want: []string{"/preview/"}},
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
			errors.As(Check(snapshot, fixtures.Day()), &clashes)
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
// draft's page, a preview's, a post out after the post linking it, or an anchor no section of its page takes, a feed's
// among them, is a fault at its file's line. Links to the résumé, a tab, a project's page and its tab, the blog's
// lists, a heading, a series' group and a feed lead somewhere, and to a post out before the post linking it, or with
// it. A draft's links are held to the site as its author reads it once every post dated later is out, where a draft's
// page leads somewhere; a preview's, to the site its link's reader reads then, where its own page leads somewhere and
// no draft's or other preview's does.
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
			change: fixtures.KeepFolder, want: []faultSpot{},
		},
		{name: "the list filtered by a tag", body: "[a](/blog?tag=terminals)", change: fixtures.KeepFolder, want: []faultSpot{}},
		{name: "an older version and its heading", body: "[a](/blog/reading-slowly/2026-11-10) [b](/blog/reading-slowly/2026-11-10#what-a-reader-keeps)", change: fixtures.KeepFolder, want: []faultSpot{}},
		{name: "the blog's feed and a series'", body: "[a](/blog.atom) [b](/blog/series/reading.atom)", change: fixtures.KeepFolder, want: []faultSpot{}},
		{
			name: "a translation, its heading, and its language's list and feed", body: "[a](/ru/blog/reading-slowly) [b](/ru/blog/reading-slowly#потом) [c](/ru/blog) [d](/ru/blog.atom)",
			change: fixtures.KeepFolder, want: []faultSpot{},
		},
		{name: "a heading of the original its translation has not", body: "[a](/ru/blog/reading-slowly#what-a-reader-keeps)", change: fixtures.KeepFolder, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a post in a language it has no translation in", body: "[a](/ru/blog/notes-on-tools)", change: fixtures.KeepFolder, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a translation's link to no page", body: "[a](/)", change: translationLinkingNowhere, want: []faultSpot{{file: translationFile, line: translationBodyLine}}},
		{name: "a draft translation's link to no page", body: "[a](/)", change: draftTranslationLinkingNowhere, want: []faultSpot{{file: translationFile, line: translationBodyLine}}},
		{name: "a translation preview's link to no page", body: "[a](/)", change: previewTranslationLinkingNowhere, want: []faultSpot{{file: translationFile, line: translationBodyLine}}},
		{name: "a section of a feed", body: "[a](/blog.atom#reading)", change: fixtures.KeepFolder, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a revision's link to no page", body: "Its text.", change: revisionLinkingNowhere, want: []faultSpot{{file: sampleRevisionFile, line: sampleRevisionLine}}},
		{name: "the list filtered by a tag no post carries", body: "[a](/blog?tag=nothing)", change: fixtures.KeepFolder, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a path no page takes", body: "[a](/blog/no-such-post)", change: fixtures.KeepFolder, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a post's number", body: "[a](/blog/1)", change: fixtures.KeepFolder, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a page's markdown", body: "[a](/p/tidy-tool.md)", change: fixtures.KeepFolder, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a draft's page", body: "[a](/blog/reading-slowly)", change: fixturePostDrafted, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a post out after the post linking it", body: "[a](/blog/reading-slowly)", change: fixturePostLater, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a post out after the post linking it, both later", body: "[a](/blog/reading-slowly)", change: linkingPostSooner, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a post out before the post linking it, both later", body: "[a](/blog/reading-slowly)", change: linkingPostLater, want: []faultSpot{}},
		{name: "a post out with the post linking it, both later", body: "[a](/blog/reading-slowly)", change: bothPostsLaterTogether, want: []faultSpot{}},
		{name: "a post's link to a preview", body: "[a](/blog/reading-slowly)", change: fixturePostPreviewed, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a preview's link to its own section", body: "## Its heading\n\n[a](/blog/notes-on-tools#its-heading)", change: samplePostPreviewed, want: []faultSpot{}},
		{name: "a preview's link to no page", body: "[a](/nowhere)", change: samplePostPreviewed, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a preview's link to a draft", body: "[a](/blog/reading-slowly)", change: previewLinkingADraft, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a preview's link to another preview", body: "[a](/blog/reading-slowly)", change: bothPreviewed, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a preview's link to a post out later", body: "[a](/blog/reading-slowly)", change: previewLinkingALaterPost, want: []faultSpot{}},
		{name: "a draft's link to no page", body: "[a](/nowhere)", change: samplePostDrafted, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a draft's link to a draft", body: "[a](/blog/reading-slowly)", change: bothDrafted, want: []faultSpot{}},
		{name: "a draft's link to a post out later", body: "[a](/blog/reading-slowly)", change: draftLinkingALaterPost, want: []faultSpot{}},
		{name: "a heading no post holds", body: "[a](/blog/reading-slowly#no-such-heading)", change: fixtures.KeepFolder, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "an anchor of a page whose sections take none", body: "[a](/experience#experience)", change: fixtures.KeepFolder, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine}}},
		{name: "a link on a later line", body: "[a](/), then\n[b](/nowhere)", change: fixtures.KeepFolder, want: []faultSpot{{file: samplePostFile, line: samplePostBodyLine + 1}}},
		{
			name: "links on several lines, in their order", body: "[a](/nowhere), then\n[b](/elsewhere), then\n[c](/anywhere)", change: fixtures.KeepFolder,
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
			errors.As(Check(snapshot, fixtures.Day()), &faults)
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

// postSluggedSeries writes a post whose folder bears the name the list by series takes in the blog's folder.
//
//wrapper:a case of the table names the change it writes
func postSluggedSeries(test *testing.T, folder string) {
	fixtures.WriteIn(test, folder, sharedPostFile, sharedPostText)
}

// translationSluggedSeries writes a post and its translation whose folder bears the name the list by series takes.
func translationSluggedSeries(test *testing.T, folder string) {
	postSluggedSeries(test, folder)
	fixtures.WriteIn(test, folder, sharedTranslationFile, translationLinkingTo(publishedStatus, "/"))
}

// translationLinkingNowhere writes the fixture's translation, published, with a link to no page.
//
//wrapper:a case of the table names the change it writes
func translationLinkingNowhere(test *testing.T, folder string) {
	fixtures.WriteIn(test, folder, translationFile, translationLinkingTo(publishedStatus, nowhere))
}

// draftTranslationLinkingNowhere writes the fixture's translation, a draft, with a link to no page.
//
//wrapper:a case of the table names the change it writes
func draftTranslationLinkingNowhere(test *testing.T, folder string) {
	fixtures.WriteIn(test, folder, translationFile, translationLinkingTo(draftStatus, nowhere))
}

// previewTranslationLinkingNowhere writes the fixture's translation, a preview, with a link to no page.
//
//wrapper:a case of the table names the change it writes
func previewTranslationLinkingNowhere(test *testing.T, folder string) {
	fixtures.WriteIn(test, folder, translationFile, translationLinkingTo(previewStatus, nowhere))
}

// translationPreviewSluggedSeries writes a post and its translation, a preview, whose folder bears the name the list by
// series takes.
func translationPreviewSluggedSeries(test *testing.T, folder string) {
	postSluggedSeries(test, folder)
	fixtures.WriteIn(test, folder, sharedTranslationFile, translationLinkingTo(previewStatus, "/"))
}

// translationLinkingTo is a translation's text of a status whose body is a link to a path, on translationBodyLine.
func translationLinkingTo(status, path string) string {
	return "---\ntitle: Перевод\ndescription: О нём.\n" + status + "\npublished: 2026-11-15T10:00+03:00\nsource: " + unknownSource + "\n---\n[a](" + path + ")\n"
}

// draftSluggedSeries writes a draft whose folder bears the name the list by series takes in the blog's folder.
func draftSluggedSeries(test *testing.T, folder string) {
	postSluggedSeries(test, folder)
	fixtures.ReplaceIn(test, folder, sharedPostFile, publishedStatus, draftStatus)
}

// laterPostSluggedSeries writes a post whose folder bears the name the list by series takes in the blog's folder, dated
// after the tests' day.
func laterPostSluggedSeries(test *testing.T, folder string) {
	postSluggedSeries(test, folder)
	fixtures.ReplaceIn(test, folder, sharedPostFile, publishedField+sharedPostMoment, publishedField+afterTheDay)
}

// tabTitledBlog titles the résumé's experience as the blog, so its tab takes the name the blog's list stands at.
//
//wrapper:a case of the table names the change it writes
func tabTitledBlog(test *testing.T, folder string) {
	fixtures.ReplaceIn(test, folder, wordsFile, experienceWord, experienceAsTheBlog)
}

// tabTitledPreview titles the résumé's experience as the preview's folder, so its tab's folder would be one.
//
//wrapper:a case of the table names the change it writes
func tabTitledPreview(test *testing.T, folder string) {
	fixtures.ReplaceIn(test, folder, wordsFile, experienceWord, experienceAsThePreview)
}

// fixturePostDrafted turns the fixture's published post to a draft, which the site serves at no address.
//
//wrapper:a case of the table names the change it writes
func fixturePostDrafted(test *testing.T, folder string) {
	fixtures.ReplaceIn(test, folder, fixturePostFile, publishedStatus, draftStatus)
}

// samplePostDrafted turns the sample post to a draft, whose links its author's site holds.
//
//wrapper:a case of the table names the change it writes
func samplePostDrafted(test *testing.T, folder string) {
	fixtures.ReplaceIn(test, folder, samplePostFile, publishedStatus, draftStatus)
}

// bothDrafted turns the sample post and the fixture's post to drafts.
func bothDrafted(test *testing.T, folder string) {
	samplePostDrafted(test, folder)
	fixturePostDrafted(test, folder)
}

// fixturePostLater dates the fixture's post after the tests' day, so it comes out later than the sample post, out now.
//
//wrapper:a case of the table names the change it writes
func fixturePostLater(test *testing.T, folder string) {
	fixturePostDated(test, folder, afterTheDay)
}

// linkingPostSooner dates both posts after the tests' day, the sample post, which links, sooner than the fixture's.
func linkingPostSooner(test *testing.T, folder string) {
	fixturePostDated(test, folder, afterTheDay)
	fixtures.ReplaceIn(test, folder, samplePostFile, publishedField+sampleMoment, publishedField+soonAfterTheDay)
}

// linkingPostLater dates both posts after the tests' day, the sample post, which links, later than the fixture's.
func linkingPostLater(test *testing.T, folder string) {
	fixturePostDated(test, folder, soonAfterTheDay)
	fixtures.ReplaceIn(test, folder, samplePostFile, publishedField+sampleMoment, publishedField+afterTheDay)
}

// bothPostsLaterTogether dates both posts at one moment after the tests' day.
func bothPostsLaterTogether(test *testing.T, folder string) {
	fixturePostDated(test, folder, afterTheDay)
	fixtures.ReplaceIn(test, folder, samplePostFile, publishedField+sampleMoment, publishedField+afterTheDay)
}

// fixturePostDated dates the fixture's post a moment, which leaves it updated on no day before it, nor with a revision
// that came out before it.
func fixturePostDated(test *testing.T, folder, moment string) {
	fixtures.ReplaceIn(test, folder, fixturePostFile, publishedField+fixturePostMoment+fixturePostUpdated, publishedField+moment)
	fixtures.RemoveIn(test, folder, fixtureRevisionFile)
}

// fixturePostPreviewed turns the fixture's published post to a preview, which its link alone opens.
//
//wrapper:a case of the table names the change it writes
func fixturePostPreviewed(test *testing.T, folder string) {
	fixtures.ReplaceIn(test, folder, fixturePostFile, publishedStatus, previewStatus)
}

// samplePostPreviewed turns the sample post to a preview, whose links its link's reader follows.
//
//wrapper:a case of the table names the change it writes
func samplePostPreviewed(test *testing.T, folder string) {
	fixtures.ReplaceIn(test, folder, samplePostFile, publishedStatus, previewStatus)
}

// previewLinkingADraft turns the sample post to a preview, and the fixture's post to a draft.
func previewLinkingADraft(test *testing.T, folder string) {
	samplePostPreviewed(test, folder)
	fixturePostDrafted(test, folder)
}

// bothPreviewed turns the sample post and the fixture's post to previews.
func bothPreviewed(test *testing.T, folder string) {
	samplePostPreviewed(test, folder)
	fixturePostPreviewed(test, folder)
}

// previewLinkingALaterPost turns the sample post to a preview, and dates the fixture's post after the tests' day.
func previewLinkingALaterPost(test *testing.T, folder string) {
	samplePostPreviewed(test, folder)
	fixturePostLater(test, folder)
}

// previewSluggedSeries writes a preview whose folder bears the name the list by series takes in the blog's folder.
func previewSluggedSeries(test *testing.T, folder string) {
	postSluggedSeries(test, folder)
	fixtures.ReplaceIn(test, folder, sharedPostFile, publishedStatus, previewStatus)
}

// draftLinkingALaterPost turns the sample post to a draft, and dates the fixture's post after the tests' day.
func draftLinkingALaterPost(test *testing.T, folder string) {
	samplePostDrafted(test, folder)
	fixturePostLater(test, folder)
}
