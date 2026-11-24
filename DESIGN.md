# The design

paran.in is Pavel Paranin's résumé: one profile served as an SSH app, a web page, markdown, a man page and a PDF
(PRODUCT.md). The web page follows the rules of the web and looks like the terminal app: one mono face on a grid of
cells, block letters, the Catppuccin palettes, one accent. Its readers are hirers; the page has done its job when they
write to him.

The owner approved this design on 2026-11-17, and on 2026-11-22 the blog's, on the page and in the TUI, with the two
faces made one style. Everything here is the target. What holds today is
`design/tokens.yaml`, `web/stylelint.config.js` and the page itself; the page grew into this design part by part,
each part with the tokens and the rules it draws by; `TODO.md` lists the parts still to come.

- A token named here lives in `design/tokens.yaml`, which writes it into `web/src/tokens/tokens.css`; once it is
  there, the file's value is the one that counts, and this document names it without its value.
- A stylesheet names a token, never a value (stylelint, `.claude/rules/page.md`).
- No size is in pixels or another absolute unit, a media query's width neither (stylelint), so the whole page grows
  with the text size its reader set (owner, 2026-11-18): a fixed length is in rem, a space inside a line in cells,
  `ch`. The only pixels are a picture's own, as the favicon's.

## Content

- Write as the profile writes: plain English, short sentences, numbers as figures.
- The site is its author's: where it speaks, it speaks in the first person, "Write to me", as About does; a fact
  states itself.
- Sentence case everywhere. Section names, tabs, fact labels and chips are lower case, as the TUI writes them:
  `experience`, `looking for`. The block-letter titles are capitals because the font has no others.
- A button's words start with a verb and say what happens: "Write to me", "Copy", "Send". A toast says what was
  done, in the past: "Copied ssh paran.in".
- An error says what to fix: "Write the whole address, as name@company.com." Never "Invalid input".
- No small line over a heading, no numbered markers, no emoji, no "→" after a link. An article's paragraphs are
  numbered, as places a reader links to, not as markers.
- Every fact comes from `content/profile.yaml`. A fact or a point that is only a gap still to fill (`[later]`) is left
  out of the web until the author fills it; a gap inside a sentence shows `faint`.
- Dates are months, `MM/YYYY` (`08/2020`), with the machine's `YYYY-MM` in `<time datetime>`. A range joins two with an
  en dash and spaces, `08/2020 – 01/2022`; a role still held ends in `present`. A post and its versions are dated by
  the day, `21 Nov 2026`, with `YYYY-MM-DD` in `<time datetime>`.
- A post's tag is its word after a `#`, `#payments`, as readers know a tag that gathers posts. What a project is made
  of is no tag: its words as the profile writes them, the parts of a line.
- A heading joins a title and its subtitle with an em dash, `Seznam.cz — Backend Developer`; a line joins its parts
  with a middle dot, `Python · Browser-based code review`; a label ends in a colon where text follows it.
- Lengths of time as the profile writes them: `7y`, `1y 7m`, `11m`, or the words `side` and `learning`.

## The grid

- Space is counted in rows of 1.5rem (`row`) and cells of 1ch (`cell`): vertical space in rows, horizontal space inside
  a line in cells. The text read sits on the rows, but for an article's words, whose lines are taller, as a
  proportional face reads best (`reading`); every other style's line is its own (Type).
- The page fills the screen up to `content` with `gutter` either side, `gutter_phone` on a phone. Only running prose is
  held to `measure`, an article's to `column`; the layout itself uses the width.

| token | value | for |
| --- | --- | --- |
| `nudge` | `design/tokens.yaml` | an offset inside a row: an underline's distance |
| `row_half` | `design/tokens.yaml` | gaps inside a component |
| `row` | `design/tokens.yaml` | between a heading and what it heads |
| `row_2` | `design/tokens.yaml` | between two entries, the brief's top |
| `row_3` | `design/tokens.yaml` | between the name and the brief's columns |
| `row_4` | `design/tokens.yaml` | between two top-level sections, above the first screen |
| `cell`, `cell_2`, `cell_3` | `design/tokens.yaml` | a gap inside a chip; a control's side padding; a list's indent |
| `content` | `design/tokens.yaml` | the widest the content runs |
| `gutter`, `gutter_phone` | `design/tokens.yaml` | either side of the content, and on a phone |
| `measure` | `design/tokens.yaml` | the longest line of prose |
| `brief`, `brief_mid` | `design/tokens.yaml` | the brief beside the name, from `wide` and from `phone` |
| `lead` | `design/tokens.yaml` | the longest line of the headline |
| `post_lead` | `design/tokens.yaml` | the longest line of a post's lead |
| `when` | `design/tokens.yaml` | an entry's dates column |
| `span_thin` | `design/tokens.yaml` | a role's bar in the phone's career list |
| `dialog` | `design/tokens.yaml` | the widest a window over the page runs, a Dialog's or a Hint's |
| `pick` | `design/tokens.yaml` | the least width of a Pick's window |
| `textarea` | 7.5rem | a message field's least height |
| `rail`, `column`, `aside` | `design/tokens.yaml` | a post's Outline, its text and its sources beside it |
| `title_text_measure` | `design/tokens.yaml` | the longest line of a title set in text |
| `code_in_words` | `design/tokens.yaml` | code among an article's words, a share of their size |
| `tab_stop` | `design/tokens.yaml` | how far apart a tab's stops stand in code |
| `group_name`, `group_name_mid`, `tile` | `design/tokens.yaml` | the blog's list: the groups' names, and below `wide`; the least width of a tile |

## Sizes

Every control is one of the sizes below; the size fixes its height, its padding, the gap inside it and its words. No
other size: a control that seems to need one is the wrong control.

