# TODO

The order of work is `docs/FOUNDATION.md` §8; each item is one chunk, started on the owner's word. An item done is removed, not ticked off.

## The foundation

- The faces, each its own commit (§8, step 5): the rest of `face/tui`: `:hire` once the form and the CV are there, since without them it repeats the ways the résumé opens with. Each reads the first version under _old/tree while the face is written, or in the archive where _old/tree lacks it (the TUI), and deletes what it replaces; the archive `paranin-first-version-2026-11-01.tar.gz` beside the repository keeps it whole.
- The page (§8, step 6), each part landing with its first user (D18), in this order:
  - The form's salary range, an island, waits for `inbox` (step 10).
- At 53.75 to about 60rem wide the foot line wraps to two rows, more than the command line's menu keeps back (`row_3`), so in a window under about 175px tall the menu's top row is cut. Keeping back the line's real height would need the script to measure it; rare enough to leave until a reader meets it.
- A tab's name keeps every letter of its title, so a site in another script serves its tabs at addresses in that script.
- A TUI frame narrower than what its top edge sets in, a widget's name and `● live` or a code block's language, runs past its edge: the frame drawing the edge's words in the room it has would hold it.
- A fault in a widget's props stands at the line YAML counts, which a body holding U+2028 or U+2029, a break YAML reads and markdown does not, sets a line or more past the line markdown counts: no check refuses either character.
- At 256 colours Latte's `code_builtin` and `code_variable` fall to one index, 161, and Macchiato's `code_number` and `code_operator` to 116, so the TUI's code shows those two kinds alike there: `scripts/tokens` picking distinct indices for the code colours of a theme would part them.
- A post's markdown and manual write neither its series, nor the posts beside it, nor where it is discussed, nor its other versions, which its document holds and its page and the TUI draw; an older version's write the newer alone.
- The document keeps no list's tightness, so a post's loose list of one paragraph an item comes out tight in its markdown.
- The manual writes a link outside the site by its words alone, so a post read with `curl` loses where its links lead: `docs/FOUNDATION.md` §5.4 plans numbered links in a terminal. The TUI's post leads by OSC 8 alone, as the résumé does, which a terminal that takes none drops: the numbered links would serve both.
- The blog (§8, step 8), as `docs/FOUNDATION.md` §5.4, D23 to D27 and `DESIGN.md` have it; its look was approved on 2026-11-22. Left to build for a widget, with the owner's first post, not before (owner, 2026-11-24): the window's room kept before the widget comes, which the widget's size gives (the page moves today as a widget comes alive); the page's language handed to a widget, its words in a typed table per language. `kit`'s controls, play, pause, step and a slider, wait for a look the owner picks.
- The commit hook `content-check` runs `paranin check content` with no build, so neither a widget whose code the build lacks nor props its schema refuses fails a commit: only CI and the server refuse them.
- A widget's schema's `pattern`, which valibot's `regex` writes in JavaScript's dialect, is compiled by Go's `regexp`, so a lookahead or a backreference refuses the build as its load compiles it: the validator takes another engine (`UseRegexpEngine`), regexp2 among them, once a widget needs one.
- No gate bounds the size of the page's build, which `docs/REQUIREMENTS.md` §5.5 asks of an island: a gate of our own is the owner's word.
- A draft's own mark, drawn on a canvas for the owner to pick; until then its version reads `draft` (`DESIGN.md`, PostHead).
- A filtered list is a page of a kind of its own (`PageTagFilter`), where a post's older version and its language are fields of its page (`Page.Revision`, `Page.Translation`): a check of a page's kind by `==` escapes `exhaustive`, so weigh making a filter a field of the list's page too.
- The tag filter's way out, drawn on a canvas for the owner to pick; until then it is the Link's underline under the pointer (`DESIGN.md`, Tag).
- A diagram, as `DESIGN.md`'s Article draws it, an SVG in `face` with boxes on `sunk`, cannot be a post's picture: a page shows a picture by `<img>`, in which a drawing loads no font and reads no token, so it keeps neither the page's face nor a theme's colours. Drawn into the page it would, which the load's refusal of raw HTML and the page's policy refuse today; how a diagram is written and drawn is the owner's word. A drawing in dark ink with no ground of its own vanishes on the dark theme's ground: whether a picture gets a ground of its own, or the writer gives it one, is his word too.
- A page, a feed and a post's picture are sent with `no-cache` and no validator, so a cache that asks again gets the whole answer each time: an `ETag` worked out once a snapshot, answered with `304`, would send only what changed. A feed reader polls a feed, up to twenty posts whole, again and again, so this lands before hosting at the latest.
- The mouse in the TUI (owner, 2026-11-25): a click on a tab, a window of the foot line, and a link the keys stand on (the list's groupings, tags and tiles, `laidTab.links`), and the wheel scrolling the page; nothing in the documents decides it yet, so the chunk writes it into `DESIGN.md`'s The TUI and `.claude/rules/tui.md`.
- A letter typed in a search costs the TUI a few milliseconds on the largest terminal, 2.8 ms on a long post, as it paints the tab again, so a client flooding at the input's pace spends up to about half a core a session: SEC-7 of `docs/REQUIREMENTS.md` bounds what one connection costs, and SEC-6 how many connect. A cheaper search key, or a pace counted by what an input costs rather than by its bytes, is weighed with hosting's SEC-6.
- The content, changed together with the owner (§8, step 9).
- Hosting, last (§8, step 10): SEC-6 of `docs/REQUIREMENTS.md`; the files of the snapshot before kept served beside the current ones, so a page fetched before a reload finds its stylesheet and chunks after it (ADR-0003), and with it whether a release switched in by the link and SIGHUP is watched too, since the watch would load the same release a second time and push out the snapshot before; `inbox` with delivery to Telegram, and the form the first screen's primary button opens, which does nothing until then; `face/print` on Typst, the CV listed in `llms.txt` with it, as the first version listed it, among the foot line's ways, its look the owner's to pick then, and before the profiles that end the first screen's brief (DESIGN.md, FootLine and Pages), which leave it out until it is served; the Dockerfile with both ports; deploy, with SEC-8, SEC-10's backup and SEC-11; the SEO of hosting (`docs/SEO.md`, At hosting).
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

- On the Russian list the blog's name stands as plain text in its `h1`, since the titles' block letters draw no Cyrillic (`list.templ`'s `titleHeading`, `internal/face/html/letters.go`): Cyrillic capitals in `internal/look/glyphs.go`, or a text title in the soft gradient, picked on a canvas.
- Nothing leads between `/blog` and `/ru/blog` but a post's Switch: whether the list's head offers the other languages' lists, and how, is his eye.
- On a phone the Switch's links stand 16px tall, under the 24px WCAG 2.5.8 asks for a target: the Switch's own shape, the list's groupings' too.
- The note of a translation older than its original (`DESIGN.md`, PostHead) stands in an older version's form for the way to the newer, `↑ старше оригинала`: its own form, drawn on a canvas for the owner to pick, as the draft's mark and the tag filter's way out wait to be.
- For his real `ru.yaml`: CLDR's `one` takes 21, 31 and on as well as 1, so a count's `one` form reads for both, `все {} часть` as `все 21 часть`; and a day writes its month short, `15 нояб. 2026`, which his `months` may write whole, `ноября`.

- A published post whose `updated:` names a day still to come gives its feed an `<updated>` in the future, since the load reads no clock and nothing holds the day back: his eye, whether the check refuses such a day for a post out, or the feed waits for it.
- `tint_strong`'s values (`design/tokens.yaml`, the way from an older version of a post to the newer, under the pointer) are the author's: in Latte the step from `tint` reads 1.13:1, and in Macchiato `text` over it 4.52:1, which no gate holds; a gate on text over a blend is his word.
- A post's head gives the day of its version read, its `updated:` where it says one, while the list's tiles and the panes give the day it was published, and that first day stands only inside the versions' Pick: his eye.
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
