// What stands over the page, which Esc puts away one thing a press, from the top (DESIGN.md, CommandLine): a field
// being typed in, which answers its own Esc, then a tooltip shown, then a window open, a way's hint or a series' pick,
// then the page's own, a search kept. Each Esc listener asks it, so one press never puts away two things.

// The fields of a page a reader types in, which keep the keys typed: never a box or a button, as the vim keys' switch
// is, which a press leaves the focus on
const TYPED_IN =
  "textarea, select, input:not([type=checkbox], [type=radio], [type=button], [type=submit], [type=reset])";

/** What stands on top of the page for a key: a field typed in, the tooltips shown, a window open, or the page. */
export type Layer =
  { kind: "field" } | { kind: "tooltips"; shown: readonly HTMLElement[] } | { kind: "window" } | { kind: "page" };

// The layers that carry nothing but what they are
const FIELD: Layer = { kind: "field" };
const WINDOW: Layer = { kind: "window" };
const PAGE: Layer = { kind: "page" };

/** Whether a key comes while the reader types in a field of the page, which keeps its keys. */
export function typedIn(target: EventTarget | null): boolean {
  return target instanceof HTMLElement && (target.isContentEditable || target.matches(TYPED_IN));
}

/** What stands over a page: the tooltips it shows now, and whether a window stands open over it. */
export class Overlay {
  constructor(
    private readonly shownTooltips: () => readonly HTMLElement[],
    private readonly windowOpen: () => boolean,
  ) {}

  /** What stands on top for a key come to a target, which an Esc there puts away. */
  top(target: EventTarget | null): Layer {
    if (typedIn(target)) {
      return FIELD;
    }
    const shown = this.shownTooltips();
    if (shown.length > 0) {
      return { kind: "tooltips", shown };
    }
    return this.windowOpen() ? WINDOW : PAGE;
  }
}
