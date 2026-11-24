package content

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.yaml.in/yaml/v3"
)

// The folder of the posts, a folder for each named by its slug, and what a post's folder holds: its text in English,
// the original (D12), and in each language it is translated into, each named post.<language>.md (textFile); its
// revisions, older versions of its text in each; and its widgets, which come with a step of the blog of their own
// (docs/FOUNDATION.md §5.4)
const (
	postsFolder     = "posts"
	postFilePrefix  = "post."
	markdownEnding  = ".md"
	englishPostFile = postFilePrefix + english + markdownEnding
	revisionsFolder = "revisions"
	widgetsFolder   = "widgets"
)

// textFile is the name of a post's text in a language, by its tag, in the post's folder.
func textFile(language string) string {
	return postFilePrefix + language + markdownEnding
}

// The fields the checks name more than once: of a post's front matter, a post's and a series' title, and a post's and
// a revision's description
const (
	titleField       = "title"
	descriptionField = "description"
	idField          = "id"
	publishedField   = "published"
	updatedField     = "updated"
)

// Post is an article of the blog in one language: its slug, its folder's name; what its front matter says; its body.
type Post struct {
	Slug string `yaml:"-"`
	// Translation is the language its file's name gives a translation, as ru; none for the original, in English
	Translation string `yaml:"-"`
	// ID is the post's number, which blog/<number> reaches as its slug does: given once, never changed (D26)
	ID    int    `yaml:"id"`
	Title string `yaml:"title"`
	// Description is the post's lead, and its page's description
	Description string `yaml:"description"`
	Status      Status `yaml:"status"`
	// Published is when the post comes out, which a draft may leave out
	Published Moment `yaml:"published"`
	// Updated is the day the post was last changed, which a post may leave out
	Updated     Day          `yaml:"updated"`
	Tags        []string     `yaml:"tags"`
	Discussions []Discussion `yaml:"discussions"`
	Body        Markdown     `yaml:"-"`
	// Pictures are the pictures its body and its revisions show, files of the post's folder, by their names
	Pictures map[string]Picture `yaml:"-"`
	// SiteLinks are the body's links inside the site, by their nodes, which compose reads into pages and checks
	SiteLinks map[*ast.Link]WrittenLink `yaml:"-"`
	// Revisions are the post's older versions, oldest first
	Revisions []Revision `yaml:"-"`
}

// Status is who may read a post (D7): a draft, its author alone; a preview, whoever holds its secret link;
// published, everyone, wherever the site lists posts; unlisted, whoever has its address, listed nowhere.
type Status string

const (
	StatusDraft     Status = "draft"
	StatusPreview   Status = "preview"
	StatusPublished Status = "published"
	StatusUnlisted  Status = "unlisted"
)

// statuses are the statuses a post may have, each named once, as levels names a check's.
func statuses() map[Status]bool {
	return map[Status]bool{StatusDraft: true, StatusPreview: true, StatusPublished: true, StatusUnlisted: true}
}

// Discussion is a place a post is discussed, which stands in for comments the site has none of (D13): the site, by
// name, and the thread's address there.
type Discussion struct {
	Site    string     `yaml:"site"`
	Address WebAddress `yaml:"address"`
}

// fileLanguage is the language a text's file is named by: the one it is translated into, or English for the original.
//
//wrapper:names the language a file's name writes for a text, which the original's leaves unsaid as no translation
func fileLanguage(translation string) string {
	return cmp.Or(translation, english)
}

// translatedFields are a translation's front matter: its own text and dates; its number and tags are its original's.
type translatedFields struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Status      Status `yaml:"status"`
	Published   Moment `yaml:"published"`
	Updated     Day    `yaml:"updated"`
}

// writtenPost is a post's text in one language as its file writes it, the line its body starts on, and its revisions.
type writtenPost struct {
	file        string
	post        Post
	frontMatter *yaml.Node
	bodyLine    int
	revisions   []writtenRevision
}

// writtenFolder is a post's folder: its original, its translations, and the stray files of its revisions folder.
type writtenFolder struct {
	slug           string
	original       writtenPost
	translations   []writtenPost
	strayRevisions Faults
}

// readPosts reads every post's folder, in their slugs' order, with the faults that kept a post from being read.
//
// The faults of the files beside the posts' folders come apart: they keep no post from being read, so a load reports
// them with the faults of what the posts say.
func readPosts(folder fs.FS, languages []string) ([]writtenFolder, Faults, Faults) {
	entries, unreadable := readFolder(folder, postsFolder)
	postParser := parser.NewParser(postFormat()...)
	folders := make([]writtenFolder, 0, len(entries))
	misplaced := Faults{}
	for _, entry := range entries {
		if !entry.IsDir() {
			misplaced = append(misplaced, fileFault(path.Join(postsFolder, entry.Name()), strayPostProblem))
			continue
		}
		written, postFaults := readPostFolder(folder, postParser, entry.Name(), languages)
		folders = append(folders, written)
		unreadable = append(unreadable, postFaults...)
	}
	return folders, unreadable, misplaced
}

