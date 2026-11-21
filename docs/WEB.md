# Web face

The contract between the web server and its readers: what each address answers, in which format and with which
headers, and what the page does in a browser. How it is built is [ARCHITECTURE](ARCHITECTURE.md), "Phase 2
component design (web)"; how it looks and acts, `DESIGN.md`. The examples are the golden files the tests hold, drawn from the
fixture profile: `internal/face/text/testdata` for the text formats, `internal/face/html/testdata` for the page.

Three readers, three formats, from one document per page:

- a browser gets **HTML**, one tab of the page at a time;
- an agent gets **markdown**, where it asks for it or opens a `.md` file;
- a terminal, `curl` and anyone else get **the manual**: a man page laid out as text, 78 columns, no colour codes.

## Addresses

A page is the résumé or a project's page. Each section past a page's opening is a tab, named by its title's
letters and digits in small letters, a hyphen between each run of them (`Experience` is `experience`).

| Method, path | Answer |
| --- | --- |
| `GET /`, `GET /p/{slug}` | The page, in the format the reader's request chooses: the text formats write it whole, the HTML shows its opening tab. |
| `GET /{tab}`, `GET /p/{slug}/{tab}` | One tab, chosen the same way: the text formats write it alone, the HTML shows it inside its page. |
| `GET /index.md`, `/p/{slug}.md`, `/{tab}.md`, `/p/{slug}/{tab}.md` | The same in markdown, `text/markdown`. |
| `GET /index.txt`, `/p/{slug}.txt`, `/{tab}.txt`, `/p/{slug}/{tab}.txt` | The same as the manual, `text/plain`. |
| `GET /llms.txt` | The index of the markdown files ([llmstxt.org](https://llmstxt.org)): markdown, sent as `text/plain`, as a name ending in `.txt` promises. |
| `GET /assets/{name}-{hash}.js`, `.css` | A file of the page's build, by the name the build's manifest gives it; never the manifest itself. |
| `GET /favicon.svg`, `/apple-touch-icon.png`, `/favicon.ico` | The site's icon: a drawing in the reader's theme, a PNG of 180px and an icon file of 32px, both in the dark theme. |
| any other `GET` | `404`, the page of no page: the status, the path asked for, quoted, and where the site begins; text or HTML, chosen as a page is. |
| any other method | `405`, `Allow: GET, HEAD`, from `net/http`. |

`HEAD` answers as `GET` does, without the body. A path with an empty segment or a dot segment (`//`, `/a/../b`) is
redirected to its clean form by `net/http`.

How long a cache keeps an answer:

| Answer | `Cache-Control` |
| --- | --- |
| A page in any format, the icon, the page of no page | `no-cache`: a reload can change any of them, and the icon keeps its path when the author's name changes |
| A built file | `public, max-age=31536000, immutable`: a changed file is built under another name |

### Negotiation for `/`, `/p/{slug}` and their tabs

The format the `Accept` header ranks highest, by the closest range that names it (RFC 9110, 12.5.1): HTML
(`text/html`), markdown (`text/markdown`) or the manual (`text/plain`), each with `; charset=utf-8`. Where it ranks
two alike the manual goes first, then markdown, then HTML, so `curl`, which sends `*/*`, and a request with no
`Accept` get the manual (D21).

Every negotiated response sets `Vary: Accept` alone, the page of no page among them. Every format of a page, at its
own path and where the reader chooses, sets `Link: <canonical URL>; rel="canonical"`, the URL where the reader
chooses; the HTML names it in its head too, with `rel="alternate"` to its markdown. Links inside the HTML lead to the
site's pages by their path from the root; the canonical URL and the text formats' links are whole.

## What each format holds

The text formats write every block of their document, a gap still to fill as the author wrote it (`[notice
period]`). The HTML shows a page as a reader of the finished work sees it: a fact or a point that is only a gap is
left out, and a gap among written words stands marked as one.

- **The manual** opens with `NAME` and `DESCRIPTION`, then a section for each of the page's, as `man(1)` lays a page
  out. A link to a page of the site is its text and its address between `<` and `>`; a card's page, its address on a
  line of its own.
- **Markdown** is CommonMark, the page's sections as headings.
- **`llms.txt`** is the author's name, the headline, then `## Résumé` and `## Projects`, a link to each page's
  markdown.
- **The HTML** of every page holds, in its head, the page's title, its description where it has one, the site's icon
  and what runs the site's script. A page with an address adds its canonical link; Open Graph's `og:type`
  (`website`), `og:site_name`, `og:title`, `og:url`, and `og:description` where it has one; and its record as
  JSON-LD, `ProfilePage` about a `Person` for the résumé and its tabs, `SoftwareSourceCode` for a project's page and
  its tabs. Every page ends with the foot line (`DESIGN.md`, FootLine).

## Security headers

Every answer that reaches the server's handler carries these headers, a page in each format, the page of no page and
a refusal the mux writes itself (405, a redirect) among them; a request `net/http` refuses before it, as a `400`
for a malformed `Host`, carries neither:

```
Content-Security-Policy: default-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'
X-Content-Type-Options: nosniff
```

A browser loads scripts, styles, images, fonts and what a script fetches from the site's own files alone. It refuses
an inline script or handler, an inline style (a `style` attribute, a `<style>` element, one a script adds) and a
`data:` address for an image or a font. `base-uri`, `form-action` and `frame-ancestors` are named, since
`default-src` does not reach them. What the policy lets be: JSON blocks (`application/ld+json`, `application/json`),
which are data, not scripts; module scripts of the site and their preloads; a style a script sets through
`element.style`, and `element.animate()`.

How the page's build keeps to it is `.claude/rules/page.md`. The page loads no font: its face is a list of families
(`design/tokens.yaml`, `face`), each the reader's own where they have it. No `Referrer-Policy` is sent:
`strict-origin-when-cross-origin` is every browser's own default.

## The page in a browser

Every page runs one script, `web/src/entries/site.ts`. It runs the parts whose look and behaviour `DESIGN.md` gives:
ThemeSwitch, Command, Hint, Tooltip and Toast, VimKeys and CommandLine. With no script the page reads whole, and the
controls the script runs stay hidden.

What the script keeps in the browser; where the browser keeps nothing, a choice holds for its page alone:

| Storage | Key | Holds |
| --- | --- | --- |
| `localStorage` | `theme` | `light` or `dark`, the theme the reader chose |
| `localStorage` | `vim-keys` | `on` while the vim keys' switch is ticked |
| `sessionStorage` | `search` | the search the reader kept, for the pages opened after it in the same tab |

### Theme

A page follows the system's theme, light or dark, as `color-scheme: light dark` on its root lets it. The site's
script is a module with `blocking="render"`, so a browser paints nothing until it has run; where the reader chose a
theme, it sets `<html data-theme="light">` or `"dark"`, which `ui.css` turns into the colour scheme. A browser that
knows no such hold (Firefox) may show the system's theme for a moment first.

### Keys

- The vim keys answer a key typed on the page outside a field a reader types in, with no Ctrl, Alt or Meta held.
- `:e {slug}` and `:edit {slug}` open a project's page, and the line refuses `:e` with no page in Neovim's words,
  `E471: Argument required`, and a page that is no project's, `E447: Can't find file "{slug}" in path`.
- Esc hides a tooltip showing until the pointer and the focus have both left its control (WCAG 1.4.13).
