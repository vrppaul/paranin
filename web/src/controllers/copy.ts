// The command chips and a post's blocks of code (DESIGN.md, Command, Article; Tooltip and Toast), over the markup
// internal/face/html draws: text the reader copies whole, words they may select, and a button the server draws hidden,
// which this script shows. A press anywhere on a command's chip copies the command, unless the reader selected words in
// it, which they mean to copy themselves; a block of code, whose lines a reader selects, copies on its button alone.
// The button then says it copied, and the toast at the foot of the screen says what, both for the toast's time from
// the last press. A copy from a hint's window (DESIGN.md, Hint) closes the window, its work done, and the toast says it
// where every copy is said. A copy the browser refuses selects the text instead, for the reader to copy by hand, its
// window left open: an error is never a toast, and the selection shows what to copy.

// The parts as the server draws them: a command's chip and the words it copies, a block of code and the code it
// copies, and the button of either, which carries the word it turns to and what the toast says; the toast, the place
// its words go and the template each word is drawn from; the hint's window a chip may stand in
const CHIP = '[data-part="command"]';
const COMMAND_WORDS = '[data-part="command-words"]';
const CODE = '[data-part="code"]';
const CODE_WORDS = '[data-part="code-words"]';
const COPY = '[data-part="copy"]';
const TOAST = '[data-part="toast"]';
const TOAST_WORDS = '[data-part="toast-words"]';
const TOAST_WORD = '[data-part="toast-word"]';
const HINT = '[data-part="hint"]';
// What parts the toast's words, and the one place its lines break: never at a command's hyphens
const SPACE = " ";
// The attribute that says a button has just copied, which the stylesheet colours, and its value
const STATE = "data-state";
const COPIED_STATE = "copied";

/** The browser's clipboard, of which a part writes text alone. */
export type TextClipboard = Pick<Clipboard, "writeText">;

/**
 * A kind of part a page copies from: where it stands, the words it copies, and whether a press anywhere on it copies,
 * as on a command's chip, or on its button alone, as on a block of code whose lines a reader selects.
 */
interface CopiedKind {
  part: string;
  words: string;
  pressedWhole: boolean;
}

// Every kind of part a page copies from
const COPIED_KINDS: readonly CopiedKind[] = [
  { part: CHIP, words: COMMAND_WORDS, pressedWhole: true },
  { part: CODE, words: CODE_WORDS, pressedWhole: false },
];

/** A word of the toast's, drawn from the element of its template. */
function drawnWord(written: string, wordElement: Element): Node {
  const drawn = wordElement.cloneNode();
  drawn.textContent = written;
  return drawn;
}

/** What the toast says, each word drawn whole and a space as text between two, where alone a line may break. */
function drawnWords(said: string, wordElement: Element): (Node | string)[] {
  const [first = "", ...rest] = said.split(SPACE);
  return [drawnWord(first, wordElement), ...rest.flatMap((written) => [SPACE, drawnWord(written, wordElement)])];
}

/** The toast a page says what was done in, for a time from when it last began to say it. */
class Toast {
  private hiding: ReturnType<typeof setTimeout> | undefined;

  constructor(
    private readonly toast: HTMLElement,
    private readonly words: HTMLElement,
    private readonly wordElement: Element,
    private readonly time: number,
  ) {}

  show(said: string): void {
    clearTimeout(this.hiding);
    this.words.replaceChildren(...drawnWords(said, this.wordElement));
    this.toast.hidden = false;
    this.hiding = setTimeout(() => this.hide(), this.time);
  }

  private hide(): void {
    this.toast.hidden = true;
    this.words.textContent = "";
  }
}

/** The words of a part's button, before a copy and after it, and what the toast says it copied. */
interface ButtonWords {
  copy: string;
  copied: string;
  toastSays: string;
}

/** A part a page copies from, which copies its text on a press and says so for the toast's time from the last press. */
class CopiedPart {
  private turningBack: ReturnType<typeof setTimeout> | undefined;

  constructor(
    private readonly part: HTMLElement,
    private readonly text: HTMLElement,
    private readonly button: HTMLButtonElement,
    private readonly pressedWhole: boolean,
    private readonly words: ButtonWords,
    private readonly clipboard: TextClipboard,
    private readonly toast: Toast,
    private readonly toastTime: number,
  ) {}

  /** Shows the button, and lets a press on it copy, or anywhere on the part where the whole part is pressed. */
  start(): void {
    const pressed = this.pressedWhole ? this.part : this.button;
    pressed.addEventListener("click", () => void this.copy());
    this.button.hidden = false;
  }

  /**
   * Copies the text, unless the reader selected words in a part pressed whole: the button says it copied, the window
   * it stands in closes, and the toast says what. Where the browser refuses, the text is selected instead.
   */
  private async copy(): Promise<void> {
    const selection = this.part.ownerDocument.getSelection();
    if (this.pressedWhole && selection !== null && !selection.isCollapsed && this.part.contains(selection.anchorNode)) {
      return;
    }
    try {
      await this.clipboard.writeText(this.text.textContent);
    } catch {
      selection?.selectAllChildren(this.text);
      return;
    }
    clearTimeout(this.turningBack);
    this.button.textContent = this.words.copied;
    this.button.setAttribute(STATE, COPIED_STATE);
    this.part.closest<HTMLElement>(HINT)?.hidePopover();
    this.toast.show(this.words.toastSays);
    this.turningBack = setTimeout(() => this.turnBack(), this.toastTime);
  }

  private turnBack(): void {
    this.button.textContent = this.words.copy;
    this.button.removeAttribute(STATE);
  }
}

/**
 * Runs every part of a kind on a page that holds a toast, its text copied over the clipboard; a part drawn without one
 * of its parts keeps its text as words to select.
 */
function startKind(page: Document, kind: CopiedKind, clipboard: TextClipboard, toast: Toast, toastTime: number): void {
  for (const part of page.querySelectorAll<HTMLElement>(kind.part)) {
    const text = part.querySelector<HTMLElement>(kind.words);
    const button = part.querySelector<HTMLButtonElement>(COPY);
    const copied = button?.dataset.copied;
    const toastSays = button?.dataset.toast;
    if (text === null || button === null || copied === undefined || toastSays === undefined) {
      continue;
    }
    const words = { copy: button.textContent, copied, toastSays };
    new CopiedPart(part, text, button, kind.pressedWhole, words, clipboard, toast, toastTime).start();
  }
}

/**
 * Runs every command chip and block of code of a page that holds a toast, the toast staying for its time, in
 * milliseconds. A page with no toast keeps them all as words to select.
 */
export function startCopying(page: Document, clipboard: TextClipboard, toastTime: number): void {
  const toastPart = page.querySelector<HTMLElement>(TOAST);
  const toastWords = toastPart?.querySelector<HTMLElement>(TOAST_WORDS);
  const wordElement = toastPart?.querySelector<HTMLTemplateElement>(TOAST_WORD)?.content.firstElementChild;
  if (toastPart == null || toastWords == null || wordElement == null) {
    return;
  }
  const toast = new Toast(toastPart, toastWords, wordElement, toastTime);
  for (const kind of COPIED_KINDS) {
    startKind(page, kind, clipboard, toast, toastTime);
  }
}
