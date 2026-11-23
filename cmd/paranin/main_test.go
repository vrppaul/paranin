package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/release"
)

// The exit code is what CI and the content's hook read: it is the command's contract.
func TestTheExitCodeSaysWhatTheCommandDid(test *testing.T) {
	faulty := test.TempDir()
	if err := os.WriteFile(filepath.Join(faulty, "profile.yaml"), []byte("person:\n  name: \"\"\n"), 0o600); err != nil {
		test.Fatalf("write a faulty profile: %v", err)
	}
	// content that loads, but whose résumé's experience takes the blog's address
	sharing := fixtures.Copy(test, fixtures.ContentFolder(test))
	fixtures.ReplaceIn(test, sharing, "i18n/en.yaml", "experience: Experience", "experience: Blog")
	// content that loads, but whose post links to a page the site does not serve
	leading := fixtures.Copy(test, fixtures.ContentFolder(test))
	fixtures.ReplaceIn(test, leading, "posts/reading-slowly/post.en.md", "A page kept", "[A page](/blog/no-such-post) kept")
	// content that loads, but whose draft links to a page the site does not serve
	drafting := fixtures.Copy(test, fixtures.ContentFolder(test))
	fixtures.ReplaceIn(test, drafting, "posts/notes-on-tools/post.en.md", "A draft,", "A [draft](/blog/no-such-post),")
	cases := []struct {
		name      string
		arguments []string
		want      int
	}{
		{name: "content that loads", arguments: []string{"paranin", "check", fixtures.ContentFolder(test)}, want: succeeded},
		{name: "content with faults", arguments: []string{"paranin", "check", faulty}, want: failed},
		{name: "content whose pages share an address", arguments: []string{"paranin", "check", sharing}, want: failed},
		{name: "a release whose pages share an address", arguments: []string{"paranin", "check", sharing, fixtures.BuildFolder(test)}, want: failed},
		{name: "content whose post links to no page", arguments: []string{"paranin", "check", leading}, want: failed},
		{name: "a release whose post links to no page", arguments: []string{"paranin", "check", leading, fixtures.BuildFolder(test)}, want: failed},
		{name: "content whose draft links to no page", arguments: []string{"paranin", "check", drafting}, want: failed},
		{name: "no folder named", arguments: []string{"paranin", "check"}, want: usageMistaken},
		{name: "a release that loads", arguments: []string{"paranin", "check", fixtures.ContentFolder(test), fixtures.BuildFolder(test)}, want: succeeded},
		{name: "a release whose build has faults", arguments: []string{"paranin", "check", fixtures.ContentFolder(test), faulty}, want: failed},
		{name: "a check of three folders", arguments: []string{"paranin", "check", fixtures.ContentFolder(test), fixtures.BuildFolder(test), faulty}, want: usageMistaken},
		{name: "a site on content with faults", arguments: []string{"paranin", "serve", faulty, fixtures.BuildFolder(test)}, want: failed},
		{name: "a site on a build with faults", arguments: []string{"paranin", "serve", fixtures.ContentFolder(test), faulty}, want: failed},
		{name: "a site with no build named", arguments: []string{"paranin", "serve", fixtures.ContentFolder(test)}, want: usageMistaken},
		{name: "a site with its drafts on content with faults", arguments: []string{"paranin", "serve", "--drafts", faulty, fixtures.BuildFolder(test)}, want: failed},
		{name: "a site with a flag it has not", arguments: []string{"paranin", "serve", "--everything", fixtures.ContentFolder(test), fixtures.BuildFolder(test)}, want: usageMistaken},
		{name: "a site with three folders", arguments: []string{"paranin", "serve", fixtures.ContentFolder(test), fixtures.BuildFolder(test), faulty}, want: usageMistaken},
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

// serve serves the drafts too only where its flag asks, before the content folder and the build folder.
func TestServeServesTheDraftsWhereAsked(test *testing.T) {
	folders := release.Folders{Content: "content", Build: "build/web"}
	cases := []struct {
		name  string
		words []string
		want  servingAsked
	}{
		{name: "the folders alone", words: []string{"content", "build/web"}, want: servingAsked{folders: folders, drafts: false}},
		{name: "the drafts asked", words: []string{"--drafts", "content", "build/web"}, want: servingAsked{folders: folders, drafts: true}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got, isWritten := servingOf(testCase.words); got != testCase.want || !isWritten {
				test.Errorf("servingOf(%q) = %+v, written right: %t; want %+v, written right", testCase.words, got, isWritten, testCase.want)
			}
		})
	}
}
