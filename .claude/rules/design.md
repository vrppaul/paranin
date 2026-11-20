---
paths:
  - "DESIGN.md"
  - "design/**"
  - "web/src/**"
  - "internal/face/html/**"
  - "internal/face/tui/**"
---

# How a look is designed

The look itself is `DESIGN.md`; its tokens are `design/tokens.yaml`.

- Nothing is coded for a look before the owner has picked it from a storyboard on a design canvas.
- Before a design, load the `frontend-design` skill and Impeccable's.
- A design starts from what the element does and every state it can be in, read from the code and from any design of it made before.
- It is drawn as a sheet of states: at rest, under the pointer, with the keyboard on it, current, off, a phone's width, and for the TUI a roomy screen and 80×24; in both themes, each captioned with what changed and why.
- It is drawn on the page's own markup and built stylesheet, or on a capture of the TUI itself, never on a lookalike, and looked at in a shot before it is shown.
- Variants are drawn only where the choice is open; the one chosen is then drawn in every state.
- Once the owner picks, the canvas keeps what was chosen, grouped by part, and drops the rest.
- What a design needs and the tokens lack is added to `design/tokens.yaml` in its own logic: named once, with the reason a reader needs it.
