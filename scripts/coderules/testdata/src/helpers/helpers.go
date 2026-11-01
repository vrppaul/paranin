package helpers

import (
	"errors"
	"iter"
	"maps"
	"slices"
	"strings"
	"testing"
	"testing/synctest"
)

func sorted(words []string) []string {
	slices.SortFunc(words, func(left, right string) int { return strings.Compare(left, right) })
	return words
}

func closer(done chan struct{}) func() {
	return func() { close(done) }
}

// One simple statement handed over is a lambda: an assignment, a send.
func closing(release func() error) (err error) {
	defer func() { err = errors.Join(err, release()) }()
	return nil
}

func serving(serve func() error) chan error {
	failures := make(chan error)
	go func() { failures <- serve() }()
	return failures
}

func stepper(count *int) func() {
	return func() { *count++ }
}

func counter() func() int {
	count := 0
	return func() int { // want `a helper with logic is a function of the package`
		count++
		return count
	}
}

func named(value int) int {
	double := func(value int) int { return value * 2 } // want `a function given a name inside another is a helper`
	return double(value)
}

func declared(value int) int {
	var triple = func(value int) int { return value * 3 } // want `a function given a name inside another is a helper`
	return triple(value)
}

type settings struct {
	onClose func()
}

// A literal set as a field is handed over, whichever way the field is set.
func configured(done chan struct{}) settings {
	made := settings{onClose: func() { close(done) }}
	made.onClose = func() { close(done) }
	made.onClose = func() { // want `a helper with logic is a function of the package`
		close(done)
		close(done)
	}
	return settings{onClose: func() { // want `a helper with logic is a function of the package`
		close(done)
		close(done)
	}}
}

func later(done chan struct{}) {
	go func() { // want `a helper with logic is a function of the package`
		<-done
		close(done)
	}()
}

// An iterator is the value its function makes.
func numbers(limit int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for value := range limit {
			if !yield(value) {
				return
			}
		}
	}
}

// So is an iterator of pairs.
func entries(table map[string]int) iter.Seq2[string, int] {
	return func(yield func(string, int) bool) {
		for key, value := range maps.All(table) {
			if !yield(key, value) {
				return
			}
		}
	}
}

func cases(test *testing.T) {
	test.Run("the body handed to Run is the test", func(test *testing.T) {
		test.Log("first")
		test.Log("second")
	})
	test.Run("a helper inside it is still a helper", func(test *testing.T) {
		check := func(value int) bool { // want `a function given a name inside another is a helper`
			value++
			return value > 0
		}
		test.Log(check(1))
	})
	synctest.Test(test, func(test *testing.T) {
		test.Log("first")
		test.Log("second")
	})
}

func benchmarks(benchmark *testing.B) {
	benchmark.Run("a benchmark's body too", func(benchmark *testing.B) {
		benchmark.Log("first")
		benchmark.Log("second")
	})
}

func fuzzing(fuzz *testing.F) {
	fuzz.Fuzz(func(test *testing.T, input string) {
		test.Log(input)
		test.Log(input)
	})
}

type runner struct{}

func (runner) Run(name string, body func()) { body() }

func others(named runner) {
	named.Run("only testing's Run hands over a test", func() { // want `a helper with logic is a function of the package`
		named.Run("inner", nil)
		named.Run("again", nil)
	})
}

// A literal outside every function is no helper inside one.
var _ = func() int {
	count := 1
	return count
}
