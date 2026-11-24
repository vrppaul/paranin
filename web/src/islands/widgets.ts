// A post's widgets come alive (docs/FOUNDATION.md §5.4, Layout of a post; DESIGN.md, Article), over the markup
// internal/face/html draws: each island is a widget's figure, named by its component, whose code the entry started to
// load from among the widgets of the build, which it alone knows. The component runs in the body of the figure's
// window with the props the server wrote there as data; once it runs, the mark that it does shows and the alt text,
// which stands for it where nothing runs, stands aside, still describing the figure to a screen reader. A widget whose
// code the build lacks, or whose code does not come, keeps its alt text. Every widget of the page is told the page it
// is shown in (src/kit/page.ts): its theme, read by the page's own rule as the reader or their system turns it, and
// whether the reader asks for motion held still.
import { mount, type Component } from "svelte";
import { prefersReducedMotion } from "svelte/motion";
import { createSubscriber } from "svelte/reactivity";
import { shownTheme, THEME_ATTRIBUTE, type SystemScheme } from "../controllers/theme";
import { SHOWN_PAGE_KEY, type ShownPage } from "../kit/page";
import type { Theme } from "../tokens/colours";
import { ISLAND, ISLAND_NAME } from "./markup";

// A widget's window's body, the mark that it runs and its alt text, as the server draws them; its props, the window's
// block of JSON, which templ writes with no part of its own; and the attribute that names what describes the figure
const BODY = '[data-part="widget-body"]';
const LIVE = '[data-part="widget-live"]';
const ALT = '[data-part="widget-alt"]';
const PROPS = 'script[type="application/json"]';
const DESCRIBED_BY = "aria-describedby";

/**
 * A widget's props, as the post writes them: an object of JSON, every widget's of the type its schema gives it, which
 * the release the page was drawn from held them to, so the page takes them as they stand.
 */
export type WidgetProps = Record<string, unknown>;

/** A widget's component, once its code has come. */
export type WidgetCode = Promise<Component<WidgetProps>>;

/** The code of a widget by its component's name, loading already; none where the build holds no such widget. */
export type WidgetCodes = (name: string) => WidgetCode | undefined;

/** The part of an island a selector finds, which the server draws in every island. */
function partOf(island: Element, selector: string): HTMLElement {
  const part = island.querySelector<HTMLElement>(selector);
  if (part === null) {
    throw new Error(`the island has no ${selector}`);
  }
  return part;
}

/** Whether what the props' block holds is an object of JSON, as the server writes every widget's props. */
function isProps(written: unknown): written is WidgetProps {
  return typeof written === "object" && written !== null && !Array.isArray(written);
}

/** The props the server wrote in an island. */
function propsOf(island: Element): WidgetProps {
  const written: unknown = JSON.parse(partOf(island, PROPS).textContent);
  if (!isProps(written)) {
    throw new Error("the widget's props are no object");
  }
  return written;
}

/** Runs the widget of one island, shown in the page, once its code has come, where the build holds its code. */
async function startWidget(island: Element, codes: WidgetCodes, shown: ShownPage): Promise<void> {
  const code = codes(island.getAttribute(ISLAND_NAME) ?? "");
  if (code === undefined) {
    return;
  }
  mount(await code, {
    target: partOf(island, BODY),
    props: propsOf(island),
    context: new Map([[SHOWN_PAGE_KEY, shown]]),
  });
  const alt = partOf(island, ALT);
  island.setAttribute(DESCRIBED_BY, alt.id);
  alt.hidden = true;
  partOf(island, LIVE).hidden = false;
}

/**
 * Runs the widget of every island of a page, each by itself, so one whose code does not come stops no other; once
 * every one has run or failed, fails with those that failed. Each is shown the page in the theme it is shown in,
 * chosen on its root or else the system's, and as either turns.
 */
export async function startWidgets(page: Document, codes: WidgetCodes, system: SystemScheme): Promise<void> {
  const shown = new PageShown(page.documentElement, system);
  const started = await Promise.allSettled(
    Array.from(page.querySelectorAll(ISLAND), (island) => startWidget(island, codes, shown)),
  );
  const failures = started.flatMap((result): unknown[] => (result.status === "rejected" ? [result.reason] : []));
  if (failures.length > 0) {
    throw new AggregateError(failures, "a widget did not run");
  }
}

/**
 * The page as its widgets are shown it: its theme read by the page's rule, from the choice on its root or else the
 * system's, each time either turns while a widget reads it, and the reader's ask for still motion. Nothing keeps a
 * theme of its own, so none can stand apart from the page's.
 */
class PageShown implements ShownPage {
  readonly #root: Element;
  readonly #system: SystemScheme;
  readonly #subscribe: () => void;

  constructor(root: Element, system: SystemScheme) {
    this.#root = root;
    this.#system = system;
    this.#subscribe = createSubscriber((update) => watchTheme(root, system, update));
  }

  get theme(): Theme {
    this.#subscribe();
    return shownTheme(this.#root, this.#system);
  }

  get reducedMotion(): boolean {
    return prefersReducedMotion.current;
  }
}

/** Calls update whenever the page's theme may turn, by its root's choice or the system's; gives back what stops it. */
function watchTheme(root: Element, system: SystemScheme, update: () => void): () => void {
  const observer = new MutationObserver(update);
  observer.observe(root, { attributeFilter: [THEME_ATTRIBUTE] });
  system.addEventListener("change", update);
  return () => stopWatchingTheme(observer, system, update);
}

/** Stops calling update as the page's theme may turn. */
function stopWatchingTheme(observer: MutationObserver, system: SystemScheme, update: () => void): void {
  observer.disconnect();
  system.removeEventListener("change", update);
}
