// The script every page runs (internal/assets, Site), before the page is first painted: the stylesheet every page
// is drawn in, then the theme, which the page is first seen in, and the switch of the vim keys, which the foot line is
// first seen with; the controllers a page needs only once it is seen load after, by a dynamic import of their own,
// each handed what it waits on from the design's times.
import "../ui/ui.css";
import { startTheme } from "../controllers/theme";
import { startVimSwitch } from "../controllers/vimswitch";
import { TIMES } from "../tokens/times";

// The system's own theme, which a page follows until its reader chooses one
const SYSTEM_DARK = "(prefers-color-scheme: dark)";

startTheme(document, () => localStorage, matchMedia(SYSTEM_DARK));
startVimSwitch(document, () => localStorage);
void Promise.all([import("../controllers/commandline"), import("../controllers/search")]).then(
  ([{ startCommandLine }, { HighlightMarks, PageLayout }]) =>
    startCommandLine(
      document,
      (address) => location.assign(address),
      new HighlightMarks(CSS.highlights),
      new PageLayout(window),
      () => sessionStorage,
    ),
);
void import("../controllers/tooltip").then(({ startTooltips }) => startTooltips(document));
void import("../controllers/copy").then(({ startCopying }) => startCopying(document, navigator.clipboard, TIMES.toast));