| size | height | either side | gap inside | words | controls |
| --- | --- | --- | --- | --- | --- |
| regular | `control`, two rows | `cell_2` | `cell` | `strong` for an action, `body` for a value | Button, Choice, Command, Field, Toast |
| compact | `control_compact`, one row | `cell` | `nudge` | `caption` | ThemeSwitch, Tooltip |
| line | a row less two `nudge`, inside the foot line's row | `cell` | `cell` | `caption`; `small` in a post's head and the blog's | the FootLine's chips, the Switch, the series' Pick |
| tick | three cells, `[x]` or `(*)`, on a row | label `cell` away | | `body` | Checkbox, Radio |

- Corners `radius` on buttons, chips, fields and dialogs; bars, rules and the current tab are square.
- Controls side by side stand `cell_2` apart, rows of them `row_half` apart.
- A dialog pads `row` and a half above, `row` below, `cell_3` either side.

## Colour

- One palette, Catppuccin Latte for light and Macchiato for dark; every colour is a token named for what it does
  (`design/tokens.yaml`), and nothing else is written.
- `background` is the page; `sunk` lies a shade below it (the foot line, a command chip, code); `deep` below that (a
  pressed quiet control). In the TUI the menu over the command line stands on `quiet`, apart from the page and the
  panel.
- `text` for what is read, `muted` for what stands beside it, `faint` for what a reader may pass over. In Latte `muted`
  is subtext1 and `faint` subtext0, one tone darker than the TUI first had them, so each reads at the least contrast
  `design/tokens.yaml` gives it over the grounds it names there, which `scripts/tokens` holds; `faint`, the lower, is
  never the only ink of something a reader needs. Neither stands on `tint`.
- A fact's label is `muted`. `label` blue rings the focus: blue, at its least contrast over the grounds the tokens
  name, in both themes; in an article it numbers the sources too, in the text and beside it.
- `accent` is spent where the eye should go first: the primary button, the cursor, the current role, the near end of
  the name, and a mark beside words that says where to look (the arrow to a newer version, the Outline's heading
  read). In Latte it is a fill, a display colour and such a mark, never small text (2.64:1).
- `host` is the site's host and the far end of the name. `tint` marks a result worth seeing first, and the way from an
  older version of a post to the newer. `error`, `passed` and `warned` only for states, always with words. In Latte
  `host` and `error` are Catppuccin's mauve and red a shade darker, so each reads as small text at the 4.5:1 WCAG asks,
  over the grounds `design/tokens.yaml` names.
- Code colours each kind of word in the hue Catppuccin's own style guide gives it, a token for each (`code_keyword`
  to `code_variable`), each reading at 4.5:1 on `sunk`, the code block's ground. In Latte each hue is a shade darker,
  as `host` and `error` are. Numbers stand in teal where the guide has peach, since `accent` is never small text. A
  comment is `muted`, words read at the contrast small text asks; punctuation `faint`.
- Borders: `quiet` for rules between parts, `muted` for a control's border and for a pane's or a tile's frame, `edge`
  for the corner crosses and for what lies over the page, a tooltip, a toast and a pick's window. A pane or a tile
  that leads nowhere is framed in `quiet` at half its strength (`quiet_half`).
- `scrim` fades the page behind a dialog towards its own ground: Latte's and Macchiato's background at 85%.

## Type

- One face (`face`), monospace, ligatures off, set once on the body for every element to inherit, served from the
  site itself (FOUNDATION §5.6), so every reader sees the face the cells are counted in. A second face, iA Writer Duo
  (`reading_face`), sets an article's words alone, its paragraphs, lists and quotes; everything else on a post's page,
  its title, headings, code and notes, stays in `face`.
- A text style is its size, its line and its weight, set as three properties (`web/stylelint.config.js`).

| style | size / line, rem | weight | for |
| --- | --- | --- | --- |
| `mail`, `mail_phone` | `design/tokens.yaml` | `design/tokens.yaml` | the mail heading the brief, and on a phone |
| `lead`, `lead_phone` | `design/tokens.yaml` | `design/tokens.yaml` | the headline under the name, and on a phone |
| `heading` | `design/tokens.yaml` | `design/tokens.yaml` | an `h2` on a page that is not the résumé |
| `body` | `design/tokens.yaml` | `design/tokens.yaml` | everything read; the default |
| `strong` | `design/tokens.yaml` | `design/tokens.yaml` | a role, a company, a project's name, a button's words, a word made strong |
| `small` | `design/tokens.yaml` | `design/tokens.yaml` | the foot line, a field's hint |
| `caption` | `design/tokens.yaml` | `design/tokens.yaml` | the career line's names and years, a chip's word, a tooltip, a source's number where a line cites it; never prose |
| `title_text`, `title_text_phone` | `design/tokens.yaml` | `design/tokens.yaml` | a title set in text: a post's, and a group's on the blog's list; on a phone |
| `title_text_mid` | `design/tokens.yaml` | `design/tokens.yaml` | a group's title on the blog's list below `wide` |
| `reading` | `design/tokens.yaml` | `design/tokens.yaml` | an article's words, in `reading_face`, on a line of its own taller than a row |
| `source` | `design/tokens.yaml` | `design/tokens.yaml` | a source beside an article's text, `caption`'s size on a line short enough to read as one note |

- The page's one `h1`, a tab's title among them, and the `h2` of the résumé's first section under its first screen are
  block letters; a post's title is its page's `h1` in `title_text`, since a sentence in block letters would not fit;
  another page's `h2` is `heading`; every heading below is `strong`, its level shown by the space before it (`row_2`
  before `h3`, `row` before `h4` to `h6`). Levels never skip.
- Code is told by its `sunk` ground, since the whole page is one face.
- Weights: 400 for text, 700 for what names (a role, a company, a button) and for a word made strong, wherever it
  stands, 800 only for the mail, its @ left at 400.

