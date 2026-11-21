// Finding words on a page with the vim keys, as the TUI's search finds them (internal/face/tui/search.go), with no
// page to touch: where a search matches a text, which match the page goes to by where it stands, a step round the
// matches, and what the line says of it in Neovim's words.

/**
 * The places a search starts in a text, whatever its case. Matches never overlap, as Neovim's do not: the next is
 * sought after the end of the one before.
 */
export function matchStarts(text: string, query: string): number[] {
  const starts: number[] = [];
  if (query === "") {
    return starts;
  }
  const wanted = query.toLowerCase();
  for (let start = 0; start + query.length <= text.length;) {
    if (text.slice(start, start + query.length).toLowerCase() === wanted) {
      starts.push(start);
      start += query.length;
    } else {
      start += 1;
    }
  }
  return starts;
}

/** The place of the first match at a point down the page or after it, or the first of all after the last. */
export function placeAtOrAfter(tops: readonly number[], point: number): number {
  return Math.max(
    tops.findIndex((top) => top >= point),
    0,
  );
}

/** The place of the last match before a point down the page, or the last of all before the first. */
export function placeBefore(tops: readonly number[], point: number): number {
  const before = tops.findLastIndex((top) => top < point);
  return before < 0 ? tops.length - 1 : before;
}

/** The place a step from another goes to among a count of matches, round from the last to the first. */
export function steppedPlace(place: number, step: number, count: number): number {
  // the remainder taken twice, so a step back from the first comes round to the last
  return (((place + step) % count) + count) % count;
}

/** Which match of how many a place among them is, as Neovim counts them: "[2/5]". */
export function matchCount(place: number, count: number): string {
  return `[${String(place + 1)}/${String(count)}]`;
}

/** What the line says once a search shows a match: the search, then which of how many. */
export function foundSaid(query: string, place: number, count: number): string {
  return `/${query} ${matchCount(place, count)}`;
}

/** What the line says of a search that found nothing, as Neovim says it. */
export function notFoundSaid(query: string): string {
  return `E486: Pattern not found: ${query}`;
}
