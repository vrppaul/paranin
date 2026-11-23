// The tooltips (DESIGN.md, Tooltip and Toast), over the markup internal/face/html draws: a word or two about a
// control, which the page's stylesheet shows while the control's part is under the pointer or holds the keyboard's
// focus (src/ui/ui.css). Esc hides the tooltip showing until the pointer and the focus have both left its control's
// part, as WCAG 1.4.13 asks of what shows on hover and on focus, so a reader can put it away without moving. A tooltip
// shown stands over all but a field typed in (./overlay.ts): the Esc that hides it goes no further.
import type { Overlay } from "./overlay";

// A tooltip as the server draws it, inside the part of the control it is about; and the key that hides one
const TOOLTIP = '[data-part="tooltip"]';
const DISMISS_KEY = "Escape";
// What says a control's part shows its tooltip, as the stylesheet asks it: the pointer over it, or the keyboard's
// focus on its control. A control focused by a click holds no keyboard's focus, and shows no tooltip
const UNDER_POINTER = ":hover";
const HOLDS_FOCUS = ":has(> :focus-visible)";

/**
 * Shows each tooltip of a page again once its control's part is left, after Esc hid it; the tooltips that show at a
 * moment, found once, for the overlay they top.
 */
export function startTooltips(page: Document): () => readonly HTMLElement[] {
  const tooltips = [...page.querySelectorAll<HTMLElement>(TOOLTIP)];
  for (const tooltip of tooltips) {
    const part = tooltip.parentElement;
    part?.addEventListener("pointerleave", () => showAgain(tooltip, part, HOLDS_FOCUS));
    part?.addEventListener("focusout", () => showAgain(tooltip, part, UNDER_POINTER));
  }
  return () => tooltips.filter(shows);
}

/**
 * Lets Esc hide the tooltips that show while they stand on top of the overlay. jsdom shows no tooltip while a field
 * holds the focus, so their yielding to a field typed in is held by tests/controllers/overlay.test.ts alone.
 */
export function hideOnEsc(page: Document, overlay: Overlay): void {
  // heard first, as the top hears Esc before what stands under it
  page.addEventListener("keydown", (event) => hideShowing(event, overlay), { capture: true });
}

/** Whether a tooltip shows: Esc has not hidden it, and its control's part is under the pointer or holds the focus. */
function shows(tooltip: HTMLElement): boolean {
  const part = tooltip.parentElement;
  return !tooltip.hidden && (part?.matches(UNDER_POINTER) === true || part?.matches(HOLDS_FOCUS) === true);
}

/**
 * Hides, on Esc while tooltips stand on top, each that shows; the key then reaches nothing under them, neither the
 * page's listeners nor the browser, which would close a window.
 */
function hideShowing(event: KeyboardEvent, overlay: Overlay): void {
  if (event.key !== DISMISS_KEY) {
    return;
  }
  const layer = overlay.top(event.target);
  if (layer.kind !== "tooltips") {
    return;
  }
  for (const tooltip of layer.shown) {
    tooltip.hidden = true;
  }
  event.preventDefault();
  event.stopPropagation();
}

/**
 * Lets a hidden tooltip show again once its control's part is left by the pointer or the focus, unless the other
 * still holds it: the tooltip would show at once, after the reader put it away.
 */
function showAgain(tooltip: HTMLElement, part: HTMLElement, stillHeld: string): void {
  if (!part.matches(stillHeld)) {
    tooltip.hidden = false;
  }
}
