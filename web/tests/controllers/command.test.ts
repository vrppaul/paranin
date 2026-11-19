// The command chips, run over the résumé the server draws: the HTML face's golden page (internal/face/html/testdata),
// which that face's own tests hold to what it writes. A fake clipboard stands for the browser's, and the clock is
// the test's own, so the toast's time passes when a test says.
import { afterEach, beforeEach, expect, test, vi } from "vitest";
import { startCommands, type CommandClipboard } from "../../src/controllers/command";
import resumePage from "../../../internal/face/html/testdata/index.html?raw";

// The chip, the words it copies and its button as the server draws them; the toast and its words
const CHIP = '[data-part="command"]';
const COMMAND_WORDS = '[data-part="command-words"]';
const COPY = '[data-part="copy"]';
const TOAST = '[data-part="toast"]';
const TOAST_WORDS = '[data-part="toast-words"]';

// How long a toast stays in these tests, in milliseconds, and half of it, when a second press comes; the command the
// fixture's chip copies, with what its button and toast say, in the fixture's words
const TOAST_TIME = 2400;
const HALF_TOAST_TIME = 1200;
const COMMAND = "ssh example.test";
const COPY_WORD = "copy";
const COPIED_WORD = "copied";
const TOAST_SAYS = "Copied ssh example.test";
// The fixture's project's install line, a chip of its own further down the page, and what its toast says
const INSTALL = "go install example.test/tidy@latest";
const INSTALL_TOAST_SAYS = "Copied go install example.test/tidy@latest";

/** The browser's clipboard as the chip uses it: what it was handed, and whether it refuses. */
class FakeClipboard implements CommandClipboard {
  written: string[] = [];
  refuses = false;

  writeText = (text: string): Promise<void> => {
    if (this.refuses) {
      return Promise.reject(new Error("the clipboard refused"));
    }
    this.written.push(text);
    return Promise.resolve();
  };
}

/** The document's element a selector finds, which the golden page draws. */
function found(selector: string): HTMLElement {
  const element = document.querySelector<HTMLElement>(selector);
  if (element === null) {
    throw new Error(`the page has no ${selector}`);
  }
  return element;
}

/** The chips of the golden page, run over a fake clipboard, which the test reads. */
function started(): FakeClipboard {
  const clipboard = new FakeClipboard();
  startCommands(document, clipboard, TOAST_TIME);
  return clipboard;
}

/** What a reader sees of the chip's button and the toast: its word, and the toast's words, or none where it is hidden. */
function shown(): { button: string; toast: string | null } {
  return { button: found(COPY).textContent, toast: found(TOAST).hidden ? null : found(TOAST_WORDS).textContent };
}

/** Presses the chip, as a reader clicks it, and lets the clipboard answer. */
async function press(): Promise<void> {
  found(CHIP).click();
  await vi.advanceTimersByTimeAsync(0);
}

beforeEach(() => {
  vi.useFakeTimers();
  const parsed = new DOMParser().parseFromString(resumePage, "text/html");
  document.documentElement.innerHTML = parsed.documentElement.innerHTML;
});

afterEach(() => {
  document.getSelection()?.removeAllRanges();
  vi.useRealTimers();
});

test("the button the server draws hidden shows once the script runs", () => {
  started();
  expect(found(COPY).hidden).toBe(false);
});

test("a press copies the command, the button says so, and the toast says what was copied", async () => {
  const clipboard = started();
  await press();
  expect({ written: clipboard.written, ...shown() }).toEqual({
    written: [COMMAND],
    button: COPIED_WORD,
    toast: TOAST_SAYS,
  });
});

test("each chip copies its own command, and the toast says which", async () => {
  const clipboard = started();
  found(`${COPY}[aria-label="copy ${INSTALL}"]`).click();
  await vi.advanceTimersByTimeAsync(0);
  expect({ written: clipboard.written, toast: found(TOAST_WORDS).textContent }).toEqual({
    written: [INSTALL],
    toast: INSTALL_TOAST_SAYS,
  });
});

test("after the toast's time the button and the toast turn back", async () => {
  started();
  await press();
  await vi.advanceTimersByTimeAsync(TOAST_TIME);
  expect(shown()).toEqual({ button: COPY_WORD, toast: null });
});

test("a second press keeps the button's word and the toast their whole time from that press", async () => {
  started();
  await press();
  await vi.advanceTimersByTimeAsync(HALF_TOAST_TIME);
  await press();
  await vi.advanceTimersByTimeAsync(HALF_TOAST_TIME);
  expect(shown()).toEqual({ button: COPIED_WORD, toast: TOAST_SAYS });
});

test("a press on a chip whose words the reader selected copies nothing, so the selection stays", async () => {
  const clipboard = started();
  document.getSelection()?.selectAllChildren(found(COMMAND_WORDS));
  await press();
  expect({ written: clipboard.written, ...shown() }).toEqual({ written: [], button: COPY_WORD, toast: null });
});

test("a press on the chip copies while words elsewhere on the page are selected", async () => {
  const clipboard = started();
  document.getSelection()?.selectAllChildren(found("footer"));
  await press();
  expect(clipboard.written).toEqual([COMMAND]);
});

test("a press that leaves the caret in the chip, selecting nothing, copies", async () => {
  const clipboard = started();
  document.getSelection()?.collapse(found(COMMAND_WORDS), 0);
  await press();
  expect(clipboard.written).toEqual([COMMAND]);
});

test("a copy the browser refuses selects the command for the reader to copy, and says nothing was copied", async () => {
  const clipboard = started();
  clipboard.refuses = true;
  await press();
  expect({ selected: document.getSelection()?.toString(), ...shown() }).toEqual({
    selected: COMMAND,
    button: COPY_WORD,
    toast: null,
  });
});
