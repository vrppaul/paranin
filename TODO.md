# TODO

The order of work is `docs/FOUNDATION.md` §8; each item is one chunk, started on the owner's word. An item done is removed, not ticked off.

## The foundation

- The faces, each its own commit (§8, step 5): the rest of `face/tui`: `:hire` once the form and the CV are there, since without them it repeats the ways the résumé opens with. Each reads the first version under _old/tree while the face is written, or in the archive where _old/tree lacks it (the TUI), and deletes what it replaces; the archive `paranin-first-version-2026-11-01.tar.gz` beside the repository keeps it whole.
- The page (§8, step 6), each part landing with its first user (D18), in this order:
  - The first island, the form's salary range, waits for `inbox` (step 10). It brings Svelte, and svelte-check in place of the hook `page-types`'s `tsc`, and widens stylelint's glob to its components' styles, which the build writes to files, never into the page, since the page's policy refuses an inline style (`docs/WEB.md`, "Security headers"); its script loads by a dynamic import, not with the site's entry, which holds the first paint.
- At 53.75 to about 60rem wide the foot line wraps to two rows, more than the command line's menu keeps back (`row_3`), so in a window under about 175px tall the menu's top row is cut. Keeping back the line's real height would need the script to measure it; rare enough to leave until a reader meets it.
- A tab's name keeps every letter of its title, so a site in another script serves its tabs at addresses in that script.
- The TUI's code blocks framed, their language in the edge, coloured as the page's (`DESIGN.md`, The TUI): they stand unframed in `muted`, a block's language not shown, and a tab in a post's code reaches the terminal as it is, where the manual writes the spaces to its stop (`look.TabStop`). At 256 colours Latte's `code_builtin` and `code_variable` fall to one index, 161, and Macchiato's `code_number` and `code_operator` to 116: the TUI's code colours must part them. With the TUI's post, whose SSH tests then compose posts: chroma's regexps start a clock goroutine of their own (regexp2's fastclock) that no synctest bubble outlives, so the release those tests open is composed outside their bubble.
- A post's markdown and manual write neither its series, nor the posts beside it, nor where it is discussed, which its document holds and its page draws. With the TUI's post, which draws them too.
- The document keeps no list's tightness, so a post's loose list of one paragraph an item comes out tight in its markdown.
- The manual writes a link outside the site by its words alone, so a post read with `curl` loses where its links lead: `docs/FOUNDATION.md` §5.4 plans numbered links in a terminal. With the TUI's post, which needs the same.
- The blog (§8, step 8), as `docs/FOUNDATION.md` §5.4, D23 to D27 and `DESIGN.md` have it; its look was approved on 2026-11-22. Built, each its own chunk, in this order: the rest of the page's post: its headings and its series' parts among the places of the command line's `:` menu (`DESIGN.md`, The TUI); the TUI's blog session, its list (`Posts` in the TUI's layout draws nothing yet), a post at `:e blog/<slug>`, `:e blog/<number>` and the SSH user name `blog/<number>`; a post's pictures served from its folder, and the figure (ADR-0003), which the load refuses until then; a post's link inside the site read back into a page of it, which the load refuses until then; statuses and dates (drafts, preview links, publishing on a date), with which a draft's or a preview's own page shows its series, itself the part read, so a preview link of part 3 says it is part 3; a tag or a series filtering the blog's list, at `/blog?tag=` and `/blog?series=` (`docs/FOUNDATION.md` §5.4), sent with `noindex` and naming `/blog` as canonical, the filter reversed in the list's head, with which the head's tags and a post's lead to the list they filter (both stand as words until then), and a tile's series to its series' group; `llms.txt` listing the blog's list and its posts (`docs/FOUNDATION.md` §5.4, Status); revisions, with which a post's version is its Pick; the feeds; the first widget, with Svelte and `kit`; translations, with which a post's head shows its language's Switch.
- SEO and Open Graph for the whole site (D27, `docs/FOUNDATION.md` §5.4): `robots.txt`, naming `llms.txt` in a comment; `sitemap.xml` with every page's, tab's and post's address and `lastmod`; `BreadcrumbList`; a post's `BlogPosting` and `og:type` `article`, which its page writes none of yet; `noindex` and `301` where §5.4 names them; each page's card as `og:image`, which needs Typst (ADR-0005), so when the cards come, before hosting or with it, is the owner's word.
- The content, changed together with the owner (§8, step 9).
- Hosting, last (§8, step 10): SEC-6 of `docs/REQUIREMENTS.md`; the files of the snapshot before kept served beside the current ones, so a page fetched before a reload finds its stylesheet and chunks after it (ADR-0003), and with it whether a release switched in by the link and SIGHUP is watched too, since the watch would load the same release a second time and push out the snapshot before; `inbox` with delivery to Telegram, and the form the first screen's primary button opens, which does nothing until then; `face/print` on Typst, the CV listed in `llms.txt` with it, as the first version listed it, among the foot line's ways, its look the owner's to pick then, and before the profiles that end the first screen's brief (DESIGN.md, FootLine and Pages), which leave it out until it is served; the Dockerfile with both ports; deploy, with SEC-8, SEC-10's backup and SEC-11; the crawlers' `Accept` read off the server's log, for the owner to settle D21 against SEO (`docs/FOUNDATION.md` §5.4, SEO and Open Graph).
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

- A fault of a YAML file, a post's front matter among them, names its field by a path written by hand beside the yaml tags (`experience[0].end`), where a fault of a post's body names its line. yaml.v3's `Node` knows the line and column of each value, which would name the place (`profile.yaml:171:`) with no second copy of the schema; every file is parsed into a `Node` already, for the check of its characters, and a fault has its line.
- Content's faults are told apart by their words in tests (`TestEveryFaultSaysItsFieldAndProblem`), which `.claude/rules/tests.md` refuses: a kind for each fault would let a test read which fault came.

## Before the design work

- Measure again what `docs/FOUNDATION.md` §3.4 says of the page (heights, the overflowing link, the date column at 390px), with the shots kept beside the numbers.

## The owner's

- The blog's line, "What I build, and what I learn building it." (`blog_line` in `content/i18n/en.yaml`), is not the owner's own: his words, later.
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
