// The switch that turns the vim keys on, run over the pages the server draws: the HTML face's golden pages
// (internal/face/html/testdata), which that face's own tests hold to what it writes.
import { beforeEach, expect, test } from "vitest";
import type { ChoiceStore } from "../../src/controllers/choice";
import { startVimSwitch } from "../../src/controllers/vimswitch";
import resumePage from "../../../internal/face/html/testdata/index.html?raw";

// The switch as the server draws it: the label, hidden until the script shows it, and its box
const SWITCH = '[data-part="vim-keys"]';
const BOX = `${SWITCH} input`;

/** A browser that keeps nothing: its storage refuses to be reached, as it does where a reader blocks it. */
function refusedStorage(): ChoiceStore {
  throw new DOMException("storage is blocked", "SecurityError");
}

/** A page as a browser parses what the server sent, its switch run over a storage. */
function started(storage: () => ChoiceStore): Document {
  const page = new DOMParser().parseFromString(resumePage, "text/html");
  startVimSwitch(page, storage);
  return page;
}

/** What a reader sees of the switch: whether it shows, and whether it is ticked. */
function seen(page: Document): { shown: boolean; ticked: boolean } {
  const label = page.querySelector<HTMLElement>(SWITCH);
  const box = page.querySelector<HTMLInputElement>(BOX);
  if (label === null || box === null) {
    throw new Error("the page has no switch");
  }
  return { shown: !label.hidden, ticked: box.checked };
}

/** Ticks or unticks the switch, as a reader's press does. */
function press(page: Document): void {
  page.querySelector<HTMLInputElement>(BOX)?.click();
}

beforeEach(() => localStorage.clear());

test("the switch shows unticked until a reader ticks it", () => {
  expect(seen(started(() => localStorage))).toEqual({ shown: true, ticked: false });
});

test("a tick holds on every next page, and so does taking it back", () => {
  press(started(() => localStorage));
  const ticked = seen(started(() => localStorage));
  press(started(() => localStorage));
  expect({ ticked, after: seen(started(() => localStorage)) }).toEqual({
    ticked: { shown: true, ticked: true },
    after: { shown: true, ticked: false },
  });
});

test("a browser that keeps nothing still ticks the switch, for the page alone", () => {
  const page = started(refusedStorage);
  press(page);
  expect(seen(page)).toEqual({ shown: true, ticked: true });
});
