// Finding words on a page with the vim keys, as the TUI's search finds them (internal/face/tui/search.go), with no
// page to touch: where a search matches the words of a passage, which match the page goes to by where it stands, a
// step round the matches, and what the line says of it in Neovim's words.

/** A piece of a passage: a text of the page, its words as they stand. */
interface Worded {
  readonly data: string;
}

/** A place in a passage: the piece a letter stands in, and how far into it. */
interface Place<Piece extends Worded> {
  piece: Piece;
  offset: number;
}

/** The words a block of the page sets, its pieces joined, and the place of each of their letters. */
export interface Passage<Piece extends Worded> {
  words: string;
  places: Place<Piece>[];
}

/** A match in a passage: the place of its first letter, and the place just after its last. */
export interface Match<Piece extends Worded> {
  start: Place<Piece>;
  end: Place<Piece>;
}

// The space the page keeps two words on one row with, as before a dot, and a Mac types with Option, written as its
// code so the code shows it; and the space a reader types
const NO_BREAK_SPACE = "\u00a0";
const SPACE = " ";

/** A block's pieces as a passage, a no-break space read as the space a reader types. */
export function passageOf<Piece extends Worded>(pieces: readonly Piece[]): Passage<Piece> {
  return {
    words: pieces
      .map((piece) => piece.data)
      .join("")
      .replaceAll(NO_BREAK_SPACE, SPACE),
    places: pieces.flatMap((piece) => [...Array(piece.data.length).keys()].map((offset) => ({ piece, offset }))),
  };
}

/** The matches of a search in a passage, whatever their case and their spaces' kind: a phrase parted is found whole. */
export function matchesIn<Piece extends Worded>(passage: Passage<Piece>, query: string): Match<Piece>[] {
  const wanted = query.replaceAll(NO_BREAK_SPACE, SPACE);
  return matchStarts(passage.words, wanted).flatMap((start) => matchAt(passage, start, wanted.length));
}

/** The match of a length that starts at a letter of a passage; none where it runs past the passage's letters. */
function matchAt<Piece extends Worded>(passage: Passage<Piece>, start: number, length: number): Match<Piece>[] {
  const first = passage.places[start];
  const last = passage.places[start + length - 1];
  if (first === undefined || last === undefined) {
    return [];
  }
  return [{ start: first, end: { piece: last.piece, offset: last.offset + 1 } }];
}

/**
 * The places a search starts in a text, whatever its case. Matches never overlap, as Neovim's do not: the next is
 * sought after the end of the one before.
 */
function matchStarts(text: string, query: string): number[] {
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
