package content

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.yaml.in/yaml/v3"
)

// How a revision's file is named in its post's revisions folder: the day its version came out, as a post writes the
// day it was updated, then the language its text is written in and the ending of a post's file, YYYY-MM-DD.en.md
const revisionNameJoint = "."

// The line that opens and closes front matter, as a revision taken writes it
const frontMatterLine = frontMatterFence + string(lineFeed)

// Revision is an older version of a post, a frozen copy its author took before rewriting it (docs/FOUNDATION.md §5.4,
// Revisions): the day the version came out, its file's name; the title, the lead and the body that version had; and its
// links inside the site, as a post's. Its post's facts, its number, status and tags, are the post's alone.
type Revision struct {
	Day         Day                       `yaml:"-"`
	Title       string                    `yaml:"title"`
	Description string                    `yaml:"description"`
	Body        Markdown                  `yaml:"-"`
	SiteLinks   map[*ast.Link]WrittenLink `yaml:"-"`
}

// writtenRevision is a revision as its file writes it, as writtenPost is a post.
type writtenRevision struct {
	file        string
	revision    Revision
	frontMatter *yaml.Node
	bodyLine    int
}

// readRevisions reads the revisions of the post of a slug by the language each is in, English or one of the languages
// it may be translated into, oldest first, as their names order them, with the faults that kept one from being read
// and those of the folder's other files, which keep none from being read; none where the post has no revisions folder.
func readRevisions(folder fs.FS, postParser parser.Parser, slug string, languages []string) (map[string][]writtenRevision, Faults, Faults) {
	place := path.Join(postsFolder, slug, revisionsFolder)
	entries, unreadable := readFolder(folder, place)
	revisions, misplaced := map[string][]writtenRevision{}, Faults{}
	for _, entry := range entries {
		file := path.Join(place, entry.Name())
		day, language, named := revisionName(entry.Name())
		switch {
		case entry.IsDir() || !named:
			misplaced = append(misplaced, fileFault(file, strayRevisionProblem))
		case language != english && !slices.Contains(languages, language):
			misplaced = append(misplaced, fileFault(file, noWordsProblem))
		default:
			revision, faults := readRevision(folder, postParser, file, day)
			revisions[language] = append(revisions[language], revision)
			unreadable = append(unreadable, faults...)
		}
	}
	return revisions, unreadable, misplaced
}

// revisionName reads a revision's file's name: the day its version came out and the language it is written in, and
// whether the name is a revision's at all.
func revisionName(name string) (Day, string, bool) {
	stem, isMarkdown := strings.CutSuffix(name, markdownEnding)
	written, language, parted := strings.Cut(stem, revisionNameJoint)
	day := readDay(written)
	return day, language, isMarkdown && parted && day.known
}

// readRevision reads a revision's file as readPost reads a post's: its front matter decoded strictly and its body
// parsed once by the posts' parser.
func readRevision(folder fs.FS, postParser parser.Parser, file string, day Day) (writtenRevision, Faults) {
	// what a revision read with faults is never inspected, since its faults end the load
	var unread writtenRevision
	data, err := fs.ReadFile(folder, file)
	if err != nil {
		return unread, Faults{unreadFault(file, err)}
	}
	parts, splitFaults := splitPost(file, data)
	if len(splitFaults) > 0 {
		return unread, splitFaults
	}
	revision, frontMatter, decodingFaults := decodeStrictly[Revision](file, parts.frontMatter)
	revision.Day = day
	revision.Body = Markdown{Source: parts.body, Tree: postParser.Parse(text.NewReader(parts.body))}
	return writtenRevision{file: file, revision: revision, frontMatter: frontMatter, bodyLine: parts.bodyLine}, decodingFaults
}

// inspectRevisions checks a post's revisions as its text is checked, each by itself, and each dated before its current
// version's: the revisions as the snapshot keeps them, the pictures their bodies show, and the faults found.
func inspectRevisions(folder fs.FS, site WebAddress, post Post, written writtenPost) ([]Revision, shownPictures, Faults) {
	shown := shownPictures{read: map[string]Picture{}, named: map[string]bool{}}
	faults := Faults{}
	revisions := make([]Revision, 0, len(written.revisions))
	for _, taken := range written.revisions {
		revision := taken.revision
		inspection := &inspection{file: taken.file, faults: nil}
		inspection.requireLine(titleField, revision.Title)
		inspection.requireLine(descriptionField, revision.Description)
		inspectRevisionDay(inspection, post, revision.Day)
		pictures, links := inspectBody(inspection, writtenBody{
			body: revision.Body, firstLine: taken.bodyLine, folder: folder, place: path.Dir(written.file), site: site,
		})
		revision.SiteLinks = links
		shown.add(pictures)
		faults = slices.Concat(faults, inspection.faults, inspectCharacters(taken.file, taken.frontMatter))
		revisions = append(revisions, revision)
	}
	return revisions, shown, faults
}

// inspectRevisionDay holds a revision's day to no earlier than the day its post was published, its first version's,
// and to before its current version's, which came after it; so a post with revisions is dated by an update.
func inspectRevisionDay(inspection *inspection, post Post, day Day) {
	switch current := post.VersionDay(); {
	case post.Published.known && day.start.Before(post.Published.day()):
		inspection.add("", earlyRevisionProblem, day.written, post.Published.written)
	case current.known && !day.start.Before(current.start):
		inspection.add("", lateRevisionProblem, day.written, current.written)
	}
}

