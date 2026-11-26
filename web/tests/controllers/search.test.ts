// The vim keys' search, run over the page the server draws (./linepage.ts), laid out and marked by the tests' own
// layout and marks.
import { beforeEach, expect, test } from "vitest";
import { found, input, press, started, type } from "./linepage";

// The parts as the server draws them: the line, its field, the colon's mark and the slash's, what the line says, the
// keys it names for a place and for a search, and the count of a search's matches
const LINE = '[data-part="command-line"]';
const FIELD = '[data-part="command-field"]';
const COLON_MARK = '[data-part="command-mark"]';
const SLASH_MARK = '[data-part="search-mark"]';
const SAID = '[data-part="command-said"]';
// words a page marks as English, whatever its own language, and the names the field is labelled by in each mode
const ENGLISH = '[lang="en"]';
const FIELD_NAMES = '[data-part="command-name"], [data-part="search-name"]';
const COMMAND_KEYS = '[data-part="command-keys"]';
const SEARCH_KEYS = '[data-part="search-keys"]';
const COUNT = '[data-part="search-count"]';

// The box of the switch that turns the vim keys on
const BOX = '[data-part="vim-keys"] input';

// The theme button's part, which the server draws hidden until the site's script shows it, the button, and its name,
// its tooltip
const THEME = '[data-part="theme"]';
const THEME_BUTTON = `${THEME} button`;
const THEME_NAME = `${THEME} [data-part="tooltip"]`;

// The attribute that marks what the line says of a match shown, which is no refusal
const STATE = "data-state";

// A search the fixture's experience matches twice, "Kept the queue moving" and "Kept the build green", one it
// matches twice in the roles' names, and one it never matches
const KEPT = "kept";
const WIND = "wind";
const NOWHERE = "zebra";

// A phrase the fixture's first point parts with a mark, "Kept the <mark>queue</mark> moving"; and the end of the
// first role's heading run on into its place, "Northwind — Senior Engineer" and "Brno", which the served page sets
// side by side with nothing between, as the golden page, set out a tag a line, does not
const PARTED = "the queue";
const ACROSS_BLOCKS = "engineerbrno";
const ROLE_HEADING = "main .entry-body > :first-child";
const ROLE_PLACE = "main .entry-place";

// Each test opens its first page in a session of its own, which the pages it opens after share
beforeEach(() => sessionStorage.clear());

// Where the session keeps the search for the pages after, as a page opened since would leave it
const SESSION_SEARCH = "search";

/** The browser shows the page again from its cache, after a page opened since. */
function shownAgain(): void {
  window.dispatchEvent(new PageTransitionEvent("pageshow", { persisted: true }));
}

/** A search typed after a slash and kept with Enter. */
function searched(query: string): void {
  press("/");
  type(query);
  press("Enter");
}

/** What the line says, where it says something. */
function said(): string | null {
  return found(LINE).hidden || found(SAID).hidden ? null : found(SAID).textContent;
}

/** The name a screen reader gives a field: the words of the element that labels it, among the field's names. */
function nameOf(field: HTMLElement): string | null | undefined {
  const label = field.getAttribute("aria-labelledby");
  return [...document.querySelectorAll(FIELD_NAMES)].find((name) => name.id === label)?.textContent;
}

test("the slash opens the line for a search: its mark and its keys, the field named for it and holding the keys", () => {
  started();
  // a place chosen on the colon's menu first, which the search's field must not go on naming
  press(":");
  press("Tab");
  press("Escape");
  press("/");
  const field = found(FIELD);
  expect({
    open: !found(LINE).hidden,
    slash: !found(SLASH_MARK).hidden,
    colon: !found(COLON_MARK).hidden,
    searchKeys: !found(SEARCH_KEYS).hidden,
    commandKeys: !found(COMMAND_KEYS).hidden,
    role: field.getAttribute("role"),
    name: nameOf(field),
    describedBy: field.getAttribute("aria-describedby"),
    menu: [
      field.getAttribute("aria-activedescendant"),
      field.getAttribute("aria-autocomplete"),
      field.getAttribute("aria-controls"),
    ],
    focused: document.activeElement === field,
  }).toEqual({
    open: true,
    slash: true,
    colon: false,
    searchKeys: true,
    commandKeys: false,
    role: "searchbox",
    name: "search",
    describedBy: found(SEARCH_KEYS).id,
    menu: [null, null, null],
    focused: true,
  });
});

test.each([
  { name: "with the vim keys off", ticked: false, held: {} },
  { name: "with a modifier held", ticked: true, held: { ctrlKey: true } },
])("the slash opens nothing $name", ({ ticked, held }) => {
  started(ticked);
  press("/", held);
  expect(found(LINE).hidden).toBe(true);
});

