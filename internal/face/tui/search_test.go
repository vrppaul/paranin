package tui

import (
	"strings"
	"testing"
	"testing/synctest"

	tea "charm.land/bubbletea/v2"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// The keys that reach the experience tab, which opens with a role, its point, a blank line, then the next role
func toExperience() []string {
	return []string{"l"}
}

// A search finds its text in the tab, whatever its case, scrolls to the first line holding it at or after where
// the page stood, and says which match of how many it shows.
func TestASearchScrollsToItsFirstMatch(test *testing.T) {
	cases := []struct {
		name  string
		query string
		holds []string
		lacks string
	}{
		{name: "a word of one line", query: "parser", holds: []string{"Wrote the first parser", "/parser [1/1]"}, lacks: "Northwind"},
		{name: "a word in another case", query: "PARSER", holds: []string{"Wrote the first parser", "/PARSER [1/1]"}, lacks: "Northwind"},
		{name: "a word of two lines", query: "kept", holds: []string{"Kept the queue moving", "/kept [1/2]"}, lacks: "Northwind"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(toExperience(), "/")...)
			screen := screenOf(test, pressed(test, typed(test, model, testCase.query), "enter"))
			for _, want := range testCase.holds {
				if !strings.Contains(screen, want) {
					test.Errorf("after searching %q the screen =\n%s\nwant it to hold %q", testCase.query, screen, want)
				}
			}
			if strings.Contains(screen, testCase.lacks) {
				test.Errorf("after searching %q the screen =\n%s\nwant it scrolled past %q", testCase.query, screen, testCase.lacks)
			}
		})
	}
}

// n and N show the next match and the one before, round from the last to the first, each match a step of its own.
// The experience tab holds "the" four times, the last two on one line.
func TestTheMatchKeysGoRoundTheMatches(test *testing.T) {
	cases := []struct {
		name  string
		keys  []string
		holds string
	}{
		{name: "n shows the next match", keys: []string{"n"}, holds: "/the [2/4]"},
		{name: "two matches of one line are two steps", keys: []string{"n", "n", "n"}, holds: "/the [4/4]"},
		{name: "n goes round to the first", keys: []string{"n", "n", "n", "n"}, holds: "/the [1/4]"},
		{name: "N goes round to the last", keys: []string{"N"}, holds: "/the [4/4]"},
		{name: "N goes back", keys: []string{"n", "n", "N"}, holds: "/the [2/4]"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(toExperience(), "/")...)
			model = pressed(test, pressed(test, typed(test, model, "the"), "enter"), testCase.keys...)
			if screen := screenOf(test, model); !strings.Contains(screen, testCase.holds) {
				test.Errorf("after %q the screen =\n%s\nwant it to hold %q", testCase.keys, screen, testCase.holds)
			}
		})
	}
}

// A search starts from where the page stands: its first match is the first at or after the page's top line, the
// top line itself among them, counted so while the reader types as once they search.
func TestASearchStartsWhereThePageStands(test *testing.T) {
	cases := []struct {
		name  string
		keys  []string
		query string
		after []string
		holds string
	}{
		{name: "a match after the top line", keys: []string{"3", "j"}, query: "kept", after: []string{"enter"}, holds: "/kept [2/2]"},
		{name: "a match on the top line", keys: []string{"j"}, query: "the", after: []string{"enter"}, holds: "/the [1/4]"},
		{name: "the count while the reader types", keys: []string{"3", "j"}, query: "kept", after: nil, holds: "[2/2]"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(append(toExperience(), testCase.keys...), "/")...)
			if screen := screenOf(test, pressed(test, typed(test, model, testCase.query), testCase.after...)); !strings.Contains(screen, testCase.holds) {
				test.Errorf("after %q, searching %q and %q the screen =\n%s\nwant it to hold %q", testCase.keys, testCase.query, testCase.after, screen, testCase.holds)
			}
		})
	}
}

