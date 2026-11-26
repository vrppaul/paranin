package compose

import (
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// Each page a person reads has its record, the facts a search engine reads as data: the résumé the person, with
// where they work now, the skills they use and not those still learning, the languages and the public
// profiles; a project's page the project, with its author; a post its current version, when it came out and when that
// version did, at the start of its day in the post's zone but never before the post, its author, its language, its
// tags and the series its page shows, by its group on the list by series in its language. An unlisted part of a
// series, which its series leaves out, names none; a draft not yet dated names no moment. Each names its own page.
func TestEachPageHasItsRecord(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	resume := document.PageOf(document.PageResume)
	author := Author{Name: "Mira Example", Page: resume}
	tidyTool := document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""}
	post, translation := document.PostPage(fixturePostSlug), document.PostPage(fixturePostSlug).InLanguage(russian)
	person := PersonRecord{
		Name:           "Mira Example",
		AlternateNames: []string{"M. Example"},
		Title:          "Software Engineer",
		Email:          "mira@example.test",
		Employer:       "Northwind",
		Skills:         []string{"Go", "Shell scripts for the build machines"},
		Languages:      []string{"English"},
		Profiles:       []string{"https://github.test/mira"},
		Page:           resume,
	}
	work := WorkRecord{
		Name:        "tidy-tool",
		Description: "A tool that tidies files.",
		Language:    "Lua",
		Repository:  "https://github.test/mira/tidy",
		Author:      author,
		Page:        tidyTool,
	}
	postOfASeries := PostRecord{
		Title: "Reading slowly", Description: "Why a page read twice is read once.", Page: post,
		Published: fixtures.MomentOf(test, "2026-11-10T09:00+01:00"), VersionOut: fixtures.MomentOf(test, "2026-11-12T00:00+01:00"),
		Author: author, Language: "en", Tags: []string{"reading", "terminals"},
		Series: &PostSeries{Title: "Reading", Group: document.Anchor{Page: document.PageOf(document.PageBlogSeries), Name: "reading"}},
	}
	unlistedPart := postOfASeries
	unlistedPart.Series = nil
	cases := []struct {
		name     string
		snapshot *content.Snapshot
		audience Audience
		page     document.Page
		want     Record
	}{
		{name: "the résumé", snapshot: snapshot, audience: publicOnTheDay(), page: resume, want: person},
		{name: "a project's page", snapshot: snapshot, audience: publicOnTheDay(), page: tidyTool, want: work},
		{name: "a post of a series", snapshot: snapshot, audience: publicOnTheDay(), page: post, want: postOfASeries},
		{
			name: "a translation", snapshot: snapshot, audience: publicOnTheDay(), page: translation,
			want: PostRecord{
				Title: "Медленное чтение", Description: "Почему страницу, прочитанную дважды, читают один раз.", Page: translation,
				Published: fixtures.MomentOf(test, "2026-11-15T10:00+03:00"), VersionOut: fixtures.MomentOf(test, "2026-11-15T10:00+03:00"),
				Author: author, Language: russian, Tags: []string{"reading", "terminals"},
				Series: &PostSeries{Title: "Чтение", Group: document.Anchor{Page: document.PageOf(document.PageBlogSeries).InLanguage(russian), Name: "reading"}},
			},
		},
		{name: "an unlisted part of a series", snapshot: changedFixture(test, fixturePostUnlisted), audience: publicOnTheDay(), page: post, want: unlistedPart},
		{
			name: "a draft its author reads", snapshot: snapshot, audience: Audience{At: fixtures.Day(), Drafts: true}, page: document.PostPage(samplePostSlug),
			want: PostRecord{
				Title: "Notes on tools", Description: "The tools a writer keeps.", Page: document.PostPage(samplePostSlug),
				Published: time.Time{}, VersionOut: time.Time{}, Author: author, Language: "en", Tags: nil, Series: nil,
			},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := recordOf(test, SiteOf(Prepare(testCase.snapshot), testCase.audience), testCase.page); !reflect.DeepEqual(got, testCase.want) {
				test.Errorf("the record of %+v = %+v, want %+v", testCase.page, got, testCase.want)
			}
		})
	}
}

// A record's text is plain: the marks of the profile's name, title, company and skill are read and gone.
func TestARecordReadsTheMarksAway(test *testing.T) {
	marked := *fixtures.LoadedContent(test)
	profile := &marked.Profile
	profile.Person.Name, profile.Person.Title = "==Mira== Example", "==Software== Engineer"
	profile.Experience = slices.Clone(profile.Experience)
	currentRole := slices.IndexFunc(profile.Experience, func(role content.Role) bool { return role.End.Kind() == content.MonthCurrent })
	profile.Experience[currentRole].Company = "==North==wind"
	profile.SkillGroups = []content.SkillGroup{{Title: "Languages", Skills: []content.Skill{{Name: "==Go==", Duration: content.Duration{}, Sources: nil, Learning: false}}}}
	record := as[PersonRecord](test, recordOf(test, SiteOf(Prepare(&marked), publicOnTheDay()), document.PageOf(document.PageResume)))
	got := markedFacts{Name: record.Name, Title: record.Title, Employer: record.Employer, Skills: record.Skills}
	want := markedFacts{Name: "Mira Example", Title: "Software Engineer", Employer: "Northwind", Skills: []string{"Go"}}
	if !reflect.DeepEqual(got, want) {
		test.Errorf("the marked facts of the résumé's record = %+v, want %+v", got, want)
	}
}

// markedFacts are the facts of a person's record a test marks in the profile.
type markedFacts struct {
	Name     string
	Title    string
	Employer string
	Skills   []string
}

// A person whose every role has ended works nowhere now, and the résumé's record names no employer.
func TestAPersonWithNoRoleGoingOnHasNoEmployer(test *testing.T) {
	ended := *fixtures.LoadedContent(test)
	ended.Profile.Experience = slices.DeleteFunc(slices.Clone(ended.Profile.Experience), func(role content.Role) bool {
		return role.End.Kind() == content.MonthCurrent
	})
	if record := as[PersonRecord](test, recordOf(test, SiteOf(Prepare(&ended), publicOnTheDay()), document.PageOf(document.PageResume))); record.Employer != "" {
		test.Errorf("the employer of a person with no role going on = %q, want none", record.Employer)
	}
}

// recordOf is the record compose makes beside the document of a page of a site: the résumé, a project's page or a
// post in any language.
func recordOf(test *testing.T, site Site, page document.Page) Record {
	composed := site.Pages
	for _, blog := range site.Blogs() {
		composed = slices.Concat(composed, blog.Posts)
	}
	for _, found := range composed {
		if found.Document.Head.Page == page {
			return found.Record
		}
	}
	test.Fatalf("compose made no page %+v", page)
	return nil
}
