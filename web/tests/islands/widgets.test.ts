// A post's widgets, run over the post the server draws: the HTML face's golden page (internal/face/html/testdata), which
// that face's own tests hold to what it writes. Its widgets are the tests' own (tests/widgets), found by their names as
// the site's entry finds the build's, by a glob of their folder, and loading already as the entry starts them.
import { beforeEach, expect, test } from "vitest";
import type { Component } from "svelte";
import { ISLAND, ISLAND_NAME } from "../../src/islands/markup";
import { startWidgets, type WidgetCode, type WidgetProps } from "../../src/islands/widgets";
import postPage from "../../../internal/face/html/testdata/blog/reading-slowly.html?raw";

// The body of an island's window its widget runs in, the mark that it runs and its alt text, as the server draws
// them; what the tests' widget draws of its props; and the attribute that names what describes a figure
const BODY = '[data-part="widget-body"]';
const LIVE = '[data-part="widget-live"]';
const ALT = '[data-part="widget-alt"]';
const READING_PACE = '[data-part="reading-pace"]';
const DESCRIBED_BY = "aria-describedby";
// The pace the fixture post writes for each of its widgets
const PACE = "slow";
// The tests' widget, and a name a test gives a widget whose code does not come
const WIDGET = "ReadingPace";
const FAILING = "Failing";

// The tests' widgets, each by its file's path
const TEST_WIDGETS = import.meta.glob<Component<WidgetProps>>("../widgets/*.svelte", { import: "default" });

/**
 * What a reader sees of an island: the words its widget draws in its window's body, none where none runs; whether it
 * shows its mark that it runs and its alt text; and whether a screen reader hears its alt text as the figure's
 * description.
 */
function shown(island: Element): { drawn: string | null; live: boolean; alt: boolean; described: boolean } {
  const alt = island.querySelector<HTMLElement>(ALT);
  return {
    drawn: island.querySelector(`${BODY} ${READING_PACE}`)?.textContent ?? null,
    live: !island.querySelector<HTMLElement>(LIVE)?.hidden,
    alt: !alt?.hidden,
    described: alt !== null && island.getAttribute(DESCRIBED_BY) === alt.id,
  };
}

/** What a reader sees of every island of the page. */
function shownIslands(): ReturnType<typeof shown>[] {
  return Array.from(document.querySelectorAll(ISLAND), shown);
}

/** The tests' widget's code, loading as the entry starts it. */
function loadedWidget(): WidgetCode | undefined {
  return TEST_WIDGETS[`../widgets/${WIDGET}.svelte`]?.();
}

beforeEach(() => {
  document.documentElement.innerHTML = new DOMParser().parseFromString(postPage, "text/html").documentElement.innerHTML;
});

test("each widget runs with the props the post writes, shows that it runs, and its alt text describes it", async () => {
  const code = loadedWidget();
  await startWidgets(document, () => code);
  expect(shownIslands()).toEqual([
    { drawn: PACE, live: true, alt: false, described: true },
    { drawn: PACE, live: true, alt: false, described: true },
  ]);
});

test("a widget whose code the build lacks keeps its alt text, and shows no mark that it runs", async () => {
  await startWidgets(document, () => undefined);
  expect(shownIslands()).toEqual([
    { drawn: null, live: false, alt: true, described: false },
    { drawn: null, live: false, alt: true, described: false },
  ]);
});

test("a widget whose code does not come keeps its alt text, and another still runs", async () => {
  document.querySelectorAll(ISLAND)[1]?.setAttribute(ISLAND_NAME, FAILING);
  const codes = new Map([
    [WIDGET, loadedWidget()],
    [FAILING, Promise.reject(new Error("the code did not come"))],
  ]);
  await expect(startWidgets(document, (name) => codes.get(name))).rejects.toThrow();
  expect(shownIslands()).toEqual([
    { drawn: PACE, live: true, alt: false, described: true },
    { drawn: null, live: false, alt: true, described: false },
  ]);
});