// n and N go on from the match shown even where the page can scroll no further, so no match is skipped; each tab
// keeps the match it showed while the reader looks at another; and a tab that has shown none starts at its top line,
// as Enter does, and N there goes to the match before it. The experience tab holds an e 31 times, six of them on its
// top line, and the projects tab 11 times, its top line among them.
func TestTheMatchShownIsWhereTheNextGoesOnFrom(test *testing.T) {
	cases := []struct {
		name  string
		keys  []string
		holds string
	}{
		{name: "back over the last page", keys: []string{"N", "N"}, holds: "/e [30/31]"},
		{name: "back to a tab left", keys: []string{"n", "n", "n", "l", "h", "n"}, holds: "/e [5/31]"},
		{name: "on a tab with no match shown", keys: []string{"l", "n"}, holds: "/e [1/11]"},
		{name: "back from a tab's top line with no match shown", keys: []string{"l", "N"}, holds: "/e [11/11]"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(toExperience(), "/")...)
			model = pressed(test, pressed(test, typed(test, model, "e"), "enter"), testCase.keys...)
			if screen := screenOf(test, model); !strings.Contains(screen, testCase.holds) {
				test.Errorf("after %q the screen =\n%s\nwant it to hold %q", testCase.keys, screen, testCase.holds)
			}
		})
	}
}

// A search that finds nothing says so, as Neovim does, and leaves the page where it was.
func TestASearchThatFindsNothingTellsTheReader(test *testing.T) {
	model := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(toExperience(), "/")...)
	screen := screenOf(test, pressed(test, typed(test, model, "zebra"), "enter"))
	if !strings.Contains(screen, "E486: Pattern not found: zebra") || !strings.Contains(screen, "Northwind") {
		test.Errorf("after searching what is not there the screen =\n%s\nwant E486 and the page where it was", screen)
	}
}

// What the reader types is shown after the slash as they type it, with which of how many matches the page shows: a
// backspace takes the last letter back, one on nothing ends the search, and Esc ends it unsought.
func TestTheSearchLineFollowsTheTyping(test *testing.T) {
	cases := []struct {
		name  string
		text  string
		keys  []string
		holds string
		lacks string
	}{
		{name: "typing shows after the slash", text: "pars", keys: nil, holds: "/pars", lacks: "tab next"},
		{name: "the matches are counted as the reader types", text: "the", keys: nil, holds: "[1/4]", lacks: "E486"},
		{name: "a backspace takes a letter back", text: "parserx", keys: []string{"backspace", "enter"}, holds: "/parser [1/1]", lacks: "parserx"},
		{name: "a backspace takes a letter of two code points back whole", text: "parser" + accentedLetter, keys: []string{"backspace", "enter"}, holds: "/parser [1/1]", lacks: "E486"},
		{name: "a backspace on nothing ends the search", text: "", keys: []string{"backspace"}, holds: "Northwind", lacks: "esc cancel"},
		{name: "Esc ends the search unsought", text: "parser", keys: []string{"esc"}, holds: "Northwind", lacks: "/parser"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(toExperience(), "/")...)
			screen := screenOf(test, pressed(test, typed(test, model, testCase.text), testCase.keys...))
			if !strings.Contains(screen, testCase.holds) || strings.Contains(atTheFoot(screen), testCase.lacks) {
				test.Errorf("after typing %q and %q the screen =\n%s\nwant it to hold %q, below its page not %q", testCase.text,
					testCase.keys, screen, testCase.holds, testCase.lacks)
			}
		})
	}
}

// What the command line says lasts one key, and the search lasts until it is taken away: Esc in normal mode takes
// it, so n finds nothing to go to; a second slash opens on nothing typed.
func TestTheSearchAndWhatTheCommandLineSaidEnd(test *testing.T) {
	cases := []struct {
		name  string
		keys  []string
		lacks string
	}{
		{name: "a key wipes what the command line said", keys: []string{"j"}, lacks: "[1/4]"},
		{name: "Esc takes the search away", keys: []string{"esc", "n"}, lacks: "/the"},
		{name: "a second slash opens on nothing typed", keys: []string{"/"}, lacks: "the"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(toExperience(), "/")...)
			model = pressed(test, pressed(test, typed(test, model, "the"), "enter"), testCase.keys...)
			if below := atTheFoot(screenOf(test, model)); strings.Contains(below, testCase.lacks) {
				test.Errorf("after %q the footer = %q, want no %q", testCase.keys, below, testCase.lacks)
			}
		})
	}
}

