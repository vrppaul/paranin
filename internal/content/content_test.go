package content

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"unicode"
	"unicode/utf8"
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
		{name: "a skill put forward that no group lists", original: "put_forward: [Go, ", replacement: "put_forward: [Haskell, ", want: `profile.yaml: experience.put_forward[0] "Haskell" is no skill of a group`},
		{name: "a month past twelve", original: "start: 2023-04", replacement: "start: 2023-13", want: `profile.yaml: experience.roles[0].start "2023-13" is no month, YYYY-MM`},
		{name: "a role whose start is a gap", original: "start: 2023-04", replacement: `start: "[2023]"`, want: `profile.yaml: experience.roles[0].start "[2023]" is no month, YYYY-MM`},
		{name: "a role ended by the old word", original: "end: 2023-03", replacement: "end: present", want: `profile.yaml: experience.roles[1].end "present" is no month, YYYY-MM, nor current`},
		{name: "a role whose end is left empty", original: "end: 2023-03", replacement: "end:", want: `profile.yaml: experience.roles[1].end "" is no month, YYYY-MM, nor current`},
		{name: "a role that ends before it starts", original: "end: 2023-03", replacement: "end: 2019-01", want: `profile.yaml: experience.roles[1].end "2019-01" comes before the start, "2020-01"`},
		{name: "a slug with capitals", original: "slug: tidy-tool", replacement: "slug: Tidy-Tool", want: `profile.yaml: projects[1].slug "Tidy-Tool" is no slug: small letters and digits, joined by hyphens`},
		{name: "a slug two projects share", original: "slug: other-tool", replacement: "slug: tidy-tool", want: `profile.yaml: projects[1].slug "tidy-tool" names another project too`},
		{name: "a stack with no name", original: "stack: [Penlight, LPeg]", replacement: `stack: [Penlight, ""]`, want: "profile.yaml: projects[1].stack[1] is empty"},
		{name: "a stack naming the language", original: "stack: [Penlight, LPeg]", replacement: "stack: [Lua, LPeg]", want: `profile.yaml: projects[1].stack[0] "Lua" is the project's language already`},
		{name: "an install with no command", original: "command: go install example.test/tidy@latest", replacement: `command: ""`, want: "profile.yaml: projects[1].install.command is empty"},
		{name: "another way to install with no note", original: "note: On a machine with no Go\n          ", replacement: "", want: "profile.yaml: projects[1].install.alternatives[0].note is empty"},
		{name: "another way to install with no command", original: "command: curl -fsSL https://example.test/tidy.sh | sh", replacement: `command: ""`, want: "profile.yaml: projects[1].install.alternatives[0].command is empty"},
		{name: "a release with no changelog", original: "\n      changelog: https://github.test/mira/tidy/blob/HEAD/CHANGELOG.md", replacement: "", want: `profile.yaml: projects[1].release.changelog "" is no absolute web address`},
		{name: "a release with no newest", original: "latest: https://github.test/mira/tidy/releases/latest\n      ", replacement: "", want: `profile.yaml: projects[1].release.latest "" is no absolute web address`},
		{name: "a group with no title", original: "title: Tools", replacement: `title: ""`, want: "profile.yaml: skill_groups[1].title is empty"},
		{name: "a group written twice", original: "title: Tools", replacement: "title: Languages", want: `profile.yaml: skill_groups[1].title "Languages" names another group too`},
		{name: "a group with no skills", original: "    skills:\n      - {name: Docker, duration: side, sources: [Northwind], learning: true}\n", replacement: "    skills: []\n", want: "profile.yaml: skill_groups[1].skills holds no skill"},
		{name: "a skill with no name", original: "name: Docker", replacement: `name: ""`, want: "profile.yaml: skill_groups[1].skills[0].name is empty"},
		{name: "a skill another group names", original: "name: Docker", replacement: "name: Go", want: `profile.yaml: skill_groups[1].skills[0].name "Go" names another skill too`},
		{name: "a duration of no span", original: "duration: 3y", replacement: "duration: three years", want: `profile.yaml: skill_groups[0].skills[1].duration "three years" is no span, as 1y 7m, nor side, learning or a [gap]`},
		{name: "a span of months before years", original: "duration: 3y", replacement: "duration: 7m 1y", want: `profile.yaml: skill_groups[0].skills[1].duration "7m 1y" is no span, as 1y 7m, nor side, learning or a [gap]`},
		{name: "a span of no months", original: "duration: 3y", replacement: "duration: 0m", want: `profile.yaml: skill_groups[0].skills[1].duration "0m" is no span, as 1y 7m, nor side, learning or a [gap]`},
		{name: "a span with more after its months", original: "duration: 3y", replacement: "duration: 1y 7mo", want: `profile.yaml: skill_groups[0].skills[1].duration "1y 7mo" is no span, as 1y 7m, nor side, learning or a [gap]`},
		{name: "a duration left empty", original: "duration: 3y", replacement: `duration: ""`, want: `profile.yaml: skill_groups[0].skills[1].duration "" is no span, as 1y 7m, nor side, learning or a [gap]`},
		{name: "a duration left out", original: "duration: 3y, ", replacement: "", want: `profile.yaml: skill_groups[0].skills[1].duration "" is no span, as 1y 7m, nor side, learning or a [gap]`},
		{name: "a span with no space between years and months", original: "duration: 3y", replacement: "duration: 1y7m", want: `profile.yaml: skill_groups[0].skills[1].duration "1y7m" is no span, as 1y 7m, nor side, learning or a [gap]`},
		{name: "a span with a space after it", original: "duration: 3y", replacement: `duration: "1y "`, want: `profile.yaml: skill_groups[0].skills[1].duration "1y " is no span, as 1y 7m, nor side, learning or a [gap]`},
		{name: "a span with a sign", original: "duration: 3y", replacement: "duration: +1y", want: `profile.yaml: skill_groups[0].skills[1].duration "+1y" is no span, as 1y 7m, nor side, learning or a [gap]`},
		{name: "a span with a nought before its count", original: "duration: 3y", replacement: "duration: 01y", want: `profile.yaml: skill_groups[0].skills[1].duration "01y" is no span, as 1y 7m, nor side, learning or a [gap]`},
		{name: "a span of a year's months after years", original: "duration: 3y", replacement: "duration: 1y 12m", want: `profile.yaml: skill_groups[0].skills[1].duration "1y 12m" is no span, as 1y 7m, nor side, learning or a [gap]`},
		{name: "a span of a year's months alone", original: "duration: 3y", replacement: "duration: 12m", want: `profile.yaml: skill_groups[0].skills[1].duration "12m" is no span, as 1y 7m, nor side, learning or a [gap]`},
		{name: "a skill still being learnt by its duration alone", original: "duration: 3y", replacement: "duration: learning", want: `profile.yaml: skill_groups[0].skills[1].duration "learning" says the skill is still being learnt, but learning is not true`},
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

