// Command typst installs Typst at the version the cards are drawn with (internal/face/card's TypstVersion, ADR-0005),
// in development and in CI alike: it downloads that version's release for this machine, refuses it unless its sha256
// is the one written here, and unpacks the program into the folder named, which the path must hold, since the tests
// and the build find typst there. Unpacking runs tar, since the release is packed with xz, which Go's library lacks.
//
// Run it in the repository's root, as CI does:
//
//	go run ./scripts/typst ~/.local/bin
//
// It prints the path of the program it installed, and exits 0 when it is installed, 1 when the download, its check or
// the unpacking failed or the whole took longer than five minutes, and 2 when the command is written wrong or no
// release is pinned for this machine.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/vrppaul/paranin/internal/face/card"
)

// What the program exits with
const (
	succeeded     = 0
	failed        = 1
	usageMistaken = 2
)

// How the command is written, what it says on a machine no release is pinned for, how a machine is named, as Go names
// its system and its processor with a slash between, where a release is downloaded from, by its version and its name,
// and how long the install may take, its download among it
const (
	usage          = "usage: go run ./scripts/typst <folder>"
	noRelease      = "no release of Typst is pinned for %s\n"
	platformJoint  = "/"
	releaseURL     = "https://github.com/typst/typst/releases/download/v%s/%s.tar.xz"
	installTimeout = 5 * time.Minute
)

// How tar unpacks the program from the release read on its standard input to its standard output: the release's folder
// holds it under the program's name, a slash between them; and how the program is installed, for anyone to run
const (
	unpacker      = "tar"
	extractOption = "--extract"
	unpackOption  = "--xz"
	toOutput      = "--to-stdout"
	fileOption    = "--file"
	standardInput = "-"
	memberJoint   = "/"
	programName   = "typst"
	programMode   = 0o755
)

// The faults of an install
var errSum = errors.New("the download's sha256 is not the one pinned")

// release is Typst's release for one machine: its name, and the sha256 of its archive at TypstVersion.
type release struct {
	name   string
	sha256 string
}

// releases are the releases pinned, by the machine each runs on, as Go names it: the owner's and CI's.
func releases() map[string]release {
	return map[string]release{
		"linux/amd64": {name: "typst-x86_64-unknown-linux-musl", sha256: "a6d077d0a95eed5a2eba715b2dae06be954f624ccbf85758a03f389ded33118c"},
	}
}

func main() {
	ctx, stop := context.WithTimeout(context.Background(), installTimeout)
	code := run(ctx, os.Args[1:], runtime.GOOS+platformJoint+runtime.GOARCH, download)
	stop()
	os.Exit(code)
}

// run installs the release of a machine into the folder the arguments name, from what fetch downloads.
func run(ctx context.Context, arguments []string, platform string, fetch func(context.Context, string) ([]byte, error)) int {
	if len(arguments) != 1 {
		fmt.Fprintln(os.Stderr, usage)
		return usageMistaken
	}
	pinned, known := releases()[platform]
	if !known {
		fmt.Fprintf(os.Stderr, noRelease, platform)
		return usageMistaken
	}
	installed, err := install(ctx, pinned, arguments[0], fetch)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return failed
	}
	fmt.Println(installed)
	return succeeded
}

// install downloads a release, checks its sha256 and unpacks its program into a folder, and is the program's path.
func install(ctx context.Context, pinned release, folder string, fetch func(context.Context, string) ([]byte, error)) (string, error) {
	address := fmt.Sprintf(releaseURL, card.TypstVersion, pinned.name)
	archive, err := fetch(ctx, address)
	if err != nil {
		return "", err
	}
	if sum := sha256.Sum256(archive); hex.EncodeToString(sum[:]) != pinned.sha256 {
		return "", fmt.Errorf("%w: %s is %x, want %s", errSum, address, sum, pinned.sha256)
	}
	var program, said bytes.Buffer
	member := pinned.name + memberJoint + programName
	unpack := exec.CommandContext(ctx, unpacker, extractOption, unpackOption, toOutput, fileOption, standardInput, member) //nolint:gosec // a release's name of the table above
	unpack.Stdin, unpack.Stdout, unpack.Stderr = bytes.NewReader(archive), &program, &said
	if err := unpack.Run(); err != nil {
		return "", fmt.Errorf("unpack %s: %w: %s", address, err, said.String())
	}
	installed := filepath.Join(folder, programName)
	if err := os.WriteFile(installed, program.Bytes(), programMode); err != nil { //nolint:gosec // a program, which must run
		return "", fmt.Errorf("install %s: %w", installed, err)
	}
	return installed, nil
}

// download is what an address answers.
func download(ctx context.Context, address string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, fmt.Errorf("ask for %s: %w", address, err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", address, err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: %s", address, response.Status)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", address, err)
	}
	return body, nil
}
