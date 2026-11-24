# TODO

The order of work is `docs/FOUNDATION.md` §8; each item is one chunk, started on the owner's word. An item done is removed, not ticked off.

## The foundation

- The faces, each its own commit (§8, step 5): the rest of `face/tui`: `:hire` once the form and the CV are there, since without them it repeats the ways the résumé opens with. Each reads the first version under _old/tree while the face is written, or in the archive where _old/tree lacks it (the TUI), and deletes what it replaces; the archive `paranin-first-version-2026-11-01.tar.gz` beside the repository keeps it whole.
- The page (§8, step 6), each part landing with its first user (D18), in this order:
  - The first island, the form's salary range, waits for `inbox` (step 10). It brings Svelte, and svelte-check in place of the hook `page-types`'s `tsc`, and widens stylelint's glob to its components' styles, which the build writes to files, never into the page, since the page's policy refuses an inline style (`docs/WEB.md`, "Security headers"); its script loads by a dynamic import, not with the site's entry, which holds the first paint.
- At 53.75 to about 60rem wide the foot line wraps to two rows, more than the command line's menu keeps back (`row_3`), so in a window under about 175px tall the menu's top row is cut. Keeping back the line's real height would need the script to measure it; rare enough to leave until a reader meets it.
- A tab's name keeps every letter of its title, so a site in another script serves its tabs at addresses in that script.
- At 256 colours Latte's `code_builtin` and `code_variable` fall to one index, 161, and Macchiato's `code_number` and `code_operator` to 116, so the TUI's code shows those two kinds alike there: `scripts/tokens` picking distinct indices for the code colours of a theme would part them.
- A post's markdown and manual write neither its series, nor the posts beside it, nor where it is discussed, nor its other versions, which its document holds and its page and the TUI draw; an older version's write the newer alone.
- The document keeps no list's tightness, so a post's loose list of one paragraph an item comes out tight in its markdown.
- The manual writes a link outside the site by its words alone, so a post read with `curl` loses where its links lead: `docs/FOUNDATION.md` §5.4 plans numbered links in a terminal. The TUI's post leads by OSC 8 alone, as the résumé does, which a terminal that takes none drops: the numbered links would serve both.
- The blog (§8, step 8), as `docs/FOUNDATION.md` §5.4, D23 to D27 and `DESIGN.md` have it; its look was approved on 2026-11-22. Built, each its own chunk, in this order: the first widget, with Svelte and `kit`; the rest of translations (§5.4, Translations), each its own chunk: the TUI's language, `:set lang=ru`, with which a session opens a translation in a buffer and draws its frame in the language's words, its days in its months and the note of a translation older than its original among them, where now it reads the site in English and leads to a translation's address on the web; the résumé in another language, `profile.ru.yaml`, with which `/ru` and the index in a language come.
- A draft's own mark, drawn on a canvas for the owner to pick; until then its version reads `draft` (`DESIGN.md`, PostHead).
- A filtered list is a page of a kind of its own (`PageTagFilter`, `PageSeriesFilter`), where a post's older version and its language are fields of its page (`Page.Revision`, `Page.Translation`): a check of a page's kind by `==` escapes `exhaustive`, so weigh making a filter a field of the list's page too.
- The series filter's chip and the filter's way out, drawn on a canvas for the owner to pick; until then the chip is the reversed tag's form with the series in the site's words, and the way out the Link's underline under the pointer (`DESIGN.md`, BlogHead).
- A diagram, as `DESIGN.md`'s Article draws it, an SVG in `face` with boxes on `sunk`, cannot be a post's picture: a page shows a picture by `<img>`, in which a drawing loads no font and reads no token, so it keeps neither the page's face nor a theme's colours. Drawn into the page it would, which the load's refusal of raw HTML and the page's policy refuse today; how a diagram is written and drawn is the owner's word. A drawing in dark ink with no ground of its own vanishes on the dark theme's ground: whether a picture gets a ground of its own, or the writer gives it one, is his word too.
- A page, a feed and a post's picture are sent with `no-cache` and no validator, so a cache that asks again gets the whole answer each time: an `ETag` worked out once a snapshot, answered with `304`, would send only what changed. A feed reader polls a feed, up to twenty posts whole, again and again, so this lands before hosting at the latest.
- SEO and Open Graph for the whole site (D27, `docs/FOUNDATION.md` §5.4): `robots.txt`, naming `llms.txt` in a comment; `sitemap.xml` with every page's, tab's and post's address and `lastmod`, a post's from its current version's day, its `updated:`, written by `face/feed` beside the feeds; `BreadcrumbList`; a post's `BlogPosting` and `og:type` `article`, which its page writes none of yet; `noindex` and `301` where §5.4 names them; each page's card as `og:image`, which needs Typst (ADR-0005), so when the cards come, before hosting or with it, is the owner's word.
- The content, changed together with the owner (§8, step 9).
- Hosting, last (§8, step 10): SEC-6 of `docs/REQUIREMENTS.md`; the files of the snapshot before kept served beside the current ones, so a page fetched before a reload finds its stylesheet and chunks after it (ADR-0003), and with it whether a release switched in by the link and SIGHUP is watched too, since the watch would load the same release a second time and push out the snapshot before; `inbox` with delivery to Telegram, and the form the first screen's primary button opens, which does nothing until then; `face/print` on Typst, the CV listed in `llms.txt` with it, as the first version listed it, among the foot line's ways, its look the owner's to pick then, and before the profiles that end the first screen's brief (DESIGN.md, FootLine and Pages), which leave it out until it is served; the Dockerfile with both ports; the preview key on the server, a file `PARANIN_PREVIEW_KEY` names, kept as a systemd credential or a container's secret, never in the image, the repository or CI, and a proxy before the server that logs no path of `/preview/`, which holds a link's token (`docs/FOUNDATION.md` §5.4, Status); deploy, with SEC-8, SEC-10's backup and SEC-11; the crawlers' `Accept` read off the server's log, for the owner to settle D21 against SEO (`docs/FOUNDATION.md` §5.4, SEO and Open Graph).
- Nothing moves the pins forward: the actions' commits in `.github/workflows/gates.yml`, the hooks' revs and Python packages in `.pre-commit-config.yaml`, the hashes of `.github/pre-commit-requirements.txt`. Dependabot for actions, pip, npm and gomod comes with the repository's remote, at hosting, since SEC-12 of `docs/REQUIREMENTS.md` asks for dependency updates in CI.
- Phase 3 (§8, step 11).
- A review of every test against `.claude/rules/tests.md`: which named behaviour each holds, and whether another test already holds it; a list to delete, merge or keep, for the owner, before any is changed. Its place in the order is the owner's.
- The name rule of `scripts/coderules` lets `conn` and `conn2` pass: "conn" is not among the abbreviations it refuses.
- SEC-13's trail lacks the SSH server's other refusals: a second session on a connection, a subsystem, a connection closed for starting no session in time, a handshake that failed. The log records only a session's refused requests.
- `internal/hosting/ssh`'s tests of a stop and of a panic read the log's words (its messages, `remote=`), which `.claude/rules/tests.md` refuses: the recording handler the other log tests use would read their values.
- `scripts/tokens`' table of faulty files of tokens tells each fault by its words, which `.claude/rules/tests.md` refuses: each fault a kind of its own, as the least contrast's are, would let the table read which fault came.
- The command line's tests open each page of a test in jsdom's one document, so the line of a page opened before it still listens there and answers the keys too; a test of the page opened next passes while both answer.
- A window narrowed to a phone's width hides the vim keys' switch, and with it the keys, while a search kept stays marked there with no key to put it away; nothing watches the width, which lives in the stylesheet's media query alone.
- `page.templ`'s page and `notfound.templ`'s each write the whole frame round a page by hand, the doctype, the language, the head's opening, the main part and the foot line, so a part of the frame added to one can be left out of the other, as the toast was: one frame template taking the head's extras and the main part, as templ's own docs build a layout, would hold it once.
- The theme and the vim keys' switch read the reader's choice once, as a page opens: a page the browser shows again from its cache keeps the choice it opened with, though a page opened after it changed it.
- `internal/hosting/web/web.go`'s `rankedFirst` calls `quality` once per format, each splitting and parsing the whole Accept header again: parsing it once a request into its ranges would work it out once (`.claude/rules/code.md`).
- `design/tokens.yaml` writes a Catppuccin hue again for each role it plays, as Macchiato's blue for `label`, `time_bar` and `code_function`: the palettes' hues named once per theme, which the role tokens refer to, would write each value once.
- `scripts/tokens` holds each colour's least contrast in true colour alone, so at 256 colours `host` (index 98) reads 3.89:1 and `error` (161) 4.46:1 on the TUI's `sunk` panel, and some letters of the name's gradient fall to grey in Latte: a gate over the 256 shades the generator picks would hold them.
- The mail's ExtraBold, `JetBrainsMono-ExtraBold-ASCII.woff2`, is a file derived from the released face that no gate holds to its source: `scripts/cut_extrabold.sh` writes it, and nothing checks it. A hook running the cut in a check mode, which cuts into a copy and compares, with fonttools and brotli pinned as the instruction files' hook pins its packages, would hold it. On the owner's word, since it is a gate of our own.

