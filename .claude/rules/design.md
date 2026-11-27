---
paths:
  - "DESIGN.md"
  - "design/**"
  - "web/src/**"
  - "internal/face/html/**"
  - "internal/face/tui/**"
---

# How a look is designed

The look itself is `DESIGN.md`; its tokens are `design/tokens.yaml`; the steps from the owner's ask to his yes are `.claude/designer.md`, read whole before the first frame.

## The canvas

- Nothing is coded for a look before the owner has picked it from a storyboard on a design canvas.
- Before a design, load the `frontend-design` skill and Impeccable's, and run Impeccable's playbook for the request.
- A design starts from what the element does and every state it can be in, read from the code and from any design of it made before.
- A part already approved is drawn as approved. Variants are drawn only where the choice is open; the one chosen is then drawn in every state.
- It is drawn as a sheet of states: at rest, under the pointer, with the keyboard on it, current, off, a phone's width, and for the TUI a roomy screen and 80×24; in both themes, each captioned with what changed and why.
- It is drawn on the page's own markup and built stylesheet, or on a capture of the TUI itself, never on a lookalike, and looked at in a shot before it is shown.
- A frame shows its edge cases: one item, the most a row holds, the longest title. What is invented to show them says so on the canvas.
- Once the owner picks, the canvas keeps what was chosen, grouped by part, and drops the rest.
- What a design needs and the tokens lack is added to `design/tokens.yaml` in its own logic: named once, with the reason a reader needs it.

## What a frame holds

- Each element answers what its reader does with it on this page. One with no answer, or there because a page like it carries one, is left out.
- Nothing the owner did not name is added, removed, moved or resized; such a change goes to him as a question, apart from the frame.
- A fact, an action or a way to reach him stands once on a page.
- The main thing comes first and takes the larger share; what adds to it stands beside it or under a rule.
- In a group, what the reader acts on comes first, then a rule, then what tells about it.
- Levels show by size, weight and how much each says, never by a name over them or a ground of their own.
- A heading, a page's title too, stands on a row of its own, nothing level with it, no wider than the column it heads.
- The width is filled with content on the page's one grid, its edges meeting those above: never a box stretched or split to fill it, never two equal columns as a magazine's.
- A new part is built from what the page already draws and in its tokens; one that stands out from the page is redrawn.
- A part takes another's look only where it does the same job, and keeps what tells its own kind: a link stays underlined.
- Things of different kinds stand apart: a project is not a role on the work's timeline.
- A hint is smaller and quieter than what stands beside it.
- A control is seen at rest, and its words say what it opens; no count or time that is not true of it.
- A page ends with where its reader goes next.
- A way to every part of the site is seen on the page, so a reader who knows none of its keys finds it, and it is reached without a mouse.
- Words the owner has not given are drafted in two to four wordings, each with its reason, and marked not approved; a fact about him the profile lacks stays a gap until he gives it.