## Block letters

- The name and the five sections' titles are drawn in the TUI's block fonts, the same glyphs
  (`internal/look/letters.go`), as SVG rectangles: they need no font and shrink to their column, never past it.
- Natural size: `name_cell` a cell for the name, `title_cell` for a title.
- One gradient, `accent` to `host`, left to right. `shape-rendering: crispEdges`, so no seam shows between two
  rectangles. A title set in text, a post's and a group's on the blog's list, takes the same gradient softened, each
  end mixed a fifth towards `text` (`title_soft_start`, `title_soft_end`), so a title of letters reads apart from one
  of blocks.
- The title font's X is drawn on three cells (▀▄▀ over ▄▀▄), in the TUI too: on five it read as )(.
- The cursor after the name is an underscore a letter wide on the letters' foot, in `accent`, and it blinks, as a
  terminal's does: the one thing on the page that moves on its own, and it stops for a reader who asks for less motion.
- The element carries the words as `aria-label`; the drawing is hidden from screen readers.

## Shape, depth and motion

- Lines are `hairline`. No shadows: depth is a shade of ground (`sunk`, `deep`) and a hairline.
- What stands over the page as it scrolls lies on a layer of `design/tokens.yaml`, the layers listed lowest first:
  a live tile over a tile to come beside it on `tile`, so the line the two share is the live one's, what is set into
  a tile to come's edge on `tile_name`, the tile the keyboard is on on `tile_focus`, so nothing beside it cuts its
  ring, and the foot line on `foot`, over them all.
- Nothing moves on its own but the cursor. Colours change under the pointer in `quick`; with reduced motion, at
  once. A toast stays `toast`.

## States

| state | how it shows |
| --- | --- |
| rest | as its component says |
| hover | a quiet control one shade down (`sunk`); the primary turns `host`; a link's underline turns `accent` and `underline_hover` thick; a field's border `text` |
| pressed | one shade further (`deep`); the primary turns `text` |
| focus | a `focus_width` ring in `label`, as far from what it rings, on every control, by keyboard |
| current, chosen | reversed, `text` ground and `background` words, bold, as the TUI's tab: a mark of form, not of colour alone |
| disabled | `sunk` ground or `quiet` border, `faint` words, `aria-disabled`, and the page says why; a pane or a tile its frame in `quiet_half` and no ground, a shade down being the live ones' hover |
| error | an `error` border and words under the control saying what to fix |
| busy | the words say what is happening ("Sending…") and the control waits |

- A visited link looks as any link.

## Layout

- From `wide` (80rem) a section's title stands beside its content where it fits at its full size, and over it where
  it would shrink (Frame and Section). Below `wide` the brief narrows to `brief_mid` and every section's title stands
  above its content.
- Below `phone` (53.75rem) everything is one column, the gutters `gutter_phone`, the foot line wraps, the career line
  becomes a list, a skill table folds (Table) and a table that names its columns scrolls inside itself. No page is
  ever wider than its screen.
- The two widths are the only ones a media query may hold, in rem, since a custom property cannot stand in one;
  stylelint holds them from the first media query on, and refuses any width until then. A media query's rem is the
  reader's text size too, so a reader who sets larger text gets the narrower layout sooner.
- A page shows one tab of its document, as the TUI does: the opening at the page's own address, each tab after it at
  its own. Every page has one `main`, then one `nav`, the foot line; with nothing before the main part, it needs no
  skip link.
- A position worked out from data (the career line) is an SVG attribute: the site's policy refuses inline styles.
- A post's page is one grid from its head to its foot: the Outline, `rail` wide, then the paragraphs' numbers, the
  text at `column`, and the sources, `aside` wide. Below `wide` the Outline gives way; below `phone` one column, the
  numbers in a narrow margin and the sources at the text's end.
- The blog's list is a grid of two columns: the groups' names, `group_name` wide, `group_name_mid` below `wide`, and
  the tiles, as many in a row as hold `tile` each. Below `phone` the name stands over its tiles and the tiles take one
  column.

## Iconography

- No icon set. The page uses the terminal's glyphs: `[x]` and `(*)` for a checkbox and a radio, ❯ before a command
  (`prompt`), the cursor after the name, ✓ in a toast (`passed`), and `-` before a point (`faint`). The blog adds `#`
  before a tag (`faint`), `↑` before the way to a newer version (`accent`), `▾` after a pick that opens a window
  (`faint`), `▌` before the Outline's heading read (`accent`), `●` before `live` (`passed`) and `*   *   *` for a
  pause (`faint`).
- Two drawn marks, SVGs in `currentColor`, `box` square, one stroke of 1.5 of its 16 units: the theme button's
  half-filled circle and the dialog's close cross. The first screen's corner crosses are drawn the same way, a `cross`
  square with two hairlines across it, the TUI's + on the page. A mark that needs drawing is drawn as these are, never
  taken from a set.

## Components

### Name

The page's `h1`, only on the first screen: PAVEL over PARANIN in the name font, a blank row between, the gradient across
both lines, the cursor after the last letter. Don't recolour a letter, set the name in plain text beside it, or put
anything over it.

### SectionTitle

A tab's title in the title font, the page's `h1`: Experience, Projects, Skills, Contact, and each tab of a project's
page; About is the `h2` under the first screen. A subsection (Spoken languages, Education, a skill group) is a
`strong` heading.

### Mail

The address as the site writes it: the box in `accent`, @ in `faint` at the body's weight, the host in `host`. It heads
the brief in `mail`; elsewhere the same three colours at `body` size. A `mailto:` link; under the pointer the box and
the host underline, `underline_hover` thick, `nudge` below.

### Favicon

