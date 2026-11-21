// Command tokens writes the design's tokens, design/tokens.yaml, into the files the faces read them from
// (docs/FOUNDATION.md §5.5), and checks that the files there are the ones the tokens make:
//
//   - internal/look/palette.go, for the TUI: each colour becomes a constant of look.Colour and a shade in each theme
//     for each kind of terminal, its value whole, the nearest of xterm's 256, and the one of sixteen the file names;
//     a blend is worked out into a colour of its own; then the page's own colour in each theme, the ground of the
//     site's icon, which the page's face draws in Go;
//   - web/src/tokens/tokens.css, for the page: each colour becomes a custom property, its value in each theme, and a
//     blend its colour at its strength, which the browser lays over whatever is under it; then what the page alone
//     reads: the face, each text style as its size, its line's height and its weight, each length, each time, and
//     each layer as its place among the layers, the z-index of what stands on it;
//   - web/src/tokens/times.ts, for the page's script: each time in milliseconds, a number a script that waits reads,
//     since the build writes the stylesheet's times the shortest way.
//
// Its usage:
//
//	go run ./scripts/tokens           checks every file against the tokens: the pre-commit hook
//	go run ./scripts/tokens -write    writes them, as `go generate ./internal/look` does
//
// It refuses, as a fault of the tokens, a colour that reads under the least contrast it needs over a ground it names.
//
// Run it in the repository's root. It prints a fault of the tokens, or which generated file is out of step, and
// exits 0 when the files are the ones the tokens make or have just been written, 1 when one is not or the tokens
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
	writeFlag      = "-write"
	usage          = "usage: go run ./scripts/tokens [-write]"
	tokensPath     = "design/tokens.yaml"
	palettePath    = "internal/look/palette.go"
	stylesheetPath = "web/src/tokens/tokens.css"
	timesPath      = "web/src/tokens/times.ts"
	// a generated file may be read by anyone, as the rest of the tree
	generatedMode = 0o644
)

// What the program says when a generated file is out of step
const outOfStep = "%s is out of step with %s: run go generate ./internal/look\n"

// generatedFile is a file the tokens make: where it lives, from the repository's root, and what it holds.
type generatedFile struct {
	path string
	body []byte
}

func main() {
	os.Exit(run(os.Args[1:], "."))
}

// run checks or writes the generated files under a root, as the arguments say.
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
	code := succeeded
	for _, file := range generated {
		if writing {
			code = max(code, write(filepath.Join(root, file.path), file.body))
			continue
		}
		code = max(code, check(folder, file))
	}
	return code
}

// modeOf is whether the arguments ask for the files to be written, and whether they say anything else.
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

// generate is every file the tokens of a folder make.
func generate(folder fs.FS) ([]generatedFile, error) {
	data, err := fs.ReadFile(folder, tokensPath)
	if err != nil {
		return nil, fmt.Errorf("read the tokens: %w", err)
	}
	read, err := readTokens(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", tokensPath, err)
	}
	palette, err := source(read)
	if err != nil {
		return nil, err
	}
	return []generatedFile{
		{path: palettePath, body: palette},
		{path: stylesheetPath, body: stylesheet(read)},
		{path: timesPath, body: timesScript(read)},
	}, nil
}

func write(target string, generated []byte) int {
	if err := os.WriteFile(target, generated, generatedMode); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return failed
	}
	return succeeded
}

func check(folder fs.FS, generated generatedFile) int {
	written, err := fs.ReadFile(folder, generated.path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return failed
	}
	if !bytes.Equal(written, generated.body) {
		fmt.Fprintf(os.Stderr, outOfStep, generated.path, tokensPath)
		return failed
	}
	return succeeded
}
