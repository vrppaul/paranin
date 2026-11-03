package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vrppaul/paranin/internal/fixtures"
)

// The exit code is what CI and the content's hook read: it is the command's contract.
func TestTheExitCodeSaysWhatTheCommandDid(test *testing.T) {
	faulty := test.TempDir()
	if err := os.WriteFile(filepath.Join(faulty, "profile.yaml"), []byte("person:\n  name: \"\"\n"), 0o600); err != nil {
		test.Fatalf("write a faulty profile: %v", err)
	}
	cases := []struct {
		name      string
		arguments []string
		want      int
	}{
		{name: "content that loads", arguments: []string{"paranin", "check", fixtures.ContentFolder(test)}, want: succeeded},
		{name: "content with faults", arguments: []string{"paranin", "check", faulty}, want: failed},
		{name: "no folder named", arguments: []string{"paranin", "check"}, want: usageMistaken},
		{name: "a release that loads", arguments: []string{"paranin", "check", fixtures.ContentFolder(test), fixtures.BuildFolder(test)}, want: succeeded},
		{name: "a release whose build has faults", arguments: []string{"paranin", "check", fixtures.ContentFolder(test), faulty}, want: failed},
		{name: "a check of three folders", arguments: []string{"paranin", "check", fixtures.ContentFolder(test), fixtures.BuildFolder(test), faulty}, want: usageMistaken},
		{name: "a site on content with faults", arguments: []string{"paranin", "serve", faulty, fixtures.BuildFolder(test)}, want: failed},
		{name: "a site on a build with faults", arguments: []string{"paranin", "serve", fixtures.ContentFolder(test), faulty}, want: failed},
		{name: "a site with no build named", arguments: []string{"paranin", "serve", fixtures.ContentFolder(test)}, want: usageMistaken},
		{name: "no command", arguments: []string{"paranin"}, want: usageMistaken},
		{name: "a command it has not", arguments: []string{"paranin", "render", faulty}, want: usageMistaken},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := run(testCase.arguments); got != testCase.want {
				test.Errorf("run(%q) = %d, want %d", testCase.arguments, got, testCase.want)
			}
		})
	}
}
