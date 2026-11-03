// The page's one build (docs/ARCHITECTURE.md, ADR-0004, D22): every entry the site names, written with a hash in
// its name into build/web, and a manifest naming each, which the Go server reads (internal/assets) and serves the
// files from. In development `pnpm watch` writes the same files again on every change, so the page is always
// served as it will be, under the same Content-Security-Policy. vitest runs the tests under tests/ in jsdom, a
// document with no browser.
import { defineConfig } from "vitest/config";

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
      input: ["src/entries/site.ts"],
    },
  },
  test: {
    environment: "jsdom",
    include: ["tests/**/*.test.ts"],
  },
  // vitest reads files through Vite, which reads none outside web/ unless told: the controllers' tests run over the
  // pages the HTML face draws, its golden files, and over nothing else of the repository
  server: {
    fs: {
      allow: [".", "../internal/face/html/testdata"],
    },
  },
});
