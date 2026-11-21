// The page the command line's tests run it over: the HTML face's golden page of the résumé's experience tab
// (internal/face/html/testdata), which that face's own tests hold to what it writes. The addresses the line goes to
// are handed to the test, in place of the browser; and as jsdom lays nothing out, the page's layout and marks are the
// test's own: a match stands as far down as the letters of the page's main part before it.
import { startCommandLine } from "../../src/controllers/commandline";
import type { Layout, Marks } from "../../src/controllers/search";
import experiencePage from "../../../internal/face/html/testdata/experience.html?raw";

// The switch that turns the vim keys on, its box, and the line's field
const SWITCH = '[data-part="vim-keys"]';
const BOX = `${SWITCH} input`;
const FIELD = '[data-part="command-field"]';

// A part of the page the browser shows nothing of
const HIDDEN = "[hidden]";

/** The page as the line lays it out: the words outside a hidden part shown, where each match and the page stand. */
class FakeLayout implements Layout {
  standsAt = 0;

  constructor(private readonly main: HTMLElement) {}

  shows(text: Node): boolean {
    return !text.parentElement?.closest(HIDDEN);
  }

  topOf(match: Range): number {
    const before = this.main.ownerDocument.createRange();
    before.setStart(this.main, 0);
    before.setEnd(match.startContainer, match.startOffset);
    return before.toString().length;
  }

  bringIntoSight(match: Range): void {
    this.standsAt = this.topOf(match);
  }

  scrolled(): number {
    return this.standsAt;
  }

  scrollTo(point: number): void {
    this.standsAt = point;
  }
}

/** The page's marks as a reader sees them: where the match the count names stands and its words, and the others. */
class FakeMarks implements Marks {
  current: number | undefined = undefined;
  currentWords = "";
  others: number[] = [];

  constructor(private readonly layout: FakeLayout) {}

  mark(current: Range, others: readonly Range[]): void {
    this.current = this.layout.topOf(current);
    this.currentWords = current.toString();
    this.others = others.map((match) => this.layout.topOf(match));
  }

  clear(): void {
    this.current = undefined;
    this.currentWords = "";
    this.others = [];
  }
}

/** The line run over the golden page: the addresses it went to, and the page's layout and marks. */
export interface StartedLine {
  went: string[];
  layout: FakeLayout;
  marks: FakeMarks;
}

/** The element a selector finds in the document, or in a page a test changes, which the golden page draws. */
export function found(selector: string, page: ParentNode = document): HTMLElement {
  const element = page.querySelector<HTMLElement>(selector);
  if (element === null) {
    throw new Error(`the page has no ${selector}`);
  }
  return element;
}

/** An input of the golden page a selector finds: the switch's box or the line's field. */
export function input(selector: string): HTMLInputElement {
  const element = document.querySelector<HTMLInputElement>(selector);
  if (element === null) {
    throw new Error(`the page has no ${selector}`);
  }
  return element;
}

/** The golden page with its command line run, the vim keys ticked or not, once a change of the test's is made to it. */
export function started(ticked = true, changed: (page: Document) => void = () => undefined): StartedLine {
  const parsed = new DOMParser().parseFromString(experiencePage, "text/html");
  changed(parsed);
  document.body.replaceWith(document.adoptNode(parsed.body));
  const went: string[] = [];
  const layout = new FakeLayout(found("main"));
  const marks = new FakeMarks(layout);
  startCommandLine(document, (address) => went.push(address), marks, layout);
  // the switch's own script shows it on the page
  found(SWITCH).hidden = false;
  input(BOX).checked = ticked;
  return { went, layout, marks };
}

/** A key pressed where the focus is, or on the page where nothing holds it; the event, for what became of it. */
export function press(key: string, held: KeyboardEventInit = {}): KeyboardEvent {
  const target = document.activeElement ?? document.body;
  const event = new KeyboardEvent("keydown", { key, bubbles: true, cancelable: true, ...held });
  target.dispatchEvent(event);
  return event;
}

/** Letters typed on the line, after what is there. */
export function type(letters: string): void {
  const field = input(FIELD);
  field.value += letters;
  field.dispatchEvent(new Event("input", { bubbles: true }));
}