// A face passes text on as the content writes it, so a character a terminal or a browser acts on rather than shows
// is a fault at the field that holds it, wherever the file writes it, as it is or by an escape; a line feed is text.
func TestACharacterAReaderWouldActOnIsAFaultAtItsField(test *testing.T) {
	cases := []struct {
		name        string
		file        string
		original    string
		replacement string
		field       string
		value       string
	}{
		{
			name: "an escape that links and clears the screen", file: profileFile, original: "name: Mira Example",
			replacement: `name: "Mira \e]8;;https://evil.test\e\\Example\e[2J"`, field: "person.name",
			value: "Mira \x1b]8;;https://evil.test\x1b\\Example\x1b[2J",
		},
		{
			name: "an escape written as binary", file: profileFile, original: "name: Mira Example",
			replacement: "name: !!binary TWlyYSAbXTg7O2h0dHBzOi8vZXZpbC50ZXN0G1xFeGFtcGxlG1sySg==", field: "person.name",
			value: "Mira \x1b]8;;https://evil.test\x1b\\Example\x1b[2J",
		},
		{name: "a byte of no character written as binary", file: profileFile, original: "name: Mira Example", replacement: "name: !!binary mw==", field: "person.name", value: "\x9b"},
		{name: "a bell in a point", file: profileFile, original: `- "Kept the ==queue== moving"`, replacement: `- "Kept the ==queue== moving\a"`, field: "experience.roles[0].bullets[0]", value: "Kept the ==queue== moving\a"},
		{name: "a nul in advice", file: profileFile, original: "advice: look at it", replacement: `advice: "look\0 at it"`, field: "health[0].checks[1].advice", value: "look\x00 at it"},
		{name: "a tab in a skill's name", file: profileFile, original: "name: Docker,", replacement: `name: "Doc\tker",`, field: "skill_groups[1].skills[0].name", value: "Doc\tker"},
		{name: "a carriage return in a description", file: profileFile, original: "description: A tool that ==tidies files==.", replacement: `description: "A tool that\r\n==tidies files==."`, field: "projects[1].description", value: "A tool that\r\n==tidies files==."},
		{name: "a delete in a fact", file: profileFile, original: "value: 5 years", replacement: `value: "5 years\x7f"`, field: "facts[0].value", value: "5 years\x7f"},
		{name: "a control of the upper half in the summary", file: profileFile, original: "summary: Writes software for people who read it.", replacement: `summary: "Writes software\u009b2J"`, field: "summary", value: "Writes software\u009b2J"},
		{name: "an override of direction in a name", file: profileFile, original: "name: tidy-tool", replacement: `name: "tidy‮-tool"`, field: "projects[1].name", value: "tidy‮-tool"},
		{name: "an isolate written as it is in a stack", file: profileFile, original: "LPeg]", replacement: "LP⁦eg]", field: "projects[1].stack[1]", value: "LP⁦eg"},
		{name: "a mark of direction in the headline", file: profileFile, original: "headline: Builds", replacement: "headline: ‏Builds", field: "person.headline", value: "‏Builds small tools that last."},
		{name: "an isolate in a word of the site", file: wordsFile(english), original: "about: About", replacement: `about: "Ab⁧out"`, field: "about", value: "Ab⁧out"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			faults := faultsOf(test, changedFile(test, fixtureFolder(test), testCase.file, testCase.original, testCase.replacement))
			want := Faults{{File: testCase.file, Field: testCase.field, Problem: fmt.Sprintf(characterProblem, testCase.value)}}
			if !slices.Equal(faults, want) {
				test.Errorf("faults with %q = %q, want %q alone", testCase.replacement, faults, want)
			}
		})
	}
}

