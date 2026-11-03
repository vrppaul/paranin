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

// The fixture content, and the repository's own, from this package
const (
	fixturePath      = "testdata/content"
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
		{name: "a base address of no web scheme", original: "base_url: https://example.test", replacement: "base_url: ftp://example.test", want: `profile.yaml: site.base_url "ftp://example.test" is no absolute web address`},
		{name: "a base address of no host", original: "base_url: https://example.test", replacement: "base_url: https:///cv", want: `profile.yaml: site.base_url "https:///cv" is no absolute web address`},
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
			faults := faultsOf(test, changedProfile(test, testCase.original, testCase.replacement))
			if !slices.ContainsFunc(faults, func(fault Fault) bool { return fault.String() == testCase.want }) {
				test.Errorf("faults = %q, want one %q", faults, testCase.want)
			}
		})
	}
}

// The site's base address is read once, as the content writes it, and keeps its host, which names the site.
func TestTheBaseAddressKeepsItsHost(test *testing.T) {
	cases := []struct {
		name string
		base string
		want string
	}{
		{name: "a bare address", base: "https://example.test", want: "example.test"},
		{name: "an address with a path", base: "https://example.test/cv", want: "example.test"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			snapshot, err := Load(changedProfile(test, "base_url: https://example.test", "base_url: "+testCase.base))
			if err != nil {
				test.Fatalf("Load with base_url %q = %v, want no error", testCase.base, err)
			}
			if address := snapshot.Profile.Site.BaseURL; address.Written() != testCase.base || address.Host() != testCase.want {
				test.Errorf("base_url %q reads as %q of the host %q, want it as written, of the host %q", testCase.base,
					address.Written(), address.Host(), testCase.want)
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
		{name: "no profile in the folder", folder: fstest.MapFS{wordsFile(english): fixtureFolder(test)[wordsFile(english)]}, want: "profile.yaml: file does not exist"},
		{name: "an empty profile", folder: writtenFile(fixtureFolder(test), profileFile, ""), want: "profile.yaml: is empty"},
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
	folder := changedProfile(test, "headline:", "tagline:")
	faults := faultsOf(test, changedFile(test, folder, profileFile, "location: Brno, Czechia", "place: Brno, Czechia"))
	named := []bool{
		slices.ContainsFunc(faults, func(fault Fault) bool { return strings.Contains(fault.Problem, "tagline") }),
		slices.ContainsFunc(faults, func(fault Fault) bool { return strings.Contains(fault.Problem, "place") }),
	}
	if len(faults) != len(named) || slices.Contains(named, false) {
		test.Errorf("faults = %v, want one for each of %q and %q", faults, "tagline", "place")
	}
}

func TestEveryFaultIsReportedAtOnceOneALine(test *testing.T) {
	folder := changedProfile(test, "ssh: ssh example.test", `ssh: ""`)
	_, err := Load(changedFile(test, folder, profileFile, "title: Software Engineer", `title: ""`))
	want := "profile.yaml: site.ssh is empty\nprofile.yaml: person.title is empty"
	if err == nil || err.Error() != want {
		test.Errorf("Load = %v, want %q", err, want)
	}
}

// The site's words are content too: each one is there, and the catalogue holds no word the site has not.
func TestEveryWordOfTheSiteIsRequired(test *testing.T) {
	cases := []struct {
		name        string
		original    string
		replacement string
		want        string
	}{
		{name: "a word left empty", original: "experience: Experience", replacement: `experience: ""`, want: "i18n/en.yaml: experience is empty"},
		{name: "a word left out", original: "experience: Experience\n", replacement: "", want: "i18n/en.yaml: experience is empty"},
		{name: "a word the site has not", original: "facts: Facts", replacement: "facts: Facts\nsummary: Summary", want: "i18n/en.yaml: summary is no word of the site"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			faults := faultsOf(test, changedFile(test, fixtureFolder(test), wordsFile(english), testCase.original, testCase.replacement))
			if !slices.ContainsFunc(faults, func(fault Fault) bool { return fault.String() == testCase.want }) {
				test.Errorf("faults = %q, want one %q", faults, testCase.want)
			}
		})
	}
}

// fixtureFolder is the fixture content in a folder in memory, which a test may change.
func fixtureFolder(test *testing.T) fstest.MapFS {
	folder := fstest.MapFS{}
	for _, file := range []string{profileFile, wordsFile(english)} {
		data, err := os.ReadFile(filepath.Join(fixturePath, file))
		if err != nil {
			test.Fatalf("read the fixture's %s: %v", file, err)
		}
		folder[file] = &fstest.MapFile{Data: data}
	}
	return folder
}

// changedFile writes one piece of a file of the folder otherwise; the piece must be there once.
func changedFile(test *testing.T, folder fstest.MapFS, file, original, replacement string) fstest.MapFS {
	text := string(folder[file].Data)
	if strings.Count(text, original) != 1 {
		test.Fatalf("%s holds %q %d times, want once", file, original, strings.Count(text, original))
	}
	return writtenFile(folder, file, strings.Replace(text, original, replacement, 1))
}

// changedProfile is the fixture content with one piece of its profile written otherwise.
func changedProfile(test *testing.T, original, replacement string) fstest.MapFS {
	return changedFile(test, fixtureFolder(test), profileFile, original, replacement)
}

func writtenFile(folder fstest.MapFS, file, text string) fstest.MapFS {
	folder[file] = &fstest.MapFile{Data: []byte(text)}
	return folder
}

// faultsOf loads a folder and hands back the faults it was refused with.
func faultsOf(test *testing.T, folder fstest.MapFS) Faults {
	_, err := Load(folder)
	var faults Faults
	if !errors.As(err, &faults) {
		test.Fatalf("Load = %v, want faults", err)
	}
	return faults
}
