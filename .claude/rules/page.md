---
paths:
  - "web/**"
  - "internal/assets/**"
---

# The page

How the page is built is said at the top of `web/vite.config.ts`; how the server reads the build, in the comment of `internal/assets`. The rules for any code are in `.claude/rules/code.md`.

- Every entry the site names is an input of `web/vite.config.ts` and a constant of `internal/assets`, and only there: a build that lacks one is refused when it loads, and CI loads each build it makes (`paranin check content build/web`).
- A colour is a token: `web/src/tokens/tokens.css` is written by `scripts/tokens` from `design/tokens.yaml`, never by hand, and every other stylesheet names a token, never a value (stylelint).
- A size, a time or a weight with a meaning is a token of `design/tokens.yaml` too, and a stylesheet names its property (stylelint).
- Outside `web/src/tokens/tokens.css` no unit stands but those a part has needed; a weight, a line, an opacity or a layer is a `var()`.
- A unit, a media query's width or a container query joins stylelint's lists with its first user, in the chunk that writes it.
- A kind of token the generator does not take yet, it learns with the token's first user.
- A text style is set as its size, its line and its weight: `font` and `font-variant` take `inherit` alone, and `all` is never set, since each sets the ligatures back on (stylelint).
- The page's policy refuses an inline script, an inline style and a `data:` address (`docs/WEB.md`, "Security headers"): Vite inlines no file, and a template writes no `<style>` and no `style` attribute.
- The parts of `web/src/` build on each other in one direction, as `LAYERS` in `web/eslint.config.js` lists them; a part joins the list with its first module.
- TypeScript is strict, with `noUncheckedIndexedAccess` and `exactOptionalPropertyTypes` (`web/tsconfig.json`): a value read by key or position may be missing, and the code says what then.
- The site's entry holds a page's first paint (`blocking="render"`): it runs only what a page must have done before it is seen, and anything else loads by a dynamic import.
- A controller runs over the markup the server draws and finds it by `data-part`, never by an id (eslint); its tests run it over the HTML face's golden pages, so the two never drift apart.
- A choice of theme is an attribute on the page's root that `web/src/ui/ui.css` turns into `color-scheme`, never a style a script sets (eslint): the build writes each `light-dark()` as variables the stylesheet's rules switch.
- The page is looked at in a browser, in both themes and at a phone's width, before a change to what it shows is called done: in the report, never in a test.
