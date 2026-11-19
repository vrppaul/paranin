# TODO

The order of work is `docs/FOUNDATION.md` §8; each item is one chunk, started on the owner's word. An item done is removed, not ticked off.

## The foundation

- The faces, each its own commit (§8, step 5): the rest of `face/tui`: `:hire` once the form and the CV are there, since without them it repeats the ways the résumé opens with. Each reads the first version under _old/tree while the face is written, or in the archive where _old/tree no longer holds it (the TUI), and deletes what it replaces; the archive `paranin-first-version-2026-11-01.tar.gz` beside the repository keeps it whole.
- The page (§8, step 6), each part landing with its first user (D18), in this order:
  - "The look" below, drawn on the page's own markup and stylesheet, the theme button among it; then the islands one at a time, each with the part of `domain` and `ui` it needs: the hint, the matcher with the document's `Widget` and tygo, the vim mode. The first island brings Svelte, and svelte-check in place of the hook `page-types`'s `tsc`, and widens stylelint's glob to its components' styles. An island's script loads by a dynamic import, not with the site's entry, which holds the first paint. The form and its salary range wait for `inbox` (step 7).
- The look, before hosting. The owner approved the first screen and the style guide on 2026-11-17; `DESIGN.md` holds it, and its tokens arrive in `design/tokens.yaml` each with its first user. Left, each a chunk, in this order:
  - the page's markup and stylesheet, part by part: the top bar kept over the first screen as it scrolls under it (the first screen, positioned for its crosses, paints over the sticky bar), with a layer token `scripts/tokens` learns; Contact's ways, the mail in its three colours; the 404 page and a project's head.
- `scripts/tokens` refuses a colour below the least contrast DESIGN.md's "Colour" gives it over the grounds it names (`muted`, `faint`, `label`), today worked out by hand; waits for the owner's word.
- Hosting (§8, step 7): SEC-6 and SEC-15 of `docs/REQUIREMENTS.md`; the files of the snapshot before kept served beside the current ones, so a page fetched before a reload finds its stylesheet and chunks after it (ADR-0003), and with it whether a release switched in by the link and SIGHUP is watched too, since the watch would load the same release a second time and push out the snapshot before; `inbox` with delivery to Telegram, and the form the first screen's primary button opens, which does nothing until then; `face/print` on Typst, the CV listed in `llms.txt` with it, as the first version listed it, and in the page's footer (DESIGN.md, Footer), which leaves it out until it is served; the Dockerfile with both ports; deploy.
- The blog (§8, step 8) and phase 3 (§8, step 9).

## Documents

- `docs/ARCHITECTURE.md`'s overview still draws the first version, every face reading `internal/content`: rewrite it from FOUNDATION §5.2's layers; it waits for no package.
- `docs/ARCHITECTURE.md`'s configuration lists `PARANIN_MESSAGES_PATH` and `PARANIN_TRUST_PROXY`, which nothing reads yet: rewrite them with `inbox` and SEC-6 (§8, step 7).

- `docs/WEB.md` copies facts of the résumé (the email, the employer) as examples: rewrite it with `face/html` against the fixture profile. Its skill groups still carry a `Key` and are grouped from a flat list by `skillGroups` (both gone by D16), and its client data the salary slider's `step` and `bigStep` (gone by D17, the form works them out of the range).

## Content

- A fault names its field by a path written by hand beside the yaml tags (`experience[0].end`). yaml.v3's `Node` knows the line and column of each value, which would name the place (`profile.yaml:171:`) with no second copy of the schema; weigh it before the posts' loader is written.

## Before the design work

- Measure again what `docs/FOUNDATION.md` §3.4 says of the page (heights, the overflowing link, the date column at 390px), with the shots kept beside the numbers.

## The owner's


- The HTML face's goldens hold every block-letter rectangle's position, and since the career line each role's share of it, which `.claude/rules/tests.md` refuses for a look: a fixture name no font draws and a fixture with no roles of known months (every face's fixture changes), or an exception for the drawings. His choice.
- On a phone each section's block title shrinks to its own column, so the five stand at different heights (320px: 27 to 40px), as DESIGN.md asks; and below about 375px the top bar's section names take a third row. His eye.
- The light theme's `muted` takes index 60 (#5f5f87, lavender) in a terminal of 256 colours, between grey text and grey faint; #5f5f5f would keep it grey. His eye.
- In the footer on a phone a long `curl` command wraps at a hyphen inside its code; DESIGN.md's Command never wraps, its Code does not say. His eye.
- DESIGN.md's Table mutes a length, a skill's duration, but a table's cells carry no kind in the document, so the page draws every cell after the first in `text` (the TUI mutes them all): a cell kind in `document.Table` (a change to the document model, his word) would let each face style a length as such; the stylesheet does not read compose's column order.
- At 1280 to 1440 a section's block title shrinks to the left column the blocks leave it (Experience 274 of its 480px at 1280), and at full size again below 80rem, where it stands above: the blocks' column is as wide as an Entry so points keep `measure` beside their dates; the approved artifact held the column to `measure`, which keeps the titles whole and squeezes points to about 41 cells. His eye.
- On a phone a skill table's rows scroll inside it, and a long name (`Embeddings and vector search`) leaves the sources past the screen's edge until scrolled to; at 320px two tables are a pixel too wide and scroll for nothing. His eye.
- At 1280 and 1440 semantic-code-mcp's install line is a few cells wider than its fact's value column and scrolls inside its chip, though the blocks' column would hold it under its label. His eye.
- Fill the profile's placeholders (`[later]`, `[remote?]`, `[notice period]`, `[service]`, `[N]`, `[20XX]`): no code, and it blocks any promotion.

## Upstream

- Bubble Tea writes a panic it catches and its stack straight to standard error, and wish's middleware logs the error `Run` ended with through charm's own logger: neither takes ours. Our log holds a panic of the TUI's model (`internal/hosting/ssh`), but a panic in a command Bubble Tea runs reaches theirs alone. A way to hand Bubble Tea a logger, or a hook for the panic it catches, would make slog the one log ADR-0001 asks for.
- claude-review's `scripts/check_instruction_files.py` takes a word for a path by `/` or by a suffix it knows, and `.go` is not among them: a bare `main.go` in a rule is never checked. Proposed there; copied back here whole.
- The same script finds a file by the file system, ignored folders included: on the owner's machine `**/*_test.go` was matched by the old tree's tests before this repository held one. Asking git for its files would match only what a clone has.
