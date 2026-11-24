// The reader's theme, run over the pages the server draws: the HTML face's golden pages
// (internal/face/html/testdata), which that face's own tests hold to what it writes, so a change to the markup the
// controller runs over fails here too.
import { beforeEach, expect, test } from "vitest";
import type { ChoiceStore } from "../../src/controllers/choice";
import { startTheme } from "../../src/controllers/theme";
import { FakeSystem } from "../browser";
import resumePage from "../../../internal/face/html/testdata/index.html?raw";
import missingPage from "../../../internal/face/html/testdata/notfound.html?raw";

// What the controller writes on a page, as the page's stylesheet and a screen reader read it
const THEME_ATTRIBUTE = "data-theme";
const PRESSED = "aria-pressed";
// Where the controller keeps a choice in the browser's storage
const STORAGE_KEY = "theme";
// The button's markup as the server draws it: the part it stands in, and the button
const SWITCH = '[data-part="theme"]';
const BUTTON = `${SWITCH} button`;

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

/** The page's theme button. */
function buttonOf(page: Document): HTMLButtonElement {
  const button = page.querySelector<HTMLButtonElement>(BUTTON);
  if (button === null) {
    throw new Error("the page has no theme button");
  }
  return button;
}

/** Whether the page's theme button says it is pressed, as it does while the page is dark. */
function pressed(page: Document): string | null {
  return buttonOf(page).getAttribute(PRESSED);
}

beforeEach(() => localStorage.clear());

test("a press turns the page to the theme it is not in, and every next page with it, before it is seen", () => {
  const first = parsed(resumePage);
  startTheme(first, () => localStorage, new FakeSystem(false));
  buttonOf(first).click();
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
    firstPressed: "true",
    next: "dark",
    nextPressed: "true",
    missing: "dark",
  });
});

test("a press that lands back on the system's theme forgets the choice: the page follows the system again", () => {
  const page = parsed(resumePage);
  startTheme(page, () => localStorage, new FakeSystem(false));
  buttonOf(page).click();
  buttonOf(page).click();
  const next = parsed(resumePage);
  const system = new FakeSystem(false);
  startTheme(next, () => localStorage, system);
  system.turn(true);
  expect({
    theme: themeOf(page),
    pressed: pressed(page),
    kept: localStorage.getItem(STORAGE_KEY),
    nextTheme: themeOf(next),
    nextPressedOnceTheSystemIsDark: pressed(next),
  }).toEqual({
    theme: null,
    pressed: "false",
    kept: null,
    nextTheme: null,
    nextPressedOnceTheSystemIsDark: "true",
  });
});

test("a page the reader chose no theme on follows the system's, and says whether it is dark", () => {
  const page = parsed(resumePage);
  const system = new FakeSystem(true);
  startTheme(page, () => localStorage, system);
  const whileDark = pressed(page);
  system.turn(false);
  expect({ theme: themeOf(page), whileDark, whileLight: pressed(page) }).toEqual({
    theme: null,
    whileDark: "true",
    whileLight: "false",
  });
});

test("a word kept that names no theme is no choice: the page follows the system's", () => {
  localStorage.setItem(STORAGE_KEY, "blue");
  const page = parsed(resumePage);
  startTheme(page, () => localStorage, new FakeSystem(true));
  expect({ theme: themeOf(page), pressed: pressed(page) }).toEqual({ theme: null, pressed: "true" });
});

test("the button shows only where the script runs: the server draws it hidden", () => {
  const page = parsed(resumePage);
  const drawnHidden = page.querySelector<HTMLElement>(SWITCH)?.hidden;
  startTheme(page, () => localStorage, new FakeSystem(false));
  expect({
    drawnHidden,
    shown: page.querySelector<HTMLElement>(SWITCH)?.hidden === false,
  }).toEqual({ drawnHidden: true, shown: true });
});

test("a choice the browser will not keep still holds for the page, and still lands back on the system's", () => {
  const page = parsed(resumePage);
  startTheme(page, refusedStorage, new FakeSystem(false));
  buttonOf(page).click();
  const pressedOnce = { theme: themeOf(page), pressed: pressed(page) };
  buttonOf(page).click();
  expect({ pressedOnce, pressedTwice: { theme: themeOf(page), pressed: pressed(page) } }).toEqual({
    pressedOnce: { theme: "dark", pressed: "true" },
    pressedTwice: { theme: null, pressed: "false" },
  });
});
