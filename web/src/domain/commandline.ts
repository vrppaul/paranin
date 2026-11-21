// The command line the vim keys open (DESIGN.md, CommandLine), as the TUI's command line works
// (internal/face/tui/commands.go), with no page to touch: what Tab goes round for what is typed, a step round it, and
// what Enter does with the line. A place goes by its name, a tab of the page shown or a project's page; Neovim's :e
// opens a project's page by its slug too. What the line says of a command it refuses is Neovim's own words.

// What a place is, as the server writes it: a tab of the page shown, or a project's page, which :e opens too
const PLACE_KINDS = ["tab", "project"] as const;

/** What a place is: a tab of the page shown, or a project's page. */
export type PlaceKind = (typeof PLACE_KINDS)[number];

/** Somewhere the command line goes by its name: what it is, its address, and whether it is the page shown. */
export interface Place {
  name: string;
  kind: PlaceKind;
  address: string;
  current: boolean;
}

/** A page's places, worked out once: each name with the first place that has it, and the projects' pages. */
export interface Places {
  named: readonly Place[];
  projects: readonly Place[];
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

/** Why the line refuses a command: one it does not know, :e with no page, or :e of a page that is no project's. */
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

// The names of Neovim's command that opens a file, here a project's page
const EDIT_NAMES: readonly string[] = ["e", "edit"];

// What the line says of a command it refuses, as Neovim says it
const SAID: Record<Fault, (typed: string) => string> = {
  unknown: (typed) => `E492: Not an editor command: ${typed}`,
  argument: () => "E471: Argument required",
  missing: (typed) => `E447: Can't find file "${typed}" in path`,
};

/** The kind a word names, if it names one. */
export function placeKindOf(word: string | undefined): PlaceKind | undefined {
  return PLACE_KINDS.find((kind) => kind === word);
}

/**
 * A page's places: a name two share goes to the first, as the TUI's places do, so a project's page named as a tab is
 * goes by :e alone.
 */
export function placesOf(all: readonly Place[]): Places {
  const named = all.filter((place, index) => all.findIndex((each) => each.name === place.name) === index);
  return { named, projects: all.filter((place) => place.kind === "project") };
}

/**
 * What Tab goes round for a line typed, none of it chosen: the places a name typed starts, never the tab shown, and
 * before a letter the tabs alone; after :e and a space, the projects' pages its argument starts.
 */
export function completionsOf(places: Places, typed: string): Completions {
  const { name, argument, argued } = partsOf(typed);
  if (argued) {
    const listed = EDIT_NAMES.includes(name)
      ? places.projects
          .filter((place) => place.name.startsWith(argument))
          .map((place) => ({ place, written: name + SPACE + place.name }))
      : [];
    return { typed, listed, matched: argument, chosen: NONE_CHOSEN };
  }
  const listed = places.named
    .filter((place) => !place.current && place.name.startsWith(name) && (name !== "" || place.kind === "tab"))
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
  const project = places.projects.find((place) => place.name === argument);
  return project === undefined ? failed("missing", argument) : wentTo(project);
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
