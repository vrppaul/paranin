// A card, the picture a link's preview shows of a page (DESIGN.md, Social cards). internal/face/card hands it the
// card's facts as JSON in sys.inputs.facts, and the template fits them to the card: it measures the words it sets, cuts
// them where they run past their lines, and breaks a name in block letters where it does not fit on one. A card is
// 1200×630 at 72 ppi, so a point is a pixel.
//
// Two looks, so the résumé and the blog read as two sides of the site: the résumé's side, the résumé, a project and
// the blog's list, a title in the name's block letters over the line under it; and a post, a header across the top,
// its title and lead under it. Every card ends in the way in, at one place on all of them.

#import "tokens.typ": face, colours as written-colours

#let facts = json(bytes(sys.inputs.facts))

// The design's colours, as the tokens write each, #rrggbb, read as colours
#let colours = written-colours.pairs().map(((name, written)) => (name, rgb(written))).to-dict()

#let card-width = 1200pt
#let card-height = 630pt

// The card's rem, the page's 16px grown by 1.75, so a card shown 500px wide on a phone keeps its words at least 17px; a
// row is a rem and a half and a cell three fifths of one, as on the page
#let rem = 28pt
#let row = rem * 1.5
#let cell = rem * 0.6
#let hairline = 2pt
// The design's radius, 0.125rem; and the room inside the way in's chip, wider than the page's so its words stand clear
#let radius = rem * 0.125
#let chip-inset = (x: cell * 2.2, y: row * 0.55)

// The content's side margin, six cells, and the room between the margins
#let side = cell * 6
#let room = card-width - 2 * side

// The space over a title, from the card's top, or from a post's header rule; the gap between a title and its line; and
// where the way in stands, from the card's top, its foot as far over the card's as a feed may crop, 40px
#let space-over-title = 64pt
#let title-gap = row * 0.9
#let way-top = 506pt

// Every word on a card but a title, at one size
#let words-style = (size: rem * 1.5, weight: 400)

// The line under a title on the résumé's side, on two lines at most
#let line-lines = 2

// A cell of the block letters on each card of the résumé's side: the résumé's name and a project's at one size, so
// the two read as one set and a project's line of two fits over the way in; the cards' mark larger, as one short word
#let letter-cells = (resume: 22pt, project: 22pt, list: 40pt)

// A drawing's cell is three fifths as wide as a row is tall, as the page draws one, so a half row is five sixths of a
// cell's width
#let half-row-share = 5 / 6

// A post's header: its height, to its rule, and the cell of the cards' mark at its left
#let header-height = 132pt
#let mark-cell = 11pt

// A post's title: 104px where it stands whole on one line, else 90px on two lines at most; and the lead's lines under a
// title of one line and under a title of two
#let title-style-one-line = (size: rem * 3.7, weight: 700)
#let title-style-two-lines = (size: rem * 3.2, weight: 700)
#let title-lines = 2
#let lead-lines-under-one = 2
#let lead-lines-under-two = 1

// What ends words cut short, and the stop before it that is dropped
#let ellipsis = "…"
#let stop-before-ellipsis = regex("[,.:;]$")

// The gradient of the name's block letters, accent to host, across what it fills; a title set in text takes it softened
#let letters-gradient = gradient.linear(colours.accent, colours.host, relative: "parent")
#let title-gradient = gradient.linear(colours.title_soft_start, colours.title_soft_end, relative: "parent")

#set page(width: card-width, height: card-height, margin: 0pt, fill: colours.background)
#set text(font: face, size: rem, fill: colours.text, ligatures: false, discretionary-ligatures: false)
#set par(leading: row - rem, spacing: row - rem)

// ---------------------------------------------------------------------------------------------------------------------
// Words fitted to their lines

// Words in a style, in an ink
#let styled(style, ink, words) = text(size: style.size, weight: style.weight, fill: ink, words)

// The letters a balanced line may run past its share, so a word that just fits never drops to a line of its own
#let balance-slack = 2