A square of the name's gradient and nothing more: `accent` at its left edge to `host` at its right, filling it edge to
edge, in the light theme's colours for every reader, since they stand apart from a light tab and a dark one alike. One
picture at every size: as SVG, a PNG of 180px for a phone's home screen, whose corners the phone rounds, and one of
32px at /favicon.ico. The picture a shared link shows is Typst's, drawn with the CV (ADR-0005).

### Button

Does one thing its words say, verb first.

- Primary: `accent` fill, `on_accent` words; under the pointer `host`, pressed `text`, each with `background` words.
  At most one a screen: writing to Pavel. Until the form is there (with `inbox`) it does nothing.
- Quiet: no fill, a `muted` hairline, `text` words. Everything else.
- Regular size, `strong` words. A link that leaves the page is a link, not a button.

### Command

A command the reader can run: ❯ in `prompt`, the command in `text`, then "copy" in `text` and `small` after a `muted`
rule. `sunk` ground and a `muted` hairline, so it shows on any ground; under the pointer the hairline turns `text` and
the ground `deep`. The whole chip is the button; the command stays selectable text. On copy the word turns to "copied"
in `passed` and a Toast says what was copied. A command never wraps: on a narrow screen it scrolls inside itself. The
button inside it is what a keyboard reaches, and its focus rings the whole chip; without the script the button is
hidden and the command is words to select, and a copy the browser refuses selects the command.

### Link

The `text` colour, underlined in `muted`; under the pointer the underline turns `accent` and `underline_hover` thick.
Never coloured: the underline says it is a link. No arrow after a link's words. A link that is a whole box, a tile or
a pane, is told by its frame, and underlines its title under the pointer.

### FootLine

One line of tmux's at the foot of every page, after its main part: it sticks to the screen's foot as the page scrolls
under it, and on a page shorter than the screen stands at the screen's foot. `sunk` ground, a `quiet` hairline over
it, one `row` tall and twice `focus_width` under it, so a ring is never cut, `small`, on the content's grid. A part
wraps whole to a row more where the screen is too narrow for the line; a part of the main part the keyboard reaches
scrolls `row_2` clear of it, `row_4` on a phone, and the line's own parts, always in sight, scroll nothing.

- The sessions, as tmux's: `[paran.in]`, leading to where the site begins, and `[blog]`, leading to the blog's list,
  two cells apart; the one the page belongs to in `host`, the other `muted`.
- The document's tabs as windows, as the TUI numbers its tabs: `1:about`, the number `faint`, the name lower case, a
  cell of air either side. The window shown is reversed and bold, as the TUI's tab, and marked `*`, which a screen
  reader skips: the link is the current page. The blog's windows are `1:posts`, its list, and the post read by its
  slug, `2:own-way-to-pay`.
- Under the pointer a session and a window turn `deep`; nothing in the line has a pressed shade, `deep` being the
  lowest ground.
- At the line's end the ways the site is served, as the line's chips (Sizes): `❯ ssh` and `❯ curl`, ❯ in `prompt`,
  each opening its Hint, reversed and bold while it is open, as the window shown; then the markdown of the page whole
  by its file's name, a link, the résumé's on a page that has none. `deep` ground, a `quiet` hairline turning `text`
  under the pointer, `caption` words.
- Then the VimKeys switch and the ThemeSwitch, its Tooltip over it. The Toast stands a row over the line, and the
  CommandLine, open, on the row over it.
- On a phone the sessions and the theme button share the first row, the windows take the rows `row_half` below by their
  names alone, a `nudge` either side, wrapping, and the ways and the VimKeys switch give way to them.

### Hint

A window of tmux's over the foot line, which a way's chip opens: the way's command whole to copy and what it gives.
`ssh` gives the résumé as an app, `curl` the page shown, whole, as a manual. It stands a `nudge` over the foot line's
hairline, whatever its panel holds, the open command line too, its right edge at the ways' chips', on whatever row
the line wraps them to, so the ThemeSwitch's Tooltip, past the VimKeys switch, shows beside it, never under it; where
the screen leaves it no room there, it starts at the content's left edge. On `sunk` with an `edge` hairline, as all
that lies over the page; it pads as a Dialog does. Its title, the program, bold, and `esc close` in `muted`, turning
`text` under the pointer, are set into its top edge, half above it. Inside, the Command, then what it gives, `small`
in `muted`.

- The browser's own popover: it opens with no script; Esc, a press outside and the other way's chip close it, Esc
  handing the focus back to the chip, and closing it alone, after a Tooltip shown over the line (CommandLine). No
  scrim: nothing is interrupted.
- The keyboard goes from the chip to the copy, then to close.
- A copy closes it, and the Toast says what was copied, as for every copy. A copy the browser refuses leaves it open,
  the command selected.
- On a phone the ways give way to the tabs, and with them the hints.

### VimKeys

`[ ] vim keys` at the foot line's end, before the ThemeSwitch: a checkbox as a terminal writes one, the brackets
`muted`, the tick `x` bold in `text`, then its words; the whole label is the target. Under the pointer it turns `deep`
as a window does; the keyboard's focus rings the label. Ticked, the keys of the CommandLine answer, on every page after
too; unticked, a page has no keys of its own. The script shows it; without it, and on a phone, which has no keys, it
is gone, and with it the keys, so none answers where the reader cannot untick them.

### CommandLine

Neovim's command line with no row of its own, as the TUI's: hidden until `:` is typed with the vim keys on, then a row
of the foot line's panel over the line's own, the two one panel, `small`. After the colon what is typed in `text`, the
cursor an underscore in `accent`, as the cursor after the name, which is the field's focus, so no ring; at the right
the keys it names, `tab next` (while a place fits), `↵ go`, `esc cancel`, each key a cap of `caption` words on `deep`
in a `quiet` hairline, `radius`, what it does `muted`.

