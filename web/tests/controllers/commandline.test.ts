// The command line the vim keys open, run over the page the server draws (./linepage.ts): a place or a command after
// the colon.
import { expect, test } from "vitest";
import { found, input, press, started, type } from "./linepage";
import listPage from "../../../internal/face/html/testdata/blog.html?raw";
import postPage from "../../../internal/face/html/testdata/blog/reading-slowly.html?raw";
import partPage from "../../../internal/face/html/testdata/blog/second.html?raw";

// The parts as the server draws them: the switch and its box, the line, its field, each place on its menu with the
// letters typed of its name, what the line says of a command that failed, and the key that goes round the places
const SWITCH = '[data-part="vim-keys"]';
const BOX = `${SWITCH} input`;
const LINE = '[data-part="command-line"]';
const FIELD = '[data-part="command-field"]';
const PLACES = '[data-part="places"]';
const PLACE = '[data-part="place"]';
const TYPED = '[data-part="place-typed"]';
const SAID = '[data-part="command-said"]';
const NEXT = '[data-part="command-next"]';

/** The places the menu shows. */
function shownPlaces(): HTMLElement[] {
  return found(PLACES).hidden ? [] : [...document.querySelectorAll<HTMLElement>(PLACE)].filter((item) => !item.hidden);
}

/** What a reader sees of the line: whether it is open to type, what it holds, the places on its menu, the one chosen,
 * and whether it says why it refused a command. */
function seen(): { open: boolean; line: string; places: string[]; chosen: string | undefined; said: boolean } {
  const chosen = [...document.querySelectorAll<HTMLElement>(PLACE)].find(
    (item) => item.getAttribute("aria-selected") === "true",
  );
  return {
    open: !found(LINE).hidden && found(SAID).hasAttribute("hidden"),
    line: input(FIELD).value,
    places: shownPlaces().map((item) => item.dataset["name"] ?? ""),
    chosen: chosen?.dataset["name"],
    said: !found(LINE).hidden && !found(SAID).hidden,
  };
}

test("the colon opens the line with the tabs but the one shown, the field holding the keys", () => {
  started();
  press(":");
  expect({ ...seen(), focused: document.activeElement === found(FIELD) }).toEqual({
    open: true,
    line: "",
    places: ["about", "projects", "skills", "contact"],
    chosen: undefined,
    said: false,
    focused: true,
  });
});

test.each([
  { name: "the colon, with the vim keys off", key: ":", ticked: false, held: {} },
  { name: "the colon, with a modifier held", key: ":", ticked: true, held: { ctrlKey: true } },
  { name: "another key", key: ";", ticked: true, held: {} },
])("$name opens nothing", ({ key, ticked, held }) => {
  started(ticked);
  press(key, held);
  expect(seen().open).toBe(false);
});

test("the colon opens the line again after Esc closed it, with nothing typed and nothing chosen", () => {
  started();
  press(":");
  type("s");
  press("Tab");
  press("Escape");
  press(":");
  const field = input(FIELD);
  expect({
    open: seen().open,
    line: field.value,
    expanded: field.getAttribute("aria-expanded"),
    chosen: field.getAttribute("aria-activedescendant"),
  }).toEqual({ open: true, line: "", expanded: "true", chosen: null });
});

test("the field tells a screen reader the menu shows and which place is chosen, and that it closed", () => {
  started();
  press(":");
  press("Tab");
  const field = input(FIELD);
  const open = { expanded: field.getAttribute("aria-expanded"), chosen: field.getAttribute("aria-activedescendant") };
  press("Escape");
  expect({ open, closed: field.getAttribute("aria-expanded") }).toEqual({
    open: { expanded: "true", chosen: found(`${PLACE}[data-name="about"]`).id },
    closed: "false",
  });
});

test("the colon opens nothing where the switch is gone from the page, as it is on a phone", () => {
  started();
  found(SWITCH).style.display = "none";
  press(":");
  expect(seen().open).toBe(false);
});

test("the colon opens the line right after a press ticks the switch, which keeps the focus", () => {
  started(false);
  input(BOX).click();
  input(BOX).focus();
  press(":");
  expect(seen().open).toBe(true);
});

test("the colon opens nothing while the reader types in a field of the page", () => {
  started();
  const other = document.createElement("textarea");
  document.body.append(other);
  other.focus();
  press(":");
  expect(seen().open).toBe(false);
});

test("the menu narrows to the places a letter starts, the letter typed apart", () => {
  started();
  press(":");
  type("t");
  const typed = shownPlaces().map((item) => item.querySelector(TYPED)?.textContent);
  expect({ places: seen().places, typed }).toEqual({ places: ["tidy-tool"], typed: ["t"] });
});

test.each([
  { name: "a post's headings", golden: postPage, typed: "a", want: ["afterwards"] },
  { name: "a post's other parts", golden: partPage, typed: "t", want: ["third", "tidy-tool"] },
  {
    name: "the list's posts, tags and series, a tag and a series named alike each standing",
    golden: listPage,
    typed: "r",
    want: ["reading-slowly", "reading", "reading"],
  },
])("the menu offers $name among the places a letter starts", ({ golden, typed, want }) => {
  started(true, () => undefined, golden);
  press(":");
  type(typed);
  expect(seen().places).toEqual(want);
});

