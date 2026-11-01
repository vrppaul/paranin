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

// A fault is read by whoever edits the content: the line it writes is the behaviour. It names the field, and
// quotes what was written there where something was.
func TestEveryFaultSaysItsFieldAndProblem(test *testing.T) {
	cases := []struct {
		name        string
		original    string
		replacement string
		want        string
	}{
		{name: "an empty name", original: "name: Mira Example", replacement: `name: ""`, want: "profile.yaml: person.name is empty"},
		{name: "a mail address with a name round it", original: "email: mira@example.test", replacement: "email: Mira <mira@example.test>", want: `profile.yaml: person.email "Mira <mira@example.test>" is no bare mail address`},
		{name: "a base address ending in a slash", original: "base_url: https://example.test", replacement: "base_url: https://example.test/", want: `profile.yaml: site.base_url "https://example.test/" ends in a slash, which every address adds itself`},
		{name: "a link that is no address", original: "url: https://github.test/mira\n", replacement: "url: github mira\n", want: `profile.yaml: person.links[0].url "github mira" is no absolute web address`},
		{name: "a month past twelve", original: "start: 2023-04", replacement: "start: 2023-13", want: `profile.yaml: experience[0].start "2023-13" is no month, YYYY-MM`},
		{name: "a role whose start is a gap", original: "start: 2023-04", replacement: `start: "[2023]"`, want: `profile.yaml: experience[0].start "[2023]" is no month, YYYY-MM`},
		{name: "a role ended by the old word", original: "end: 2023-03", replacement: "end: present", want: `profile.yaml: experience[1].end "present" is no month, YYYY-MM, nor current`},
		{name: "a role whose end is left empty", original: "end: 2023-03", replacement: "end:", want: `profile.yaml: experience[1].end "" is no month, YYYY-MM, nor current`},
		{name: "a role that ends before it starts", original: "end: 2023-03", replacement: "end: 2019-01", want: `profile.yaml: experience[1].end "2019-01" comes before the start, "2020-01"`},
		{name: "a school month neither a month nor a gap", original: `end: "[20XX]"`, replacement: "end: later", want: `profile.yaml: education[0].end "later" is no month, YYYY-MM, nor a [gap]`},
		{name: "a school that ends before it starts", original: `end: "[20XX]"`, replacement: "end: 2010-01", want: `profile.yaml: education[0].end "2010-01" comes before the start, "2016-09"`},
		{name: "a slug with capitals", original: "slug: tidy-tool", replacement: "slug: Tidy-Tool", want: `profile.yaml: projects[1].slug "Tidy-Tool" is no slug: small letters and digits, joined by hyphens`},
		{name: "a slug two projects share", original: "slug: other-tool", replacement: "slug: tidy-tool", want: `profile.yaml: projects[1].slug "tidy-tool" names another project too`},
		{name: "a group with no title", original: "title: Tools", replacement: `title: ""`, want: "profile.yaml: skill_groups[1].title is empty"},
		{name: "a group written twice", original: "title: Tools", replacement: "title: Languages", want: `profile.yaml: skill_groups[1].title "Languages" names another group too`},
		{name: "a group with no skills", original: "    skills:\n      - {name: Docker, duration: side, aliases: [docker], sources: [Northwind], learning: true}\n", replacement: "    skills: []\n", want: "profile.yaml: skill_groups[1].skills holds no skill"},
		{name: "a skill with no name", original: "name: Docker", replacement: `name: ""`, want: "profile.yaml: skill_groups[1].skills[0].name is empty"},
		{name: "a level of no kind", original: "level: warn", replacement: "level: bad", want: `profile.yaml: health[0].checks[1].level "bad" is none of error, ok, warn`},
		{name: "a salary from nothing", original: "min: 1000", replacement: "min: 0", want: "profile.yaml: contact.salary.min 0 is not above 0"},
		{name: "a salary range upside down", original: "soft_max: 9000", replacement: "soft_max: 500", want: "profile.yaml: contact.salary.soft_max 500 is not above min, 1000"},
		{name: "a salary range of no width", original: "soft_max: 9000", replacement: "soft_max: 1000", want: "profile.yaml: contact.salary.soft_max 1000 is not above min, 1000"},
		{name: "a salary in no currency", original: "currencies: [EUR]", replacement: "currencies: []", want: "profile.yaml: contact.salary.currencies is empty"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			faults := faultsOf(test, changedFixture(test, fixtureText(test), testCase.original, testCase.replacement))
			if !slices.ContainsFunc(faults, func(fault Fault) bool { return fault.String() == testCase.want }) {
				test.Errorf("faults = %q, want one %q", faults, testCase.want)
			}
		})
	}
}

// A profile missing or empty is a fault of the file, which is named once.
func TestAMissingOrEmptyProfileIsAFaultOfTheFile(test *testing.T) {
	cases := []struct {
		name   string
		folder fstest.MapFS
		want   string
	}{
		{name: "no profile in the folder", folder: fstest.MapFS{}, want: "profile.yaml: file does not exist"},
		{name: "an empty profile", folder: fstest.MapFS{profileFile: {Data: []byte("")}}, want: "profile.yaml: is empty"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if _, err := Load(testCase.folder); err == nil || err.Error() != testCase.want {
				test.Errorf("Load = %v, want %q", err, testCase.want)
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
	text := changedFixture(test, fixtureText(test), "ssh: ssh example.test", `ssh: ""`)
	text = changedFixture(test, text, "title: Software Engineer", `title: ""`)
	_, err := Load(fstest.MapFS{profileFile: {Data: []byte(text)}})
	want := "profile.yaml: site.ssh is empty\nprofile.yaml: person.title is empty"
	if err == nil || err.Error() != want {
		test.Errorf("Load = %v, want %q", err, want)
	}
}

func TestASiteCannotOpenOnFaultyContent(test *testing.T) {
	folder := fstest.MapFS{profileFile: {Data: []byte(changedFixture(test, fixtureText(test), "name: Mira Example", `name: ""`))}}
	if store, err := Open(folder); store != nil || err == nil {
		test.Errorf("Open(faulty) = %v, %v; want no store and faults", store, err)
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
