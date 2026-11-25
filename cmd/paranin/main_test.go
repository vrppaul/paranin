package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
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
	// where a site would listen were it to start, so a site that should not start and does waits on no real port
	startedAside := []environmentVariable{
		{name: webAddressVariable, value: anyPort},
		{name: sshAddressVariable, value: anyPort},
		{name: hostKeyVariable, value: filepath.Join(test.TempDir(), "host_key")},
	}
	contentFolder, buildFolder, cardsFolder := fixtures.ContentFolder(test), fixtures.BuildFolder(test), fixtures.CardsFolder(test)
	cases := []struct {
		name        string
		arguments   []string
		environment []environmentVariable
		want        int
	}{
		{
			name: "a site with an IndexNow key of another form", arguments: []string{"paranin", "serve", contentFolder, buildFolder, cardsFolder},
			environment: append([]environmentVariable{{name: indexNowKeyVariable, value: "key_of_another_form"}}, startedAside...), want: failed,
		},
		// the fixture's translation was made from an older version of its original, which the check warns of alone
		{name: "content that loads, a translation older than its original", arguments: []string{"paranin", "check", contentFolder}, want: succeeded},
		{name: "content with faults", arguments: []string{"paranin", "check", faulty}, want: failed},
		{name: "content whose pages share an address", arguments: []string{"paranin", "check", sharing}, want: failed},
		{name: "a release whose pages share an address", arguments: []string{"paranin", "check", sharing, buildFolder, cardsFolder}, want: failed},
		{name: "content whose post links to no page", arguments: []string{"paranin", "check", leading}, want: failed},
		{name: "a release whose post links to no page", arguments: []string{"paranin", "check", leading, buildFolder, cardsFolder}, want: failed},
		{name: "content whose draft links to no page", arguments: []string{"paranin", "check", drafting}, want: failed},
		{name: "no folder named", arguments: []string{"paranin", "check"}, want: usageMistaken},
		{name: "a release that loads, a translation older than its original", arguments: []string{"paranin", "check", contentFolder, buildFolder, cardsFolder}, want: succeeded},
		{name: "a release whose build has faults", arguments: []string{"paranin", "check", contentFolder, faulty, cardsFolder}, want: failed},
		{name: "a release whose cards have faults", arguments: []string{"paranin", "check", contentFolder, buildFolder, faulty}, want: failed},
		{name: "a check of the content and a build alone", arguments: []string{"paranin", "check", contentFolder, buildFolder}, want: usageMistaken},
		{name: "a check of four folders", arguments: []string{"paranin", "check", contentFolder, buildFolder, cardsFolder, faulty}, want: usageMistaken},
		{name: "a site on content with faults", arguments: []string{"paranin", "serve", faulty, buildFolder, cardsFolder}, want: failed},
		{name: "a site on a build with faults", arguments: []string{"paranin", "serve", contentFolder, faulty, cardsFolder}, want: failed},
		{name: "a site on cards with faults", arguments: []string{"paranin", "serve", contentFolder, buildFolder, faulty}, want: failed},
		{name: "a site with no cards named", arguments: []string{"paranin", "serve", contentFolder, buildFolder}, want: usageMistaken},
		{name: "a site with its drafts on content with faults", arguments: []string{"paranin", "serve", "--drafts", faulty, buildFolder, cardsFolder}, want: failed},
		{name: "a site with a flag it has not", arguments: []string{"paranin", "serve", "--everything", contentFolder, buildFolder, cardsFolder}, want: usageMistaken},
		{name: "a site with four folders", arguments: []string{"paranin", "serve", contentFolder, buildFolder, cardsFolder, faulty}, want: usageMistaken},
		// go test runs the command in its package's folder, where no faces the cards are set in stand
		{name: "cards drawn outside the repository's root", arguments: []string{"paranin", "cards", contentFolder, test.TempDir()}, want: failed},
		{name: "cards with no folder to draw them into", arguments: []string{"paranin", "cards", contentFolder}, want: usageMistaken},
		{name: "cards with a flag they have not", arguments: []string{"paranin", "cards", "--drafts", contentFolder, test.TempDir()}, want: usageMistaken},
		{name: "no command", arguments: []string{"paranin"}, want: usageMistaken},
		{name: "a command it has not", arguments: []string{"paranin", "render", faulty}, want: usageMistaken},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			for _, variable := range testCase.environment {
				test.Setenv(variable.name, variable.value)
			}
			if got := run(testCase.arguments); got != testCase.want {
				test.Errorf("run(%q) with %+v = %d, want %d", testCase.arguments, testCase.environment, got, testCase.want)
			}
		})
	}
}

// fixtureRelease are the folders of a release of the content in a folder, the build's fixture and the cards'.
func fixtureRelease(test *testing.T, contentFolder string) release.Folders {
	return release.Folders{Content: contentFolder, Build: fixtures.BuildFolder(test), Cards: fixtures.CardsFolder(test)}
}

