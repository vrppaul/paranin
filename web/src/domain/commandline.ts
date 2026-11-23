// The command line the vim keys open (DESIGN.md, CommandLine), as the TUI's command line works
// (internal/face/tui/commands.go), with no page to touch: what Tab goes round for what is typed, a step round it, and
// what Enter does with the line. A place goes by its name, a tab of the page shown, a place of the page's own or a
// project's page; Neovim's :e opens a page by its name, a project's or one of the blog's. What the line says of a
// command it refuses is Neovim's own words.

// The kind of a tab of the page shown, which Tab offers before a letter is typed; every other kind the line shows as
// the server names it (document.PlaceKind), and treats alike
const TAB_KIND = "tab";

/**
 * Somewhere the command line goes: its name, its kind as the menu names it, its address, whether it is the page shown,
 * and whether :e opens it, where its name alone does not.
 */
export interface Place {
  name: string;
  kind: string;
  address: string;
  current: boolean;
  edited: boolean;
}

/**
 * A folder of the site where :e opens a page by its name though the menu lists none: the name a terminal writes for
 * it, and its address, whose every page the site answers, a missing one by its page of no page.
 */
export interface Folder {
  name: string;
  address: string;
}

/**
 * A page's places, worked out once: each name with the first place that has it, the pages :e opens, and the folder
 * where it opens any other.
 */
export interface Places {
  named: readonly Place[];
  edited: readonly Place[];
  folder: Folder;
}

/** A place Tab goes round, with what it writes on the line where it is chosen. */
interface Completion {
  place: Place;
  written: string;
}

/**
 * What Tab goes round for what was typed: the line as typed, the places that fit it, what of each one's name is
 * typed, and the place in the round of the one chosen, or NONE_CHOSEN while the line holds what was typed.
 */
export interface Completions {
  typed: string;
  listed: readonly Completion[];
  matched: string;
  chosen: number;
}

/** Why the line refuses a command: one it does not know, :e with no page, or :e of a page it does not open. */
type Fault = "unknown" | "argument" | "missing";

/** What Enter does: goes to an address, stays on the page shown, does nothing, or refuses, saying why. */
export type Outcome =
  | { kind: "go"; address: string }
  | { kind: "stay" }
  | { kind: "nothing" }
  | { kind: "failed"; fault: Fault; message: string };

// The place of the item chosen where none is, the line holding what the reader typed
const NONE_CHOSEN = -1;

// What typed takes in the round Tab goes: one place, before the first item and after the last
const TYPED_PLACES = 1;

// What parts a command's name from its argument
const SPACE = " ";

// The names of Neovim's command that opens a file, here a page by its name
const EDIT_NAMES: readonly string[] = ["e", "edit"];

// What the line says of a command it refuses, as Neovim says it
const SAID: Record<Fault, (typed: string) => string> = {
  unknown: (typed) => `E492: Not an editor command: ${typed}`,
  argument: () => "E471: Argument required",
  missing: (typed) => `E447: Can't find file "${typed}" in path`,
};

/**
 * A page's places: those it goes to by name, a name two share going to the first, as the TUI's places do, so a
 * project's page named as a tab is goes by :e alone; the pages :e opens; and the folder where :e opens a page the
 * menu does not list.
 */
export function placesOf(all: readonly Place[], folder: Folder): Places {
  const byName = all.filter((place) => !place.edited);
  const named = byName.filter((place, index) => byName.findIndex((each) => each.name === place.name) === index);
  return { named, edited: all.filter((place) => place.edited), folder };
}

/**
 * What Tab goes round for a line typed, none of it chosen: the places a name typed starts, never the tab shown, and
 * before a letter the tabs alone; after :e and a space, the pages it opens its argument starts.
 */
export function completionsOf(places: Places, typed: string): Completions {
  const { name, argument, argued } = partsOf(typed);
  if (argued) {
    const listed = EDIT_NAMES.includes(name)
      ? places.edited
          .filter((place) => place.name.startsWith(argument))
          .map((place) => ({ place, written: name + SPACE + place.name }))
      : [];
    return { typed, listed, matched: argument, chosen: NONE_CHOSEN };
  }
  const listed = places.named
    .filter((place) => !place.current && place.name.startsWith(name) && (name !== "" || place.kind === TAB_KIND))
    .map((place) => ({ place, written: place.name }));
  return { typed, listed, matched: name, chosen: NONE_CHOSEN };
}

/**
 * The completions a step on, as Neovim's wildmenu goes round, what was typed a place in the round before the first
 * and after the last.
 */
export function stepped(completions: Completions, step: number): Completions {
  const round = completions.listed.length + TYPED_PLACES;
  // the remainder taken twice, so a step back from the first place in the round comes round to the last
  const chosen = ((((completions.chosen + TYPED_PLACES + step) % round) + round) % round) - TYPED_PLACES;
  return { ...completions, chosen };
}

/** What the line holds: the item chosen, written, or what the reader typed. */
export function lineOf(completions: Completions): string {
  return completions.listed[completions.chosen]?.written ?? completions.typed;
}

/** What Enter does with a line: goes to the place it names or :e names, stays, does nothing, or refuses. */
export function ran(places: Places, line: string): Outcome {
  const typed = line.trim();
  if (typed === "") {
    return { kind: "nothing" };
  }
  const named = places.named.find((place) => place.name === typed);
  if (named !== undefined) {
    return wentTo(named);
  }
  const { name, argument } = partsOf(typed);
  if (!EDIT_NAMES.includes(name)) {
    return failed("unknown", typed);
  }
  if (argument === "") {
    return failed("argument", argument);
  }
  const opened = places.edited.find((place) => place.name === argument);
  if (opened !== undefined) {
    return wentTo(opened);
  }
  // the site answers in the blog's folder for every post, by its slug or its number, unlisted ones too
  const { folder } = places;
  return argument.startsWith(folder.name)
    ? { kind: "go", address: folder.address + argument.slice(folder.name.length) }
    : failed("missing", argument);
}

/** Where going to a place takes the reader: to its address, or nowhere where it is the page shown. */
function wentTo(place: Place): Outcome {
  return place.current ? { kind: "stay" } : { kind: "go", address: place.address };
}

/** A refusal for a fault, in Neovim's words for it. */
function failed(fault: Fault, typed: string): Outcome {
  return { kind: "failed", fault, message: SAID[fault](typed) };
}

/** A line read as Neovim reads its own: a command's name, then, after a space, its argument. */
function partsOf(typed: string): { name: string; argument: string; argued: boolean } {
  const line = typed.trimStart();
  const space = line.indexOf(SPACE);
  if (space < 0) {
    return { name: line, argument: "", argued: false };
  }
  return { name: line.slice(0, space), argument: line.slice(space + SPACE.length).trimStart(), argued: true };
}
