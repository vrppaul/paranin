package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// A download whose sha256 is not the one pinned is refused, and nothing is installed.
func TestADownloadOfAnotherSumIsRefused(test *testing.T) {
	folder := test.TempDir()
	pinned := releases()["linux/amd64"]
	_, err := install(test.Context(), pinned, folder, func(context.Context, string) ([]byte, error) { return []byte("another release"), nil })
	if !errors.Is(err, errSum) {
		test.Errorf("install of another release = %v, want %v", err, errSum)
	}
	if _, err := os.Stat(filepath.Join(folder, programName)); !errors.Is(err, os.ErrNotExist) {
		test.Errorf("stat %s = %v, want nothing installed", programName, err)
	}
}
