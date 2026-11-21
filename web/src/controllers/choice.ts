// A reader's choice that holds on every page after, the theme, the vim keys or the search kept, where the browser
// keeps it under a key of its storage. storage is asked for the browser's storage each time it is needed, since a
// browser may refuse to hand it over, or to read or keep what it holds: a choice then holds for its page alone.

/** Where a choice is kept: the browser's storage, of which a choice reads, writes and forgets one key. */
export type ChoiceStore = Pick<Storage, "getItem" | "setItem" | "removeItem">;

/** The choice kept under a key, or nothing where none is or the browser will not read it. */
export function storedChoice(storage: () => ChoiceStore, key: string): string | null {
  try {
    return storage().getItem(key);
  } catch {
    // The browser hands over no storage, or will not read it: the reader has chosen nothing it can tell
    return null;
  }
}

/** Keeps a choice under a key, where the browser will keep it. */
export function keepChoice(storage: () => ChoiceStore, key: string, choice: string): void {
  try {
    storage().setItem(key, choice);
  } catch {
    // The browser keeps nothing: the choice holds for this page alone
  }
}

/** Forgets the choice under a key. */
export function forgetChoice(storage: () => ChoiceStore, key: string): void {
  try {
    storage().removeItem(key);
  } catch {
    // The browser keeps nothing, so it holds no choice to forget
  }
}
