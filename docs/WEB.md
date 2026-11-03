# Web face

> This is the contract of the first version, kept on the owner's machine while its
> replacement lands. ADR-0002 to ADR-0004 of [ARCHITECTURE](ARCHITECTURE.md) change it
> where they differ; it is rewritten with `face/html` and `hosting/web`. Rewritten so
> far, and the contract the code holds: "Negotiation for `/` and `/p/{slug}`".

`paran.in` serves one résumé to three kinds of reader from one profile:

- **people** — HTML with two modes: *normies* (default, everything clickable)
  and *vim* (the same page driven from the keyboard);
- **agents and terminals** — markdown and a man-page style plain text;
- **print** — `cv.pdf`, printed from a dedicated HTML page.

The SSH app (`ssh paran.in`) reads the same profile. This document is the
contract between the Go server, the templates and the browser code.

## Packages

| Path | Holds |
| --- | --- |
| `internal/content/profile.yaml` | All facts. Nothing factual lives anywhere else. |
| `internal/content/resume.go` | `Resume` types, `Load`, `MustLoad`, `Parse`, validation. |
| `internal/content/profile.go` | Adapter from `Resume` to what the TUI needs. |
| `internal/web/*.go` | HTTP server: routes, negotiation, renderers, contact API. |
| `internal/web/templates/` | `html/template` files. |
| `internal/web/static/` | `ui.css` (design library), `site.css`, `theme.js`, `site.js`, built `cv.pdf` and `og.png`. |
| `internal/tui/` | The SSH app. |
| `cmd/paranin/` | Starts the SSH and HTTP servers. |

## Inline marks

Any text from the profile may carry two marks.

| Mark | HTML | Markdown | Plain text |
| --- | --- | --- | --- |
| `==text==` | `<mark class="on-tint">text</mark>` | `**text**` | `text` |
| `[text]` | `<span class="on-faint">[text]</span>` | `[text]` | `[text]` |

`content.ParseMarks` is the only parser; the web and the TUI render its pieces.
A gap may sit inside a tint. A `[text]` directly followed by `(` is a markdown
link and is not a gap.

## Routes

| Method, path | Response |
| --- | --- |
| `GET /` | Index. Negotiated: HTML, markdown or plain text. |
| `GET /index.md` | Index as markdown. |
| `GET /p/{slug}` | Project page for projects with `page: true`. Negotiated. |
| `GET /p/{slug}.md` | Project page as markdown. |
| `GET /cv` | Print page used to make `cv.pdf`. `X-Robots-Tag: noindex`. |
| `GET /og` | 1200×630 card used to make `og.png`. `X-Robots-Tag: noindex`. |
| `GET /cv.pdf`, `GET /og.png` | Built assets from `static/`; 404 until built. |
| `GET /llms.txt` | Short map of the markdown files. |
| `GET /robots.txt` | Allows everything; points to the sitemap. |
| `GET /sitemap.xml` | `/` and every project page, `lastmod` = `site.updated`. |
| `GET /static/{file}` | Embedded static files. |
| `POST /api/contact` | Contact form. JSON or `application/x-www-form-urlencoded`. |
| `GET /healthz` | `ok`. |
| anything else | 404: HTML template, or plain text for terminals. |

### Negotiation for `/` and `/p/{slug}`

The format the `Accept` header ranks highest, by the closest range that names
it (RFC 9110, 12.5.1): HTML (`text/html`), markdown (`text/markdown`) or the
manual (`text/plain`), each with `; charset=utf-8`. Where it ranks two alike
the manual goes first, then markdown, then HTML, so `curl`, which sends
`*/*`, and a request with no `Accept` get the manual (D21).

Every negotiated response sets `Vary: Accept` alone. Every format of a page,
at its own path and where the reader chooses, sets
`Link: <canonical URL>; rel="canonical"`, the URL where the reader chooses;
the HTML names it in its head too, with `rel="alternate"` to its markdown.
Links inside the HTML lead to the site's pages by their path from the root;
the canonical URL and the text formats' links are whole.

## Templates

Pages are parsed as `layout.html` + every `partials/*.html` (if any) + the page
file, and executed as `"layout"`. The page file defines `"main"`.

| File | Defines | Route |
| --- | --- | --- |
| `layout.html` | `layout` | all HTML pages |
| `index.html` | `main` | `/` |
| `project.html` | `main` | `/p/{slug}` |
| `notfound.html` | `main` | 404 |
| `cv.html` | `cv` (a full document, parsed alone) | `/cv` |
| `og.html` | `og` (a full document, parsed alone) | `/og` |

