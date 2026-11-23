// What stands over the page, which Esc puts away from the top, asked of the parts of the page the server draws: the
// HTML face's golden page (internal/face/html/testdata), its command line's field and its theme button's tooltip.
// jsdom shows no tooltip under a pointer and opens no window, so each case says which stand.
import { expect, test } from "vitest";
import { Overlay, type Layer } from "../../src/controllers/overlay";
import experiencePage from "../../../internal/face/html/testdata/experience.html?raw";

// The line's field, a field a reader types in; the theme button's tooltip; and the button, which a key may come to
const FIELD = '[data-part="command-field"]';
const TOOLTIP = '[data-part="theme"] [data-part="tooltip"]';
const BUTTON = '[data-part="theme"] button';

/** A key's target and what stands over the page as it comes, and the kind of layer on top. */
interface RankCase {
  name: string;
  target: string;
  shownTooltips: string[];
  windowOpen: boolean;
  top: Layer["kind"];
}

const RANKS: RankCase[] = [
  {
    name: "a field typed in, over a tooltip and a window",
    target: FIELD,
    shownTooltips: [TOOLTIP],
    windowOpen: true,
    top: "field",
  },
  {
    name: "a tooltip shown, over a window",
    target: BUTTON,
    shownTooltips: [TOOLTIP],
    windowOpen: true,
    top: "tooltips",
  },
  { name: "a window open, over the page", target: BUTTON, shownTooltips: [], windowOpen: true, top: "window" },
  { name: "the page, with nothing over it", target: BUTTON, shownTooltips: [], windowOpen: false, top: "page" },
];

/** The golden page's element a selector finds. */
function found(page: Document, selector: string): HTMLElement {
  const element = page.querySelector<HTMLElement>(selector);
  if (element === null) {
    throw new Error(`the page has no ${selector}`);
  }
  return element;
}

test.each(RANKS)("on top: $name", ({ target, shownTooltips, windowOpen, top }) => {
  const page = new DOMParser().parseFromString(experiencePage, "text/html");
  const shown = shownTooltips.map((selector) => found(page, selector));
  const overlay = new Overlay(
    () => shown,
    () => windowOpen,
  );
  expect(overlay.top(found(page, target)).kind).toBe(top);
});
