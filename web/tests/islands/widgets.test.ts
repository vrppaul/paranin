// A post's widgets, run over the post the server draws: the HTML face's golden page (internal/face/html/testdata),
// which that face's own tests hold to what it writes, parsed afresh for each test, so no widget of one hears the page
// of another turn. Its widgets are the tests' own (tests/widgets), found by their names as the site's entry finds the
// build's, by a glob of their folder, and loading already as the entry starts them. jsdom lays nothing out and has no
// canvas, so a test lays each widget's stage out at a box of its own, and the stage draws with a pen that keeps the
// texts it draws.
import "../jsdom";
import { afterEach, expect, test, vi } from "vitest";
import { flushSync, type Component } from "svelte";
import { THEME_ATTRIBUTE } from "../../src/controllers/theme";
import { ISLAND, ISLAND_NAME } from "../../src/islands/markup";
import { startWidgets, type WidgetCode, type WidgetProps } from "../../src/islands/widgets";
import { COLOURS } from "../../src/tokens/colours";
import { FakeSystem, layOut, RecordingPen } from "../browser";
import postPage from "../../../internal/face/html/testdata/blog/reading-slowly.html?raw";

// The body of an island's window its widget runs in, the mark that it runs and its alt text, as the server draws
// them; what a widget draws on; and the attribute that names what describes a figure
const BODY = '[data-part="widget-body"]';
const LIVE = '[data-part="widget-live"]';
const ALT = '[data-part="widget-alt"]';
const CANVAS = "canvas";
// The box a test lays each widget's stage out at, in CSS pixels
const STAGE_BOX = { width: 320, height: 80 };
const DESCRIBED_BY = "aria-describedby";
// The pace the fixture post writes for each of its widgets
const PACE = "slow";
// The tests' widget, and a name a test gives a widget whose code does not come
const WIDGET = "ReadingPace";
const FAILING = "Failing";

// The tests' widgets, each by its file's path
const TEST_WIDGETS = import.meta.glob<Component<WidgetProps>>("../widgets/*.svelte", { import: "default" });

/**
 * What a reader sees of an island: whether its widget draws in its window's body, whether it shows its mark that it
 * runs and its alt text, and whether a screen reader hears its alt text as the figure's description.
 */
function shown(island: Element): { drawing: boolean; live: boolean; alt: boolean; described: boolean } {
  const alt = island.querySelector<HTMLElement>(ALT);
  return {
    drawing: island.querySelector(`${BODY} ${CANVAS}`) !== null,
    live: !island.querySelector<HTMLElement>(LIVE)?.hidden,
    alt: !alt?.hidden,
    described: alt !== null && island.getAttribute(DESCRIBED_BY) === alt.id,
  };
}

/** What a reader sees of every island of a page. */
function shownIslands(page: Document): ReturnType<typeof shown>[] {
  return Array.from(page.querySelectorAll(ISLAND), shown);
}

/** The post as a browser parses what the server sent. */
function parsedPost(): Document {
  return new DOMParser().parseFromString(postPage, "text/html");
}

/** The tests' widget's code, loading as the entry starts it. */
function loadedWidget(): WidgetCode | undefined {
  return TEST_WIDGETS[`../widgets/${WIDGET}.svelte`]?.();
}

/** A pen every widget's canvas draws with. */
function penForEveryCanvas(): RecordingPen {
  const pen = new RecordingPen();
  vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue(pen.asContext());
  return pen;
}

/** Lays every widget's stage of a page out, as a browser lays it out once the widget runs, and lets each draw. */
function layOutStages(page: Document): void {
  page.querySelectorAll(CANVAS).forEach((canvas) => layOut(canvas, STAGE_BOX));
  flushSync();
}

/** Waits until a change to a page's root is heard, as a MutationObserver hears it, and drawn. */
async function rootChangeDrawn(): Promise<void> {
  await Promise.resolve();
  flushSync();
}

afterEach(() => vi.restoreAllMocks());

test("each widget runs with the props the post writes, shows that it runs, and its alt text describes it", async () => {
  const pen = penForEveryCanvas();
  const page = parsedPost();
  const code = loadedWidget();
  await startWidgets(page, () => code, new FakeSystem(false));
  layOutStages(page);
  expect({ islands: shownIslands(page), drawn: pen.drawnTexts.map(({ text }) => text) }).toEqual({
    islands: [
      { drawing: true, live: true, alt: false, described: true },
      { drawing: true, live: true, alt: false, described: true },
    ],
    drawn: [PACE, PACE],
  });
});

test("a widget whose code the build lacks keeps its alt text, and shows no mark that it runs", async () => {
  const page = parsedPost();
  await startWidgets(page, () => undefined, new FakeSystem(false));
  expect(shownIslands(page)).toEqual([
    { drawing: false, live: false, alt: true, described: false },
    { drawing: false, live: false, alt: true, described: false },
  ]);
});

test("a widget whose code does not come keeps its alt text, and another still runs", async () => {
  penForEveryCanvas();
  const page = parsedPost();
  page.querySelectorAll(ISLAND)[1]?.setAttribute(ISLAND_NAME, FAILING);
  const codes = new Map([
    [WIDGET, loadedWidget()],
    [FAILING, Promise.reject(new Error("the code did not come"))],
  ]);
  await expect(startWidgets(page, (name) => codes.get(name), new FakeSystem(false))).rejects.toThrow();
  expect(shownIslands(page)).toEqual([
    { drawing: true, live: true, alt: false, described: true },
    { drawing: false, live: false, alt: true, described: false },
  ]);
});

test("a widget draws in the theme the page is shown in, and again as the reader turns it or the system's", async () => {
  const pen = penForEveryCanvas();
  const page = parsedPost();
  const system = new FakeSystem(false);
  const code = loadedWidget();
  await startWidgets(page, () => code, system);
  layOutStages(page);
  page.documentElement.setAttribute(THEME_ATTRIBUTE, "dark");
  await rootChangeDrawn();
  page.documentElement.removeAttribute(THEME_ATTRIBUTE);
  await rootChangeDrawn();
  system.turn(true);
  flushSync();
  expect(pen.drawnTexts.map(({ colour }) => colour)).toEqual([
    COLOURS.light.text,
    COLOURS.light.text,
    COLOURS.dark.text,
    COLOURS.dark.text,
    COLOURS.light.text,
    COLOURS.light.text,
    COLOURS.dark.text,
    COLOURS.dark.text,
  ]);
});
