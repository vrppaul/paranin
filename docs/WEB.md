# Web face

The contract between the web server and its readers: what each address answers, in which format and with which
headers, and what the page does in a browser. How it is built is [ARCHITECTURE](ARCHITECTURE.md), "Phase 2
component design (web)"; how it looks and acts, `DESIGN.md`. The examples are the golden files the tests hold, drawn from the
fixture profile: `internal/face/text/testdata` for the text formats, `internal/face/html/testdata` for the page.

Three readers, three formats, from one document per page:

- a browser, and whoever names no format by its exact type but a terminal's client, crawlers and link previews among
  them, gets **HTML**, one tab of the page at a time;
- an agent gets **markdown**, where it asks for it or opens a `.md` file;
- a terminal's client, `curl` among them, and whoever asks for plain text get **the manual**: a man page laid out as
  text, 78 columns, no colour codes.

Beside them a feed reader polls the blog's feeds, in **Atom**, each at an address of its own.

## Addresses

A page is the résumé, a project's page, a post or the blog's list. Each section past the opening of the résumé or a
project's page is a tab, named by its title's letters and digits in small letters, a hyphen between each run of them
(`Experience` is `experience`); a post and the list have no tabs. A post is served while it is published or unlisted
and its `published` time has come, and as a draft only under `paranin serve --drafts`, to its author; the list names
the posts published and out alone. A post dated later comes out at its time with no reload, and is a draft until then
in who reads it and where it is listed, its version marked by its day, never the word `draft`, which a post written
as a draft alone reads (`docs/FOUNDATION.md` §5.4, Status).

A post's translation, and the blog in its language, stand in the language's folder, for each language the content
has words for (`content/i18n/`), its tag two or three small letters: every address of the blog below with
`/{language}` before it, as `/ru/blog/{slug}`, `/ru/blog.md?tag={tag}` and `/ru/blog.atom`. The blog in a language
lists the posts out in it alone, in its words; English, the site's own language, has no folder. Nothing answers by
`Accept-Language`: the address alone picks the language.