- The menu of the places that fit what is typed stands over it as soon as the colon is typed and narrows with each
  letter: before a letter the page's tabs but the one shown, then every tab, place of the page's own (The TUI) and
  project's page the letters start, after `:e ` the pages it opens: the projects', `project`, the blog's lists,
  `list`, and on the blog's own pages each post the list names, `post`, which no other page carries for its weight, by
  the names a terminal reaches them by, as the TUI's `:e`. `:e` opens any name in the blog's folder, a post by its
  number or one no list names among them, by going to its address, where the site answers for it. A hairline
  box in `edge` on `sunk`, as a Tooltip lies over the page, `radius`; each name under where the line writes it, a cell
  of air either side, its kind `muted` in a column of the kinds four cells after the widest name, as The TUI names it.
  The letters typed are bold, underlined in `accent`, which Latte allows as a fill and never as small text, on the
  TUI's menu too.
- Tab and Shift+Tab go round them, writing the one chosen on the line, reversed and bold as the window shown. The menu
  holds as many whole rows as the screen has over the foot line's two, a row of air above, stands a `nudge` clear of
  its hairline, and scrolls in the browser's own thin bar, keeping the one chosen in sight; the page does not move.
- Enter goes where the line says; a command it refuses it names on the line in `error`, as Neovim words it (`E492: Not
  an editor command: x`), until the next key. Esc, Backspace on an empty line and the focus leaving close it.
- `/` opens the same line for a search, as the TUI's: what is typed marks the page's words at once, whatever their
  case, the one the count names on `accent` in `on_accent` words, the others underlined in `accent`,
  `underline_hover` thick, a `nudge` under the letters, as the menu's typed letters are, a phrase a mark or a link
  parts among them but none across two blocks, and brings the page to the first match at or after where it stood,
  `row_2` clear of the panel; the count, `[1/4]`, bold in `text`, stands before `↵ find` and `esc cancel`. Enter
  keeps the search, and the line says `/Python [1/4]` in `text` until the next key; n and N go round the matches,
  from where the page stands once it is scrolled away, saying each; a search that finds nothing says `E486: Pattern
  not found: x` in `error`. Esc while typing puts back the search kept before and the page where it stood; Esc on the
  page puts the search away once nothing stands over it. A search kept marks every page the reader opens after it in
  the browser's tab, as the TUI marks every tab, each match underlined before n or N names one, until Esc on a page
  puts it away; with the vim keys off a page marks none.
- Esc puts away one thing a press, from the top: the line being typed in, then a Tooltip shown, then a Hint or a
  Pick open, then the search kept.

### ThemeSwitch

One quiet button at the end of the foot line: the half-filled circle, no words, a `control_compact` square, no border
and no ground, `muted` turning `text` under the pointer. The page starts in the system's theme; a press turns it to
the other, sets `data-theme` on the root and keeps it for the next visit; a press that lands back on the system's
theme forgets the choice. It is a toggle button named "Dark theme", `aria-pressed` while the page is dark, its name
in a Tooltip over it under the pointer. Without the script it is hidden and the page follows the system.

### Facts

Label and value, two cells apart: the label `muted`, one or two lower-case words as the profile writes them; the value
in `text`, on the label's baseline. On the first screen the first four the profile lists, its author's choice
(experience, stack, now, languages); the rest stand in About under the summary (owner, 2026-11-18). On a phone a
value that is a Command stands under its label across the whole width, as a Card's install line does at any width.

### CareerLine

Every role on one line of time, each as long as it lasted: ended roles in `edge`, the current one in `accent`, its name
bold. Names alternate between two rows, so a short role keeps its name; under the line stand only its ends, the first
year and the site's word for the present (`present`). It runs to the day the page is drawn. Below `phone` it becomes a
list: years in a column as wide as the widest, company, and a bar as long as the role beside the longest. A role with a
month still to fill is left out of both.

### Entry

One role or school: its dates in a column of their own (`when`), then its heading, `strong` at any level (`h3` for a
role, `h4` for a school), its place in `muted`, its points. The current role's dates bold in `text`, ended ones
`muted`. Points start with a `faint` dash, at most `measure` wide. Entries are set apart by a `quiet` hairline, never
boxed.

### Card

A piece of work: its name as a `strong` heading with a link to its page, its language and line in one `muted` row, its
summary, its facts. An install line is a fact whose value is Code alone; the page draws it as a Command, under its
label across the card's width, since beside the label a long one outruns the value's column. Cards are
`row_2` apart with a `quiet` hairline between, never boxed.

### Table

Rows of a few cells, as a skill group's skills: the first cell `strong`, the rest `text`, a duration `muted`; cells
`cell_3` apart, rows set apart by a `quiet` hairline; numbers line up (`tabular-nums`). No header row, no stripes. A
post's table may name its columns: the page draws that row as the table's first, in `strong`. A cell but the last keeps to one line, so on a narrow screen the table scrolls inside itself.
On a phone a table that names no columns, a skill group's, folds: each row's last cell, a skill's sources, stands under
the others across its width, and its first cell may wrap, so it fits rather than scroll.
Whatever scrolls inside itself, a table, a Command, a code block, shows a thin scrollbar.

### Code

In a line, Code on a `sunk` ground; on lines of its own, a block padded `row_half` and `cell_2`, `radius`, scrolling
sideways inside itself, never wrapping, a tab's stops `tab_stop` apart, as an editor sets them, not the browser's
eight cells; the manual writes a tab as the spaces to the same stops.

### Contents and PageHead

- Contents: the pages a page leads to, one a line, `row_half` apart, each a link by its title, then a colon and its
  line in `muted`. No bullets.
- PageHead: a page that is not the résumé opens with its title as the `h1` in the title font, its line in `lead`, what
  it is made of as `muted` words joined by ` · ` (not chips and not Tags: they do nothing when pressed and gather
  nothing), its lead, and where its code lives as Facts.