test.each([
  {
    name: "a heading, at its heading",
    golden: postPage,
    typed: "in-a-terminal",
    want: ["/blog/reading-slowly#in-a-terminal"],
  },
  { name: "a post, at its page", golden: listPage, typed: "reading-slowly", want: ["/blog/reading-slowly"] },
  { name: "another part, at its page", golden: partPage, typed: "first", want: ["/blog/first"] },
  {
    name: "a tag, at the list it filters, the first of the places its name names",
    golden: listPage,
    typed: "reading",
    want: ["/blog?tag=reading"],
  },
])("Enter goes to $name", ({ golden, typed, want }) => {
  const { went } = started(true, () => undefined, golden);
  press(":");
  type(typed);
  press("Enter");
  expect(went).toEqual(want);
});

test("Enter goes to a series, at its group on the list by series, where Tab chose it beside the tag named alike", () => {
  const { went } = started(true, () => undefined, listPage);
  press(":");
  type("reading");
  press("Tab");
  press("Tab");
  press("Tab");
  press("Enter");
  expect(went).toEqual(["/blog/series#reading"]);
});

test.each([
  {
    name: "a page of the résumé, each project's and the blog's lists",
    golden: undefined,
    want: ["tidy-tool", "blog", "blog/series"],
  },
  {
    name: "a page of the blog, the blog's posts too",
    golden: listPage,
    want: ["tidy-tool", "blog", "blog/series", "blog/reading-slowly"],
  },
])("the menu after :e and a space on $name", ({ golden, want }) => {
  started(true, () => undefined, golden);
  press(":");
  type("e ");
  expect(seen().places).toEqual(want);
});

test("the menu hides, with the key that goes round it, where no place fits", () => {
  started();
  press(":");
  type("x");
  expect({ menu: found(PLACES).hidden, next: found(NEXT).hidden }).toEqual({ menu: true, next: true });
});

test("Backspace with letters typed leaves the line open, for the letter it takes back", () => {
  started();
  press(":");
  type("s");
  press("Backspace");
  expect(seen().open).toBe(true);
});

test.each([
  { name: "Tab chooses the first and writes it", keys: [{ key: "Tab", shiftKey: false }], line: "about" },
  {
    name: "Tab again chooses the next",
    keys: [
      { key: "Tab", shiftKey: false },
      { key: "Tab", shiftKey: false },
    ],
    line: "projects",
  },
  { name: "Shift+Tab chooses the last", keys: [{ key: "Tab", shiftKey: true }], line: "contact" },
])("$name, the focus kept on the line", ({ keys, line }) => {
  started();
  press(":");
  keys.forEach((held) => press(held.key, held));
  expect({ line: seen().line, chosen: seen().chosen, focused: document.activeElement === found(FIELD) }).toEqual({
    line,
    chosen: line,
    focused: true,
  });
});

test.each([
  { name: "a place by its name", typed: "skills", want: ["/skills"] },
  { name: "the project's page :e names", typed: "e tidy-tool", want: ["/p/tidy-tool"] },
  { name: "a post :e names by its number, in the blog's folder", typed: "e blog/1", want: ["/blog/1"] },
  { name: "nowhere on the tab shown", typed: "experience", want: [] },
])("Enter goes to $name and closes the line", ({ typed, want }) => {
  const { went } = started();
  press(":");
  type(typed);
  press("Enter");
  expect({ went, open: seen().open }).toEqual({ went: want, open: false });
});

test("Enter says why it refuses a command until the next key, the focus handed back, and goes nowhere", () => {
  const { went } = started();
  const link = found('[aria-current="page"]');
  link.focus();
  press(":");
  type("x");
  press("Enter");
  const refused = { said: seen().said, focused: document.activeElement === link };
  press("j");
  expect({ went, refused, after: seen().said }).toEqual({
    went: [],
    refused: { said: true, focused: true },
    after: false,
  });
});

test.each([
  { name: "on Esc", close: () => press("Escape") },
  { name: "on Backspace with nothing typed", close: () => press("Backspace") },
  { name: "when the focus leaves it", close: () => found(FIELD).blur() },
])("the line closes $name, going nowhere", ({ close }) => {
  const { went } = started();
  press(":");
  close();
  expect({ went, open: seen().open }).toEqual({ went: [], open: false });
});

test("the line closing as the focus leaves it leaves the focus where it went", () => {
  started();
  found('[aria-current="page"]').focus();
  press(":");
  const other = document.createElement("textarea");
  document.body.append(other);
  other.focus();
  expect({ open: seen().open, focused: document.activeElement === other }).toEqual({ open: false, focused: true });
});

test("Esc hands the focus back where it was before the colon", () => {
  started();
  const link = found('[aria-current="page"]');
  link.focus();
  press(":");
  press("Escape");
  expect(document.activeElement === link).toBe(true);
});

// The places of the page's one project, by its name and as :e opens it
const PROJECT = `${PLACE}[data-place="project"]`;
const EDITED_PROJECT = `${PROJECT}[data-edit]`;

/** The page with its one project named as a tab is, skills, by its name and as :e opens it. */
function projectNamedAsTab(page: Document): void {
  page.querySelectorAll<HTMLElement>(PROJECT).forEach((project) => project.setAttribute("data-name", "skills"));
}

test("a press on a project's page :e lists goes to the project, though a tab takes its name", () => {
  const { went } = started(true, projectNamedAsTab);
  press(":");
  type("e ");
  found(EDITED_PROJECT).click();
  expect(went).toEqual(["/p/tidy-tool"]);
});

test("a press on a place goes there", () => {
  const { went } = started();
  press(":");
  found(`${PLACE}[data-name="skills"]`).click();
  expect(went).toEqual(["/skills"]);
});
