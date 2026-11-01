package numbers

import "time"

// A number named in a constant says what it is.
const (
	width   = 80
	timeout = 30 * time.Second
)

type digest [32]byte // want `a number with a meaning is a named constant`

func sizes(items []int) []int {
	limit := 24 // want `a number with a meaning is a named constant`
	limit += 3  // want `a number with a meaning is a named constant`
	const height = 24
	_ = items[2] // want `a number with a meaning is a named constant`
	scale := 1.0
	_ = scale
	return append(items, width, height, limit, 0, 1, len(items)-1)
}

func blended(share float64) (float64, complex128) {
	return share * 0.16, 2i // want `a number with a meaning is a named constant` `a number with a meaning is a named constant`
}

func screens() []int {
	return []int{80, 24} // want `a number with a meaning is a named constant` `a number with a meaning is a named constant`
}
