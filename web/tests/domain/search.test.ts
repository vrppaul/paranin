// Finding words on the page, as the TUI's search finds them: where a search matches a passage, which match a step
// goes to, and what the line says.
import { expect, test } from "vitest";
import { foundSaid, matchesIn, passageOf, placeAtOrAfter, placeBefore, steppedPlace } from "../../src/domain/search";

// n's step on to the next match, and N's back
const ONWARD = 1;
const BACK = -1;

// A project's line as the page writes it, a no-break space before the dot, and as a reader reads it; and the same
// search typed with a plain space and with a no-break one
const DOTTED_AS_WRITTEN = "Python\u00a0· tools";
const DOTTED_AS_READ = "Python · tools";
const DOTTED_TYPED = "python ·";
const DOTTED_TYPED_UNBROKEN = "python\u00a0·";

/** A place in a passage, by the words of the piece it stands in. */
function placed(words: string, offset: number): { piece: { data: string }; offset: number } {
  return { piece: { data: words }, offset };
}

test.each([
  {
    name: "each match, whatever its case",
    pieces: ["Python 2 to python 3"],
    query: "PYTHON",
    want: [
      { start: placed("Python 2 to python 3", 0), end: placed("Python 2 to python 3", "Python".length) },
      {
        start: placed("Python 2 to python 3", "Python 2 to ".length),
        end: placed("Python 2 to python 3", "Python 2 to python".length),
      },
    ],
  },
  {
    name: "matches never overlapping, the next sought after the one before",
    pieces: ["aaa"],
    query: "aa",
    want: [{ start: placed("aaa", 0), end: placed("aaa", "aa".length) }],
  },
  { name: "nothing for an empty search", pieces: ["Python"], query: "", want: [] },
  { name: "nothing where the search is longer than the words", pieces: ["Py"], query: "Python", want: [] },
  {
    name: "a phrase a mark parts, from inside the first piece to inside the second",
    pieces: ["Kept the ", "queue", " moving"],
    query: "the queue",
    want: [{ start: placed("Kept the ", "Kept ".length), end: placed("queue", "queue".length) }],
  },
  {
    name: "a match ending where its piece ends, ending in that piece",
    pieces: ["the ", "queue", " moving"],
    query: "queue",
    want: [{ start: placed("queue", 0), end: placed("queue", "queue".length) }],
  },
  {
    name: "a match across an empty piece",
    pieces: ["Pyt", "", "hon"],
    query: "python",
    want: [{ start: placed("Pyt", 0), end: placed("hon", "hon".length) }],
  },
  {
    name: "a no-break space the page writes, as the space a reader types",
    pieces: [DOTTED_AS_WRITTEN],
    query: DOTTED_TYPED,
    want: [{ start: placed(DOTTED_AS_WRITTEN, 0), end: placed(DOTTED_AS_WRITTEN, DOTTED_TYPED.length) }],
  },
  {
    name: "a no-break space typed, as a space",
    pieces: [DOTTED_AS_READ],
    query: DOTTED_TYPED_UNBROKEN,
    want: [{ start: placed(DOTTED_AS_READ, 0), end: placed(DOTTED_AS_READ, DOTTED_TYPED.length) }],
  },
])("a search matches in a passage: $name", ({ pieces, query, want }) => {
  expect(matchesIn(passageOf(pieces.map((data) => ({ data }))), query)).toEqual(want);
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
  expect(foundSaid("Python", second, MATCH_COUNT)).toEqual([
    { text: "/", echoed: false },
    { text: "Python", echoed: true },
    { text: " [2/4]", echoed: false },
  ]);
});
