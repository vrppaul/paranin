// The reader's choice of theme, run over the pages the server draws: the HTML face's golden pages
// (internal/face/html/testdata), which that face's own tests hold to what it writes, so a change to the markup the
// controller runs over fails here too.
import { beforeEach, expect, test } from "vitest";
import { startTheme, type ChoiceStore, type SystemScheme } from "../../src/controllers/theme";
import resumePage from "../../../internal/face/html/testdata/index.html?raw";
import missingPage from "../../../internal/face/html/testdata/notfound.html?raw";

// What the controller writes on a page, as the page's stylesheet and a screen reader read it
const THEME_ATTRIBUTE = "data-theme";
const PRESSED = "aria-pressed";
// Where the controller keeps a choice in the browser's storage
const STORAGE_KEY = "theme";
// The choice's markup as the server draws it
const CHOICE = '[data-part="theme"]';
const BUTTONS = "button[data-theme-choice]";

/** The system's theme, dark or not, which a test turns as a reader turns their system's. */
class FakeSystem extends EventTarget implements SystemScheme {
  matches: boolean;

  constructor(dark: boolean) {
    super();
    this.matches = dark;
  }

  turn(dark: boolean): void {
    this.matches = dark;
    this.dispatchEvent(new Event("change"));
  }
}

/** A browser that keeps nothing: its storage refuses to be reached, as it does where a reader blocks it. */
function refusedStorage(): ChoiceStore {
  throw new DOMException("storage is blocked", "SecurityError");
}

/** A page as a browser parses what the server sent. */
function parsed(markup: string): Document {
  return new DOMParser().parseFromString(markup, "text/html");
}

/** The theme a page is set in, where it is set in one. */
function themeOf(page: Document): string | null {
  return page.documentElement.getAttribute(THEME_ATTRIBUTE);
}

/** Whether each button of a page's choice says it is pressed, by the theme it chooses. */
function pressed(page: Document): Record<string, string | null> {
  const buttons = [...page.querySelectorAll<HTMLButtonElement>(BUTTONS)];
  return Object.fromEntries(
    buttons.map((button) => [String(button.dataset["themeChoice"]), button.getAttribute(PRESSED)]),
  );
}

/** The button of a page's choice that chooses a theme. */
function buttonFor(page: Document, theme: string): HTMLButtonElement {
  const found = [...page.querySelectorAll<HTMLButtonElement>(BUTTONS)].find(
    (button) => button.dataset["themeChoice"] === theme,
  );
  if (found === undefined) {
    throw new Error(`the page has no button choosing ${theme}`);
  }
  return found;
}

beforeEach(() => localStorage.clear());

test("a theme the reader chooses is the page's at once, and every next page's before it is seen", () => {
  const first = parsed(resumePage);
  startTheme(first, () => localStorage, new FakeSystem(false));
  buttonFor(first, "dark").click();
  const next = parsed(resumePage);
  startTheme(next, () => localStorage, new FakeSystem(false));
  const missing = parsed(missingPage);
  startTheme(missing, () => localStorage, new FakeSystem(false));
  expect({
    first: themeOf(first),
    firstPressed: pressed(first),
    next: themeOf(next),
    nextPressed: pressed(next),
    missing: themeOf(missing),
  }).toEqual({
    first: "dark",
    firstPressed: { light: "false", dark: "true" },
    next: "dark",
    nextPressed: { light: "false", dark: "true" },
    missing: "dark",
  });
});

test("a page the reader chose no theme on follows the system's, and says which it is in", () => {
  const page = parsed(resumePage);
  const system = new FakeSystem(true);
  startTheme(page, () => localStorage, system);
  const whileDark = pressed(page);
  system.turn(false);
  expect({
    theme: themeOf(page),
    whileDark,
    whileLight: pressed(page),
  }).toEqual({
    theme: null,
    whileDark: { light: "false", dark: "true" },
    whileLight: { light: "true", dark: "false" },
  });
});

test("a word kept that names no theme is no choice: the page follows the system's", () => {
  localStorage.setItem(STORAGE_KEY, "blue");
  const page = parsed(resumePage);
  startTheme(page, () => localStorage, new FakeSystem(true));
  expect({ theme: themeOf(page), pressed: pressed(page) }).toEqual({
    theme: null,
    pressed: { light: "false", dark: "true" },
  });
});

test("the choice shows only where the script runs: the server draws it hidden", () => {
  const page = parsed(resumePage);
  const drawnHidden = page.querySelector<HTMLElement>(CHOICE)?.hidden;
  startTheme(page, () => localStorage, new FakeSystem(false));
  expect({
    drawnHidden,
    shown: page.querySelector<HTMLElement>(CHOICE)?.hidden === false,
  }).toEqual({
    drawnHidden: true,
    shown: true,
  });
});

test("a choice the browser will not keep still holds for the page", () => {
  const page = parsed(resumePage);
  startTheme(page, refusedStorage, new FakeSystem(false));
  buttonFor(page, "dark").click();
  expect({ theme: themeOf(page), pressed: pressed(page) }).toEqual({
    theme: "dark",
    pressed: { light: "false", dark: "true" },
  });
});
