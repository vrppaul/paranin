// The top bar's current section (DESIGN.md, TopBar), over the markup internal/face/html draws: the bar's links to the
// résumé's sections, and the sections they lead to. As the reader scrolls, the section they are reading is marked on
// its link as the current place, which the stylesheet reverses, as the TUI's tab (src/ui/ui.css). A section is read
// from the moment its top reaches the line a followed link stops it at, its scroll margin under the bar, which the
// stylesheet sets; it stays read until the next one's top reaches that line. At the page's end, where a section may
// never reach the line, the section the address names is current while its top stands on the screen, or else the
// last. A page without the sections, every page but the résumé, marks none.

// The bar's links and the sections as the server draws them, and how a link names the place it leads to
const LINKS = '[data-part="sections"] a';
const SECTIONS = '[data-part="section"]';
const PLACE_MARK = "#";
// What says a link leads to the current place on the page, as ARIA writes it
const CURRENT = "aria-current";
const CURRENT_PLACE = "location";
// How far past the line a section's top may stand and still have reached it, for a browser that stops a followed
// link a fraction of a pixel short
const LINE_TOLERANCE = 1;

/** The window as the bar reads it while the reader scrolls. */
export interface Reading {
  /** How tall the screen is. */
  screenHeight(): number;
  /** How far below the screen's top a section is current from: where a followed link stops it. */
  line(): number;
  /** Whether the page has scrolled as far down as it goes. */
  atEnd(): boolean;
  /** The place the page's address names, as "#skills", or nothing. */
  place(): string;
  /** Calls back, once a frame at the most, whenever the reader scrolls or follows a link to a place. */
  onMove(listener: () => void): void;
}

/** A section of the résumé and the bar's link to it. */
interface LinkedSection {
  section: Element;
  link: HTMLAnchorElement;
}

/** Marks the link of the section the reader is in, as the page opens and whenever they move through it. */
export function startSections(page: Document, reading: Reading): void {
  const links = [...page.querySelectorAll<HTMLAnchorElement>(LINKS)];
  const linked = [...page.querySelectorAll(SECTIONS)].flatMap((section) => linkedSection(section, links));
  if (linked.length === 0) {
    return;
  }
  reading.onMove(() => markCurrent(linked, reading));
  markCurrent(linked, reading);
}

/**
 * The browser's window as the bar reads it. The line is the sections' scroll margin, which the stylesheet sets, read
 * once and again only after the window is resized, as a phone's width may change it; the reader's moves reach the
 * listeners once a frame. The window is listened to from the first listener on, so a page without the sections asks
 * it for nothing.
 */
export class WindowReading implements Reading {
  private readonly view: Window;
  private readonly listeners: (() => void)[] = [];
  private knownLine: number | undefined = undefined;
  private frameAsked = false;

  constructor(view: Window) {
    this.view = view;
  }

  screenHeight(): number {
    return this.view.innerHeight;
  }

  line(): number {
    this.knownLine ??= this.readLine();
    return this.knownLine;
  }

  atEnd(): boolean {
    const page = this.view.document.documentElement;
    return this.view.scrollY + this.view.innerHeight >= page.scrollHeight - LINE_TOLERANCE;
  }

  place(): string {
    return this.view.location.hash;
  }

  onMove(listener: () => void): void {
    if (this.listeners.length === 0) {
      this.listenToWindow();
    }
    this.listeners.push(listener);
  }

  /** Asks for a frame whenever the reader scrolls or follows a link to a place, and forgets the line on a resize. */
  private listenToWindow(): void {
    this.view.addEventListener("scroll", () => this.askFrame(), { passive: true });
    this.view.addEventListener("hashchange", () => this.askFrame());
    this.view.addEventListener("resize", () => this.forgetLine());
  }

  /** Asks the browser to tell the listeners before its next frame, unless it is asked already. */
  private askFrame(): void {
    if (!this.frameAsked) {
      this.frameAsked = true;
      this.view.requestAnimationFrame(() => this.tell());
    }
  }

  private tell(): void {
    this.frameAsked = false;
    for (const listener of this.listeners) {
      listener();
    }
  }

  /** Forgets the line, and marks again by the one the resized window holds. */
  private forgetLine(): void {
    this.knownLine = undefined;
    this.askFrame();
  }

  /** The scroll margin the stylesheet gives the sections; the screen's top where the page has none. */
  private readLine(): number {
    const section = this.view.document.querySelector(SECTIONS);
    return section === null ? 0 : Number.parseFloat(this.view.getComputedStyle(section).scrollMarginTop);
  }
}

/** A section with the bar's link to it, the link whose place is the section's id; none where no link leads to it. */
function linkedSection(section: Element, links: HTMLAnchorElement[]): LinkedSection[] {
  const link = links.find((candidate) => candidate.hash === PLACE_MARK + section.id);
  return link === undefined ? [] : [{ section, link }];
}

/** Marks the current section's link, and no other. */
function markCurrent(linked: LinkedSection[], reading: Reading): void {
  const current = currentSection(linked, reading);
  for (const { link } of linked) {
    if (link === current?.link) {
      link.setAttribute(CURRENT, CURRENT_PLACE);
    } else {
      link.removeAttribute(CURRENT);
    }
  }
}

/**
 * The section the reader is in: at the page's end, the one its address names while its top stands on the screen, or
 * else the last; elsewhere the last whose top has reached the line; none above the first.
 */
function currentSection(linked: LinkedSection[], reading: Reading): LinkedSection | undefined {
  if (reading.atEnd()) {
    const named = linked.find(({ link }) => link.hash === reading.place());
    return named !== undefined && topOnScreen(named.section, reading) ? named : linked.at(-1);
  }
  const line = reading.line() + LINE_TOLERANCE;
  return linked.findLast(({ section }) => section.getBoundingClientRect().top <= line);
}

/** Whether a section's top stands on the screen. */
function topOnScreen(section: Element, reading: Reading): boolean {
  const top = section.getBoundingClientRect().top;
  return top >= 0 && top < reading.screenHeight();
}
