// What a widget draws on its stage by (Stage.svelte): the stage's size, the design's colours in the theme the page is
// shown in, and whether the reader asks for motion held still.
import type { COLOURS, Theme } from "../tokens/colours";

/** The design's colours in one theme, each by its token's name. */
type Palette = (typeof COLOURS)[Theme];

/** What a widget draws by: the stage's size in CSS pixels, the theme's colours, and whether motion is held still. */
export interface Scene {
  width: number;
  height: number;
  palette: Palette;
  reducedMotion: boolean;
}

/** A widget's drawing of the whole stage, on its pen, in CSS pixels. */
export type Draw = (pen: CanvasRenderingContext2D, scene: Scene) => void;
