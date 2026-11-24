// The vim keys' search on a page (DESIGN.md, CommandLine), as the TUI's (internal/face/tui/search.go): the words of
// the page's main part a search matches, marked on the page, the page brought to the one shown, and n and N going
// round them. The browser's highlights and layout stand behind interfaces the tests stand in for: jsdom has neither;
// which element sets a text's words apart is read off its computed display, which jsdom gives by its default sheet.
import {
  foundSaid,
  matchCount,
  matchesIn,
  notFoundSaid,
  passageOf,
  placeAtOrAfter,
  placeBefore,
  steppedPlace,
  type Match,
  type Passage,
} from "../domain/search";
import type { SaidPiece } from "../domain/said";
import { forgetChoice, keepChoice, storedChoice, type ChoiceStore } from "./choice";

// Where the browser keeps the search the reader kept, for the pages they open after in the same session
const KEPT_SEARCH = "search";

// The foot line, which a match brought into sight stands clear of
const FOOT = '[data-part="foot"]';

// The highlights the stylesheet paints a search's matches by (::highlight): the one the count names, and the rest
const CURRENT_MATCH = "match-current";
const OTHER_MATCHES = "match";

// n's step on to the next match, and N's back
const ONWARD_STEP = 1;
const BACK_STEP = -1;

/** Where a search's matches are marked on the page: the one the count names, and the rest. */
export interface Marks {
  mark(current: Range | undefined, others: readonly Range[]): void;
  clear(): void;
}

/** What a search needs of the page's layout: which words it shows, where a match stands, and the page's scroll. */
export interface Layout {
  /** Whether the page shows a text, whose words a search then matches. */
  shows(text: Node): boolean;
  /** How far down the page a match stands. */
  topOf(match: Range): number;
  /** Scrolls the page the least that stands a match in sight, clear of the foot line. */
  bringIntoSight(match: Range): void;
  /** How far down the page is scrolled. */
  scrolled(): number;
  scrollTo(point: number): void;
}

/** What the line says of a search or a command: its words, Neovim's and the reader's, and whether they say it failed. */
export interface Said {
  pieces: SaidPiece[];
  failed: boolean;
}

/** The way n or N goes: its step from the match shown, and the match it picks by where the page stands. */
export interface MatchWay {
  step: number;
  byThePage: (tops: readonly number[], point: number) => number;
}

/** n's way: on to the next match, or the first at or after where the page stands. */
export const ONWARD: MatchWay = { step: ONWARD_STEP, byThePage: placeAtOrAfter };

/** N's way: back to the match before, or the last before where the page stands. */
export const BACK: MatchWay = { step: BACK_STEP, byThePage: placeBefore };

/** A search and its matches, with the place among them of the one it shows. */
interface Found {
  query: string;
  matches: readonly Range[];
  place: number;
}

// A search of nothing, which matches nothing
const NOTHING: Found = { query: "", matches: [], place: 0 };

/** The page's marks as the browser's highlights, which the stylesheet paints. */
export class HighlightMarks implements Marks {
  constructor(private readonly registry: HighlightRegistry) {}

  mark(current: Range | undefined, others: readonly Range[]): void {
    if (current === undefined) {
      this.registry.delete(CURRENT_MATCH);
    } else {
      this.registry.set(CURRENT_MATCH, new Highlight(current));
    }
    this.registry.set(OTHER_MATCHES, new Highlight(...others));
  }

  clear(): void {
    this.registry.delete(CURRENT_MATCH);
    this.registry.delete(OTHER_MATCHES);
  }
}

/**
 * The page as the browser lays it out; a match brought into sight stands as far clear of the foot line as the
 * stylesheet keeps a focused part clear of the screen's foot.
 */
export class PageLayout implements Layout {
  constructor(private readonly view: Window) {}

  shows(text: Node): boolean {
    return text.parentElement?.checkVisibility({ visibilityProperty: true }) ?? false;
  }

  topOf(match: Range): number {
    return match.getBoundingClientRect().top + this.view.scrollY;
  }

  bringIntoSight(match: Range): void {
    const holder = match.startContainer.parentElement;
    if (holder === null) {
      return;
    }
    const clear = parseFloat(this.view.getComputedStyle(holder).scrollMarginBottom);
    const footTop = this.view.document.querySelector(FOOT)?.getBoundingClientRect().top ?? this.view.innerHeight;
    const box = match.getBoundingClientRect();
    if (box.bottom > footTop - clear) {
      this.view.scrollBy(0, box.bottom - footTop + clear);
    } else if (box.top < clear) {
      this.view.scrollBy(0, box.top - clear);
    }
  }

  scrolled(): number {
    return this.view.scrollY;
  }

  scrollTo(point: number): void {
    this.view.scrollTo(0, point);
  }
}

/** The matches of a search in a page's passages, in their order. */
function matchesOf(passages: readonly Passage<Text>[], query: string, page: Document): Range[] {
  return passages.flatMap((passage) => matchesIn(passage, query)).map((match) => rangeOf(match, page));
}

/** A match as a range of the page, from its first letter's text to its last's. */
function rangeOf(match: Match<Text>, page: Document): Range {
  const range = page.createRange();
  range.setStart(match.start.piece, match.start.offset);
  range.setEnd(match.end.piece, match.end.offset);
  return range;
}

