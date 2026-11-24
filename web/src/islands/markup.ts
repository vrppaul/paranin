// How the server marks an island in its markup (internal/face/html): a figure naming its component, which the entry
// reads to start each shown widget's code and the islands read to run it. It imports no Svelte, so the script every
// page runs takes it whole and nothing more.

// An island, and the attribute that names its component
export const ISLAND = "[data-island]";
export const ISLAND_NAME = "data-island";

/** The names of the components a page's islands show, each once, in the order the page first shows them. */
export function islandNames(page: ParentNode): string[] {
  return [...new Set(Array.from(page.querySelectorAll(ISLAND), (island) => island.getAttribute(ISLAND_NAME) ?? ""))];
}
