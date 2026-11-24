// The script every page runs (internal/assets, Site), before the page is first painted: the stylesheet every page
// is drawn in, then the theme, which the page is first seen in, and the switch of the vim keys, which the foot line is
// first seen with; the controllers a page needs only once it is seen load after, by a dynamic import of their own,
// each handed what it waits on from the design's times, and the outline the reading line it measures by. A page that
// shows a widget loads the islands, Svelte with them, and at once the code of each widget it shows, found by its name
// among the widgets the build holds, each a chunk of its own.
import type { Component } from "svelte";
import "../ui/ui.css";
import { startTheme } from "../controllers/theme";
import { startVimSwitch } from "../controllers/vimswitch";
import { islandNames } from "../islands/markup";
import type { WidgetProps } from "../islands/widgets";
import { SCREEN_SHARES } from "../tokens/shares";
import { TIMES } from "../tokens/times";

// The system's own theme, which a page follows until its reader chooses one
const SYSTEM_DARK = "(prefers-color-scheme: dark)";
// A window the browser opens over the page, a way's hint or a series' pick, which it closes on Esc after the page's
// listeners have heard the key
const OPEN_WINDOW = ":popover-open";
// Every widget's code, by its component's path, each loaded only when called: the folder internal/assets finds a
// widget's code in the manifest by
const WIDGETS = import.meta.glob<Component<WidgetProps>>("../widgets/*.svelte", { import: "default" });

startTheme(document, () => localStorage, matchMedia(SYSTEM_DARK));
startVimSwitch(document, () => localStorage);
// what stands over the page, which the tooltips and the command line ask before they answer Esc
const loadedOverlay = Promise.all([import("../controllers/tooltip"), import("../controllers/overlay")]).then(
  ([{ startTooltips }, { Overlay }]) =>
    new Overlay(startTooltips(document), () => document.querySelector(OPEN_WINDOW) !== null),
);
void Promise.all([import("../controllers/tooltip"), loadedOverlay]).then(([{ hideOnEsc }, overlay]) =>
  hideOnEsc(document, overlay),
);
void Promise.all([import("../controllers/commandline"), import("../controllers/search"), loadedOverlay]).then(
  ([{ startCommandLine }, { HighlightMarks, PageLayout }, overlay]) =>
    startCommandLine(
      document,
      (address) => location.assign(address),
      new HighlightMarks(CSS.highlights),
      new PageLayout(window),
      () => sessionStorage,
      overlay,
    ),
);
void import("../controllers/copy").then(({ startCopying }) => startCopying(document, navigator.clipboard, TIMES.toast));
// the components the page's islands show, each one's code loading beside the islands', not after them
const shownComponents = islandNames(document);
if (shownComponents.length > 0) {
  const loading = new Map(shownComponents.map((name) => [name, WIDGETS[`../widgets/${name}.svelte`]?.()]));
  void import("../islands/widgets").then(({ startWidgets }) => startWidgets(document, (name) => loading.get(name)));
}
void import("../controllers/outline").then(({ startOutline }) =>
  startOutline(
    document,
    window,
    (callback, options) => new IntersectionObserver(callback, options),
    SCREEN_SHARES.readingLine,
  ),
);