// A fault's line reaches a terminal, and a key or the decoder's words in it come from the file: each fault stands on a
// line of its own, holding no character a terminal acts on and no byte of no character, each written as %q writes it.
func TestAFaultsLineHoldsNothingATerminalActsOn(test *testing.T) {
	cases := []struct {
		name        string
		file        string
		original    string
		replacement string
		shown       string
	}{
		{name: "an escape in a key the site has no word for", file: wordsFile(english), original: "about: About", replacement: "about: About\n\"evil\\e[2J\": \"x\\e\"", shown: `\x1b`},
		{name: "a line feed in a key the site has no word for", file: wordsFile(english), original: "about: About", replacement: "about: About\n\"ev\\nil\": x", shown: `\n`},
		{name: "a byte of no character in a key", file: wordsFile(english), original: "about: About", replacement: "about: About\n!!binary mw==: x", shown: `\x9b`},
		{name: "a replacement character in a key, which is text", file: wordsFile(english), original: "about: About", replacement: "about: About\n\"ev\\uFFFDil\": x", shown: "ev�il"},
		{name: "an escape in a key no type knows", file: profileFile, original: "headline:", replacement: "\"head\\e[2Jline\":", shown: `\x1b`},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			faults := faultsOf(test, changedFile(test, fixtureFolder(test), testCase.file, testCase.original, testCase.replacement))
			lines := strings.Split(faults.Error(), "\n")
			if len(lines) != len(faults) || slices.ContainsFunc(lines, actedOnByATerminal) || !strings.Contains(faults.Error(), testCase.shown) {
				test.Errorf("faults with %q write %q, want %d lines with no control character, mark of direction or byte of no character, showing %q",
					testCase.replacement, lines, len(faults), testCase.shown)
			}
		})
	}
}

