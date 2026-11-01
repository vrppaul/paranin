package records

type pair [2]string // want `an array of words stands in for a record`

type label string

func pairs() [][2]string { // want `an array of words stands in for a record`
	return nil
}

func labels() [3]label { // want `an array of words stands in for a record`
	return [3]label{} // want `an array of words stands in for a record`
}

func anything() [2]any { // want `an array of words stands in for a record`
	return [...]any{1, "one"} // want `an array of words stands in for a record`
}

// An array of numbers is a vector or a digest, and a slice of words is a list: neither is a record.
func digest() ([32]byte, [3]float64, []string) {
	return [32]byte{}, [3]float64{}, nil
}