// Words in a style held to the room, on as few lines as it allows, each about as long as the others: the face is mono,
// so a line's width is its letters counted, each as wide as the face's cell
#let balanced(style, ink, words) = {
  let letter = measure(styled(style, ink, "X")).width
  let letters = words.clusters().len()
  let lines = calc.ceil(letters * letter / room)
  block(width: calc.min(room, (calc.ceil(letters / lines) + balance-slack) * letter), styled(style, ink, words))
}

// The first count of words, an ellipsis after them where words are left out, a stop before it dropped
#let words-shown(words, count) = {
  let kept = words.slice(0, count).join(" ")
  if count < words.len() { kept.trim(stop-before-ellipsis) + ellipsis } else { kept }
}

// Words balanced in a style on some lines at most: past them, cut after the last whole word that leaves room for an
// ellipsis
#let fitted(style, ink, words, lines) = context {
  let most = measure(block(width: room, styled(style, ink, range(lines).map(_ => "X").join(linebreak())))).height
  let all = words.split(" ")
  let count = all.len()
  while count > 1 and measure(balanced(style, ink, words-shown(all, count))).height > most {
    count -= 1
  }
  balanced(style, ink, words-shown(all, count))
}

// Whether a title set in text stands whole on one line at its larger size. Measures, so it needs a context
#let stands-on-one-line(title) = measure(styled(title-style-one-line, colours.text, title)).width <= room

// A title set in text, as a post's: in the softened gradient, at its larger size where it stands on one line, else at
// its smaller on two lines at most
#let text-title(title, one-line) = {
  let style = if one-line { title-style-one-line } else { title-style-two-lines }
  fitted(style, title-gradient, title, title-lines)
}

// ---------------------------------------------------------------------------------------------------------------------
// Block letters, as face/card draws them from the name's font (internal/look): each piece of a name its half rows, each
// a column's half filled or not; a line of pieces may break once between two pieces, never inside one

// How many cells wide a piece is, and pieces side by side, a gap between two
#let piece-width(piece) = piece.half_rows.first().len()
#let row-width(pieces, gap) = pieces.map(piece-width).sum() + gap * (pieces.len() - 1)

// How wide the wider of two rows is, the pieces broken before one of them
#let wider-of-two(pieces, at, gap) = calc.max(row-width(pieces.slice(0, at), gap), row-width(pieces.slice(at), gap))

// The rows a line of pieces stands on at a width in cells: one where it fits, else two, broken where the wider of the
// two is narrowest, which may then stand at a smaller cell; a line of one piece stays whole
#let rows-of(line, columns, gap) = {
  let pieces = line.pieces
  if pieces.len() == 1 or row-width(pieces, gap) <= columns { return (pieces,) }
  let at = range(1, pieces.len()).sorted(key: at => wider-of-two(pieces, at, gap)).first()
  (pieces.slice(0, at), pieces.slice(at))
}

// The half rows of a row of pieces, side by side, a gap of empty columns between two
#let joined(pieces, gap) = range(pieces.first().half_rows.len()).map(half-row => {
  pieces.map(piece => piece.half_rows.at(half-row)).intersperse(range(gap).map(_ => false)).flatten()
})

// The runs of a half row's filled cells, each its first column and how many columns it fills
#let filled-runs(filled) = {
  let runs = ()
  let start = none
  for column in range(filled.len() + 1) {
    let on = column < filled.len() and filled.at(column)
    if on and start == none { start = column }
    if not on and start != none {
      runs.push((start: start, length: column - start))
      start = none
    }
  }
  runs
}

// The half rows of rows of pieces, top to bottom, the font's gap of half rows between two rows
#let drawn-half-rows(rows, letters) = {
  let half-rows = ()
  for row in rows {
    if half-rows.len() > 0 { half-rows += range(letters.line_gap_half_rows).map(_ => ()) }
    half-rows += joined(row, letters.gap_cells)
  }
  half-rows
}

// How many half rows from the top hold ink: a font may leave its last half row empty
#let inked(half-rows) = half-rows.len() - half-rows.rev().position(filled => filled.contains(true))

