package compose

import (
	"os"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// The content package's fixture, which compose reads as the site would
const fixturePath = "../content/testdata/content"

// Every fact of the profile a person reads reaches the résumé, as a phrase of its own: a face that shows every
// phrase shows every fact.
func TestEveryFactOfTheProfileReachesTheResume(test *testing.T) {
	phrases := document.Phrases(Resume(fixtureSnapshot(test)))
	cases := []struct {
		name string
		want string
	}{
		{name: "the name", want: "Mira Example"},
		{name: "the title", want: "Software Engineer"},
		{name: "the place", want: "Brno, Czechia"},
		{name: "the headline", want: "Builds small tools that last."},
		{name: "the summary", want: "Writes software for people who read it."},
		{name: "a fact's label", want: "available"},
		{name: "a fact's value", want: "5 years"},
		{name: "a gap", want: "[notice period]"},
		{name: "the mail address", want: "mira@example.test"},
		{name: "a profile's label", want: "GitHub"},
		{name: "a profile's address as written", want: "github.test/mira"},
		{name: "the ssh command", want: "ssh example.test"},
		{name: "a company", want: "Northwind"},
		{name: "a role", want: "Senior Engineer"},
		{name: "a role's place", want: "Prague"},
		{name: "a point", want: "Wrote the first parser"},
		{name: "a highlighted result", want: "queue"},
		{name: "a project", want: "tidy-tool"},
		{name: "a tagline", want: "Tidies files."},
		{name: "a description", want: "A tool that tidies files."},
		{name: "a project's language", want: "Lua"},
		{name: "a repository", want: "https://github.test/mira/tidy-tool"},
		{name: "an install command", want: "go install example.test/tidy-tool@latest"},
		{name: "a group of skills", want: "Tools"},
		{name: "a skill", want: "Docker"},
		{name: "a skill's duration", want: "side"},
		{name: "a skill's sources", want: "tidy-tool, a hobby"},
		{name: "a school", want: "A University"},
		{name: "a field of study", want: "Computer Science and the Engineering of Systems That Last"},
		{name: "a month still to fill", want: "[20XX]"},
		{name: "a language", want: "English"},
		{name: "a language's level", want: "fluent"},
		{name: "a word of the site", want: "Experience"},
		{name: "a label of the site", want: "repository"},
		{name: "the word for what still goes on", want: "present"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !slices.Contains(phrases, testCase.want) {
				test.Errorf("phrases = %q, want one %q", phrases, testCase.want)
			}
		})
	}
}

// The résumé's parts stand in one order, decided once here, which every face follows: the summary, then each
// section under the site's word for it.
func TestTheResumeKeepsOneOrder(test *testing.T) {
	blocks := Resume(fixtureSnapshot(test)).Blocks
	as[document.Paragraph](test, blocks[0])
	got := []string{}
	for _, block := range blocks[1:] {
		got = append(got, titleOf(test, as[document.Section](test, block)))
	}
	want := []string{"Facts", "Experience", "Projects", "Skills", "Education", "Languages"}
	if !slices.Equal(got, want) {
		test.Errorf("sections = %q, want %q", got, want)
	}
}

// A role's months reach the document typed, so each face writes them its own way, and a role still held stays
// current.
func TestEachRoleKeepsItsPeriod(test *testing.T) {
	experience := sectionTitled(test, Resume(fixtureSnapshot(test)), "Experience")
	got := []document.Period{}
	for _, block := range experience.Blocks {
		got = append(got, as[document.Entry](test, block).Period)
	}
	want := []document.Period{
		{Start: document.KnownMonth{Start: firstMoment(2023, time.April)}, End: document.CurrentMonth{Text: "present"}},
		{Start: document.KnownMonth{Start: firstMoment(2020, time.January)}, End: document.KnownMonth{Start: firstMoment(2023, time.March)}},
	}
	if !reflect.DeepEqual(got, want) {
		test.Errorf("periods = %+v, want %+v", got, want)
	}
}

