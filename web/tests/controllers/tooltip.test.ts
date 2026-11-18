// The tooltips, run over the résumé the server draws: the HTML face's golden page (internal/face/html/testdata),
// which that face's own tests hold to what it writes. The page is the test's own document, so focus moves in it as
// it does in a browser.
import { beforeEach, expect, test } from "vitest";
import { startTooltips } from "../../src/controllers/tooltip";
import resumePage from "../../../internal/face/html/testdata/index.html?raw";

// The tooltip as the server draws it, the part of the control whose tooltip it is, which the server draws hidden,
// the control, and the key that hides the tooltip
const TOOLTIP = '[data-part="tooltip"]';
const SHOWN_PART = '[data-part="theme"]';
const CONTROL = `${SHOWN_PART} button`;
const DISMISS_KEY = "Escape";

/** The document's element a selector finds, which the golden page draws. */
function found(selector: string): HTMLElement {
  const element = document.querySelector<HTMLElement>(selector);
  if (element === null) {
    throw new Error(`the page has no ${selector}`);
  }
  return element;
}

/** Presses a key where the focus is, as a reader does. */
function press(key: string): void {
  (document.activeElement ?? document.body).dispatchEvent(new KeyboardEvent("keydown", { key, bubbles: true }));
}

beforeEach(() => {
  const parsed = new DOMParser().parseFromString(resumePage, "text/html");
  document.documentElement.innerHTML = parsed.documentElement.innerHTML;
  // the control's part shows, as the site's script shows it, so its control takes the focus
  found(SHOWN_PART).hidden = false;
  startTooltips(document);
});

test("Esc hides the tooltip of the control in focus, until the focus leaves it", () => {
  found(CONTROL).focus();
  press(DISMISS_KEY);
  const afterEsc = found(TOOLTIP).hidden;
  found(CONTROL).blur();
  expect({ afterEsc, afterLeaving: found(TOOLTIP).hidden }).toEqual({ afterEsc: true, afterLeaving: false });
});

test("the pointer leaving the control's part shows no tooltip Esc hid while the focus still holds it", () => {
  found(CONTROL).focus();
  press(DISMISS_KEY);
  found(SHOWN_PART).dispatchEvent(new PointerEvent("pointerleave"));
  expect(found(TOOLTIP).hidden).toBe(true);
});

test("Esc leaves a tooltip alone while its control is neither in focus nor under the pointer", () => {
  press(DISMISS_KEY);
  expect(found(TOOLTIP).hidden).toBe(false);
});

test("another key hides no tooltip", () => {
  found(CONTROL).focus();
  press("Enter");
  expect(found(TOOLTIP).hidden).toBe(false);
});
