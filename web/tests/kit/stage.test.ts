// A widget's stage, mounted as the island mounts a widget, with the page it is shown in: jsdom lays nothing out and has
// no canvas, so a test lays the stage out at a box of its own and the stage draws with a pen that keeps what it is
// asked.
import "../jsdom";
import { afterEach, beforeEach, expect, test, vi } from "vitest";
import { flushSync, mount } from "svelte";
import Stage from "../../src/kit/Stage.svelte";
import { SHOWN_PAGE_KEY } from "../../src/kit/page";
import type { Draw, Scene } from "../../src/kit/scene";
import { COLOURS } from "../../src/tokens/colours";
import { layOut, RecordingPen, type Box } from "../browser";
import { StageProps, TurnedPage } from "./stage.svelte";

// The screen's pixels to a CSS pixel, as a phone's; the stage's width over its height; the boxes a test lays it out at
const PIXEL_RATIO = 2;
const ASPECT = 4;
const FIRST_BOX = { width: 320, height: 80 };
const WIDER_BOX = { width: 480, height: 120 };

/** A widget's drawing that keeps what it is handed each time the stage asks for it. */
class KeptScenes {
  readonly scenes: Scene[] = [];
  readonly draw: Draw = (_pen, scene) => this.scenes.push(scene);
}

/** Mounts a stage, in an element of its own, on a page, drawing with a pen, as a widget drawing by it; its canvas. */
function mountStage(page: TurnedPage, props: StageProps, pen: RecordingPen): HTMLCanvasElement {
  vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue(pen.asContext());
  const target = document.createElement("div");
  mount(Stage, { target, props, context: new Map([[SHOWN_PAGE_KEY, page]]) });
  flushSync();
  const canvas = target.querySelector("canvas");
  if (canvas === null) {
    throw new Error("the stage drew no canvas");
  }
  return canvas;
}

/** Lays the stage's canvas out at a box, and lets it draw. */
function layOutAndDraw(canvas: HTMLCanvasElement, box: Box): void {
  layOut(canvas, box);
  flushSync();
}

beforeEach(() => vi.stubGlobal("devicePixelRatio", PIXEL_RATIO));

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

test("nothing is drawn before the stage is laid out; then its canvas holds the screen's pixels of its box", () => {
  const pen = new RecordingPen();
  const kept = new KeptScenes();
  const canvas = mountStage(new TurnedPage(), new StageProps(ASPECT, kept.draw), pen);
  const drawnBeforeLaidOut = kept.scenes.length;
  layOutAndDraw(canvas, FIRST_BOX);
  expect({
    drawnBeforeLaidOut,
    canvas: { width: canvas.width, height: canvas.height },
    scale: pen.scale,
    drawnAt: kept.scenes.map(({ width, height }) => ({ width, height })),
  }).toEqual({
    drawnBeforeLaidOut: 0,
    canvas: { width: FIRST_BOX.width * PIXEL_RATIO, height: FIRST_BOX.height * PIXEL_RATIO },
    scale: PIXEL_RATIO,
    drawnAt: [FIRST_BOX],
  });
});

test("the widget draws again at the stage's new box", () => {
  const kept = new KeptScenes();
  const canvas = mountStage(new TurnedPage(), new StageProps(ASPECT, kept.draw), new RecordingPen());
  layOutAndDraw(canvas, FIRST_BOX);
  layOutAndDraw(canvas, WIDER_BOX);
  expect(kept.scenes.map(({ width, height }) => ({ width, height }))).toEqual([FIRST_BOX, WIDER_BOX]);
});

test("the widget draws in the theme shown, and again in the theme turned to", () => {
  const page = new TurnedPage();
  const kept = new KeptScenes();
  layOutAndDraw(mountStage(page, new StageProps(ASPECT, kept.draw), new RecordingPen()), FIRST_BOX);
  page.theme = "dark";
  flushSync();
  expect(kept.scenes.map(({ palette }) => palette)).toEqual([COLOURS.light, COLOURS.dark]);
});

test("the widget is told whether the reader asks for motion held still, and draws again as they turn it", () => {
  const page = new TurnedPage();
  const kept = new KeptScenes();
  layOutAndDraw(mountStage(page, new StageProps(ASPECT, kept.draw), new RecordingPen()), FIRST_BOX);
  page.reducedMotion = true;
  flushSync();
  expect(kept.scenes.map(({ reducedMotion }) => reducedMotion)).toEqual([false, true]);
});

test("a new drawing the widget hands the stage is drawn at once, and the old one is not asked again", () => {
  const first = new KeptScenes();
  const second = new KeptScenes();
  const props = new StageProps(ASPECT, first.draw);
  layOutAndDraw(mountStage(new TurnedPage(), props, new RecordingPen()), FIRST_BOX);
  props.draw = second.draw;
  flushSync();
  expect({ first: first.scenes.length, second: second.scenes.length }).toEqual({ first: 1, second: 1 });
});
