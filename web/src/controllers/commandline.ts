// The command line the vim keys open (DESIGN.md, CommandLine), over the markup internal/face/html draws: with the
// switch ticked and in sight, a colon typed on the page opens the line over the foot line for a place or a command, a
// slash for a search (./search.ts); its field takes the keys. A colon's menu shows the places that fit what is typed,
// and Tab and Shift+Tab go round them, as the TUI's do (src/domain/commandline.ts). Enter goes where the line says,
// keeps the search, or says why it will not, until the next key; n and N go round the search kept, which marks the
// pages opened after it too, Esc on the page puts it away once nothing stands over the page (./overlay.ts); Esc,
// Backspace on an empty line and the focus leaving close the line.
import {
  completionsOf,
  lineOf,
  placesOf,
  ran,
  stepped,
  type Completion,
  type Completions,
  type Folder,
  type Place,
  type Places,
} from "../domain/commandline";
import type { ChoiceStore } from "./choice";
import { typedIn, type Overlay } from "./overlay";
import { BACK, ONWARD, Search, type Layout, type Marks, type MatchWay, type Said } from "./search";

// The parts as the server draws them: the switch and its box; the line, the part a reader types in, its field and the
// menu of its places, each place with the letters typed of its name and the rest; the colon's mark and the slash's;
// what the line says; the keys it names for a place, and the one that goes round the places; the keys it names for a
// search, and the count of its matches; and the page's main part, which a search finds words in
const SWITCH = '[data-part="vim-keys"]';
const BOX = "input";
const LINE = '[data-part="command-line"]';
const ENTRY = '[data-part="command-entry"]';
const FIELD = '[data-part="command-field"]';
const PLACES = '[data-part="places"]';
const PLACE = '[data-part="place"]';
const TYPED = '[data-part="place-typed"]';
const REST = '[data-part="place-rest"]';
const COLON_MARK = '[data-part="command-mark"]';
const SLASH_MARK = '[data-part="search-mark"]';
const SAID = '[data-part="command-said"]';
const KEYS = '[data-part="command-keys"]';
const NEXT = '[data-part="command-next"]';
const SEARCH_KEYS = '[data-part="search-keys"]';
const COUNT = '[data-part="search-count"]';
const MAIN = "main";

// How the stylesheet takes a part out of the page, as it takes the switch on a phone
const GONE = "none";

// What the server writes of a place being the page shown, and of one :e opens
const CURRENT = "current";
const EDITED = "edit";

// The keys the line answers, as a browser names them
const COLON = ":";
const SLASH = "/";
const NEXT_MATCH = "n";
const PREVIOUS_MATCH = "N";
const TAB = "Tab";
const ENTER = "Enter";
const ESCAPE = "Escape";
const BACKSPACE = "Backspace";

// Tab's step round the places, and Shift+Tab's
const FORWARD = 1;
const BACKWARD = -1;

// What the field says to a screen reader: what it is, its name, the keys it names, whether it offers a menu, whether
// the menu shows, and the place chosen; a search is a field of its own kind, with no menu
const ROLE = "role";
const SEARCH_ROLE = "searchbox";
const NAME = "aria-label";
const DESCRIBED_BY = "aria-describedby";
const AUTOCOMPLETE = "aria-autocomplete";
const CONTROLS = "aria-controls";
const EXPANDED = "aria-expanded";
const ACTIVE = "aria-activedescendant";
const SELECTED = "aria-selected";

// The state of what the line says that the stylesheet colours as no refusal: a search's match shown
const STATE = "data-state";
const FOUND_STATE = "found";

// The cells of the line before the name its menu lists, which the stylesheet stands the menu's names under
const TYPED_CELLS = "--typed-cells";

/** Where the line goes: an address on the site, handed to the browser. */
type Visit = (address: string) => void;

/** What the line is open for: a place or a command after the colon, a search after the slash. */
type Mode = "command" | "search";

/** How the field presents itself to a screen reader in a mode: what it is, its name, and its menu if any. */
interface FieldKind {
  role: string;
  name: string;
  menu: FieldMenu | undefined;
}

