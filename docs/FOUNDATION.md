# Foundation: review, target construction and plan

What we found in the code as it stands on 2026-11-01, what we take from
claude-review and livereel, the construction we build towards, and the order
of work. Every ADR this plan names starts from a section here. §10 records
the reviews this document has been through.

It is the review of one day and the plan made from it. What is current lives
elsewhere: each decision, once recorded, in its ADR (`docs/ARCHITECTURE.md`);
the work still to do, in `TODO.md`. Where they and this document differ, they
hold.

Each decision carries one of three marks:

- **Decided**: Pavel said so. It changes only on his word.
- **Proposed**: the plan's answer, waiting for Pavel.
- **Open**: a question nobody has answered yet.

## 1. Decided

| # | Decision | Date |
| --- | --- | --- |
| D1 | Nothing goes into git that we can already name as wrong. The first commit is a clean foundation that passes every gate, not a snapshot of today's tree. | 2026-11-01 |
| D2 | Go is written to the same standard as any other code: whole words for every name. One-letter habits of Go code (`w`, `r`, `m`, `i`, `s`) are not taken over. | 2026-11-01 |
| D3 | Plain JavaScript is not the base of the front end. The site will grow (articles, widgets, chat), and the front end is chosen for that, with any tools it needs. | 2026-11-01 |
| D4 | The plan looks ahead to every phase of `docs/ROADMAP.md`, not only to the next step. | 2026-11-01 |
| D5 | Posts are published without rebuilding or redeploying the binary. | 2026-11-01 |
| D6 | Posts live in this repository, which is public. Drafts and edits being visible there is fine. | 2026-11-01 |
| D7 | A post is `draft`, `published` or `unlisted` (§5.4). | 2026-11-24 |
| D8 | A post is written once, in one format, and every face draws that same text, the SSH TUI included (§5.4). | 2026-11-01 |
| D9 | Go draws every page with templ; the interactive parts are Svelte 5 islands (option A, §5.1). | 2026-11-01 |
| D10 | Typst makes `cv.pdf`, the tailored CV and the social cards, in place of headless Chrome (§5.2). | 2026-11-01 |
| D11 | `err` and `ctx` are the only abbreviations in Go names, beside the short words of §6.1 and `_`. | 2026-11-01 |
| D12 | The blog is in English first; translations come later, on the model of §5.4, with `lang` and the word catalogue there from the start. | 2026-11-01 |
| D13 | Posts go first to LinkedIn, HN and Lobsters, Reddit, dev.to, and X or Bluesky (§5.7). | 2026-11-01 |
| D14 | An initialism of two letters, written as Go writes it (`ID`, `IO`, `IP`, `OS`, `UI`, `VM`), is a word inside a name: `clientIP`, `uiState`. Alone it stays refused, but for `id` (D11). | 2026-11-01 |
| D15 | Time is the `time` package's, in code and in tests; a test holds a moment or a wait inside `testing/synctest`. No `Clock` is handed to the parts. | 2026-11-01 |
| D16 | The profile nests skills under their groups: a group is its title and its skills, each in its order. No key joins a skill to its group. | 2026-11-01 |
| D17 | The content holds facts, not how a face shows them: the links' short keys and the salary slider's steps leave the profile. A face shows a link by its label; the form works its steps out of `min` and `soft_max` (the author's choice, on Pavel's leave to choose). | 2026-11-01 |
| D18 | Nothing lands before its user: `document` and `compose` arrive in one chunk, `compose` making the résumé their first user; `routes` arrives with `face/text`, the first face to write an address; `look` and the token generator with `face/tui`, the first that needs a colour. | 2026-11-01 |
| D19 | Every face draws the résumé, the project pages and the posts from one document model, the HTML page and the islands included. The document has the résumé's own kinds beside markdown's, and its kinds are a closed set every face switches over. Its parts, in ADR-0003, are the author's, on Pavel's leave to choose. | 2026-11-01 |
| D20 | The layers of ADR-0003 are accepted. | 2026-11-01 |
| D21 | A page's own address chooses its format by the Accept header's ranking alone, with `Vary: Accept`, HTML among the formats once it is served: no reader is told apart by its agent (the author's choice, on Pavel's leave to choose). | 2026-11-01 |
| D22 | The one front-end build of ADR-0004 is accepted as its bullet is written: one Vite build writing hashed files, one manifest and each widget's props as JSON Schema, files served from a folder and none embedded, the manifest and the props checked when a snapshot loads, `vite build --watch` in development. Embedding would ship a post's widgets only with a new binary (D5), Vite's dev server breaks the CSP (§2), and two builds would ship Svelte twice. | 2026-11-03 |
| D23 | The blog holds the author's own thoughts and what he learned, each written up as an article; the original in English, translations beside it (D12). Interactive parts at least on the web, a reduced post over SSH (§5.4). | 2026-11-22 |
| D24 | A post has revisions: the reader sees which date's version they read and can open an older one (§5.4, Revisions). | 2026-11-22 |
| D25 | Posts group into series, and carry tags; the blog lists its posts by year or by series, and a tag filters the list (§5.4, Series, tags and lists; the tag alone, owner, 2026-11-24). | 2026-11-22 |
| D26 | A post has a slug and a short number, both reached as `blog/…` in every face: `paran.in/blog/12`, `:e blog/12`, `ssh blog/12@paran.in` (§5.4, Addresses). | 2026-11-22 |
| D27 | The whole site, the blog among it, has good SEO and Open Graph (§5.4, SEO and Open Graph). | 2026-11-22 |

## 2. What must hold whatever we build

Taken from `docs/WEB.md`, `docs/REQUIREMENTS.md` and `docs/ROADMAP.md`:

- One source of facts, many faces: HTML, markdown, the man page for `curl`,
  the SSH TUI, `cv.pdf`, `llms.txt`.
- The web résumé reads whole without JavaScript, and in a terminal through
  `curl`. Search engines and LLM crawlers see the same facts.
- `script-src 'self'`: no inline script, no inline handler, no script from
  another origin, in development as in production.
- Phase 3, the brain: an LLM grounded in the content, with caps on cost,
  in the browser and over SSH (phase 4), answers streamed.
- Phase 5: a CV tailored to a job description, made as a PDF on the server
  on request, never inventing experience.
- Phase 6, if it ever comes: a draft-and-approve helper for recruiters'
  messages, another caller of the brain.

## 3. Review of the code as it stands

The paths below are the first version's. It is not in git: it is archived
beside the repository, and what the rebuild has not yet replaced waits in
_old/tree on the owner's machine; `internal/tui` is in the archive alone.

### 3.1 Architecture

1. **Content is part of the binary.** `internal/content/resume.go:19` embeds
   `profile.yaml`; `internal/web/server.go:20` embeds templates and static
   files. A post cannot change without a new build, which breaks D5.
2. **Each face lays out the content by hand.** `index.html`,
   `internal/web/markdown.go`, `internal/web/manpage.go`,
   `internal/web/page.go` (client JSON, JSON-LD) and `internal/tui/pages.go`
   each order the sections and pick the facts again.
   `TestTextCopiesCoverTheResume` (`internal/web/text_test.go:9`) holds the
   markdown and the man page to the profile's facts; no test holds the HTML
   or the TUI to them.
3. **No one map of addresses.** `"/p/"` is built in 10 places outside tests,
   in `internal/web/{server,files,page,manpage,markdown}.go` and
   `internal/content/profile.go`.
4. **`content` is not neutral.** `content.Profile` is a view model for the
   TUI; `FromResume` writes a web route (`base + "/#contact"`); `Month`,
   `Span`, `Years`, `HostPath` and the skill groups' titles are presentation.
5. **`"present"` is a magic string**, checked in `internal/content/resume.go:252`,
   `internal/content/format.go:28`, `internal/web/page.go:262` and
   `internal/web/templates/index.html:9`.
6. **`web.Page` serves every template.** `Project` is nil on the index;
   `NotFoundPath` exists for the 404 alone; `ClientData` (skills, health,
   salary) ships on every page, the 404 included.
7. **JSON-LD is untyped.** `profileJSONLD` (`internal/web/page.go:252`)
   builds it as `map[string]any` from typed input.
8. **The TUI rebuilds the page document on every key.** `currentDoc()` is
   called from `follow`, `copyCurrent`, `search` and `showMatch`
   (`internal/tui/keys.go:107,184,203`, `internal/tui/model.go:186`), and
   `render` builds it again (`internal/tui/render.go:99`). `tui.Model` holds
   24 flat fields: modes, blink, count, search, prompt.
9. **The palette is written twice**: `internal/web/static/ui.css` and
   `internal/tui/theme.go:20` hold the same hex values, and the tint and
   cursorline strengths (0.16, 0.07) are written again in Go.
10. **The health levels are written three times**: the check in
    `internal/content/resume.go:287`, the table in
    `internal/tui/pages.go:159-161` and `HEALTH_LEVELS` in `site.js:7`, each
    with its own words for them.
11. **Plain text is chosen by guessing.** `negotiateVary`
    (`internal/web/server.go:150`) picks the format by a list of User-Agent
    prefixes and sets `Vary: Accept, User-Agent`, which splits a CDN's cache
    by every browser version.
12. **The package `server` is the SSH server**; HTTP lives in `web`. The name
    says less than it should.
13. **Imports are clean today** (`cmd → server, web → tui → content`), held by
    nothing.

### 3.2 Code against the owner's rules