| Method, path | Answer |
| --- | --- |
| `GET /`, `GET /p/{slug}` | The page, in the format the reader's request chooses: the text formats write it whole, the HTML shows its opening tab. |
| `GET /{tab}`, `GET /p/{slug}/{tab}` | One tab, chosen the same way: the text formats write it alone, the HTML shows it inside its page. |
| `GET /blog/{slug}` | A post, whole, in the format the reader's request chooses. |
| `GET /blog`, `GET /blog/series` | The blog's list, every post published, by year, newest first, and by series, the posts in no series last, chosen the same way. No two pages share an address: content where two would, as a post slugged `series`, is refused as it loads. |
| `GET /blog?tag={tag}` | The list by year filtered to the posts of a tag, chosen the same way, for each tag of a post the list names. On the list by year the query names one tag; any other key, and the query of any other path, is passed over. A tag named twice, or one no list has, is `404`. Each is drawn the first time a request asks for it, and kept. |
| `GET /index.md`, `/p/{slug}.md`, `/{tab}.md`, `/p/{slug}/{tab}.md`, `/blog/{slug}.md`, `/blog.md`, `/blog/series.md`, `/blog.md?tag={tag}` | The same in markdown, `text/markdown`. |
| `GET /index.txt`, `/p/{slug}.txt`, `/{tab}.txt`, `/p/{slug}/{tab}.txt`, `/blog/{slug}.txt`, `/blog.txt`, `/blog/series.txt`, `/blog.txt?tag={tag}` | The same as the manual, `text/plain`. |
| `GET /blog/{slug}/{date}`, `.md`, `.txt` | An older version of a post out, published or unlisted, by the day it came out (`2026-11-10`), its revision's name (`docs/FOUNDATION.md` §5.4, Revisions), chosen and in each format as the post is; drawn the first time a request asks for it, and kept. A draft's, a post's dated later among them, are served nowhere. The current version's day, the name the post's versions list it by, is `302` to the same face at `/blog/{slug}`, with `Cache-Control: no-cache`, since that day names an older version once the post is updated. |
| `GET /blog/{slug}/{picture}` | A picture a version of the post shows, from its folder, as its type: `image/png`, `image/jpeg`, `image/webp` or `image/svg+xml`. Nothing else of the folder is served, and no path leads out of it. A picture's name ends in its type's ending, which no day and no format's ending is, so a version and a picture never share a path. |
| `GET /blog/{number}` | `301` to `/blog/{slug}`, by its path from the root, with `Cache-Control: no-cache`, since a post's slug is its folder's name, which may change. |
| `GET {path}/` | A path this table serves, a file's among them, with a slash at its end: `301` to the path without it, its query kept as written; where that path is sent on, as a post's number is, straight to where it is sent, with its status; with `Cache-Control: no-cache`, since a reload may take the path away. `/` is the résumé; any other path with a slash at its end whose path without it is served nowhere, the page of no page. |
| `GET /blog.atom`, `GET /blog/series/{name}.atom` | The blog's feed and a series', in Atom, `application/atom+xml`, whatever the request asks: each the twenty newest posts the list names, or the series' parts among them, for each series with a part out. `404` where the list names none. |
| `GET /{language}/blog…` | The same in a language, for each address above in the blog's folder and the feeds': the blog of the posts translated into it, `/ru/blog/{number}` sent on to `/ru/blog/{slug}`. `404` in a language the content has no words for, and for a post not translated into it. |
| `GET /llms.txt` | The index of the markdown files ([llmstxt.org](https://llmstxt.org)): markdown, sent as `text/plain`, as a name ending in `.txt` promises. |
| `GET /robots.txt` | What a crawler may read (RFC 9309), `text/plain`, whatever the request asks: every crawler let in and nothing disallowed, the sitemap and, in a comment, `llms.txt`, each by its whole address (`docs/SEO.md`, Crawling). |
| `GET /sitemap.xml` | Every page a search engine may index ([sitemaps.org](https://www.sitemaps.org/protocol.html)), `application/xml`, whatever the request asks: each by its whole address, with the day its text last changed where it has one (`docs/SEO.md`, Crawling). |
| `GET /{key}.txt` | The IndexNow key the site is known by, where it is served with one, never with its drafts (ARCHITECTURE, Configuration): its body the key alone, `text/plain`, whatever the request asks, which tells a search engine that reads IndexNow the site's word of a change is the site's own ([indexnow.org](https://www.indexnow.org/documentation); `docs/SEO.md`, Crawling). |
| `GET /assets/{name}-{hash}.js`, `.css` | A file of the page's build, by the name the build's manifest gives it; never the manifest itself, nor the widgets' props' schemas. |
| `GET /cards/{hash}.png` | A page's card, the picture a link's preview shows of it (`docs/SEO.md`, A card), `image/png`, whatever the request asks: by the name the cards' manifest gives it, the start of the sha256 of what it shows; never the manifest itself, nor any other file of the cards' folder. |
| `GET /favicon.svg`, `/apple-touch-icon.png`, `/favicon.ico` | The site's icon, one picture for every reader: a drawing, a PNG of 180px and an icon file of 32px. |
| any other `GET` | `404`, the page of no page: the status, the address asked for, quoted, its path decoded and its query as written, and where the site begins; text or HTML, chosen as a page is, the HTML in the frame of the language whose folder the path opens with, the site's own where none does. |
| any other method | `405`, `Allow: GET, HEAD`, from `net/http`. |
| a request whose line and headers pass the server's bound, 16 KB and the 4 KB of slack `net/http` adds | `431`, from `net/http`, before the handler reads it. |

`HEAD` answers as `GET` does, without the body. A path with an empty segment or a dot segment (`//`, `/a/../b`) is
redirected to its clean form by `net/http`.

How long a cache keeps an answer:

| Answer | `Cache-Control` |
| --- | --- |
| A page in any format, a feed, robots.txt and the sitemap, the IndexNow key, a post's picture, the icon, the page of no page, a post's number sent on | `no-cache`: a reload, or a post coming out, can change any of them but the icon, which keeps its path when a new build of the server draws it anew, and a post's number its own when the post's slug does |
| A built file, a card | `public, max-age=31536000, immutable`: a changed file is built, and a changed card drawn, under another name |

### Negotiation for `/`, `/p/{slug}`, their tabs, `/blog/{slug}` and the blog's list, filtered or not

One function chooses the format, HTML (`text/html`), markdown (`text/markdown`) or the manual (`text/plain`), each
sent with `; charset=utf-8`; why, `docs/SEO.md`, Formats (D21). Each format takes the quality of the closest range of
the `Accept` header that reaches it (RFC 9110, 12.5.1): its exact type, which names it, or a wildcard, its family
(`text/*`) or `*/*`.

- A format at quality 0, or one no range reaches, is refused and never ranked.
- Each agent has a format of its own: the manual for a terminal's client, one whose `User-Agent` opens with curl,
  Wget, HTTPie or xh, the name before its first `/`, in capitals or small letters; HTML for anyone else.
- Where the header names by its exact type no format it does not refuse, as where it is `*/*` or `text/*` alone, empty
  or absent, the agent's own goes; where it refuses that, as `text/html;q=0, */*` does HTML, the first it accepts in
  the order below; where it refuses every format, the agent's own all the same. So `curl` gets the manual, and a
  crawler sending `*/*` the HTML.
- Otherwise the format the header ranks highest goes; at equal quality one named by its exact type before one a
  wildcard reaches; among those named, markdown, then the manual, then HTML, since a browser names neither of the
  first two; among those a wildcard reaches, the agent's own first, then in that order. So
  `text/markdown, text/html, */*` gets markdown, `text/html, */*` HTML, and `text/markdown;q=0.5, */*` the agent's own.

The page of no page is chosen the same way, its text where the manual or markdown would be. Every negotiated response
sets `Vary: Accept, User-Agent`, the page of no page among them: a filter's query is part of the address a cache keys
on already. Every format of a page, at its own path and where the reader chooses, sets
`Link: <canonical URL>; rel="canonical"`, the URL where the reader chooses; the HTML names it in its head too, with
`rel="alternate"` to its markdown. Three kinds of page are not indexed: a filtered list, a view of the list and no
page of its own (`docs/FOUNDATION.md` §5.4); an older version of a post, which a newer stands over; and an unlisted
post, read by whoever has its address (owner, 2026-11-24). Each sends `X-Robots-Tag: noindex` in every format, and its
HTML says so in its head too (`<meta name="robots">`). None names a canonical URL. A filtered list and an older
version have no address of their own, so they carry no Open Graph tags. An unlisted post keeps its preview. Links
inside the HTML lead to the site's pages by their path from the root; the canonical URL and the text formats' links
are whole.

## What each format holds

The text formats write every block of their document, a gap still to fill as the author wrote it (`[notice
period]`). The HTML shows a page as a reader of the finished work sees it: a fact or a point that is only a gap is
left out, and a gap among written words stands marked as one.

- **The manual** opens with `NAME` and `DESCRIPTION`, then a section for each of the page's, as `man(1)` lays a page
  out, and `NOTES` where a post cites any, each by its number as `[1]`, its title line naming the manual's section,
  `User Commands`, each in the page's language, as its catalogue writes them, `ИМЯ` and `Команды пользователя` in
  Russian; its last line holds the day of a post's version read in the middle, as `2026-11-10`, or `draft` for a
  draft, where `man(1)` writes a page's date. An older
  version's `DESCRIPTION` opens with the newer that stands over it, `newer: 2026-11-21` and its address, and a
  translation older than its original's with the site's words saying so and the original's address. A post read in
  more than one language lists them under its lead, each by its tag, the others with their addresses,
  `en · ru <…/ru/blog/{slug}.txt>`. A link to a page of the site,
  or to a section of one, is its text and its page's address between `<` and `>`, since neither curl nor man reads an
  anchor; a card's page, its address on a line of its own. A post's table stands in columns under its header where it fits the 78
  columns, a bullet a row where not; a quote behind `│`; a picture its alt text, its address between `<` and `>` and
  its caption, a line each, and a widget so, its address the live one's, the post's page at the widget,
  `<https://paran.in/blog/{slug}#PageCache>`; a tab in code as the spaces to its stop, every eight columns. The blog's list stands a
  section a group, each post its title with its day at the right, then its lead, its part and series, its tags and
  its address; a filtered list is named as the list, `BLOG(1)`, its filter in `NAME` after the blog's,
  `blog — posts tagged go`.
- **Markdown** is CommonMark, the page's sections as headings; a post's head gives its version's day under its title,
  `draft` for a draft, and an older version's the newer beside it, `newer: [2026-11-21](address)`, a translation older
  than its original the site's words saying so, a link to the original's markdown; under it the
  languages it is read in, the one read strong and the others links, `**en** · [ru](address)`; its tables, code with its language and footnotes as GFM writes them, `[^1]`, the notes at the end, and a picture as
  `![alt](address "caption")`, its address whole, and a widget as the same link to the live one with no picture's
  mark, `[alt](address "caption")`. The blog's list
  stands a heading a group, each post a heading under it, a link to its markdown, then its day, part and series, its
  tags and its lead; a filtered list writes its filter on the line under the blog's name.
- **A feed** is Atom (RFC 4287), every address in it whole: its id, a tag URI under the profile's `site.tag_authority`
  and its path (`tag:paran.in,2026:blog.atom`), its title in the site's words by the author's name, the blog's line or
  the series' as its subtitle, when its newest version came out, a link to itself and to the list it follows, the
  series' group on the list by series for a series', and its author, by name and the résumé's address; a language's
  feed in its words, `xml:lang` its tag, its id under its folder (`tag:paran.in,2026:ru/blog.atom`). Each post is an
  entry, newest first: its id by its number (`tag:paran.in,2026:blog/12`, a translation's `…:ru/blog/12`), which
  neither a renamed folder nor a moved site changes, its title, a link to its page, when it came out and when its current version did, the start of its
  `updated:` day in its own zone, its lead as the summary and its body whole as HTML, every address whole, its code in
  no colour, its notes numbered at its end, a widget its alt text as a link to the live one over its caption, with no id, paragraph's number, point's mark, button or script, which a
  reader would show as stray words or run nothing for.
- **`llms.txt`** is the author's name, the headline, then `## Résumé`, `## Projects` and `## blog`, a link to each
  page's markdown: under the blog, its list by year and by series, each by the blog's line, then each post the list
  names, newest first, by its description; never a draft, a post unlisted, one not yet out or an older
  version.
- **The HTML** of every page holds, in its head, the page's title, its description where it has one, the site's icon
  and what runs the site's script; `<html lang>` is its language's tag, and the frame's own words, Neovim's and tmux's
  (`vim keys`, `esc close`, the command line's keys, kinds and names, Neovim's words on its line), stand in
  `lang="en"` on every page, and what the reader types and the line echoes in the page's own. A page with an address of
  its own adds Open Graph, which a link preview reads, and, where a search engine may index it, its canonical link, its
  languages and its record (`docs/SEO.md`). A page of the blog names the feeds that follow it,
  `<link rel="alternate" type="application/atom+xml">` with each feed's title, a series' first on a post's page that
  shows the series, then the blog's, which the résumé and its tabs name too, for a feed
  reader handed the page, or the site's address, to find them. A post's page is the post whole: its title the one
  `h1`, its version's day in a `<time datetime>`, a draft's version the word `draft`; a post out its version's day a button
  opening, as a popover, the list of its versions, each a link to it but the one read, and an older version under it a
  link to the newer, a translation older than its original a link to the original, `hreflang` its language; its day
  in its language's words, `15 нояб. 2026`; its languages a Switch, each a link by its tag with `hreflang`, the one read
  `aria-current="page"`; each paragraph a place `#_p1` names and its
  number a link to it, each heading a place its anchor names, by its words; every id of the page's own opens with `_`, which no anchor holds, so no two
  places of the page share one name; a note's number raised where a line cites it, the note beside the line that first
  cites it, and every note once at the text's end, which a phone shows in their place; a picture a `<figure>`, its
  `<img>` holding its alt text and its size, so the page keeps its room before it comes, fetched as it nears the
  screen, and its caption a `<figcaption>`; a widget a `<figure>` its anchor names, by the widget's name, counted from
  the second of a name (`#PageCache-2`), its window holding its name, the mark that it runs, hidden until the script
  runs it, its body, where the widget runs, holding its alt text, which describes the figure (`aria-describedby`) once
  the widget runs, and its props, a block of JSON (`application/json`), its caption a `<figcaption>`. The blog's list
  has the blog's name as its `h1`, each group's name an `h2`, a series' group a place its name names, each post a link
  to it, its day in a `<time datetime>` then its title, stretched over its tile, its series a link to its group on the
  list by series, a part to come no link; it has no record (`docs/SEO.md`). The list's tags and a post's are
  links to the list each filters, a post's only where the lists name it; the filter shown is `aria-current="true"`, leads back to
  the list and is named to a screen reader as the way there (`posts tagged go, show all posts`); a filtered list is called by the blog's name and its filter, `blog — posts tagged go`. Every page ends with the
  foot line (`DESIGN.md`, FootLine), a post's and the list's in the blog's session.

## Security headers

Every answer that reaches the server's handler carries these headers, a page in each format, the page of no page and
a refusal the mux writes itself (405, a redirect) among them; a request `net/http` refuses before it, as a `400`
for a malformed `Host` or a `431` for a head past its bound, carries neither:

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

A post's picture is sent under a policy of its own in place of that one, so a drawing opened by itself, which may hold
a script, loads and runs nothing and stands as a document of no site; its own styles stay, which a drawing is drawn by.
A page shows it with `<img>`, where a browser runs nothing in a drawing either:

```
Content-Security-Policy: default-src 'none'; style-src 'unsafe-inline'; sandbox
```

How the page's build keeps to it is `.claude/rules/page.md`. The page's face, JetBrains Mono, is a built file of the
site like any other, from `design/fonts` (`design/tokens.yaml`, `face`), so every reader sees the copy the cells are
counted in; a mono face of the reader's own shows the text until it has come. An article's words are set in iA
Writer Duo, served the same way (`reading_face`). No `Referrer-Policy` is sent:
`strict-origin-when-cross-origin` is every browser's own default.

## The page in a browser

Every page runs one script, `web/src/entries/site.ts`. It runs the parts whose look and behaviour `DESIGN.md` gives:
ThemeSwitch, Command and a post's code that copies as it does, Hint, Tooltip and Toast, VimKeys and CommandLine. A
page that shows a widget loads Svelte and the widget's component at once, each a built file of its own that no other
page loads, and runs the component in the widget's window with the props the page holds as data, read by
`JSON.parse` and taken as they stand, since the release the page was drawn from held them to the widget's schema;
`● live` then shows, and the alt text stands aside, still the figure's description to a screen reader. The widget
follows the theme the page is shown in as the reader or their system turns it, and whether the reader asks for reduced
motion; one that draws on a canvas, the kit's stage, draws in that theme's colours (owner, 2026-11-24).
With no script the page reads whole, a widget its alt text, and the controls the script runs stay hidden.

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
