// Command tokens writes the design's tokens, design/tokens.yaml, into internal/look/palette.go, the Go the faces
// read them from (docs/FOUNDATION.md §5.5), and checks that the file there is the one the tokens make. Each colour
// becomes a constant of look.Colour and a shade in each theme for each kind of terminal: its value whole, the
// nearest of xterm's 256, and the one of sixteen the file names; a blend is worked out into a colour of its own.
//
//	go run ./scripts/tokens           checks internal/look/palette.go against the tokens: the pre-commit hook
//	go run ./scripts/tokens -write    writes it, as `go generate ./internal/look` does
//
// Run it in the repository's root. It prints a fault of the tokens, or that the generated file is out of step,
// and exits 0 when the file is the one the tokens make or has just been written, 1 when it is not or the tokens
// have a fault, and 2 when the command is written wrong.
package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// What the program exits with
const (
	succeeded     = 0
	failed        = 1
	usageMistaken = 2
)

// How the command is written, and the files it reads and writes, from the repository's root
const (
	writeFlag   = "-write"
	usage       = "usage: go run ./scripts/tokens [-write]"
	tokensPath  = "design/tokens.yaml"
	palettePath = "internal/look/palette.go"
	// the generated file may be read by anyone, as the rest of the tree
	paletteMode = 0o644
)

// What the program says when the generated file is out of step
const outOfStep = "%s is out of step with %s: run go generate ./internal/look\n"

func main() {
	os.Exit(run(os.Args[1:], "."))
}

// run checks or writes the generated file under a root, as the arguments say.
func run(arguments []string, root string) int {
	writing, understood := modeOf(arguments)
	if !understood {
		fmt.Fprintln(os.Stderr, usage)
		return usageMistaken
	}
	folder := os.DirFS(root)
	generated, err := generate(folder)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return failed
	}
	target := filepath.Join(root, palettePath)
	if writing {
		return write(target, generated)
	}
	return check(folder, generated)
}

// modeOf is whether the arguments ask for the file to be written, and whether they say anything else.
func modeOf(arguments []string) (writing, understood bool) {
	switch len(arguments) {
	case 0:
		return false, true
	case 1:
		return arguments[0] == writeFlag, arguments[0] == writeFlag
	default:
		return false, false
	}
}

// generate is the Go the tokens of a folder make.
func generate(folder fs.FS) ([]byte, error) {
	data, err := fs.ReadFile(folder, tokensPath)
	if err != nil {
		return nil, fmt.Errorf("read the tokens: %w", err)
	}
	read, err := readTokens(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", tokensPath, err)
	}
	return source(paletteOf(read))
}

func write(target string, generated []byte) int {
	if err := os.WriteFile(target, generated, paletteMode); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return failed
	}
	return succeeded
}

func check(folder fs.FS, generated []byte) int {
	written, err := fs.ReadFile(folder, palettePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return failed
	}
	if !bytes.Equal(written, generated) {
		fmt.Fprintf(os.Stderr, outOfStep, palettePath, tokensPath)
		return failed
	}
	return succeeded
}
