// The outline (DESIGN.md, Outline), over the markup internal/face/html draws: the headings of a post's top level,
// each a link beside the article. As the article scrolls, the heading read, as DESIGN.md's Outline says which it is,
// by the reading line its entry hands it (design/tokens.yaml), is marked as the reader's place, which the stylesheet
// draws bold, a bar before it; a page read without the script marks none. The place gone to is the one the address
// names, a link followed to it. Every heading is measured when one crosses the band above the line, which an observer
// says as the page scrolls, when a scroll ends, since a jump past the whole band crosses nothing the observer watches,
// and when the address names another place.

// The outline's links and the headings they lead to, as the server draws them; what a link writes before the name
// of its heading; the place the address names, as the browser marks it; and the attribute that marks the heading
// read, with its value
const OUTLINE_LINK = '[data-part="outline-link"]';
const OUTLINE_HEADING = '[data-part="outline-heading"]';
const FRAGMENT_MARK = "#";
const TARGET = ":target";
const CURRENT = "aria-current";
const CURRENT_PLACE = "location";
// The screen's whole height, in percent, which the reading line is a share of
const WHOLE_PERCENT = 100;
// The events a scroll ends with and the address names another place with
const SCROLL_END = "scrollend";
const HASH_CHANGE = "hashchange";
// How far short of its end, in pixels, a page counts as scrolled to it: a zoomed page's scroll stops a fraction short
const END_SLACK = 1;

/** The browser's observer of the headings as the outline uses it: told to watch a heading. */
export type HeadingObserver = Pick<IntersectionObserver, "observe">;

/** Makes an observer that calls back as a watched heading crosses the band its options draw, as the browser's does. */
export type ObserverMaker = (
  callback: IntersectionObserverCallback,
  options: IntersectionObserverInit,
) => HeadingObserver;

/**
 * The window the page scrolls in, as the outline reads it: how tall it is, how far it is scrolled, the place its
 * address names, and the end of each scroll and each change of that place.
 */
export type ScrolledView = Pick<Window, "innerHeight" | "scrollY" | "location" | "addEventListener">;

/** A place of the outline: its link, and the heading the link leads to. */
interface OutlinePlace {
  link: HTMLAnchorElement;
  heading: HTMLElement;
}

/**
 * Marks, as the page scrolls in its window, the outline's link to the heading read, by a reading line a share of the
 * screen's height down it, in percent; the observer watches the band from the screen's top down to that line.
 */
export function startOutline(
  page: Document,
  view: ScrolledView,
  makeObserver: ObserverMaker,
  readingLinePercent: number,
): void {
  const places = placesOf(page);
  const readingBand = `0px 0px ${String(readingLinePercent - WHOLE_PERCENT)}% 0px`;
  const observer = makeObserver(() => markRead(places, page, view, readingLinePercent), { rootMargin: readingBand });
  for (const place of places) {
    observer.observe(place.heading);
  }
  view.addEventListener(SCROLL_END, () => markRead(places, page, view, readingLinePercent));
  view.addEventListener(HASH_CHANGE, () => markRead(places, page, view, readingLinePercent));
}

/** The places of a page's outline: each link with the heading it leads to, a link to no heading of the page left out. */
function placesOf(page: Document): OutlinePlace[] {
  const headings = [...page.querySelectorAll<HTMLElement>(OUTLINE_HEADING)];
  const places: OutlinePlace[] = [];
  for (const link of page.querySelectorAll<HTMLAnchorElement>(OUTLINE_LINK)) {
    const heading = headings.find((candidate) => FRAGMENT_MARK + candidate.id === link.getAttribute("href"));
    if (heading !== undefined) {
      places.push({ link, heading });
    }
  }
  return places;
}

/**
 * Marks the link to the heading read (DESIGN.md, Outline), by a reading line a share of the screen's height down it,
 * and no other; the place gone to is the one the address names, a heading or a paragraph under it.
 */
function markRead(places: OutlinePlace[], page: Document, view: ScrolledView, readingLinePercent: number): void {
  const atEnd = view.scrollY + view.innerHeight >= page.documentElement.scrollHeight - END_SLACK;
  const line = readingLine(view, readingLinePercent);
  const reach = atEnd ? view.innerHeight : line;
  const target = namedPlace(page);
  const named =
    target !== null && standsAbove(target, reach) ? places.find((place) => holds(place, target)) : undefined;
  const read = named ?? (atEnd ? places.at(-1) : lastPassed(places, line));
  for (const place of places) {
    if (place === read) {
      place.link.setAttribute(CURRENT, CURRENT_PLACE);
    } else {
      place.link.removeAttribute(CURRENT);
    }
  }
}

/** The place of the page its address names, as the browser marks it; none where it names none. */
function namedPlace(page: Document): HTMLElement | null {
  return page.querySelector<HTMLElement>(TARGET);
}

/** Whether a place of the outline holds an element: its heading's section, the heading's parent, holds it. */
function holds(place: OutlinePlace, element: HTMLElement): boolean {
  return place.heading.parentElement?.contains(element) ?? false;
}

/** The last place whose heading's top stands above a line down the screen; none where the line has passed none. */
function lastPassed(places: OutlinePlace[], line: number): OutlinePlace | undefined {
  return places.filter((place) => place.heading.getBoundingClientRect().top < line).at(-1);
}

/** How far down the screen its reading line runs, in pixels, a share of its height in percent. */
function readingLine(view: ScrolledView, readingLinePercent: number): number {
  return (view.innerHeight * readingLinePercent) / WHOLE_PERCENT;
}

/** Whether a heading's top stands on the screen, above a height down it. */
function standsAbove(heading: HTMLElement, reach: number): boolean {
  const top = heading.getBoundingClientRect().top;
  return top >= 0 && top < reach;
}
