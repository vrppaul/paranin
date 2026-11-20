package proxies

import "testing"

// A test is found by its name; a helper of the tests is not.
func TestShout(test *testing.T) {
	test.Log(factMark)
}

func shoutTwice(text string) string { // want `shoutTwice only passes its call on`
	return shout(text)
}