### Highlight

A result worth seeing first in `tint`; a gap still to fill in `faint`, with its brackets. A gap inside a result is in
the result's `text`, told by its brackets alone, since `faint` never stands on `tint`.

### Field, Choice, Checkbox and Radio

- Field: the label above it in `muted`, a real `label`; a hint in `muted`; the box with a `muted` hairline, regular
  size; a message field at least `textarea` tall and growing. An error turns the border `error` and says what to do.
  A placeholder only shows an example.
- Choice: one choice among up to five short ones, as a row of chips, each a radio inside its label, the chosen one
  reversed. A longer list is Radio.
- Checkbox: `[x]` before its label, as a terminal writes it, the brackets `muted`, the x `text` and bold. The input is
  hidden and stays the control; the whole label is the target; the focus ring goes round the brackets.
- Radio: `( )` before each choice, `(*)` before the one chosen, as `dialog(1)`'s radiolist writes it, drawn as
  Checkbox is.

### FormStates

Busy: the button's words become "Sending…" and it waits; nothing spins. Sent: ✓ in `passed`, then what happened. Not
sent: `error` words saying why and what to do, with the mail as the way that always works.

### Dialog

A window over the page, as tmux draws a popup: on `sunk` with an `edge` hairline, as the Hint and all that lies over
the page, the title set into its top edge on the left and the close cross into it on the right, both half above the frame. The page behind fades by `scrim`.
A native `dialog` opened with `showModal()`: focus moves in, Esc or the cross closes it, focus goes back. The cross is
named "Close"; its tooltip says Esc does the same. Never opens on its own.

### Tooltip and Toast

- Tooltip: a word or two about a control, on hover and on focus, two `nudge` over it; `sunk` ground, an `edge`
  hairline, `caption`; never the only place something is said. Esc hides it, in its turn (CommandLine), until
  its control is left.
- Toast: ✓ in `passed`, then what was done, in the past tense, a row over the FootLine for `toast`; regular size,
  `sunk` ground and an `edge` hairline, as the Tooltip's. Its words stand `row_half` from its top and foot, a line or
  several, and wrap at their spaces alone, so a long command breaks between its words, never at a hyphen; the ✓
  stands by their first line, as a mark beside words does. Errors are not toasts: they stand where they happened.

### Frame and Section

- Four crosses mark the first screen's corners, `cross` in `edge`, half a gutter outside the content and `row_2` in
  from its top and foot: only there and round a post's head, which stands as a first screen, and not on a phone, whose
  gutter leaves them no room.
- A section: its block title at the left, its content at the right, as wide as an Entry: `when`, `cell_3` and
  `measure`; a title too wide for the room the content leaves stands over the content in its column at its full size,
  never shrunk to fit beside it, so the content stands in one column on every tab; `row_2` of air above and below it,
  `row` on a phone. Its blocks stand a row apart, a subsection two. A tab of the résumé stands as a section, its title
  the page's `h1`.

### Tag

`#` in `faint`, then the word in `muted`, `small`; tags stand `cell_2` apart and wrap whole. On the blog a tag is a
link that filters the list, told by its hash, not an underline: under the pointer its ground goes `sunk` a `nudge`
past its word and the word turns `text`; the tag the list is filtered by is reversed and bold, its hash too, and leads
back to the list, which a screen reader hears it named as; under the pointer it takes the Link's underline, `accent`
and `underline_hover` thick, until the owner draws its way out (`TODO.md`). A post the lists name nowhere, a draft, a preview or one unlisted, keeps its tags as words. On a
tile the tags are words of the tile, `caption`, all `faint`. Only posts have tags.

### Switch

A few addresses of one thing, a post's language or how the list groups its posts: their words in one row of `quiet`
hairline, line size, `radius`, a hairline between each two, the one shown reversed and bold as the window shown; under
the pointer `deep`. They are links, since each leads to an address.

### Pick

A word that opens a window over the page, as the Hint's chip does: a post's versions, its series' parts.

- The version: its date bold in `muted`, underlined in `edge`, then `▾` in `faint`; under the pointer `text`,
  underlined in `accent`; reversed while its window is open, as the series' chip. Its window is titled `versions`, a
  row a version's date, newest first, its state `current` or, for an older one read, `reading`, in `muted`; a post of
  one version lists that one. The window opens under the whole fact, so an older version's way to the newer stays in
  sight.
- The series: the FootLine's chip, its name, then `1 of 11` in `muted` and `▾` in `faint`; reversed while its window
  is open, as a way's chip. In the Panes' heading, which names the series already, it says what it opens, `all 11
  parts`.
- The window stands as the Hint does, on `sunk` in an `edge` hairline, `radius`, `small`, its title bold and `esc
  close` set into its top edge. A row a line: a part's number `faint` and its title, or a version's date; its state
  at the right (`reading`, `current`, `next`, `soon`). The row read is reversed and bold; a part out is a
  link, `deep` under the pointer; the next part bold; a part to come `muted`, no link. On a phone the window stands
  across the screen over the foot line.

### PostHead

A post's first screen, as the résumé's. Left, over the Outline and the text, its title (`title_text` in `title_soft`,
a few words a line, balanced) and its lead, `post_lead` at most. Right, in the sources' column, its Facts `row_half`
apart: version (the Pick), series (the Pick), language (the Switch), tags (Tags). An older version says under its date
that a newer stands: `↑` in `accent`, then `newer: 21 Nov 2026` bold in `text`, on `tint`, line size, `radius`, the
whole a link, `tint_strong` under the pointer. On a phone one column, the facts under the lead, a `quiet`
hairline under the head. A draft, which its author alone reads under `paranin serve --drafts`, gives the word `draft`
as its version, a fact's value like any other, where a post out gives its day; the TUI and the text faces write it
there too. A preview, read at its link, gives its date as a post out does. It is the plainest mark the tokens give, standing until the owner draws one.