test("each letter typed marks the page's matches, the first named by the count and brought into sight", () => {
  const { marks, layout } = started();
  press("/");
  type(KEPT);
  const [other] = marks.others;
  expect({
    others: marks.others.length,
    count: found(COUNT).textContent,
    firstNamed: (marks.current ?? Infinity) < (other ?? 0),
    inSight: layout.standsAt === marks.current,
  }).toEqual({ others: 1, count: "[1/2]", firstNamed: true, inSight: true });
});

test("a search starts where the page stands: the first match at or after it, which Enter keeps", () => {
  const { marks, layout } = started();
  press("/");
  type(KEPT);
  const [second] = marks.others;
  press("Escape");
  layout.standsAt = second ?? 0;
  press("/");
  type(KEPT);
  const typing = { current: marks.current, count: found(COUNT).textContent };
  press("Enter");
  expect({ typing, kept: said() }).toEqual({ typing: { current: second, count: "[2/2]" }, kept: "/kept [2/2]" });
});

test("a search never matches what the page does not show", () => {
  started(true, (page) => page.querySelector("main li")?.setAttribute("hidden", ""));
  searched(KEPT);
  expect(said()).toBe("/kept [1/1]");
});

test("a search finds a phrase a mark parts, and marks it whole", () => {
  const { marks } = started();
  searched(PARTED);
  expect({ said: said(), marked: marks.currentWords }).toEqual({ said: `/${PARTED} [1/1]`, marked: PARTED });
});

test("a search never joins the words of two blocks, however close the markup sets them", () => {
  started(true, (page) => found(ROLE_HEADING, page).append(found(ROLE_PLACE, page)));
  searched(ACROSS_BLOCKS);
  expect(said()).toBe(`E486: Pattern not found: ${ACROSS_BLOCKS}`);
});

test("the line says Neovim's own words in English, and what the reader typed in the page's own language", () => {
  started();
  searched(NOWHERE);
  const english = [...found(SAID).querySelectorAll(ENGLISH)].map((words) => words.textContent).join("");
  expect({ english, whole: said() }).toEqual({
    english: "E486: Pattern not found: ",
    whole: `E486: Pattern not found: ${NOWHERE}`,
  });
});

test("Enter keeps the search, the line saying which match the page shows until the next key, the marks staying", () => {
  const { marks } = started();
  searched(KEPT);
  const kept = { words: said(), keys: !found(SEARCH_KEYS).hidden };
  press("j");
  expect({ kept, after: said(), marked: marks.current !== undefined }).toEqual({
    kept: { words: "/kept [1/2]", keys: false },
    after: null,
    marked: true,
  });
});

test("Enter on an empty search leaves the line, saying nothing", () => {
  started();
  press("/");
  press("Enter");
  expect({ open: !found(LINE).hidden, said: said() }).toEqual({ open: false, said: null });
});

test("n goes on to the next match and round from the last to the first, N back, the line saying which", () => {
  started();
  searched(KEPT);
  press("n");
  const second = said();
  press("n");
  const roundAgain = said();
  press("N");
  expect({ second, roundAgain, back: said() }).toEqual({
    second: "/kept [2/2]",
    roundAgain: "/kept [1/2]",
    back: "/kept [2/2]",
  });
});

test("n goes from where the page stands once the reader scrolls it away: past the last match, round to the first", () => {
  const { marks, layout } = started();
  searched(KEPT);
  const first = marks.current;
  const [second] = marks.others;
  layout.standsAt = (second ?? 0) + 1;
  press("n");
  expect({ said: said(), current: marks.current }).toEqual({ said: "/kept [1/2]", current: first });
});

test("N goes from where the page stands once the reader scrolls it away: the last match before it", () => {
  const { marks, layout } = started();
  searched(KEPT);
  press("n");
  const second = marks.current;
  const [first] = marks.others;
  layout.standsAt = (second ?? 0) - 1;
  press("N");
  expect({ said: said(), current: marks.current }).toEqual({ said: "/kept [1/2]", current: first });
});

test("Tab while a search is typed keeps the focus and writes no place on the line", () => {
  started();
  press("/");
  type(KEPT);
  const tab = press("Tab");
  expect({ kept: tab.defaultPrevented, line: input(FIELD).value }).toEqual({ kept: true, line: KEPT });
});

test("a search that matches nothing marks nothing, counts nothing, and Enter says it failed", () => {
  const { marks } = started();
  press("/");
  type(NOWHERE);
  const typing = { marked: marks.current, count: found(COUNT).textContent };
  press("Enter");
  expect({ typing, failed: said() !== null && !found(SAID).hasAttribute(STATE) }).toEqual({
    typing: { marked: undefined, count: "" },
    failed: true,
  });
});

test("the page goes back where it stood once what is typed matches nothing", () => {
  const { layout } = started();
  const stood = layout.standsAt;
  press("/");
  type(KEPT);
  const brought = layout.standsAt;
  type(NOWHERE);
  expect({ moved: brought !== stood, back: layout.standsAt }).toEqual({ moved: true, back: stood });
});