/** The menu a field offers, as a screen reader is told of it: how it completes, and which element it is. */
interface FieldMenu {
  autocomplete: string;
  controls: string;
}

/**
 * The parts of the line the server draws, how its field presents itself for a place and for a search, and the folder
 * its field names, where :e opens a page the menu does not list.
 */
interface Parts {
  switchLabel: HTMLElement;
  box: HTMLInputElement;
  line: HTMLElement;
  entry: HTMLElement;
  field: HTMLInputElement;
  menu: HTMLElement;
  colonMark: HTMLElement;
  slashMark: HTMLElement;
  said: HTMLElement;
  keys: HTMLElement;
  next: HTMLElement;
  searchKeys: HTMLElement;
  count: HTMLElement;
  kinds: Record<Mode, FieldKind>;
  folder: Folder;
}

/** A place on the menu: where it goes, and its element with the letters typed of its name and the rest. */
interface Item {
  place: Place;
  element: HTMLElement;
  typed: HTMLElement;
  rest: HTMLElement;
}

/**
 * Runs a page's command line, where the page draws one, going where it says through visit, its search's matches
 * marked by marks on the page as layout lays it out, the search kept for the pages after in storage, and Esc left to
 * what stands over the page while overlay says anything does.
 */
export function startCommandLine(
  page: Document,
  visit: Visit,
  marks: Marks,
  layout: Layout,
  storage: () => ChoiceStore,
  overlay: Overlay,
): void {
  const parts = partsOf(page);
  const main = page.querySelector<HTMLElement>(MAIN);
  if (parts === undefined || main === null) {
    return;
  }
  const items = [...parts.menu.querySelectorAll<HTMLElement>(PLACE)].flatMap((element) => itemOf(element) ?? []);
  const line = new CommandLine(page, parts, items, visit, new Search(main, marks, layout, storage), overlay);
  line.listen();
  line.markKept();
}

/** The line's parts, or nothing where the page lacks one. */
function partsOf(page: Document): Parts | undefined {
  const switchLabel = page.querySelector<HTMLElement>(SWITCH);
  const box = switchLabel?.querySelector<HTMLInputElement>(BOX);
  const line = page.querySelector<HTMLElement>(LINE);
  const entry = line?.querySelector<HTMLElement>(ENTRY);
  const field = line?.querySelector<HTMLInputElement>(FIELD);
  const menu = line?.querySelector<HTMLElement>(PLACES);
  const colonMark = line?.querySelector<HTMLElement>(COLON_MARK);
  const slashMark = line?.querySelector<HTMLElement>(SLASH_MARK);
  const said = line?.querySelector<HTMLElement>(SAID);
  const keys = line?.querySelector<HTMLElement>(KEYS);
  const next = line?.querySelector<HTMLElement>(NEXT);
  const searchKeys = line?.querySelector<HTMLElement>(SEARCH_KEYS);
  const count = line?.querySelector<HTMLElement>(COUNT);
  const role = field?.getAttribute(ROLE);
  const name = field?.getAttribute(NAME);
  const searchName = field?.dataset.searchName;
  const folderName = field?.dataset.folder;
  const folderAddress = field?.dataset.folderAddress;
  const autocomplete = field?.getAttribute(AUTOCOMPLETE);
  const controls = field?.getAttribute(CONTROLS);
  if (
    !switchLabel ||
    !box ||
    !line ||
    !entry ||
    !field ||
    !menu ||
    !colonMark ||
    !slashMark ||
    !said ||
    !keys ||
    !next ||
    !searchKeys ||
    !count ||
    !role ||
    !name ||
    !searchName ||
    !folderName ||
    !folderAddress ||
    !autocomplete ||
    !controls
  ) {
    return undefined;
  }
  const kinds: Record<Mode, FieldKind> = {
    command: { role, name, menu: { autocomplete, controls } },
    search: { role: SEARCH_ROLE, name: searchName, menu: undefined },
  };
  return {
    switchLabel,
    box,
    line,
    entry,
    field,
    menu,
    colonMark,
    slashMark,
    said,
    keys,
    next,
    searchKeys,
    count,
    kinds,
    folder: { name: folderName, address: folderAddress },
  };
}