// readPostFolder reads a post's folder: its text in English and in each language it is translated into, with revisions.
func readPostFolder(folder fs.FS, postParser parser.Parser, slug string, languages []string) (writtenFolder, Faults) {
	revisions, faults, strayRevisions := readRevisions(folder, postParser, slug, languages)
	original, originalFaults := readText(folder, postParser, slug, "")
	original.revisions = revisions[english]
	written := writtenFolder{slug: slug, original: original, translations: nil, strayRevisions: strayRevisions}
	faults = append(faults, originalFaults...)
	for _, language := range languages {
		if _, err := fs.Stat(folder, path.Join(postsFolder, slug, textFile(language))); errors.Is(err, fs.ErrNotExist) {
			for _, revision := range revisions[language] {
				written.strayRevisions = append(written.strayRevisions, fileFault(revision.file, noTextProblem))
			}
			continue
		}
		translation, translationFaults := readText(folder, postParser, slug, language)
		translation.revisions = revisions[language]
		written.translations = append(written.translations, translation)
		faults = append(faults, translationFaults...)
	}
	return written, faults
}

// readText reads a post's text in a language, the original where it names none: front matter strictly, body once.
func readText(folder fs.FS, postParser parser.Parser, slug, translation string) (writtenPost, Faults) {
	// what a post read with faults is never inspected, since its faults end the load
	var unread writtenPost
	file := path.Join(postsFolder, slug, textFile(fileLanguage(translation)))
	data, err := fs.ReadFile(folder, file)
	if err != nil {
		return unread, Faults{unreadFault(file, err)}
	}
	parts, splitFaults := splitPost(file, data)
	if len(splitFaults) > 0 {
		return unread, splitFaults
	}
	post, frontMatter, decodingFaults := decodePost(file, parts.frontMatter, translation)
	post.Slug, post.Translation = slug, translation
	post.Body = Markdown{Source: parts.body, Tree: postParser.Parse(text.NewReader(parts.body))}
	return writtenPost{file: file, post: post, frontMatter: frontMatter, bodyLine: parts.bodyLine, revisions: nil}, decodingFaults
}

// decodePost decodes a post's front matter strictly: the original's whole, a translation's own fields alone.
func decodePost(file string, frontMatter []byte, translation string) (Post, *yaml.Node, Faults) {
	if translation == "" {
		return decodeStrictly[Post](file, frontMatter)
	}
	fields, document, faults := decodeStrictly[translatedFields](file, frontMatter)
	return Post{
		Slug: "", Translation: translation, ID: 0, Title: fields.Title, Description: fields.Description, Status: fields.Status,
		Published: fields.Published, Updated: fields.Updated, Tags: nil, Discussions: nil, Body: Markdown{Source: nil, Tree: nil},
		Pictures: nil, SiteLinks: nil, Revisions: nil,
	}, document, faults
}

// partFaults say why each part of a post's folder but its texts, revisions and pictures shown is not read.
func partFaults(folder fs.FS, written writtenFolder, shown map[string]bool) Faults {
	place := path.Dir(written.original.file)
	texts := map[string]bool{englishPostFile: true}
	for _, translation := range written.translations {
		texts[path.Base(translation.file)] = true
	}
	entries, faults := readFolder(folder, place)
	for _, entry := range entries {
		file := path.Join(place, entry.Name())
		switch name := entry.Name(); {
		case texts[name] || name == revisionsFolder || shown[name]:
			// the post itself, its translations and its revisions, which readPostFolder reads, and a picture one of its
			// versions shows, which its inspection reads
		case name == widgetsFolder:
			faults = append(faults, fileFault(file, widgetsProblem))
		case strings.HasPrefix(name, postFilePrefix) && strings.HasSuffix(name, markdownEnding):
			faults = append(faults, fileFault(file, noWordsProblem))
		default:
			faults = append(faults, fileFault(file, strayPartProblem))
		}
	}
	return faults
}

