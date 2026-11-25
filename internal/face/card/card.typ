// A card, the picture a link's preview shows of a page (DESIGN.md, Social cards). internal/face/card hands it the
// card's facts as JSON in sys.inputs.facts, block letters already laid out as rectangles, and the template fits them to
// the card: it measures the words it sets and cuts them where they run past their lines, chooses a title's size, and
// chooses how a name in block letters stands and at what cell. It states each choice as metadata labelled <choice>,
// which the face's tests query. A card's size in pixels is handed in as sys.inputs.width and height; at 72 ppi a point
// is a pixel.
//
// Two looks, so the résumé and the blog read as two sides of the site: the résumé's side, the résumé, a project and
// the blog's list, a title in the name's block letters over the line under it; and a post, a header across the top,
// its title and lead under it. Every card ends in the way in, at one place on all of them.

#import "tokens.typ": face, colours as written-colours, rems

#let facts = json(bytes(sys.inputs.facts))

// The design's colours, as the tokens write each, #rrggbb, read as colours
#let colours = written-colours.pairs().map(((name, written)) => (name, rgb(written))).to-dict()

#let card-width = int(sys.inputs.width) * 1pt
#let card-height = int(sys.inputs.height) * 1pt

// The page's rem at a browser's own text size, and the card's, grown by 1.75, so a card shown 500px wide on a phone
// keeps its words at least 17px
#let page-rem = 16pt
#let rem = page-rem * 1.75

// The page's row, radius and hairline, as the tokens write them in rem, at the card's rem
#let row = rem * rems.row
#let radius = rem * rems.radius
#let hairline = rem * rems.hairline

// The content's side margin, in cells
#let margin-cells = 6

// The space over a title, from the card's top, or from a post's header rule; the gap between a title and its line;
// and where the way in stands, from the card's top
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

// A post's header: its height, to its rule, and the cell of the cards' mark at its left
#let header-height = 132pt
#let mark-cell = 11pt

// A post's title: its larger size where it stands whole on one line, else its smaller on two lines at most; and the
// lead's lines under a title of one line and under a title of two
#let title-style-one-line = (size: rem * 3.7, weight: 700)
#let title-style-two-lines = (size: rem * 3.2, weight: 700)
#let title-lines = 2
#let lead-lines-under-one = 2
#let lead-lines-under-two = 1

// The room inside the way in's chip, in cells at its sides and rows over and under its words, more than the page's,
// so its words stand clear
#let chip-inset-cells = 2.2
#let chip-inset-rows = 0.55

// What ends words cut short, and the stop before it that is dropped
#let ellipsis = "…"
#let stop-before-ellipsis = regex("[,.:;]$")

// The letters a balanced line may run past its share, so a word that just fits never drops to a line of its own
#let balance-slack = 2

// The gradient of the name's block letters, accent to host, across what it fills; a title set in text takes it softened
#let letters-gradient = gradient.linear(colours.accent, colours.host, relative: "parent")
#let title-gradient = gradient.linear(colours.title_soft_start, colours.title_soft_end, relative: "parent")

#set page(width: card-width, height: card-height, margin: 0pt, fill: colours.background)
#set text(font: face, size: rem, fill: colours.text, ligatures: false, discretionary-ligatures: false)
#set par(leading: row - rem, spacing: row - rem)

// A cell of the card, as the page's is the face's ch: the width of its zero at the card's rem; the side margin, and
// the room between the margins. Each measures, so it needs a context
#let cell() = measure(text(size: rem, "0")).width
#let side() = cell() * margin-cells
#let room() = card-width - 2 * side()

// A choice the card made, stated for whoever queries it: the part it made it for and what it chose
#let chosen(choice) = [#metadata(choice) <choice>]

// ---------------------------------------------------------------------------------------------------------------------
// Words fitted to their lines

// Words in a style, in an ink
#let styled(style, ink, words) = text(size: style.size, weight: style.weight, fill: ink, words)

// Words in a style held to the room, on as few lines as it allows, each about as long as the others: the face is mono,
// so a line's width is its letters counted, each as wide as the face's cell
#let balanced(style, ink, words) = {
  let letter = measure(styled(style, ink, "X")).width
  let letters = words.clusters().len()
  let lines = calc.ceil(letters * letter / room())
  block(width: calc.min(room(), (calc.ceil(letters / lines) + balance-slack) * letter), styled(style, ink, words))
}

// Words with the first one wider than the room cut inside it, after the last letter that leaves room for an ellipsis,
// and the words after it left out, so nothing runs past the margin
#let within-room(style, ink, all) = {
  let wide = all.position(word => measure(styled(style, ink, word)).width > room())
  if wide == none { return all }
  let letters = all.at(wide).clusters()
  let kept = letters.len() - 1
  while kept > 1 and measure(styled(style, ink, letters.slice(0, kept).join() + ellipsis)).width > room() {
    kept -= 1
  }
  all.slice(0, wide) + (letters.slice(0, kept).join() + ellipsis,)
}

// The first count of words, an ellipsis after them where words are left out, a stop before it dropped
#let words-shown(words, count) = {
  let kept = words.slice(0, count).join(" ")
  if count < words.len() { kept.trim(stop-before-ellipsis) + ellipsis } else { kept }
}

// How tall some lines of words in a style stand in the room, measured on as many lines of one letter
#let lines-height(style, ink, lines) = {
  measure(block(width: room(), styled(style, ink, range(lines).map(_ => "X").join(linebreak())))).height
}

// How many lines words balanced in a style take
#let line-count(style, ink, shown) = {
  let height = measure(balanced(style, ink, shown)).height
  let count = 1
  while lines-height(style, ink, count) < height {
    count += 1
  }
  count
}

