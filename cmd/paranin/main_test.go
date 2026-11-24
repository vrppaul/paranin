package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/hosting/web"
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
		// the fixture's translation was made from an older version of its original, which the check warns of alone
		{name: "content that loads, a translation older than its original", arguments: []string{"paranin", "check", fixtures.ContentFolder(test)}, want: succeeded},
		{name: "content with faults", arguments: []string{"paranin", "check", faulty}, want: failed},
		{name: "content whose pages share an address", arguments: []string{"paranin", "check", sharing}, want: failed},
		{name: "a release whose pages share an address", arguments: []string{"paranin", "check", sharing, fixtures.BuildFolder(test)}, want: failed},
		{name: "content whose post links to no page", arguments: []string{"paranin", "check", leading}, want: failed},
		{name: "a release whose post links to no page", arguments: []string{"paranin", "check", leading, fixtures.BuildFolder(test)}, want: failed},
		{name: "content whose draft links to no page", arguments: []string{"paranin", "check", drafting}, want: failed},
		{name: "no folder named", arguments: []string{"paranin", "check"}, want: usageMistaken},
		{name: "a release that loads, a translation older than its original", arguments: []string{"paranin", "check", fixtures.ContentFolder(test), fixtures.BuildFolder(test)}, want: succeeded},
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
			return checkRelease(fixtures.ContentFolder(test), fixtures.BuildFolder(test), warned)
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

// The secret the tests sign preview links with, and how long a test's link lasts or has lapsed by
const (
	previewSecret = "a secret the tests sign preview links with"
	linkLasts     = time.Hour
)

// preview-link prints the link of a post previewed now, signed with the key in the file the environment names, for as
// long as it is asked, ninety days at most; it fails for a post out, a draft or no post, whose link would open nothing,
// and where the environment names no key or one too short; written wrong, a lifetime of none or past ninety days among
// it, it says how it is written.
func TestAPreviewLinkIsPrintedForAPreviewAlone(test *testing.T) {
	folder := fixtures.PreviewedFolder(test)
	key, short := fixtures.KeyFile(test, previewSecret), fixtures.KeyFile(test, "short")
	cases := []struct {
		name    string
		keyPath string
		words   []string
		want    int
	}{
		{name: "a preview", keyPath: key, words: []string{folder, fixtures.PreviewSlug}, want: succeeded},
		{name: "a preview for an hour", keyPath: key, words: []string{"--lasts", "1h", folder, fixtures.PreviewSlug}, want: succeeded},
		{name: "a post out", keyPath: key, words: []string{folder, "reading-slowly"}, want: failed},
		{name: "a draft", keyPath: key, words: []string{folder, "notes-on-tools"}, want: failed},
		{name: "no post", keyPath: key, words: []string{folder, "no-such-post"}, want: failed},
		{name: "no key named", keyPath: "", words: []string{folder, fixtures.PreviewSlug}, want: failed},
		{name: "a key too short", keyPath: short, words: []string{folder, fixtures.PreviewSlug}, want: failed},
		{name: "no slug", keyPath: key, words: []string{folder}, want: usageMistaken},
		{name: "two slugs", keyPath: key, words: []string{folder, fixtures.PreviewSlug, fixtures.PreviewSlug}, want: usageMistaken},
		{name: "a lifetime of none", keyPath: key, words: []string{"--lasts", "0s", folder, fixtures.PreviewSlug}, want: usageMistaken},
		{name: "the longest lifetime", keyPath: key, words: []string{"--lasts", "2160h", folder, fixtures.PreviewSlug}, want: succeeded},
		{name: "a lifetime past the longest", keyPath: key, words: []string{"--lasts", "2161h", folder, fixtures.PreviewSlug}, want: usageMistaken},
		{name: "a lifetime that does not read", keyPath: key, words: []string{"--lasts", "soon", folder, fixtures.PreviewSlug}, want: usageMistaken},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			test.Setenv(previewKeyVariable, testCase.keyPath)
			arguments := append([]string{"paranin", "preview-link"}, testCase.words...)
			if got := run(arguments); got != testCase.want {
				test.Errorf("run(%q) with the key %q = %d, want %d", arguments, testCase.keyPath, got, testCase.want)
			}
		})
	}
}

