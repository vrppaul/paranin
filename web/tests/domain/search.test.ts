// Finding words on the page, as the TUI's search finds them: where a search matches a text, which match a step goes
// to, and what the line says.
import { expect, test } from "vitest";
import { foundSaid, matchStarts, placeAtOrAfter, placeBefore, steppedPlace } from "../../src/domain/search";

// n's step on to the next match, and N's back
const ONWARD = 1;
const BACK = -1;

test.each([
  {
    name: "each place a search starts, whatever its case",
    text: "Python 2 to python 3",
    query: "PYTHON",
    want: [0, "Python 2 to ".length],
  },
  {
    name: "matches never overlapping, the next sought after the one before",
    text: "aaaa",
    query: "aa",
    want: [0, "aa".length],
  },
  { name: "nothing for an empty search", text: "Python", query: "", want: [] },
  { name: "nothing where the search is longer than the text", text: "Py", query: "Python", want: [] },
])("a search matches $name", ({ text, query, want }) => {
  expect(matchStarts(text, query)).toEqual(want);
});

// Where four matches stand down the page, top to bottom: the first a little way down, each a screen below the one
// before
const MATCH_COUNT = 4;
const FIRST_TOP = 100;
const SCREEN = 300;
const TOPS = [...Array(MATCH_COUNT).keys()].map((place) => FIRST_TOP + place * SCREEN);

test.each([
  { name: "the first at or after a point", way: placeAtOrAfter, point: 400, want: 1 },
  { name: "the first of all, after the last", way: placeAtOrAfter, point: 1200, want: 0 },
  { name: "the last before a point", way: placeBefore, point: 700, want: 1 },
  { name: "the last of all, before the first", way: placeBefore, point: 50, want: 3 },
])("a match is picked by where the page stands: $name", ({ way, point, want }) => {
  expect(way(TOPS, point)).toBe(want);
});

test.each([
  { name: "n on to the next", place: 1, step: ONWARD, want: 2 },
  { name: "n round from the last to the first", place: 3, step: ONWARD, want: 0 },
  { name: "N round from the first to the last", place: 0, step: BACK, want: 3 },
])("a step goes $name", ({ place, step, want }) => {
  expect(steppedPlace(place, step, TOPS.length)).toBe(want);
});

test("the line says the search and which match of how many the page shows, as Neovim counts them", () => {
  const second = 1;
  expect(foundSaid("Python", second, MATCH_COUNT)).toBe("/Python [2/4]");
});