### Data: `web.Page`

```go
type Page struct {
	Resume      *content.Resume
	Project     *content.Project // project page only
	Title       string           // <title>
	Description string           // meta description, ≤ 160 chars
	Canonical   string           // absolute URL of this page
	MarkdownURL string           // "/index.md" or "/p/{slug}.md"; empty on 404
	OGImage     string           // absolute URL of og.png
	JSONLD      template.JS      // JSON for <script type="application/ld+json">
	ClientData  template.JS      // JSON for <script type="application/json" id="on-data">
	Buffers     []Buffer         // "/" first, then every project with page: true
	NotFoundPath string          // 404 only: the requested path
	ContactResult string         // "sent" or "error" after a form post without JavaScript
}

type Buffer struct {
	Index   int // 1-based
	Title   string
	URL     string
	Current bool
}
```

### Template functions

| Name | Signature | Example |
| --- | --- | --- |
| `inline` | `string → template.HTML` | applies inline marks |
| `span` | `start, end string → string` | `09/2012 – 06/2016`; gaps and `present` stay as they are |
| `period` | `content.Role → string` | `01/2024 – present` |
| `years` | `content.Role → string` | `2024 – now`, `2020 – 2022` |
| `skillGroups` | `[]content.Skill → []SkillGroup` | groups in order languages, backend, ai, design |
| `hostpath` | `string → string` | `https://github.com/vrppaul` → `github.com/vrppaul` |

`SkillGroup` is `{Key, Title string; Skills []content.Skill}` with titles
`Languages`, `Backend`, `AI and ML`, `Design`.

### Client data (`#on-data`)

The browser code reads everything it needs from this JSON block and from the DOM.

```json
{
  "ssh": "ssh paran.in",
  "cvUrl": "/cv.pdf",
  "markdownUrl": "/index.md",
  "pages": [{"title": "index", "url": "/"}, {"title": "claude-review", "url": "/p/claude-review"}],
  "skills": [{"name": "Python", "duration": "7y", "aliases": ["python"], "sources": ["Seznam.cz"], "learning": false}],
  "health": [{"name": "experience", "checks": [{"level": "ok", "text": "…", "advice": ""}]}],
  "about": ["A role", "A contract", "A project", "Something else"],
  "salary": {"min": 100000, "softMax": 300000, "step": 10000, "bigStep": 50000, "currencies": ["USD", "EUR"], "period": "year"}
}
```

## Contact API

`POST /api/contact`, body at most 32 KB.

```json
{
  "name": "Jane Novak",
  "email": "jane@example.com",
  "about": "A role",
  "message": "Hi Pavel, …",
  "salary": {"currency": "USD", "min": 140000, "max": 190000},
  "website": ""
}
```

- `salary` is optional; `max: null` means no upper limit.
- `website` is a honeypot. When it is not empty the server answers `200` and
  stores nothing.
- Checks: `name` 1–100 characters; `email` has one `@`, a dot in the domain,
  ≤ 254 characters; `about` is empty or one of `contact.about`; `message`
  10–5000 characters; `salary.currency` is one of the currencies; `salary.min`
  ≥ `contact.salary.min`; `salary.max` is null or ≥ `salary.min`.
- Form posts carry no salary: the salary range needs JavaScript.
- Rate limit per 10 minutes: 5 accepted messages per client and 50 in total.
  A client is its IPv4 address or its IPv6 /64. The address is the connection's;
  the right-most `X-Forwarded-For` entry is used only when
  `PARANIN_TRUST_PROXY=1`.
- Cross-site posts are rejected with `403`.
- Storage for the prototype: one JSON line per message appended to
  `PARANIN_MESSAGES_PATH` (default `data/messages.jsonl`) with `received_at`.
  Logs say that a message arrived, never what it says. Delivery (Telegram,
  e-mail) is decided later.

| Result | JSON request | Form request |
| --- | --- | --- |
| accepted | `200 {"ok": true}` | `303` to `/?sent=1#contact` |
| invalid | `400 {"ok": false, "errors": {"email": "Add the domain, like jane@example.com"}}` | `303` to `/?error=1#contact` |
| too many | `429 {"ok": false, "error": "Too many messages. Try again in a few minutes."}` | plain text, `429` |
| too long | `413 {"ok": false, "error": "The message is too long."}` | `303` to `/?error=1#contact` |
| unreadable JSON | `400 {"ok": false, "error": "…"}` | — |
| other content type | `415` | `415` |
| storage failed | `500` | `500` |

