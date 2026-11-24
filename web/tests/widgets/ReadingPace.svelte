<!--
  The tests' widget, which the fixture post shows (internal/content/testdata): a widget as the islands mount one, its
  one prop drawn on the kit's stage, in the text's colour on the page's ground, typed by the props' schema beside it as
  the post writes them, which is what it gets. It stands with the tests, so the page's own build holds no widget until
  a post of the site shows one.
-->
<script lang="ts">
  import type { InferInput } from "valibot";
  import Stage from "../../src/kit/Stage.svelte";
  import type { Scene } from "../../src/kit/scene";
  import type { schema } from "./ReadingPace.props";

  // The stage's width over its height: a band across the window
  const ASPECT = 4;
  // The pace's type: the size of the page's text, in the system's monospace, as the tests' widget needs no face of the
  // design's
  const PACE_FONT = "1em monospace";
  // Where the pace stands: the stage's middle, as a share of its width and of its height
  const MIDDLE = 0.5;

  const { pace }: InferInput<typeof schema> = $props();

  /** Draws the pace in the stage's middle, on the page's ground. */
  function drawPace(pen: CanvasRenderingContext2D, { width, height, palette }: Scene): void {
    pen.fillStyle = palette.background;
    pen.fillRect(0, 0, width, height);
    pen.font = PACE_FONT;
    pen.textAlign = "center";
    pen.textBaseline = "middle";
    pen.fillStyle = palette.text;
    pen.fillText(pace, width * MIDDLE, height * MIDDLE);
  }
</script>

<Stage aspect={ASPECT} draw={drawPace} />
