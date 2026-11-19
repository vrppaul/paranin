# The design

paran.in is Pavel Paranin's résumé: one profile served as an SSH app, a web page, markdown, a man page and a PDF
(PRODUCT.md). The web page follows the rules of the web and looks like the terminal app: one mono face on a grid of
cells, block letters, the Catppuccin palettes, one accent. Its readers are hirers; the page has done its job when they
write to him.

The owner approved this design on 2026-11-17. Everything here is the target. What holds today is
`design/tokens.yaml`, `web/stylelint.config.js` and the page itself; the page grew into this design part by part,
each part with the tokens and the rules it draws by; `TODO.md` lists the parts still to come.

- A token named here lives in `design/tokens.yaml`, which writes it into `web/src/tokens/tokens.css`; once it is
  there, the file's value is the one that counts, and this document names it without its value.
- A stylesheet names a token, never a value (stylelint, `.claude/rules/page.md`).
- No size is in pixels or another absolute unit, a media query's width neither (stylelint), so the whole page grows
  with the text size its reader set (owner, 2026-11-18): a fixed length is in rem, a space inside a line in cells,
  `ch`. The only pixels are a picture's own, as the favicon's.

## Content

- Write as the profile writes: plain English, short sentences, numbers as figures. The page says "I" only in About;
  elsewhere it states facts.
- Sentence case everywhere. Section names, tabs, fact labels and chips are lower case, as the TUI writes them:
  `experience`, `looking for`. The block-letter titles are capitals because the font has no others.
- A button's words start with a verb and say what happens: "Write to Pavel", "Copy", "Send". A toast says what was
  done, in the past: "Copied ssh paran.in".
- An error says what to fix: "Write the whole address, as name@company.com." Never "Invalid input".
- No small line over a heading, no numbered markers, no emoji, no "→" after a link.
- Every fact comes from `content/profile.yaml`. A fact or a point that is only a gap still to fill (`[later]`) is left
  out of the web until the author fills it; a gap inside a sentence shows `faint`.
- Dates are months, `MM/YYYY` (`08/2020`), with the machine's `YYYY-MM` in `<time datetime>`. A range joins two with an
  en dash and spaces, `08/2020 – 01/2022`; a role still held ends in `present`.
- A heading joins a title and its subtitle with an em dash, `Seznam.cz — Backend Developer`; a line joins its parts
  with a middle dot, `Python · Browser-based code review`; a label ends in a colon where text follows it.
- Lengths of time as the profile writes them: `7y`, `1y 7m`, `11m`, or the words `side` and `learning`.

## The grid

- Space is counted in rows of 1.5rem (`row`) and cells of 1ch (`cell`): vertical space in rows, horizontal space inside
  a line in cells. The text read sits on the rows; every other style's line is its own (Type).
- The page fills the screen up to `content` with `gutter` either side, `gutter_phone` on a phone. Only running prose is
  held to `measure`; the layout itself uses the width.

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
| `when` | `design/tokens.yaml` | an entry's dates column |
| `span_thin` | `design/tokens.yaml` | a role's bar in the phone's career list |
| `bar` | `design/tokens.yaml` | the top bar |
| `dialog` | 35rem | the widest a dialog runs |
| `textarea` | 7.5rem | a message field's least height |

## Sizes

Every control is one of two sizes; the size fixes its height, its padding, the gap inside it and its words. No third
size: a control that seems to need one is the wrong control.

| size | height | either side | gap inside | words | controls |
| --- | --- | --- | --- | --- | --- |
| regular | `control`, two rows | `cell_2` | `cell` | `strong` for an action, `body` for a value | Button, Choice, Command, Field, Toast |
| compact | `control_compact`, one row | `cell` | `nudge` | `caption` | ThemeSwitch, Tooltip |
| tick | three cells, `[x]` or `(*)`, on a row | label `cell` away | | `body` | Checkbox, Radio |

- Corners `radius` on buttons, chips, fields and dialogs; bars, rules and the current tab are square.
- Controls side by side stand `cell_2` apart, rows of them `row_half` apart.
- A dialog pads `row` and a half above, `row` below, `cell_3` either side.

## Colour

- One palette, Catppuccin Latte for light and Macchiato for dark; every colour is a token named for what it does
  (`design/tokens.yaml`), and nothing else is written.
- `background` is the page; `sunk` lies a shade below it (the top bar, a command chip, code); `deep` below that (a
  pressed quiet control).
- `text` for what is read, `muted` for what stands beside it, `faint` for what a reader may pass over. In Latte `muted`
  is subtext1 and `faint` subtext0, one tone darker than the TUI first had them: `muted` reads 4.7:1 or more on
  `background`, `sunk` and `deep`, `faint` 3.7:1 or more, so `faint` is never the only ink of something a reader needs.
  Neither stands on `tint`.
- A fact's label is `muted`. `label` blue rings the focus: blue, and 3:1 or more on every ground in both themes.
- `accent` is spent where the eye should go first: the primary button, the cursor, the current role, the near end of
  the name. In Latte it is a fill and a display colour, never small text (2.64:1).
