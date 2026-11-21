// The switch that turns the vim keys on (DESIGN.md, VimKeys), over the markup internal/face/html draws: a checkbox the
// server draws hidden, which this script shows, ticked where the reader ticked it on an earlier page. The keys
// themselves (src/controllers/commandline.ts) read the box when a key comes. It runs before the page is first seen, as
// the theme does, so the switch shows with the foot line and moves none of its parts after.
import { forgetChoice, keepChoice, storedChoice, type ChoiceStore } from "./choice";

// The switch as the server draws it: the label it stands in, and its box
const SWITCH = '[data-part="vim-keys"]';
const BOX = "input";
// Where the browser keeps the reader's tick, and what it keeps
const STORAGE_KEY = "vim-keys";
const TICKED = "on";

/** Shows a page's switch, ticked where the reader ticked it before, and keeps each tick for the pages after. */
export function startVimSwitch(page: Document, storage: () => ChoiceStore): void {
  const label = page.querySelector<HTMLElement>(SWITCH);
  const box = label?.querySelector<HTMLInputElement>(BOX);
  if (label === null || box === null || box === undefined) {
    return;
  }
  box.checked = storedChoice(storage, STORAGE_KEY) === TICKED;
  box.addEventListener("change", () => keepTick(storage, box.checked));
  label.hidden = false;
}

/** Keeps a tick for the pages after, or forgets it. */
function keepTick(storage: () => ChoiceStore, ticked: boolean): void {
  if (ticked) {
    keepChoice(storage, STORAGE_KEY, TICKED);
  } else {
    forgetChoice(storage, STORAGE_KEY);
  }
}
