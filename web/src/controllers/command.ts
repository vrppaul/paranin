// The command chips (DESIGN.md, Command; Tooltip and Toast), over the markup internal/face/html draws: a command the
// reader can run, its words code they may select, and a button the server draws hidden, which this script shows. A
// press anywhere on the chip copies the command, unless the reader selected words in it, which they mean to copy
// themselves; the button then says it copied, and the toast at the foot of the screen says what, both for the
// toast's time from the last press. A copy from a hint's window (DESIGN.md, Hint) closes the window, its work done,
// and the toast says it where every copy is said. A copy the browser refuses selects the command instead, for the
// reader to copy by hand, its window left open: an error is never a toast, and the selection shows what to copy.

// The chip as the server draws it: the chip, the words it copies, and its button, which carries the word it turns to
// and what the toast says; the toast and the place its words go; the hint's window a chip may stand in
const CHIP = '[data-part="command"]';
const COMMAND_WORDS = '[data-part="command-words"]';
const COPY = '[data-part="copy"]';
const TOAST = '[data-part="toast"]';
const TOAST_WORDS = '[data-part="toast-words"]';
const HINT = '[data-part="hint"]';
// The attribute that says a button has just copied, which the stylesheet colours, and its value
const STATE = "data-state";
const COPIED_STATE = "copied";

/** The browser's clipboard, of which a chip writes text alone. */
export type CommandClipboard = Pick<Clipboard, "writeText">;

/** The toast a page says what was done in, for a time from when it last began to say it. */
class Toast {
  private hiding: ReturnType<typeof setTimeout> | undefined;

  constructor(
    private readonly toast: HTMLElement,
    private readonly words: HTMLElement,
    private readonly time: number,
  ) {}

  show(said: string): void {
    clearTimeout(this.hiding);
    this.words.textContent = said;
    this.toast.hidden = false;
    this.hiding = setTimeout(() => this.hide(), this.time);
  }

  private hide(): void {
    this.toast.hidden = true;
    this.words.textContent = "";
  }
}

/** The words of a chip's button, before a copy and after it, and what the toast says it copied. */
interface ChipWords {
  copy: string;
  copied: string;
  toastSays: string;
}

/** A command's chip, which copies its command on a press and says so for the toast's time from the last press. */
class CommandChip {
  private turningBack: ReturnType<typeof setTimeout> | undefined;

  constructor(
    private readonly chip: HTMLElement,
    private readonly command: HTMLElement,
    private readonly button: HTMLButtonElement,
    private readonly words: ChipWords,
    private readonly clipboard: CommandClipboard,
    private readonly toast: Toast,
    private readonly toastTime: number,
  ) {}

  /** Shows the button, and lets a press anywhere on the chip copy. */
  start(): void {
    this.chip.addEventListener("click", () => void this.copy());
    this.button.hidden = false;
  }

  /**
   * Copies the command, unless the reader selected words in the chip: the button says it copied, the window it stands
   * in closes, and the toast says what. Where the browser refuses, the command is selected instead.
   */
  private async copy(): Promise<void> {
    const selection = this.chip.ownerDocument.getSelection();
    if (selection !== null && !selection.isCollapsed && this.chip.contains(selection.anchorNode)) {
      return;
    }
    try {
      await this.clipboard.writeText(this.command.textContent);
    } catch {
      selection?.selectAllChildren(this.command);
      return;
    }
    clearTimeout(this.turningBack);
    this.button.textContent = this.words.copied;
    this.button.setAttribute(STATE, COPIED_STATE);
    this.chip.closest<HTMLElement>(HINT)?.hidePopover();
    this.toast.show(this.words.toastSays);
    this.turningBack = setTimeout(() => this.turnBack(), this.toastTime);
  }

  private turnBack(): void {
    this.button.textContent = this.words.copy;
    this.button.removeAttribute(STATE);
  }
}

/**
 * Runs every command chip of a page that holds a toast, the toast staying for its time, in milliseconds. A page with
 * no toast, or a chip drawn without one of its parts, keeps its commands as words to select.
 */
export function startCommands(page: Document, clipboard: CommandClipboard, toastTime: number): void {
  const toastPart = page.querySelector<HTMLElement>(TOAST);
  const toastWords = toastPart?.querySelector<HTMLElement>(TOAST_WORDS);
  if (toastPart == null || toastWords == null) {
    return;
  }
  const toast = new Toast(toastPart, toastWords, toastTime);
  for (const chip of page.querySelectorAll<HTMLElement>(CHIP)) {
    const command = chip.querySelector<HTMLElement>(COMMAND_WORDS);
    const button = chip.querySelector<HTMLButtonElement>(COPY);
    const copied = button?.dataset.copied;
    const toastSays = button?.dataset.toast;
    if (command === null || button === null || copied === undefined || toastSays === undefined) {
      continue;
    }
    const words = { copy: button.textContent, copied, toastSays };
    new CommandChip(chip, command, button, words, clipboard, toast, toastTime).start();
  }
}
