package content

import (
	"bytes"
	"fmt"
	"io/fs"
	"maps"
	"path"
	"slices"
	"strings"
)

// Where a project's own files stand, a folder of each project's named by its slug, as a post's pictures stand beside
// the post; and the one type of recording a project's page plays, an MP4 of H.264 with AAC, as the clips are encoded
const (
	projectsFolder     = "projects"
	recordingEnding    = ".mp4"
	recordingMediaType = "video/mp4"
)

// The most a recording may hold, in bytes: each is held in memory with its snapshot and sent whole to a reader who
// plays it; the clips, 1280px wide with their sound, are encoded to about 1.5 MB, and this leaves them room
const largestRecording = 4 << 20

// sectionList is a list of a project's page's sections at its field of the profile: its story's or its details'.
type sectionList struct {
	field    string
	sections []Section
}

// sectionListsOf are a project's page's sections, its story's and its details', at their fields of the profile.
func sectionListsOf(projectIndex int, project Project) []sectionList {
	return []sectionList{
		{field: fmt.Sprintf("projects[%d].sections", projectIndex), sections: project.Sections},
		{field: fmt.Sprintf("projects[%d].details", projectIndex), sections: project.Details},
	}
}

// readClips reads each clip of a project's page from the project's folder, its recording's length from the file, and
// keeps both files with the clip; a fault at the clip's field where a file is missing, of no type it may be, too
// large, or unread.
func readClips(folder fs.FS, projects []Project) Faults {
	inspection := &inspection{file: profileFile, faults: nil}
	for projectIndex, project := range projects {
		place := path.Join(projectsFolder, project.Slug)
		for _, list := range sectionListsOf(projectIndex, project) {
			for sectionIndex, section := range list.sections {
				if section.Clip != nil {
					field := fmt.Sprintf("%s[%d].clip", list.field, sectionIndex)
					readRecording(inspection, folder, place, field+".file", section.Clip)
					readStill(inspection, folder, place, field+".still", section.Clip)
				}
			}
		}
	}
	return inspection.faults
}

// readRecording reads a clip's recording from its project's folder and its length from the file, and keeps it.
func readRecording(inspection *inspection, folder fs.FS, place, field string, clip *Clip) {
	name := clip.File
	if path.Ext(name) != recordingEnding || !namedInFolder(name, recordingEnding) {
		inspection.add(field, recordingNameProblem, name)
		return
	}
	body, read := readProjectFile(inspection, folder, place, field, name, largestRecording)
	if !read {
		return
	}
	length, err := recordingLength(body)
	if err != nil {
		inspection.add(field, unmeasuredRecordingProblem, name, err)
		return
	}
	clip.Recording = Recording{MediaType: recordingMediaType, Length: length, Body: body}
}

// readStill reads a clip's still from its project's folder, a picture of a type a post's picture may be, and keeps it.
func readStill(inspection *inspection, folder fs.FS, place, field string, clip *Clip) {
	name := clip.Still
	types := pictureTypes()
	format, known := types[path.Ext(name)]
	switch {
	case !known:
		inspection.add(field, pictureTypeProblem, name, strings.Join(slices.Sorted(maps.Keys(types)), kindSeparator))
		return
	case !namedInFolder(name, path.Ext(name)):
		inspection.add(field, projectFileNameProblem, name)
		return
	}
	body, read := readProjectFile(inspection, folder, place, field, name, largestPicture)
	if !read {
		return
	}
	size, err := format.readSize(bytes.NewReader(body))
	if err != nil {
		inspection.add(field, unreadPictureProblem, name, err)
		return
	}
	clip.Picture = Picture{MediaType: format.mediaType, Width: size.Width, Height: size.Height, Body: body}
}

// readProjectFile reads a file of a project's folder by its name, no larger than the most it may hold, and whether it
// could; a fault at its field where it could not.
func readProjectFile(inspection *inspection, folder fs.FS, place, field, name string, most int) ([]byte, bool) {
	body, err := readHeld(folder, path.Join(place, name), most)
	switch {
	case err != nil:
		inspection.add(field, projectFileReadProblem, name, place, err)
		return nil, false
	case len(body) > most:
		inspection.add(field, largeProjectFileProblem, name, most)
		return nil, false
	}
	return body, true
}
