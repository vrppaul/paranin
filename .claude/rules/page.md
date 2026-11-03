---
paths:
  - "web/**"
  - "internal/assets/**"
---

# The page

How the page is built is said at the top of `web/vite.config.ts`; how the server reads the build, in the comment of `internal/assets`. The rules for any code are in `.claude/rules/code.md`.

- Every entry the site names is an input of `web/vite.config.ts` and a constant of `internal/assets`, and only there: a build that lacks one is refused when it loads, and CI loads each build it makes (`paranin check content build/web`).
- A colour is a token: `web/src/tokens/tokens.css` is written by `scripts/tokens` from `design/tokens.yaml`, never by hand, and every other stylesheet names a token, never a value (stylelint).
- A size, a time or a weight with a meaning is a custom property, with a comment saying what it is.
- The page's policy refuses an inline script, an inline style and a `data:` address (`docs/WEB.md`, "Security headers"): Vite inlines no file, and a template writes no `<style>` and no `style` attribute.
- The page is looked at in a browser, in both themes and at a phone's width, before a change to what it shows is called done: in the report, never in a test.
