# Architecture

## ADR-0001 — Language and TUI/SSH stack

**Status:** accepted (2026-06-02); amended 2026-11-01, when the TUI was rebuilt: the log is slog's, and the SSH
layer is `charm.land/ssh`, where Charm moved it.

### Context

`paranin` is, in its first phase, a personal résumé served as an interactive
TUI over SSH. A secondary, explicit goal is for the author to **learn Go**, so
the whole project is to be written in Go where reasonable.

### Decision

Use **Go 1.26** with the **Charm v2** ecosystem:

| Library | Import path | Version (at decision) | Role |
| --- | --- | --- | --- |
| Wish | `charm.land/wish/v2` | v2.x | SSH server + middleware |
| Bubble Tea | `charm.land/bubbletea/v2` | v2.0.6 | TUI framework (The Elm Architecture) |
| Lip Gloss | `charm.land/lipgloss/v2` | v2.0.3 | styling & layout |
| Bubbles | `charm.land/bubbles/v2` | v2.0.0 | reusable components |
| Log | `log/slog` | the standard library | structured logging, the one log of every part |
| ssh | `charm.land/ssh` | pinned | low-level SSH session layer |

> **Important:** Charm v2 moved its module paths from
> `github.com/charmbracelet/...` to **`charm.land/...`**. Most tutorials online
> are v1 and will not compile against v2 (notably `View()` returns `tea.View`,
> not `string`; alt-screen is a `View` field, not a program option).

### Rationale

- **Wish + Bubble Tea are built by the same team to compose** — Wish exists
  specifically to serve Bubble Tea programs over SSH. This is what Charm's own
  production apps (Soft Serve, `terminal.shop`) use.
- **Secure-by-default for our use case:** a Wish server with only the Bubble Tea
  middleware serves *the TUI and nothing else* — no shell, no exec, no SFTP, no
  port/agent/X11 forwarding unless explicitly added. "Kiosk, not a shell."
- **Single language** keeps the later phases (web, chat, CV) in Go too.
- **Active, modern ecosystem** with a fast v2 renderer and rich components.

### Alternatives considered

- **`tview` / `gocui` / `termui`** for the TUI — mature but less momentum, and
  no first-class SSH-serving story. Rejected.
- **`gliderlabs/ssh` or `golang.org/x/crypto/ssh`** directly — lower level,
  more plumbing, no middleware. Wish wraps `charm.land/ssh` with sane
  defaults. Rejected for v1.

### Consequences

- Tied to `charm.land/*` v2 APIs (declarative `View`, richer input model).
- Must keep an eye on the still-young v2 indirect deps (`ultraviolet` renderer).

---

## ADR-0002 — Who draws the HTML

**Status:** accepted (2026-11-01). The review behind it: [FOUNDATION](FOUNDATION.md) §3 and §5.1, D9.

### Context

The first version drew its pages with `html/template` from one struct every
template shares, and ran the page's behaviour from one plain script of 1404
lines. The site will grow: articles with interactive widgets, a chat (phase
3), a CV made to a job description (phase 5). The résumé must still read
whole without JavaScript and through `curl`, and the CSP allows no script
but our own (`script-src 'self'`).

### Decision

- Go draws every page, with **templ**: typed components the compiler checks.
- What a reader works on in the browser is an **island of Svelte 5**,
  TypeScript in strict mode, mounted into server markup at a `data-island`:
  the salary range, the chat, an article's widgets.
  Behaviour over server markup that draws nothing of its own is a
  TypeScript controller, as the vim keys' command line is: the server draws
  its places, the controller shows those that fit.
- One CSS library, `ui.css`, gives templ and Svelte components the same
  classes; stylelint refuses those classes in any other stylesheet.
- The page's code lives in `web/`, a pnpm package built by Vite.

### Rationale

- The résumé and the prose of an article are documents, and every face of
  them (markdown, man page, TUI, PDF) is Go's already: Go drawing the HTML
  keeps one process, one binary and one place for the facts.
- The interactive parts are islands with a framework of their own, so they
  can grow without the page's script growing with them. Svelte, since
  claude-review's rules, eslint layers and component library carry over.
- templ removes string comparisons in templates and template errors that
  surface only at run time.

### Alternatives considered

- **SvelteKit on Node, Go as an API.** One component system for the web, but
  two runtimes in production, every page a call from Node to Go, and a proxy
  in front to send `curl` to Go. Rejected; to be weighed again only if most
  of a page comes to be drawn in the browser.
