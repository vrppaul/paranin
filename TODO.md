# TODO

The order of work is `docs/FOUNDATION.md` §8; each item is one chunk, started on the owner's word. An item done is removed, not ticked off.

## The foundation

- The faces, each its own commit (§8, step 5): the `Link: rel="canonical"` header of a page's other formats comes with `face/html`; the rest of `face/tui`, a chunk each: the command line with completion, the window of keys, and `:e <slug>` opening a project's page; the extras (`:checkhealth`, `:Lazy`, `:hire`) as documents `compose` makes (ADR-0003); copying a value (OSC 52), the blinking cursor and the banner, which deletes `_old/tree/internal/tui`; `face/html` on templ, on the same `hosting/web`. Each reads the first version under _old/tree while the face is written, and deletes what it replaces; the archive `paranin-first-version-2026-11-01.tar.gz` beside the repository keeps it whole.
- The page (§8, step 6): `web/` with Vite, TypeScript and Svelte 5, its gates, `domain`, `ui`, then the islands one at a time; the rules of the page and the design arrive with it.
- Hosting (§8, step 7): SEC-6 and SEC-15 of `docs/REQUIREMENTS.md`; the reload on a change in the content folder once the changes settle (ADR-0003), beside SIGHUP; `inbox` with delivery to Telegram; `face/print` on Typst, the CV listed in `llms.txt` with it, as the first version listed it; the Dockerfile with both ports; deploy.
- The blog (§8, step 8) and phase 3 (§8, step 9).

## Documents

- `docs/ARCHITECTURE.md` beyond ADR-0001 and the SSH face still draws the first version: the web's component design (`internal/web`, embedded templates, `make assets`), the project layout's `internal/web` and `scripts/build-assets.sh`, and `PARANIN_MESSAGES_PATH` and `PARANIN_TRUST_PROXY`, which nothing reads yet. Rewrite each with the part that replaces it.

- `docs/WEB.md` copies facts of the résumé (the email, the employer) as examples: rewrite it with `face/html` against the fixture profile. Its skill groups still carry a `Key` and are grouped from a flat list by `skillGroups` (both gone by D16), and its client data the salary slider's `step` and `bigStep` (gone by D17, the form works them out of the range).

## Content

- `routes` parses the site's base address again to name the site after its host, though content's check parsed it already: content could keep a web address parsed, as it keeps a month, and hand routes its host.

- A fault names its field by a path written by hand beside the yaml tags (`experience[0].end`). yaml.v3's `Node` knows the line and column of each value, which would name the place (`profile.yaml:171:`) with no second copy of the schema; weigh it before the posts' loader is written.

## Before the design work

- Measure again what `docs/FOUNDATION.md` §3.4 says of the page (heights, the overflowing link, the date column at 390px), with the shots kept beside the numbers.

## The owner's


- Fill the profile's placeholders (`[later]`, `[remote?]`, `[notice period]`, `[service]`, `[N]`, `[20XX]`): no code, and it blocks any promotion.

## Hosting

- SEC-13 asks for one log: a panic inside a session's TUI is caught by Bubble Tea and logged by wish through its own logger, on standard error, beside the slog log. Serving the TUI through `bubbletea.MiddlewareWithProgramHandler` with a handler of our own that logs `Run`'s error would put it in ours.

## Upstream

- claude-review's `scripts/check_instruction_files.py` takes a word for a path by `/` or by a suffix it knows, and `.go` is not among them: a bare `main.go` in a rule is never checked. Proposed there; copied back here whole.
- The same script finds a file by the file system, ignored folders included: on the owner's machine `**/*_test.go` was matched by the old tree's tests before this repository held one. Asking git for its files would match only what a clone has.
