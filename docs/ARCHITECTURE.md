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
name one of its tabs, shown alone. Amended 2026-11-22, with the posts' loader: `content` reads a post once.
Amended 2026-11-22, with the first post served: the posts' blocks and spans, a footnote as a note where a line cites
it, and the post's head. Amended 2026-11-23, with the blog's list: a document in two groupings, its groups sections
holding the posts they name. Amended 2026-11-23, on the owner's choice of its look: a post's pause. Amended
2026-11-23, with code coloured: a block of code holds its text in runs of the kinds of word a reader of code tells
apart. Amended 2026-11-23, with the command line's places: a page's own places are read off its document, `compose`
names each anchor once, the kinds of place are `document`'s, and the page's own ids take a form no anchor takes. Amended
2026-11-23, with the post's pictures: the figure, a picture read from the post's folder into the snapshot. Amended
2026-11-23, with a post's links inside the site: a link's anchor, a target of its own, and the site's check of them.
Amended 2026-11-23, with statuses and dates: `compose` composes for an audience, a moment and whether drafts are read,
the hosts keep what they draw until the next post dated later comes out, and the check runs over each moment the site
changes. Amended 2026-11-23: `compose` splits a snapshot's code once (`Prepare`). Amended 2026-11-24, with the list's
filters: a filtered list is a page of two kinds of its own, which `compose` composes with the site and the web host
draws when first asked for, a head holds the filters it offers, and `routes` reads a request's query into its address.
Amended 2026-11-24, with revisions: `content` reads a
post's revisions with it, the version read is one of three kinds and lists the post's versions, an older version is a
page of its own a post's page names by its revision, which `compose` composes with the site and the web host draws when
first asked for, and a page says whether a search engine indexes it. Amended 2026-11-24, with the feeds: `compose`
composes the blog's feed and each series' with the site, a head names the feeds that follow its page, a feed is a link's
target of its own, `routes` names what the site serves for good and each format's media type, and `face/feed` writes
Atom, each post's body as `face/html` draws it for a feed reader. Amended 2026-11-24, with translations, the model the
owner's, ways weighed below: a snapshot holds an edition in each language, its words
and its posts, a page names the language it is translated into, `routes` writes its every address in that language's
folder, `compose` composes a blog in each language, its frame in its words, whose links read the pages of every
language, and a post's head lists the languages it is read in. Amended 2026-11-24, with the site's words in every
language: an edition holds its series in its language, a catalogue writes a count in each plural form and the months,
read by x/text's CLDR rules, and the frame carries how a day and the manual's headings are written in its language.
Amended 2026-11-24, with a translation's source: `content` holds a translation to the text of its original it was made
from and keeps what it warns of beside a snapshot, and a translation older than its original says so under its day.
Amended 2026-11-24, on the owner's word: a heading's anchor is its words, whatever else its page names alike.
Amended 2026-11-24, on the owner's word: a tag alone filters the list, a filtered list a page of one kind.
Amended 2026-11-24, on the owner's word: `paranin revise` dates the post as updated on its day, `content` writing its
file so.
Amended 2026-11-24, on the owner's word: the composition root prepares a snapshot once for both hosts.
Amended 2026-11-24, on the owner's word: a widget lands with the blog's first, before the form's salary range.
Amended 2026-11-24, on the owner's word: no preview status; a post dated later is a draft until it comes out.
Amended 2026-11-24, with the widget block: a widget is a node of a post's tree and a block of the document, its anchor
named by `compose`, a place a link may name.
Amended 2026-11-24, with the first island: a widget's props reach the document as JSON, and a release refuses a post
showing a widget whose code its build lacks.
Amended 2026-11-24, with the props' schema: a release refuses a post showing a widget with props its schema refuses.

### Context

The first version embedded `profile.yaml` in the binary and parsed it once,
so no post could change without a build (against D5). Each face laid the
content out again by hand: the HTML page, the markdown, the man page and the
TUI each picked the facts and their order, and `content` held presentation
and web routes.

### Decision

- **Content lives on disk.** `content` loads a folder (the profile, the
  posts and their series, the word catalogues), checks it whole and makes an immutable
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
- **A post is read once, by `content`** (amended 2026-11-22), which owns
  the posts' format as it owns the profile's. It decodes a post's front
  matter strictly, parses its body once with goldmark set up as the format
  of FOUNDATION §5.4 says, CommonMark with GFM's tables and footnotes
  alone, and refuses what falls outside it, each fault at its file and
  line. The snapshot keeps the body's source and goldmark's tree, which
  `compose` walks into blocks. The tree is read and never changed after
  the load, by no transformer, attribute or id set later, since two
  goroutines may compose one snapshot (`release.Derived`), and `compose`'s
  tests run under `-race` in a hook of their own. One table in `content`
  (`PostKinds`) names the kinds of node a post may hold: the load refuses
  any other (FOUNDATION §5.4), raw HTML among them, and `compose`'s walk
  has a case for each, a test holding the two to one list, so a kind an
  extension brings is refused, never dropped. Ways weighed: `content`
  keeping the text alone for `compose` to parse, two parses of one text
  and the format set up in two layers; `content` turning the tree into one
  of its own, the document model written twice. goldmark keeps to the
  CommonMark spec and is built to be extended; blackfriday and gomarkdown
  keep to none. The posts stand in their slugs' order and the series in
  their names', as their folders list them. A post's revisions, older
  versions of its text (FOUNDATION §5.4, Revisions), are read with it, each
  its own title, lead and body, held to what the post's text is held to,
  and dated by its file's name from the day the post was published on and
  before the post's current version, which its `updated:` dates (amended
  2026-11-24); a picture one of them alone shows is the post's, served
  with it. `content.RevisionOf` writes the revision of the current version
  that `paranin revise` saves, a day after that version came out at the
  soonest, and `content.UpdatedText` the post's file dated as updated that
  day, its other bytes as written, which `revise` puts in the file's place
  whole, so the post's format is written in one layer.
