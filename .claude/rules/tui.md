---
paths:
  - "internal/face/tui/**"
---

# The TUI

The rules for Go are in `.claude/rules/go.md`.

- A tab is laid out once for a size, and painted once for a theme, a terminal's colours and the search, the frame around it and the window of keys with it.
- On the blog's list a key chooses which group's painted name stands beside the painted rows, as it chooses the rows.
- A key press chooses which painted lines show, fades the page's last two, and writes where the page stands (ruler, scrollbar) and the command line as typed, with the menu of what Tab goes round for it.
- Nothing else is laid out or painted for a key, but for the search and a buffer shown.
- A key that changes what the search line holds, or opens or leaves it, paints the chosen tab's page again; a search confirmed or taken away paints every tab again, once.
- n and N, naming another match, paint again only the lines of the match named before and of the one named now.
- On the blog's list, j, k or a move that takes the page from the tile chosen, choosing another, paint again only the lines of the tile left and of the one chosen.
- Showing another buffer, by `:e`, `:checkhealth` or `:Lazy`, by Enter on a tile of the blog's list, or by the key `q` or `:q` on any page but the résumé, lays out and paints the buffer shown, once.
- A tab shown before a buffer is shown, and still shown after, keeps its layout and its paint.
- What the reader types reaches the screen only as letters a terminal shows as they are: a control character typed is passed over, never echoed.
- The layout says what a piece of text is, a kind; the paint says how each kind looks. No colour, weight or style is written in the layout.
- With sixteen colours or none, what colour alone would tell is told by reversing or underlining as well.
- The facts and the site's words come from the document.
- The TUI's own words are its frame's, as Neovim's and tmux's are: the ruler, the footer's marks round the session and the tabs, the keys it names, the kinds its menu names, `help`.
- On the page the TUI's own words are only Neovim's for a value the document types: `:checkhealth`'s levels (`OK`, `WARNING`, `ERROR`) and `ADVICE:`.
- The page may count `:checkhealth`'s checks by level in those words, as `4 OK · 1 WARNING`: no other word of the TUI's stands with the counts.
