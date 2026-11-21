package compose

import (
	"reflect"
	"slices"
	"testing"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// Each page a person reads has its record, the facts a search engine reads as data: the résumé the person, with
// where they work now, the schools, the skills they use and not those still learning, the languages and the public
// profiles; a project's page the project, with its author and the page that is theirs. Each names its own page.
func TestEachPageHasItsRecord(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	tidyTool := document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""}
	person := PersonRecord{
		Name:           "Mira Example",
		AlternateNames: []string{"M. Example"},
		Title:          "Software Engineer",
		Email:          "mira@example.test",
		Employer:       "Northwind",
		Schools:        []string{"A University"},
		Skills:         []string{"Go", "Shell scripts for the build machines"},
		Languages:      []string{"English"},
		Profiles:       []string{"https://github.test/mira"},
		Page:           resumePage(),
	}
	work := WorkRecord{
		Name:        "tidy-tool",
		Description: "A tool that tidies files.",
		Language:    "Lua",
		Repository:  "https://github.test/mira/tidy",
		Author:      "Mira Example",
		AuthorPage:  resumePage(),
		Page:        tidyTool,
	}
	cases := []struct {
		name string
		page document.Page
		want Record
	}{
		{name: "the résumé", page: resumePage(), want: person},
		{name: "a project's page", page: tidyTool, want: work},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := recordOf(test, snapshot, testCase.page); !reflect.DeepEqual(got, testCase.want) {
				test.Errorf("the record of %+v = %+v, want %+v", testCase.page, got, testCase.want)
			}
		})
	}
}

// A record's text is plain: the marks of the profile's name, title, company, school and skill are read and gone.
func TestARecordReadsTheMarksAway(test *testing.T) {
	marked := *fixtures.LoadedContent(test)
	profile := &marked.Profile
	profile.Person.Name, profile.Person.Title = "==Mira== Example", "==Software== Engineer"
	profile.Experience = slices.Clone(profile.Experience)
	currentRole := slices.IndexFunc(profile.Experience, func(role content.Role) bool { return role.End.Kind() == content.MonthCurrent })
	profile.Experience[currentRole].Company = "==North==wind"
	profile.Education = slices.Clone(profile.Education)
	for index := range profile.Education {
		profile.Education[index].Name = "==" + profile.Education[index].Name + "=="
	}
	profile.SkillGroups = []content.SkillGroup{{Title: "Languages", Skills: []content.Skill{{Name: "==Go==", Duration: content.Duration{}, Sources: nil, Learning: false}}}}
	record := as[PersonRecord](test, recordOf(test, &marked, resumePage()))
	got := markedFacts{Name: record.Name, Title: record.Title, Employer: record.Employer, Schools: record.Schools, Skills: record.Skills}
	want := markedFacts{Name: "Mira Example", Title: "Software Engineer", Employer: "Northwind", Schools: []string{"A University"}, Skills: []string{"Go"}}
	if !reflect.DeepEqual(got, want) {
		test.Errorf("the marked facts of the résumé's record = %+v, want %+v", got, want)
	}
}

// markedFacts are the facts of a person's record a test marks in the profile.
type markedFacts struct {
	Name     string
	Title    string
	Employer string
	Schools  []string
	Skills   []string
}

// A person whose every role has ended works nowhere now, and the résumé's record names no employer.
func TestAPersonWithNoRoleGoingOnHasNoEmployer(test *testing.T) {
	ended := *fixtures.LoadedContent(test)
	ended.Profile.Experience = slices.DeleteFunc(slices.Clone(ended.Profile.Experience), func(role content.Role) bool {
		return role.End.Kind() == content.MonthCurrent
	})
	if record := as[PersonRecord](test, recordOf(test, &ended, resumePage())); record.Employer != "" {
		test.Errorf("the employer of a person with no role going on = %q, want none", record.Employer)
	}
}

// recordOf is the record compose makes beside the document of a page.
func recordOf(test *testing.T, snapshot *content.Snapshot, page document.Page) Record {
	for _, composed := range Pages(snapshot) {
		if composed.Document.Head.Page == page {
			return composed.Record
		}
	}
	test.Fatalf("compose made no page %+v", page)
	return nil
}
