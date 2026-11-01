package main

import (
	"os"
	"path/filepath"
	"testing"
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
		{name: "content that loads", arguments: []string{"paranin", "check", "../../internal/content/testdata/content"}, want: succeeded},
		{name: "content with faults", arguments: []string{"paranin", "check", faulty}, want: failed},
		{name: "no folder named", arguments: []string{"paranin", "check"}, want: usageMistaken},
		{name: "a site on content with faults", arguments: []string{"paranin", "serve", faulty}, want: failed},
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
