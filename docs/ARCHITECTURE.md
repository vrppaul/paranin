# Architecture

## ADR-0001 — Language and TUI/SSH stack

**Status:** accepted (2026-06-02)

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
| Log | `charm.land/log/v2` | v2.0.0 | structured logging |
| ssh | `github.com/charmbracelet/ssh` | pinned | low-level SSH session layer |

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
  more plumbing, no middleware. Wish wraps `charmbracelet/ssh` with sane
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
  the vim chrome, the salary range, the job matcher, the chat, an article's
  widgets. Behaviour over server markup that draws nothing of its own is a
  TypeScript controller.
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

**Status:** accepted in part (2026-11-01): content published without a new binary (D5) and one format for
every face (D8) are decided; the store, the document model and the layers are proposed, waiting for Pavel.
[FOUNDATION](FOUNDATION.md) §5.2.

### Context

The first version embedded `profile.yaml` in the binary and parsed it once,
so no post could change without a build (against D5). Each face laid the
content out again by hand, and `content` held presentation and web routes.

### Decision

- **Content lives on disk.** `content` loads a folder (the profile, the
  posts, the word catalogues), checks it whole and makes an immutable
  snapshot. A store holds the snapshot in an `atomic.Pointer`. On SIGHUP or a
  change in the folder, after the changes settle, it loads the whole folder
  again; a snapshot that fails its check is refused, the old one stays, and
  the error is logged. Nothing of the content is embedded; tests read
  fixtures.
- **`paranin check`** runs the same loader, so CI refuses what the server
  would refuse.
- **One document model**, `document`: headings, paragraphs, lists, code,
  facts, figures, widgets, links. `compose` turns content into documents: a
  post's markdown, through goldmark, and the résumé, its sections written in
  one order once. Every face draws documents; the HTML résumé keeps a layout
  of its own and takes its sections from `compose`.
- **`routes`** writes every address.
- **The layers**: `content`, `document` and `look` at the foot, side by
  side; `routes` on them, and `compose` on `routes`; `inbox` on the foot
  alone, beside both; the faces and `brain` side by side on those, no face on
  another; `hosting` over everything but the composition root. depguard holds
  them (`.golangci.yml`).

### Rationale

- A snapshot swapped whole means no request sees content half loaded, and a
  bad push leaves the site as it was.
- With one document model a test can hold the rule the first version kept by
  discipline: every fact `compose` emits appears in every face.

### Alternatives considered

- **A database for content.** Content is text people edit in git (D6);
  rejected. Runtime state (chat sessions, cost caps) gets SQLite with the
  brain.
- **Keeping an embedded copy as a fallback.** Two sources of the same facts;
  rejected.

### Consequences

- The binary cannot start without a content folder; development points it
  at the repository's own.
- A face is written against `document`, never against `content`'s types
  directly, but for the HTML résumé's own layout.

---

## ADR-0004 — Publishing, and the one front-end build

**Status:** accepted in part (2026-11-01): publishing without a new binary, posts in this repository, the four
statuses, one format and translations later (D5–D8, D12) are decided; the pipeline and the one front-end
build are proposed, waiting for Pavel. [FOUNDATION](FOUNDATION.md) §5.3 and §5.4.

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
  catalogues every face reads, there from the first face.

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

## System overview — one brain, many faces

> The sections from here to the deployment describe the first version, kept on the owner's
> machine while its replacements land. ADR-0002 to ADR-0005 and [FOUNDATION](FOUNDATION.md) §5
> replace them, package by package; each section is rewritten as its package lands.

```
                       ┌────────────────────────────┐
                       │  internal/content (profile) │  single source of truth
                       └──────────────┬─────────────┘
                                      │
        ┌──────────────┬─────────────┼───────────────┬─────────────────┐
        ▼              ▼             ▼               ▼                 ▼
   SSH TUI        Web résumé     CV generator   LinkedIn helper   (future faces)
   (Phase 1)      (Phase 2)      (Phase 5)      (Phase 6)
```

Every face reads from `internal/content`. Phase 1 builds the SSH face only, but
the content package is designed to be face-agnostic from day one.

## Phase 1 component design (SSH TUI)

```
   ssh client ──TCP:22/23234──▶ ┌─────────────────────────────────────┐
                                │ internal/server (Wish)              │
                                │  middleware chain (outer→inner):    │
                                │   recover → logging → activeterm →  │
                                │   bubbletea.Middleware(teaHandler)  │
                                └───────────────┬─────────────────────┘
                                                │ per connection
                                                ▼
                                ┌─────────────────────────────────────┐
                                │ internal/tui (Bubble Tea program)   │
                                │  root model: menu + page router     │
                                │  pages render from content          │
                                │  styles: lipgloss theme             │
                                └─────────────────────────────────────┘
```

**Connection flow:** Wish accepts the TCP connection, runs the middleware chain,
and the Bubble Tea middleware starts a fresh `tea.Program` bound to that
session's PTY. The program runs The Elm Architecture: `Init → Update(msg) →
View` until the user quits or disconnects. Window-resize events from the SSH
client are forwarded into the program as `tea.WindowSizeMsg` automatically.

**Per-connection isolation:** each session gets its own model instance; a panic
in one session is contained by the recover middleware and does not affect others
or crash the server.

## Phase 2 component design (web)

```
   browser, curl, agent ──HTTP──▶ ┌──────────────────────────────────────┐
                                  │ internal/web (net/http)              │
                                  │  GET /, /p/{slug}: negotiate →       │
                                  │    HTML template | markdown | man    │
                                  │  POST /api/contact: validate →       │
                                  │    rate limit → data/messages.jsonl  │
                                  └──────────────────┬───────────────────┘
                                                     │ reads
                                                     ▼
                                  internal/content: Resume from profile.yaml
```

- `cmd/paranin` runs the SSH and HTTP servers in one process and stops both
  on `SIGINT` or `SIGTERM`.
- `content.Load` parses `profile.yaml` once at start-up. The HTML pages, the
  markdown, the man page, `llms.txt`, the sitemap and JSON-LD all render from
  that one `Resume`.
- Templates and static files are embedded with `go:embed`. Templates are
  parsed at start-up, so a broken one stops the process before it serves.
- `make assets` prints `/cv` to `static/cv.pdf` and captures `/og` to
  `static/og.png` with headless Chrome; the next build embeds them.
- The contract between server, templates and browser code is
  [WEB](WEB.md).

## Project layout

```
cmd/paranin/main.go      # wire config + start both servers; handle signals/shutdown
internal/server/         # Wish server construction, middleware, security policy
internal/tui/            # Bubble Tea root model, page views, lipgloss styles
internal/web/            # HTTP routes, renderers, contact API, templates, static files
internal/content/        # profile.yaml + Resume types (the "brain"); face-agnostic
scripts/build-assets.sh  # cv.pdf and og.png from the running server
```

The MVP may start with fewer files and grow into this layout.

## Configuration

MVP reads from env with sane defaults (12-factor):

| Var | Default | Meaning |
| --- | --- | --- |
| `PARANIN_HOST` | `localhost` | bind host |
| `PARANIN_PORT` | `23234` | bind port |
| `PARANIN_HOST_KEY_PATH` | `.ssh/id_ed25519` | host key path (persisted, gitignored) |
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
