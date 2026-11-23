# Reading slowly

2026-11-10

reading · terminals

> Why a page read twice is read once.

A page read twice is read once, if the first reading was *slow*.[^1]

## What a reader keeps

| Reading | Kept |
| --- | --- |
| Fast | Little |
| Slow | **Most** |

- the margins
- the notes
  - in pencil[^2]
  - in ink

> A book is read with a pencil.
>
> Twice, with a pen.

![A noted margin](https://example.test/blog/reading-slowly/margin.png "The margin, where a slow reader writes back.")

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

See [the reader's notes](https://notes.example.test/slow), or write to [mira@example.test](mailto:mira@example.test).[^1]

## Afterwards

A page kept is a page read again, as [what a reader keeps](https://example.test/blog/reading-slowly.md#what-a-reader-keeps) says, and [the tool](https://example.test/p/tidy-tool.md) that tidies it.

***

What stays is what was argued with.

[^1]: Mira Example, on her own reading.
[^2]: A soft one, which rubs out.
