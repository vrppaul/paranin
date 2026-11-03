// The script every page runs (internal/assets, Site), before the page is first painted: the stylesheet every page
// is drawn in, then the controllers over the markup the server draws.
import "../ui/ui.css";
import { startTheme } from "../controllers/theme";

// The system's own theme, which a page follows until its reader chooses one
const SYSTEM_DARK = "(prefers-color-scheme: dark)";

startTheme(document, () => localStorage, matchMedia(SYSTEM_DARK));
