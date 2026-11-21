// The reader's theme (docs/WEB.md, "Theme"), over the markup internal/face/html draws: one button, hidden until this
// script shows it, which turns the page to the theme it is not in. A page follows the system's theme until the
// reader turns it; the theme turned to is set on the page's root as an attribute, which the page's stylesheet turns
// into its colour scheme (src/ui/ui.css), and kept in the browser's storage for every page after. A press that lands
// back on the system's theme forgets the choice, so the page follows the system again. The script holds the page's
// first paint until it has run, so a page is first seen in the theme chosen; a browser that knows no such hold
// (Firefox) may show it in the system's theme for a moment first. The colour scheme is never set as a style: the
// build writes each token's light-dark() pair as variables the stylesheet's rules switch, and a style set by script
// would pass them by.
import { forgetChoice, keepChoice, storedChoice, type ChoiceStore } from "./choice";

/** The themes a page is shown in, each as the stylesheet names it. */
const THEMES = ["light", "dark"] as const;
type Theme = (typeof THEMES)[number];

/** The theme a press turns each theme to. */
const OTHER_THEME: Record<Theme, Theme> = { light: "dark", dark: "light" };

// Where a theme chosen is kept between pages, and the attribute on the page's root the stylesheet reads it from
const STORAGE_KEY = "theme";
const THEME_ATTRIBUTE = "data-theme";
// The button as the server draws it: the part it stands in, hidden until this script shows it, and the button
const SWITCH = '[data-part="theme"]';
const BUTTON = "button";
// What says the button is pressed, to a screen reader: it is named for the dark theme, and pressed while the page is
// dark
const PRESSED = "aria-pressed";
const PRESSED_THEME: Theme = "dark";

/** The system's theme as a page asks it: whether it is dark, and a word whenever that changes. */
export type SystemScheme = Pick<MediaQueryList, "matches" | "addEventListener">;

/**
 * Sets a page in the theme its reader chose, where they chose one, and runs the page's theme button where it has
 * one: shows it, presses it while the page is dark, and turns the page to the other theme on a press. storage is asked
 * for the browser's storage each time it is needed, since a browser may refuse to hand it over, or to read or keep
 * what it holds: a page with no storage it may reach still turns, for itself alone.
 */
export function startTheme(page: Document, storage: () => ChoiceStore, system: SystemScheme): void {
  const root = page.documentElement;
  const chosen = themeNamed(storedChoice(storage, STORAGE_KEY));
  if (chosen !== undefined) {
    root.setAttribute(THEME_ATTRIBUTE, chosen);
  }
  const themeSwitch = page.querySelector<HTMLElement>(SWITCH);
  if (themeSwitch === null) {
    return;
  }
  const button = themeSwitch.querySelector<HTMLButtonElement>(BUTTON);
  if (button === null) {
    return;
  }
  button.addEventListener("click", () => turn(root, button, storage, system));
  system.addEventListener("change", () => pressShown(root, button, system));
  pressShown(root, button, system);
  themeSwitch.hidden = false;
}

/**
 * Turns a page to the theme it is not in and presses its button for it: the theme is kept as the reader's choice,
 * unless it is the system's, which forgets the choice so the page follows the system again.
 */
function turn(root: HTMLElement, button: HTMLButtonElement, storage: () => ChoiceStore, system: SystemScheme): void {
  const turned = OTHER_THEME[shownTheme(root, system)];
  if (turned === systemTheme(system)) {
    root.removeAttribute(THEME_ATTRIBUTE);
    forgetChoice(storage, STORAGE_KEY);
  } else {
    root.setAttribute(THEME_ATTRIBUTE, turned);
    keepChoice(storage, STORAGE_KEY, turned);
  }
  pressShown(root, button, system);
}

/** Presses the button while the page is shown dark: in the theme chosen, or else the system's. */
function pressShown(root: HTMLElement, button: HTMLButtonElement, system: SystemScheme): void {
  button.setAttribute(PRESSED, String(shownTheme(root, system) === PRESSED_THEME));
}

/** The theme a page is shown in: the one chosen, or else the system's. */
function shownTheme(root: HTMLElement, system: SystemScheme): Theme {
  return themeNamed(root.getAttribute(THEME_ATTRIBUTE)) ?? systemTheme(system);
}

function systemTheme(system: SystemScheme): Theme {
  return system.matches ? "dark" : "light";
}

/** The theme a word names, if it names one. */
function themeNamed(word: string | null): Theme | undefined {
  return THEMES.find((theme) => theme === word);
}