// Words in block letters on the rows they take at a cell, smaller only where a row, or the last with the cursor after
// it, would not fit; in a block as tall as their ink, or as the cursor where one follows them: a column apart, as wide
// as their last letter, on the drawing's last half row
#let block-letters(letters, letter-cell, ink) = {
  let gap = letters.gap_cells
  let rows = letters.lines.map(line => rows-of(line, room / letter-cell, gap)).join()
  let cursor-room = if letters.cursor_cells > 0 { gap + letters.cursor_cells } else { 0 }
  let widest = calc.max(..rows.map(row => row-width(row, gap)), row-width(rows.last(), gap) + cursor-room)
  let shown-cell = calc.min(letter-cell, room / widest)
  let half = shown-cell * half-row-share
  let half-rows = drawn-half-rows(rows, letters)
  let shown-half-rows = if letters.cursor_cells > 0 { half-rows.len() } else { inked(half-rows) }
  let width = calc.max(..half-rows.map(filled => filled.len())) + cursor-room
  block(width: width * shown-cell, height: shown-half-rows * half, {
    for (half-row, filled) in half-rows.enumerate() {
      for run in filled-runs(filled) {
        place(dx: run.start * shown-cell, dy: half-row * half,
          rect(width: run.length * shown-cell, height: half, fill: ink, stroke: none))
      }
    }
    if letters.cursor_cells > 0 {
      place(dx: (half-rows.last().len() + letters.gap_cells) * shown-cell, dy: (half-rows.len() - 1) * half,
        rect(width: letters.cursor_cells * shown-cell, height: half, fill: colours.accent, stroke: none))
    }
  })
}

// ---------------------------------------------------------------------------------------------------------------------
// The card's parts

// The way in, as the page's Command chip: ❯ in the prompt's colour and the command, on a shade below the page in a
// hairline; at its place, from the left margin
#let way-in(command) = place(dx: side, dy: way-top, box(
  fill: colours.sunk, stroke: hairline + colours.muted, radius: radius, inset: chip-inset,
  text(size: words-style.size)[#text(fill: colours.prompt)[❯] #text(fill: colours.text, command)],
))

// The résumé's side: the title in block letters at its kind's cell, the cards' mark on the list's, a name set as a
// post's title is where the font does not draw every letter of it; the line under it, then the way in
#let side-card(facts) = {
  place(dx: side, dy: space-over-title, block(width: room, {
    if facts.title.lines.len() > 0 {
      block-letters(facts.title, letter-cells.at(facts.kind), letters-gradient)
    } else {
      context text-title(facts.title.words, stands-on-one-line(facts.title.words))
    }
    v(title-gap)
    fitted(words-style, colours.text, facts.line, line-lines)
  }))
  way-in(facts.way_in)
}

// A post's header: the cards' mark at its left, the day and the series at its right, each in the middle of the
// header's height by what it shows, the letters by their ink and the words from the cap line to the baseline; a quiet
// rule under it across the whole card
#let post-header(facts) = {
  let mark = block-letters(facts.mark, mark-cell, colours.faint)
  let day = text(fill: colours.muted, facts.day)
  let series = text(fill: colours.faint, " · ") + text(fill: colours.muted, facts.series)
  let when = if facts.series == "" { day } else { day + series }
  place(dx: side, block(width: room, height: header-height, grid(
    columns: (auto, 1fr), rows: header-height, align: (left + horizon, right + horizon),
    mark, text(size: words-style.size, top-edge: "cap-height", bottom-edge: "baseline", when),
  )))
  place(dy: header-height, line(length: card-width, stroke: hairline + colours.quiet))
}

// A post: the header, its title under the rule, the lead under the title, then the way in
#let post-card(facts) = {
  post-header(facts)
  place(dx: side, dy: header-height + space-over-title, block(width: room, context {
    let one-line = stands-on-one-line(facts.title)
    text-title(facts.title, one-line)
    v(title-gap)
    fitted(words-style, colours.text, facts.lead, if one-line { lead-lines-under-one } else { lead-lines-under-two })
  }))
  way-in(facts.way_in)
}

#if facts.kind == "post" { post-card(facts) } else { side-card(facts) }