/** A place as the server draws it on the menu, its kind the word it names it by, or nothing where it lacks a part. */
function itemOf(element: HTMLElement): Item | undefined {
  const { name, place: kind, address } = element.dataset;
  const typed = element.querySelector<HTMLElement>(TYPED);
  const rest = element.querySelector<HTMLElement>(REST);
  if (name === undefined || address === undefined || kind === undefined || typed === null || rest === null) {
    return undefined;
  }
  const place = { name, kind, address, current: CURRENT in element.dataset, edited: EDITED in element.dataset };
  return { place, element, typed, rest };
}

/**
 * Scrolls the menu alone the least that brings a place on it into sight. The browser's own scroll into view would
 * scroll the page too.
 */
function keptInSight(menu: HTMLElement, item: HTMLElement): void {
  const itemEnd = item.offsetTop + item.offsetHeight;
  if (item.offsetTop < menu.scrollTop) {
    menu.scrollTop = item.offsetTop;
  } else if (itemEnd > menu.scrollTop + menu.clientHeight) {
    menu.scrollTop = itemEnd - menu.clientHeight;
  }
}

/** The command line of a page: closed, open for a place or a search with the reader typing, or saying something. */
class CommandLine {
  private readonly places: Places;
  private completions: Completions;
  private mode: Mode = "command";
  private saying = false;
  // where the focus was before the line opened, which it goes back to
  private before: Element | null = null;

  constructor(
    private readonly page: Document,
    private readonly parts: Parts,
    private readonly items: readonly Item[],
    private readonly visit: Visit,
    private readonly search: Search,
    private readonly overlay: Overlay,
  ) {
    this.places = placesOf(
      items.map((item) => item.place),
      parts.folder,
    );
    this.completions = completionsOf(this.places, "");
  }

  listen(): void {
    this.page.addEventListener("keydown", (event) => this.pageKey(event));
    this.parts.field.addEventListener("keydown", (event) => this.fieldKey(event));
    this.parts.field.addEventListener("input", () => this.edited());
    this.parts.field.addEventListener("blur", () => this.left());
    // a press on the menu keeps the focus on the field, so the line stays open for the click that follows
    this.parts.menu.addEventListener("pointerdown", (event) => event.preventDefault());
    for (const item of this.items) {
      item.element.addEventListener("click", () => this.pressed(item.place));
    }
    // the marks follow the keys turned on or off, and the session's search as a page opened since left it
    this.parts.box.addEventListener("change", () => this.markKept());
    this.page.defaultView?.addEventListener("pageshow", (event) => this.shownAgain(event));
  }

  /** Marks the page with the search the session keeps, where the keys are on to put it away; with none where not. */
  markKept(): void {
    if (this.keysOn()) {
      this.search.resume();
    } else {
      this.search.hide();
    }
  }

  /**
   * The browser shows the page again from its cache, as it was before a page opened since: what the line said of a
   * match it named is put away with the name, as a key would put it away.
   */
  private shownAgain(event: PageTransitionEvent): void {
    if (!event.persisted) {
      return;
    }
    if (this.saying) {
      this.close();
    }
    this.markKept();
  }

  /** A press on a place goes where the line would with the place written on it and chosen. */
  private pressed(place: Place): void {
    const completion = this.completions.listed.find((listed) => listed.place === place);
    this.went(completion?.written ?? place.name, completion);
  }

