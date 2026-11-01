# TODO

The order of work is `docs/FOUNDATION.md` §8; each item is one chunk, started on the owner's word. An item done is removed, not ticked off.

## The foundation

- `document` and `compose` (§8, step 4; ADR-0003): the résumé as a document, `compose` their first user (D18). Each package's layer rule in `.golangci.yml` is tried on it as it lands.
- The faces, each its own commit (§8, step 5): `face/text` with `routes`, `hosting/web` serving its formats, the project pages and the index, the depguard rule that keeps `content` out of every face, and one that holds `routes` to `document` alone; `face/tui` with `hosting/ssh`, `look` and the token generator from `design/tokens.yaml`; `face/html` on templ, on the same `hosting/web`. The site's word catalogue arrives with the first of them. Each reads the first version under _old/tree while the face is written, and deletes what it replaces; the archive `paranin-first-version-2026-11-01.tar.gz` beside the repository keeps it whole.
- The page (§8, step 6): `web/` with Vite, TypeScript and Svelte 5, its gates, `domain`, `ui`, then the islands one at a time; the rules of the page and the design arrive with it.
- Hosting (§8, step 7): SEC-6 and SEC-15 of `docs/REQUIREMENTS.md`; `inbox` with delivery to Telegram; `face/print` on Typst; the Dockerfile with both ports; deploy.
- The blog (§8, step 8) and phase 3 (§8, step 9).

## Documents

- `docs/WEB.md` copies facts of the résumé (the email, the employer) as examples: rewrite it with `face/html` against the fixture profile. Its skill groups still carry a `Key` and are grouped from a flat list by `skillGroups` (both gone by D16), and its client data the salary slider's `step` and `bigStep` (gone by D17, the form works them out of the range).

## Content

- A fault names its field by a path written by hand beside the yaml tags (`experience[0].end`). yaml.v3's `Node` knows the line and column of each value, which would name the place (`profile.yaml:171:`) with no second copy of the schema; weigh it before the posts' loader is written.

## Before the design work

- Measure again what `docs/FOUNDATION.md` §3.4 says of the page (heights, the overflowing link, the date column at 390px), with the shots kept beside the numbers.

## The owner's

- Fill the profile's placeholders (`[later]`, `[remote?]`, `[notice period]`, `[service]`, `[N]`, `[20XX]`): no code, and it blocks any promotion.

## Upstream

- claude-review's `scripts/check_instruction_files.py` takes a word for a path by `/` or by a suffix it knows, and `.go` is not among them: a bare `main.go` in a rule is never checked. Proposed there; copied back here whole.
- The same script finds a file by the file system, ignored folders included: on the owner's machine `**/*_test.go` was matched by the old tree's tests before this repository held one. Asking git for its files would match only what a clone has.