- `host` is the site's host and the far end of the name. `tint` marks a result worth seeing first. `error`, `passed`
  and `warned` only for states, always with words.
- Borders: `quiet` for rules between parts, `muted` for a control's border, `edge` for the corner crosses and for
  what lies over the page, a tooltip and a toast.
- `scrim` fades the page behind a dialog towards its own ground: Latte's and Macchiato's background at 85%.

## Type

- One face (`face`), monospace, ligatures off, set once on the body for every element to inherit. No second face on
  the résumé and its pages; an article's reading face is chosen with the blog (FOUNDATION §5.6).
- A text style is its size, its line and its weight, set as three properties (`.claude/rules/page.md`).

| style | size / line, rem | weight | for |
| --- | --- | --- | --- |
| `mail`, `mail_phone` | `design/tokens.yaml` | `design/tokens.yaml` | the mail heading the brief, and on a phone |
| `lead`, `lead_phone` | `design/tokens.yaml` | `design/tokens.yaml` | the headline under the name, and on a phone |
| `heading` | `design/tokens.yaml` | `design/tokens.yaml` | an `h2` on a page that is not the résumé |
| `body` | `design/tokens.yaml` | `design/tokens.yaml` | everything read; the default |
| `strong` | `design/tokens.yaml` | `design/tokens.yaml` | a role, a company, a project's name, a button's words |
| `small` | `design/tokens.yaml` | `design/tokens.yaml` | the top bar, a field's hint, the footer |
| `caption` | `design/tokens.yaml` | `design/tokens.yaml` | the career line's names and years, a chip's word, a tooltip; never prose |

- The page's one `h1` and the résumé's five `h2` are block letters; another page's `h2` is `heading`; every heading
  below is `strong`, its level shown by the space before it (`row_2` before `h3`, `row` before `h4` to `h6`). Levels
  never skip.
- Code is told by its `sunk` ground, since the whole page is one face.
- Weights: 400 for text, 700 for what names (a role, a company, a button), 800 only for the mail, its @ left at 400.

## Block letters

- The name and the five sections' titles are drawn in the TUI's block fonts, the same glyphs
  (`internal/look/letters.go`), as SVG rectangles: they need no font and shrink to their column, never past it.
- Natural size: `name_cell` a cell for the name, `title_cell` for a title.
- One gradient, `accent` to `host`, left to right: the only gradient on the page. `shape-rendering: crispEdges`, so no
  seam shows between two rectangles.
- The title font's X is drawn on three cells (▀▄▀ over ▄▀▄), in the TUI too: on five it read as )(.
- The cursor after the name is an underscore a letter wide on the letters' foot, in `accent`, and it blinks, as a
  terminal's does: the one thing on the page that moves on its own, and it stops for a reader who asks for less motion.
- The element carries the words as `aria-label`; the drawing is hidden from screen readers.

## Shape, depth and motion

- Lines are `hairline`. No shadows: depth is a shade of ground (`sunk`, `deep`) and a hairline.
- What stands over the page as it scrolls lies on a layer of `design/tokens.yaml`, the layers listed lowest first:
  the top bar on `bar`.
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
| disabled | `sunk` ground or `quiet` border, `faint` words, `aria-disabled`, and the page says why |
| error | an `error` border and words under the control saying what to fix |
| busy | the words say what is happening ("Sending…") and the control waits |

- A visited link looks as any link.

## Layout

- Below `wide` (80rem) the brief narrows to `brief_mid` and a section's title stands above its content.
- Below `phone` (53.75rem) everything is one column, the gutters `gutter_phone`, the top bar scrolls away, the career
  line becomes a list, a table scrolls inside itself. No page is ever wider than its screen.
- The two widths are the only ones a media query may hold, in rem, since a custom property cannot stand in one;
  stylelint holds them from the first media query on, and refuses any width until then. A media query's rem is the
  reader's text size too, so a reader who sets larger text gets the narrower layout sooner.
- Every page has a skip link, one `header` (the top bar), one `main` and one `footer`. A section a link leads to stops
  a row below the sticky bar.
- A position worked out from data (the career line) is an SVG attribute: the site's policy refuses inline styles.

## Iconography

- No icon set. The page uses the terminal's glyphs: `[x]` and `(*)` for a checkbox and a radio, ❯ before a command
  (`prompt`), the cursor after the name, ✓ in a toast (`passed`), and `-` before a point (`faint`).
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

A top-level section's `h2` in the title font: only About, Experience, Projects, Skills and Contact. A subsection
(Spoken languages, Education, a skill group) is a `strong` heading.

### Mail

The address as the site writes it: the box in `accent`, @ in `faint` at the body's weight, the host in `host`. It heads
the brief in `mail`; elsewhere the same three colours at `body` size. A `mailto:` link; under the pointer the box and
the host underline, `underline_hover` thick, `nudge` below.

### Favicon

