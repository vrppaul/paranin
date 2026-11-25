---
paths:
  - "web/**"
  - "internal/assets/**"
---

# The page

How the page is built is read in `web/vite.config.ts`; how the server reads the build, in `internal/assets`. What eslint and stylelint refuse, and why, is in `web/eslint.config.js` and `web/stylelint.config.js`.

- Every entry the site names is an input of `web/vite.config.ts` and a constant of `internal/assets`, and only there: a build that lacks one is refused when it loads, and CI loads each build (`paranin check content build/web build/cards`).
- The look of a part is its section of `DESIGN.md`, a reference read for the part in hand, not whole.
- A colour, a size, a time or a weight with a meaning is a token of `design/tokens.yaml`, which `scripts/tokens` writes into `web/src/tokens/tokens.css`; every other stylesheet names a token, never a value.
- A unit, a media query's width or a container query joins stylelint's lists with its first user, in the chunk that writes it; a kind of token the generator does not take yet, it learns with the token's first user.
- The page's policy refuses an inline script, an inline style and a `data:` address (`docs/WEB.md`, "Security headers"): Vite inlines no file, and a template writes no `<style>` and no `style` attribute.
- The parts of `web/src/` build on each other in one direction, as `LAYERS` in `web/eslint.config.js` lists them; a part joins the list with its first module.
- A value read by key or position may be missing (`noUncheckedIndexedAccess`): the code says what then.
- The site's entry holds a page's first paint (`blocking="render"`): it runs only what a page must have done before it is seen, and anything else loads by a dynamic import.
- A controller runs over the markup the server draws and finds it by `data-part`, never by an id; its tests run it over the HTML face's golden pages, so the two never drift apart.
- A choice of theme is an attribute on the page's root that `web/src/ui/ui.css` turns into `color-scheme`, never a style a script sets.
- The page is looked at in a browser, in both themes and at a phone's width, before a change to what it shows is called done: in the report, never in a test.
