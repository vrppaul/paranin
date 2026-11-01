package content

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

// The folder of the fixture profile, and the repository's own content, from this package
const (
	fixtureFolder    = "testdata/profile"
	repositoryFolder = "../../content"
)

func TestTheRepositoryContentLoads(test *testing.T) {
	if _, err := Load(os.DirFS(repositoryFolder)); err != nil {
		test.Errorf("Load(%q) = %v, want no error", repositoryFolder, err)
	}
}

func TestEveryFaultIsReportedAtItsField(test *testing.T) {
	cases := []struct {
		name        string
		original    string
		replacement string
		field       string
	}{
		{name: "an empty name", original: "name: Mira Example", replacement: `name: ""`, field: "person.name"},
		{name: "a mail address with a name round it", original: "email: mira@example.test", replacement: "email: Mira <mira@example.test>", field: "person.email"},
		{name: "a base address ending in a slash", original: "base_url: https://example.test", replacement: "base_url: https://example.test/", field: "site.base_url"},
		{name: "a link that is no address", original: "url: https://github.test/mira\n", replacement: "url: github mira\n", field: "person.links[0].url"},
		{name: "a month past twelve", original: "start: 2023-04", replacement: "start: 2023-13", field: "experience[0].start"},
		{name: "a role whose start is a gap", original: "start: 2023-04", replacement: `start: "[2023]"`, field: "experience[0].start"},
		{name: "a role ended by the old word", original: "end: 2023-03", replacement: "end: present", field: "experience[1].end"},
		{name: "a role whose end is left empty", original: "end: 2023-03", replacement: "end:", field: "experience[1].end"},
		{name: "a role that ends before it starts", original: "end: 2023-03", replacement: "end: 2019-01", field: "experience[1].end"},
		{name: "a school month neither a month nor a gap", original: `end: "[20XX]"`, replacement: "end: later", field: "education[0].end"},
		{name: "a school that ends before it starts", original: `end: "[20XX]"`, replacement: "end: 2010-01", field: "education[0].end"},
		{name: "a slug with capitals", original: "slug: tidy-tool", replacement: "slug: Tidy-Tool", field: "projects[1].slug"},
		{name: "a slug two projects share", original: "slug: other-tool", replacement: "slug: tidy-tool", field: "projects[1].slug"},
		{name: "a skill in no group", original: "group: languages, duration: 3y", replacement: "group: music, duration: 3y", field: "skills[0].group"},
		{name: "a group written twice", original: "key: tools", replacement: "key: languages", field: "skill_groups[1].key"},
		{name: "a level of no kind", original: "level: warn", replacement: "level: bad", field: "health[0].checks[1].level"},
		{name: "a salary range upside down", original: "soft_max: 9000", replacement: "soft_max: 500", field: "contact.salary"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			faults := faultsOf(test, changedFixture(test, fixtureText(test), testCase.original, testCase.replacement))
			if !slices.ContainsFunc(faults, func(fault Fault) bool { return fault.Field == testCase.field }) {
				test.Errorf("faults = %v, want one at %q", faults, testCase.field)
			}
		})
	}
}

func TestEveryUnknownKeyIsAFault(test *testing.T) {
	text := changedFixture(test, fixtureText(test), "headline:", "tagline:")
	text = changedFixture(test, text, "location: Brno, Czechia", "place: Brno, Czechia")
	faults := faultsOf(test, text)
	named := []bool{
		slices.ContainsFunc(faults, func(fault Fault) bool { return strings.Contains(fault.Problem, "tagline") }),
		slices.ContainsFunc(faults, func(fault Fault) bool { return strings.Contains(fault.Problem, "place") }),
	}
	if len(faults) != len(named) || slices.Contains(named, false) {
		test.Errorf("faults = %v, want one for each of %q and %q", faults, "tagline", "place")
	}
}

func TestEveryFaultIsReportedAtOnceOneALine(test *testing.T) {
	text := changedFixture(test, fixtureText(test), "name: Mira Example", `name: ""`)
	text = changedFixture(test, text, "level: warn", "level: bad")
	_, err := Load(fstest.MapFS{profileFile: {Data: []byte(text)}})
	want := "profile.yaml: person.name is empty\nprofile.yaml: health[0].checks[1].level is none of ok, warn, error"
	if err == nil || err.Error() != want {
		test.Errorf("Load = %v, want %q", err, want)
	}
}

func TestTheStoreKeepsTheLastGoodSnapshot(test *testing.T) {
	folder := fstest.MapFS{profileFile: {Data: []byte(fixtureText(test))}}
	store, err := Open(folder)
	if err != nil {
		test.Fatalf("Open(fixture) = %v, want no error", err)
	}
	folder[profileFile] = &fstest.MapFile{Data: []byte(changedFixture(test, fixtureText(test), "name: Mira Example", `name: ""`))}
	if err := store.Reload(); err == nil || store.Current().Profile.Person.Name != "Mira Example" {
		test.Errorf("a faulty reload = %v with the name %q, want faults and the old name kept", err, store.Current().Profile.Person.Name)
	}
	folder[profileFile] = &fstest.MapFile{Data: []byte(changedFixture(test, fixtureText(test), "name: Mira Example", "name: Mira Other"))}
	if err := store.Reload(); err != nil || store.Current().Profile.Person.Name != "Mira Other" {
		test.Errorf("a good reload = %v with the name %q, want the new name", err, store.Current().Profile.Person.Name)
	}
}

func fixtureText(test *testing.T) string {
	data, err := os.ReadFile(filepath.Join(fixtureFolder, profileFile))
	if err != nil {
		test.Fatalf("read the fixture: %v", err)
	}
	return string(data)
}

// changedFixture is a profile with one piece of it written otherwise; the piece must be there once.
func changedFixture(test *testing.T, text, original, replacement string) string {
	if strings.Count(text, original) != 1 {
		test.Fatalf("the profile holds %q %d times, want once", original, strings.Count(text, original))
	}
	return strings.Replace(text, original, replacement, 1)
}

// faultsOf loads a profile and hands back the faults it was refused with.
func faultsOf(test *testing.T, profileText string) Faults {
	_, err := Load(fstest.MapFS{profileFile: {Data: []byte(profileText)}})
	var faults Faults
	if !errors.As(err, &faults) {
		test.Fatalf("Load = %v, want faults", err)
	}
	return faults
}