// environmentVariable is a variable of the environment a test sets for the command it runs.
type environmentVariable struct {
	name  string
	value string
}

// An address the system picks a free port at
const anyPort = "localhost:0"

// A site takes the IndexNow key the environment sets, of IndexNow's form, and none where it sets none or the site
// serves its drafts, a key of another form among them.
func TestServeTakesTheIndexNowKeyOfItsForm(test *testing.T) {
	cases := []struct {
		name      string
		key       string
		drafts    bool
		want      string
		isRefused bool
	}{
		{name: "none", key: "", drafts: false, want: "", isRefused: false},
		{name: "a key", key: "a1b2c3d4-key", drafts: false, want: "a1b2c3d4-key", isRefused: false},
		{name: "a key of another form", key: "key_of_another_form", drafts: false, want: "", isRefused: true},
		{name: "a key, with the drafts", key: "a1b2c3d4-key", drafts: true, want: "", isRefused: false},
		{name: "a key of another form, with the drafts", key: "key_of_another_form", drafts: true, want: "", isRefused: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			test.Setenv(indexNowKeyVariable, testCase.key)
			if got, err := indexNowKeyOf(testCase.drafts); got != testCase.want || (err != nil) != testCase.isRefused {
				test.Errorf("indexNowKeyOf(%t) with the key %q = %q, %v; want %q, refused: %t", testCase.drafts, testCase.key, got, err,
					testCase.want, testCase.isRefused)
			}
		})
	}
}

// Where check warns the fixture's translation, made from an older version of its original, stands: its file and field
const staleTranslationPlace = "posts/reading-slowly/post.ru.md: source "

// check of the content alone and of a release prints what the content warns of, at its place, and passes.
func TestACheckPrintsWhatTheContentWarnsOfAndPasses(test *testing.T) {
	cases := []struct {
		name  string
		check func(warned io.Writer) int
	}{
		{name: "the content alone", check: func(warned io.Writer) int { return check(fixtures.ContentFolder(test), warned) }},
		{name: "a release", check: func(warned io.Writer) int {
			return checkRelease(fixtureRelease(test, fixtures.ContentFolder(test)), warned)
		}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			warned := &strings.Builder{}
			if code := testCase.check(warned); code != succeeded || !strings.Contains(warned.String(), staleTranslationPlace) {
				test.Errorf("check of %s = %d, printing %q; want %d, warning at %q", testCase.name, code, warned.String(), succeeded,
					staleTranslationPlace)
			}
		})
	}
}

// The fixture's post out, the revision revise takes of it, dated by its update of the 12th, and a moment late on that
// day in the post's zone; and the language of the fixture's translation, by its tag
const (
	postOut             = "reading-slowly"
	revisionTaken       = "posts/reading-slowly/revisions/2026-11-12.en.md"
	onTheVersionsDay    = "2026-11-12T22:00+01:00"
	translationLanguage = "ru"
)

// The fixture's draft, its file, and its status as written and as a case dates it, before the tests' day
const (
	draftSlug      = "notes-on-tools"
	draftFile      = "posts/notes-on-tools/post.en.md"
	draftStatus    = "status: draft"
	draftDatedText = "status: draft\npublished: 2026-11-01T09:00+01:00"
)

// revise takes a revision of a post out, its current version, beside the post; it refuses a post not out, a draft or
// none at all, whose version no reader read, and a version on the day it came out, a change of which is an edit of it,
// the one a revision taken that day dates among them. Each case runs on a folder of its own.
func TestARevisionIsTakenOfAPostOutAlone(test *testing.T) {
	out := document.PostPage(postOut)
	cases := []struct {
		name    string
		post    document.Page
		before  func(test *testing.T, folder string)
		moment  time.Time
		written bool
	}{
		{name: "a post out", post: out, before: fixtures.KeepFolder, moment: fixtures.Day(), written: true},
		{name: "a translation out", post: out.InLanguage(translationLanguage), before: fixtures.KeepFolder, moment: fixtures.Day(), written: true},
		{name: "a draft", post: document.PostPage(draftSlug), before: fixtures.KeepFolder, moment: fixtures.Day(), written: false},
		{name: "a draft dated before the moment", post: document.PostPage(draftSlug), before: func(test *testing.T, folder string) {
			fixtures.ReplaceIn(test, folder, draftFile, draftStatus, draftDatedText)
		}, moment: fixtures.Day(), written: false},
		{name: "no post", post: document.PostPage("no-such-post"), before: fixtures.KeepFolder, moment: fixtures.Day(), written: false},
		{name: "a post in a language it has no translation in", post: document.PostPage(draftSlug).InLanguage(translationLanguage), before: fixtures.KeepFolder, moment: fixtures.Day(), written: false},
		{name: "its version taken that day", post: out, before: revisedOnTheDay, moment: fixtures.Day(), written: false},
		{name: "on its version's day", post: out, before: fixtures.KeepFolder, moment: fixtures.MomentOf(test, onTheVersionsDay), written: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			folder := fixtures.Copy(test, fixtures.ContentFolder(test))
			testCase.before(test, folder)
			if _, err := reviseAt(folder, testCase.post, testCase.moment); (err == nil) != testCase.written {
				test.Errorf("reviseAt(%+v) at %s = %v, want a revision written: %t", testCase.post, testCase.moment, err, testCase.written)
			}
		})
	}
}