| Rule | Where it breaks |
| --- | --- |
| Whole names | Go: `r`, `p`, `w`, `s`, `g`, `d`, `m`, `l`, `c`, `f`, `e` throughout, struct fields too (`internal/tui/doc.go:95-96`); TUI kinds `kText…`, `seg`, `pick`, `stem`. JS: `el`, `at`, `i`. |
| No function inside a function | `validate` → `require`, `monthOrGap` (`resume.go`); `parseTemplates` → `withLayout` (`page.go`); `appendGaps` → `add` (`internal/content/marks.go`); `blend` → `mix` (`theme.go`); `statusline` → `label` (`render.go`); `hero` → `bar` (`doc.go`); `buildChrome` → `find`, and six helpers inside `initSalary` (`site.js`). |
| Types, not tuples | `[][2]string` in `internal/tui/render.go:240-274` (`hints`, `keyHelp`) and `internal/web/manpage.go:80-278` (`pairs`, `factPairs`); arrays for records in `site.js` (`HEALTH_LEVELS`, `HINTS`). |
| No meaningful literal in an expression | `max(columnWidth(...), 14)`, `padded(label, 9)`, `gap(2)` in `pages.go`; 80 and 24 in `tui.New` (`internal/tui/model.go:66,69`); blend strengths in `theme.go`; breakpoints in `site.css`; `eq .ContactResult "sent"`. |
| No HTML as strings in code | `htmlMarks` in `internal/web/marks.go`; `innerHTML` at `site.js:583` (`CHROME_HTML`), `:1260` and `:1261`. |
| Facts live in the profile alone | "Pavel" in `index.html:20,197,201,203` and `site.js:27,312,350`; `TECH_WORDS`, 86 words of vocabulary, in `site.js:13`. |
| No global mutable state | one `state` object (`site.js:122`) for a 1404-line script. |
| Elements found by what they are | 10 calls of `getElementById` in `site.js`; styles written straight onto elements at `site.js:446,1071,1072`. |
| Tests hold behaviour, with no branch | 63 Go tests hold 143 `if`; most are assertions, the rest are not sorted yet. |
| Tests stand on their own data | `resume_test.go:19-33` edits the live `profile.yaml` by string replacement; `server_test.go:147` expects the real projects in the sitemap. Editing the résumé breaks the suite. |

Done well and kept: strict YAML (`KnownFields`), the profile checked at
start, templates parsed at start, ETags, a rate limiter that groups IPv6 by
/64 and trusts `X-Forwarded-For` only when told to, a cross-site check on the
form, `%q` in the 404 a terminal reads, the TUI palette falling back to 256,
16 and no colours.

### 3.3 Operations and the tree

- `data/messages.jsonl` (`cmd/paranin/main.go:60`) collects contact messages
  and nobody reads it. The site's main channel is deaf.
- HTTP defaults to `localhost:8080` (`cmd/paranin/main.go:67`) and the
  `Dockerfile` exposes only 23234: the container serves no web.
  `docker-compose.yml` is a development setup.
- Open in `docs/REQUIREMENTS.md`: SEC-6 (per-IP SSH limits), SEC-15 (log
  injection). No CI, no `govulncheck`, no linter config.
- JetBrains Mono comes from Google Fonts (`layout.html:36-38`, `cv.html:9`,
  `og.html:10`): a request to Google from an EU site (the 2022 Munich ruling
  on Google Fonts) and a third party in the path of every page.
- `cv.pdf` (93 KB) and `og.png` (31 KB) are build outputs in
  `internal/web/static/`, made by `scripts/build-assets.sh` with a running
  server, Google Chrome and `pdftotext`. Under D1 they never enter git.
- The git index is stale: it holds an older generation of 13 files
  (`internal/tui/styles.go`, an older `internal/tui/model.go`). No commit is
  built from it.
- `site.updated` (`profile.yaml:13`) is kept by hand, and the sitemap's
  `lastmod` reads it.
- `README.md` says phase 1 is early; `REQUIREMENTS.md` covers SSH alone; the
  blog is in no document.

### 3.4 Design

Kept: one identity across web, TUI, PDF and the social card. Monospace, the
Catppuccin latte and macchiato palettes, the 1ch × 24px grid, the cursor, the
corner crosses, the vim and tmux metaphor. The rules at the top of `ui.css`.

To change (seen in Playwright shots at 1440px and 390px on 2026-11-01;
measured again before design work starts):

1. `ssh paran.in`, the site's best hook, lives in a pop-up that covers the
   summary at 1440px and the foot of the screen on a phone. It belongs under
   the name in the first screen.
2. After the first screen there is no hierarchy: eight sections of one
   weight, each an upper-case header with a count and a rule. The page is
   about 4100px tall on a desktop and 6400px on a phone.
3. Placeholders show on the page: `[later]`, `[remote?]`, `[notice period]`,
   `[service]`, `[N]`, `[20XX]` (`profile.yaml:42-188`). A production start
   should refuse a profile with gaps.
4. Monospace tires a reader over long prose. Articles need a reading layout:
   a second, proportional face for prose, about 65–70ch, figures and widgets
   wider than the text.
5. Bugs: the repository link overflows the project page's side column at
   1440px; the `semantic-code-mcp` install command is cut with no sign it
   scrolls; the buffer tabline shows on project pages alone; there is no
   favicon (`href="data:,"`); at 390px the 12ch date column leaves about 25
   characters for the text.

### 3.5 Distribution

Present: Open Graph and Twitter tags, one `og.png`, JSON-LD `Person`,
`llms.txt`, a sitemap, `rel="me"`. Missing: a card per page, a feed (Atom),
`og:type=article` and `BlogPosting` for posts, analytics, links to the places
posts will go, a canonical link back from cross-posts, static stand-ins for
widgets where a platform runs no script.

## 4. What we take from claude-review and livereel

Both hold their rules with tools at every commit. paranin holds them in the
author's head, which is why one review found so many breaks.

| Taken | From | Where it lands here |
| --- | --- | --- |
| `AGENTS.md` with `CLAUDE.md` as its symlink; rules per area in `.claude/rules/*.md`, each loaded by its `paths:` | both | §7 |
| "How we work": talk short, small chunks, step back first, each hands over, understand before writing, tests hold behaviour, review before presenting, weigh then decide | both | `AGENTS.md` |
| "Only better": construction first, simplest first, root cause first, facts not adjectives, the verdict is somebody else's, a regression found becomes a gate, what can be named as worse is not committed | both | `AGENTS.md` |
| `scripts/check_instruction_files.py`, taken whole | claude-review, as livereel took it | gate |
| Layers held by a linter (`LAYERS`, import-linter) | both | `depguard`, eslint `LAYERS` |
| Short names, nested helpers and numbers refused by a linter (`id-length`, `NO_NESTED_HELPERS`, `no-magic-numbers`, complexity 10) | both | golangci, `scripts/coderules`, eslint |
| Colours only in `tokens.css`; sizes off one scale; motion only in named rules | claude-review, livereel's stage | stylelint |
| A component library: every kind of element in one place, refused elsewhere | claude-review `lib/ui` | `face/html/ui`, `web/src/ui` |
| A script that finds elements by `data-part` and writes no style, only custom properties | livereel `stage.js` | `web/src/dom`, `controllers` |
| Types for what the server sends generated from the server, checked by a hook | claude-review `api:check` | tygo |
| Design rules from the owner's canvas; a sheet of states in both themes before a design is built | claude-review `design*.md` | `DESIGN.md`, `.claude/rules/design.md` |
| Test rules: one behaviour per test, no branch, no look, the cheapest level, a code rule is a gate and never a test | both | `.claude/rules/tests.md` |
| `TODO.md` (a done item is removed), `CHANGELOG.md`, Conventional Commits, typos | both | repo root, gates |

Not taken: client rendering of the whole page (claude-review is a client
app; paranin is a document first), and livereel's plain JS with JSDoc (D3).

## 5. Target construction

### 5.1 Who draws the HTML — **Decided: A** (D9)

| Option | For | Against |
| --- | --- | --- |
| **A. Go draws every page (templ); interactive parts are Svelte islands** | One process, one binary. Facts and every face stay in Go. Without JavaScript everything but the islands works. | Two component systems; one CSS library and a gate keep them alike. |
| B. SvelteKit draws the web on Node; Go serves an API, SSH and the text faces | One component system for the whole web; widgets rendered on the server | Two runtimes in production; every page a call from Node to Go; a proxy in front to send `curl` to Go; twice the deploy |
| C. Astro or another static generator | Pleasant to write articles (MDX) | A second brain outside Go; the TUI, man page and chat still need Go; every post a rebuild |

Why A: the static part (the résumé, the prose of an article) is a document,
and Go draws documents well. The interactive part (vim chrome, salary range,
chat, article widgets) is a set of islands, and they get a real
framework. Svelte 5, because claude-review's page rules, eslint layers and
`lib/ui` library carry over almost whole.

Revisit B if most of a page comes to be drawn in the browser. Nothing in the
roadmap points there.

templ replaces `html/template`: components become typed Go functions the
compiler checks. That removes the one struct every template shares
(`web.Page`), string comparisons in templates, and template errors that
surface only at run time. If templ is ever dropped, the fallback is
`html/template` with a check of its parse trees, at the cost of weaker types.

### 5.2 Back end — **Proposed**

Each layer builds only on those below it; `depguard` refuses the rest.

