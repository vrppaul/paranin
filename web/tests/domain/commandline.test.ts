// The command line's logic, as the TUI's command line has it: what Tab goes round for what is typed, a step round
// it, and what Enter does with the line.
import { expect, test } from "vitest";
import { completionsOf, lineOf, placesOf, ran, stepped, type Outcome } from "../../src/domain/commandline";

// The places of a page showing the experience tab: its tabs, and three projects' pages, one starting as a tab does
// and one named as a tab is; then the pages :e opens, those three and the blog's list; and the blog's folder, where :e
// opens any post
const PLACES = placesOf(
  [
    { name: "about", kind: "tab", address: "/", current: false, edited: false },
    { name: "experience", kind: "tab", address: "/experience", current: true, edited: false },
    { name: "skills", kind: "tab", address: "/skills", current: false, edited: false },
    { name: "sorted-list", kind: "project", address: "/p/sorted-list", current: false, edited: false },
    { name: "tidy-tool", kind: "project", address: "/p/tidy-tool", current: false, edited: false },
    { name: "skills", kind: "project", address: "/p/skills", current: false, edited: false },
    { name: "sorted-list", kind: "project", address: "/p/sorted-list", current: false, edited: true },
    { name: "tidy-tool", kind: "project", address: "/p/tidy-tool", current: false, edited: true },
    { name: "skills", kind: "project", address: "/p/skills", current: false, edited: true },
    { name: "blog", kind: "list", address: "/blog", current: false, edited: true },
  ],
  { name: "blog/", address: "/blog/" },
);

// Tab's step forward and Shift+Tab's back
const FORWARD = 1;
const BACKWARD = -1;

test.each([
  { name: "before anything is typed, the tabs but the one shown", typed: "", want: ["about", "skills"] },
  {
    name: "the places a letter typed starts, a project's among them, a name two share leading apart for each",
    typed: "s",
    want: ["skills", "sorted-list", "skills"],
  },
  { name: "never the tab shown, typed whole", typed: "experience", want: [] },
  { name: "never a page :e alone opens", typed: "b", want: [] },
  {
    name: "after :e and a space, every page it opens, a project's named as a tab among them, and the blog's",
    typed: "e ",
    want: ["sorted-list", "tidy-tool", "skills", "blog"],
  },
  { name: "after :e, the pages its argument starts", typed: "e ti", want: ["tidy-tool"] },
  { name: "after a command that takes no page, nothing", typed: "x ", want: [] },
])("Tab goes round $name", ({ typed, want }) => {
  expect(completionsOf(PLACES, typed).listed.map((completion) => completion.place.name)).toEqual(want);
});

test.each([
  { name: "Tab writes the first", typed: "s", steps: [FORWARD], want: "skills" },
  { name: "Tab again writes the next", typed: "s", steps: [FORWARD, FORWARD], want: "sorted-list" },
  {
    name: "past the last, Tab writes what was typed again",
    typed: "s",
    steps: [FORWARD, FORWARD, FORWARD, FORWARD],
    want: "s",
  },
  { name: "Shift+Tab writes the last", typed: "s", steps: [BACKWARD], want: "skills" },
  { name: "Tab writes a page :e opens after the command", typed: "e t", steps: [FORWARD], want: "e tidy-tool" },
  { name: "with nothing to go round, Tab leaves what was typed", typed: "x", steps: [FORWARD], want: "x" },
])("$name", ({ typed, steps, want }) => {
  const completions = steps.reduce((current, step) => stepped(current, step), completionsOf(PLACES, typed));
  expect(lineOf(completions)).toBe(want);
});

// what came of the line, a failure by its fault alone, never its words
test.each<{ name: string; typed: string; want: Partial<Outcome> }>([
  {
    name: "goes to a place by its name, the first that has it",
    typed: "skills",
    want: { kind: "go", address: "/skills" },
  },
  {
    name: "goes to the project's page :e names as a tab is",
    typed: "e skills",
    want: { kind: "go", address: "/p/skills" },
  },
  { name: "goes to a project's page by its slug", typed: "tidy-tool", want: { kind: "go", address: "/p/tidy-tool" } },
  { name: "goes to a project's page :e names", typed: "e tidy-tool", want: { kind: "go", address: "/p/tidy-tool" } },
  {
    name: "goes to a page :edit names, spaces round it",
    typed: "  edit   tidy-tool ",
    want: { kind: "go", address: "/p/tidy-tool" },
  },
  { name: "stays on the tab shown", typed: "experience", want: { kind: "stay" } },
  { name: "does nothing with nothing typed", typed: "  ", want: { kind: "nothing" } },
  { name: "refuses a command it does not know", typed: "x", want: { kind: "failed", fault: "unknown" } },
  { name: "refuses :e with no page", typed: "e", want: { kind: "failed", fault: "argument" } },
  { name: "goes to the blog's page :e names", typed: "e blog", want: { kind: "go", address: "/blog" } },
  {
    name: "goes to a post :e names in the blog's folder",
    typed: "e blog/12",
    want: { kind: "go", address: "/blog/12" },
  },
  {
    name: "refuses a page :e alone opens, by its name alone",
    typed: "blog",
    want: { kind: "failed", fault: "unknown" },
  },
  { name: "refuses :e of a page it does not open", typed: "e about", want: { kind: "failed", fault: "missing" } },
])("Enter $name", ({ typed, want }) => {
  expect(ran(PLACES, typed, undefined)).toMatchObject(want);
});

test.each([
  {
    name: "goes to the first place a name two share names, typed",
    steps: [],
    want: { kind: "go", address: "/skills" },
  },
  {
    name: "goes to the place Tab chose among those a name two share names",
    steps: [FORWARD, FORWARD],
    want: { kind: "go", address: "/p/skills" },
  },
])("Enter $name", ({ steps, want }) => {
  const completions = steps.reduce((current, step) => stepped(current, step), completionsOf(PLACES, "skills"));
  expect(ran(PLACES, lineOf(completions), completions.listed[completions.chosen])).toMatchObject(want);
});
