// The tooltips (DESIGN.md, Tooltip and Toast), over the markup internal/face/html draws: a word or two about a
// control, which the page's stylesheet shows while the control's part is under the pointer or holds the keyboard's
// focus (src/ui/ui.css). Esc hides the tooltip showing until the pointer and the focus have both left its control's
// part, as WCAG 1.4.13 asks of what shows on hover and on focus, so a reader can put it away without moving.

// A tooltip as the server draws it, inside the part of the control it is about; and the key that hides one
const TOOLTIP = '[data-part="tooltip"]';
const DISMISS_KEY = "Escape";
// What says a control's part shows its tooltip, as the stylesheet asks it: the pointer over it, or the keyboard's
// focus on its control. A control focused by a click holds no keyboard's focus, and shows no tooltip
const UNDER_POINTER = ":hover";
const HOLDS_FOCUS = ":has(> :focus-visible)";

/** Lets Esc hide every tooltip of a page that shows, and shows each again once its control's part is left. */
export function startTooltips(page: Document): void {
  const tooltips = [...page.querySelectorAll<HTMLElement>(TOOLTIP)];
  for (const tooltip of tooltips) {
    const part = tooltip.parentElement;
    part?.addEventListener("pointerleave", () => showAgain(tooltip, part, HOLDS_FOCUS));
    part?.addEventListener("focusout", () => showAgain(tooltip, part, UNDER_POINTER));
  }
  page.addEventListener("keydown", (event) => hideShowing(event, tooltips));
}

/**
 * Hides, on Esc, each tooltip whose control's part is under the pointer or holds the focus: those the stylesheet
 * shows.
 */
function hideShowing(event: KeyboardEvent, tooltips: HTMLElement[]): void {
  if (event.key !== DISMISS_KEY) {
    return;
  }
  for (const tooltip of tooltips) {
    const part = tooltip.parentElement;
    if (part?.matches(UNDER_POINTER) === true || part?.matches(HOLDS_FOCUS) === true) {
      tooltip.hidden = true;
    }
  }
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