```
cmd/paranin            composition root: config, wiring; commands serve | check | render
  hosting/web          net/http: negotiation, CSP, limits, SSE for the chat
  hosting/ssh          wish
    brain              phases 3, 4, 6: Claude API, cost caps; takes a Grounding, streams markdown
    face/html          templ pages and the component library
    face/text          markdown, the man page for curl, llms.txt; also the brain's Grounding
    face/feed          Atom, sitemap
    face/tui           Bubble Tea
    face/print         Typst: cv.pdf, the tailored CV, social cards as PNG
      compose          content → documents, and the records no person reads (ADR-0003)
      routes           every address, written from a document's link target in each format
      inbox            contact messages: delivery to Telegram
        document       the face-neutral model (ADR-0003)
        content        the facts: Profile, Project, Post, Release; load from a folder, check, snapshot
        look           design tokens (generated)
```

- **Content on disk.** `content.Store` swaps whole snapshots and keeps the
  last good one; the composition root reloads it on SIGHUP and on a change in
  the folder once the changes settle (a debounce, never a reload per file),
  and logs a refusal (ADR-0003). `paranin check ./content` runs the same
  loader, so CI refuses what the server would refuse. No copy of the content is embedded:
  tests read `testdata/`.
- **One document model** (D19, ADR-0003). Every face draws the résumé and
  the posts from `document`, the HTML page included, each block in a layout
  of its own. This makes a behaviour test possible: every fact `compose`
  emits appears in every face, HTML and TUI included, so the `curl` reader
  sees what the browser reader sees.
- **The brain beside the faces.** It needs the content as text, which is
  `face/text`'s work, so the composition root builds its `Grounding` from
  `face/text` and hands it over. It streams markdown; the chat island and
  the TUI draw it with the same inline renderer the faces use. Phase 6
  calls the same brain. Its threat model (cost abuse through SSE, prompt
  injection) gets its own section before any phase 3 code.
- **Dates and levels are types.** A role's end is a month or "current", not
  the string `"present"`; health levels are one enumeration that every face
  names.
- **Plain text by contract.** `Accept` and a `.txt` route, with
  `Vary: Accept` alone (D21); negotiated routes send
  `Cache-Control: no-cache` until the contract is settled in `WEB.md`.
- **Typst replaces headless Chrome** (Decided, D10) for `cv.pdf`, the tailored CV (phase 5)
  and a social card per page and post. One tool makes PDF and PNG from
  markup, with no browser. It runs as a program beside ours, so the runtime
  image and CI carry a pinned `typst` and the fonts. Cards and `cv.pdf` are
  drawn when a snapshot loads, into a disk cache, by a worker, never inline
  in the reload. A tailored CV is drawn on request by a bounded worker with
  a per-IP limit. The `pdftotext` check from `scripts/build-assets.sh`
  becomes a test.
- **Runtime state.** The contact form delivers to Telegram and appends to a
  file until the brain arrives. With the brain comes SQLite
  (`modernc.org/sqlite`, pure Go) for chat sessions, cost caps and
  counters; messages move into it then. Content stays files in git.
- **Explicit dependencies.** Constructors take parameters; no
  package-level mutable state. Time is the `time` package's, and tests set
  it inside `testing/synctest` (D15). Errors have kinds, and one table
  turns a kind into an HTTP status. JSON-LD is typed structs.

### 5.3 Front end — **Decided** where marked, the rest **Proposed**

A pnpm package `web/`: Vite, TypeScript strict with `noUncheckedIndexedAccess`
and `exactOptionalPropertyTypes`, Svelte 5.

```
web/src/
  tokens/        tokens.ts, tokens.css: generated from one source
  domain/        pure TS, no DOM: vim motions and counts, the :command parser, search, the salary range
  dom/           what the page does to the document: find by data-part, focus, scroll
  ui/            the Svelte library: Button, Field, Range, Chip, Statusline…
  controllers/   behaviour over server markup: vim, theme, hint, form submit
  islands/       what runs each island's component in its data-island, a post's widgets first; the site's own: VimChrome, SalaryRange, Chat
  kit/           the widget kit for articles: Stage (canvas, DPR, resize, reduced motion), play/pause/step, sliders
  widgets/       the articles' widgets, named for the whole site, as a post's block names one: PageCache.svelte
  entries/       site.ts, holding the island registry: every widget's code by a lazy glob of widgets/
```

- eslint holds the layers: `domain` knows no DOM, `ui` knows no page, the
  articles' widgets in `widgets/` build on `kit` and `ui` alone.
- Vim mode becomes a pure state machine in `domain` with a `KEYMAP` table,
  tested in vitest with no browser, and a Svelte island for its chrome.
  `CHROME_HTML` goes.
- The client data's types are generated from the Go structs (tygo); a hook
  fails when they drift.
- **One build, one manifest** (Decided, D22). A single Vite build takes the
  site's entries and every post's widgets and writes hashed files and one
  manifest into the release folder (§5.4). Go serves every front-end file
  from that folder; the binary embeds none. Loading a snapshot checks that
  the manifest holds every entry the pages name. A change to the islands
  ships like a post, with no new binary.
- **Development uses the same path** (Decided, D22): `vite build --watch`
  writes the same files and manifest, Go serves them, and the strict CSP
  holds. No Vite dev server, whose scripts come from another origin, and no
  second way of writing script tags.
- One element, one look: templ components and Svelte components put on the
  classes of one library, `ui.css`. stylelint refuses those classes in any
  other stylesheet.

### 5.4 Articles — **Decided** where marked, the rest **Proposed**

#### One format for every face (Decided, D8)

A post is written once, in one format, and every face draws that same text:
the web page, `.md`, the man page for `curl`, the feed and the SSH TUI.