`/?sent=1` and `/?error=1` set `Page.ContactResult`, and the form's status line
says what happened.

## Plain text (terminals)

A man page, 78 columns, no colour codes:

```
PARANIN(1)                     User Commands                    PARANIN(1)

NAME
       paranin — Senior Software Engineer in Prague, Czechia. Builds backend
       services, and tools for people who live in the terminal.

SYNOPSIS
       ssh paran.in                 # the interactive version
       curl -O https://paran.in/cv.pdf

EXPERIENCE
       Kiwi.com — Senior Software Engineer            01/2024 – present
              - Owned [service] handling [N] requests/day
…
PROJECTS, SKILLS, EDUCATION, LANGUAGES, CONTACT

SEE ALSO
       https://paran.in, github.com/vrppaul, linkedin.com/in/pavel-paranin

paran.in                          2026-09-13                     PARANIN(1)
```

`DESCRIPTION` holds the summary and the facts. A project page in plain text uses
the same frame with `NAME`, `DESCRIPTION`, `INSTALL`, `EXAMPLE` (the recording,
when there is one), the sections and `SEE ALSO`. Long commands wrap with a
trailing `\`.

## Markdown

```markdown
# Pavel Paranin

Senior Software Engineer · Prague, Czechia

Builds backend services, and tools for people who live in the terminal.

