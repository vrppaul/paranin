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
| D7 | A post is `draft`, `preview`, `published` or `unlisted` (§5.4). | 2026-11-01 |
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

The paths below are the first version's. It is not in git: it waits in
_old/tree on the owner's machine, archived beside the repository.

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
| "How we work": small chunks, understand before writing, tests hold behaviour, review before presenting, ask first | both | `AGENTS.md` |
| "Only better": construction first, simplest first, root cause first, facts not adjectives, the verdict is somebody else's, a regression found becomes a gate, what can be named as worse is not committed | both | `AGENTS.md` |
| `scripts/check_instruction_files.py`, taken whole | claude-review, as livereel took it | gate |
| Layers held by a linter (`LAYERS`, import-linter) | both | `depguard`, eslint `LAYERS` |
| Short names, nested helpers and numbers refused by a linter (`id-length`, `NO_NESTED_HELPERS`, `no-magic-numbers`, complexity 10) | both | golangci, `scripts/coderules`, eslint |
| Colours only in `tokens.css`; sizes off one scale; motion only in named rules | claude-review, livereel's stage | stylelint |
| A component library: every kind of element in one place, refused elsewhere | claude-review `lib/ui` | `face/html/ui`, `web/src/ui` |
| A script that finds elements by `data-part` and writes no style, only custom properties | livereel `stage.js` | `web/src/dom`, `controllers` |
| Types for what the server sends generated from the server, checked by a hook | claude-review `api:check` | tygo |
| Design rules from the owner's canvas; a sheet of states in both themes before a design is built | claude-review `design*.md` | `.claude/rules/design.md` |
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
job matcher, chat, article widgets) is a set of islands, and they get a real
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

### 5.3 Front end — **Proposed**

A pnpm package `web/`: Vite, TypeScript strict with `noUncheckedIndexedAccess`
and `exactOptionalPropertyTypes`, Svelte 5.

```
web/src/
  tokens/        tokens.ts, tokens.css: generated from one source
  domain/        pure TS, no DOM: vim motions and counts, the :command parser, search, the matcher, the salary range
  dom/           what the page does to the document: find by data-part, focus, scroll
  ui/            the Svelte library: Button, Field, Range, Chip, Statusline…
  controllers/   behaviour over server markup: vim, theme, hint, form submit
  islands/       mounted into data-island: VimChrome, SalaryRange, Matcher, Chat
  kit/           the widget kit for articles: Stage (canvas, DPR, resize, reduced motion), play/pause/step, sliders
  entries/       site.ts, post.ts: the island registry
```

- eslint holds the layers: `domain` knows no DOM, `ui` knows no page, an
  article's widgets build on `kit` and `ui` alone.
- Vim mode becomes a pure state machine in `domain` with a `KEYMAP` table,
  tested in vitest with no browser, and a Svelte island for its chrome.
  `CHROME_HTML` goes.
- The job matcher stays in the browser, so the text of a job description
  never reaches the server. Its vocabulary moves into the content and comes
  to the page with the client data, a change to the contract in `WEB.md`.
- The client data's types are generated from the Go structs (tygo); a hook
  fails when they drift.
- **One build, one manifest.** A single Vite build takes the site's entries
  and every post's widgets and writes hashed files and one manifest into the
  release folder (§5.4). Go serves every front-end file from that folder;
  the binary embeds none. Loading a snapshot checks that the manifest holds
  every entry the pages name. A change to the islands ships like a post,
  with no new binary.
- **Development uses the same path**: `vite build --watch` writes the same
  files and manifest, Go serves them, and the strict CSP holds. No Vite dev
  server, whose scripts come from another origin, and no second way of
  writing script tags.
- One element, one look: templ components and Svelte components put on the
  classes of one library, `ui.css`. stylelint refuses those classes in any
  other stylesheet.

### 5.4 Articles — **Decided** where marked, the rest **Proposed**

#### One format for every face (Decided, D8)

A post is written once, in one format, and every face draws that same text:
the web page, `.md`, the man page for `curl`, the feed and the SSH TUI.

The format is CommonMark with GFM tables, YAML front matter and widget
blocks, and nothing else. `paranin check` refuses what falls outside it, raw
HTML first of all: a terminal cannot draw it and the CSP would not run it.

