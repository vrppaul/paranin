<!--
  A widget's stage: a canvas as wide as its window's body, as tall as the widget's shape makes it, sharp at the
  screen's pixel density, which the widget draws on in CSS pixels, in the colours of the theme the page is shown in. It
  is drawn once laid out, and again when its size or the screen's pixel density changes, when the theme turns or the
  reader's ask for still motion does, and when the widget hands it another drawing. The figure it stands in is named by
  its description already, so a screen reader passes the canvas by.
-->
<script lang="ts">
  import { devicePixelRatio } from "svelte/reactivity/window";
  import { COLOURS } from "../tokens/colours";
  import { readShownPage } from "./page";
  import type { Draw } from "./scene";

  interface Props {
    /** The stage's width over its height. */
    aspect: number;
    draw: Draw;
  }

  const { aspect, draw }: Props = $props();
  const page = readShownPage();
  let canvas: HTMLCanvasElement;
  /** The stage's box in CSS pixels, none until it is laid out. */
  let laidOut: DOMRectReadOnly | undefined = $state();

  // Sizing the canvas clears it and resets its pen, so each drawing starts on a clear stage, scaled to CSS pixels.
  // Its pixels are its box's in the screen's, rounded: devicePixelContentBoxSize would give them exactly, but Safari
  // has none, so a reader there would see no drawing
  $effect(() => {
    const pixelRatio = devicePixelRatio.current;
    if (laidOut === undefined || pixelRatio === undefined) {
      return;
    }
    canvas.width = Math.round(laidOut.width * pixelRatio);
    canvas.height = Math.round(laidOut.height * pixelRatio);
    const pen = canvas.getContext("2d");
    if (pen === null) {
      throw new Error("the browser draws on no canvas");
    }
    pen.setTransform(pixelRatio, 0, 0, pixelRatio, 0, 0);
    draw(pen, {
      width: laidOut.width,
      height: laidOut.height,
      palette: COLOURS[page.theme],
      reducedMotion: page.reducedMotion,
    });
  });
</script>

<canvas class="stage" bind:this={canvas} bind:contentRect={laidOut} style:--aspect={aspect} aria-hidden="true"></canvas>

<style>
  /* a block, so no line's room stands under it */
  .stage {
    display: block;
    width: 100%;
    aspect-ratio: var(--aspect);
  }
</style>