// actedOnByATerminal is whether a line holds a character a terminal acts on rather than shows, or bytes of no
// character, which a terminal may read as one.
func actedOnByATerminal(line string) bool {
	return !utf8.ValidString(line) || strings.ContainsFunc(line, func(character rune) bool {
		return unicode.IsControl(character) || unicode.Is(unicode.Bidi_Control, character)
	})
}

// How long a skill was used reads as what the profile writes: a span, years before months, as the months it comes to;
// a skill used on the side, one still being learnt, by its duration and its learning alike, or a gap still to fill,
// as one of no months.
func TestASkillsDurationReadsAsItsKindAndMonths(test *testing.T) {
	// Go's line of the fixture, whose duration each case but the one on Rust's line writes anew
	const goDuration = "name: Go, duration: 3y"
	kindNames := map[DurationKind]string{
		durationUnread: "unread", DurationMeasured: "span", DurationSide: "side", DurationLearning: "learning", DurationGap: "gap",
	}
	cases := []struct {
		name        string
		original    string
		replacement string
		skill       string
		kind        DurationKind
		months      int
	}{
		{name: "years", original: goDuration, replacement: goDuration, skill: "Go", kind: DurationMeasured, months: 36},
		{name: "years and months", original: goDuration, replacement: "name: Go, duration: 1y 7m", skill: "Go", kind: DurationMeasured, months: 19},
		{name: "years of two counts and months", original: goDuration, replacement: "name: Go, duration: 10y 11m", skill: "Go", kind: DurationMeasured, months: 131},
		{name: "months", original: goDuration, replacement: "name: Go, duration: 11m", skill: "Go", kind: DurationMeasured, months: 11},
		{name: "on the side", original: goDuration, replacement: "name: Go, duration: side", skill: "Go", kind: DurationSide, months: 0},
		{name: "a gap", original: goDuration, replacement: `name: Go, duration: "[2y?]"`, skill: "Go", kind: DurationGap, months: 0},
		{name: "still being learnt", original: "name: Rust", replacement: "name: Rust", skill: "Rust", kind: DurationLearning, months: 0},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			snapshot, err := Load(changedProfile(test, testCase.original, testCase.replacement))
			if err != nil {
				test.Fatalf("Load with %q = %v, want no error", testCase.replacement, err)
			}
			duration := skillNamed(test, snapshot, testCase.skill).Duration
			if duration.Kind() != testCase.kind || duration.Months() != testCase.months {
				test.Errorf("%s's duration = a %s of %d months, want a %s of %d", testCase.skill, kindNames[duration.Kind()],
					duration.Months(), kindNames[testCase.kind], testCase.months)
			}
		})
	}
}

// skillNamed is the skill of a name, in whichever group the profile lists it.
func skillNamed(test *testing.T, snapshot *Snapshot, name string) Skill {
	for _, group := range snapshot.Profile.SkillGroups {
		for _, skill := range group.Skills {
			if skill.Name == name {
				return skill
			}
		}
	}
	test.Fatalf("the profile lists no skill %q", name)
	return Skill{Name: name, Duration: Duration{}, Sources: nil, Learning: false}
}

