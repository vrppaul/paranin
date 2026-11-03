// The reader's choice of theme (docs/WEB.md, "Theme"), over the markup internal/face/html draws: a group of two
// buttons, hidden until this script shows it, each choosing a theme. A page follows the system's theme until the
// reader chooses one; the theme chosen is set on the page's root as an attribute, which the page's stylesheet
// turns into its colour scheme (src/ui/ui.css), and kept in the browser's storage for every page after. The
// script holds the page's first paint until it has run, so a page is first seen in the theme chosen; a browser that
// knows no such hold (Firefox) may show it in the system's theme for a moment first. The colour scheme is
// never set as a style: the build writes each token's light-dark() pair as variables the stylesheet's rules
// switch, and a style set by script would pass them by.

/** The themes a reader may choose, each as the stylesheet and the server's markup name it. */
const THEMES = ["light", "dark"] as const;
type Theme = (typeof THEMES)[number];

// Where a theme chosen is kept between pages, and the attribute on the page's root the stylesheet reads it from
const STORAGE_KEY = "theme";
const THEME_ATTRIBUTE = "data-theme";
// The choice as the server draws it: the group, and each button by the theme it chooses
const CHOICE = '[data-part="theme"]';
const BUTTONS = "button[data-theme-choice]";
// What says a button's theme is the page's, to a screen reader and, once the look draws it, to the eye
const PRESSED = "aria-pressed";

/** Where a choice is kept: the browser's storage, of which the choice reads and writes one key. */
export type ChoiceStore = Pick<Storage, "getItem" | "setItem">;

/** The system's theme as a page asks it: whether it is dark, and a word whenever that changes. */
export type SystemScheme = Pick<MediaQueryList, "matches" | "addEventListener">;

/** A theme's button in the choice. */
interface ThemeButton {
  element: HTMLButtonElement;
  theme: Theme;
}

/**
 * Sets a page in the theme its reader chose, where they chose one, and runs the page's choice where it has one:
 * shows it, presses the theme the page is in, and keeps a theme pressed as the reader's choice. storage is asked
 * for the browser's storage each time it is needed, since a browser may refuse to hand it over, or to read or keep
 * what it holds: a page with no storage it may reach still takes a choice, for itself alone.
 */
export function startTheme(page: Document, storage: () => ChoiceStore, system: SystemScheme): void {
  const root = page.documentElement;
  const chosen = storedTheme(storage);
  if (chosen !== undefined) {
    root.setAttribute(THEME_ATTRIBUTE, chosen);
  }
  const choice = page.querySelector<HTMLElement>(CHOICE);
  if (choice === null) {
    return;
  }
  const buttons = themeButtons(choice);
  for (const button of buttons) {
    button.element.addEventListener("click", () => choose(root, storage, buttons, button.theme));
  }
  system.addEventListener("change", () => pressShown(root, buttons, system));
  pressShown(root, buttons, system);
  choice.hidden = false;
}

/** The buttons of a choice, each with the theme it chooses. */
function themeButtons(choice: HTMLElement): ThemeButton[] {
  return [...choice.querySelectorAll<HTMLButtonElement>(BUTTONS)].flatMap(themeButton);
}

/** A button of a choice with the theme it chooses, or none where it names no theme. */
function themeButton(element: HTMLButtonElement): ThemeButton[] {
  const theme = themeNamed(element.dataset["themeChoice"]);
  return theme === undefined ? [] : [{ element, theme }];
}

/** Sets a page in a theme the reader chose, keeps it for the pages after, and presses its button. */
function choose(root: HTMLElement, storage: () => ChoiceStore, buttons: ThemeButton[], theme: Theme): void {
  root.setAttribute(THEME_ATTRIBUTE, theme);
  keepTheme(storage, theme);
  press(buttons, theme);
}

/** Presses the button of the theme a page is shown in: the one chosen, or else the system's. */
function pressShown(root: HTMLElement, buttons: ThemeButton[], system: SystemScheme): void {
  press(buttons, themeNamed(root.getAttribute(THEME_ATTRIBUTE)) ?? systemTheme(system));
}

function press(buttons: ThemeButton[], theme: Theme): void {
  for (const button of buttons) {
    button.element.setAttribute(PRESSED, String(button.theme === theme));
  }
}

function systemTheme(system: SystemScheme): Theme {
  return system.matches ? "dark" : "light";
}

/** The theme a word names, if it names one. */
function themeNamed(word: string | null | undefined): Theme | undefined {
  return THEMES.find((theme) => theme === word);
}

/** The theme the reader chose on an earlier page, where the browser kept it. */
function storedTheme(storage: () => ChoiceStore): Theme | undefined {
  try {
    return themeNamed(storage().getItem(STORAGE_KEY));
  } catch {
    // The browser hands over no storage, or will not read it: the reader has chosen nothing it can tell
    return undefined;
  }
}

/** Keeps a theme chosen for the pages after, where the browser will keep it. */
function keepTheme(storage: () => ChoiceStore, theme: Theme): void {
  try {
    storage().setItem(STORAGE_KEY, theme);
  } catch {
    // The browser keeps nothing: the choice holds for this page alone
  }
}