// A match the search scrolls to stays in sight under what the command line says of it, on a screen with no row of
// crosses for the command line to stand on as on one with them: here the experience tab's last line.
func TestAMatchShowsBesideWhatTheCommandLineSays(test *testing.T) {
	// a screen shorter than the experience tab, at a phone's width as at a common one
	const shorterThanTheTab = 10
	cases := []struct {
		name  string
		width int
	}{
		{name: "a phone's screen, with no crosses", width: phoneWidth},
		{name: "a common screen, with crosses", width: commonWidth},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, fixtureModel(test, testCase.width, shorterThanTheTab), append(toExperience(), "/")...)
			screen := screenOf(test, pressed(test, typed(test, model, "outages"), "enter"))
			if page := aboveTheCommandLine(screen); !strings.Contains(page, "outages") || !strings.Contains(commandLineOf(screen), "/outages") {
				test.Errorf("after searching the last line the screen =\n%s\nwant the match on the page over the command line", screen)
			}
		})
	}
}

// n goes on from the page as the reader left it: from the match it showed while the page stays, from the page's
// top once they scroll it.
func TestTheNextMatchFollowsThePage(test *testing.T) {
	cases := []struct {
		name  string
		keys  []string
		holds string
	}{
		{name: "from the match shown", keys: []string{"n"}, holds: "/the [2/4]"},
		{name: "a count goes as many matches on", keys: []string{"2", "n"}, holds: "/the [3/4]"},
		{name: "from the top line after a scroll to the end", keys: []string{"G", "n"}, holds: "/the [3/4]"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(toExperience(), "/")...)
			model = pressed(test, pressed(test, typed(test, model, "the"), "enter"), testCase.keys...)
			if screen := screenOf(test, model); !strings.Contains(screen, testCase.holds) {
				test.Errorf("after %q the screen =\n%s\nwant it to hold %q", testCase.keys, screen, testCase.holds)
			}
		})
	}
}

// A new search forgets the matches the old one showed: back on a tab, n starts by the page's top again. The match
// of an e before the last leaves the experience tab at its last page, where the old match would send n round to the
// first match of "kept" rather than on to the second.
func TestANewSearchForgetsTheMatchesShown(test *testing.T) {
	model := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(toExperience(), "/")...)
	model = pressed(test, typed(test, model, "e"), "enter", "N", "N", "l", "/")
	model = pressed(test, typed(test, model, "kept"), "enter", "h", "n")
	if screen := screenOf(test, model); !strings.Contains(screen, "/kept [2/2]") {
		test.Errorf("after a new search and back on the tab the screen =\n%s\nwant the second line holding kept", screen)
	}
}

// A resize forgets the matches shown, which were lines of the layout gone: n starts by the page's top again.
func TestAResizeForgetsTheMatchesShown(test *testing.T) {
	model := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(toExperience(), "/")...)
	resized, _ := pressed(test, typed(test, model, "e"), "enter").Update(tea.WindowSizeMsg{Width: commonWidth, Height: shortHeight})
	if screen := screenOf(test, pressed(test, as[Model](test, resized), "n")); !strings.Contains(screen, "/e [1/31]") {
		test.Errorf("after a resize and n the screen =\n%s\nwant the match on the top line", screen)
	}
}

// Esc typed just before a key reaches the program glued to it, as the key with Alt: it is read as Esc, then the
// key, so the search typed ends and n goes on.
func TestEscapeGluedToAKeyIsReadAsBoth(test *testing.T) {
	model := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(toExperience(), "/")...)
	model = pressed(test, pressed(test, typed(test, model, "the"), "enter"), "/")
	next, _ := model.Update(tea.KeyPressMsg{Code: 'n', Mod: tea.ModAlt})
	if screen := screenOf(test, as[Model](test, next)); !strings.Contains(screen, "/the [2/4]") {
		test.Errorf("after Esc glued to n the screen =\n%s\nwant the search ended and the next match", screen)
	}
}