### Outline

A post's headings of its top level in `rail`, `small`, standing in sight as the article scrolls, from its head on:
each `muted`, a link with no underline, `text` under the pointer; the heading read bold in `text`, `▌` in `accent`
hanging before it. Gone below `wide`; the `:` menu holds the headings too. The heading read is the one whose section
holds the place last gone to, by a link or the menu, while that place stands between the screen's top and
`reading_line`, or anywhere on the screen at the page's end; else, at the page's end, the last; else the last whose
top has passed `reading_line`; none before the first has.

### Article

A post's body, in the text's column.

- Its words in `reading`; each paragraph numbered in the margin, `caption` in `faint`, the number a link to the
  paragraph, under the pointer `text`, underlined in `accent`. Headings: `h2` in `heading` a row below the block
  before it, `h3` in `strong` half a row below, since an article's blocks already stand a row apart. Code among the
  words stands at `code_in_words` of their size.
- A source: its number raised after the word, `caption` in `label`, no link; beside the text, at the line that first
  cites it, under a `quiet` hairline, its number in `label` and its words `source` in `muted`. On a phone none stands
  beside the text: the sources stand at its end, each once, under a `quiet` hairline, `small`.
- A quote: a bar of `edge`, `underline_hover` wide, at its left; who said it under it, `small` in `muted`.
- A pause, a thematic break: three stars as a book sets them, `*   *   *`, in `faint` in the page's face, in the
  middle of the text's column, a row of air above and below on top of the row its block before keeps; no paragraph's
  number.
- Points as the résumé's, a `faint` dash hanging in the margin; a list in order, its numbers `faint`.
- Code: in a line, Code; on lines of its own, Code's block in `small`, its language in `faint` and `copy` in `muted`
  at its top right, under the pointer `text` underlined in `accent`, copying as a Command does; the page copies by
  the button alone, so a reader selects the lines. Each kind of word in its colour (Colour): a keyword mauve, a
  string green, a number or a constant teal, a function's name and a YAML key blue, a type yellow, a builtin red, an
  operator sky, a shell's variable maroon, a comment `muted`, punctuation `faint`, every other word `text`. Code in a
  language the site does not colour, or in none, stands in `text`.
- A picture: a block of its own in the text's column, as wide as its file and never wider than the column, its shape
  kept and its room held before it comes; its caption under it, half a row below, `small` in `muted`. A diagram is an
  SVG in `face`: boxes on `sunk` in an `edge` hairline, the one that is ours in `accent`, lines in `edge`; on a phone it
  scrolls inside itself.
- A widget: a window as tmux draws a popup, a `muted` hairline on the page's ground, `radius`; its name bold, `small`,
  set into its top edge at the left, `● live` at the right, the dot in `passed`, the word `muted`; live inside. A pick
  among a few is a Choice; an answer stands on `sunk`, its verdict in a state's colour, always with its words. Where
  nothing runs, its alt text and a link to it.

### Panes

A post's foot, after a `quiet` hairline: a heading in `strong` saying where the panes lead, `Other parts of stint` with
the series' Pick, or `Other posts` with a link to the list; then two panes as tmux splits a window.

- Each pane square, framed in a `muted` hairline, the two sharing the line between them; the whole pane is the link,
  `sunk` under the pointer, its title's underline then `accent` and `underline_hover` thick; the keyboard's ring parts
  where the name is set in, as the frame does.
- Its name set into its top edge, bold, `small`: `part 4`, `older`. Its title bold, underlined in `muted`; the side
  after's title, the next part's or the newer post's, in `heading`. Its lead `small` in `muted`; beside other posts, a
  post's date `caption` in `faint` at the pane's foot, since the blog orders them by date.
- A part to come, a title its series names under `to_come`: its title `muted`, `soon` after it, no link. A side with
  none: framed in `quiet_half`, its name and its words `faint`, `none: this is the first part`.
- Under them where the post is discussed: `discuss` in `muted`, then the links.
- On a phone one pane under the other, the next first, the empty side last.

### BlogHead, Group and Tile

- BlogHead: BLOG in the title font, the gradient across it, in the names' column; beside it the blog's line in `lead`, and under it `by` in
  `muted`, the Switch of `years` and `series`, and the Tags, each a filter of the list. On a phone the tags take a row
  of their own. On a filtered list the Switch shows `years`, the grouping it filters; a list filtered by a series
  opens its tags with the filter in the site's words, `parts of Stint`, reversed and bold as the tag a list is
  filtered by, with no hash, leading back to the list: the plainest form the tokens give, standing until the owner
  draws one. A filtered list's page is called `blog — posts tagged go`.
- Group: its name in `title_text_mid` (`title_text` from `wide`), `title_soft`, staying in sight at the top while its
  own tiles scroll, but on a phone, where it stands over them; under it what it
  holds, `small` in `muted`: `6 posts`, or a series' line and `6 of 11 out`. By year, newest first, at `/blog`;
  by series at `/blog/series`, the posts in no series last, as `other posts`.
- Tile: a post as a pane, square, framed in a `muted` hairline, the tiles sharing their lines, no gap. Its date set
  into its top edge, `small` in `muted`, or in a series' group its part, bold in `text`, then the date. Its title bold,
  balanced; its lead `small` in `muted`. At its foot, apart from the post, where it stands: its series where the
  group does not say it, `part 6 · stint`, the part bold in `text`, the dot `faint`, the name `text`, a long name
  wrapping, the name a Link to the series' group on the list by series, standing over the tile's; then its Tags. The
  whole tile is the link, its title's, which holds its date set into the edge, stretched over it, `sunk` under the pointer, its title underlined in `accent`; the keyboard's
  ring parts where the date is set in, as the frame does. A
  part to come: framed in `quiet_half`, its date and title `faint`, not bold, no link.