- **Astro or another static generator.** Pleasant for articles, but a second
  brain outside Go, and every post a rebuild. Rejected.
- **Plain JavaScript with JSDoc**, as livereel's stage is written. Rejected
  by D3: it does not hold up as the page grows.

### Consequences

- Two component systems; the shared CSS library and its gate keep one
  element one look.
- templ is a generator with a version to pin; its output is checked against
  its source by a hook. Should templ ever go, `html/template` with a check of
  its parse trees is the fallback.

---

## ADR-0003 — The content store and the document model

**Status:** accepted (2026-11-01): content published without a new binary (D5), one format for every face
(D8), the store, one document model every face draws (D19), and the layers (D20). The document's parts below
are the author's, on Pavel's leave to choose. [FOUNDATION](FOUNDATION.md) §5.2. Amended 2026-11-06: the head
gains a greeting and the title of its opening, for the TUI Pavel chose to redo. Amended 2026-11-07, on Pavel's
choice for a project's page in the TUI: the head gains what a project is made of and where its code lives.
Amended 2026-11-07, on Pavel's choice for the TUI's :Lazy: a skill carries its months beside how long.
Amended 2026-11-20, on Pavel's choice that each tab is a page of its own, as in the TUI: a page of the site may
name one of its tabs, shown alone.

### Context

The first version embedded `profile.yaml` in the binary and parsed it once,
so no post could change without a build (against D5). Each face laid the
content out again by hand: the HTML page, the markdown, the man page and the
TUI each picked the facts and their order, and `content` held presentation
and web routes.

### Decision

- **Content lives on disk.** `content` loads a folder (the profile, the
  posts, the word catalogues), checks it whole and makes an immutable
  snapshot. `release.Store` holds it beside the page's build (ADR-0004),
  the two loaded as one snapshot, so the pages of a snapshot name only files
  that snapshot serves; it swaps snapshots whole, in an `atomic.Pointer`,
  and keeps the last good one when a load of either part fails (amended
  2026-11-03, when the build arrived; the store was `content.Store` until
  then). A page fetched before a swap that asks for a file after it finds
  the file gone, until the files of the snapshot before are kept beside the
  current ones (TODO). The composition root calls `Reload` on SIGHUP and on
  a change in the folders once the changes settle, and logs a refusal.
  Nothing of the content is embedded; tests read fixtures.
- **`paranin check`** runs the same loader, of the content or of a whole
  release, so CI refuses what the server would refuse.
- **One document, every face.** `compose` turns content into documents: the
  résumé, a page for each project that has one, each post, the index of the
  site that `llms.txt` lists, and the extras of the TUI and the vim mode.
  Every face draws them, the HTML page included, each part in a layout of its
  own; no face picks the facts or their order again. `compose` is a function
  of a snapshot, and of the time where a post's date makes it one; who holds
  what it makes is the hosts' to say.