  /** A key on the page: it puts away what the line said, and where the keys are on, answers the vim keys. */
  private pageKey(event: KeyboardEvent): void {
    if (this.saying) {
      this.close();
    }
    const modified = event.ctrlKey || event.metaKey || event.altKey;
    if (modified || !this.keysOn() || !this.parts.line.hidden || typedIn(event.target)) {
      return;
    }
    switch (event.key) {
      case COLON:
        this.open("command");
        break;
      case SLASH:
        this.open("search");
        break;
      case NEXT_MATCH:
        if (!this.goRound(ONWARD)) {
          return;
        }
        break;
      case PREVIOUS_MATCH:
        if (!this.goRound(BACK)) {
          return;
        }
        break;
      case ESCAPE:
        // Esc puts away only what stands on top: a tooltip or a window over the page takes it, the search waits
        if (this.overlay.top(event.target).kind === "page") {
          this.search.end();
        }
        return;
      default:
        return;
    }
    event.preventDefault();
  }

  /** Whether the keys are on: the switch ticked, and in sight, so the reader can untick it. */
  private keysOn(): boolean {
    return this.parts.box.checked && getComputedStyle(this.parts.switchLabel).display !== GONE;
  }

  /** A key in the field: Tab, Enter, and Esc or Backspace on an empty line, which leave it. */
  private fieldKey(event: KeyboardEvent): void {
    switch (event.key) {
      case TAB:
        // a search has no places to go round, and keeps the focus, as the TUI's does
        if (this.mode === "command") {
          this.step(event.shiftKey ? BACKWARD : FORWARD);
        }
        break;
      case ENTER:
        this.entered();
        break;
      case ESCAPE:
        this.leave();
        break;
      case BACKSPACE:
        if (this.parts.field.value !== "") {
          return;
        }
        this.leave();
        break;
      default:
        return;
    }
    event.preventDefault();
    // the page's own listener runs after, and must not take the key again
    event.stopPropagation();
  }

  /** Opens the line for a place or a search, its mark, keys and field the mode's. */
  private open(mode: Mode): void {
    this.mode = mode;
    this.before = this.page.activeElement;
    const searching = mode === "search";
    const kind = this.parts.kinds[mode];
    this.parts.colonMark.hidden = searching;
    this.parts.slashMark.hidden = !searching;
    this.parts.keys.hidden = searching;
    this.parts.searchKeys.hidden = !searching;
    this.parts.field.setAttribute(ROLE, kind.role);
    this.parts.field.setAttribute(NAME, kind.name);
    this.parts.field.setAttribute(DESCRIBED_BY, (searching ? this.parts.searchKeys : this.parts.keys).id);
    if (kind.menu === undefined) {
      this.parts.field.removeAttribute(AUTOCOMPLETE);
      this.parts.field.removeAttribute(CONTROLS);
      this.parts.field.removeAttribute(EXPANDED);
      this.parts.field.removeAttribute(ACTIVE);
    } else {
      this.parts.field.setAttribute(AUTOCOMPLETE, kind.menu.autocomplete);
      this.parts.field.setAttribute(CONTROLS, kind.menu.controls);
    }
    this.parts.field.value = "";
    this.parts.line.hidden = false;
    if (searching) {
      this.search.begin();
    }
    this.edited();
    // the line stands in the foot line, always in sight: a focus that scrolled would take the page to its end
    this.parts.field.focus({ preventScroll: true });
  }

  /** Enter: a place or a command is run, a search kept; Enter on an empty search leaves it. */
  private entered(): void {
    if (this.mode === "command") {
      this.went(this.parts.field.value, this.completions.listed[this.completions.chosen]);
    } else if (this.parts.field.value === "") {
      this.leave();
    } else {
      this.answer(this.search.keep());
    }
  }

  /** n or N, where a search is kept: the line takes its row first, so the match is brought into sight over it. */
  private goRound(way: MatchWay): boolean {
    if (!this.search.holds()) {
      return false;
    }
    this.standSaying();
    this.say(this.search.next(way));
    return true;
  }

  /** The reader leaves the line, the focus handed back while the field can still let it go, which leaves it too. */
  private leave(): void {
    this.handBack();
    this.left();
  }

  /** The focus left the field for somewhere else, where it stays: the line closes, unless it is saying something. */
  private left(): void {
    if (this.saying || this.parts.line.hidden) {
      return;
    }
    if (this.mode === "search") {
      this.search.leave();
    }
    this.close();
  }

