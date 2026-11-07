package tui

import (
	"math"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/vrppaul/paranin/internal/document"
)

// :Lazy shows the list of skills in place of the résumé, as lazy.nvim lists its plugins: each skill after a mark of
// whether the author uses it or is still learning it.
func TestLazyListsEverySkill(test *testing.T) {
	screen := screenOf(test, commanded(test, fixtureModel(test, commonWidth, commonHeight), "Lazy"))
	if strings.Contains(screen, resumeHeadline) {
		test.Errorf("after :Lazy the screen =\n%s\nwant the résumé gone", screen)
	}
	cases := []struct {
		name string
		want string
	}{
		{name: "a skill in use", want: "● Go"},
		{name: "a skill still being learnt", want: "○ Rust"},
		{name: "another skill still being learnt", want: "○ Docker"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !strings.Contains(screen, testCase.want) {
				test.Errorf("after :Lazy the screen =\n%s\nwant %q", screen, testCase.want)
			}
		})
	}
}

// Each skill's months stand as a bar after its name, on one scale for the whole list, how long as the profile writes
// it after the bar: a skill used for a share of the longest one's months has a bar of that share of its cells, to the
// nearest cell.
func TestEachSkillsBarIsAsLongAsItsMonths(test *testing.T) {
	// the fixture's two skills of a measured time, written 3y and 2y
	threeYears, twoYears := measuredSkill{written: "3y", months: 36}, measuredSkill{written: "2y", months: 24}
	sevenYears := document.Skills{List: []document.Skill{
		{Name: textOf("Python"), Duration: textOf("7y"), Months: 84, Sources: textOf("Northwind"), Learning: false},
		{Name: textOf("Kotlin"), Duration: textOf("1y 7m"), Months: 19, Sources: textOf("Southwind"), Learning: false},
	}}
	cases := []struct {
		name    string
		model   Model
		longer  measuredSkill
		shorter measuredSkill
	}{
		{name: "the fixture's list", model: commanded(test, fixtureModel(test, commonWidth, tallHeight), "Lazy"), longer: threeYears, shorter: twoYears},
		{name: "the fixture's list on a narrow page", model: commanded(test, fixtureModel(test, narrowWidth, tallHeight), "Lazy"), longer: threeYears, shorter: twoYears},
		{name: "a list of short names on a wide page", model: handResume(test, "Mira", wideWidth, tallHeight, shortSkills()), longer: threeYears, shorter: twoYears},
		{
			name: "a share of no whole cells", model: handResume(test, "Mira", wideWidth, tallHeight, sevenYears),
			longer: measuredSkill{written: "7y", months: 84}, shorter: measuredSkill{written: "1y 7m", months: 19},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			read := readRows(aboveTheStatusLine(screenOf(test, testCase.model)))
			longer, shorter := barCells(read, space+testCase.longer.written), barCells(read, space+testCase.shorter.written)
			share := float64(testCase.shorter.months) / float64(testCase.longer.months)
			if want := int(math.Round(float64(longer) * share)); longer == 0 || shorter != want {
				test.Errorf("on the page\n%s\nthe bar of %s is %d cells, of %s %d, want it %d", read, testCase.longer.written,
					longer, testCase.shorter.written, shorter, want)
			}
		})
	}
}

// A wider page never draws the bars shorter: they stand beside the names only where they keep the length they would
// have under them.
func TestAWiderPageNeverDrawsShorterBars(test *testing.T) {
	longest := 0
	for width := narrowWidth; width <= commonWidth; width++ {
		read := readRows(aboveTheStatusLine(screenOf(test, commanded(test, fixtureModel(test, width, tallHeight), "Lazy"))))
		cells := barCells(read, " 3y")
		if cells < longest {
			test.Fatalf("at %d columns the longest bar is %d cells, shorter than the %d of a narrower page:\n%s", width, cells,
				longest, read)
		}
		longest = cells
	}
	if longest == 0 {
		test.Errorf("no page from %d to %d columns draws a bar before 3y", narrowWidth, commonWidth)
	}
}

// measuredSkill is how long a skill of a test's list was used: as the list writes it, and in months.
type measuredSkill struct {
	written string
	months  int
}