// Words balanced in a style on some lines at most, for a part of the card: past them, cut after the last whole word
// that leaves room for an ellipsis; nothing where there are none. It states the words shown, whether they were cut and
// how many lines they take
#let fitted(part, style, ink, words, lines) = context {
  if words == "" { return chosen((part: part, words: words, cut: false, lines: 0)) }
  let most = lines-height(style, ink, lines)
  let all = within-room(style, ink, words.split(" "))
  let count = all.len()
  while count > 1 and measure(balanced(style, ink, words-shown(all, count))).height > most {
    count -= 1
  }
  let shown = words-shown(all, count)
  chosen((part: part, words: shown, cut: shown != words, lines: line-count(style, ink, shown)))
  balanced(style, ink, shown)
}

// Whether a title set in text stands whole on one line at its larger size. Measures, so it needs a context
#let stands-on-one-line(title) = measure(styled(title-style-one-line, colours.text, title)).width <= room()

// A title set in text, as a post's: in the softened gradient, at its larger size where it stands on one line, else at
// its smaller on two lines at most
#let text-title(title, one-line) = {
  let style = if one-line { title-style-one-line } else { title-style-two-lines }
  fitted("title", style, title-gradient, title, title-lines)
}

// ---------------------------------------------------------------------------------------------------------------------
// Block letters, each arrangement of them laid out by face/card from internal/look as the page's are: its lines'
// words, its width in cells, how far down its ink reaches in half rows, its runs of filled cells, its cursor, and a
// cell's shape, as wide and as tall as its units

// The arrangement block letters stand in at a width in cells: the first, on one line, where it fits or is the only
// one; else the narrowest of the rest, the name broken onto two lines where the wider of them is narrowest
#let arranged(arrangements, columns) = {
  let whole = arrangements.first()
  if whole.width <= columns or arrangements.len() == 1 { return whole }
  arrangements.slice(1).sorted(key: drawing => drawing.width).first()
}

// Block letters for a part of the card in an arrangement at a cell, smaller only where it would not fit, in a block as
// tall as their ink, the cursor among it where one follows them. It states the rows they stand on and whether they
// shrank
#let block-letters(part, arrangements, letter-cell, ink) = context {
  let drawing = arranged(arrangements, room() / letter-cell)
  let shown-cell = calc.min(letter-cell, room() / drawing.width)
  let half = shown-cell * drawing.half_row_units / drawing.cell_units
  chosen((part: part, rows: drawing.lines, shrank: shown-cell < letter-cell))
  block(width: drawing.width * shown-cell, height: drawing.foot * half, {
    for run in drawing.rectangles {
      place(dx: run.left * shown-cell, dy: run.top * half,
        rect(width: run.width * shown-cell, height: run.height * half, fill: ink, stroke: none))
    }
    let cursor = drawing.cursor
    if cursor != none {
      place(dx: cursor.left * shown-cell, dy: cursor.top * half,
        rect(width: cursor.width * shown-cell, height: cursor.height * half, fill: colours.accent, stroke: none))
    }
  })
}

// ---------------------------------------------------------------------------------------------------------------------
// The card's parts

// The way in, as the page's Command chip: ❯ in the prompt's colour and the command, on a shade below the page in a
// hairline; at its place, from the left margin
#let way-in(command) = context place(dx: side(), dy: way-top, box(
  fill: colours.sunk, stroke: hairline + colours.muted, radius: radius,
  inset: (x: cell() * chip-inset-cells, y: row * chip-inset-rows),
  text(size: words-style.size)[#text(fill: colours.prompt)[❯] #text(fill: colours.text, command)],
))

// The résumé's side: the title in block letters at its kind's cell, the cards' mark on the list's, a name set as a
// post's title is where the font does not draw every letter of it; the line under it, then the way in
#let side-card(facts) = {
  context place(dx: side(), dy: space-over-title, block(width: room(), {
    let title = facts.title
    if title.arrangements.len() > 0 {
      block-letters("title", title.arrangements, letter-cells.at(facts.kind), letters-gradient)
    } else {
      text-title(title.words, stands-on-one-line(title.words))
    }
    v(title-gap)
    fitted("line", words-style, colours.text, facts.line, line-lines)
  }))
  way-in(facts.way_in)
}

// A post's header: the cards' mark at its left, the day and the series at its right, each in the middle of the
// header's height by what it shows, the letters by their ink and the words from the cap line to the baseline; a quiet
// rule under it across the whole card
#let post-header(facts) = {
  let day = text(fill: colours.muted, facts.day)
  let series = text(fill: colours.faint, " · ") + text(fill: colours.muted, facts.series)
  let when = if facts.series == "" { day } else { day + series }
  context place(dx: side(), block(width: room(), height: header-height, grid(
    columns: (auto, 1fr), rows: header-height, align: (left + horizon, right + horizon),
    block-letters("mark", (facts.mark,), mark-cell, colours.faint),
    text(size: words-style.size, top-edge: "cap-height", bottom-edge: "baseline", when),
  )))
  place(dy: header-height, line(length: card-width, stroke: hairline + colours.quiet))
}

// A post: the header, its title under the rule, the lead under the title, then the way in
#let post-card(facts) = {
  post-header(facts)
  context place(dx: side(), dy: header-height + space-over-title, block(width: room(), {
    let one-line = stands-on-one-line(facts.title)
    text-title(facts.title, one-line)
    v(title-gap)
    fitted("lead", words-style, colours.text, facts.lead, if one-line { lead-lines-under-one } else { lead-lines-under-two })
  }))
  way-in(facts.way_in)
}

#if facts.kind == "post" { post-card(facts) } else { side-card(facts) }
