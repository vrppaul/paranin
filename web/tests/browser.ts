// What a browser shows a page by that jsdom lacks, or that a test turns as a reader would: a question put to the
// reader's system, a canvas's pen, and the box an element is laid out at.
import type { SystemScheme } from "../src/controllers/theme";

/** A question put to the reader's system, as their dark scheme or their ask for still motion, which a test turns. */
export class FakeSystem extends EventTarget implements SystemScheme {
  matches: boolean;

  constructor(matches: boolean) {
    super();
    this.matches = matches;
  }

  turn(matches: boolean): void {
    this.matches = matches;
    this.dispatchEvent(new Event("change"));
  }
}

/** A text a pen drew, and the colour it drew it in. */
export interface DrawnText {
  text: string;
  colour: string;
}

/**
 * A canvas's pen that keeps every text drawn with it, in order, and the scale it last drew at; whatever else it is
 * asked to draw, it draws nowhere. jsdom has no canvas to draw on.
 */
export class RecordingPen {
  fillStyle = "";
  scale = 1;
  readonly drawnTexts: DrawnText[] = [];

  setTransform(horizontalScale: number): void {
    this.scale = horizontalScale;
  }

  fillRect(): void {
    // A ground drawn is nothing a test reads
  }

  fillText(text: string): void {
    this.drawnTexts.push({ text, colour: this.fillStyle });
  }

  /** The pen as a canvas hands it out. */
  asContext(): CanvasRenderingContext2D {
    return this as unknown as CanvasRenderingContext2D;
  }
}

/** An element's box, in CSS pixels. */
export interface Box {
  width: number;
  height: number;
}

/** That a test laid an element out, which jsdom lays out none of, at a box of its own. */
class LaidOut extends Event {
  static readonly type = "laid-out";
  readonly box: Box;

  constructor(box: Box) {
    super(LaidOut.type);
    this.box = box;
  }
}

/** Lays an element out at a box, as a browser lays it out: every observer of its size hears it. */
export function layOut(element: Element, box: Box): void {
  element.dispatchEvent(new LaidOut(box));
}

/** An observer of an element's size, which hears it each time a test lays it out, and never before. */
export class LayoutObserver {
  readonly #callback: ResizeObserverCallback;

  constructor(callback: ResizeObserverCallback) {
    this.#callback = callback;
  }

  observe(target: Element): void {
    target.addEventListener(LaidOut.type, this);
  }

  unobserve(target: Element): void {
    target.removeEventListener(LaidOut.type, this);
  }

  handleEvent(laidOut: LaidOut): void {
    const entry = { target: laidOut.target, contentRect: laidOut.box } as unknown as ResizeObserverEntry;
    this.#callback([entry], this as unknown as ResizeObserver);
  }
}
