// The page's one build (docs/ARCHITECTURE.md, ADR-0004, D22): every entry the site names, written with a hash in
// its name into build/web, and a manifest naming each, which the Go server reads (internal/assets) and serves the
// files from. In development `pnpm watch` writes the same files again on every change, so the page is always
// served as it will be, under the same Content-Security-Policy.
import { defineConfig } from "vite";

export default defineConfig({
  build: {
    outDir: "../build/web",
    emptyOutDir: true,
    // At the build's root, where internal/assets reads it; the server serves only the files it names, never it
    manifest: "manifest.json",
    // The page's policy refuses a data: address (docs/WEB.md, "Security headers"), so no file is inlined
    assetsInlineLimit: 0,
    rolldownOptions: {
      // Every entry the site names, each written once here and once in internal/assets, whose load refuses a
      // build lacking one
      input: ["src/ui/ui.css"],
    },
  },
});
