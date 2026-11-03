// Command templates formats the templ templates under internal and writes the Go each one makes, or checks that
// both are as templ writes them (docs/FOUNDATION.md §6.2). templ's own commands hold no check of the format alone:
// `templ fmt -fail` rewrites the template it fails on, and a gate never changes the working tree, so the format is
// checked here, on each template formatted in memory; the Go is checked by `templ generate -check`. Both run in
// this process, through templ's own commands at the version go.mod names. The Go each template makes names its
// template by its path under internal, so it is written and checked from there alone.
//
//	go run ./scripts/templates           checks every template and its Go: the pre-commit hook
//	go run ./scripts/templates -write    formats each and writes its Go, as `go generate ./internal/face/html` does
//
// Run it in the repository's root. It prints each template out of its format and each Go file out of step with
// its template, and exits 0 when every one is as templ writes it or has just been written, 1 when one is not or
// templ fails, and 2 when the command is written wrong.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/a-h/templ/cmd/templ/fmtcmd"
	"github.com/a-h/templ/cmd/templ/generatecmd"
)

// What the program exits with
const (
	succeeded     = 0
	failed        = 1
	usageMistaken = 2
)

// How the command is written, and where the templates are, from the repository's root
const (
	writeFlag       = "-write"
	usage           = "usage: go run ./scripts/templates [-write]"
	templatesFolder = "internal"
	templateEnding  = ".templ"
)

// How templ's generator is told to check, where to look, and to say only what failed
const (
	checkFlag    = "-check"
	pathFlag     = "-path"
	logLevelFlag = "-log-level"
	failuresOnly = "warn"
)

// What the program says of a template out of its format
const unformatted = "%s is not formatted as templ formats it: run go generate ./internal/face/html\n"

func main() {
	os.Exit(run(os.Args[1:], "."))
}

// run checks or writes the templates under a root and their Go, as the arguments say.
func run(arguments []string, root string) int {
	writing, understood := modeOf(arguments)
	if !understood {
		fmt.Fprintln(os.Stderr, usage)
		return usageMistaken
	}
	folder := filepath.Join(root, templatesFolder)
	if writing {
		return write(folder)
	}
	return check(folder)
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

// templatesIn are the paths of every template in a folder and below, from the folder, in order.
func templatesIn(folder fs.FS) ([]string, error) {
	templates := []string{}
	err := fs.WalkDir(folder, ".", func(path string, entry fs.DirEntry, err error) error {
		return keepTemplate(&templates, path, entry, err)
	})
	if err != nil {
		return nil, fmt.Errorf("find the templates: %w", err)
	}
	slices.Sort(templates)
	return templates, nil
}

// keepTemplate adds a path the walk meets to the templates, where it is one.
func keepTemplate(templates *[]string, path string, entry fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if !entry.IsDir() && strings.HasSuffix(path, templateEnding) {
		*templates = append(*templates, path)
	}
	return nil
}

// write formats every template in the folder and below in place, then writes the Go of each.
func write(folder string) int {
	arguments := fmtcmd.Arguments{Files: []string{folder}}
	if err := fmtcmd.Run(logger(), nil, io.Discard, arguments); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return failed
	}
	writing := []string{pathFlag, folder, logLevelFlag, failuresOnly}
	if err := generatecmd.Run(context.Background(), io.Discard, os.Stderr, writing); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return failed
	}
	return succeeded
}

// check says each template in the folder out of its format, and fails where one is or where any Go is out of step
// with its template.
func check(folder string) int {
	templates := os.DirFS(folder)
	paths, err := templatesIn(templates)
	if err != nil {
		fmt.Fprintln(os.Stderr, folder, err)
		return failed
	}
	result := succeeded
	for _, path := range paths {
		whole := filepath.Join(folder, filepath.FromSlash(path))
		formatted, err := isFormatted(templates, path, whole)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return failed
		}
		if !formatted {
			fmt.Fprintf(os.Stderr, unformatted, whole)
			result = failed
		}
	}
	checking := []string{checkFlag, pathFlag, folder, logLevelFlag, failuresOnly}
	if err := generatecmd.Run(context.Background(), io.Discard, os.Stderr, checking); err != nil {
		fmt.Fprintln(os.Stderr, err)
		result = failed
	}
	return result
}

// isFormatted is whether a template of a folder, at a path in it and a whole path, reads as templ formats it,
// formatted in memory.
func isFormatted(folder fs.FS, path, whole string) (bool, error) {
	written, err := fs.ReadFile(folder, path)
	if err != nil {
		return false, fmt.Errorf("read %s: %w", whole, err)
	}
	var formatted bytes.Buffer
	arguments := fmtcmd.Arguments{StdinFilepath: whole}
	if err := fmtcmd.Run(logger(), bytes.NewReader(written), &formatted, arguments); err != nil {
		return false, fmt.Errorf("format %s: %w", whole, err)
	}
	return bytes.Equal(written, formatted.Bytes()), nil
}

// logger is where templ's formatter says what failed; what it did, it keeps to itself.
func logger() *slog.Logger {
	options := &slog.HandlerOptions{AddSource: false, Level: slog.LevelWarn, ReplaceAttr: nil}
	return slog.New(slog.NewTextHandler(os.Stderr, options))
}