// While the reader types a search, the page goes to its first match at or after where the page stood, as
// Neovim's incsearch does, and back there while what is typed matches nothing, or once the reader leaves it
// unsought. Three lines down, the page stands past the first line holding "kept".
func TestThePageFollowsTheSearchTyped(test *testing.T) {
	cases := []struct {
		name   string
		before []string
		text   string
		keys   []string
		holds  string
		lacks  string
	}{
		{name: "typing goes to the first match", before: nil, text: "parser", keys: nil, holds: "Wrote the first parser", lacks: "Northwind"},
		{name: "typing goes to the first match after the page", before: []string{"3", "j"}, text: "kept", keys: nil, holds: "Kept the build", lacks: "Kept the queue"},
		{name: "typing what matches nothing goes back", before: nil, text: "parserz", keys: nil, holds: "Northwind", lacks: "Wrote the first parser"},
		{name: "a backspace goes to the match again", before: nil, text: "parserz", keys: []string{"backspace"}, holds: "Wrote the first parser", lacks: "Northwind"},
		{name: "a backspace seeks from where the search began", before: nil, text: "kept the b", keys: []string{"backspace"}, holds: "Kept the queue", lacks: "Kept the build"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(append(toExperience(), testCase.before...), "/")...)
			screen := screenOf(test, pressed(test, typed(test, model, testCase.text), testCase.keys...))
			if !strings.Contains(screen, testCase.holds) || strings.Contains(screen, testCase.lacks) {
				test.Errorf("after typing %q and %q the screen =\n%s\nwant it to hold %q and not %q", testCase.text, testCase.keys,
					screen, testCase.holds, testCase.lacks)
			}
		})
	}
}

// A search typed and left unconfirmed leaves the search before it as it was: n goes on through its matches from the
// one shown. Esc and Ctrl+C leave it, as do a backspace on nothing and Enter on nothing.
func TestASearchLeftKeepsTheOneBefore(test *testing.T) {
	cases := []struct {
		name string
		text string
		keys []string
	}{
		{name: "Esc", text: "parser", keys: []string{"esc"}},
		{name: "Ctrl+C", text: "parser", keys: []string{"ctrl+c"}},
		{name: "a backspace on nothing", text: "p", keys: []string{"backspace", "backspace"}},
		{name: "Enter on nothing", text: "p", keys: []string{"backspace", "enter"}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(toExperience(), "/")...)
			model = pressed(test, typed(test, pressed(test, typed(test, model, "the"), "enter", "/"), testCase.text), testCase.keys...)
			if screen := screenOf(test, pressed(test, model, "n")); !strings.Contains(screen, "/the [2/4]") {
				test.Errorf("after typing %q, %q and n the screen =\n%s\nwant the second match of the search before", testCase.text,
					testCase.keys, screen)
			}
		})
	}
}

// A resize while the reader types a search paints every tab again, each with the search it marks: the tab typed on
// marks what is typed, the rest the search before. Left unsought, the search before goes on in another tab, where
// "the" stands once.
func TestAResizeWhileTypingKeepsTheSearchBefore(test *testing.T) {
	model := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(toExperience(), "/")...)
	model = typed(test, pressed(test, typed(test, model, "the"), "enter", "/"), "parser")
	resized, _ := model.Update(tea.WindowSizeMsg{Width: commonWidth, Height: shortHeight})
	if screen := screenOf(test, pressed(test, as[Model](test, resized), "esc", "l", "n")); !strings.Contains(screen, "/the [1/1]") {
		test.Errorf("after a resize while typing, Esc, the next tab and n the screen =\n%s\nwant the search before", screen)
	}
}

// Matches never overlap, as Neovim's do not: the next is sought after the end of the one before, so "aaa" holds
// one match of "aa", and "aa" a second.
func TestMatchesNeverOverlap(test *testing.T) {
	composed := document.Document{
		Head:   fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, nil),
		Blocks: []document.Block{document.Paragraph{Spans: []document.Span{document.Text{Text: "aaa aa"}}}},
	}
	model := pressed(test, handModel(test, commonWidth, commonHeight, composed), "/")
	if screen := screenOf(test, pressed(test, typed(test, model, "aa"), "enter")); !strings.Contains(screen, "/aa [1/2]") {
		test.Errorf("after searching %q in %q the screen =\n%s\nwant two matches", "aa", "aaa aa", screen)
	}
}

