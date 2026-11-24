// What a post's widget reads of the page it is shown in, as a Svelte context the island mounts the widget with and the
// widget, or a part of the kit it draws by, reads: the theme the page is shown in and whether the reader asks for
// motion held still, each current as it turns. The kit knows nothing of how the page decides its theme: the island
// does, by the page's own rule, and alone keeps it.
import { getContext } from "svelte";
import type { Theme } from "../tokens/colours";

/** The key a widget is mounted with the page it is shown in under, among its contexts. */
export const SHOWN_PAGE_KEY = Symbol("shown page");

/** The page as its reader is shown it, which a widget reads and never turns. */
export interface ShownPage {
  /** The theme the page is shown in. */
  readonly theme: Theme;
  /** Whether the reader asks for motion held still. */
  readonly reducedMotion: boolean;
}

/** The page a widget is shown in, read while the widget, or a part of the kit in it, starts. */
export function readShownPage(): ShownPage {
  const page = getContext<ShownPage | undefined>(SHOWN_PAGE_KEY);
  if (page === undefined) {
    throw new Error("the widget was mounted with no page");
  }
  return page;
}