/** The texts a part of the page shows, in the page's order, in groups of those one block sets one after another. */
function textsByBlock(part: HTMLElement, layout: Layout): Text[][] {
  const view = part.ownerDocument.defaultView;
  // a page with no window lays nothing out, so it shows no words
  if (view === null) {
    return [];
  }
  const walker = part.ownerDocument.createTreeWalker(part, NodeFilter.SHOW_TEXT);
  const groups: Text[][] = [];
  let groupBlock: Element | null = null;
  for (let text = walker.nextNode(); text !== null; text = walker.nextNode()) {
    if (!(text instanceof Text) || !layout.shows(text)) {
      continue;
    }
    const block = blockOf(text, view);
    const group = groups.at(-1);
    if (group !== undefined && block === groupBlock) {
      group.push(text);
    } else {
      groups.push([text]);
    }
    groupBlock = block;
  }
  return groups;
}

// How an element is laid out when it sets its words among those around it, as a mark or a link: any other display,
// a block, a flex item, a cell, an inline-block, sets its words apart
const AMONG_WORDS = "inline";

/** The block that sets a text's words: the nearest element holding it that is not set among words. */
function blockOf(text: Text, view: Window): Element | null {
  let holder = text.parentElement;
  while (holder !== null && view.getComputedStyle(holder).display === AMONG_WORDS) {
    holder = holder.parentElement;
  }
  return holder;
}

/** A page's search, as the TUI's: what is typed, the search kept, their matches, and the match shown. */
export class Search {
  // where the page stood when the slash was typed: what is typed seeks from it, and leaving goes back to it
  private begunAt = 0;
  // the passages the page showed then, which every letter typed seeks in
  private passages: Passage<Text>[] = [];
  private typing = NOTHING;
  private kept = NOTHING;
  // the match the kept search shows, and where the page stood then: n goes on from it while the page stays there
  private shown: number | undefined = undefined;
  private shownAt = 0;

  constructor(
    private readonly main: HTMLElement,
    private readonly marks: Marks,
    private readonly layout: Layout,
    private readonly storage: () => ChoiceStore,
  ) {}

  /** The search the session keeps marks the page, none of its matches named, as the TUI marks a tab; or none does. */
  resume(): void {
    const query = storedChoice(this.storage, KEPT_SEARCH);
    this.shown = undefined;
    this.kept =
      query === null
        ? NOTHING
        : { query, matches: matchesOf(this.passagesShown(), query, this.main.ownerDocument), place: 0 };
    this.mark(this.kept.matches, undefined);
  }

  /** The page marks no search and n has none to go round, while the session keeps its own for the pages after. */
  hide(): void {
    this.kept = NOTHING;
    this.shown = undefined;
    this.marks.clear();
  }

  /** The slash is typed: what is typed seeks from where the page stands, among the words it shows. */
  begin(): void {
    this.begunAt = this.layout.scrolled();
    this.passages = this.passagesShown();
  }

  /** What is typed changed: its matches marked and the first at or after the start shown; the count, or none. */
  typed(query: string): string {
    const matches = matchesOf(this.passages, query, this.main.ownerDocument);
    this.typing = { query, matches, place: placeAtOrAfter(this.topsOf(matches), this.begunAt) };
    if (!this.show(this.typing.matches, this.typing.place)) {
      this.marks.clear();
      this.layout.scrollTo(this.begunAt);
      return "";
    }
    return matchCount(this.typing.place, matches.length);
  }

  /** Enter keeps what is typed, the match it shows: what the line says of it. */
  keep(): Said {
    this.kept = this.typing;
    keepChoice(this.storage, KEPT_SEARCH, this.kept.query);
    return this.shownSaid(this.kept.place);
  }

  /** The reader leaves what is typed unkept: the kept search marked again, the page back where it stood. */
  leave(): void {
    this.mark(this.kept.matches, this.shown);
    this.layout.scrollTo(this.begunAt);
  }

  /** Whether a search is kept, for n and N to go round. */
  holds(): boolean {
    return this.kept.query !== "";
  }

  /** n or N: the match a step from the one shown while the page stays there, else the one the way picks. */
  next(way: MatchWay): Said {
    const point = this.layout.scrolled();
    const place =
      this.shown !== undefined && point === this.shownAt
        ? steppedPlace(this.shown, way.step, this.kept.matches.length)
        : way.byThePage(this.topsOf(this.kept.matches), point);
    return this.shownSaid(place);
  }

  /** Esc on the page: the search put away, its marks with it, on the pages after too. */
  end(): void {
    this.hide();
    forgetChoice(this.storage, KEPT_SEARCH);
  }

  /** Shows the kept search's match at a place: which of how many, or that the search found none. */
  private shownSaid(place: number): Said {
    if (!this.show(this.kept.matches, place)) {
      this.shown = undefined;
      this.marks.clear();
      return { pieces: notFoundSaid(this.kept.query), failed: true };
    }
    this.shown = place;
    this.shownAt = this.layout.scrolled();
    return { pieces: foundSaid(this.kept.query, place, this.kept.matches.length), failed: false };
  }

  /** Marks matches and brings the one at a place into sight; false where none stands there. */
  private show(matches: readonly Range[], place: number): boolean {
    const current = matches[place];
    if (current === undefined) {
      return false;
    }
    this.mark(matches, place);
    this.layout.bringIntoSight(current);
    return true;
  }

  /** Marks matches, the one at a place named apart; none named where no place is given. */
  private mark(matches: readonly Range[], place: number | undefined): void {
    const current = place === undefined ? undefined : matches[place];
    this.marks.mark(
      current,
      matches.filter((match) => match !== current),
    );
  }

  /** The words the page shows now, a passage for each block that sets them. */
  private passagesShown(): Passage<Text>[] {
    return textsByBlock(this.main, this.layout).map((texts) => passageOf(texts));
  }

  private topsOf(matches: readonly Range[]): number[] {
    return matches.map((match) => this.layout.topOf(match));
  }
}