// inspectPosts checks every post and each translation of it, and that no two posts share a number: the originals, the
// translations by their language, and the faults found.
func inspectPosts(folder fs.FS, site WebAddress, folders []writtenFolder) ([]Post, map[string][]Post, Faults) {
	originals := make([]Post, 0, len(folders))
	translations := map[string][]Post{}
	faults := Faults{}
	numbered := map[int]bool{}
	for _, written := range folders {
		original := written.original
		folderInspection := &inspection{file: path.Dir(original.file), faults: nil}
		folderInspection.requireName(written.slug)
		originalInspection := &inspection{file: original.file, faults: nil}
		inspectPostFields(originalInspection, original.post)
		if numbered[original.post.ID] {
			originalInspection.add(idField, sharedNumberProblem, original.post.ID)
		}
		numbered[original.post.ID] = true
		post, shown, textFaults := inspectText(folder, site, originalInspection, original)
		faults = slices.Concat(faults, folderInspection.faults, textFaults)
		originals = append(originals, post)
		for _, translation := range written.translations {
			translated, translationShown, translationFaults := inspectTranslation(folder, site, post, translation)
			translations[translated.Translation] = append(translations[translated.Translation], translated)
			shown.add(translationShown)
			faults = append(faults, translationFaults...)
		}
		faults = slices.Concat(faults, written.strayRevisions, partFaults(folder, written, shown.named))
	}
	return originals, translations, faults
}

// inspectTranslation checks a translation as inspectText does, with its original's number, tags and discussions.
func inspectTranslation(folder fs.FS, site WebAddress, original Post, written writtenPost) (Post, shownPictures, Faults) {
	written.post.ID, written.post.Tags, written.post.Discussions = original.ID, original.Tags, original.Discussions
	translationInspection := &inspection{file: written.file, faults: nil}
	inspectTextFields(translationInspection, written.post)
	return inspectText(folder, site, translationInspection, written)
}

// inspectText checks a post's text in one language and its revisions, its fields inspected already.
func inspectText(folder fs.FS, site WebAddress, inspection *inspection, written writtenPost) (Post, shownPictures, Faults) {
	post := written.post
	place := path.Dir(written.file)
	shown, links := inspectBody(inspection, writtenBody{body: post.Body, firstLine: written.bodyLine, folder: folder, place: place, site: site})
	revisions, revisionsShown, revisionFaults := inspectRevisions(folder, site, post, written)
	shown.add(revisionsShown)
	post.Pictures, post.SiteLinks, post.Revisions = shown.read, links, revisions
	return post, shown, slices.Concat(inspection.faults, inspectCharacters(written.file, written.frontMatter), revisionFaults)
}

// inspectPostFields checks what an original's front matter says, field by field.
func inspectPostFields(inspection *inspection, post Post) {
	if post.ID <= 0 {
		inspection.add(idField, positiveProblem, post.ID)
	}
	inspectTextFields(inspection, post)
	inspectTags(inspection, post.Tags)
	for index, discussion := range post.Discussions {
		place := fmt.Sprintf("discussions[%d].", index)
		inspection.requireLine(place+"site", discussion.Site)
		inspection.requireSecureWebAddress(place+"address", discussion.Address)
	}
}

// inspectTextFields checks the title, lead, status and dates of a post's text in one language.
func inspectTextFields(inspection *inspection, post Post) {
	inspection.requireLine(titleField, post.Title)
	inspection.requireLine(descriptionField, post.Description)
	if !statuses()[post.Status] {
		inspection.add("status", noneOfProblem, post.Status, kindList(statuses()))
	}
	inspectDates(inspection, post)
}

// inspectDates holds a post to when it was published, which a draft may leave out, and to a day it was updated on
// no earlier than that.
func inspectDates(inspection *inspection, post Post) {
	switch {
	case post.Published.written == "" && post.Status == StatusDraft:
		// a draft not yet dated
	case post.Published.written == "":
		inspection.add(publishedField, emptyProblem)
	case !post.Published.known:
		inspection.add(publishedField, unknownMoment, post.Published.written)
	}
	switch {
	case post.Updated.written == "":
		// a post never updated
	case !post.Updated.known:
		inspection.add(updatedField, unknownDay, post.Updated.written)
	case post.Published.known && post.Updated.start.Before(post.Published.day()):
		inspection.add(updatedField, updatedProblem, post.Updated.written, post.Published.written)
	}
}

// inspectTags holds each tag to a slug's form, which an address filters the posts by, and to a tag of its own.
func inspectTags(inspection *inspection, tags []string) {
	tagForm := regexp.MustCompile(slugPattern)
	seen := map[string]bool{}
	for index, tag := range tags {
		field := fmt.Sprintf("tags[%d]", index)
		switch {
		case !tagForm.MatchString(tag):
			inspection.add(field, tagProblem, tag)
		case seen[tag]:
			inspection.add(field, sharedTagProblem, tag)
		}
		seen[tag] = true
	}
}