// The site's tag authority (RFC 4151) takes one form: a domain, a comma, and a day its author held it on, a year or a
// date. A profile with none, or one of another form RFC 4151 has, a month or a mail address, is refused.
func TestATagAuthorityIsADomainAndAYearOrADate(test *testing.T) {
	cases := []struct {
		name      string
		authority string
		loads     bool
	}{
		{name: "a domain and a year", authority: "example.test,2026", loads: true},
		{name: "a domain and a date", authority: "example.test,2026-11-24", loads: true},
		{name: "none", authority: "", loads: false},
		{name: "a domain alone", authority: "example.test", loads: false},
		{name: "a domain and a month", authority: "example.test,2026-11", loads: false},
		{name: "a domain and a date of no day", authority: "example.test,2026-13-01", loads: false},
		{name: "a mail address", authority: "mira@example.test,2026", loads: false},
		{name: "a web address", authority: "https://example.test,2026", loads: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			_, err := Load(changedProfile(test, "tag_authority: example.test,2026", "tag_authority: "+testCase.authority))
			if loads := err == nil; loads != testCase.loads {
				test.Errorf("a profile whose tag authority is %q loads: %t (%v), want %t", testCase.authority, loads, err, testCase.loads)
			}
		})
	}
}

// The host a terminal reaches the site at over ssh is a domain alone, which every ssh command the site writes names: a
// command, a user or a port written into it is refused, as is none.
func TestTheSSHHostIsADomain(test *testing.T) {
	cases := []struct {
		name  string
		host  string
		loads bool
	}{
		{name: "a domain", host: "example.test", loads: true},
		{name: "none", host: `""`, loads: false},
		{name: "a command", host: "ssh example.test", loads: false},
		{name: "a user", host: "mira@example.test", loads: false},
		{name: "a port", host: "example.test:2222", loads: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			_, err := Load(changedProfile(test, "ssh_host: example.test", "ssh_host: "+testCase.host))
			if loads := err == nil; loads != testCase.loads {
				test.Errorf("a profile whose ssh host is %q loads: %t (%v), want %t", testCase.host, loads, err, testCase.loads)
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
	folder := changedProfile(test, "ssh_host: example.test", `ssh_host: ""`)
	_, err := Load(changedFile(test, folder, profileFile, "title: Software Engineer", `title: ""`))
	want := "profile.yaml: site.ssh_host is empty\nprofile.yaml: person.title is empty"
	if err == nil || err.Error() != want {
		test.Errorf("Load = %v, want %q", err, want)
	}
}

// The site's words are content too: each one is there, in every language, and a catalogue holds no word the site has
// not.
func TestEveryWordOfTheSiteIsRequired(test *testing.T) {
	cases := []struct {
		name        string
		file        string
		original    string
		replacement string
		want        string
	}{
		{name: "a word left empty", file: wordsFile(english), original: "experience: Experience", replacement: `experience: ""`, want: "i18n/en.yaml: experience is empty"},
		{name: "a word left out", file: wordsFile(english), original: "experience: Experience\n", replacement: "", want: "i18n/en.yaml: experience is empty"},
		{name: "a word left out of a translation's words", file: wordsFile("ru"), original: "experience: Опыт\n", replacement: "", want: "i18n/ru.yaml: experience is empty"},
		{name: "a word the site has not", file: wordsFile(english), original: "about: About", replacement: "about: About\nsummary: Summary", want: "i18n/en.yaml: summary is no word of the site"},
		{
			name: "a phrase with no place for its value", file: wordsFile(english), original: `copied_command: "Copied {}"`, replacement: "copied_command: Copied",
			want: `i18n/en.yaml: copied_command "Copied" writes its place, {}, 0 times, where its values ask for 1`,
		},
		{
			name: "a phrase with two places", file: wordsFile(english), original: `copied_command: "Copied {}"`, replacement: `copied_command: "Copied {} {}"`,
			want: `i18n/en.yaml: copied_command "Copied {} {}" writes its place, {}, 2 times, where its values ask for 1`,
		},
		{
			name: "a phrase of two values with one place", file: wordsFile(english), original: `part_of: "{} of {}"`, replacement: `part_of: "{} of"`,
			want: `i18n/en.yaml: part_of "{} of" writes its place, {}, 1 times, where its values ask for 2`,
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			faults := faultsOf(test, changedFile(test, fixtureFolder(test), testCase.file, testCase.original, testCase.replacement))
			if !slices.ContainsFunc(faults, func(fault Fault) bool { return fault.String() == testCase.want }) {
				test.Errorf("faults = %q, want one %q", faults, testCase.want)
			}
		})
	}
}

// A count is written in the form its language's plural rule (CLDR) gives the number, the number in its first place:
// English writes two forms, Russian three for a whole count.
func TestACountIsWrittenInTheFormItsLanguageGivesTheNumber(test *testing.T) {
	snapshot, err := Load(fixtureFolder(test))
	if err != nil {
		test.Fatalf("Load of the fixture = %v, want no error", err)
	}
	wordsIn := map[string]Words{}
	for _, edition := range snapshot.Editions() {
		wordsIn[edition.Words.Language()] = edition.Words
	}
	cases := []struct {
		name     string
		language string
		count    Count
		number   int
		values   []string
		want     string
	}{
		{name: "one post in English", language: english, count: CountPosts, number: 1, values: nil, want: "1 post"},
		{name: "no post in English", language: english, count: CountPosts, number: 0, values: nil, want: "0 posts"},
		{name: "two posts in English", language: english, count: CountPosts, number: 2, values: nil, want: "2 posts"},
		{name: "one post in Russian", language: russian, count: CountPosts, number: 1, values: nil, want: "1 пост"},
		{name: "three posts in Russian", language: russian, count: CountPosts, number: 3, values: nil, want: "3 поста"},
		{name: "five posts in Russian", language: russian, count: CountPosts, number: 5, values: nil, want: "5 постов"},
		{name: "eleven posts in Russian", language: russian, count: CountPosts, number: 11, values: nil, want: "11 постов"},
		{name: "twenty-one posts in Russian", language: russian, count: CountPosts, number: 21, values: nil, want: "21 пост"},
		{name: "twenty-two posts in Russian", language: russian, count: CountPosts, number: 22, values: nil, want: "22 поста"},
		{name: "a part of two out in Russian", language: russian, count: CountPartsOut, number: 1, values: []string{"2"}, want: "вышла 1 из 2"},
		{name: "every part of a series of one", language: english, count: CountAllParts, number: 1, values: nil, want: "1 part"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := wordsIn[testCase.language].Count(testCase.count, testCase.number, testCase.values...); got != testCase.want {
				test.Errorf("Count(%d, %d, %q) in %s = %q, want %q", testCase.count, testCase.number, testCase.values, testCase.language, got, testCase.want)
			}
		})
	}
}

// Every fault of a catalogue's count or months stands at its key, and at the form or the month it names: a form the
// language's rule gives a whole count and the catalogue leaves out, or one the rule never gives.
func TestEveryFaultOfACatalogueStandsAtItsPlace(test *testing.T) {
	cases := []struct {
		name        string
		file        string
		original    string
		replacement string
		want        faultPlace
	}{
		{
			name: "a count written as one text", file: wordsFile(english), original: "posts_count:\n  one: \"{} post\"\n  other: \"{} posts\"",
			replacement: `posts_count: "{} posts"`, want: faultPlace{file: wordsFile(english), field: "posts_count", line: 0},
		},
		{
			name: "a form its language gives left out", file: wordsFile(russian), original: "  few: \"{} поста\"\n", replacement: "",
			want: faultPlace{file: wordsFile(russian), field: "posts_count.few", line: 0},
		},
		{
			name: "a form its language never gives a whole count", file: wordsFile(english), original: "  other: \"{} posts\"",
			replacement: "  other: \"{} posts\"\n  few: \"{} posts\"", want: faultPlace{file: wordsFile(english), field: "posts_count.few", line: 0},
		},
		{
			name: "a form of no name CLDR gives", file: wordsFile(english), original: "  other: \"{} posts\"",
			replacement: "  other: \"{} posts\"\n  several: \"{} posts\"", want: faultPlace{file: wordsFile(english), field: "posts_count.several", line: 0},
		},
		{
			name: "a form with no place for its count", file: wordsFile(english), original: `one: "{} post"`, replacement: `one: "a post"`,
			want: faultPlace{file: wordsFile(english), field: "posts_count.one", line: 0},
		},
		{
			name: "a word written as forms", file: wordsFile(english), original: "about: About", replacement: "about: {one: About}",
			want: faultPlace{file: wordsFile(english), field: "about", line: 0},
		},
		{
			name: "eleven months", file: wordsFile(english), original: "Nov, Dec]", replacement: "Nov]",
			want: faultPlace{file: wordsFile(english), field: "months", line: 0},
		},
		{
			name: "a month left empty", file: wordsFile(english), original: "Apr,", replacement: `"",`,
			want: faultPlace{file: wordsFile(english), field: "months[3]", line: 0},
		},
		{
			name: "months written as one text", file: wordsFile(english), original: "months: [Jan, Feb, Mar, Apr, May, Jun, Jul, Aug, Sep, Oct, Nov, Dec]",
			replacement: "months: Jan", want: faultPlace{file: wordsFile(english), field: "months", line: 0},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			faults := faultsOf(test, changedFile(test, fixtureFolder(test), testCase.file, testCase.original, testCase.replacement))
			if !slices.ContainsFunc(faults, func(fault Fault) bool { return placeOf(fault) == testCase.want }) {
				test.Errorf("faults with %q = %q, want one at %+v", testCase.replacement, faults, testCase.want)
			}
		})
	}
}

