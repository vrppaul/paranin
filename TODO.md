# TODO

The order of work is `docs/FOUNDATION.md` §8; each item is one chunk, started on the owner's word. An item done is removed, not ticked off.

## The foundation

- The faces, each its own commit (§8, step 5): the rest of `face/tui`: `:hire` once the form and the CV are there, since without them it repeats the ways the résumé opens with. Each reads the first version under _old/tree while the face is written, or in the archive where _old/tree lacks it (the TUI), and deletes what it replaces; the archive `paranin-first-version-2026-11-01.tar.gz` beside the repository keeps it whole.
- The page (§8, step 6), each part landing with its first user (D18), in this order:
  - The TUI's search marks the match its count names reversed, as the page's does (owner, 11-21), where today every match stands on the accent alike.
  - The first island, the form's salary range, waits for `inbox` (step 10). It brings Svelte, and svelte-check in place of the hook `page-types`'s `tsc`, and widens stylelint's glob to its components' styles; its script loads by a dynamic import, not with the site's entry, which holds the first paint.
- A tab is served at the name its title makes, and nothing refuses a title that takes another's address: a résumé's tab named `index` or `llms` takes `/index.md` or `/llms.txt`, two sections of a project with one name share one, a title of no letter takes its page's own. No title does today; the owner left it unchecked.
- At 53.75 to about 60rem wide the foot line wraps to two rows, more than the command line's menu keeps back (`row_3`), so in a window under about 175px tall the menu's top row is cut. Keeping back the line's real height would need the script to measure it; rare enough to leave until a reader meets it.
- The sitemap, when it is served, lists every tab's address beside its page's.
- A tab's name keeps every letter of its title, so a site in another script serves its tabs at addresses in that script.
- A security audit of the whole site (§8, step 7), before anything is served.
- The blog (§8, step 8): first agreed with the owner, how it works and what it holds; then built, before hosting.
- The content, changed together with the owner (§8, step 9).
- Hosting, last (§8, step 10): SEC-6 and SEC-15 of `docs/REQUIREMENTS.md`; the files of the snapshot before kept served beside the current ones, so a page fetched before a reload finds its stylesheet and chunks after it (ADR-0003), and with it whether a release switched in by the link and SIGHUP is watched too, since the watch would load the same release a second time and push out the snapshot before; `inbox` with delivery to Telegram, and the form the first screen's primary button opens, which does nothing until then; `face/print` on Typst, the CV listed in `llms.txt` with it, as the first version listed it, among the foot line's ways, its look the owner's to pick then, and before the profiles that end the first screen's brief (DESIGN.md, FootLine and Pages), which leave it out until it is served; the Dockerfile with both ports; deploy.
- Phase 3 (§8, step 11).
- A review of every test against `.claude/rules/tests.md`: which named behaviour each holds, and whether another test already holds it; a list to delete, merge or keep, for the owner, before any is changed. Its place in the order is the owner's.
- `scripts/tokens`' table of faulty files of tokens tells each fault by its words, which `.claude/rules/tests.md` refuses: each fault a kind of its own, as the least contrast's are, would let the table read which fault came.
- The command line's tests open each page of a test in jsdom's one document, so the line of a page opened before it still listens there and answers the keys too; a test of the page opened next passes while both answer.
- A window narrowed to a phone's width hides the vim keys' switch, and with it the keys, while a search kept stays marked there with no key to put it away; nothing watches the width, which lives in the stylesheet's media query alone.
- The theme and the vim keys' switch read the reader's choice once, as a page opens: a page the browser shows again from its cache keeps the choice it opened with, though a page opened after it changed it.

## Documents

- `docs/ARCHITECTURE.md`'s overview still draws the first version, every face reading `internal/content`: rewrite it from FOUNDATION §5.2's layers; it waits for no package.
- `docs/ARCHITECTURE.md`'s configuration lists `PARANIN_MESSAGES_PATH` and `PARANIN_TRUST_PROXY`, which nothing reads yet: rewrite them with `inbox` and SEC-6 (§8, step 10).

- `docs/WEB.md` copies facts of the résumé (the email, the employer) as examples: rewrite it with `face/html` against the fixture profile. Its skill groups still carry a `Key` and are grouped from a flat list by `skillGroups` (both gone by D16), and its client data the salary slider's `step` and `bigStep` (gone by D17, the form works them out of the range).

## Content

- A fault names its field by a path written by hand beside the yaml tags (`experience[0].end`). yaml.v3's `Node` knows the line and column of each value, which would name the place (`profile.yaml:171:`) with no second copy of the schema; weigh it before the posts' loader is written.

## Before the design work

- Measure again what `docs/FOUNDATION.md` §3.4 says of the page (heights, the overflowing link, the date column at 390px), with the shots kept beside the numbers.

## The owner's


