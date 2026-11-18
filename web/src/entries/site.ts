// The script every page runs (internal/assets, Site), before the page is first painted: the stylesheet every page
// is drawn in, then the theme, which the page is first seen in; the controllers a page needs only once it is seen
// load after, by a dynamic import of their own.
import "../ui/ui.css";
import { startTheme } from "../controllers/theme";

// The system's own theme, which a page follows until its reader chooses one
const SYSTEM_DARK = "(prefers-color-scheme: dark)";

startTheme(document, () => localStorage, matchMedia(SYSTEM_DARK));
void import("../controllers/tooltip").then(({ startTooltips }) => startTooltips(document));
void import("../controllers/sections").then(({ startSections, WindowReading }) =>
  startSections(document, new WindowReading(window)),
);