How the TUI draws each kind of block: headings, paragraphs, lists and
quotes as text; code with its syntax colours from the tokens; tables laid
out to the terminal's width; links as OSC 8 hyperlinks with the address
beside them; an image as its alt text and its address; a widget as its
caption, its alt text and a link to the live version. A post opens with
`:e <slug>`.

#### Layout of a post

```
content/posts/2026-11-page-cache/
  post.en.md
  post.ru.md
  widgets/PageCache.svelte
  widgets/PageCache.messages.ts
```

1. A widget in markdown is a fenced block, ` ```widget PageCache `, with its
   props as YAML inside, `caption` and `alt` among them. It is valid
   CommonMark and shows as code on GitHub. A goldmark extension turns it
   into a `Widget` node of the document. `alt` is required: it is the
   widget in the TUI, the man page, the feed and a screen reader.
2. Each face draws `Widget` its own way. HTML: a `<figure data-island>`
   holding a still image the island brings to life. Every other face: the
   caption, the alt text and a link to the live version.
3. **One rule for props.** Each widget writes its props' schema in valibot;
   the build turns it into JSON Schema beside the manifest, and the Go
   loader checks every post's props against it. `paranin check`, CI and the
   server refuse the same typo.
4. Playwright takes every still image in CI, canvas and DOM widgets alike:
   one way to make them. For LinkedIn and X, livereel films the widget.

#### Where posts live (Decided, D6)

In this repository, which is public. Drafts and the history of every edit
are visible there, and that is fine. Widgets are type-checked with `kit`,
and a change to `kit` rebuilds and checks the old posts. A draft is checked
like any post, so a broken draft fails the commit, not the publishing of
others.

#### Status (Decided, D7)

```yaml
status: draft | preview | published | unlisted
published: 2026-11-10T09:00+01:00
updated: 2026-11-12
```

| Status | Seen where |
| --- | --- |
| `draft` | Locally alone: `paranin serve --drafts` marks the page as a draft. Never shipped. |
| `preview` | On the server, at a secret link alone. Not in the feed, the sitemap, `llms.txt`, the lists or the TUI; sent with `noindex`. |
| `published` | Everywhere: the page, the feed, the sitemap, `llms.txt`, the TUI, its card. |
| `unlisted` | At its address, listed nowhere: a post taken off the lists without breaking links to it. |

- **The preview link is not in git.** It is `/preview/<token>/<slug>`, the
  token an HMAC of the slug under the server's secret. `paranin
  preview-link <slug>` prints it on the owner's machine, which holds the
  same secret. CI never sees the secret and never prints a link: the logs
  of a public repository are public too. A new secret ends every old
  preview link at once.
- **Publishing on a date.** A `published` date in the future keeps the post
  a preview until then. The status is worked out per request from the
  time, with the feed and the sitemap cached per snapshot and minute, so
  the post comes out on time with no push.

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

#### Translations (Decided: later, D12; the model Proposed)

- **One folder per post, one file per language**; widgets shared. English
  is the main language and every post carries `lang` from the first one.
- **Addresses:** `/blog/page-cache` for English, `/ru/blog/page-cache` for
  Russian; the slug is the folder's name. Nothing redirects by
  `Accept-Language`: the address alone picks the language, so a crawler and
  `curl` get one text per address.
- **Each translation has its own status**: the English can be published
  while the Russian is still a preview.
- **A stale translation shows.** Its front matter names the original it was
  made from, `source: en@<hash of the file>`. When the original changes,
  `paranin check` warns, and the page says the translation is older than the
  original.
- **Markup:** `<link rel="alternate" hreflang>` with `x-default`, the same
  alternates in the sitemap, a feed per language, `og:locale`. The language
  switch lists only the translations that exist.
- **Words inside widgets** live in a typed table,
  `PageCache.messages.ts` (`satisfies Record<Lang, Messages>`): a missing
  translation of a label fails to compile.
- **Words of the site** (Published, Read in Russian, vim hints, man page
  headings) live in one catalogue per language, `content/i18n/en.yaml` and
  `ru.yaml`, used by every face, the TUI (`:set lang=ru`) and
  `curl paran.in/ru/…` included. `paranin check` requires every key in
  every language. The catalogue arrives with `compose`, the first part that
  writes a word of the site, not with the first translation, so no template
  is written twice.
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
| A preview link with a signed token | Ghost, Next.js draft mode, headless CMSs |
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
`.claude/rules/design.md`; what a tool can hold (one radius, one line weight,
sizes in rows and ch, colours only from tokens) stylelint holds.

### 5.6 Design work — **Proposed**

Each goes through a sheet of states in both themes before code:

- `ssh paran.in` in the hero, the pop-up gone or no longer covering text.
- A hierarchy after the first screen; fewer facts; dates stacked under 560px.
- The reading layout for articles with a second, proportional face, chosen
  on a real article.
- Constant navigation: résumé, projects, blog.
- The layout bugs of §3.4 and a favicon.
- JetBrains Mono served from our own server (OFL), in the repository, so
  the web, Typst and the stills use one copy.

### 5.7 Distribution — **Proposed**

A feed (Atom); `og:type=article`, `article:published_time` and JSON-LD
`BlogPosting`; a card per page and post (§5.2, Typst); analytics without
cookies on our own server (GoatCounter), which keeps `script-src 'self'`,
and a count of SSH sessions; cross-posts with a canonical link to paran.in;
links to the profiles where posts go: LinkedIn, HN, Lobsters, Reddit
(r/golang, r/commandline), dev.to with a canonical link, and X or Bluesky
for short notices (D13). LinkedIn runs no script, so each post with a
widget brings its still or a livereel film. First topics: the site itself (a
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
  - complexity: `cyclop` (10), `gocognit` (15), `nestif` (4);
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
    replaces `mnd`, which passes a number assigned to a variable.

  Generated files are passed over. Each rule was seen to fail its test when
  broken, over three rounds of changes to it by the author and two by
  reviewers. A change that survived a round is now a case of its test, but
  for an entry of a table: no test holds a declaration.
- **Go, besides:** `go mod tidy -diff`, `go test ./...`; `govulncheck`
  v1.9.0 in CI alone, since it needs the network. templ, when it comes,
  through `go tool` (its generated code must match its version), with
  `templ fmt`.
- **Page:** `tsc` strict; `svelte-check`; eslint with `LAYERS`,
  `id-length`, `NO_NESTED_HELPERS`, `no-magic-numbers`, complexity 10, and
  `innerHTML` and `getElementById` refused; stylelint as in §5.5; knip;
  Prettier; vitest; Playwright.
- **Generated files:** templ, tokens, tygo and the widgets' JSON Schema each
  checked against their source.
- **Common:** pre-commit as the one list (`.pre-commit-config.yaml`), run
  again by CI on what a commit carries (`.github/workflows/gates.yml`), typos,
  the standard file checks with `detect-private-key`, Conventional Commits,
  `check_instruction_files.py`.

### 6.3 Tests — **Proposed**

One behaviour per test, at the cheapest level, written first and seen
failing. No branch but the assertion. Cases of one rule in one table test.
No test of a look. Tests run on a fixture profile in `testdata/`; one test
checks the real profile loads. Text faces are checked against golden files,
since their text is what a `curl` reader gets. The TUI is tested at the
model level, the browser with Playwright.

## 7. Instruction files — **Proposed**

`AGENTS.md` (with `CLAUDE.md` as its symlink): commands, how we work,
layers, rules for all code, only better, rules by area. Under
`.claude/rules/`: `go.md`, `tui.md`, `page.md`, `design.md`, `tests.md`,
`scripts.md`, `instructions.md`.

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
7. Hosting: SEC-6 and SEC-15; `inbox` with delivery to Telegram;
   `face/print` on Typst; the `Dockerfile` with both ports and
   `PARANIN_HTTP_ADDR`; CI; deploy.
8. The blog: posts in the one format, their status and preview links,
   publishing on a date, `kit`, the first widget from study, the feed,
   the cards, posts in the TUI. The site's word catalogue arrives with
   `compose` in step 4; a second language is an addition, not a step.
9. Phase 3: the brain's threat model, SQLite, `brain`, the chat island,
   then the chat in the TUI.

Filling the profile's placeholders needs no code and blocks any promotion;
it can happen at any time.

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