- The HTML face's goldens hold every block-letter rectangle's position, and since the career line each role's share of it, which `.claude/rules/tests.md` refuses for a look: a fixture name no font draws and a fixture with no roles of known months (every face's fixture changes), or an exception for the drawings. His choice.
- The light theme's `muted` takes index 60 (#5f5f87, lavender) in a terminal of 256 colours, between grey text and grey faint; #5f5f5f would keep it grey. His eye.
- DESIGN.md's Table mutes a length, a skill's duration, but a table's cells carry no kind in the document, so the page draws every cell after the first in `text` (the TUI mutes them all): a cell kind in `document.Table` (a change to the document model, his word) would let each face style a length as such; the stylesheet does not read compose's column order.
- At 1280 to 1440 a section's block title shrinks to the left column the blocks leave it (Experience 274 of its 480px at 1280), and at full size again below 80rem, where it stands above: the blocks' column is as wide as an Entry so points keep `measure` beside their dates; the approved artifact held the column to `measure`, which keeps the titles whole and squeezes points to about 41 cells. His eye.
- In the light theme the foot line's session, `host` on `sunk` at `small`, reads 4.45:1, under the 4.5 the text's size asks; the canvas drew it so. A darker `host` would change the mail's host and the letters' gradient too. His eye.
- In the light theme the TUI's menu over the command line writes the letters typed in `accent` on `quiet`, about 1.9:1, and its kinds in `muted`, about 4.0:1; DESIGN.md never lets the accent be small text in Latte on the page, while the TUI writes it as text in either theme (code, the mailbox, the current role). Bold `text` for the typed letters in Latte would read. His eye.
- In the light theme what the page's command line says of a refused command, `error` on `sunk` at `small`, reads 4.46:1, under the 4.5 the text's size asks; Latte has no darker red, so a shade of its own would change `error` wherever it stands. His eye.
- The TUI's `:e` goes round the projects' pages in the order of their letters, which it reads off its buffers; the page's in the profile's, from `compose.Frame`. The TUI reading `Frame.Projects` too would give both one order. His word on which.
- The headline, "Builds backend services, and tools for people who live in the terminal.", is the profile's one line in the third person: "Builds" reads as "he builds". It heads the first screen and is every page's description, llms.txt's and the man page's too. "I build …" would speak as the rest of the site does. His words, his call.
- No page links `llms.txt`; the canvas left it out, and agents find it at the root by its name. His call.
- At 320px the toast of a long command wraps it at its hyphens ("add -" / "-scope user semantic-"), its lines flush against its hairlines, since a toast has no padding above and below; DESIGN.md's Command never wraps, its Toast does not say. His eye.
- On a phone a skill table's rows scroll inside it, and a long name (`Embeddings and vector search`) leaves the sources past the screen's edge until scrolled to; at 320px two tables are a pixel too wide and scroll for nothing. His eye.
- The site's icon in a picture of one theme, a phone's home screen and `/favicon.ico`, is drawn in the dark theme; the drawing a browser prefers follows the reader's. Its ground is a square, where the approved artifact rounded it (`rx` 2) and ringed it in `quiet`: a phone rounds its picture itself, and a ring a unit wide is half a pixel at 16. At 16px the P stands half a pixel high, since its 30 units of 32 cannot share the two left over in whole pixels. His eye.
- At 1280 and 1440 semantic-code-mcp's install line is a few cells wider than its fact's value column and scrolls inside its chip, though the blocks' column would hold it under its label. His eye.
- With a hint open, the theme button's tooltip stands under the window, which covers the line's right end over it; a window whose right edge stops before the theme button would leave it in sight. His eye.
- DESIGN.md's Dialog is a `muted` frame on the page's ground, its Hint an `edge` frame on `sunk`, as all that lies over the page; the Dialog, not yet built, could take the Hint's, or keep its own over its scrim. His eye.
- With a hint open and a search kept, one Esc closes the hint and puts the search away; Esc could close only what stands on top. His call.
- Fill the profile's placeholders (`[later]`, `[remote?]`, `[notice period]`, `[service]`, `[N]`, `[20XX]`): no code, and it blocks any promotion.

## Upstream

- Bubble Tea writes a panic it catches and its stack straight to standard error, and wish's middleware logs the error `Run` ended with through charm's own logger: neither takes ours. Our log holds a panic of the TUI's model (`internal/hosting/ssh`), but a panic in a command Bubble Tea runs reaches theirs alone. A way to hand Bubble Tea a logger, or a hook for the panic it catches, would make slog the one log ADR-0001 asks for.
- claude-review's `scripts/check_instruction_files.py` takes a word for a path by `/` or by a suffix it knows, and `.go` is not among them: a bare `main.go` in a rule is never checked. Proposed there; copied back here whole.
- The same script finds a file by the file system, ignored folders included: on the owner's machine `**/*_test.go` was matched by the old tree's tests before this repository held one. Asking git for its files would match only what a clone has.