// VersionDay is the day a post's current version came out, the day its text last changed: the day it was updated,
// where it says one, else the day it was published, in its own zone; unknown for a draft not yet dated.
func (post Post) VersionDay() Day {
	if post.Updated.written != "" {
		return post.Updated
	}
	if !post.Published.known {
		return Day{written: "", start: time.Time{}, known: false}
	}
	start := post.Published.day()
	return Day{written: start.Format(dayLayout), start: start, known: true}
}

// Current is a post's current version as its revisions hold theirs: the day it came out, its title, its lead, its body
// and its links inside the site.
func (post Post) Current() Revision {
	return Revision{Day: post.VersionDay(), Title: post.Title, Description: post.Description, Body: post.Body, SiteLinks: post.SiteLinks}
}

// The faults of a revision asked of a post: its text is its newest revision's already, no version to take; or its
// current version came out on the day it is asked or later, so a change on that day is an edit of it, and the one
// version a day it holds is taken the next day at the soonest
var (
	errNoChange = errors.New("the post's title, lead and text are its newest revision's already")
	errSameDay  = errors.New("the post's current version came out today or later: a change on its day is an edit of it")
)

// RevisionFile is a revision as `paranin revise` writes it beside its post: its path in the content folder, and what
// it holds.
type RevisionFile struct {
	Path string
	Text []byte
}

// RevisionOf is the revision of a post's current version its author takes at a moment, before rewriting the post: its
// title and lead as front matter, then its body as written, at the day the version came out, the post's file then
// dated by UpdatedText; a fault where the version came out on the moment's day or later, in the post's own zone, or the
// post's text is its newest revision's.
func RevisionOf(post Post, now time.Time) (RevisionFile, error) {
	taken := post.Current()
	if !taken.Day.start.Before(post.Published.dayAt(now)) {
		return RevisionFile{Path: "", Text: nil}, errSameDay
	}
	if newest := len(post.Revisions) - 1; newest >= 0 && bytes.Equal(post.Revisions[newest].text(), taken.text()) {
		return RevisionFile{Path: "", Text: nil}, errNoChange
	}
	// its front matter is its two strings alone, which always encode
	fields, _ := yaml.Marshal(taken)
	name := taken.Day.written + revisionNameJoint + fileLanguage(post.Translation) + markdownEnding
	return RevisionFile{
		Path: path.Join(postsFolder, post.Slug, revisionsFolder, name),
		Text: slices.Concat([]byte(frontMatterLine), fields, []byte(frontMatterLine), taken.Body.Source),
	}, nil
}

// PostFile is where a post's text in its language is written in the content folder.
//
//wrapper:names a loaded post's file for whoever writes it, as the loader names it by its slug and language
func PostFile(post Post) string {
	return textPath(post.Slug, post.Translation)
}

// What parts a field's name from its value, as a post's file is dated writes a field of its own; and where YAML counts
// a node's line and column from
const (
	fieldJoint     = ": "
	yamlCountsFrom = 1
)

// UpdatedText is a post's file as its author wrote it, dated as updated on the day a moment falls on in the post's own
// zone, as `paranin revise` dates it once it takes the version that stood (docs/FOUNDATION.md §5.4, Revisions): its
// updated: written in place of the day it said, or on a line of its own after its published: where it said none;
// every other byte, a comment and each field's form among them, kept, which a document encoded again would not keep.
func UpdatedText(post Post, written []byte, now time.Time) ([]byte, error) {
	file := PostFile(post)
	parts, faults := splitPost(file, written)
	if len(faults) > 0 {
		return nil, faults
	}
	var frontMatter yaml.Node
	if err := yaml.Unmarshal(parts.frontMatter, &frontMatter); err != nil {
		return nil, fmt.Errorf("read the front matter of %q: %w", file, err)
	}
	day := post.Published.dayAt(now).Format(dayLayout)
	lines := bytes.SplitAfter(written, []byte{lineFeed})
	if updated, said := fieldValue(&frontMatter, updatedField); said {
		row, start := updated.Line-yamlCountsFrom, updated.Column-yamlCountsFrom
		lines[row] = slices.Concat(lines[row][:start], []byte(day), lines[row][start+writtenLength(updated):])
		return bytes.Join(lines, nil), nil
	}
	published, said := fieldValue(&frontMatter, publishedField)
	// a post out says when it was published as it loads, but its file may change before it is dated
	if !said {
		return nil, fmt.Errorf("%q names no published: to write its updated: after", file)
	}
	ownLine := []byte(updatedField + fieldJoint + day + string(lineFeed))
	return bytes.Join(slices.Insert(lines, published.Line, ownLine), nil), nil
}

// fieldValue is the value of a field of a front matter's top level, by its name, and whether it says one.
func fieldValue(frontMatter *yaml.Node, name string) (*yaml.Node, bool) {
	if len(frontMatter.Content) == 0 {
		return nil, false
	}
	return valueOf(frontMatter.Content[0], name)
}

// writtenLength is how many bytes a scalar takes in its file: its value, between its quotes where it is quoted, which
// a day, holding no character to escape, writes as it is.
func writtenLength(scalar *yaml.Node) int {
	if scalar.Style&(yaml.SingleQuotedStyle|yaml.DoubleQuotedStyle) != 0 {
		return len(scalar.Value) + quotesAround
	}
	return len(scalar.Value)
}

// The quotes a quoted scalar stands between
const quotesAround = 2

// What parts a version's title, its lead and its body in its text: neither a title nor a lead holds a line feed
const textJoint = string(lineFeed)

// text is a version's text, the words a revision keeps and a translation translates: its title, its lead and its body
// as its file writes them, each parted from the next. Its post's facts, a status, a day or a tag, are none of it.
func (revision Revision) text() []byte {
	return slices.Concat([]byte(revision.Title+textJoint+revision.Description+textJoint), revision.Body.Source)
}