The first letter of the author's name in the name font, the P, its cells square, in the name's gradient, and a cursor
beside it, a beam in `accent` as tall as the letter's lower three half rows, on the page's ground, a square. As SVG with
crisp edges, in the reader's theme; a PNG of 180px for a phone's home screen, the letter a tenth of its side in from
each edge so the phone's rounded corners cut none of it, and one of 32px at /favicon.ico, both in the dark theme. The
picture a shared link shows is Typst's, drawn with the CV (ADR-0005).

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
Never coloured: the underline says it is a link. Nothing is marked with an arrow.

### TopBar

The site, the five sections, the theme button. `sunk` ground, a `quiet` hairline under it, `bar` tall. It sticks to the
top on a wide screen; on a phone it scrolls away and the sections take a second row, wrapping to a third on a screen
too narrow for them all (below about 375px). The current section is reversed,
as the TUI's tab, and follows the reader's scroll. Section names are lower case.

### ThemeSwitch

One quiet button at the end of the top bar: the half-filled circle, no words, a `control_compact` square, no border and
no ground, `muted` turning `text` under the pointer. The page starts in the system's theme; a press turns it to the
other, sets `data-theme` on the root and keeps it for the next visit; a press that lands back on the system's theme
forgets the choice. It is a toggle button named "Dark theme", `aria-pressed` while the page is dark, its name in a
Tooltip under the pointer. Without the script it is hidden and the page follows the system.

### SkipLink and Footer

- The skip link is the first thing a keyboard reaches: hidden until focused, then at the top left over the bar,
  reversed, regular size.
- The footer ends every page: a `quiet` hairline above, `row_2` of air, `small` and `muted`, the ways the same résumé
  is served: the commands that reach it as Code (ssh, curl), its files as links (markdown, llms.txt, cv.pdf). Social
  profiles are in Contact.

### Facts

Label and value, two cells apart: the label `muted`, one or two lower-case words as the profile writes them; the value
in `text`, on the label's baseline. On the first screen the first four the profile lists, its author's choice
(experience, stack, now, languages); the rest stand in About under the summary (owner, 2026-11-18). On a phone a
value that is a Command stands under its label across the whole width.

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
summary, its facts. An install line is a fact whose value is Code alone; the page draws it as a Command. Cards are
`row_2` apart with a `quiet` hairline between, never boxed.

### Table

Rows of a few cells, as a skill group's skills: the first cell `strong`, the rest `text`, a length `muted`; cells
`cell_3` apart, rows set apart by a `quiet` hairline; numbers line up (`tabular-nums`). No header row, no stripes. A cell but the last keeps to one line, so on a narrow screen the table scrolls inside itself. Whatever
scrolls inside itself, a table, a Command, a code block, shows a thin scrollbar.

### Code

In a line, Code on a `sunk` ground; on lines of its own, a block padded `row_half` and `cell_2`, `radius`, scrolling
sideways inside itself, never wrapping.

### Contents and PageHead

- Contents: the pages a page leads to, one a line, `row_half` apart, each a link by its title, then a colon and its
  line in `muted`. No bullets.
- PageHead: a page that is not the résumé opens with its title as the `h1` in the title font, its line in `lead`, what
  it is made of as `muted` words joined by ` · ` (not chips: they do nothing when pressed), its lead, and where its
  code lives as Facts.

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

A window over the page, as tmux draws a popup: a `muted` hairline frame on the page's ground, the title set into its top
edge on the left and the close cross into it on the right, both half above the frame. The page behind fades by `scrim`.
A native `dialog` opened with `showModal()`: focus moves in, Esc or the cross closes it, focus goes back. The cross is
named "Close"; its tooltip says Esc does the same. Never opens on its own.

### Tooltip and Toast

- Tooltip: a word or two about a control, on hover and on focus; `sunk` ground, an `edge` hairline, `caption`; never
  the only place something is said.
- Toast: ✓ in `passed`, then what was done, in the past tense, at the foot of the screen for `toast`; regular size,
  `sunk` ground and an `edge` hairline, as the Tooltip's. Errors are not toasts: they stand where they happened.

### Frame and Section

- Four crosses mark the first screen's corners, `cross` in `edge`, half a gutter outside the content and `row_2` in
  from its top and foot: only there, and not on a phone, whose gutter leaves them no room.
- A section: its block title in the left column, its content in the right, as wide as an Entry: `when`, `cell_3` and
  `measure`, the title taking what is left; sections `row_4` apart, `row_2` on a phone. Its blocks stand a row apart,
  a subsection two. Its `id` is its tab's name, so the top bar links to it.

## Pages

- The first screen: left, the greeting, the Name and the headline in `lead`; right, `brief` wide, the Mail, the role,
  the place, the Facts, the primary button and the Command, then cv.pdf, GitHub and LinkedIn. Below, the CareerLine
  across the whole width; the four crosses at the corners. Nothing more: projects, skills and each role's details come
  further down.
- The address of no page: its status in block letters, the address quoted as Code, a link to the start; no apology and
  no joke. The top bar and the footer stand as on every page.