## Documents

- `docs/ARCHITECTURE.md`'s configuration lists `PARANIN_MESSAGES_PATH` and `PARANIN_TRUST_PROXY`, which nothing reads yet: rewrite them with `inbox` and SEC-6 (§8, step 10).

## Content

- The content holds the site's words in English alone, `content/i18n/en.yaml`: `ru.yaml`, every word of it, comes with the owner's first Russian post, as the tests' fixture holds one.
- A fault of a YAML file, a post's front matter among them, names its field by a path written by hand beside the yaml tags (`experience[0].end`), where a fault of a post's body names its line. yaml.v3's `Node` knows the line and column of each value, which would name the place (`profile.yaml:171:`) with no second copy of the schema; every file is parsed into a `Node` already, for the check of its characters, and a fault has its line.
- Content's faults are told apart by their words in tests (`TestEveryFaultSaysItsFieldAndProblem`), which `.claude/rules/tests.md` refuses: a kind for each fault would let a test read which fault came.

## Before the design work

- Measure again what `docs/FOUNDATION.md` §3.4 says of the page (heights, the overflowing link, the date column at 390px), with the shots kept beside the numbers.

## The owner's

- The translations' model (`docs/FOUNDATION.md` §5.4, Translations, ADR-0003): a blog of its own in each language, `/ru/blog` listing the posts translated alone, chosen with its ways weighed; his to confirm.
- On the Russian list the blog's name stands as plain text in its `h1`, since the titles' block letters draw no Cyrillic (`list.templ`'s `titleHeading`, `internal/face/html/letters.go`): Cyrillic capitals in `internal/look/glyphs.go`, or a text title in the soft gradient, picked on a canvas.
- Nothing leads between `/blog` and `/ru/blog` but a post's Switch: whether the list's head offers the other languages' lists, and how, is his eye.
- On a phone the Switch's links stand 16px tall, under the 24px WCAG 2.5.8 asks for a target: the Switch's own shape, the list's groupings' too.
- The note of a translation older than its original (`DESIGN.md`, PostHead) stands in an older version's form for the way to the newer, `↑ старше оригинала`: its own form, drawn on a canvas for the owner to pick, as the draft's mark and the series filter's chip wait to be.
- For his real `ru.yaml`: CLDR's `one` takes 21, 31 and on as well as 1, so a count's `one` form reads for both, `все {} часть` as `все 21 часть`; and a day writes its month short, `15 нояб. 2026`, which his `months` may write whole, `ноября`.
- The Switch names a language in its own words, `English` and `Русский`, a word of each catalogue (`language_name`); short tags, `en · ru`, are the same word written otherwise.

- The tag authority's date (`site.tag_authority` in `content/profile.yaml`, `paran.in,2026` until he answers, which RFC 4151 reads as 2026-01-01): a day he held paran.in; ids never change once readers subscribe.
- A published post whose `updated:` names a day still to come gives its feed an `<updated>` in the future, since the load reads no clock and nothing holds the day back: his eye, whether the check refuses such a day for a post out, or the feed waits for it.
- `tint_strong`'s values (`design/tokens.yaml`, the way from an older version of a post to the newer, under the pointer) are the author's: in Latte the step from `tint` reads 1.13:1, and in Macchiato `text` over it 4.52:1, which no gate holds; a gate on text over a blend is his word.
- A post's head gives the day of its version read, its `updated:` where it says one, while the list's tiles and the panes give the day it was published, and that first day stands only inside the versions' Pick: his eye.
- The blog's line, "What I build, and what I learn building it." (`blog_line` in `content/i18n/en.yaml`), is not the owner's own: his words, later.
- Nothing links to `/blog?series=`; a reader reaches it only by its address: lead to it, or drop it.
- Fill the profile's placeholders (`[later]`, `[remote?]`, `[notice period]`, `[service]`, `[N]`, `[20XX]`): no code, and it blocks any promotion.
- In Latte the TUI's menu writes its kinds in `muted` on `quiet` at 4.05:1, under the 4.7 `muted` is held to over its grounds, among which `quiet` is not: another ground for the menu, or another ink for its kinds. His eye.

## Upstream

- `sessionContext` (`internal/hosting/ssh`) works round two faults: charm.land/ssh's `Session.Context()` is the connection's, not the session's, and wish's window loop (`bubbletea/tea.go:66-87`) reads its window changes without checking the channel is open, so it spins once a session closes. The type goes when either is mended.

- charm.land/ssh writes a session's window on each window-change (`session.go:412`) with no lock, while `Pty()` reads it: our `open` and wish's `MakeOptions` both read it as the session opens, and `go test -race` reports it. Reading the size from our own bounded pty-req would take away our read but not wish's, so the fix is upstream. Nothing here runs `-race` but compose's tests, by a hook of their own.

- ultraviolet's terminal reader gathers a bracketed paste with no bound until it closes (`terminal_reader.go`), so a paste never closed grows for as long as input comes: here our channel handler paces a session's input (`pacedChannel`, `internal/hosting/ssh`).
- charm.land/ssh keeps every variable of the environment a session is sent (`session.go:341`), however many and however long: here our channel handler takes only the colour profile's, each once and short (`checkedChannel`).


- Bubble Tea writes a panic it catches and its stack straight to standard error, and wish's middleware logs the error `Run` ended with through charm's own logger: neither takes ours. Our log holds a panic of the TUI's model (`internal/hosting/ssh`), but a panic in a command Bubble Tea runs reaches theirs alone. A way to hand Bubble Tea a logger, or a hook for the panic it catches, would make slog the one log ADR-0001 asks for.
- claude-review's `scripts/check_instruction_files.py` takes a word for a path by `/` or by a suffix it knows, and `.go` is not among them: a bare `main.go` in a rule is never checked. Proposed there; copied back here whole.
- The same script finds a file by the file system, ignored folders included: on the owner's machine `**/*_test.go` was matched by the old tree's tests before this repository held one. Asking git for its files would match only what a clone has.