  private close(): void {
    this.saying = false;
    this.parts.line.hidden = true;
    this.parts.said.hidden = true;
    // an alert reads its words out only as they change: the same refusal said again must be heard again
    this.parts.said.textContent = "";
    this.parts.said.removeAttribute(STATE);
    this.parts.entry.hidden = false;
    if (this.mode === "command") {
      this.parts.field.setAttribute(EXPANDED, String(false));
    }
  }

  /**
   * Hands the focus back where it was before the line opened; where nothing held it, the field lets it go, since a
   * field hidden with the focus would keep the keys from the page.
   */
  private handBack(): void {
    if (this.page.activeElement === this.parts.field) {
      this.parts.field.blur();
    }
    if (this.before instanceof HTMLElement) {
      this.before.focus({ preventScroll: true });
    }
  }

  /** What is typed changed: a place's menu narrows, a search marks the page and counts its matches. */
  private edited(): void {
    if (this.mode === "search") {
      this.parts.count.textContent = this.search.typed(this.parts.field.value);
      this.parts.menu.hidden = true;
      return;
    }
    this.completions = completionsOf(this.places, this.parts.field.value);
    this.show();
  }

  private step(step: number): void {
    this.completions = stepped(this.completions, step);
    this.parts.field.value = lineOf(this.completions);
    this.show();
  }

  /** Goes where a line says, to the place chosen where one is, or says why it will not. */
  private went(line: string, chosen: Completion | undefined): void {
    const outcome = ran(this.places, line, chosen);
    switch (outcome.kind) {
      case "go":
        this.leave();
        this.visit(outcome.address);
        break;
      case "stay":
      case "nothing":
        this.leave();
        break;
      case "failed":
        this.answer({ words: outcome.message, failed: true });
        break;
    }
  }

  /** Answers what was typed: saying, the line stays open as the field lets the focus go, before the field hides. */
  private answer(said: Said): void {
    this.saying = true;
    this.handBack();
    this.say(said);
  }

  /** The line takes the row it says things in, in place of the field and its keys. */
  private standSaying(): void {
    this.saying = true;
    this.parts.line.hidden = false;
    this.parts.said.hidden = false;
    this.parts.entry.hidden = true;
    this.parts.keys.hidden = true;
    this.parts.searchKeys.hidden = true;
  }

  /** Says something on the line until the next key. */
  private say(said: Said): void {
    this.standSaying();
    this.parts.said.textContent = said.words;
    if (said.failed) {
      this.parts.said.removeAttribute(STATE);
    } else {
      this.parts.said.setAttribute(STATE, FOUND_STATE);
    }
  }

  /** Shows on the menu each place that fits what is typed, the letters typed apart, and the one chosen. */
  private show(): void {
    const { listed, matched, chosen } = this.completions;
    const chosenPlace = listed[chosen]?.place;
    for (const item of this.items) {
      item.element.hidden = !listed.some((completion) => completion.place === item.place);
      item.typed.textContent = matched;
      item.rest.textContent = item.place.name.slice(matched.length);
      item.element.setAttribute(SELECTED, String(item.place === chosenPlace));
    }
    const shown = listed.length > 0;
    this.parts.menu.hidden = !shown;
    this.parts.next.hidden = !shown;
    this.parts.field.setAttribute(EXPANDED, String(shown));
    const chosenItem = this.items.find((item) => item.place === chosenPlace);
    if (chosenItem === undefined) {
      this.parts.field.removeAttribute(ACTIVE);
    } else {
      this.parts.field.setAttribute(ACTIVE, chosenItem.element.id);
      keptInSight(this.parts.menu, chosenItem.element);
    }
    // the names stand under where the line writes them: after what comes before the name on the line
    const cellsBefore = this.parts.field.value.length - (chosenPlace?.name.length ?? matched.length);
    this.parts.menu.style.setProperty(TYPED_CELLS, String(Math.max(cellsBefore, 0)));
  }
}