- **A document is a head and its blocks.** The head holds the page it is,
  where it is one; the language it is written in; its title, a subtitle, a place and a lead (the résumé's
  are the name, the title, where the author lives, the headline); on a
  project's page its tags, the language it is written in and then its
  stack, as the profile lists them; a description for the page's
  metadata; and on a project's page where its code lives, a repository
  typed as its address and the site's word for it, since every face shows
  it first. The ways to reach the author (mail, the profiles, `ssh`, the
  CV) left the head (amended 2026-11-15, the owner's word): they are facts
  of their own kind, in the résumé's last section under the site's word
  for contact, which a face finds by that kind and shows where it chooses.
  One more is the site's word, which a face shows only where it has a use
  for it, so `Phrases` leaves it out: a greeting before the title. The
  résumé carries it; the TUI and the page show it.
  The résumé has five sections, as a reader asks of its author (amended
  2026-11-15, the owner's word, after the TUI's storyboard E): about them,
  the summary and the facts; their experience; their projects; their
  skills, the groups and then the languages and the education, each a
  section nested in it; and how to reach them. Every face follows them.
  The TUI's opening shows three phrases by its form, as neofetch prints a
  system (amended 2026-11-06, with the look the owner chose): the name in
  the block font where the page has room for it and the font draws every
  letter of it, so a search finds the name only where it reads as text;
  the résumé's first section, where it holds facts, taken into the first
  tab, which its title names, its facts as a brief beside the name and
  what else it holds under the lead; and the first way to reach the
  author, the mail, at the brief's head without its label, the other ways
  on the contact section's tab alone. Under them a career's line lays out
  the roles, the entries of a role's kind in any of its top-level
  sections, and the lead comes last. The brief is found by its place in
  the document, the roles and the ways by their kind (amended 2026-11-15,
  the owner's word). The page's first screen opens from the same reading,
  which `document.BriefOf` makes for both faces: beside the name the
  mail, the title, the place and the first four facts in the profile's
  order, the author's choice of which come first; the facts after them
  stay in their section, where the brief stood (amended 2026-11-18, the
  owner's word). `document.TabsOf` parts a document into its tabs, the
  opening and a tab for each section left, the TUI's; `SlugOf` names
  each, and the ids of the page's sections (amended 2026-11-20, the
  owner's word). Each section of the résumé after the opening stands
  under its title, in a smaller block font where the page has room for
  it, so a search finds a section's title on its page only where it reads
  as text.
  A project's page opens as glow draws its README (amended 2026-11-07,
  with the look the owner chose): its name in that smaller font where the
  page has room for it, so a search finds the name on its page only where
  it reads as text, and its repository by its form, the owner and name its
  address's path gives, a link to it, without its label; the address is
  written out only where its path gives no owner and name.
- **The blocks** are a closed set (`//sumtype:decl`), and every face switches
  over all of them (gochecksumtype), so a new kind reaches every face:
  - the résumé's: a section (a title, an anchor and its blocks), a
    paragraph, a table (each skill group is a section holding one), facts
    (their kind, details or the ways to reach the author; labels and their
    values), an entry (its kind, role or school; a period, a title, a
    subtitle, a place and its points) and a card (a title, a subtitle, a
    tag, a summary, its facts and its page: a project). A kind is a field,
    not a block of its own, as the owner chose for the entry (2026-11-15):
    every face draws both kinds alike, and the one that places them apart
    reads the field;
  - the index's: contents, the pages a document leads to, each by its title
    and a line on what it is, which `llms.txt` lists;
  - the health report's: checks, a group of them under its name, each check
    at a level (OK, a warning, an error) with what it says and its advice;
  - the skill list's: skills, every skill in one list, each with its name,
    how long and, where that is a span, in months, where it was proven,
    and whether the author is still learning it (amended 2026-11-07, for
    the TUI's :Lazy the owner chose: `content` reads a duration when the
    profile loads, a span, `side`, `learning` or a gap, and refuses one it
    cannot read, and `learning` on a skill whose `learning` is not true;
    `compose` writes a span and a gap as the profile does, and `side` and
    `learning` in the site's words, as it writes `current`);
  - a widget (a name, its props, a caption and an alt text), the form and
    the salary range among them, so `compose` places them too;
  - the posts' (step 8): a list, ordered or not, whose items hold blocks; a
    quote; code with its language; a figure. A table gains a header row, and
    a post's headings nest into sections.
- **Text inside a block** is a tree of spans, a closed set too: emphasis,
  strong, a link and a highlight (`==result==`) hold spans; text, code and a
  gap (`[text]`, still to fill) are leaves. A gap may sit inside a highlight,
  as the profile has it. The profile's marks and a post's markdown become the
  same spans.
- **A period is a type**: a start and an end, each a month or a gap, and the
  end may be current; the kinds of month are a closed set too. Each face
  writes a known month its own way (`2024`, `01/2024`); the end of what still
  goes on carries the site's word for it, `present`, which every face writes
  as it is. `document` imports nothing of `content`, so `compose` carries the
  months across.
- **A link names its target**: an address outside the site, a mailbox by
  its address, which a face may part at its at and `routes` writes under
  `mailto:`, or a page of the site by its kind and slug, and the tab of it
  shown alone where it names one, with a format
  where the link names one (the CV as PDF, `index.md`, `llms.txt`) and an
  anchor where it names a place inside the page. A section's anchor is the
  one a link names. The kinds of page are `document`'s type; `routes`
  builds on `document` and writes a target's address in each format
  (`/p/claude-review`, `/p/claude-review.md`,
  `/p/claude-review/what-it-does.md`, `/experience`), so the markdown face
  links to markdown and the TUI can open the page in a buffer. A tab is
  named by its title (`document.SlugOf`).
- **The islands' text is a document's too.** The extras of the TUI and the
  vim mode (`:checkhealth`, `:Lazy`, `:hire`) are documents, which the page
  draws on the server for the island to show. What no person reads comes
  typed: a widget's props, each widget's own type (the salary range's
  numbers), and a record `compose` makes for JSON-LD
  and the sitemap's dates. No face imports `content`; depguard refuses it,
  from the first face on.
- **What lands when.** The head, the résumé's blocks and its spans (text,
  code, link, highlight, gap) land with `compose` and the résumé, and the
  site's word catalogue with them, since `compose` writes the first word of
  the site. The head's page lands with `routes`, its language (the tag of
  the words' catalogue, which a page's `lang` names) and its description
  with `face/html`, the CV among the ways with `face/print`; a section's
  anchor and a link's format and anchor with the first link that names
  one; project pages and the index with `face/text`, the first face to
  serve them; the extras with `face/tui`, and the head's greeting with its
  redesign (2026-11-06), a project's tags and repository with its page
  there (2026-11-07), the entry's and the facts' kinds and the five
  sections with the owner's word (2026-11-15); the records for JSON-LD,
  the person's and a project's, with `face/html`, and the sitemap's dates
  with `face/feed`, the `dateModified` of the résumé's `ProfilePage` with
  them; a widget with the first island that draws one, the form's salary
  range (step 10), since `face/html` came without one; the posts' blocks and spans with the blog (D18).
- **`routes`** writes every address.
- **The layers**: `content`, `document`, `look` and `assets` (the page's
  build) at the foot, side by side; `routes` on `document`; `compose` on the
  foundation, and on `routes` once it reads a post's links back into
  targets; `inbox` on the foot alone, beside both; `release`, which holds the
  content and the build as one snapshot, on the foot beside them; the faces
  and `brain` side by side on those, no face on another nor on `release`;
  `hosting` over everything but the composition root. depguard holds each
  border once its package lands (`.golangci.yml`).

### Rationale

- A snapshot swapped whole means no request sees content half loaded, and a
  bad push leaves the site as it was.
- With one document a test can hold the rule the first version kept by
  discipline: every fact of a document appears in every face. The HTML page
  differs from the TUI in layout and in its own words, never in the facts it
  shows.
- A head, typed periods and typed links keep faces from finding the name by
  position, and keep date formats and addresses out of `compose`, in the
  faces and in `routes`, which own them.

### Alternatives considered

- **A database for content.** Content is text people edit in git (D6);
  rejected. Runtime state (chat sessions, cost caps) gets SQLite with the
  brain.
- **Keeping an embedded copy as a fallback.** Two sources of the same facts;
  rejected.
- **The HTML résumé drawn from `content`'s types, `compose` giving it only
  the order of the sections**, as this ADR's first draft had it. Two paths to
  one page, and the HTML held to the facts by a test of its own; rejected.
- **The blocks of markdown alone**, the résumé written as headings, lists and
  paragraphs. The TUI's columns and the HTML layout would read the structure
  back out of text; rejected.
- **The islands' text from `content`, as the first version's client data
  carried it.** One text on two paths, and a face importing `content` that
  only a reviewer could tell from one that should not; rejected.

### Consequences

- The binary cannot start without a content folder; development points it
  at the repository's own.
- A face draws from documents and from `compose`'s records alone; the HTML
  face reads beside them the page's build (`assets`, ADR-0004), for the
  files its pages link, and chooses which entry each page links.
- A new kind of block or span is a change to every face, which the compiler
  and gochecksumtype point at.
- A month's shape is written twice, in `content` and in `document`, and so
  are a check's levels: the price of layers that stand side by side.

---

## ADR-0004 — Publishing, and the one front-end build

**Status:** accepted in part: publishing without a new binary, posts in this repository, the four statuses, one
format and translations later (D5–D8, D12) on 2026-11-01; the one front-end build on 2026-11-03 (D22). The
pipeline is proposed, to be settled with hosting (§8, step 10). [FOUNDATION](FOUNDATION.md) §5.3 and §5.4.

### Context

Posts must go live without a new binary (D5), live in this public repository
(D6), carry a status (D7) and be written once for every face, the SSH TUI
included (D8). Their widgets are code, built by the page's toolchain.

### Decision

- **One format**: CommonMark with GFM tables, YAML front matter and widget
  blocks (a fenced ` ```widget Name ` block with its props as YAML, `alt`
  required). `paranin check` refuses anything else, raw HTML first.
- **Status**: `draft` (never shipped), `preview` (a link signed by the
  server's secret, `noindex`, listed nowhere), `published`, `unlisted`. A
  `published` date in the future keeps a post a preview until then, worked
  out per request from the time (tests hold it in `testing/synctest`). The
  preview link is printed on the owner's machine; CI never sees the secret.
- **One front-end build**: one Vite build takes the site's entries and every
  post's widgets, and writes hashed files, one manifest and each widget's
  props as JSON Schema into the release folder. Go serves every front-end
  file from there and embeds none. Loading a snapshot checks the manifest
  holds every entry the pages name and every post's props against their
  schema. Development runs `vite build --watch` into the same shape, so the
  strict CSP holds there too.
- **Pipeline**: a push to `content/` or `web/` runs CI: `paranin check`, the
  types and the build, Typst's cards, rsync into `releases/<commit>/`, the
  `current` link switched, SIGHUP; CI then reads `/version` and fails if the
  server does not name the new release. Rolling back switches the link back.
- **Translations**, when they come (D12): a file per language beside the
  original, an address per language under its prefix, `hreflang`, a status
  per translation, the original's hash in each to show a stale one, and word
  catalogues every face reads, there from `compose` on.

### Rationale

- Files in git published by CI, release folders behind a link, reload on
  SIGHUP keeping the old state, signed preview links and a language per
  subpath are each common practice ([FOUNDATION](FOUNDATION.md) §5.4 names
  where). What is ours is one post drawn in a browser, `curl` and SSH.
- One build and one manifest, so Svelte and the widget kit are built once,
  and a change to the islands ships like a post.

### Alternatives considered

- **Posts in a private repository.** Not needed: drafts being public is fine
  (D6), and one repository type-checks the widgets with the kit they use.
- **An admin screen.** Git is the editor.
- **Vite's dev server.** Its scripts come from another origin, which the CSP
  refuses, and it would be a second way of writing script tags.

### Consequences

- The CI job that publishes needs Node, a browser, Typst and SSH to the
  host: no new binary does not mean no build.
- A broken draft fails the commit, since every post is checked.

---

## ADR-0005 — Typst for the PDF and the cards

**Status:** accepted (2026-11-01). [FOUNDATION](FOUNDATION.md) §5.2, D10.

### Context

The first version printed `cv.pdf` and captured `og.png` with headless
Chrome from a running server, by hand. Phase 5 makes a CV per job
description on request, and every page and post needs a card of its own.

### Decision

- **Typst** draws `cv.pdf`, the tailored CV and a social card per page and
  post, from markup `face/print` writes out of documents, as PDF and PNG.
- It runs as a program beside ours, pinned in the runtime image and in CI,
  with JetBrains Mono from the repository.
- Cards and `cv.pdf` are drawn when a snapshot loads, into a disk cache, by a
  worker, never inline in the reload. A tailored CV is drawn on request by a
  bounded worker with a per-IP limit.
- The text of the PDF is checked by a test, as `pdftotext` checked it.

### Rationale

- One tool for PDF and PNG, with no browser in production or in the build.
- Typst's markup is text Go writes from the document model like any face.

### Alternatives considered

- **Headless Chrome**, as the first version did: heavy for a request in
  production, and a browser in the path of every card. Rejected.
- **Drawing the cards with a Go image library.** Possible for cards, but the
  PDF would still need a second tool. Rejected.

### Consequences

- A second program in the runtime image, with its version to pin; a new
  version can move the layout, so its text check and a look at the PDF come
  with every bump.

---

## System overview — one source, many faces

> Of the sections from here to the deployment, the configuration's last two rows still
> describe the first version, kept on the owner's machine while their replacements land;
> the rest holds today.

```
cmd/paranin
  hosting/web · hosting/ssh
    face/html · face/text · face/tui
      compose · routes · release
        document · content · look · assets
```

Each part builds only on those below it, as ADR-0003 draws them; what each
does is in the project layout below. A face draws documents and never reads
the content they are made from, nor the release a host hands it; `release`
stands beside `compose` and `routes`, `compose` coming to build on `routes`
once it reads a post's links, and the hosts bring them together. The parts
still to come land with their steps of [FOUNDATION](FOUNDATION.md) §8, where
§5.2 places them: `face/feed` and `face/print` beside the faces, `inbox`
beside `compose`, the brain beside the faces.

## Phase 1 component design (SSH TUI)

```
   ssh client ───TCP:23234────▶ ┌─────────────────────────────────────┐
                                │ internal/hosting/ssh (Wish)         │
                                │  middleware chain (outer→inner):    │
                                │   logging → recover → activeterm →  │
                                │   bubbletea.Middleware(open)        │
                                └───────────────┬─────────────────────┘
                                                │ per connection
                                                ▼
                                ┌─────────────────────────────────────┐
                                │ internal/face/tui (Bubble Tea)      │
                                │  a buffer of tabs for each document │
                                │  laid out once a size, painted once │
                                │  a theme and a search, from look    │
                                └─────────────────────────────────────┘
```

**Connection flow:** Wish accepts the TCP connection, runs the middleware chain,
and the Bubble Tea middleware starts a fresh `tea.Program` bound to that
session's PTY. The program runs The Elm Architecture: `Init → Update(msg) →
View` until the user quits, closes the session or disconnects: a session's
program ends with its channel. A connection opens one session, and a second is
refused; it ends with that session, or once it has started none in its time to
start one (SEC-5). The size of the terminal a client asks for, and each size it
changes to, is bounded as the session's requests come in, before
`charm.land/ssh` reads them, so no session draws a screen larger than the bound
(SEC-7); Wish forwards each change into the program as `tea.WindowSizeMsg`.

**Per-connection isolation:** each session gets its own model instance; a panic
in one session is contained, by Bubble Tea inside the program and by the recover
middleware around it, and does not affect others or crash the server. No
session's program answers the server's own signals: the server alone ends its
sessions on a stop.

## Phase 2 component design (web)

```
   browser, curl, agent ──HTTP──▶ ┌──────────────────────────────────────┐
                                  │ internal/hosting/web (net/http)      │
                                  │  GET, HEAD at the addresses of WEB   │
                                  │  own address: Accept ranks →         │
                                  │    HTML | markdown | manual          │
                                  │  its .md, .txt; llms.txt, each built │
                                  │    file, the icon: one format each   │
                                  └──────────────────┬───────────────────┘
                                                     │ for each snapshot of
                                                     ▼   internal/release's store
                     internal/compose, internal/routes: documents, records,
                       the site's addresses; internal/assets: the build
                                                     │ handed to
                                                     ▼
                     internal/face/html (templ) · internal/face/text
```

- `cmd/paranin` runs the SSH and HTTP servers in one process over one
  `release.Store`, the content folder and the page's build, reloads it on
  `SIGHUP` and once a change in the folders holds still for a look
  (`release.Watch`), and stops both on `SIGINT` or `SIGTERM`.
- `hosting/web` composes each snapshot of the store with `compose` and
  `routes`, the first time a request meets it, and hands the documents and
  the build to the faces to draw (two requests meeting a new snapshot at
  once may each draw it, `release.Derived`). A request only reads the
  drawing; the page of no page alone is drawn for each request, since it
  quotes the path asked for.
- `face/html` draws each document with templ, in the build's stylesheet
  and running its script, with the frame round each page in the site's
  words (`compose.FrameOf`), and its record as JSON-LD, and the site's
  icon its pages link, in SVG and in PNG, drawn from the author's initial;
  `face/text` draws the markdown, the manual, `llms.txt` and its own page
  of no page. `routes` writes every address, the icon's among them.
- templ compiles the templates into Go, and no file is embedded. `web/`
  is the page's pnpm package: Vite builds it into `build/web` with a
  manifest (ADR-0004, D22); `internal/assets` reads the manifest and every
  file it names into memory, and `hosting/web` serves each at its path for
  a cache to keep for good, its name changing with what it holds. The
  site's script, `src/entries/site.ts`, is the build's one entry so far, the
  stylesheet among what it imports; `assets` works out once what a page
  includes for it, as Vite's guide to backend integration walks a manifest.
  The islands come with step 6, the CV and the cards with Typst (ADR-0005).
- Every answer tells a browser what it may do: a strict
  `Content-Security-Policy` and `nosniff`, set where every request passes.
- The contract between the server and its readers is [WEB](WEB.md): each
  address and what it answers, the headers, and what the page does in a
  browser.

## Project layout

```
cmd/paranin/main.go      # check the content, or serve it with the build: both servers, the signals, the stop
internal/hosting/ssh/    # Wish server construction, middleware, security policy
internal/hosting/web/    # HTTP: each page drawn once a snapshot, the format the reader's Accept ranks
internal/face/tui/       # Bubble Tea model: a buffer per document, its tabs laid out and painted
internal/face/html/      # templ pages: each document, its record as JSON-LD, the page of no page, the icon
internal/face/text/      # markdown, the manual for curl, llms.txt, the page of no page as text
internal/compose/        # content → documents and their records
internal/routes/         # every address of the site
internal/release/        # the content and the build as one snapshot: load, the store, what is derived once, the watch
internal/document/       # the face-neutral model
internal/content/        # the content folder: load, check
internal/look/           # the design's colours, generated from design/tokens.yaml, and its block fonts
internal/assets/         # the page's build: its manifest and files, read and checked
internal/golden/         # the golden files' helper, for tests alone
internal/fixtures/       # the shared fixtures, the content's and the build's, for tests alone
scripts/tokens/          # the generator of internal/look and the page's tokens.css, and its check
scripts/templates/       # templ's format and generated Go, checked or written
scripts/coderules/       # the code rules no linter here holds
scripts/check_instruction_files.py  # claude-review's gate of the instruction files
content/                 # the site's content: the profile, the site's words
design/tokens.yaml       # the design's colours, which internal/look and tokens.css are written from, and the page's face, text styles, lengths, times and layers, which only the page takes
web/                     # the page: Vite, its stylesheet, the tokens' CSS; built into build/web
```

Each part is held to its layer by depguard (`.golangci.yml`, `.claude/rules/go.md`).

## Configuration

MVP reads from env with sane defaults (12-factor):

| Var | Default | Meaning |
| --- | --- | --- |
| `PARANIN_SSH_ADDR` | `localhost:23234` | SSH listen address |
| `PARANIN_SSH_HOST_KEY` | `.paranin/ssh_host_ed25519_key` | host key path (persisted, gitignored; made when missing) |
| `PARANIN_HTTP_ADDR` | `localhost:8080` | HTTP listen address |
| `PARANIN_MESSAGES_PATH` | `data/messages.jsonl` | where contact messages are appended |
| `PARANIN_TRUST_PROXY` | off | `1` counts the rate limit by `X-Forwarded-For` |

No secrets in the repo. Future API keys (Phase 3+) come from env only.

## Deployment architecture (Phase 1 → public)

```
   ssh client ─────────────────────────────────▶  Hetzner VPS
                  (DNS A record for paran.in)        └ paranin (systemd)
                  Cloudflare "DNS only" for SSH         port 22 → TUI

   browser ──▶ Cloudflare (proxy/CDN/DDoS) ──▶ Hetzner VPS  (Phase 2 web face)
```

- **Host:** Hetzner VPS (small). App ships as a single static Go binary, run
  under **systemd** (or a container). Host key persisted on a stable path.
- **Cloudflare + the SSH caveat:** Cloudflare's standard proxy (orange cloud)
  is **HTTP/HTTPS only** and will **not** proxy plain SSH on port 22. Options:
  1. **DNS-only (grey cloud)** A record → straight to the Hetzner IP. Plain
     `ssh paran.in` works, but Cloudflare's DDoS shield does **not** cover SSH
     and the origin IP is exposed. ← simplest; relies on our own SSH hardening.
  2. **Cloudflare Spectrum** — proxies raw TCP/SSH (DDoS-protected), but it's a
     paid/Enterprise feature.
  3. **Cloudflare Tunnel** — free, but clients must use `cloudflared`
     (`ProxyCommand`), so it's not vanilla `ssh paran.in`.
- **Decision deferred to deploy time.** Default expectation: **option 1** for
  SSH (with strong host-level hardening — see REQUIREMENTS §5) and Cloudflare in
  front of the **web** face only.
- **Firewall:** Hetzner Cloud Firewall + host firewall (ufw/nftables); expose
  only the SSH-TUI port and (later) 80/443. Consider moving the *admin* SSH (for
  server management) to a non-standard port or key-only + allowlist, separate
  from the public TUI port.

## Testing strategy

How tests are written, at which level and when they run: `.claude/rules/tests.md`. Every gate a
commit runs: `.pre-commit-config.yaml`.
