// The command line the vim keys open, run over the page the server draws: the HTML face's golden page of the
// résumé's experience tab (internal/face/html/testdata), which that face's own tests hold to what it writes. The
// addresses the line goes to are handed to the test, in place of the browser.
import { expect, test } from "vitest";
import { startCommandLine } from "../../src/controllers/commandline";
import experiencePage from "../../../internal/face/html/testdata/experience.html?raw";

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

/** The document's element a selector finds, which the golden page draws. */
function found(selector: string): HTMLElement {
  const element = document.querySelector<HTMLElement>(selector);
  if (element === null) {
    throw new Error(`the page has no ${selector}`);
  }
  return element;
}

/** An input of the golden page a selector finds: the switch's box or the line's field. */
function input(selector: string): HTMLInputElement {
  const element = document.querySelector<HTMLInputElement>(selector);
  if (element === null) {
    throw new Error(`the page has no ${selector}`);
  }
  return element;
}

/** The golden page with its command line run, the vim keys ticked or not, once a change of the test's is made to it;
 * the addresses the line goes to land in the list returned. */
function started(ticked = true, changed: (page: Document) => void = () => undefined): string[] {
  const parsed = new DOMParser().parseFromString(experiencePage, "text/html");
  changed(parsed);
  document.body.replaceWith(document.adoptNode(parsed.body));
  const went: string[] = [];
  startCommandLine(document, (address) => went.push(address));
  // the switch's own script shows it on the page
  found(SWITCH).hidden = false;
  input(BOX).checked = ticked;
  return went;
}

/** A key pressed where the focus is, or on the page where nothing holds it. */
function press(key: string, held: KeyboardEventInit = {}): void {
  const target = document.activeElement ?? document.body;
  target.dispatchEvent(new KeyboardEvent("keydown", { key, bubbles: true, cancelable: true, ...held }));
}

/** Letters typed on the line, after what is there. */
function type(letters: string): void {
  const field = input(FIELD);
  field.value += letters;
  field.dispatchEvent(new Event("input", { bubbles: true }));
}

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

test("the menu lists every project's page after :e and a space", () => {
  started();
  press(":");
  type("e ");
  expect(seen().places).toEqual(["tidy-tool"]);
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
  { name: "nowhere on the tab shown", typed: "experience", want: [] },
])("Enter goes to $name and closes the line", ({ typed, want }) => {
  const went = started();
  press(":");
  type(typed);
  press("Enter");
  expect({ went, open: seen().open }).toEqual({ went: want, open: false });
});

test("Enter says why it refuses a command until the next key, the focus handed back, and goes nowhere", () => {
  const went = started();
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
  const went = started();
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

/** The page with its one project's place named as a tab is, skills. */
function projectNamedAsTab(page: Document): void {
  const project = page.querySelector<HTMLElement>(`${PLACE}[data-place="project"]`);
  project?.setAttribute("data-name", "skills");
}

test("a press on a project's page :e lists goes to the project, though a tab takes its name", () => {
  const went = started(true, projectNamedAsTab);
  press(":");
  type("e ");
  found(`${PLACE}[data-place="project"]`).click();
  expect(went).toEqual(["/p/tidy-tool"]);
});

test("a press on a place goes there", () => {
  const went = started();
  press(":");
  found(`${PLACE}[data-name="skills"]`).click();
  expect(went).toEqual(["/skills"]);
});