// preview-link's link lasts thirty days where no lifetime is asked, and as long as one asked; it opens the post in
// the language asked, the original where none is.
func TestAPreviewLinkLastsAsAsked(test *testing.T) {
	coming := document.Page{Kind: document.PagePost, Slug: "coming", Tab: ""}
	cases := []struct {
		name  string
		words []string
		want  linkAsked
	}{
		{name: "no lifetime asked", words: []string{"content", "coming"}, want: linkAsked{post: postAsked{folder: "content", page: coming}, lasts: 30 * 24 * time.Hour}},
		{name: "an hour asked", words: []string{"--lasts", "1h", "content", "coming"}, want: linkAsked{post: postAsked{folder: "content", page: coming}, lasts: time.Hour}},
		{
			name: "a translation's", words: []string{"--language", "ru", "content", "coming"},
			want: linkAsked{post: postAsked{folder: "content", page: coming.InLanguage("ru")}, lasts: 30 * 24 * time.Hour},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got, isWritten := linkAskedOf(testCase.words); got != testCase.want || !isWritten {
				test.Errorf("linkAskedOf(%q) = %+v, written right: %t; want %+v, written right", testCase.words, got, isWritten, testCase.want)
			}
		})
	}
}

// The link preview-link prints opens the preview on a server holding the same key until it lapses, and not after.
func TestAPrintedLinkOpensThePreviewUntilTheLinkLapses(test *testing.T) {
	folder := fixtures.PreviewedFolder(test)
	key, err := web.ReadPreviewKey(fixtures.KeyFile(test, previewSecret))
	if err != nil {
		test.Fatalf("ReadPreviewKey = %v, want a key", err)
	}
	store, err := release.Open(release.Folders{Content: folder, Build: fixtures.BuildFolder(test)}, checkedNow)
	if err != nil {
		test.Fatalf("Open(fixtures) = %v, want no error", err)
	}
	// the server draws its snapshot before the bubble a link is opened in (fixtures.ServedOnTheDay)
	server := web.New(store, release.DeriveFromContent(store, compose.Prepare), false, key)
	server.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	day := fixtures.Day()
	previewed := document.Page{Kind: document.PagePost, Slug: fixtures.PreviewSlug, Tab: ""}
	cases := []struct {
		name   string
		post   document.Page
		lapses time.Time
		want   int
	}{
		{name: "before it lapses", post: previewed, lapses: day.Add(linkLasts), want: http.StatusOK},
		{name: "once it has lapsed", post: previewed, lapses: day.Add(-linkLasts), want: http.StatusNotFound},
		{name: "a translation's, before it lapses", post: previewed.InLanguage(fixtures.PreviewLanguage), lapses: day.Add(linkLasts), want: http.StatusOK},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			link, err := previewLinkOf(folder, testCase.post, key, day, testCase.lapses)
			if err != nil {
				test.Fatalf("previewLinkOf(%+v) = %v, want a link", testCase.post, err)
			}
			address, err := url.Parse(link)
			if err != nil {
				test.Fatalf("the link %q does not parse: %v", link, err)
			}
			response := fixtures.ServedOnTheDay(test, server, http.MethodGet, address.Path+".md", "")
			if response.Code != testCase.want {
				test.Errorf("GET %s.md = %d, want %d", address.Path, response.Code, testCase.want)
			}
		})
	}
}

// The fixture's post out, the revision revise takes of it, dated by its update of the 12th, and a moment late on that
// day in the post's zone
const (
	postOut          = "reading-slowly"
	revisionTaken    = "posts/reading-slowly/revisions/2026-11-12.en.md"
	onTheVersionsDay = "2026-11-12T22:00+01:00"
)

// revise takes a revision of a post out, its current version, beside the post; it refuses a post not out, a draft, a
// preview or none at all, whose version no reader read, and a version on the day it came out, a change of which is an
// edit of it, the one a revision taken that day dates among them. Each case runs on a folder of its own.
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
		{name: "a translation out", post: out.InLanguage(fixtures.PreviewLanguage), before: fixtures.KeepFolder, moment: fixtures.Day(), written: true},
		{name: "a draft", post: document.PostPage("notes-on-tools"), before: fixtures.KeepFolder, moment: fixtures.Day(), written: false},
		{name: "a preview", post: document.PostPage(fixtures.PreviewSlug), before: fixtures.KeepFolder, moment: fixtures.Day(), written: false},
		{name: "a translation previewed", post: document.PostPage(fixtures.PreviewSlug).InLanguage(fixtures.PreviewLanguage), before: fixtures.KeepFolder, moment: fixtures.Day(), written: false},
		{name: "no post", post: document.PostPage("no-such-post"), before: fixtures.KeepFolder, moment: fixtures.Day(), written: false},
		{name: "a post in a language it has no translation in", post: document.PostPage("notes-on-tools").InLanguage(fixtures.PreviewLanguage), before: fixtures.KeepFolder, moment: fixtures.Day(), written: false},
		{name: "its version taken that day", post: out, before: revisedOnTheDay, moment: fixtures.Day(), written: false},
		{name: "on its version's day", post: out, before: fixtures.KeepFolder, moment: fixtures.MomentOf(test, onTheVersionsDay), written: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			folder := fixtures.PreviewedFolder(test)
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
		{name: "a flag revise takes not", words: []string{"--lasts", "1h", folder, postOut}},
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