// A resize while the reader types a search keeps the page where it stands: a letter typed after it, and Enter, go
// from there, never from a line the new layout lacks. On a wide terminal the experience tab's last page
// holds two lines with an e, and three matches of "er" on its top line.
func TestTypingAfterAResizeKeepsThePage(test *testing.T) {
	cases := []struct {
		name string
		text string
		keys []string
	}{
		{name: "Enter", text: "", keys: []string{"enter"}},
		{name: "a letter", text: "r", keys: nil},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := typed(test, pressed(test, fixtureModel(test, narrowWidth, shortHeight), append(toExperience(), "G", "/")...), "e")
			resized, _ := model.Update(tea.WindowSizeMsg{Width: commonWidth, Height: shortHeight})
			screen := screenOf(test, pressed(test, typed(test, as[Model](test, resized), testCase.text), testCase.keys...))
			if below := atTheFoot(screen); !strings.Contains(below, "Bot") {
				test.Errorf("after a resize while typing, %q and %q the footer = %q, want the page still at its end",
					testCase.text, testCase.keys, below)
			}
		})
	}
}

// The rows at a screen's foot: the command line's, where it shows, and the footer under it
const footRows = 2

// What parts the command line's left, the mark and what is typed or what a key said, from the keys it names at its
// right: a gap wider than any typed in a test
const commandLineGap = "   "

// atTheFoot is the last rows of a screen: the command line's row, where it shows, and the footer.
func atTheFoot(screen string) string {
	rows := strings.Split(screen, lineEnd)
	return strings.Join(rows[max(len(rows)-footRows, 0):], lineEnd)
}

// aboveTheCommandLine is a screen but for its foot: the command line's row and the footer.
func aboveTheCommandLine(screen string) string {
	rows := strings.Split(screen, lineEnd)
	return strings.Join(rows[:max(len(rows)-footRows, 0)], lineEnd)
}

// aboveTheFooter is a screen but for its footer.
func aboveTheFooter(screen string) string {
	rows := strings.Split(screen, lineEnd)
	return strings.Join(rows[:len(rows)-1], lineEnd)
}

// footerOf is the footer of a screen, its last row, without the blanks around it.
func footerOf(screen string) string {
	rows := strings.Split(screen, lineEnd)
	return strings.TrimSpace(rows[len(rows)-1])
}

// commandLineOf is what the command line of a screen holds at its left, the mark and what is typed or what a key
// said, without the keys it names at its right.
func commandLineOf(screen string) string {
	rows := strings.Split(screen, lineEnd)
	left, _, _ := strings.Cut(strings.TrimSpace(rows[max(len(rows)-footRows, 0)]), commandLineGap)
	return left
}

// A control character sent as typing never reaches the reader's terminal: the command line, a search or a place typed
// on it, shows letters alone.
func TestAControlCharacterTypedNeverReachesTheScreen(test *testing.T) {
	for _, opening := range []string{"/", ":"} {
		test.Run(opening, func(test *testing.T) {
			model := pressed(test, fixtureModel(test, commonWidth, shortHeight), opening)
			next, _ := model.Update(tea.KeyPressMsg{Code: 'x', Text: "\x1b[2J"})
			if view := as[Model](test, next).View().Content; strings.Contains(view, "[2J") {
				test.Errorf("after %q the screen holds the control sequence typed: %q", opening, view)
			}
		})
	}
}

// A count before a key moves by as many: 3j three lines down, 2G to the second line.
func TestACountMovesByAsMany(test *testing.T) {
	cases := []struct {
		name  string
		keys  []string
		holds string
		lacks string
	}{
		{name: "3j goes three lines down", keys: []string{"3", "j"}, holds: "Southwind", lacks: "Kept the queue"},
		{name: "2G goes to the second line", keys: []string{"G", "2", "G"}, holds: "Kept the queue", lacks: "Northwind"},
		{name: "a zero alone counts nothing", keys: []string{"0", "j"}, holds: "Kept the queue", lacks: "Northwind"},
		{name: "a number of two digits", keys: []string{"G", "1", "0", "k"}, holds: "Northwind", lacks: "one long winter"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := pressed(test, fixtureModel(test, commonWidth, shortHeight), append(toExperience(), testCase.keys...)...)
			if screen := screenOf(test, model); !strings.Contains(screen, testCase.holds) || strings.Contains(screen, testCase.lacks) {
				test.Errorf("after %q the screen =\n%s\nwant it to hold %q and not %q", testCase.keys, screen, testCase.holds, testCase.lacks)
			}
		})
	}
}

