---
id: 1
title: Reading slowly
description: Why a page read twice is read once.
status: published
published: 2026-11-10T09:00+01:00
updated: 2026-11-12
tags: [reading, terminals]
discussions:
  - {site: Lobsters, address: "https://lobsters.example.test/s/slow"}
---
A page read twice is read once, if the first reading was *slow*.[^source]

## What a reader keeps

| Reading | Kept |
| --- | --- |
| Fast | Little |
| Slow | **Most** |

- the margins
- the notes
  - in pencil[^pencil]
  - in ink

> A book is read with a pencil.
>
> Twice, with a pen.

![A noted margin](margin.png "The margin, where a slow reader writes back.")

### In a terminal

3. open the file
4. read it with `less`

```go
// slowly reads a page twice, aloud
func slowly(page string) int {
	fmt.Println("slowly", page)
	return len(page) * 2
}
```

```sh
less "$HOME/notes"
```

See [the reader's notes](https://notes.example.test/slow), or write to <mira@example.test>.[^source]

## Afterwards

A page kept is a page read again, as [what a reader keeps](/blog/reading-slowly#what-a-reader-keeps) says, and
[the tool](/p/tidy-tool) that tidies it.

***

What stays is what was argued with.

[^pencil]: A soft one, which rubs out.
[^source]: Mira Example, on her own reading.
