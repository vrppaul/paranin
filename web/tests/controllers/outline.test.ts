// The outline, run over the post the server draws: the HTML face's golden page (internal/face/html/testdata), which
// that face's own tests hold to what it writes. jsdom lays nothing out, so a fake observer stands for the browser's,
// calling back when a test says the page scrolled, and the page, how far it is scrolled and each heading stand where a
// test puts them.
import { beforeEach, expect, test, vi } from "vitest";
import { startOutline, type HeadingObserver } from "../../src/controllers/outline";
import postPage from "../../../internal/face/html/testdata/blog/reading-slowly.html?raw";

// The outline's links and the headings they lead to, as the server draws them; the attribute that marks the heading
// read; the address of the first heading, as a followed link names it, its name after the mark of a place
const OUTLINE_LINK = '[data-part="outline-link"]';
const OUTLINE_HEADING = '[data-part="outline-heading"]';
const CURRENT = "aria-current";
const FRAGMENT_MARK = "#";
const FIRST_HEADING = FRAGMENT_MARK + "what-a-reader-keeps";

// A heading's top above the reading line, off the screen and on it, under it, and under it though in sight near the
// page's end, in pixels from the top of jsdom's window, 768 tall, whose line runs a third of the way down
const ABOVE_THE_LINE = -100;
const NEAR_THE_TOP = 100;
const UNDER_THE_LINE = 600;
const NEAR_THE_END = 300;
const NEARER_THE_END = 500;
// How tall the page is, in pixels, and how far it is scrolled at its top
const PAGE_HEIGHT = 5000;
const TOP = 0;
// The events a scroll ends with and the address names another place with
const SCROLL_END = "scrollend";
const HASH_CHANGE = "hashchange";

/** The browser's observer as the outline uses it: what it was told to watch, and the call it makes as they cross. */
class FakeObserver implements HeadingObserver {
  watched: Element[] = [];

  constructor(private readonly callback: IntersectionObserverCallback) {}

  observe = (target: Element): void => {
    this.watched.push(target);
  };

  /** Calls back as the browser does once it watches a heading, and once the page scrolls one across the band. */
  scrolled(): void {
    this.callback([], this as unknown as IntersectionObserver);
  }
}

/** Makes the outline's observer, a fake, and hands it to the test. */
class FakeObserverMaker {
  private observer: FakeObserver | undefined;

  make = (callback: IntersectionObserverCallback): FakeObserver => {
    this.observer = new FakeObserver(callback);
    return this.observer;
  };

  /** The observer the outline made. */
  made(): FakeObserver {
    if (this.observer === undefined) {
      throw new Error("the outline made no observer");
    }
    return this.observer;
  }
}

/** Puts each heading of the page where its top stands on the screen, in their order; one not named, under the line. */
function placeHeadings(tops: number[]): void {
  const headings = document.querySelectorAll<HTMLElement>(OUTLINE_HEADING);
  for (const [index, heading] of [...headings].entries()) {
    vi.spyOn(heading, "getBoundingClientRect").mockReturnValue({ top: tops[index] ?? UNDER_THE_LINE } as DOMRect);
  }
}

/** Scrolls the page as far as a test says, from its top. */
function scrollPageTo(scrolled: number): void {
  Object.defineProperty(window, "scrollY", { value: scrolled, configurable: true });
}

/** Scrolls the page to its end. */
function scrollPageToItsEnd(): void {
  scrollPageTo(PAGE_HEIGHT - window.innerHeight);
}

/** Whether each link of the outline is marked as the heading read, in their order. */
function marked(): boolean[] {
  return [...document.querySelectorAll(OUTLINE_LINK)].map((link) => link.hasAttribute(CURRENT));
}

/** The outline of the golden page, started over a fake observer, which the test scrolls. */
function started(): FakeObserver {
  const maker = new FakeObserverMaker();
  startOutline(document, window, maker.make);
  return maker.made();
}

beforeEach(() => {
  const parsed = new DOMParser().parseFromString(postPage, "text/html");
  document.documentElement.innerHTML = parsed.documentElement.innerHTML;
  Object.defineProperty(document.documentElement, "scrollHeight", { value: PAGE_HEIGHT, configurable: true });
  scrollPageTo(TOP);
  history.replaceState(null, "", location.pathname);
});

test("the outline watches every heading it leads to", () => {
  const observer = started();
  expect(observer.watched).toEqual([...document.querySelectorAll(OUTLINE_HEADING)]);
});

test("the heading the reading line has passed is marked read", () => {
  const observer = started();
  placeHeadings([ABOVE_THE_LINE, UNDER_THE_LINE]);
  observer.scrolled();
  expect(marked()).toEqual([true, false]);
});

test("of the headings the line has passed, the last is the one read", () => {
  const observer = started();
  placeHeadings([ABOVE_THE_LINE, ABOVE_THE_LINE]);
  observer.scrolled();
  expect(marked()).toEqual([false, true]);
});

test("a heading scrolled back under the line is read no more", () => {
  const observer = started();
  placeHeadings([ABOVE_THE_LINE, ABOVE_THE_LINE]);
  observer.scrolled();
  placeHeadings([ABOVE_THE_LINE, UNDER_THE_LINE]);
  observer.scrolled();
  expect(marked()).toEqual([true, false]);
});

test("a jump past the whole band, which the observer never sees, marks the heading read once the scroll ends", () => {
  started();
  placeHeadings([ABOVE_THE_LINE, UNDER_THE_LINE]);
  window.dispatchEvent(new Event(SCROLL_END));
  expect(marked()).toEqual([true, false]);
});

test("no heading is read while the line has passed none", () => {
  const observer = started();
  placeHeadings([UNDER_THE_LINE, UNDER_THE_LINE]);
  observer.scrolled();
  expect(marked()).toEqual([false, false]);
});

test("a page scrolled to its end reads its last heading, which a short last section never brings up to the line", () => {
  const observer = started();
  scrollPageToItsEnd();
  placeHeadings([ABOVE_THE_LINE, NEARER_THE_END]);
  observer.scrolled();
  expect(marked()).toEqual([false, true]);
});

test("a heading a followed link names is marked at once, where the page's end keeps it under the line", () => {
  started();
  scrollPageToItsEnd();
  placeHeadings([NEAR_THE_END, NEARER_THE_END]);
  history.replaceState(null, "", FIRST_HEADING);
  window.dispatchEvent(new HashChangeEvent(HASH_CHANGE));
  expect(marked()).toEqual([true, false]);
});

test("a heading a followed link names is read, though the next heading has passed the line too", () => {
  started();
  placeHeadings([TOP, NEAR_THE_TOP]);
  history.replaceState(null, "", FIRST_HEADING);
  window.dispatchEvent(new HashChangeEvent(HASH_CHANGE));
  expect(marked()).toEqual([true, false]);
});

test("a heading a followed link names counts no more once it has left the screen", () => {
  started();
  placeHeadings([ABOVE_THE_LINE, NEAR_THE_TOP]);
  history.replaceState(null, "", FIRST_HEADING);
  window.dispatchEvent(new HashChangeEvent(HASH_CHANGE));
  expect(marked()).toEqual([false, true]);
});

test("a page opened at a heading marks it", () => {
  history.replaceState(null, "", FIRST_HEADING);
  scrollPageToItsEnd();
  placeHeadings([NEAR_THE_END, NEARER_THE_END]);
  started().scrolled();
  expect(marked()).toEqual([true, false]);
});