- Email: paraninpasha@gmail.com
- GitHub: [github.com/vrppaul](https://github.com/vrppaul)
- LinkedIn: [linkedin.com/in/pavel-paranin](https://www.linkedin.com/in/pavel-paranin/)
- SSH: `ssh paran.in`
- PDF: [cv.pdf](https://paran.in/cv.pdf)

## Facts
## Summary
## Experience
### Kiwi.com — Senior Software Engineer
01/2024 – present · Prague
- Owned [service] handling **[N] requests/day**
## Projects
## Skills
## Education
## Languages
```

Markdown and HTML carry the same facts. Nothing appears in one and not the other.

## Other files

- `llms.txt`: `# Pavel Paranin`, a `>` line with title and headline, then
  `## Résumé` (links to `/index.md` and `/cv.pdf`) and `## Projects` (one link
  per project page `.md`, with its tagline).
- JSON-LD on `/`: `ProfilePage` whose `mainEntity` is a `Person` with `name`,
  `alternateName`, `jobTitle`, `worksFor` (the role whose `end` is `present`),
  `alumniOf`, `knowsAbout` (skills that are not `learning`), `knowsLanguage`,
  `email`, `url`, `sameAs` (link URLs). On project pages: `SoftwareSourceCode`
  with `name`, `description`, `programmingLanguage`, `codeRepository` (if any)
  and `author`.

## Security headers (HTML and static)

```
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'
X-Content-Type-Options: nosniff
Referrer-Policy: strict-origin-when-cross-origin
```

No inline scripts or inline event handlers. JSON blocks
(`application/ld+json`, `application/json`) are data, not scripts.

## Front end

### Design rules

`static/ui.css` is the design library (`paranin ui`). Use its tokens and
`on-*` components; `site.css` only lays pages out. The rules:

1. One grid: a cell is 1ch wide and one row (24px) tall. Sizes are rows or ch.
2. One radius: 2px on controls and pop-ups, 0 on surfaces. No pills, no circles.
3. One line weight: 1px. Quiet lines divide, normal lines outline controls.
4. Shadows only on things that float above the page.
5. Selection has three meanings: reverse = chosen option, tint = matched or
   important text, cursorline = the current row.
6. One accent (peach). State and mode colours are separate.
7. Focus looks the same everywhere: 2px accent outline, 2px away.
8. Text on the accent is always dark (`--on-accent`).

Text sizes: 12 and 14px; large text only `on-title`, `on-display`, `on-hero`.

### Theme

`theme.js` loads in `<head>` without `defer`. It sets `<html class="on latte">`
or `on mac` from `localStorage["on-theme"]`, else from
`prefers-color-scheme`. The server renders `class="on latte"` so the page works
without JavaScript.

### Index page, top to bottom

`#top` (top line with the mode switch, name, title, actions) · `#facts` ·
`#summary` · `#experience` · `#projects` · `#skills` · `#education` ·
`#languages` · `#match` (check against a job description) · `#contact` (form
with salary range) · footer (links to `index.md`, `llms.txt`, `cv.pdf`,
`ssh paran.in`).

### Hint pop-up

A small floating terminal window in the bottom-right corner of every HTML page
offers the other ways to read it. It always uses the Macchiato palette. Its title
and a `[q] close` button sit in the top border; a tmux status line at the bottom
switches between windows:

- `0:ssh` — `ssh paran.in` typed out at a prompt, with copy;
- `1:vim` (normies mode only) — a Vim intro screen whose command line
  `:set mode=vim` switches the page to vim mode.

Closing is remembered in `localStorage["on-hint-closed"]`. Typing animations
are CSS only and static under `prefers-reduced-motion`. Without JavaScript and
in print the window stays hidden, and `ssh paran.in` remains in the footer.

### Modes

- `<html data-mode="normies|vim">`, saved in `localStorage["on-mode"]`, normies
  by default; saved values from the old names `easy` and `hard` are mapped on
  load. The switch is a `on-seg on-seg--sm` with two buttons and
  `aria-pressed`, labelled "view". It is hidden without JavaScript.
- Normies mode has no keyboard layer.
- Vim mode adds a fixed statusline (mode block VIM) and command line at the bottom, a cursorline on
  the current item, relative numbers in a left gutter, and the keys below. Keys
  are ignored while focus is in a form field or when a modifier is held.

| Key | Action |
| --- | --- |
| `j` `k` (with a count, e.g. `5j`) | next / previous item (`[data-nav]`) |
| `gg` `G` | first / last item |
| `/` then `n` `N` | search the page; highlights via the CSS Custom Highlight API |
| `:` | command line with completion pop-up (`on-pop`) |
| `space` | key menu (`on-pop` with a header) |
| `s` | jump labels (`on-key--accent`) on links and items |
| `u` | source view: fetch `markdownUrl`, show it with line numbers and `on-syn-*` |
| `]b` `[b` | next / previous page from `pages` |
| `esc` | close whatever is open |

Commands shown in completion: `:cv`, `:mail`, `:ssh` (copies the command),
`:source`, `:e {page}`, `:theme`, `:normies`. Hidden commands (not in completion,
not in the key menu): `:checkhealth` / `:checkhealth paranin` (report from
`health`), `:Lazy` (skills as a plugin list), `:hire` (scroll to the form,
choose "A role", prefill the message with `Hiring: [role] at [company]`).
Unknown command → `E492: Not an editor command: {text}` in the command line.

### Job description check

A textarea. On input, match the text (case-insensitive, whole words) against
each skill's name and aliases; list matches with their sources. Also look for a
fixed vocabulary of common technology words in the text; those that match no
skill are listed as "not in the résumé — ask Pavel". The text never leaves the
browser.

### Contact form

Fields: name, email, about (`on-choice`, one option), salary range (optional),
message, hidden honeypot `website`. The salary range is two native
`<input type="range">` elements styled as `on-range` (thumbs are 4×24 cursor
bars). Scale from `min` to `softMax + step`; a value above `softMax` means no
limit. Until touched the value reads "any" and no salary is sent. Arrow keys
move by `step`, Shift+arrow by `bigStep`. Currency is a `on-seg on-seg--sm`.
Submit with `fetch` as JSON; show field errors with `on-help is-error` and
success with `on-msg-ok`. Without JavaScript the form posts normally.

### Print page (`cv.html`)

A4, one column, real text, no JavaScript, ligatures off. Crosses and the cursor
bar are CSS shapes. No running header or footer text. Dates in full
(`01/2024 – present`). Two pages at most.

## Build and run

| Command | Does |
| --- | --- |
| `make run` | `go run ./cmd/paranin` (SSH on `:23234`, HTTP on `localhost:8080`) |
| `make test` | `go vet ./...` and `go test ./...` |
| `make assets` | builds, starts the server, prints `/cv` to `static/cv.pdf` and captures `/og` to `static/og.png` with headless Chrome, checks the PDF text with `pdftotext` |

| Variable | Default |
| --- | --- |
| `PARANIN_HTTP_ADDR` | `localhost:8080` |
| `PARANIN_MESSAGES_PATH` | `data/messages.jsonl` |
| `PARANIN_TRUST_PROXY` | off |