## Pages

- The first screen: left, the greeting, the Name and the headline in `lead`; right, `brief` wide, the Mail, the role,
  the place, the Facts, the primary button and the Command, then cv.pdf, GitHub and LinkedIn. Below, the CareerLine
  across the whole width; the four crosses at the corners. Under it the About section; the rest of the résumé is its
  tabs, each a page of its own.
- The address of no page: its status in block letters, the address quoted as Code, a link to the start; no apology and
  no joke. The foot line stands as on every page, with no windows.
- The blog's list: the BlogHead, then its Groups of Tiles; a tag or a series filters it, the filter reversed in the
  head.
- A post: the PostHead, the Outline beside the Article, the Panes; an older version the same, its head saying a newer
  stands.

## The TUI

The TUI draws this design in a terminal, as Neovim inside tmux: the same tokens, each kind of text and part of the
frame painted in one (`internal/face/tui/paint.go`), in full colour, 256 or 16. What a cell cannot hold, a size, a
second face, a hover, it leaves out; what the page and the TUI both draw, they draw alike.

- Frames are tmux's, square, `┌┬┐├┼┤└┴┘`, each in one colour, round a command, the window of keys, panes, tiles and
  code. A command's label, a pane's and a tile's name stand set into the top edge at the left.
- A tab's title, a project's section's too, is in the title font in the gradient, or in text in the gradient where the
  letters do not fit. A post's title is bold text.
- A post's tags as the page's: `#` in `faint`, the word in `muted`. What a project is made of as the page's PageHead
  writes it, `muted` words joined by ` · `, not chips.
- Code in a line stands in `text` on `sunk`; a block of it in a `quiet` frame, its language in the edge, coloured as
  the page's. A line wider than the frame goes on under itself, never cut: each row it goes on in opens with `↪ ` in
  `faint`, as Neovim's showbreak marks it. A command is a box in a `muted` hairline with its label in the top edge,
  `❯` in `prompt` before the command, as the page's Command; y copies it once a search has found it, as it copies any
  value.
- The blog is a session of its own. The footer opens with `[paran.in] [blog]`, the one shown in `host`, then its
  windows: on the blog `1:posts` and the post read, `2:own-way-to-pay*`. `:e blog` and `ssh blog@paran.in` open the
  list; `:e blog/12` and `ssh blog/12@paran.in` the post, the list in the first window; h and l go between the two;
  q on a post goes back to the list, and on the list back to the résumé. The first window keeps the list's grouping
  the reader chose, by year or by series, beside a post; the page, which keeps no state, leads it to the list by year.
- The blog's pages run wider than the résumé's column where the screen has room, 132 cells at most, the text still
  held to its measure.
- The list: BLOG in the title font, as the page's; the groups' names in a column, 22 cells wide, 18 under 132, the
  year in the title font in the gradient,
  staying in sight at the top while its tiles scroll; under 100 cells the name stands over its tiles as one line in the
  gradient. The tiles as the page's, at least 31 cells each. j and k go from tile to tile in reading order, newest
  first; the tile chosen stands on `sunk`, its title underlined in `accent`; Enter opens its post. A move that takes
  the page from the tile chosen chooses one in sight: G the last, gg the first, a search's match its own.
- A post's version is its date among its facts, as the page's; the TUI has no Pick, so the other versions are places of
  the `:` menu, `version`, each by its date as its address writes it. An older version says under its date that a newer
  stands, as the page's: `↑` in `accent`, then `newer: 21 Nov 2026` bold on `tint`, painted as a result worth seeing
  first is, a link to it by OSC 8.
- A post's lists stand inside the text's column, so their marks never read as the paragraphs' numbers: a point's dash
  `faint`, an ordered list's numbers `muted`, the words after the widest mark.
- A post: the Outline (`rail`) on a page 132 cells wide (a screen of 140), the paragraphs' numbers (4), the text (66),
  the sources beside it (`aside`) on a page 106 cells wide (a screen of 114) and at the text's end under a `quiet`
  rule where narrower, each column four cells clear of the text. Its head's facts beside the title where the sources'
  column is, under the lead where not; its lead held to `post_lead` at every width, as the page's. The Outline stays
  in sight as the post scrolls, the heading read marked `▌` in `accent` hanging before it in the page's margin, as the
  page marks it (Outline). The panes side by side, filling the text's and the sources' columns, on a page 106 cells
  wide, one under the other narrower, the next first.
- A picture, which a cell cannot hold, is a line each: its alt text in `text`, its address in `faint`, both links to
  the picture, and its caption in `muted`, as the page sets it.
- A widget is its frame, its name and `● live` in the edge, and inside its alt text and the address of the live one. A
  diagram is drawn in box marks, as mermaid-ascii draws a flowchart, in `muted`.
- A digit chooses its tab at once; no key takes a count before it, as the page's vim keys take none (owner,
  2026-11-23).
- The `:` menu names each place by its kind: on a post `tab`, `heading`, `part` and `version`; on the list `post`,
  `tag` and `series`. The page's CommandLine offers the same places on the blog's pages. A heading goes to its
  heading, a series to its group on the list by series, a tag to the list it filters, shown in the list's window, and
  on that list back to the list; the list's head shows the filter reversed, as the page's does. A heading named as
  another place of its page, `help` among them, takes its count, `posts-2`, since a heading is an anchor; a name two
  places share stands once where both lead to one place, as the post read beside the list and the list's own, and for
  each, by its kind, where they lead apart, as a tag and a series named alike, Tab going round both. On the page and in
  the TUI alike, Enter goes to the place Tab chose, while the line holds the name Tab wrote for it, and else to the
  first the menu lists by the name typed.
