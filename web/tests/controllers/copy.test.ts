// The command chips and a post's blocks of code, run over the pages the server draws: the HTML face's golden pages
// (internal/face/html/testdata), which that face's own tests hold to what it writes. A fake clipboard stands for the
// browser's, and the clock is the test's own, so the toast's time passes when a test says.
import { afterEach, beforeEach, expect, test, vi, type Mock } from "vitest";
import { startCopying, type TextClipboard } from "../../src/controllers/copy";
import resumePage from "../../../internal/face/html/testdata/index.html?raw";
import projectPage from "../../../internal/face/html/testdata/p/tidy-tool.html?raw";
import notFoundPage from "../../../internal/face/html/testdata/notfound.html?raw";
import postPage from "../../../internal/face/html/testdata/blog/reading-slowly.html?raw";

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
// The fixture's project's install line, a chip of its own on the project's page, and what its toast says
const INSTALL = "go install example.test/tidy@latest";
const INSTALL_TOAST_SAYS = "Copied go install example.test/tidy@latest";
// A hint's window, which a way of the foot line opens: the first is ssh's, holding the command the first screen does
const HINT = '[data-part="hint"]';
// The fixture's post's block of code, the code it copies and its button, and what the toast says once it has
const CODE = '[data-part="code"]';
const CODE_WORDS = '[data-part="code-words"]';
const CODE_COPY = `${CODE} ${COPY}`;
const CODE_TEXT = 'fmt.Println("slowly")';
const CODE_TOAST_SAYS = "Copied the code";

/** The browser's clipboard as the chip uses it: what it was handed, and whether it refuses. */
class FakeClipboard implements TextClipboard {
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

/** Adds a golden page's main part after the document's own, so the document holds that page's chips too. */
function withMainOf(page: string): void {
  const main = new DOMParser().parseFromString(page, "text/html").querySelector("main");
  if (main === null) {
    throw new Error("the golden page has no main part");
  }
  found("main").after(main);
}

/** The chips of the golden page, run over a fake clipboard, which the test reads. */
function started(): FakeClipboard {
  const clipboard = new FakeClipboard();
  startCopying(document, clipboard, TOAST_TIME);
  return clipboard;
}

/** What a reader sees of the chip's button and the toast: its word, and the toast's words, or none where it is hidden. */
function shown(): { button: string; toast: string | null } {
  return { button: found(COPY).textContent, toast: found(TOAST).hidden ? null : found(TOAST_WORDS).textContent };
}

/**
 * Lends every element the hidePopover jsdom lacks, as a mock: the windows a script hid are its calls' contexts, in
 * order.
 */
function lentHidePopover(): Mock<() => void> {
  const hidePopover = vi.fn<() => void>();
  HTMLElement.prototype.hidePopover = hidePopover;
  return hidePopover;
}

/** Presses the chip, as a reader clicks it, and lets the clipboard answer. */
async function press(): Promise<void> {
  found(CHIP).click();
  await vi.advanceTimersByTimeAsync(0);
}

/** Makes the document a golden page, as the browser would show it. */
function showPage(page: string): void {
  document.documentElement.innerHTML = new DOMParser().parseFromString(page, "text/html").documentElement.innerHTML;
}

beforeEach(() => {
  vi.useFakeTimers();
  showPage(resumePage);
});

afterEach(() => {
  document.getSelection()?.removeAllRanges();
  Reflect.deleteProperty(HTMLElement.prototype, "hidePopover");
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
  withMainOf(projectPage);
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
  document.getSelection()?.selectAllChildren(found(".foot"));
  await press();
  expect(clipboard.written).toEqual([COMMAND]);
});

test("a press that leaves the caret in the chip, selecting nothing, copies", async () => {
  const clipboard = started();
  document.getSelection()?.collapse(found(COMMAND_WORDS), 0);
  await press();
  expect(clipboard.written).toEqual([COMMAND]);
});

test.each([
  { name: "the résumé", page: resumePage },
  { name: "the page of no page", page: notFoundPage },
])("on $name a copy from a hint's window closes the window, and the toast says what was copied", async ({ page }) => {
  showPage(page);
  const hidePopover = lentHidePopover();
  started();
  found(`${HINT} ${CHIP}`).click();
  await vi.advanceTimersByTimeAsync(0);
  expect({ hidden: hidePopover.mock.contexts, toast: found(TOAST_WORDS).textContent }).toEqual({
    hidden: [found(HINT)],
    toast: TOAST_SAYS,
  });
});

test("a copy from a chip on the page closes no window", async () => {
  const hidePopover = lentHidePopover();
  started();
  await press();
  expect(hidePopover.mock.contexts).toEqual([]);
});

test("a copy refused in a hint's window leaves it open, the command selected to copy by hand", async () => {
  const hidePopover = lentHidePopover();
  const clipboard = started();
  clipboard.refuses = true;
  found(`${HINT} ${CHIP}`).click();
  await vi.advanceTimersByTimeAsync(0);
  expect({ hidden: hidePopover.mock.contexts, selected: document.getSelection()?.toString() }).toEqual({
    hidden: [],
    selected: COMMAND,
  });
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

test("a block of code copies its code from its button, which says so, and the toast says the code was copied", async () => {
  showPage(postPage);
  const clipboard = started();
  found(CODE_COPY).click();
  await vi.advanceTimersByTimeAsync(0);
  expect({
    written: clipboard.written,
    button: found(CODE_COPY).textContent,
    toast: found(TOAST_WORDS).textContent,
  }).toEqual({
    written: [CODE_TEXT],
    button: COPIED_WORD,
    toast: CODE_TOAST_SAYS,
  });
});

test("a press on a block's code itself copies nothing, so a reader selects its lines", async () => {
  showPage(postPage);
  const clipboard = started();
  found(CODE_WORDS).click();
  await vi.advanceTimersByTimeAsync(0);
  expect(clipboard.written).toEqual([]);
});

test("a block's button copies the whole code while the reader has selected some of it", async () => {
  showPage(postPage);
  const clipboard = started();
  document.getSelection()?.selectAllChildren(found(CODE_WORDS));
  found(CODE_COPY).click();
  await vi.advanceTimersByTimeAsync(0);
  expect(clipboard.written).toEqual([CODE_TEXT]);
});
