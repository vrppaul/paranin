package content

import (
	"os"
	"path"
	"testing"
	"time"
)

// The folder of the fixture project whose section plays a recording
const recordingFolder = "testdata/content/projects/tidy-tool"

// A recording's length is read from its file, its movie header's duration in the units of its timescale; a file with
// no movie header has none to read.
func TestARecordingsLengthIsReadFromItsFile(test *testing.T) {
	cases := []struct {
		name   string
		file   string
		length time.Duration
		read   bool
	}{
		{name: "a recording", file: "tidying.mp4", length: 2 * time.Second, read: true},
		{name: "a file with no movie header", file: "unmeasured.mp4", length: 0, read: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			body, err := os.ReadFile(path.Join(recordingFolder, testCase.file))
			if err != nil {
				test.Fatalf("read %q: %v", testCase.file, err)
			}
			length, err := recordingLength(body)
			if (err == nil) != testCase.read || length != testCase.length {
				test.Errorf("recordingLength of %q = %v, %v; want %v, read: %t", testCase.file, length, err, testCase.length, testCase.read)
			}
		})
	}
}