- **`paranin check`** runs the same loader, of the content or of a whole
  release, so CI refuses what the server would refuse. Both hold the
  content to the site's own check more, `compose.Check`, which the
  composition root hands the release's load since `release` knows nothing
  over it (amended 2026-11-23): no two pages `compose` makes share an
  address `routes` writes, in any format, as a post slugged `series` would
  share the list by series'; and each link inside the site of a post
  leads to a page served at its path and, where it names an anchor, to a
  section of that page, a fault at the post's file and line where not.
  `content` reads a link's form and keeps those inside the site with their
  lines; only `compose` knows which pages are served and the anchors it
  named. The check takes the moment of the load, which the composition
  root reads (amended 2026-11-23): the links of every post served are
  checked over the site as it stands at each moment it changes from then
  on, as each post dated later comes out, and a draft's links and the
  addresses over the site as the author reads it with the drafts once
  every post dated later is out, where it holds the most pages. A link to
  a post that comes out after the one linking it is refused, naming when
  that post comes out, since it would lead nowhere until then; one to a
  post out before it or with it passes. Ways weighed: the check over the site
  as it stands now alone, which sees neither a post dated later nor its
  links; every link to a post dated later let pass, which serves a dead
  link until that post comes out.
- **One document, every face.** `compose` turns content into documents: the
  résumé, a page for each project that has one, each post, the blog's list
  in its two groupings, by year and by series, the index of the
  site that `llms.txt` lists, and the extras of the TUI and the vim mode.
  Every face draws them, the HTML page included, each part in a layout of its
  own; no face picks the facts or their order again. `compose` is a function
  of a snapshot and of its audience (amended 2026-11-23): the moment they
  read at, which a post dated later stays out until, a draft, and
  whether they read the drafts, as the author does under
  `paranin serve --drafts`; the hosts read the clock and hand the moment
  in. It says until when what it makes holds, the moment the next post
  dated later comes out, and each host keeps what it drew until then or
  the next snapshot (`release.Derived`), so a post comes out on time and
  no request composes. Ways weighed, on the fixture's one post: composing
  per request, about 3 ms a request to compose and draw the site, growing
  with every post; composing per snapshot and minute, as FOUNDATION
  planned, a composition each minute for the few moments the site
  changes, and a post up to a minute late.
  A composition reads its snapshot through `compose.Prepare`, which splits
  every post's code into its kinds of word once a snapshot (amended
  2026-11-23). The composition root prepares each snapshot once for both
  hosts: it hands them one `release.Derived` of the content alone
  (`release.DeriveFromContent`), which each reads `Of` the
  snapshot it draws (amended 2026-11-24, the owner's word), so neither a
  second host nor a site composed again as a post comes out lexes the
  code again. A lexer is slow, and chroma's regexps start a clock goroutine of
  their own (regexp2's fastclock), which a test's bubble would see outlive
  it. Ways weighed: splitting the code as `content` loads, which takes
  chroma out of `compose`, against this ADR; the prepared content held by
  `release`, which knows nothing of `compose`; each host preparing its
  own, the code split twice; the check's split kept with the snapshot, a
  type parameter on every part of `release` for one split a load saves.
- **A document is a head and its blocks.** The head holds the page it is,
  where it is one; the language it is written in; its title, a subtitle,
  a place and a lead (the résumé's are the name, the title, where the
  author lives, the headline); on a project's page what it is made of,
  the language it is written in and then its stack, as the profile lists
  them; a description for the page's metadata; and on a project's page
  where its code lives, a repository typed as its address and the site's
  word for it, since every face shows it first; on a post's page its
  number, its tags as filters of the list, its description as its lead, and the version read, a
  closed set (amended 2026-11-23): out, by when it came out, which each
  face writes its own way, or a draft, by the site's word for one, which a
  face writes where it would write the day; and an older version, which
  names the current one and the site's words round its day that lead
  there, which each face writes its own way (amended 2026-11-24). A version
  out, current or older, lists every version of the post a reader may go
  to, newest first, the one read among them, its current one alone where
  it has no older, each by its page, its day and its name: a post's
  page names an older version by its revision's name beside its slug
  (`Page.Revision`), so a version is a page like any other, and its places
  (`version`) are read off its head as a part's are. Ways weighed: an
  older version as a page of a kind of its own, which every switch over
  the kinds would name again for what is a post; the versions a field of
  `PostPlace`, which a draft, with no version out, carries too.
  The head holds too, in one field only a post fills, none on every other page, where it
  stands in the blog: its series, every part with where it stands for
  the post's reader; the posts beside it, which its foot leads to; and
  where it is discussed. `compose` reads them once from every post and
  series of a snapshot, so no face picks a post's neighbours again
  (FOUNDATION §5.4). The ways to reach the author (mail, the profiles,
  `ssh`, the CV) left the head (amended 2026-11-15, the owner's word):
  they are facts of their own kind, in the résumé's last section under the
  site's word for contact, which a face finds by that kind and shows where
  it chooses.
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
  owner's word). Each section of the résumé and of a project's page after
  the opening stands under its title, in a smaller block font where the
  page has room for it, so a search finds a section's title on its page
  only where it reads as text (amended 2026-11-22, the owner's word).
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
  - a widget (amended 2026-11-24, with the widget block), an island a post
    shows as a block of its own: the component of `web/src/widgets` it
    names, its alt text, its caption, and the live one, the post's page at
    the widget's anchor, which every face that runs no island leads to.
    `content` reads the block as the post is parsed, by a transformer of
    goldmark's tree, into a node of a kind of its own that carries the
    widget, which `PostKinds` names and `compose` walks, and holds its props
    to their form. It writes them as a JSON object in their written order,
    all but the alt text and the caption, which the page draws round the
    widget, and the document carries them to the island, their first reader:
    as JSON, since each widget's own types are the page's (amended
    2026-11-24, with the first island). The post keeps where it shows each
    widget, which `release`, holding the build beside the content, holds to
    the widget's code and its props to the widget's schema (amended
    2026-11-24, with the props' schema; ADR-0004). Way weighed: the widgets
    kept beside the tree by their fenced blocks, as a link is, which every
    reader of a post's code would have to tell from code. The form
    and the salary range are widgets too, so `compose` places them;
  - the posts' (step 8): a list, in order from its first number or not,
    whose items hold blocks; a quote, holding blocks; code with its
    language, its text in runs of the kinds of word (below); a pause, a
    thematic break standing between two of the post's own blocks, never
    inside a quote or a list nor at the body's start or end (amended
    2026-11-23); and a figure (amended 2026-11-23), a picture standing as a
    paragraph of its own: its file beside the post, by the page and its
    name, its alt text, its caption, which is the picture's title, none
    where it has none, and its size in pixels. `content` reads each picture
    a post shows from the post's folder as it loads, its bytes, its type and
    its size, into the snapshot with the post, so a reload swaps a picture
    with the text and a page names only pictures its snapshot serves; the
    web host alone serves them, at the post's address. A release's folders
    are read through `os.Root`, one opened for each load and closed after
    it, so no read leaves them, by a link inside among others, and every
    file of a load comes from one release, one switched in by a link read
    whole at the next load. Ways weighed: a picture as a
    span, which no face draws among words; a caption from the paragraph
    after the picture, a convention of no format, where the title is
    CommonMark's own and markdown writes it back as it was; the pictures
    read from disk as a request asks, which a reload would not swap with
    the text. A WebP's size is read by `golang.org/x/image/webp`, which
    depguard keeps to `content`; its cost: the binary grows by 170 KB, and
    its init registers WebP with `image.Decode` for the whole program. Ways
    weighed: reading WebP's head by hand, three forms of it to write and
    keep; leaving WebP out, as AVIF is, whose size Go reads by no package.
    A table gains a header row, and a post's headings nest into sections, each
    holding what follows it up to the next heading of its level or above;
  - the blog's list's (amended 2026-11-23): posts, the block a group of the
    list holds, a section under the group's name, a year's, a series' or the
    posts' in no series: a series' line, how many posts it holds in the
    site's words, and each post it names, out, by its page and when it came
    out, or a part still to come, by the site's word for it, a closed set
    of two, with its title, its lead,
    its part and its series' title where its group does not name it, and its
    tags. Its head is the blog's name, its line and the tags of the posts it
    lists; its page is one of two kinds, the list by year and by series. Ways
    weighed: a record of the list beside no document, which the text faces
    and the TUI could not draw as they draw every other page; a group as a
    block of its own with its title, which would leave the text faces'
    sections and the manual's to write a second time.
    A filtered list (amended 2026-11-24) is the list by year of the posts
    a tag keeps, a page of a kind of its own, its slug the tag, its head
    the list's with the filter in the site's words as its subtitle: every
    face draws it as the list. The page says it filters the list
    (`Page.Unfiltered`), and so that no search engine indexes it
    (`Page.OwnAddress`, `Head.Indexed`), which a face and the web host
    read to send it with `noindex` and name no canonical address, since a
    part of the list is no copy of it. The head of a list and of
    a post holds its filters (`Head.Filters`), which `compose` works out
    once: each tag, where it leads and whether it is the filter shown,
    then named to a screen reader as the way to the whole list; a tag of a
    post the lists name nowhere, a draft's or an unlisted post's, leads
    nowhere, since it may filter none. The faces draw the
    filters in that order and the list's places are its tags among them,
    so no face works out where a tag leads. `compose` composes one for
    each tag, with the site, so no request composes; the web host draws
    each filter's formats the first time a request asks for one, once a
    drawing however many ask at once (`sync.OnceValue`): on 72 posts, 45
    tags and 2 series, drawing them all with the site took 52 ms and 53 MB
    of allocations a snapshot, not 23 ms and 27 MB, where drawing them
    when asked takes 22 ms and 30 MB. Ways
    weighed: the filter a field of every page, which every page written by
    hand would carry; a filter composed per request, a request that
    composes; the TUI filtering the list's buffer in place, a list no
    other face could name; a filter by series too, which repeated the
    series' group on the list by series.
- **The feeds** (amended 2026-11-24, FOUNDATION §5.4, Feeds) are the blog's
  and each series' with a part out, which `compose` composes with the site
  (`Site.Feeds`, `compose.Feed`): each its newest posts the lists name,
  twenty at most, the document of each post's current version, when the
  post came out and when that version did, the start of its `updated:` day
  in the post's own zone, never before the post (`compose.FeedPost`); a
  blog that lists no post has no feed. A feed is a target of its own
  (`document.Feed`), its address `routes.FeedPath`'s, so a post's link
  inside the site may lead to one, and a head names the feeds that follow
  its page (`Head.Feeds`), a series' first where the page shows the series,
  the blog's on the résumé too, which the page's head links for a feed
  reader to find. A feed and each post are named for good by a tag URI
  (RFC 4151), which `routes` writes (`Site.TagURI`) under the site's tag
  authority, the profile's `site.tag_authority`, which `routes.New` takes,
  and the path it is served at, a post by its number's,
  `tag:paran.in,2026:blog/12`, so neither the site's address nor a post's
  folder moving renames them. The media type of each format is named once,
  in `routes`. `face/feed` writes Atom through typed structs `encoding/xml`
  escapes, and takes each post's body from a `BodyWriter` the host hands
  it, `face/html`'s `FeedWriter`, which writes text inside a block as the
  page does, every address whole, and the blocks plain, with none of the
  page's own parts, since one face never builds on another. Ways weighed:
  a template, whose text could slip past the escaping; a library, a
  dependency for a few structs; a post's address as its id, which moves
  with the site; the page's article as the body, whose notes stand twice
  and whose buttons and numbers show in a reader as stray words; the
  summary alone, which a feed standing in for a newsletter cannot be.
- **Translations** (amended 2026-11-24, FOUNDATION §5.4, Translations; the
  owner's model). `content` loads one profile
  and an edition in each language it has words for,
  English's apart from the rest (`Snapshot.Own`, `Snapshot.Translations`),
  so no place in a list says which is English: that language's words, its series and the posts
  written in it, a translation with its original's number, tags and
  discussions and its own text, status, dates and revisions
  (`Post.Translation`). A language's tag is stored once, in its words;
  `Words.Translation` reads it as a page names it, none for English. A page
  names the language it is translated into (`Page.Translation`), as it
  names an older version (`Page.Revision`), so a target, a head and a link
  name a page in any language by one value; `routes` writes every address
  of a translated page in its language's folder, `/ru/blog/<slug>` and a
  feed's `/ru/blog.atom`, and reads a folder back only for a language the
  site is translated into, which `routes.New` takes from the snapshot.
  `compose.SiteOf` composes one site, its index and pages
  once, and a blog in each language (`Site.Own`, `Site.Translations`) by the code that composes
  the English, each with its frame in its words and its own lists, filters,
  posts, versions and feeds, all their links read among the pages
  of every language, so a post may link to a translation; `Check` holds
  every blog's links and addresses together. A post's head lists the
  languages its reader may go to (`Head.Languages`), the one read and each
  the lists name it in, which the page draws as its Switch, its `hreflang`
  alternates and `og:locale:alternate`, and the other faces as a line of
  links, and offers among its places (`document.PlacesOf`), each by its
  tag. A frame names its language once, as a page does
  (`Frame.Translation`, `document.Untranslated` for the site's own), which
  each host keys its frames by and its list's groupings are worked out
  from (`ListWords.Groupings`). Every frame holds the one list of the
  pages `:e` opens in every language (`Frame.Pages`), so the page's `:e`
  and the TUI's open the same pages, and the page names each language's
  blog folder (`routes.Site.BlogFolders`). The hosts draw each blog in its
  own frame: the SSH host hands a session every blog's documents
  (`Blog.Served`) and each frame (`tui.Drawing.Frames`), and the TUI lays
  a page out in its language's frame, opens a page of any language by its
  terminal name, which `routes.Site.BlogNamed` reads a language's folder
  off, and holds the blog's session in the language of the list it shows,
  a post read going where its head's language leads. Ways weighed: the
  language as a kind of page, which every switch over the kinds would name
  again for each kind; a `routes.Site` for each language, whose targets
  could name no page in another, so no Switch; a site and a snapshot per
  language, each carrying a copy of the profile, the series, an index and
  pages none but English has; every language read off a path by its form,
  which read `/de/blog` as a language the site has not; the faces reading
  the words of a document's language off it, which would carry the frame's
  words into every head.
  A translation names the text of its original it was made from (amended
  2026-11-24, FOUNDATION §5.4, A stale translation shows), `source:
  en@<hash>`, the hash of the original's version's text, its title, lead
  and body, the text a revision keeps (`Revision.text`), which `content`
  holds the source to: one of no form is a fault, and one that names any
  other text a warning, a fault of the content that refuses nothing, kept
  in the snapshot (`Snapshot.Warnings`), which `paranin check` prints and
  `paranin serve` logs, and the translation stale (`Post.Stale`). `compose`
  says so in the current version's head (`VersionOut.Outdated`), the site's
  words and the original's page and language, where the Switch offers the
  original, and each face writes it under the day, as it writes an older
  version's way to the newer. Ways weighed: the source as a hash of the
  file whole, which a status or a day changes; the note a version of its
  own kind, which every switch over the versions would name again for what
  is a version out.
  A series is read in each language (amended 2026-11-24, `Edition.Series`), its own
  words where its translation's file writes them, its original's where
  none does, so `compose` reads a blog's series as it reads its posts. The
  site's words in a language are its catalogue's alone: a count
  (`content.Count`) is written in each plural form the language's CLDR rule
  gives a whole number, which `content` reads from
  `golang.org/x/text/feature/plural` with the language's tag, depguard
  keeping x/text to `content` and to those two packages; the months a day
  writes, which the frame's `DayWords` writes as `DESIGN.md` has a day,
  for the page and the TUI alike; the manual's headings, which the frame
  carries (`Frame.Manual`) and the web host hands `face/text` with the
  site's addresses; and the frame of the page of no page in a language's
  folder, which the web host draws for each language. Its cost: the binary
  grows by 0.36 MB. Ways weighed: a table of plural rules of our own,
  each language a grammar copied by hand and waiting on code; the series'
  words in the catalogue, which every catalogue would have to hold for
  every series; the manual's words in the document's head, which carries
  the frame's words into every head.
- **Text inside a block** is a tree of spans, a closed set too: emphasis,
  strong, a link, a highlight (`==result==`) and a note hold spans; text,
  code and a gap (`[text]`, still to fill) are leaves. A gap may sit inside a
  highlight, as the profile has it. The profile's marks and a post's markdown
  become the same spans.
- **A block of code is runs of kinds of word** (amended 2026-11-23): its
  text whole in runs, each one kind of a closed set, a keyword, a string, a
  number, a function's name, a type, a builtin, an operator, a variable, a
  comment, punctuation or plain, so a face colours each kind from the tokens
  and none reads a language. `compose` splits the text by chroma's lexer of
  the language, the highlighter Hugo uses, for the languages one table of
  `compose` names, each held by a test, since a lexer runs whenever a
  snapshot's posts are composed, at every load by `compose.Check`, in
  `paranin check` and on a reload, and again as the hosts draw the
  snapshot, and some never end on text they do not expect; code in any
  other language, or in none, is one plain run. depguard keeps chroma out
  of every part but `compose`, and lets `compose` take its lexers alone. Its
  cost: the binary grows from 18.4 MB to 22.2 MB, and `paranin check`'s
  peak memory from 13.2 MB to 19.5 MB. Ways weighed: Shiki in the browser,
  which a reader without script, the manual and the TUI would never get; a
  lexer of our own for the languages posts hold, a second highlighter to
  write and keep; chroma's own HTML, drawing in `compose` in colours
  outside the tokens.
- **A footnote is a note where a line cites it** (amended 2026-11-22): a
  span holding its number, the order of the first citations, and its words,
  alike at each citation. A face sets it beside the line, as the page and a
  wide TUI will, or gathers the notes at the end (`document.NotesOf`), as
  markdown and the manual do and a narrow TUI will, so no list of notes is
  kept in step with the blocks. A note is one paragraph and cites no other,
  which the load holds a footnote to: a source set beside a line reads as a
  line, and a span holds no blocks. Ways weighed: a block of notes at the
  document's end, each citation a span naming its number, which keeps two
  lists in step and leaves a face that sets notes beside the line to find
  each one; a note holding blocks, which puts blocks inside a span.
- **A page's places are read off its document** (amended 2026-11-23): the
  places its command line goes to by name beside its tabs and the projects'
  pages, a post's headings and its series' other parts out, the posts out
  the blog's list names, its tags and their series (`document.PlacesOf`), as its tabs
  and notes are, so no list of places is kept in step with the blocks. A
  heading's section carries its anchor, which `compose` names once, by the
  heading's words, counted only where an earlier heading of the post took
  them, so an anchor holds whatever else the page comes to name alike
  (amended 2026-11-24); a series' group carries the series' name, and a post
  the list names its series' name. Every face reaches a section by that
  anchor: the page writes it as the section's id, and the TUI keeps the
  line it stands at. The page's own ids, its menu's, its hints', its
  paragraphs', open with a mark `SlugOf` never writes, so no anchor takes
  one. A widget carries its anchor too (amended 2026-11-24), which `compose`
  names once, by the widget's name, counted where an earlier widget of the
  post took it, which the page writes as the widget's id and a link inside
  the site may name, as a heading's: it opens with a capital, which no
  heading's anchor holds, and with no mark of the page's own ids. Way
  weighed: an anchor by the widget's place among the post's, which a widget
  written before it renames. The kinds of place are `document`'s words, and
  `compose` hands every face the pages `:e` opens in one list, by every name
  it reads, a post's number among them, each saying whether Tab offers it.
  A name two places share stands once where both lead to one place, as the
  post read and the list's own post do, and for each, by its kind, where
  they lead apart, as a tag and a series named alike do, or a heading
  named as a window, a project, a part, a version or the TUI's `help`
  (amended 2026-11-24); Enter goes to the one Tab chose, or runs the
  command the line names, or goes to the first. A place is no anchor, so
  none takes a count. Ways weighed: the places a field of the head, which
  kept a second list beside the blocks; each face naming its anchors,
  which named a heading twice and let the page rename one its own ids
  took; a heading's anchor kept clear of the page's other places, which a
  new part, project or version named alike renamed, breaking every link to
  it.
- **A period is a type**: a start and an end, each a month or a gap, and the
  end may be current; the kinds of month are a closed set too. Each face
  writes a known month its own way (`2024`, `01/2024`); the end of what still
  goes on carries the site's word for it, `present`, which every face writes
  as it is. `document` imports nothing of `content`, so `compose` carries the
  months across.
- **A link names its target**: an address outside the site, a mailbox by
  its address, which a face may part at its at and `routes` writes under
  `mailto:`, a file served beside a page, a post's picture, by the page and
  its name, in any format (amended 2026-11-23), or a page of the site by its
  kind and slug, and the tab of it shown alone where it names one, with a format
  where the link names one (the CV as PDF, `index.md`, `llms.txt`), or a
  section of a page, by the page and the section's anchor, its own target
  beside the page (amended 2026-11-23), so a page stays the one value a
  head and a link both name. A section's anchor is the one a link names.
  A post's link inside the site, written as a path from the root, is read
  back into the page served at that path, or into a section of it, by a
  table of the served pages by their paths, which `compose` builds from
  the addresses `routes` writes, so no reading of a path is kept in step
  with `routes`. `content` reads each such link once, at the load, and
  keeps it by its node. `compose.SiteOf` composes a snapshot's pages for
  an audience, its lists, its frame and then its posts once, keeping the
  table, and both hosts draw on that one composition, `compose.Check` on
  one for each moment it checks. Ways weighed: the anchor
  a field of the page, which every comparison of a page with a head's
  would have to leave out; `routes` parsing a path back into a page, a
  second copy of its paths; a list of the served pages made apart from
  their composition, a second account of which tabs a page has; the
  table built on its own and handed beside the documents, which composed
  the pages once for it and again for the hosts. The kinds
  of page are `document`'s type; `routes` builds on `document` and writes
  a target's address in each format (`/p/claude-review`,
  `/p/claude-review.md`, `/p/claude-review/what-it-does.md`,
  `/experience`, `/blog/<slug>#<heading>`), so the markdown face
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
  with `face/html`, the CV among the ways with `face/print`; a link's
  format with the first link that names one, and a section of a page as a
  link's target with a post's links inside the site (2026-11-23); project
  pages and the index with `face/text`, the first face to serve them; the
  extras with `face/tui`, and the head's greeting with its redesign
  (2026-11-06), a project's tags and repository with its page
  there (2026-11-07), the entry's and the facts' kinds and the five
  sections with the owner's word (2026-11-15); the records for JSON-LD,
  the person's and a project's, with `face/html`, the feeds with
  `face/feed` (2026-11-24), and the sitemap's dates with the sitemap, a
  post's and a list's alone (`docs/SEO.md`, Crawling); a widget with
  the blog's first (FOUNDATION §8, step 8), drawn by every face with no
  script (2026-11-24), its props with the first island; the posts' blocks
  and spans, but the figure, and a post's head with the first post served,
  in the text faces (2026-11-22); the HTML face's and the TUI's drawings of
  them with each face's post, the figure with the post's pictures, in every
  face (2026-11-23); the
  blog's list with its page and its text faces (2026-11-23), the TUI's with
  its blog session.
- **`routes`** writes every address. A filtered list is the site's one
  address with a query, `/blog?tag=go` (amended 2026-11-24), and `Requested` reads a
  request's path and query into the address `Path` writes, by the one
  encoding both use, so the host only looks the address up: on the list
  by year the tag the query names, every other key passed over; a tag
  named twice, an address no page has. An older version of
  a post stands in the post's folder by its revision's name, a day, its
  formats beside it, `/blog/<slug>/<date>.md` (amended 2026-11-24), and a
  picture by its name, which ends in its type's ending: neither reads as
  the other. A page says whether a search engine indexes it
  (`Page.OwnAddress`, `Head.Indexed`), which a filtered list and an
  older version do not, so the web host and the page read one rule. `compose` composes each older
  version of a post out with the site (`Site.Versions`), its links read
  among the pages served, a link to it among them, and the web host draws
  each the first time a request asks for it, as it draws a filtered list.
- **The layers**: `content`, `document`, `look` and `assets` (the page's
  build) at the foot, side by side; `routes` on `document`; `compose` on the
  foundation, and on `routes`, through which it finds two pages at one
  address and reads a post's links inside the site back into targets;
  `inbox` on the foot alone, beside both; `release`, which holds the
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
- **Status**: `draft` (never shipped), `published`, `unlisted`. A
  `published` date in the future keeps a post a draft until then, worked
  out from the time, the site composed again as the post comes out
  (ADR-0003; tests hold it in `testing/synctest`; FOUNDATION §5.4,
  Status).
- **One front-end build**: one Vite build takes the site's entries and every
  post's widgets, and writes hashed files, one manifest and each widget's
  props as JSON Schema into the release folder. Go serves every front-end
  file from there and embeds none. Loading a snapshot checks the manifest
  holds every entry the pages name, the code of every widget a post shows
  (amended 2026-11-24, with the first island), and every post's props
  against their schema. Development runs `vite build --watch` into the
  same shape, so the strict CSP holds there too.
- **Props' schema** (amended 2026-11-24): a widget's schema is written once,
  in valibot, beside its component (`PageCache.props.ts`), which infers
  its props' type from it. A plugin of the build (`web/plugins/`) turns
  each into JSON Schema, draft 2020-12, written by the widget's name into
  `props.schema.json` beside the manifest, which names no such file, so the
  site never serves it. It refuses a schema by which a typo would pass, an
  object not strict, and a default, which no reader applies. `assets`
  compiles each schema as the build loads, by santhosh-tekuri/jsonschema,
  reading no schema from disk, and `release` holds every post's props to
  it, a fault at the widget's line. The island takes the props as they
  stand: the release held them already. Ways weighed: a parse in the
  browser, valibot in each widget's chunk to check props checked once;
  a schema in the component's module script, which the build would run
  Svelte to read; xeipuuv/gojsonschema, with no release since 2019, and
  kaptinlin/jsonschema and google/jsonschema-go, younger and before 1.0,
  where santhosh-tekuri's, at 6, passes the JSON Schema test suite of
  every draft but its optional tests.
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
  SIGHUP keeping the old state and a language per subpath are each common
  practice ([FOUNDATION](FOUNDATION.md) §5.4 names
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

**Status:** accepted (2026-11-01); amended (owner, 2026-11-25): the build
draws the cards. [FOUNDATION](FOUNDATION.md) §5.2, D10.

### Context

The first version printed `cv.pdf` and captured `og.png` with headless
Chrome from a running server, by hand. Phase 5 makes a CV per job
description on request, and every page and post needs a card of its own.

### Decision

- **Typst** draws `cv.pdf`, the tailored CV and a social card per page and
  post, as PDF and PNG, with JetBrains Mono from the repository.
- **The cards are drawn by the build**, beside the page's (ADR-0004): a
  command of ours composes the site as a release would and has Typst draw
  each card into a folder of its own, which the release holds and
  `hosting/web` serves, each at an address that changes with what it shows,
  so a cache keeps it for good and a preview fetches a changed card anew.
  It draws a card for every page that will show a preview, a post dated
  later among them, which comes out on the clock with no build (ADR-0004);
  loading a snapshot checks each has its card, as it checks every widget
  has its code. CI draws them on every push; development keeps them as it
  keeps the page's build, by the same command watching. The server draws
  nothing and runs no Typst.
- **A card is a template and its facts**: the template, in Typst, lives in
  the repository and holds the look (`DESIGN.md`, Social cards); `face/card`
  hands it a card's facts as data, its title, its line, its day and series,
  its way in, written from the document model like any face's. Typst
  measures the words it sets, so the template, not Go, fits a title to its
  lines.
- `cv.pdf` is drawn when a snapshot loads, into a disk cache, by a worker,
  never inline in the reload. A tailored CV is drawn on request by a
  bounded worker with a per-IP limit, and Typst, pinned in the runtime
  image, comes to the server with it.
- Typst is pinned in CI and in development; the command refuses another
  version.
- The text of the PDF is checked by a test, as `pdftotext` checked it.

### Rationale

- One tool for PDF and PNG, with no browser in production or in the build.
- A card shows only what the content says, so it is a build's output, as
  the page's files are: the server reads a finished folder and builds
  nothing (ADR-0004), with no worker or cache for the cards.
- A template is Typst written once and read as Typst; facts as data need no
  escaping of what an author writes into Typst's markup.

### Alternatives considered

- **Headless Chrome**, as the first version did: heavy for a request in
  production, and a browser in the path of every card. Rejected.
- **Drawing the cards with a Go image library.** Possible for cards, but the
  PDF would still need a second tool. Rejected.
- **The server draws the cards** when a snapshot loads, into a disk cache,
  by a worker, or on a card's first request into memory. Either puts Typst, a worker and a cache on the server for
  what the build can draw once. Rejected (owner, 2026-11-25).
- **Go writes each card's Typst markup**: every author's word escaped into
  markup, and the look split between Go and Typst. Rejected.

### Consequences

- A second program in the runtime image, with its version to pin; a new
  version can move the layout, so its text check and a look at the PDF come
  with every bump.
- Typst in every developer's setup, for the cards; a bump of it is looked
  at on the cards as on the PDF.

---

## System overview — one source, many faces

> Of the sections from here to the deployment, the configuration's last two rows still
> describe the first version, kept on the owner's machine while their replacements land;
> the rest holds today.

```
cmd/paranin
  hosting/web · hosting/ssh · hosting/indexnow
    face/html · face/text · face/feed · face/tui
      compose · routes · release
        document · content · look · assets
```

Each part builds only on those below it, as ADR-0003 draws them; what each
does is in the project layout below. A face draws documents and never reads
the content they are made from, nor the release a host hands it; `release`
stands beside `compose` and `routes`, `compose` building on `routes` to find
two pages at one address, and the hosts bring them together. The parts
still to come land with their steps of [FOUNDATION](FOUNDATION.md) §8, where
§5.2 places them: `face/card` and `face/print` beside the faces, `inbox`
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
(SEC-7); Wish forwards each change into the program as `tea.WindowSizeMsg`. A
session's input reaches the program at a reader's pace, its flood waiting in the
channel's window, and its environment takes only the variables the colour profile
reads, each once and short (SEC-7).

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
                                  │  own address: Accept, else agent →   │
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
  icon its pages link, in SVG and in PNG, a square of the name's gradient;
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
  The islands come with the blog's first widget; the cards, which the
  build draws with Typst, in a folder of their own; the CV with hosting
  (ADR-0005).
- Every answer tells a browser what it may do: a strict
  `Content-Security-Policy` and `nosniff`, set where every request passes.
- The contract between the server and its readers is [WEB](WEB.md): each
  address and what it answers, the headers, and what the page does in a
  browser.

## Project layout

```
cmd/paranin/main.go      # check the content, or serve it with the build: both servers, IndexNow, the signals, the stop
internal/hosting/indexnow/ # tells IndexNow each address that changes, as a reload takes or a post comes out
internal/hosting/ssh/    # Wish server construction, middleware, security policy
internal/hosting/web/    # HTTP: each page drawn once a snapshot, the format the reader's Accept ranks, else its agent (D21)
internal/face/tui/       # Bubble Tea model: a buffer per document, its tabs laid out and painted
internal/face/html/      # templ pages: each document, its record as JSON-LD, the page of no page, the icon
internal/face/text/      # markdown, the manual for curl, llms.txt, the page of no page as text
internal/face/feed/      # what a machine polls: the blog's feeds in Atom, the sitemap, robots.txt
internal/compose/        # content → documents and their records
internal/routes/         # every address of the site
internal/release/        # the content and the build as one snapshot: load, the store, what is derived once, the watch
internal/document/       # the face-neutral model
internal/content/        # the content folder: load, check; a post's markdown parsed once
internal/look/           # the design's colours, generated from design/tokens.yaml, and its block fonts
internal/assets/         # the page's build: its manifest and files, read and checked
internal/golden/         # the golden files' helper, for tests alone
internal/fixtures/       # the shared fixtures, the content's and the build's, for tests alone
scripts/tokens/          # the generator of internal/look and the page's tokens.css, and its check
scripts/templates/       # templ's format and generated Go, checked or written
scripts/coderules/       # the code rules no linter here holds
scripts/check_instruction_files.py  # claude-review's gate of the instruction files
scripts/cut_extrabold.sh # the mail's ExtraBold, cut to a mail's characters from the face as released
content/                 # the site's content: the profile, the site's words
design/fonts/            # the faces, each under its own licence, as released: those the site serves, the ExtraBold its cut is made from, and, to come, the TTF Typst reads, which reads no woff2
design/tokens.yaml       # the design's colours, which internal/look and tokens.css are written from, and the page's mixes of them, faces, text styles, lengths, times and layers, which only the page takes
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
| `PARANIN_INDEXNOW_KEY` | unset | the IndexNow key, 8 to 128 of `a–z`, `A–Z`, `0–9` and `-`: served at `/<key>.txt`, and IndexNow told each address that changes; none under `--drafts` |
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
