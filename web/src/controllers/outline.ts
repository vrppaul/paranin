// The outline (DESIGN.md, Outline), over the markup internal/face/html draws: the headings of a post's top level,
// each a link beside the article. As the article scrolls, the heading read, the last whose top has passed a line
// across the screen's upper part, is marked as the reader's place, which the stylesheet draws bold, a bar before it;
// a page read without the script marks none. The heading the address names, a link followed to it, is read while it
// stands above the line, though the next has passed it too; and a page scrolled to its end, where a short last
// section never brings its heading up to the line, reads that heading while it is in sight, or else its last. Every
// heading is measured when one crosses the band above the line, which an observer says as the page
// scrolls, when a scroll ends, since a jump past the whole band crosses nothing the observer watches, and when the
// address names another heading.

// The outline's links and the headings they lead to, as the server draws them; what a link writes before the name
// of its heading; and the attribute that marks the heading read, with its value
const OUTLINE_LINK = '[data-part="outline-link"]';
const OUTLINE_HEADING = '[data-part="outline-heading"]';
const FRAGMENT_MARK = "#";
const CURRENT = "aria-current";
const CURRENT_PLACE = "location";
// The line a heading counts as read once its top has passed, in percent of the screen's height from its top: the
// foot of its upper third, so a heading a link scrolls to the top is read at once, and one under the words in view is
// not yet; and the screen's whole height, in percent
const READING_LINE_PERCENT = 33;
const WHOLE_PERCENT = 100;
// The band the observer watches, from the screen's top down to the reading line; the events a scroll ends with and
// the address names another place with
const READING_BAND = `0px 0px ${String(READING_LINE_PERCENT - WHOLE_PERCENT)}% 0px`;
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

/** Marks, as the page scrolls in its window, the outline's link to the heading read. */
export function startOutline(page: Document, view: ScrolledView, makeObserver: ObserverMaker): void {
  const places = placesOf(page);
  const observer = makeObserver(() => markRead(places, page, view), { rootMargin: READING_BAND });
  for (const place of places) {
    observer.observe(place.heading);
  }
  view.addEventListener(SCROLL_END, () => markRead(places, page, view));
  view.addEventListener(HASH_CHANGE, () => markRead(places, page, view));
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
 * Marks the link to the heading read, and no other: the one the address names while it stands between the screen's
 * top and the reading line, or anywhere on the screen once the page is scrolled to its end, so a link followed to a
 * short section marks its heading, not the next one the line has passed too; else at the page's end its last, and
 * anywhere else the last the line has passed, none where it has passed none.
 */
function markRead(places: OutlinePlace[], page: Document, view: ScrolledView): void {
  const atEnd = view.scrollY + view.innerHeight >= page.documentElement.scrollHeight - END_SLACK;
  const reach = atEnd ? view.innerHeight : readingLine(view);
  const named = places.find((place) => place.link.hash === view.location.hash && standsAbove(place.heading, reach));
  const read = named ?? (atEnd ? places.at(-1) : lastPassed(places, view));
  for (const place of places) {
    if (place === read) {
      place.link.setAttribute(CURRENT, CURRENT_PLACE);
    } else {
      place.link.removeAttribute(CURRENT);
    }
  }
}

/** The last place whose heading's top stands above the screen's reading line; none where the line has passed none. */
function lastPassed(places: OutlinePlace[], view: ScrolledView): OutlinePlace | undefined {
  const line = readingLine(view);
  return places.filter((place) => place.heading.getBoundingClientRect().top < line).at(-1);
}

/** How far down the screen its reading line runs, in pixels. */
function readingLine(view: ScrolledView): number {
  return (view.innerHeight * READING_LINE_PERCENT) / WHOLE_PERCENT;
}

/** Whether a heading's top stands on the screen, above a height down it. */
function standsAbove(heading: HTMLElement, reach: number): boolean {
  const top = heading.getBoundingClientRect().top;
  return top >= 0 && top < reach;
}
