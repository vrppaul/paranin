// The command line the vim keys open (DESIGN.md, CommandLine), over the markup internal/face/html draws: with the
// switch ticked and in sight, a colon typed on the page opens the line over the foot line, its field taking the keys;
// the menu shows the places that fit what is typed, and Tab and Shift+Tab go round them, as the TUI's do
// (src/domain/commandline.ts). Enter goes where the line says, or says why it will not until the next key; Esc,
// Backspace on an empty line and the focus leaving close it.
import {
  completionsOf,
  lineOf,
  placeKindOf,
  placesOf,
  ran,
  stepped,
  type Completions,
  type Place,
  type Places,
} from "../domain/commandline";

// The parts as the server draws them: the switch and its box; the line, the part a reader types in, its field and the
// menu of its places, each place with the letters typed of its name and the rest; what the line says of a command
// that failed; the keys it names, and the one that goes round the places
const SWITCH = '[data-part="vim-keys"]';
const BOX = "input";
const LINE = '[data-part="command-line"]';
const ENTRY = '[data-part="command-entry"]';
const FIELD = '[data-part="command-field"]';
const PLACES = '[data-part="places"]';
const PLACE = '[data-part="place"]';
const TYPED = '[data-part="place-typed"]';
const REST = '[data-part="place-rest"]';
const SAID = '[data-part="command-said"]';
const KEYS = '[data-part="command-keys"]';
const NEXT = '[data-part="command-next"]';

// The fields of a page a reader types in, which keep the keys typed: never a box or a button, as the switch is, which
// a press leaves the focus on
const TYPED_IN =
  "textarea, select, input:not([type=checkbox], [type=radio], [type=button], [type=submit], [type=reset])";

// How the stylesheet takes a part out of the page, as it takes the switch on a phone
const GONE = "none";

// What the server writes of a place being the page shown
const CURRENT = "current";

// The keys the line answers, as a browser names them
const COLON = ":";
const TAB = "Tab";
const ENTER = "Enter";
const ESCAPE = "Escape";
const BACKSPACE = "Backspace";

// Tab's step round the places, and Shift+Tab's
const FORWARD = 1;
const BACKWARD = -1;

// What the field says to a screen reader: whether its menu shows, and the place chosen on it
const EXPANDED = "aria-expanded";
const ACTIVE = "aria-activedescendant";
const SELECTED = "aria-selected";

// The cells of the line before the name its menu lists, which the stylesheet stands the menu's names under
const TYPED_CELLS = "--typed-cells";

/** Where the line goes: an address on the site, handed to the browser. */
type Visit = (address: string) => void;

/** The parts of the line the server draws. */
interface Parts {
  switchLabel: HTMLElement;
  box: HTMLInputElement;
  line: HTMLElement;
  entry: HTMLElement;
  field: HTMLInputElement;
  menu: HTMLElement;
  said: HTMLElement;
  keys: HTMLElement;
  next: HTMLElement;
}

/** A place on the menu: where it goes, and its element with the letters typed of its name and the rest. */
interface Item {
  place: Place;
  element: HTMLElement;
  typed: HTMLElement;
  rest: HTMLElement;
}

/** Runs a page's command line, where the page draws one, going where it says through visit. */
export function startCommandLine(page: Document, visit: Visit): void {
  const parts = partsOf(page);
  if (parts === undefined) {
    return;
  }
  const items = [...parts.menu.querySelectorAll<HTMLElement>(PLACE)].flatMap((element) => itemOf(element) ?? []);
  new CommandLine(page, parts, items, visit).listen();
}

/** The line's parts, or nothing where the page lacks one. */
function partsOf(page: Document): Parts | undefined {
  const switchLabel = page.querySelector<HTMLElement>(SWITCH);
  const box = switchLabel?.querySelector<HTMLInputElement>(BOX);
  const line = page.querySelector<HTMLElement>(LINE);
  const entry = line?.querySelector<HTMLElement>(ENTRY);
  const field = line?.querySelector<HTMLInputElement>(FIELD);
  const menu = line?.querySelector<HTMLElement>(PLACES);
  const said = line?.querySelector<HTMLElement>(SAID);
  const keys = line?.querySelector<HTMLElement>(KEYS);
  const next = line?.querySelector<HTMLElement>(NEXT);
  if (!switchLabel || !box || !line || !entry || !field || !menu || !said || !keys || !next) {
    return undefined;
  }
  return { switchLabel, box, line, entry, field, menu, said, keys, next };
}

/** A place as the server draws it on the menu, or nothing where it lacks a part or names no kind the line knows. */
function itemOf(element: HTMLElement): Item | undefined {
  const { name, place: word, address } = element.dataset;
  const kind = placeKindOf(word);
  const typed = element.querySelector<HTMLElement>(TYPED);
  const rest = element.querySelector<HTMLElement>(REST);
  if (name === undefined || address === undefined || kind === undefined || typed === null || rest === null) {
    return undefined;
  }
  return { place: { name, kind, address, current: CURRENT in element.dataset }, element, typed, rest };
}