// A catalogue is named by its language's tag, which an address of the site writes before a translated page's path:
// a file named otherwise is a fault of its own.
func TestACatalogueIsNamedByItsLanguage(test *testing.T) {
	cases := []struct {
		name string
		file string
	}{
		{name: "a language written whole", file: "i18n/russian.yaml"},
		{name: "a language with its region", file: "i18n/pt-br.yaml"},
		{name: "a tag in capitals", file: "i18n/RU.yaml"},
		{name: "a file of no catalogue", file: "i18n/notes.txt"},
		{name: "a tag of no language", file: "i18n/xx.yaml"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			faults := faultsOf(test, writtenFile(fixtureFolder(test), testCase.file, "about: About\n"))
			want := faultPlace{file: testCase.file, field: "", line: 0}
			if !slices.ContainsFunc(faults, func(fault Fault) bool { return placeOf(fault) == want }) {
				test.Errorf("faults with %s = %q, want one at the file", testCase.file, faults)
			}
		})
	}
}

// fixtureFolder is the fixture content in a folder in memory, which a test may change.
func fixtureFolder(test *testing.T) fstest.MapFS {
	copied := fixtureCopy{source: os.DirFS(fixturePath), folder: fstest.MapFS{}}
	if err := fs.WalkDir(copied.source, ".", copied.visit); err != nil {
		test.Fatalf("read the fixture: %v", err)
	}
	return copied.folder
}

// fixtureCopy gathers the files of a fixture's folder into a folder in memory.
type fixtureCopy struct {
	source fs.FS
	folder fstest.MapFS
}

func (copied fixtureCopy) visit(file string, entry fs.DirEntry, err error) error {
	if err != nil || entry.IsDir() {
		return err
	}
	data, err := fs.ReadFile(copied.source, file)
	copied.folder[file] = &fstest.MapFile{Data: data}
	return err
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

// withoutFile is a folder with one of its files taken away.
func withoutFile(folder fstest.MapFS, file string) fstest.MapFS {
	delete(folder, file)
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