// Every link of the résumé and of a project's page leads where the profile says: a face writes the address, but
// never chooses it.
func TestEveryLinkLeadsWhereTheProfileSays(test *testing.T) {
	documents := Documents(fixtureSnapshot(test))
	resume := documentOf(test, documents, document.Page{Kind: document.PageResume, Slug: ""})
	project := documentOf(test, documents, document.Page{Kind: document.PageProject, Slug: "tidy-tool"})
	projects := sectionTitled(test, resume, "Projects")
	pageless, tidy := cardTitled(test, projects, "other-tool"), cardTitled(test, projects, "tidy-tool")
	cases := []struct {
		name string
		got  document.Target
		want document.Target
	}{
		{name: "the mail address", got: linkTarget(test, factLabelled(test, resume.Head.Ways, "mail").Value), want: document.Address{URL: "mailto:mira@example.test"}},
		{name: "a public profile", got: linkTarget(test, factLabelled(test, resume.Head.Ways, "GitHub").Value), want: document.Address{URL: "https://github.test/mira"}},
		{name: "a project's page", got: tidy.Page, want: document.Page{Kind: document.PageProject, Slug: "tidy-tool"}},
		{name: "a project with no page", got: pageless.Page, want: document.Page{Kind: document.PageNone, Slug: ""}},
		{name: "a project's repository", got: linkTarget(test, factLabelled(test, tidy.Facts, "repository").Value), want: document.Address{URL: "https://github.test/mira/tidy-tool"}},
		{name: "a project's page back to the résumé", got: linkTarget(test, factLabelled(test, detailsOf(test, project), "author").Value), want: document.Page{Kind: document.PageResume, Slug: ""}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if testCase.got != testCase.want {
				test.Errorf("target = %+v, want %+v", testCase.got, testCase.want)
			}
		})
	}
}

// documentOf is the document of a page, wherever compose put it.
func documentOf(test *testing.T, documents []document.Document, page document.Page) document.Document {
	for _, found := range documents {
		if found.Head.Page == page {
			return found
		}
	}
	test.Fatalf("no document is the page %+v", page)
	return document.Document{Head: document.Head{Page: page, Title: nil, Subtitle: nil, Place: "", Lead: nil, Ways: nil}, Blocks: nil}
}

// cardTitled is a section's card under a title, wherever compose put it.
func cardTitled(test *testing.T, section document.Section, title string) document.Card {
	for _, block := range section.Blocks {
		if card, isCard := block.(document.Card); isCard && as[document.Text](test, card.Title[0]).Text == title {
			return card
		}
	}
	test.Fatalf("the section has no card %q", title)
	return document.Card{Title: nil, Subtitle: nil, Tag: "", Summary: nil, Facts: nil, Page: document.Page{Kind: document.PageNone, Slug: ""}}
}

// detailsOf are a page's facts, wherever compose put them.
func detailsOf(test *testing.T, page document.Document) []document.Fact {
	for _, block := range page.Blocks {
		if facts, isFacts := block.(document.Facts); isFacts {
			return facts.List
		}
	}
	test.Fatalf("the page %+v has no facts", page.Head.Page)
	return nil
}

// factLabelled is the fact under a label.
func factLabelled(test *testing.T, facts []document.Fact, label string) document.Fact {
	for _, fact := range facts {
		if fact.Label == label {
			return fact
		}
	}
	test.Fatalf("no fact is labelled %q among %+v", label, facts)
	return document.Fact{Label: label, Value: nil}
}

// sectionTitled is the résumé's section under a title, wherever compose put it.
func sectionTitled(test *testing.T, resume document.Document, title string) document.Section {
	for _, block := range resume.Blocks {
		if section, isSection := block.(document.Section); isSection && titleOf(test, section) == title {
			return section
		}
	}
	test.Fatalf("the résumé has no section %q", title)
	return document.Section{Title: nil, Blocks: nil}
}

func titleOf(test *testing.T, section document.Section) string {
	return as[document.Text](test, section.Title[0]).Text
}

// linkTarget is where a value that is one link leads.
func linkTarget(test *testing.T, spans []document.Span) document.Target {
	return as[document.Link](test, spans[0]).Target
}

// as is a value as the kind a test expects of it, or the test's end, saying what it was.
func as[Kind any](test *testing.T, value any) Kind {
	typed, isKind := value.(Kind)
	if !isKind {
		test.Fatalf("%+v is no %T", value, typed)
	}
	return typed
}

func fixtureSnapshot(test *testing.T) *content.Snapshot {
	snapshot, err := content.Load(os.DirFS(fixturePath))
	if err != nil {
		test.Fatalf("Load(%q) = %v, want no error", fixturePath, err)
	}
	return snapshot
}

func firstMoment(year int, month time.Month) time.Time {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
}
