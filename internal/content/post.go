package content

import (
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.yaml.in/yaml/v3"
)

// The folder of the posts, a folder for each named by its slug, and what a post's folder holds: its text in English,
// the one language so far (D12), named post.<language>.md as every language's file will be; and two folders that come
// with steps of the blog of their own, its revisions and its widgets (docs/FOUNDATION.md §5.4)
const (
	postsFolder     = "posts"
	postFilePrefix  = "post."
	markdownEnding  = ".md"
	englishPostFile = postFilePrefix + english + markdownEnding
	revisionsFolder = "revisions"
	widgetsFolder   = "widgets"
)

// The fields the checks name more than once: of a post's front matter, and a post's and a series' title
const (
	titleField     = "title"
	idField        = "id"
	publishedField = "published"
	updatedField   = "updated"
)

// Post is an article of the blog: its slug, its folder's name; what its front matter says; and its body.
type Post struct {
	Slug string `yaml:"-"`
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
	// Pictures are the pictures the body shows, files of the post's folder, by their names
	Pictures map[string]Picture `yaml:"-"`
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

// writtenPost is a post as its folder writes it: the file, the post, the front matter's document, which the check of
// characters walks, and the file's line the body starts on, which a fault of the body counts from.
type writtenPost struct {
	file        string
	post        Post
	frontMatter *yaml.Node
	bodyLine    int
}

// readPosts reads every post's folder, in their slugs' order, with the faults that kept a post from being read.
//
// The faults of the files beside the posts' folders come apart: they keep no post from being read, so a load reports
// them with the faults of what the posts say.
func readPosts(folder fs.FS) ([]writtenPost, Faults, Faults) {
	entries, unreadable := readFolder(folder, postsFolder)
	postParser := parser.NewParser(postFormat()...)
	posts := make([]writtenPost, 0, len(entries))
	misplaced := Faults{}
	for _, entry := range entries {
		if !entry.IsDir() {
			misplaced = append(misplaced, fileFault(path.Join(postsFolder, entry.Name()), strayPostProblem))
			continue
		}
		post, postFaults := readPost(folder, postParser, entry.Name())
		posts = append(posts, post)
		unreadable = append(unreadable, postFaults...)
	}
	return posts, unreadable, misplaced
}

// readPost reads the file of the post of a slug: its front matter decoded strictly and its body parsed once by the
// posts' parser.
func readPost(folder fs.FS, postParser parser.Parser, slug string) (writtenPost, Faults) {
	file := path.Join(postsFolder, slug, englishPostFile)
	// what a post read with faults is never inspected, since its faults end the load
	var unread writtenPost
	data, err := fs.ReadFile(folder, file)
	if err != nil {
		return unread, Faults{unreadFault(file, err)}
	}
	parts, splitFaults := splitPost(file, data)
	if len(splitFaults) > 0 {
		return unread, splitFaults
	}
	post, frontMatter, decodingFaults := decodeStrictly[Post](file, parts.frontMatter)
	post.Slug = slug
	post.Body = Markdown{Source: parts.body, Tree: postParser.Parse(text.NewReader(parts.body))}
	return writtenPost{file: file, post: post, frontMatter: frontMatter, bodyLine: parts.bodyLine}, decodingFaults
}

// partFaults find every part of a post's folder but its file in English and the pictures it shows, and say why it is
// not read.
func partFaults(folder fs.FS, place string, shown map[string]bool) Faults {
	entries, faults := readFolder(folder, place)
	for _, entry := range entries {
		file := path.Join(place, entry.Name())
		switch name := entry.Name(); {
		case name == englishPostFile || shown[name]:
			// the post itself, which readPost reads, and a picture its body shows, which its inspection reads
		case name == revisionsFolder:
			faults = append(faults, fileFault(file, revisionsProblem))
		case name == widgetsFolder:
			faults = append(faults, fileFault(file, widgetsProblem))
		case strings.HasPrefix(name, postFilePrefix) && strings.HasSuffix(name, markdownEnding):
			faults = append(faults, fileFault(file, translationProblem))
		default:
			faults = append(faults, fileFault(file, strayPartProblem))
		}
	}
	return faults
}

// inspectPosts checks every post by itself, and that no two share a number, reading the pictures each shows from the
// folder: the posts as the snapshot keeps them, and the faults found.
func inspectPosts(folder fs.FS, posts []writtenPost) ([]Post, Faults) {
	checked := make([]Post, 0, len(posts))
	faults := Faults{}
	numbered := map[int]bool{}
	for _, written := range posts {
		post, place := written.post, path.Dir(written.file)
		folderInspection := &inspection{file: place, faults: nil}
		folderInspection.requireName(post.Slug)
		inspection := &inspection{file: written.file, faults: nil}
		inspectPostFields(inspection, post)
		if numbered[post.ID] {
			inspection.add(idField, sharedNumberProblem, post.ID)
		}
		numbered[post.ID] = true
		shown := inspectBody(inspection, writtenBody{body: post.Body, firstLine: written.bodyLine, folder: folder, place: place})
		post.Pictures = shown.read
		faults = slices.Concat(faults, folderInspection.faults, inspection.faults, inspectCharacters(written.file, written.frontMatter),
			partFaults(folder, place, shown.named))
		checked = append(checked, post)
	}
	return checked, faults
}

// inspectPostFields checks what a post's front matter says, field by field.
func inspectPostFields(inspection *inspection, post Post) {
	if post.ID <= 0 {
		inspection.add(idField, positiveProblem, post.ID)
	}
	inspection.requireLine(titleField, post.Title)
	inspection.requireLine("description", post.Description)
	if !statuses()[post.Status] {
		inspection.add("status", noneOfProblem, post.Status, kindList(statuses()))
	}
	inspectDates(inspection, post)
	inspectTags(inspection, post.Tags)
	for index, discussion := range post.Discussions {
		place := fmt.Sprintf("discussions[%d].", index)
		inspection.requireLine(place+"site", discussion.Site)
		inspection.requireSecureWebAddress(place+"address", discussion.Address)
	}
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