// A number pressed alone chooses the tab of that number once the reader stops typing it.
func TestANumberAloneChoosesItsTab(test *testing.T) {
	synctest.Test(test, func(test *testing.T) {
		alone, wait := fixtureModel(test, commonWidth, shortHeight).Update(keyPress(test, "3"))
		chosen, _ := alone.Update(wait())
		if screen := screenOf(test, as[Model](test, chosen)); !strings.Contains(screen, "other-tool") {
			test.Errorf("after 3 and a pause the screen =\n%s\nwant the third tab", screen)
		}
	})
}

// A number followed in time by a key counts the key, and the pause after it chooses no tab: 3l, and 2 g g, which a
// pause between the two g's does not part.
func TestANumberBeforeAKeyCountsTheKey(test *testing.T) {
	cases := []struct {
		name  string
		start []string
		digit int
		then  []string
		holds string
		lacks string
	}{
		{name: "2l goes two tabs on", start: []string{"2", "l"}, digit: 0, then: nil, holds: "other-tool", lacks: "Kept the queue"},
		{name: "2gg goes to the second line", start: append(toExperience(), "2", "g"), digit: len(toExperience()), then: []string{"g"}, holds: "Kept the queue", lacks: "Northwind"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			synctest.Test(test, func(test *testing.T) {
				model, pause := pressedWaiting(test, fixtureModel(test, commonWidth, shortHeight), testCase.start, testCase.digit)
				paused, _ := model.Update(pause())
				screen := screenOf(test, pressed(test, as[Model](test, paused), testCase.then...))
				if !strings.Contains(screen, testCase.holds) || strings.Contains(screen, testCase.lacks) {
					test.Errorf("after %q, a pause and %q the screen =\n%s\nwant it to hold %q and not %q", testCase.start,
						testCase.then, screen, testCase.holds, testCase.lacks)
				}
			})
		})
	}
}

// pressedWaiting is a model after keys, and the wait the key at a place among them started.
func pressedWaiting(test *testing.T, model Model, keys []string, waitingPlace int) (Model, tea.Cmd) {
	commands := []tea.Cmd{}
	for _, key := range keys {
		next, command := model.Update(keyPress(test, key))
		model = as[Model](test, next)
		commands = append(commands, command)
	}
	return model, commands[waitingPlace]
}

// A pause that ends a digit typed before the last one is passed over: the number still counts the key after it.
func TestAPauseAfterAnEarlierDigitPassesOver(test *testing.T) {
	synctest.Test(test, func(test *testing.T) {
		first, earlier := pressed(test, fixtureModel(test, commonWidth, shortHeight), toExperience()...).Update(keyPress(test, "1"))
		second, _ := first.Update(keyPress(test, "2"))
		paused, _ := second.Update(earlier())
		moved, _ := paused.Update(keyPress(test, "j"))
		if screen := screenOf(test, as[Model](test, moved)); !strings.Contains(screen, "one long winter") {
			test.Errorf("after 1, 2, the first digit's pause and j the screen =\n%s\nwant twelve lines down, at the end", screen)
		}
	})
}

// A number beyond the last tab, typed alone, chooses none, and counts nothing after it.
func TestANumberBeyondTheTabsChoosesNone(test *testing.T) {
	synctest.Test(test, func(test *testing.T) {
		alone, wait := pressedWaiting(test, fixtureModel(test, commonWidth, shortHeight), append(toExperience(), "9"), len(toExperience()))
		after, _ := alone.Update(wait())
		screen := screenOf(test, pressed(test, as[Model](test, after), "j"))
		if !strings.Contains(screen, "Kept the queue") || strings.Contains(screen, "one long winter") {
			test.Errorf("after 9, a pause and j the screen =\n%s\nwant the same tab a line down", screen)
		}
	})
}