// A folder revised loads again at once: revise dates the post as updated on the day it is revised.
func TestAFolderRevisedLoadsAtOnce(test *testing.T) {
	folder := fixtures.Copy(test, fixtures.ContentFolder(test))
	revisedOnTheDay(test, folder)
	if code := check(folder, io.Discard); code != succeeded {
		test.Errorf("check of the folder with %s = %d, want %d", revisionTaken, code, succeeded)
	}
}

// A revision's file is never written over: one there already is refused and kept as it was, not taken away, and the
// post is left as it was, not dated.
func TestARevisionThereAlreadyIsKept(test *testing.T) {
	const kept = "a revision there already"
	folder := fixtures.Copy(test, fixtures.ContentFolder(test))
	fixtures.WriteIn(test, folder, revisionTaken, kept)
	post, found := fixtures.LoadedContent(test).Post("", postOut)
	postFile := filepath.Join(folder, content.PostFile(post))
	before, err := os.ReadFile(postFile)
	if !found || err != nil {
		test.Fatalf("read the post %q of the fixture: found %t, %v", postOut, found, err)
	}
	if err := saveRevision(folder, post, content.RevisionFile{Path: revisionTaken, Text: []byte("another")}, fixtures.Day()); err == nil {
		test.Errorf("saveRevision over %s = nil, want a fault", revisionTaken)
	}
	if written, err := os.ReadFile(filepath.Join(folder, revisionTaken)); err != nil || string(written) != kept {
		test.Errorf("%s holds %q, %v; want %q", revisionTaken, written, err, kept)
	}
	if after, err := os.ReadFile(postFile); err != nil || !bytes.Equal(after, before) {
		test.Errorf("%s holds %q, %v; want it as it was, %q", postFile, after, err, before)
	}
}

// revisedOnTheDay takes a revision of the fixture's post out on the tests' day.
func revisedOnTheDay(test *testing.T, folder string) {
	test.Helper()
	if written, err := reviseAt(folder, document.PostPage(postOut), fixtures.Day()); err != nil || written != revisionTaken {
		test.Fatalf("reviseAt(%q) = %q, %v; want %q", postOut, written, err, revisionTaken)
	}
}

// revise takes the post in the language asked, the original where none is.
func TestReviseTakesThePostInTheLanguageAsked(test *testing.T) {
	read := document.PostPage(postOut)
	cases := []struct {
		name  string
		words []string
		want  postAsked
	}{
		{name: "the original", words: []string{"content", postOut}, want: postAsked{folder: "content", page: read}},
		{name: "a translation", words: []string{"--language", translationLanguage, "content", postOut}, want: postAsked{folder: "content", page: read.InLanguage(translationLanguage)}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got, isWritten := postAskedOf(testCase.words); got != testCase.want || !isWritten {
				test.Errorf("postAskedOf(%q) = %+v, written right: %t; want %+v, written right", testCase.words, got, isWritten, testCase.want)
			}
		})
	}
}

// revise is written with the content folder and the post's slug, and nothing else but the language of a translation.
func TestReviseTakesAFolderAndASlugAlone(test *testing.T) {
	folder := fixtures.Copy(test, fixtures.ContentFolder(test))
	cases := []struct {
		name  string
		words []string
	}{
		{name: "no slug", words: []string{folder}},
		{name: "two slugs", words: []string{folder, postOut, postOut}},
		{name: "a language with no tag", words: []string{folder, postOut, "--language"}},
		{name: "a flag revise takes not", words: []string{"--drafts", folder, postOut}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			arguments := append([]string{"paranin", "revise"}, testCase.words...)
			if got := run(arguments); got != usageMistaken {
				test.Errorf("run(%q) = %d, want %d", arguments, got, usageMistaken)
			}
		})
	}
}

// serve serves the drafts too only where its flag asks, before the content folder, the build folder and the cards'.
func TestServeServesTheDraftsWhereAsked(test *testing.T) {
	folders := release.Folders{Content: "content", Build: "build/web", Cards: "build/cards"}
	cases := []struct {
		name  string
		words []string
		want  servingAsked
	}{
		{name: "the folders alone", words: []string{"content", "build/web", "build/cards"}, want: servingAsked{folders: folders, drafts: false}},
		{name: "the drafts asked", words: []string{"--drafts", "content", "build/web", "build/cards"}, want: servingAsked{folders: folders, drafts: true}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got, isWritten := servingOf(testCase.words); got != testCase.want || !isWritten {
				test.Errorf("servingOf(%q) = %+v, written right: %t; want %+v, written right", testCase.words, got, isWritten, testCase.want)
			}
		})
	}
}
