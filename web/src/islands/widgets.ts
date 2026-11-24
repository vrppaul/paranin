// A post's widgets come alive (docs/FOUNDATION.md §5.4, Layout of a post; DESIGN.md, Article), over the markup
// internal/face/html draws: each island is a widget's figure, named by its component, whose code the entry started to
// load from among the widgets of the build, which it alone knows. The component runs in the body of the figure's
// window with the props the server wrote there as data; once it runs, the mark that it does shows and the alt text,
// which stands for it where nothing runs, stands aside, still describing the figure to a screen reader. A widget whose
// code the build lacks, or whose code does not come, keeps its alt text.
import { mount, type Component } from "svelte";
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

/** Runs the widget of one island, once its code has come, where the build holds its code. */
async function startWidget(island: Element, codes: WidgetCodes): Promise<void> {
  const code = codes(island.getAttribute(ISLAND_NAME) ?? "");
  if (code === undefined) {
    return;
  }
  mount(await code, { target: partOf(island, BODY), props: propsOf(island) });
  const alt = partOf(island, ALT);
  island.setAttribute(DESCRIBED_BY, alt.id);
  alt.hidden = true;
  partOf(island, LIVE).hidden = false;
}

/**
 * Runs the widget of every island of a page, each by itself, so one whose code does not come stops no other; once
 * every one has run or failed, fails with those that failed.
 */
export async function startWidgets(page: ParentNode, codes: WidgetCodes): Promise<void> {
  const started = await Promise.allSettled(
    Array.from(page.querySelectorAll(ISLAND), (island) => startWidget(island, codes)),
  );
  const failures = started.flatMap((result): unknown[] => (result.status === "rejected" ? [result.reason] : []));
  if (failures.length > 0) {
    throw new AggregateError(failures, "a widget did not run");
  }
}
