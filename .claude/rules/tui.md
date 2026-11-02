---
paths:
  - "internal/face/tui/**"
---

# The TUI

How the TUI is built is said in the comment of `internal/face/tui`; the rules for Go are in `.claude/rules/go.md`.

- A tab is laid out once for a size, and painted once for a theme, a terminal's colours and the search, the frame around it and the window of keys with it.
- A key press chooses which painted lines show and writes the ruler, the command line and the window of commands; nothing else is laid out or painted for a key, but for the search and a buffer shown.
- A key that changes what the search line holds, or opens or leaves it, paints the chosen tab's page again; a search confirmed or taken away paints every tab again, once.
- Showing another buffer, by `:e` or `:checkhealth`, or by the key `q` or `:q` on any page but the résumé, lays out and paints the buffer shown, once.
- What the reader types reaches the screen only as letters a terminal shows as they are: a control character typed is passed over, never echoed.
- The layout says what a piece of text is, a kind; the paint says how each kind looks. No colour, weight or style is written in the layout.
- Every colour comes from `look`, which `scripts/tokens` writes from `design/tokens.yaml`: change the tokens and run `go generate ./internal/look`, never the palette by hand.
- With sixteen colours or none, what colour alone would tell is told by reversing or underlining as well.
- Each key is a named constant, and each key's behaviour a test through the model: keys in, the screen's text out, its colours stripped.
- The facts and the site's words come from the document. The TUI's own words are those of the frame it draws, as Neovim's are: the mode, the ruler (`All`, `Top`, `Bot`), the keys the status line names.
- On the page the TUI's own words are only Neovim's for a value the document types: `:checkhealth`'s levels (`OK`, `WARNING`, `ERROR`) and `ADVICE:`.