test("Esc while a search is typed puts back the search kept before and the page where it stood", () => {
  const { marks, layout } = started();
  searched(KEPT);
  const kept = { current: marks.current, standsAt: layout.standsAt };
  press("/");
  type(WIND);
  press("Escape");
  expect({ current: marks.current, standsAt: layout.standsAt, open: !found(LINE).hidden }).toEqual({
    ...kept,
    open: false,
  });
});

test("Esc on the colon's line leaves the page where it stands, a search kept before or not", () => {
  const { layout } = started();
  searched(KEPT);
  const stands = layout.standsAt + 1;
  layout.standsAt = stands;
  press(":");
  press("Escape");
  expect(layout.standsAt).toBe(stands);
});

test("Esc on the page puts the search away, its marks with it, and n then finds nothing to go to", () => {
  const { marks } = started();
  searched(KEPT);
  press("Escape");
  press("n");
  expect({ marked: marks.current, said: said() }).toEqual({ marked: undefined, said: null });
});

test("Esc on the page puts away one thing a press: a tooltip shown, then a window open, then the search", () => {
  const { marks, windows } = started();
  searched(KEPT);
  found(THEME).hidden = false;
  found(THEME_BUTTON).focus();
  press("Escape");
  const afterFirstEsc = { tooltipHidden: found(THEME_NAME).hidden, marks: marks.others.length };
  windows.open = true;
  press("Escape");
  const marksAfterSecondEsc = marks.others.length;
  // the browser closed the window on the second Esc
  windows.open = false;
  press("Escape");
  expect({ afterFirstEsc, marksAfterSecondEsc, marksAfterThirdEsc: marks.others.length }).toEqual({
    afterFirstEsc: { tooltipHidden: true, marks: 1 },
    marksAfterSecondEsc: 1,
    marksAfterThirdEsc: 0,
  });
});

test("a search kept marks the page opened next, none of its matches named until n goes to the first", () => {
  started();
  searched(KEPT);
  const { marks } = started();
  const opened = { named: marks.current, others: marks.others.length, said: said() };
  press("n");
  expect({ opened, after: said() }).toEqual({
    opened: { named: undefined, others: 2, said: null },
    after: "/kept [1/2]",
  });
});

test("Esc on the page puts the search away on the pages opened after it too", () => {
  started();
  searched(KEPT);
  press("Escape");
  const { marks } = started();
  press("n");
  expect({ others: marks.others, said: said() }).toEqual({ others: [], said: null });
});

test("Esc while a search is typed on a page a kept search marks puts its marks back, none named", () => {
  started();
  searched(KEPT);
  const { marks } = started();
  press("/");
  type(WIND);
  press("Escape");
  expect({ named: marks.current, others: marks.others.length }).toEqual({ named: undefined, others: 2 });
});

test("unticking the vim keys takes a kept search's marks away, and ticking them again marks it, none named", () => {
  const { marks } = started();
  searched(KEPT);
  input(BOX).click();
  const unticked = { named: marks.current, others: marks.others.length };
  input(BOX).click();
  expect({ unticked, ticked: { named: marks.current, others: marks.others.length } }).toEqual({
    unticked: { named: undefined, others: 0 },
    ticked: { named: undefined, others: 2 },
  });
});

test("a page the browser shows again from its cache marks none once a page opened since put the search away", () => {
  const { marks } = started();
  searched(KEPT);
  sessionStorage.removeItem(SESSION_SEARCH);
  shownAgain();
  press("n");
  expect({ others: marks.others, said: said() }).toEqual({ others: [], said: null });
});

test("a page the browser shows again from its cache says nothing on its line of the match it named before", () => {
  started();
  searched(KEPT);
  shownAgain();
  expect(said()).toBe(null);
});

test("a page the browser shows again from its cache marks the search a page opened since kept, none named", () => {
  const { marks } = started();
  sessionStorage.setItem(SESSION_SEARCH, KEPT);
  shownAgain();
  expect({ named: marks.current, others: marks.others.length }).toEqual({ named: undefined, others: 2 });
});

test("a page opened with the vim keys off marks no search kept before", () => {
  started();
  searched(KEPT);
  const { marks } = started(false);
  expect(marks.others).toEqual([]);
});

test("the colon after a search opens the line for a place, its own mark, keys and field back", () => {
  started();
  searched(KEPT);
  press(":");
  const field = found(FIELD);
  expect({
    colon: !found(COLON_MARK).hidden,
    slash: !found(SLASH_MARK).hidden,
    commandKeys: !found(COMMAND_KEYS).hidden,
    searchKeys: !found(SEARCH_KEYS).hidden,
    role: field.getAttribute("role"),
    name: nameOf(field),
  }).toEqual({ colon: true, slash: false, commandKeys: true, searchKeys: false, role: "combobox", name: "command" });
});
