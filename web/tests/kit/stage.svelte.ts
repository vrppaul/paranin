// What a stage's tests turn as the island and a widget would: the page it is shown in and the props a widget hands it,
// each reactive, as a test cannot make them in a module of no runes.
import type { ShownPage } from "../../src/kit/page";
import type { Draw } from "../../src/kit/scene";
import type { Theme } from "../../src/tokens/colours";

/** A page whose theme and ask for still motion a test turns. */
export class TurnedPage implements ShownPage {
  theme: Theme = $state("light");
  reducedMotion = $state(false);
}

/** A stage's props, which a test changes as a widget would. */
export class StageProps {
  aspect: number;
  draw: Draw;

  constructor(aspect: number, draw: Draw) {
    this.aspect = $state(aspect);
    this.draw = $state(draw);
  }
}
