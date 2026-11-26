package content

import (
	"encoding/binary"
	"os"
	"path"
	"slices"
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

// What a test's recordings are built of: a timescale of a thousand units a second, a duration of two seconds in them,
// the 64-bit size's mark, the size a box claims past what the file holds, a large box's body, a version no header has,
// and the duration of 2^63 units, past what a length holds
const (
	testTimescale      = 1000
	testDuration       = 2000
	testLength         = 2 * time.Second
	sizePastFile       = 1000
	largeBodyBytes     = 100_000
	unknownVersion     = 2
	pastLengthDuration = uint64(1) << 63
)

// A recording's length is read through the boxes as MP4 sizes them, a box of a 64-bit size, one that runs to the file's
// end and a movie after a large box of media among them; a file whose boxes are cut short or claim more than it holds,
// or whose movie header is empty, of no version known or of no timescale, or whose length passes what one holds, has
// none to read, and none of them stops the load.
func TestARecordingsBoxesAreReadAsTheirSizesSay(test *testing.T) {
	cases := []struct {
		name   string
		body   []byte
		length time.Duration
		read   bool
	}{
		{name: "a movie header of the first version", body: box(movieBox, shortHeader(testTimescale, testDuration)), length: testLength, read: true},
		{name: "a movie header of the second version", body: box(movieBox, longHeader(testTimescale, testDuration)), length: testLength, read: true},
		{name: "a box of a 64-bit size", body: largeBox(movieBox, shortHeader(testTimescale, testDuration)), length: testLength, read: true},
		{name: "a box that runs to the file's end", body: sized(toTheEndMark, movieBox, shortHeader(testTimescale, testDuration)), length: testLength, read: true},
		{
			name: "a movie after a large box of media", body: slices.Concat(box("mdat", make([]byte, largeBodyBytes)), box(movieBox, shortHeader(testTimescale, testDuration))),
			length: testLength, read: true,
		},
		{name: "an empty file", body: []byte{}, length: 0, read: false},
		{name: "a box's header cut short", body: []byte{0, 0, 0}, length: 0, read: false},
		{name: "a box's size past the file", body: sized(sizePastFile, movieBox, shortHeader(testTimescale, testDuration)), length: 0, read: false},
		{name: "a 64-bit size past the file", body: largeSized(sizePastFile, movieBox, shortHeader(testTimescale, testDuration)), length: 0, read: false},
		{name: "a 64-bit size under its header", body: largeSized(boxHeaderBytes, movieBox, shortHeader(testTimescale, testDuration)), length: 0, read: false},
		{name: "an empty movie header", body: box(movieBox, box(movieHeaderBox)), length: 0, read: false},
		{name: "a movie header of no version known", body: box(movieBox, box(movieHeaderBox, []byte{unknownVersion, 0, 0, 0})), length: 0, read: false},
		{name: "a movie header of no timescale", body: box(movieBox, shortHeader(0, testDuration)), length: 0, read: false},
		{name: "a length past what one holds", body: box(movieBox, longHeader(testTimescale, pastLengthDuration)), length: 0, read: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			length, err := recordingLength(testCase.body)
			if (err == nil) != testCase.read || length != testCase.length {
				test.Errorf("recordingLength = %v, %v; want %v, read: %t", length, err, testCase.length, testCase.read)
			}
		})
	}
}

// box is an MP4 box as a test builds it: its size, counting it whole, its type, and what it holds.
func box(kind string, inside ...[]byte) []byte {
	body := slices.Concat(inside...)
	return sized(uint32(boxHeaderBytes+len(body)), kind, body)
}

// sized is a box that claims a size of its own, whatever it holds.
func sized(size uint32, kind string, body []byte) []byte {
	return slices.Concat(binary.BigEndian.AppendUint32(nil, size), []byte(kind), body)
}

// largeBox is a box whose size follows its type in eight bytes.
func largeBox(kind string, body []byte) []byte {
	return largeSized(uint64(boxHeaderBytes+largeSizeBytes+len(body)), kind, body)
}

// largeSized is a box of a 64-bit size that claims a size of its own, whatever it holds.
func largeSized(size uint64, kind string, body []byte) []byte {
	return slices.Concat(binary.BigEndian.AppendUint32(nil, largeSizeMark), []byte(kind), binary.BigEndian.AppendUint64(nil, size), body)
}

// shortHeader is a movie header of the first version: its version and flags, its two times, its timescale and its
// duration, four bytes each but the first.
func shortHeader(timescale, duration uint32) []byte {
	times := make([]byte, headerTimes*shortDuration)
	fields := slices.Concat([]byte{shortHeaderVersion, 0, 0, 0}, times, binary.BigEndian.AppendUint32(nil, timescale), binary.BigEndian.AppendUint32(nil, duration))
	return box(movieHeaderBox, fields)
}

// longHeader is a movie header of the second version, its times and its duration eight bytes each.
func longHeader(timescale uint32, duration uint64) []byte {
	times := make([]byte, headerTimes*longDuration)
	fields := slices.Concat([]byte{longHeaderVersion, 0, 0, 0}, times, binary.BigEndian.AppendUint32(nil, timescale), binary.BigEndian.AppendUint64(nil, duration))
	return box(movieHeaderBox, fields)
}
