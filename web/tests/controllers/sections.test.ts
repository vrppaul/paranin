// The top bar's current section, run over the pages the server draws: the HTML face's golden pages
// (internal/face/html/testdata), which that face's own tests hold to what it writes. jsdom lays nothing out, so a test
// says where each section's top stands on the screen, and a fake reading stands for the window: the screen's height,
// the line a section is current from, here the screen's top, whether the page has scrolled to its end, the place its
// address names, and the reader's scroll.
import { expect, test } from "vitest";
import { startSections, type Reading } from "../../src/controllers/sections";
import resumePage from "../../../internal/face/html/testdata/index.html?raw";
import missingPage from "../../../internal/face/html/testdata/notfound.html?raw";

// The bar's links, its current one, and the résumé's sections as the server draws them
const LINKS = '[data-part="sections"] a';
const CURRENT_LINK = `${LINKS}[aria-current]`;
const SECTIONS = '[data-part="section"]';

// The height of the screen the fake reading shows the page on
const SCREEN_HEIGHT = 900;

/** The window as the controller reads it, which a test scrolls as a reader would. */
class FakeReading implements Reading {
  end = false;
  named = "";
  listeners: (() => void)[] = [];

  screenHeight = (): number => SCREEN_HEIGHT;
  line = (): number => 0;
  atEnd = (): boolean => this.end;
  place = (): string => this.named;
  onMove = (listener: () => void): void => {
    this.listeners.push(listener);
  };

  /** Scrolls the page: each section's top, by its id, where the screen now shows it, and whether it is the end. */
  scroll(page: Document, tops: Record<string, number>, atEnd = false): void {
    for (const section of page.querySelectorAll(SECTIONS)) {
      const top = tops[section.id] ?? 0;
      section.getBoundingClientRect = () => ({ top }) as DOMRect;
    }
    this.end = atEnd;
    for (const listener of this.listeners) {
      listener();
    }
  }
}

/** A page as a browser parses what the server sent. */
function parsed(markup: string): Document {
  return new DOMParser().parseFromString(markup, "text/html");
}

/** The bar's link of the given words, which the page's address names once the reader follows it. */
function linkNamed(page: Document, words: string): HTMLAnchorElement {
  const links = [...page.querySelectorAll<HTMLAnchorElement>(LINKS)];
  const link = links.find((candidate) => candidate.textContent === words);
  if (link === undefined) {
    throw new Error(`the bar has no link ${words}`);
  }
  return link;
}

/** The words of the bar's links that say they are the current section. */
function current(page: Document): string[] {
  return [...page.querySelectorAll(CURRENT_LINK)].map((link) => link.textContent);
}

// Where the résumé's sections stand once the reader has scrolled the experience section's top to the screen's top
const AT_EXPERIENCE = { about: -500, experience: 0, projects: 900, skills: 1800, contact: 2700 };

test("the section whose top a followed link stopped at the line is the current one, on its link alone", () => {
  const page = parsed(resumePage);
  const reading = new FakeReading();
  startSections(page, reading);
  reading.scroll(page, AT_EXPERIENCE);
  expect(current(page)).toEqual(["Experience"]);
});

test("a section stays current until the next one's top reaches the line", () => {
  const page = parsed(resumePage);
  const reading = new FakeReading();
  startSections(page, reading);
  reading.scroll(page, { ...AT_EXPERIENCE, about: -1400, experience: -900, projects: 2 });
  const nextBelow = current(page);
  reading.scroll(page, { ...AT_EXPERIENCE, about: -1400, experience: -900, projects: 0 });
  expect({ nextBelow, nextAtTheLine: current(page) }).toEqual({
    nextBelow: ["Experience"],
    nextAtTheLine: ["Projects"],
  });
});

test("above the first section no section is current", () => {
  const page = parsed(resumePage);
  const reading = new FakeReading();
  startSections(page, reading);
  reading.scroll(page, AT_EXPERIENCE);
  reading.scroll(page, { about: 400, experience: 900, projects: 1800, skills: 2700, contact: 3600 });
  expect(current(page)).toEqual([]);
});

test("at the page's end the section the address names is current, though the page ended before it reached the line", () => {
  const page = parsed(resumePage);
  const reading = new FakeReading();
  startSections(page, reading);
  reading.named = linkNamed(page, "Skills").hash;
  reading.scroll(page, { about: -2200, experience: -1700, projects: -800, skills: 300, contact: 900 }, true);
  expect(current(page)).toEqual(["Skills"]);
});

test("at the page's end a named section the reader scrolled off the screen is not current, the last is", () => {
  const page = parsed(resumePage);
  const reading = new FakeReading();
  startSections(page, reading);
  reading.named = linkNamed(page, "About").hash;
  reading.scroll(page, { about: -2700, experience: -2200, projects: -1300, skills: -400, contact: 500 }, true);
  expect(current(page)).toEqual(["Contact"]);
});

test("at the page's end with no section named, the last section is current", () => {
  const page = parsed(resumePage);
  const reading = new FakeReading();
  startSections(page, reading);
  reading.scroll(page, { about: -2200, experience: -1700, projects: -800, skills: 300, contact: 900 }, true);
  expect(current(page)).toEqual(["Contact"]);
});

test("the current section is marked as the page opens, before the reader scrolls", () => {
  const page = parsed(resumePage);
  const reading = new FakeReading();
  reading.scroll(page, AT_EXPERIENCE);
  startSections(page, reading);
  expect(current(page)).toEqual(["Experience"]);
});

test("a page without the résumé's sections listens to nothing and marks nothing", () => {
  const page = parsed(missingPage);
  const reading = new FakeReading();
  startSections(page, reading);
  expect({ listening: reading.listeners.length, current: current(page) }).toEqual({ listening: 0, current: [] });
});