/** Whether a key comes while the reader types in a field of the page, which keeps its keys. */
function inField(target: EventTarget | null): boolean {
  return target instanceof HTMLElement && (target.isContentEditable || target.matches(TYPED_IN));
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

/** The command line of a page: closed, open with the reader typing, or saying why it refused a command. */
class CommandLine {
  private readonly places: Places;
  private completions: Completions;
  private saying = false;
  // where the focus was before the line opened, which it goes back to
  private before: Element | null = null;

  constructor(
    private readonly page: Document,
    private readonly parts: Parts,
    private readonly items: readonly Item[],
    private readonly visit: Visit,
  ) {
    this.places = placesOf(items.map((item) => item.place));
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
  }

  /** A press on a place goes where the line would with the place written on it. */
  private pressed(place: Place): void {
    const completion = this.completions.listed.find((listed) => listed.place === place);
    this.went(completion?.written ?? place.name);
  }

  /** A key on the page: it puts away what the line said, and a colon opens the line where the keys are on. */
  private pageKey(event: KeyboardEvent): void {
    if (this.saying) {
      this.close();
    }
    const modified = event.ctrlKey || event.metaKey || event.altKey;
    if (event.key !== COLON || modified || !this.keysOn() || !this.parts.line.hidden || inField(event.target)) {
      return;
    }
    event.preventDefault();
    this.open();
  }

  /** Whether the keys are on: the switch ticked, and in sight, so the reader can untick it. */
  private keysOn(): boolean {
    return this.parts.box.checked && getComputedStyle(this.parts.switchLabel).display !== GONE;
  }

  /** A key in the field: Tab goes round, Enter runs the line, Esc and Backspace on an empty line close it. */
  private fieldKey(event: KeyboardEvent): void {
    switch (event.key) {
      case TAB:
        this.step(event.shiftKey ? BACKWARD : FORWARD);
        break;
      case ENTER:
        this.went(this.parts.field.value);
        break;
      case ESCAPE:
        this.finish();
        break;
      case BACKSPACE:
        if (this.parts.field.value !== "") {
          return;
        }
        this.finish();
        break;
      default:
        return;
    }
    event.preventDefault();
    // the page's own listener runs after, and must not take the key again
    event.stopPropagation();
  }

  private open(): void {
    this.before = this.page.activeElement;
    this.parts.field.value = "";
    this.parts.line.hidden = false;
    this.edited();
    // the line stands in the foot line, always in sight: a focus that scrolled would take the page to its end
    this.parts.field.focus({ preventScroll: true });
  }

  /** Closes the line the reader is done with, the focus handed back first, while the field can still let it go. */
  private finish(): void {
    this.handBack();
    this.close();
  }

  /** The focus left the field: the line closes, unless it is saying why it refused a command. */
  private left(): void {
    if (!this.saying) {
      this.close();
    }
  }

  private close(): void {
    this.saying = false;
    this.parts.line.hidden = true;
    this.parts.said.hidden = true;
    // an alert reads its words out only as they change: the same refusal said again must be heard again
    this.parts.said.textContent = "";
    this.parts.entry.hidden = false;
    this.parts.keys.hidden = false;
    this.parts.field.setAttribute(EXPANDED, String(false));
  }

  /**
   * Hands the focus back where it was before the colon; where nothing held it, the field lets it go, since a field
   * hidden with the focus would keep the keys from the page.
   */
  private handBack(): void {
    if (this.page.activeElement === this.parts.field) {
      this.parts.field.blur();
    }
    if (this.before instanceof HTMLElement) {
      this.before.focus({ preventScroll: true });
    }
  }

  private edited(): void {
    this.completions = completionsOf(this.places, this.parts.field.value);
    this.show();
  }

  private step(step: number): void {
    this.completions = stepped(this.completions, step);
    this.parts.field.value = lineOf(this.completions);
    this.show();
  }

  /** Goes where a line says, or says why it will not. */
  private went(line: string): void {
    const outcome = ran(this.places, line);
    switch (outcome.kind) {
      case "go":
        this.finish();
        this.visit(outcome.address);
        break;
      case "stay":
      case "nothing":
        this.finish();
        break;
      case "failed":
        this.say(outcome.message);
        break;
    }
  }

  /** Says on the line why a command was refused, in place of the field and its keys, until the next key. */
  private say(message: string): void {
    this.saying = true;
    this.handBack();
    this.parts.said.textContent = message;
    this.parts.said.hidden = false;
    this.parts.entry.hidden = true;
    this.parts.keys.hidden = true;
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
