# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

The primary reader is someone who hires: a recruiter, a technical recruiter, an engineering manager or a team lead weighing the author for a role. They arrive from a link (a CV, LinkedIn, a post), or out of curiosity after `ssh paran.in`. Engineers who meet the site through the author's posts are a second audience. The author is a third: he edits the content and reads every face while he writes it.

## Product Purpose

paranin presents one engineer's professional record, his experience, projects and skills, so that whoever hires can see who he is and what he has built, and then writes to him or asks him to an interview. Success is that message.

## Positioning

One source of facts, drawn in every form a technical reader uses: an SSH TUI (`ssh paran.in`), a web page, markdown for agents, a man page for `curl`, `llms.txt` and a PDF; later a blog drawn in each of them. What the reader comes for is the facts of the résumé. The forms they arrive in are how they reach those facts, not the point.

## Operating Context

- The SSH TUI runs in the reader's own terminal and SSH client, down to about 80×24, mobile clients such as Termius included. It takes its theme from the terminal's background and copes with 256, 16 or no colours.
- The web page reads whole without JavaScript, under a CSP that allows only the site's own files.
- `curl` and agents choose their format by the Accept header.
- One profile in `content/` holds every fact. Each face draws it from one document model.

## Capabilities and Constraints

- The TUI moves between tabs and scrolls, and offers `/` search, `:` commands (`:e <project>`, `:checkhealth`, `:Lazy`) and a window of keys; `y` copies a value over OSC 52.
- The web page has a theme choice, a hint with the command for each way a terminal reads it, and vim keys: `:` to go to a place, `/` to find on the page. The contact form comes later, its salary range the first island.
- Planned: a blog, an LLM grounded in the content, and a CV tailored to a job description. Neither of the last two ever invents experience.
- The theme choice on the web is one button: the page starts in the system's theme, and a choice that lands back on it is forgotten (owner, 2026-11-17; `DESIGN.md`, ThemeSwitch).

## Brand Commitments

These hold in every face: TUI, web, PDF and social cards (docs/FOUNDATION.md §3.4).

- Monospace type on a grid of cells and rows.
- Catppuccin Latte for the light theme and Macchiato for the dark one; every colour comes from `design/tokens.yaml`.
- The cursor after the author's name: `▎` in the TUI, PDF and cards; on the web page an underscore that blinks, since beside the block letters `▎` read as one more letter (owner, 2026-11-17).
- Corner crosses `+`.
- The vim and tmux metaphor: modes, a status line, `:` commands and tabs.

The owner confirmed these as binding on 2026-11-06. A redesign changes the look inside them.

## Evidence on Hand

- The real profile: `content/profile.yaml`, with the site's words in `content/i18n/en.yaml`.
- It still holds placeholders (`[later]`, `[remote?]`, `[notice period]`, `[service]`, `[N]`, `[20XX]`) that only the author fills. Nothing may stand in for them.
- No testimonials, client names or metrics beyond what the profile states.

## Product Principles

- The facts lead: every screen answers "who is this and what has he built" before it shows anything off.
- One truth, many faces: no face holds a fact another lacks, and no fact lives outside the content.
- Make writing to the author easy: the ways to reach him are always within a step.
- Restraint over decoration: the medium is distinctive enough, and the craft shows in precision.

## Accessibility & Inclusion

- Keyboard-only use throughout.
- The TUI works at 80×24.
- With 16 colours or none, what colour would tell is told by reversing or underlining as well.
- The web page reads without JavaScript.
- A plain mode for screen readers and constrained terminals is planned (REQUIREMENTS NFR-7).