// The bars stand on the scale of the list's longest skill, whatever its months: a list of twice the months draws the
// same bars, and a span however short draws one, of a cell at least.
func TestTheBarsStandOnTheScaleOfTheLongestSkill(test *testing.T) {
	doubled := document.Skills{List: []document.Skill{
		{Name: textOf("Go"), Duration: textOf("6y"), Months: 72, Sources: textOf("tidy-tool"), Learning: false},
		{Name: textOf("Make"), Duration: textOf("4y"), Months: 48, Sources: textOf("Northwind"), Learning: false},
		{Name: textOf("Lua"), Duration: []document.Span{document.Gap{Text: "[1y?]"}}, Months: 0, Sources: textOf("Southwind"), Learning: false},
		{Name: textOf("Zig"), Duration: textOf("1m"), Months: 1, Sources: textOf("a hobby"), Learning: false},
	}}
	single := readRows(aboveTheStatusLine(screenOf(test, handResume(test, "Mira", wideWidth, tallHeight, shortSkills()))))
	twice := readRows(aboveTheStatusLine(screenOf(test, handResume(test, "Mira", wideWidth, tallHeight, doubled))))
	cases := []struct {
		name string
		got  int
		want int
	}{
		{name: "the longest", got: barCells(twice, " 6y"), want: barCells(single, " 3y")},
		{name: "the shorter", got: barCells(twice, " 4y"), want: barCells(single, " 2y")},
		{name: "a month, whose share of the scale rounds to no cell", got: barCells(twice, " 1m"), want: 1},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if testCase.want == 0 || testCase.got != testCase.want {
				test.Errorf("on the page of twice the months\n%s\nthe bar is %d cells, want %d, as on the page\n%s", twice,
					testCase.got, testCase.want, single)
			}
		})
	}
}

// A skill of no measured time of its own, used on the side, still being learnt or a gap still to fill, stands with a
// mark of none in place of a bar.
func TestASkillOfNoMeasuredTimeHasAMarkOfNone(test *testing.T) {
	lazy := commanded(test, fixtureModel(test, commonWidth, tallHeight), "Lazy")
	cases := []struct {
		name  string
		model Model
		want  string
	}{
		{name: "one still being learnt", model: lazy, want: "○ Rust " + unmeasuredMark + " still learning"},
		{name: "one used on the side", model: lazy, want: "○ Docker " + unmeasuredMark + " on the side"},
		{name: "a gap still to fill", model: handResume(test, "Mira", wideWidth, tallHeight, shortSkills()), want: "● Lua " + unmeasuredMark + " [1y?]"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if read := readRows(aboveTheStatusLine(screenOf(test, testCase.model))); !strings.Contains(read, testCase.want) {
				test.Errorf("the page =\n%s\nwant %q on it", read, testCase.want)
			}
		})
	}
}

// Where a skill was proven stands beside how long where the page has the room, else under its name; and its bar
// stands under its name too where the name leaves the bar its widest no room beside it.
func TestASkillsPartsStandBesideWhereThePageHasTheRoom(test *testing.T) {
	cases := []struct {
		name  string
		model Model
		want  string
	}{
		{name: "where it was proven beside", model: handResume(test, "Mira", wideWidth, tallHeight, shortSkills()), want: "[1y?] Southwind"},
		{name: "where it was proven under the name", model: commanded(test, fixtureModel(test, commonWidth, tallHeight), "Lazy"), want: "still learning\na hobby"},
		{
			name:  "the bar under the name on a narrow page",
			model: commanded(test, fixtureModel(test, narrowWidth, tallHeight), "Lazy"),
			want:  "○ Rust\n" + unmeasuredMark + " still learning\na hobby",
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if read := readRows(aboveTheStatusLine(screenOf(test, testCase.model))); !strings.Contains(read, testCase.want) {
				test.Errorf("the page =\n%s\nwant %q on it", read, testCase.want)
			}
		})
	}
}

// shortSkills is a list of skills of short names, a test builds by hand: one used three years, one two, and one whose
// time is a gap still to fill.
func shortSkills() document.Skills {
	return document.Skills{List: []document.Skill{
		{Name: textOf("Go"), Duration: textOf("3y"), Months: 36, Sources: textOf("tidy-tool"), Learning: false},
		{Name: textOf("Make"), Duration: textOf("2y"), Months: 24, Sources: textOf("Northwind"), Learning: false},
		{Name: textOf("Lua"), Duration: []document.Span{document.Gap{Text: "[1y?]"}}, Months: 0, Sources: textOf("Southwind"), Learning: false},
	}}
}

// barCells is how many cells long the bar is that a row of a page, read as readRows reads it, holds before a text;
// none where no row holds such a bar.
func barCells(read, after string) int {
	match := regexp.MustCompile("(" + barCell + "+)" + regexp.QuoteMeta(after)).FindStringSubmatch(read)
	if match == nil {
		return 0
	}
	return utf8.RuneCountInString(match[1])
}