The format is CommonMark with GFM's tables and footnotes, YAML front matter
and widget blocks, and nothing else. `paranin check` refuses what falls
outside it, raw HTML first of all: a terminal cannot draw it and the CSP
would not run it. It refuses as well a heading of level 1, which the title
is, a heading that skips a level or stands inside a quote or a list's item,
where it opens no part of the post, a pause (a thematic break) anywhere but
between two of the post's own blocks, since it parts two runs of its text
(inside a quote or a list, at the body's start or end, or with nothing
after it but footnotes and links' definitions, it parts none), a footnote
no reference cites and a reference to no footnote, and the characters the
content refuses everywhere but a tab inside code, which a face expands as it draws the
code, written or as a reference (`&#27;`); each fault at its file and line.
A link leads to an absolute web address of another site, to `mailto:` a
bare mail address, or inside the site by a path from its root, as
`/blog/<slug>#<heading>`, the one form a link inside it takes: the check
refuses a relative path, an anchor alone and the site's own absolute
address, which would lead from a staging host to the live site. A path
must be the address of a page the site serves where the reader chooses
the format, and its anchor, where it names one, the anchor of a section
of that page, a heading of a post or a series' group of the list by
series, or of a widget of a post; each face writes the page's address in
its own format. Only the whole site knows its pages, so this part of the
check runs once the content loads (ADR-0003), each post's links over the
site as it stands when the post comes out, now or at its later date: a
link to a draft or a post that comes out after the one linking it leads
nowhere, one to a post out before it or with it leads somewhere. A
draft's links are checked too, over the site as its author reads it
under `paranin serve --drafts` once every post dated later is out, where
another draft's page leads somewhere (2026-11-23).
It refuses as well what no face draws yet: a link's title; a table's row
of more cells than its header, whose cells past it GFM drops; a footnote that is more than
one paragraph or cites another, since a note stands beside the line that
cites it (ADR-0003); and a footnote cited in a heading, which is no such
line. A line broken hard it refuses for good: the two spaces that most
often break one cannot be seen in the text (owner, 2026-11-23). Any kind of node outside
the one list `content` keeps is refused, so a part of the format no face
draws never passes unseen.

A picture, `![alt](margin.png "caption")`, is a figure: it stands as a
paragraph of its own among the post's blocks, in a quote too, or in a
list's item, its title the caption under it, none where it has none. An
item that opens with a block, a picture, code, a quote, a table or a
list, stands that block's first line beside its mark on every face, as
an item of words stands its first words there (owner, 2026-11-24). The
load reads it from the post's own folder, a file named by a slug and the
ending of PNG, JPEG, WebP or SVG, and refuses a picture with no alt text,
or alt text holding a mark; one among words, inside a link or in a
footnote; an address of another site, or of any folder but the post's; a
file missing, named otherwise, of another type, that does not read as
its type, or over 1 MiB, since each picture is held in memory with its
snapshot and sent whole to a phone; an SVG that names no size; and a
picture of the folder no line shows. AVIF is left out: Go reads no size
of one. Each picture's size is read as it loads, so the page keeps its
room before it comes.

How the TUI draws each kind of block: headings, paragraphs, lists and
quotes as text; a thematic break, a pause inside a part, as the page's
three stars in the middle of the text; code with its syntax colours from
the tokens; tables laid out to the terminal's width; links as OSC 8
hyperlinks with the address beside them; an image as its alt text, its
address and its caption, a line each, the first two links to it; a widget
as its caption, its alt text and a link to the live version. A post opens with `:e blog/<slug>` or `:e blog/<number>`
(Addresses, below).

#### Layout of a post

```
content/posts/2026-11-page-cache/
  post.en.md
  post.ru.md
  page-cache.svg
  revisions/2026-11-10.en.md
```

1. A widget in markdown is a fenced block, ` ```widget PageCache `, with its
   props as YAML inside, `caption` and `alt` among them. It is valid
   CommonMark and shows as code on GitHub. The fence names a widget of
   `web/src/widgets`, `PageCache.svelte`, named for the whole site, which
   any post may show; a post's folder holds none (owner, 2026-11-24), since
   eslint, run from `web/`, reads no file outside it, and a bare import of
   `svelte` resolves from no folder of `content/`. A goldmark extension of
   `content`'s turns the block into a widget's node as the post is parsed,
   and `compose` turns that into a `Widget` block of the document, its
   anchor the widget's name, which a link inside the site may name. `alt`
   is required: it is the widget in the TUI, the man page, the feed and a
   screen reader. The load refuses a widget inside a quote, a list's item
   or a footnote; a name of no component's form, a capital letter then
   letters and digits; a body that is no one mapping; `alt` or `caption`
   that is no one line of words; and props written with an anchor, an
   alias or a tag of their own, or a key that is no text or stands twice;
   each at its line.
2. Each face draws `Widget` its own way. HTML: a `<figure data-island>`
   holding the alt text, where the island runs the widget. Every other face: the
   caption, the alt text and a link to the live version.
3. **One rule for props.** Each widget writes its props' schema in valibot;
   the build turns it into JSON Schema beside the manifest, and the Go
   loader checks every post's props against it. `paranin check`, CI and the
   server refuse the same typo.

#### Front matter (Proposed)

```yaml
---
id: 12
title: The page cache
description: Why a page read twice is read once.
status: published
published: 2026-11-10T09:00+01:00
updated: 2026-11-12
tags: [kernel, caching]
discussions:
  - {site: Lobsters, address: "https://lobste.rs/s/…"}
---
```

- `id`, `title`, `description` and `status` are required, and `published`
  of every post but a draft; the rest may be left out. A field no post has
  is a fault, as in the profile.
- `title` and `description` are a line each; the description is the post's
  lead and its page's description.
- `published` is to the minute, with its zone; `updated` is a day, no
  earlier than the day the post was published, in that zone. The post's
  current version is dated by `updated`, or by the day it was published
  where it says none (Revisions).
- `tags` are slugs, none twice; each discussion is a site's name and the
  thread's `https` address (D13).
- A post's folder holds `post.en.md`, its translations, its `revisions/`
  and its pictures alone: `widgets/` is a fault that says a widget lives in
  `web/src/widgets`, and a translation into a language the site has no
  words for one that says its catalogue comes first.

#### Where posts live (Decided, D6)

In this repository, which is public. Drafts and the history of every edit
are visible there, and that is fine. Widgets are type-checked with `kit`,
and a change to `kit` rebuilds and checks the old posts. A draft is checked
like any post, so a broken draft fails the commit, not the publishing of
others.

#### Status (Decided, D7)

```yaml
status: draft | published | unlisted
published: 2026-11-10T09:00+01:00
updated: 2026-11-12
```

| Status | Seen where |
| --- | --- |
| `draft` | Locally alone: `paranin serve --drafts` serves it at its address, listed nowhere, its page marked as a draft, for now by its version reading `draft` where a post's day stands, until the owner picks a mark (DESIGN.md, PostHead); a post dated later is served so too, but marked by its day. Never served otherwise. |
| `published` | Everywhere: the page, the feed, the sitemap, `llms.txt`, the TUI, its card. |
| `unlisted` | At its address, listed nowhere: a post taken off the lists without breaking links to it. |

- **No preview** (owner, 2026-11-24): the repository is public, so a
  post's text is public before it is out, and a secret link would hide
  its address alone. A post shown to someone before it is announced is
  `unlisted`; its author reads a post not out under
  `paranin serve --drafts`.
- **Publishing on a date.** A `published` date in the future keeps the
  post, published or unlisted, a draft until then in who reads it and
  where it is listed: served to its author alone, at its address, listed
  nowhere. Its page is marked by its day, as it will be once out, not by
  the word `draft`, which a post written as a draft alone reads. The
  status is worked out from the time: a host composes the site once a
  snapshot, as it stands when a reader first meets it, and again as the
  next post dated later comes out, so the post comes out on time with no
  push and no request composes the site (ADR-0003, 2026-11-23).

#### Editing and publishing (Decided, D5; the rest Proposed)

Git is the editor; no admin screen. A typo is a commit and a push, from a
laptop or GitHub's web editor on a phone, live in a minute or two.

1. A push to `content/` or `web/` runs CI: `paranin check`; the types and
   the one Vite build, with unchanged widgets from the cache; Typst draws
   the cards of the posts that changed.
2. rsync copies the result into a new folder,
   `/srv/paranin/releases/<commit>/`. The `current` link switches to it at
   once, and the server gets SIGHUP.
3. CI reads `/version`: the server must name the new release. If it does
   not, the build fails red and the site keeps the old snapshot.
4. The last few releases stay on disk. Rolling back is a workflow that
   switches `current` back and sends SIGHUP.

The server builds nothing; it reads a finished folder. The binary is
rebuilt only when Go code changed. The CI job needs Node, a browser, Typst
and SSH to the host: "without a new binary" does not mean without a build.

#### Revisions (Decided, D24; the model Decided, 2026-11-24)

- **A revision is a frozen copy the author takes.** Before he rewrites a
  post, `paranin revise <content folder> <slug>` copies its text to
  `revisions/<date>.<lang>.md` beside it; `post.<lang>.md` stays the current
  version. A typo fixed is an edit, not a revision. Ways turned down: the
  versions read from git (a release folder holds no `.git`, and every commit
  would become a version), and stored patches (every face would have to
  apply them).
- **Addresses:** `/blog/<slug>` is the current version, `/blog/<slug>/<date>`
  an older one. A post shows the date of its version and lists the older
  ones; an older one says a newer stands and is sent with `noindex`. A post
  has no tabs, so `/blog/<slug>/…` holds its versions and its pictures
  alone, and they never clash: a picture's name ends in its type's ending,
  which no day and no format's ending is. The current version's day,
  `/blog/<slug>/<its date>`, sends its reader on to `/blog/<slug>` in the
  same face, `302`, since that day names an older version once the post
  is updated.
- **Each language has its own revisions**, as it has its own status:
  `revisions/<date>.ru.md` is a version of `post.ru.md`, which
  `paranin revise --language ru` takes, at `/ru/blog/<slug>/<date>` (built
  2026-11-24); a revision in a language the post has no text in is
  refused.
- **The model** (2026-11-24). A version is dated by the day its text last
  changed: the current one by `updated:`, or by the day the post was
  published where it says none; an older one by its file's name, the day
  its version came out. A revision is dated from the day the post was
  published on and before its current version's day, so a post with
  revisions says `updated:`. `revise` names the copy by the current
  version's day, then dates the post itself: it writes its `updated:` as
  the day it runs, in the post's own zone, in place of the day it said or
  on a line of its own after `published:`, every other byte of the file
  as its author wrote it, so the folder loads again at once and the post
  is the new version its author goes on to rewrite (owner, 2026-11-24).
  It refuses a version on the day it came out or later, in the post's own
  zone, so a post holds one version a day, and a change on that day is an
  edit of it. `updated:` is held to no day but the one the post was published on
  or after: the load reads no clock, and a post dated later may be
  updated on a day still to come; `revise` refuses such a version until
  its day has passed. A revision's file holds the version's own `title`
  and `description` as front matter, then its body as written; the post's facts, its number,
  status, tags and series, are the post's alone. The load holds it to what
  a post's text is held to, raw HTML, links and pictures among them, and
  `compose.Check` its links inside the site over the site as the post's
  are, so a frozen copy never serves a dead link. A picture a revision
  alone shows stays in the post's folder and is served with the post's.
  `revise` refuses a post not out, a draft or one dated later,
  whose version no reader read, and a text that is its newest revision's
  already. Ways weighed: the file copied whole, front matter and all, whose
  number, status and tags would be read and do nothing; the body alone,
  which loses the title and lead the version had; the copy named by the
  day it was taken, which dates a version by the day it ended. The file is
  never written over one there, and is taken away where it could not be
  written whole; one half written fails the load as the whole one does,
  until the post is dated by a later update. The post is dated only once
  its revision is written, and written whole: beside its file first, then
  renamed over it, so a reload meets it as it was or as dated. The field
  is found where the parsed front matter places it, and its day alone
  written over. Ways weighed: the front matter decoded and encoded again,
  which rewrites the author's quotes, flow lists and comments; a line
  matched by its text, which a nested field or a quoted day would fool.
- **Where versions show.** A post out, published or unlisted, lists its
  versions, newest first, its current one alone where it has no older,
  and serves each older one, the same page, its head saying under its day
  that a newer stands, the current one; a post dated later, which no
  reader read yet, lists its current version alone to its author and a
  draft none, and neither serves an older one, though their revisions
  stay for when they are out. An older version is in no list, the
  sitemap or `llms.txt`, and sends no canonical address, as a filtered
  list sends none; a link inside the site may lead to it. Its markdown
  and manual stand beside it at `.md` and `.txt`, and write that a newer
  stands; the TUI opens it by `:e blog/<slug>/<date>` and the `:` menu's
  `version` places, and the page by its version's Pick.

#### Series, tags and lists (Decided, D25; the model Proposed)

- **A series is one file**, `content/series/<name>.yaml`, its name a slug
  as a post's is: its `title`, a `line` about it, its `posts`' slugs in
  order, one at least, and under `to_come` the titles of parts announced
  and not yet out, so an author who wants a drafted part's title on the
  site keeps it there until the part is out. The order is written once, so a post put in the middle
  renumbers nothing. A post is in one series at most. A post of a series
  shows "part 3 of 11" with the parts before and after; a part to come may
  stand in the list grouped by series.
- **What a series and a post's foot list.** A part published is out; the
  parts to come are the titles under `to_come`, which the author chose to
  announce. A part not yet published, a draft or one dated later, is
  listed nowhere, so its working title stays in the repository (owner,
  2026-11-23); its own page, a draft's as its author reads it, shows its
  series, itself the part read among the parts out, so a draft of part 3
  says it is part 3. An unlisted part is listed
  nowhere, so its series leaves it out and counts on without it, and its
  own page shows no series. Outside a series a post's foot leads to the
  posts published before and after it, by the time they came out, so
  never to a draft or an unlisted post; an unlisted post's own
  foot leads to those published either side of its time, a draft's to
  those either side of its date, or of the moment it is read where it
  has none.
- **Tags** stand in a post's front matter: `tags: [payments, agents]`.
- **One list, no pages of its own for a tag or a series** (owner,
  2026-11-22): `/blog`, every post by year, newest first, with its date,
  series and tags; `/blog/series`, the same posts by series. A tag
  filters the list, the filter reversed in the list's head; a filtered
  list is sent with `noindex` and names no canonical address
  (2026-11-24): a part of the list is no copy of
  `/blog`, and a search engine reads `noindex` beside a canonical address
  as two signs that disagree. Its address is the site's first with a
  query, `/blog?tag=payments` (owner, 2026-11-23): a filter is a view of
  the list, not a page to index. A filtered list is the list by year of
  the posts the list names that carry the tag, its head the list's with
  the filter beside the blog's name; there is one for each tag of a post
  the list names, composed with the site, and the web host draws each
  the first time a reader asks for it, since few are read; every face
  draws it: the page and its markdown and manual at `/blog.md?tag=` and
  `/blog.txt?tag=`, the TUI in the list's window. One rule reads the
  query: on the list by year the tag it names; any other key, and the
  query of any other page, passed over, as a link's tracking words are;
  a tag named twice, or one no list has (a tag no post the list names
  carries), is an address of no page, `404`, as any other, quoting the
  address as it was asked for.
  Ways weighed: an empty list for an unknown filter, a page that says
  nothing found where nothing is; the filters read per request, which
  composes in a request; a series filtering the list too, which nothing
  led to and which repeated the series' group on `/blog/series` (owner,
  2026-11-24). The head's tags and a post's lead to the list
  they filter, but a draft's and an unlisted post's, which
  the lists name nowhere and whose tag may filter none; on a filtered
  list the filter leads back to the list; a tile's series leads to its
  group on the list by series. A tag and a series named alike each stand
  in the `:` menu by their kind. `llms.txt` lists no filter. The list is a
  document every face draws: the TUI gives the blog a session of its own
  (DESIGN.md, The TUI), and its posts, tags and series stand in the `:`
  menu.
- **Feeds:** Atom for the blog in each language, and one for each series
  (built 2026-11-24): the blog's at `/blog.atom`, a series' at
  `/blog/series/<name>.atom`, where no post stands, a language's in its
  folder, `/ru/blog.atom`, each holding the twenty
  newest posts the lists name, published and out, never a draft or an
  unlisted post, each with its lead and its body whole, so a feed
  stands in for a newsletter; a blog that lists none has no feed. A feed
  changes as the lists do, when a post comes out. A feed and each post keep
  one name for good, a tag URI under `site.tag_authority`, a post by its
  number, so neither the site moving nor a folder renamed shows a post as
  new. Every page of the blog names its feeds in its head, a series' first
  on a page that shows the series, and the résumé the blog's, so a reader
  pasting the site's address finds it (ADR-0003).
- **For the first post:** footnotes as sources beside the line that cites
  them, at the post's end where the screen is narrow, numbered links in a
  terminal; anchors on headings and a table of contents for a long post;
  code coloured from the tokens, with the copy chip; pictures from the
  post's folder, their alt text in a terminal.
- **Later, when the posts are many:** a search across them, reading time,
  related posts by shared tags.
- **Not built:** comments (accounts, moderation, spam), with a post's front
  matter naming where it is discussed (HN, Lobsters: D13) instead; a mail
  newsletter, which the feeds stand in for.

#### Addresses (Decided, D26; the model Proposed)

- **A post has a slug, its folder's name, and a number** in its front
  matter, `id: 12`, given once and never changed. `paranin check` refuses a
  number two posts share, a slug of digits alone, which `blog/12` would
  read as a number, and content of which two pages share an address, as a
  post slugged `series` would share the list by series'.
- **On the web** `/blog/12` answers `301` to `/blog/<slug>`, and
  `/blog/<slug>` the post as the project pages answer, its markdown and its
  manual at `.md` and `.txt` beside it (`docs/WEB.md`).
- **In the TUI** `:e blog` opens the list, `:e blog/12` and
  `:e blog/<slug>` the post, `:e blog?tag=go` the list filtered, as a
  tag's place of the list's `:` menu shows it in the list's window.
- **Over SSH** the user name is the address: `ssh blog@paran.in` opens the
  TUI on the list, `ssh blog/12@paran.in` and `ssh blog/<slug>@paran.in`
  on the post. SSH has no path, and sends no host name a server could
  route by, but anyone may connect under any name (FR-1), and OpenSSH
  passes `blog/12` as it is written. A name that names no post opens the
  résumé, as a login name does.

#### SEO and Open Graph (Decided, D27; the model Proposed)

For every page of the site, the blog's among them:

- **The head:** a title and a description for each page, which the check
  requires of a post; Open Graph and Twitter's card tags; an `og:image`,
  each page's card (ADR-0005).
- **JSON-LD:** `BlogPosting` for a post, with `datePublished`,
  `dateModified` its current version's day (Revisions), its author linked to the résumé,
  `inLanguage`, `keywords` from its tags and `isPartOf` its series;
  `BreadcrumbList`; the résumé's `ProfilePage` as it stands.
- **Languages:** `hreflang` and `x-default` between a page and its
  translations.
- **Crawling:** `robots.txt` naming `sitemap.xml`, and `llms.txt` in a
  comment, since no page links it and no `rel` names it (owner,
  2026-11-23); the sitemap with every
  page, tab and post and their `lastmod`, per language; `noindex` on
  unlisted posts, older versions and the filtered list; `301` from
  `/blog/12` and from a path with a trailing slash.
- **The markup:** one `h1`, `<article>`, `<time datetime>`, alt text
  required of every picture.
- **At hosting:** HTTPS and HSTS, the site verified in Google's and Bing's
  consoles, IndexNow for a new post, Lighthouse on its pages.
- **Open: D21 and crawlers.** A page's own address chooses its format by
  `Accept` alone, the manual winning a tie, so a crawler sending `*/*` gets
  the manual at the page's canonical address. Bing's crawler is reported to
  send `*/*`, LinkedIn's possibly too; neither documents it. To settle from
  the server's logs once it is served, then the owner's call: keep D21, or
  rank HTML first unless `text/plain` is named, which gives a bare `curl`
  the HTML.

#### Translations (Decided, D12; the model Decided, owner, 2026-11-24; built)

- **One folder per post, one file per language**; widgets shared. English
  is the main language. A post's language is its file's name,
  `post.en.md`, from the first one, and written nowhere else (owner,
  2026-11-23). A translation, `post.ru.md`, is in a language the site has
  words for, `content/i18n/ru.yaml`, a tag of two or three small letters;
  its front matter writes its own `title`, `description`, `status`,
  `published` and `updated`, and the `source` it was made from, and its
  number, tags and discussions are its original's, so the check refuses
  them there. A post's folder with no
  `post.en.md` is refused, a translation's among them.
- **Addresses:** `/blog/page-cache` for English, `/ru/blog/page-cache` for
  Russian; the slug is the folder's name. Nothing redirects by
  `Accept-Language`: the address alone picks the language, so a crawler and
  `curl` get one text per address. Every address of the blog has its
  language's folder before it, `/ru/blog/12`, `/ru/blog.md?tag=go`,
  `/ru/blog/page-cache/2026-11-10`; a post may link to a translation, which the check holds as it holds any
  link inside the site.
- **The blog in a language is a blog of its own** (2026-11-24): its list
  at `/ru/blog`, by year and by series, filtered by its tags,
  names the posts out in that language alone, in its words, with its
  feeds; a translation's series and neighbours are the posts translated
  beside it, counted among themselves. The simplest rule that holds
  everywhere: every face, a feed and a filter included, reads the blog in
  one language as it reads the English one, by the same code, and a reader
  of Russian stays in Russian. A series in another language is a file
  beside its original, `series/reading.ru.yaml`, as a post's text is,
  writing its own `title`, `line` and `to_come`; its posts are its
  original's, which the check refuses there, and a translation of no series,
  into English or into a language with no words is refused. A series with no file in a
  language stands there as its original writes it (built 2026-11-24). Ways
  weighed for the series: its words in the catalogue, which holds the
  site's words and would ask every catalogue for every series; a block per
  language inside the series' file, a second way to translate beside the
  posts' and the profile's file per language. Ways weighed for the blog:
  no `/ru/blog`, a translation's page leading to the English list, its
  tags to English filters and its feed to an English page, so a feed per
  language would name no page of its own; one list of every post, each in
  the language it is read in, which a feed per language could not follow.
- **Each translation has its own status**: the English can be published
  while the Russian is still a draft. A post's Switch offers the
  languages the lists name it in, and its own; so a translation still a
  draft, dated later or unlisted shows on no other language's page.
- **A stale translation shows** (built 2026-11-24). Its front matter names
  the text of its original it was made from, `source: en@3f9a1c0b2d4e`: the
  original's language, then SHA-256 of the original's title, lead and body
  as its file writes them, each parted by a line feed, cut to twelve
  hexadecimal digits, as git abbreviates a commit. The hash takes the words
  a translation translates alone, the text a revision keeps, so a status, a
  day, a tag or a discussion changed changes none; a typo fixed in place
  does, as gettext marks an original edited by a letter fuzzy. A
  translation with no `source`, or one not of that form, is refused. Where
  the source is not the original's text as it stands, the translation is
  stale: `paranin check` warns, naming the version it was made from where a
  revision keeps it, and the hash that stands, and exits 0, so a commit and
  CI pass and print it (the hook runs `verbose`), and `paranin serve` logs
  it at each load. The translation's current version says under its day
  that it is older than the original, a link to the original: on the page
  in the form an older version's way to the newer takes, in markdown, in
  the manual and in the TUI. The note leads where the Switch does, so a
  translation whose original no list names says nothing; an older version
  of a translation says a newer stands already, and nothing of its
  original. The author refreshes the hash by hand, writing the one the
  warning names once the translation follows the original; a new
  translation's is the one the check's refusal names. Ways weighed: the
  file's bytes whole, which marks a translation stale when the original's
  status, day or discussions change; the body alone, which misses a title
  rewritten; a command that writes the hash into the front matter, for a
  line the warning writes out already; the warning a fault, which blocks a typo fixed in the
  original until the translation follows; the warning a log line alone,
  which `check` never prints.
- **Markup:** `<link rel="alternate" hreflang>` with `x-default`, the
  English original, on each language's page of a post, the same alternates
  in the sitemap, a feed per language, `og:locale` from each catalogue's
  `locale`, `<html lang>` the page's own. The language switch lists only
  the translations that exist, each by its tag, `en` and `ru`, on the page
  and in the line the other faces write (owner, 2026-11-24).
- **The TUI reads the blog in a language by its address** (owner,
  2026-11-24), as the web does: `ssh ru/blog@paran.in`, `ssh
  ru/blog/12@paran.in` and `:e ru/blog/page-cache` open the Russian blog,
  framed in its words, its days in its months and the note of a
  translation older than its original among them. `:e` opens the same
  pages on every face, every language's, which compose lists once, and in
  the blog's folder of each language a post no list names; Tab offers
  every language's lists and posts. Since the TUI has no Switch, a post's
  other languages are places of its `:` menu, `language`, each by its tag,
  `:ru`, as its versions are by their days; the page's command line offers
  them too. The blog's session reads one language at a time, as a reader
  of the web stays in a language's folder: a post or a list of another
  language takes the other window there, the post read going where its
  head's language leads, an older version to the translation's current
  one, and closed where it has none; the list in its grouping and its
  filter where the language has that filter, and by year where not. The
  résumé, the projects' pages and the TUI's own words (the ruler, the
  keys, the menu's kinds, Neovim's messages) stay English. Ways weighed:
  `:set lang=ru`, an option Neovim lacks (it answers `E518: Unknown
  option`), which, as `:set helplang=ru` does, makes a session's state
  choose which page one name opens, where every other face lets the
  address alone choose; `:language ru`, which in Neovim sets the language
  of its own messages, which the TUI keeps English; a language read off
  the SSH user's name alone, `ssh ru@paran.in`, a second way in beside the
  address the web already writes.
- **`llms.txt`** lists the English site, whose posts' markdown leads to
  their translations.
- **Words inside widgets** live in a typed table beside the widget,
  `web/src/widgets/PageCache.messages.ts` (`satisfies Record<Lang,
  Messages>`): a missing translation of a label fails to compile.
- **Words of the site** (Published, Read in Russian, vim hints, man page
  headings) live in one catalogue per language, `content/i18n/en.yaml` and
  `ru.yaml`, used by every face, the TUI (`ssh ru/blog@…`) and
  `curl paran.in/ru/…` included. `paranin check` requires every key in
  every language. The catalogue arrives with `compose`, the first part that
  writes a word of the site, not with the first translation, so no template
  is written twice. The repository's content holds `en.yaml` alone until
  the owner's first Russian post brings `ru.yaml`; the tests' fixture holds
  both. Every word a page in a language shows comes from its catalogue or
  its content, never a face's constant (built 2026-11-24): the manual's
  own headings, its section's name, `NAME`, `DESCRIPTION` and `NOTES`, which
  the manual writes in capitals; the months of a day, `months`, twelve
  short names as a day writes them, `21 Nov 2026`, `21 нояб. 2026`, its
  order the site's; the page of no page in a language's folder, framed in
  its words, HTTP's status in English; and a count, how many posts, all a
  series' parts and those out, written in each plural form its language's
  rule gives a whole number, as CLDR names them, English `one` and `other`,
  Russian `one`, `few` and `many`:

  ```yaml
  posts_count:
    one: "{} пост"
    few: "{} поста"
    many: "{} постов"
  ```

  The check holds a count to every form its language gives a whole number
  and refuses any other. The rule is CLDR's, read by
  `golang.org/x/text/feature/plural` (ADR-0003). Ways weighed: a table of
  our own for the languages the site has, a grammar copied by hand that a
  new language would wait on code for; message catalogues (`x/text/message`,
  go-i18n), each a format of its own for a few counts. A day's form for
  Russian follows `DESIGN.md`'s short month, CLDR's abbreviated genitive,
  `15 нояб. 2026`; the month's name is the catalogue's, so the owner may
  write it whole, `ноября`, if his eye picks it.
- **The résumé in another language later**: `profile.ru.yaml` beside
  `profile.yaml`, holding the words that translate; the check refuses a key
  the main file lacks.
- **Help with translating, after phase 3**: `paranin translate <slug> --to
  ru` drafts `post.ru.md` through Claude with `status: draft` and the
  original's hash, for a person to read before it moves on.

#### What is common practice and what is ours

| Part | Common practice |
| --- | --- |
| Posts as files in git, published by CI | Hugo, Jekyll, Astro, docs-as-code |
| `releases/<id>` and a `current` link | Capistrano, Deployer |
| Reload on SIGHUP, the old state kept when the new one fails | nginx, HAProxy, Prometheus |
| `draft` and a publish date in front matter | Hugo, Jekyll |
| `unlisted` | Ghost, Medium, dev.to, YouTube |
| A language per subpath, `hreflang`, `x-default`, no redirect by `Accept-Language` | Google's own guidance |
| A file per language beside the original (`post.ru.md`) | Hugo's translation by file name |
| Typed message tables | Paraglide, typesafe-i18n |
| A translation's original by hash | ours; gettext marks a changed original "fuzzy" in the same spirit |
| One post drawn in a browser, `curl` and an SSH TUI | ours, the point of the site |

### 5.5 Design tokens — **Proposed**

One file, `design/tokens.yaml`. A Go generator writes `tokens.css`,
`look/palette.go` and `tokens.ts` (canvas widgets cannot read CSS custom
properties cheaply, and study's slides hard-code colours). A hook checks the
generated files against the source. The rules at the top of `ui.css` move to
`DESIGN.md` (2026-11-17: the design approved that day is a document of tables
and components, longer than a rule file holds); what a tool can hold
stylelint holds: colours, sizes, times and weights only from the tokens.

### 5.6 Design work — **Proposed**

2026-11-17: the owner approved the design of the résumé and a project's page,
now `DESIGN.md`, which settles the first, second and fifth items below as it
says (a role's dates stand in their own column, the page one column under
the width `phone`), and the fourth for the résumé, whose top bar holds its
five sections. Since 2026-11-18 every size of the page is in rem, none in
pixels (`DESIGN.md`).
2026-11-22: the owner approved the blog's design, on the page and in the
TUI, with the two faces made one style, now `DESIGN.md` too: the reading
layout, its second face iA Writer Duo, and the blog as a session of its own
beside the résumé's, which settle the third and fourth items. Since
2026-11-22 the page serves JetBrains Mono from `design/fonts`, the last
item for the web; Typst takes the face from there when it comes.

Each goes through a sheet of states in both themes before code:

- `ssh paran.in` in the hero, the pop-up gone or no longer covering text.
- A hierarchy after the first screen; fewer facts; dates stacked under 560px.
- The reading layout for articles with a second, proportional face, chosen
  on a real article.
- Constant navigation: résumé, projects, blog.
- The layout bugs of §3.4 and a favicon.
- JetBrains Mono served from our own server (OFL), in the repository, so
  the web and Typst use one copy.

### 5.7 Distribution — **Proposed**

A feed (Atom); `og:type=article`, `article:published_time` and JSON-LD
`BlogPosting`; a card per page and post (§5.2, Typst); analytics without
cookies on our own server (GoatCounter), which keeps `script-src 'self'`,
and a count of SSH sessions; cross-posts with a canonical link to paran.in;
links to the profiles where posts go: LinkedIn, HN, Lobsters, Reddit
(r/golang, r/commandline), dev.to with a canonical link, and X or Bluesky
for short notices (D13). First topics: the site itself (a
résumé over SSH, vim mode in a browser, a man page over `curl`) and releases
of claude-review and semantic-code-mcp. Reading posts in the TUI is a hook
few sites have.

## 6. Code quality

### 6.1 Rules for all code

Taken from the owner's rules and both sibling repos: whole names; no helper
inside a function; types, not tuples; no meaningful literal in an
expression; no HTML or styles as strings in code; no package-level mutable
state; no alias or shim; dependencies explicit; a parameter every caller
gives has no default; a function does one thing; shared code comes with its
second user.

**Decided** (D11) exceptions to whole names in Go: `err` and `ctx`, since
`error` would hide the built-in type and `context` the package; the short
words `id`, `ok`, `to`, `at`, `of`, `on`, `by`, `is`, `as`; and `_`.
Everything else is a word: `writer`, `request`, `(model Model)`, `rowIndex`,
`test *testing.T`.

As the gate reads D11 and D14 (`scripts/coderules/names.go` holds each list
once); the small words and the table of abbreviations are **Proposed**:

- A short word alone is one of `err`, `ctx`, `id`, `ok`, `to`, `at`, `of`,
  `on`, `by`, `is`, `as`; so is its plural (`ids`).
- Inside a longer name, the small words of English may stand too: `a`, `an`,
  `in`, `or`, `if`, `no`, `up`, `do`, `go` (`oneOfAKind`).
- An initialism is a word, written as Go writes it: in capitals, its plural
  with a small s, or in small letters where a name or a part of it starts
  (`ID`, `userIDs`, `idList`, `URLs`, `clientIP`, `Example_url`), with the
  digits it carries (`HTTP2Server`); and the names of things as their owners
  spell them (`ETag`, `IPv6`, `OAuth`). Mixed case is refused (`userId`,
  `Urls`, `Http2Server`). The list is staticcheck's own, copied
  whole, and a few of this trade (`CLI`, `TUI`, `PTY`, `OSC`, `PDF`). One of
  two letters stands alone only as `id` (D14).
- A digit is no letter of a word: `p99` is refused, `utf8Text` kept.
- A word cut short is refused in the singular and the plural (`seg`, `segs`,
  `cfg`, `srv`), from a table of them; one the table lacks is the
  reviewer's.

### 6.2 Gates — built in step 2 (2026-11-01) for Go and the common files

- **Go: golangci-lint v2.15.0** (`.golangci.yml`), built by pre-commit at
  that version with this machine's Go, in an environment of its own: nothing
  of it reaches `go.mod`, and a linter built with an older Go never refuses
  the language `go.mod` names. Its linters do not read tests (`run.tests:
  false`), as in both sibling projects; its formatter does. Every finding is
  listed, where golangci-lint would show one a line and three of a kind.
  - literals: `goconst` (a string of two letters or more written twice,
    arguments of a call included); numbers are coderules';
  - layers: `depguard`, one rule to each border, between layers and between
    the parts of one layer, so every import refused is written once;
  - complexity: `gocognit` (10), `nestif` (4); `cyclop` counted each case
    of a match over every kind of a closed set, which `gochecksumtype`
    asks for, so the owner dropped it and `gocognit` went from 15 to 10 to
    hold flat branching as `cyclop` did (2026-11-02);
  - construction: `gochecknoglobals` (Go has no constant slices or maps,
    so a fixed table becomes a function that returns it), `gochecknoinits`,
    `exhaustive` (a default case does not count), `exhaustruct_v5` (for our
    own types alone: a struct of another module keeps its defaults);
  - errors and resources: `errorlint`, `wrapcheck`, `forcetypeassert`,
    `contextcheck`, `bodyclose`;
  - security: `gosec`;
  - printing: `forbidigo` refuses `fmt.Print…`, `os.Stdout`, `os.Stderr`
    and the package-level loggers (`log.Default` and `slog`'s among them)
    outside the composition root and the scripts;
  - the language: `gochecksumtype`, `ireturn`, `modernize`, `predeclared`,
    `usestdlibvars`;
  - suppressions: `nolintlint`, each one naming its linter and its reason;
  - formatting: `gofumpt` with its extra rules, as one of golangci's
    formatters.

  Every setting was seen refusing a sample module that breaks it, every
  layer border included (`gochecksumtype`, `modernize` and `usestdlibvars`
  in the second review's samples, goldmark's top package in the third's).
  goldmark's renderer is refused by its import path, tried in the fourth
  review's samples.
- **Go: `scripts/coderules`**, our own analyzer on `go/analysis`, which
  refuses:
  - a name with a word in it shorter than a word or cut short (`seg`,
    `cfg`), of every kind the code declares, struct fields, imports and the
    name a type switch binds included, with one list of the short words
    kept (§6.1). It replaces `varnamelen`, which passes a short name used
    within a few lines of where it is declared and never reads a field;
  - a function literal inside a function that is given a name, however
    short, or that holds more than one simple statement; the body handed to
    a test's `Run`, `Fuzz` or `synctest.Test` is the test, and an
    `iter.Seq` its function returns is the value it makes;
  - an array of a fixed length of strings or interfaces, a record written by
    position;
  - HTML in a string, a tag with an attribute, a closing tag or one that
    closes itself, and a colour in a string, outside tests;
  - a number other than 0 and 1 outside a constant, outside tests. It
    replaces `mnd`, which passes a number assigned to a variable;
  - a function that only passes its call on, with no `//wrapper:` reason
    above it, as stint's `scripts/check_proxies.py` refuses one; a test, a
    method a type implements an interface by, and an exported getter of an
    unexported field are passed over;
  - a comment that tells what the code was (`no longer`, `renamed from`)
    or holds a date, as claude-review's `scripts/check_prose.py` refuses
    one, the phrases `scripts/history`'s;
  - a line a comment opens wider than 120 columns, counted in letters, a
    tab one. A line of code may run wider, as a literal with every field
    written does, so `lll`, which holds every line, is not used.

  Generated files are passed over. Each rule was seen to fail its test when
  broken, over three rounds of changes to it by the author and two by
  reviewers. A change that survived a round is now a case of its test, but
  for an entry of a table: no test holds a declaration.
- **Go, besides:** `go mod tidy -diff`, `go test ./...`; `govulncheck`
  v1.9.0 in CI alone, since it needs the network. `scripts/templates` runs
  templ's own commands, so its generated code is that of the version
  `go.mod` requires; templ is a `tool` of `go.mod` too, so `go tool templ`,
  an editor's language server among its uses, is the same version. The
  script checks each template against `templ fmt` in memory, since templ's
  own `-fail` rewrites the file, and its Go with `templ generate -check`.
- **Page:** `svelte-check`, strict, over the scripts and the components; eslint with `LAYERS`,
  `id-length`, `NO_NESTED_HELPERS`, `no-magic-numbers`, complexity 10, and
  `innerHTML` and `getElementById` refused; stylelint as in §5.5; knip;
  Prettier; vitest; Playwright.
- **Generated files:** templ, tokens, tygo and the widgets' JSON Schema each
  checked against their source.
- **Common:** pre-commit as the one list (`.pre-commit-config.yaml`), run
  again by CI on what a commit carries (`.github/workflows/gates.yml`), typos,
  the standard file checks with `detect-private-key`, Conventional Commits,
  `check_instruction_files.py`; on the message, `scripts/commitmessage`,
  stint's check: a subject of 72 letters at most with no phrase of
  `scripts/history`, and 5 lines of why.

### 6.3 Tests — **Proposed**

One behaviour per test, at the cheapest level, written first and seen
failing. No branch but the assertion. Cases of one rule in one table test.
No test of a look. Tests run on a fixture profile in `testdata/`; one test
checks the real profile loads. Text faces are checked against golden files,
since their text is what a `curl` reader gets. The TUI is tested at the
model level, the browser with Playwright.

## 7. Instruction files — **Proposed**

`AGENTS.md` (with `CLAUDE.md` as its symlink): commands, how we work,
only better, rules by area. Under `.claude/rules/`: `code.md`, `go.md`,
`tui.md`, `page.md`, `design.md`, `tests.md`, `prose.md`, `scripts.md`,
`instructions.md`. Under `.claude/agents/`: the reviewer and the page's
reviewer. The design's reference, `DESIGN.md`, stands beside `PRODUCT.md`;
it is no instruction file, and is read by the section in hand.

## 8. Order of work — **Proposed**

Old code is moved onto the new construction, not cleaned in place: cleaning
`site.js` or the `html/template` pages only to throw them out is twice the
work. Each step is its own chunk, starts on Pavel's word, and ends with an
independent reviewer's verdict.

Go's `./...` and every gate read the working tree, not the index, so old
code left in the module would fail the gates of every new commit. Step 0
moves it out first.

0. **Clear the ground.** Empty the stale index (`git reset`: the working
   files stay, the older staged versions go). Move the old tree into
   `_old/`, which `./...` skips and `.gitignore` lists, to be read while
   its replacements are written and deleted piece by piece. `.gitignore`
   also lists `/build/`, where every build writes: the Vite output, Typst's
   PDF and cards, a local release folder.
1. Instruction files, `TODO.md`, `CHANGELOG.md`.
2. Gates: pre-commit, golangci and its config, `scripts/coderules`, typos,
   Conventional Commits. The page's gates arrive with `web/`.
3. ADRs: 0002 who draws the HTML; 0003 the content store and the document
   model; 0004 the publishing pipeline and the one front-end build; 0005
   Typst.
4. The first commits, each clean and every gate green: first the
   instruction files, the gates and the documents; then `content` with
   `paranin check` and the real profile; then `document` and `compose`,
   which makes the résumé their first user (D18). One commit was the plan;
   small ones follow "Small chunks" of `AGENTS.md`.
5. The faces one at a time, each its own commit, with its host, since a face
   has no reader without one. `face/text` comes first, with `routes` and
   `hosting/web` serving its formats to `curl`. Then `face/tui` with
   `hosting/ssh`, and with `look` and the token generator from
   `design/tokens.yaml`: it is the first face that needs a colour (D18).
   Then `face/html` on templ, served by the same `hosting/web`.
6. `web/`: `domain`, `ui`, then the islands one at a time.
7. A security audit of the whole site, before anything is served.
8. The blog, agreed with Pavel first, how it works and what it holds, then
   built before hosting: posts in the one format, their status, publishing
   on a date, `kit`, the first widget from study, the
   feed, the cards, posts in the TUI. The site's word catalogue arrives
   with `compose` in step 4; a second language is an addition, not a step.
9. The content, changed together with Pavel, the profile's placeholders
   filled among it: they block any promotion.
10. Hosting, last: SEC-6 and SEC-15; `inbox` with delivery to Telegram;
    `face/print` on Typst; the `Dockerfile` with both ports and
    `PARANIN_HTTP_ADDR`; CI; deploy.
11. Phase 3: the brain's threat model, SQLite, `brain`, the chat island,
    then the chat in the TUI.

Steps 7 to 10 stand in Pavel's order of 2026-11-21: hosting last.

## 9. Open questions

None. Pavel answered the last five on 2026-11-01: D9 to D13.

## 10. Reviews

- **2026-11-01, author's two passes.** Every claim checked against the code;
  three counts corrected (`"/p/"`, `innerHTML`, `TECH_WORDS`), the health
  levels found, and a step in §8 found that left the old binary unbuilt.
- **2026-11-01, Fable.** An independent design first, then a review of this
  document. It confirmed 31 citations and 6 counts and corrected one claim
  (§3.1.2). Taken: the old tree leaves the module before the first commit;
  development without Vite's dev server; the brain beside the faces;
  `varnamelen`'s settings and struct fields in `coderules`; one props rule
  in Go; one front-end build and manifest; golangci as a pinned binary;
  GoatCounter alone; SQLite with the brain; `hosting/web` with
  `face/html`; Typst at snapshot time; tests on a fixture profile; typed
  JSON-LD; plain text by contract; phase 6 named. Not taken: stills drawn
  without a browser for canvas widgets (two ways to make a still where
  one does).
- **2026-11-01, Pavel's answers.** The repository is public and posts being
  visible in it is fine (D6), which settles question 3 and removes a wrong
  line of the earlier draft: drafts "on a branch" would have been public
  too. Four statuses with `unlisted` (D7). One format for every face, SSH
  included (D8), which settles question 7.
- **2026-11-01, Pavel's last answers.** Option A (D9), Typst (D10), `err`
  and `ctx` alone (D11), English first with translations later (D12), and
  the channels (D13). No question of this plan is open.
- **2026-11-01, steps 0 to 3.** The index emptied and the first version
  moved out of the module (step 0); the instruction files (step 1); the gates
  (step 2), where trying them changed the plan: names are held by
  `scripts/coderules` in place of `varnamelen`, a lambda given a name is a
  helper, golangci-lint is built by pre-commit with this machine's Go, and
  `exhaustruct` is `exhaustruct_v5`; ADR-0002 to ADR-0005 (step 3).
- **2026-11-01, the first independent review of steps 0 to 3.** Fourteen
  findings, all answered: the format of tests held, the first version
  archived, names and helpers read as Go writes them, borders inside a layer,
  numbers and colours in coderules, five more linters, CI and SECURITY.md,
  the instruction files split by area, a reviewer agent.
- **2026-11-01, the second and third reviews,** by that reviewer
  (`.claude/agents/reviewer.md`): initialisms as Go spells them, plurals of
  abbreviations, the inbox placed, nothing outside internal, cases for every
  behaviour a rule's documentation names. Pavel's word on initialisms of two
  letters is D14.
- **2026-11-01, D15 and the fourth review.** Pavel's word on time: the
  `time` package in code, `synctest` in tests, no `Clock` (D15). The fourth
  review found the ciphers' names refused (`ed25519Key`), two holes at the
  edge of internal and stale lines here; each answered. The first commit
  became four (§8, step 4).
- **2026-11-01, D16 and D17, and two reviews of the content's answer.**
  Pavel's word: skills nested under their groups (D16); the links' keys and
  the slider's steps out of the content, what takes their place left to the
  author (D17). The reviews asked D17 to name what takes their place, the
  levels written once and held to every value by exhaustive's map check, a
  quote in every fault of a written value, a fault built from its format so
  vet sees a value missing, and SHA and UTF16 among the initialisms; each
  answered.
- **2026-11-01, D18 and D19, and the review of ADR-0003.** Pavel's word on
  the order of work and the document: nothing lands before its user (D18),
  and one document every face draws, with the résumé's own kinds (D19). The
  store is accepted. On two reviews' findings, ADR-0003 gave the document a
  head, nested its spans, named the project pages, the index and the extras,
  sent the islands' text through documents and the rest through typed props
  and `compose`'s record, moved `routes` to `face/text` and `face/text` to
  `hosting/web`, and gave a link a format and an anchor.
- **2026-11-01, D20, and `document` with `compose`.** Pavel accepted the
  layers (D20). The résumé became a document, and the site's word catalogue
  arrived with it, as `compose` writes the first word of the site (the
  author's choice, under D18). The review asked for the snapshot built whole
  again, each word's key written once, a test of where each link leads, the
  kinds of month a closed set on both sides, a card's page of its own, the
  marks read in every text that becomes spans, and phrases as a plain list;
  each answered.
- **2026-11-01, the first face and its host.** `routes`, the markdown of
  `face/text` for the résumé and each project's page, `hosting/web` serving
  each page at its address and drawing a snapshot's pages once, and
  `paranin serve`, which reloads on SIGHUP. The head's page came back with
  its first user, `hosting/web`. depguard now keeps `content` out of every
  face and `routes` to `document`. The review asked for markdown that never
  reads text as markup, code spans CommonMark keeps whole, the word for a
  role still held written as the site writes it, a project's author and
  recording kept, a server that listens before it says so and gives the
  signals back once stopping, and a test of the signals; each answered. What
  the first version gave `curl` that waits for its chunk is in TODO.
- **2026-11-01, the manual page.** `face/text` writes a document as a
  manual page too, at `.txt`; a page's own address (`/`, `/p/<slug>`) sends
  markdown to whoever asks for it and the manual to everyone else, `curl`
  among them, with `Vary: Accept` alone (§5.2). Pavel left to the author
  whether HTML brings telling `curl` from a browser by its agent: it does
  not (D21), since a browser ranks HTML first and `curl` asks for anything.
- **2026-11-01, the index and the text 404.** `compose` makes the site's
  index, and `face/text` writes it as markdown, served at `llms.txt` as
  plain text: a title, a quote of the headline, the résumé and each project's
  page listed under the site's word for it. The document gained the index's
  page and a block of contents for it (Pavel's word); a document's lead is a
  quote in markdown on every page, which is how llms.txt opens. An address of
  no page is answered in text that quotes the path and names the command
  that reads the site.
- **2026-11-01, the TUI over SSH, the first of its chunks.** Pavel's word:
  the Charm modules (Bubble Tea, Lip Gloss, wish, colorprofile, x/ansi), and
  no deploy until every part is built. `design/tokens.yaml` holds the
  colours the TUI uses, each with a fallback of sixteen; `scripts/tokens`
  writes them into `internal/look/palette.go`, blends and the nearest of 256
  worked out there, and its hook checks the file. `tokens.css` and
  `tokens.ts` come with their users (D18). `face/tui` draws the résumé as
  tabs, the head first and then each section, and `hosting/ssh` serves it
  as a kiosk beside HTTP from `paranin serve`. What both hosts draw from a
  snapshot is worked out once for each snapshot, by `content.Derived`.
  wish v2.0.5 asks for Go 1.26.8, which `go.mod` now names; the SSH tests'
  client is `golang.org/x/crypto/ssh`, already in the graph through wish.
- **2026-11-01, search and counts in the TUI.** A slash searches the tab,
  whatever the case; Enter marks every match and goes to the first at or
  after the page, n and N go round them, and the command line, a row of its
  own as Neovim's, says which of how many or E486. A number before a key
  counts it; one typed alone chooses its tab after a pause. The keys of
  normal mode are a table of actions, and the tokens gained `error`.
- **2026-11-21, the security audit (§8, step 7).** Two passes, the author's
  and an independent auditor's, each probing a server under a memory cap.
  Fourteen findings: one critical, four high, two medium, five low and two
  to note. Seven were bounds an SSH session lacked: a terminal's size, a
  paste never closed, the environment, the sessions of one connection and
  the line typed, each a way for one client to stop or hold the whole
  process, the web with it, and a silent connection and a long user name,
  each held or logged without end. The rest: control characters in the
  content, the host key's default in the folder the server starts in, a
  request's head read up to Go's 1 MB, `golang.org/x/net` flagged by
  `govulncheck`, CI's actions pinned by tag, the release's watch misled when
  the server runs out of files, and the security documents behind the code.
  Each was mended, the watch with the SSH bounds that keep files free; what
  stays open, SEC-6 first, is in `TODO.md`, and what the parts to come must
  meet is REQUIREMENTS §5.5.
