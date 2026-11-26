package content

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

// How an MP4 holds its parts (ISO/IEC 14496-12): boxes one after another, each its size in four bytes, the size
// counting the box whole, then its type in four letters, then what it holds; a size of one says the size follows the
// type in eight bytes, and a size of nought that the box runs to the file's end. The movie box holds the movie header,
// whose timescale says how many units a second holds and whose duration how long the movie runs in them
const (
	boxSizeBytes      = 4
	boxHeaderBytes    = 8
	largeSizeBytes    = 8
	largeSizeMark     = 1
	toTheEndMark      = 0
	movieBox          = "moov"
	movieHeaderBox    = "mvhd"
	timescaleBytes    = 4
	shortDuration     = 4
	longDuration      = 8
	headerVersionByte = 0
)

// A movie header opens with its version and flags, then its two times, of its making and of its last change, then its
// timescale and its duration; its two versions part in how many bytes the times and the duration take, the first four,
// the second eight
const (
	versionAndFlagsBytes      = 4
	headerTimes               = 2
	shortHeaderVersion   byte = 0
	longHeaderVersion    byte = 1
)

// timeBytesOf are how many bytes a movie header's times and duration take, by its version.
func timeBytesOf() map[byte]int {
	return map[byte]int{shortHeaderVersion: shortDuration, longHeaderVersion: longDuration}
}

// recordingLength is how long a recording runs, read from its movie header, so no length is written by hand to go
// stale; a fault where the file holds no movie header it can read.
func recordingLength(body []byte) (time.Duration, error) {
	movie, found := boxIn(body, movieBox)
	if !found {
		return 0, fmt.Errorf("no %q box", movieBox)
	}
	header, found := boxIn(movie, movieHeaderBox)
	if !found {
		return 0, fmt.Errorf("no %q box in its %q", movieHeaderBox, movieBox)
	}
	return headerLength(header)
}

// boxIn is what the first box of a type holds among boxes standing one after another, and whether there is one.
func boxIn(boxes []byte, kind string) ([]byte, bool) {
	rest := boxes
	for len(rest) >= boxHeaderBytes {
		size, headerSize, whole := boxSizes(rest)
		if !whole {
			return nil, false
		}
		if string(rest[boxSizeBytes:boxHeaderBytes]) == kind {
			return rest[headerSize:size], true
		}
		rest = rest[size:]
	}
	return nil, false
}

// boxSizes are the size of the box that opens boxes, the box whole, and the size of its header, and whether the box
// stands whole among them.
func boxSizes(boxes []byte) (uint64, uint64, bool) {
	size, headerSize := uint64(binary.BigEndian.Uint32(boxes)), uint64(boxHeaderBytes)
	switch size {
	case largeSizeMark:
		if len(boxes) < boxHeaderBytes+largeSizeBytes {
			return 0, 0, false
		}
		size, headerSize = binary.BigEndian.Uint64(boxes[boxHeaderBytes:]), boxHeaderBytes+largeSizeBytes
	case toTheEndMark:
		size = uint64(len(boxes))
	}
	return size, headerSize, size >= headerSize && size <= uint64(len(boxes))
}

// headerLength is the length a movie header holds: its duration over its timescale.
func headerLength(header []byte) (time.Duration, error) {
	if len(header) <= headerVersionByte {
		return 0, fmt.Errorf("an empty %q box", movieHeaderBox)
	}
	timeBytes, known := timeBytesOf()[header[headerVersionByte]]
	if !known {
		return 0, fmt.Errorf("a %q box of version %d", movieHeaderBox, header[headerVersionByte])
	}
	timescaleAt := versionAndFlagsBytes + headerTimes*timeBytes
	durationAt := timescaleAt + timescaleBytes
	if len(header) < durationAt+timeBytes {
		return 0, fmt.Errorf("a %q box cut short", movieHeaderBox)
	}
	timescale := uint64(binary.BigEndian.Uint32(header[timescaleAt:]))
	duration := uint64(binary.BigEndian.Uint32(header[durationAt:]))
	if timeBytes == longDuration {
		duration = binary.BigEndian.Uint64(header[durationAt:])
	}
	if timescale == 0 {
		return 0, fmt.Errorf("a %q box of no timescale", movieHeaderBox)
	}
	length := float64(duration) / float64(timescale) * float64(time.Second)
	if length >= math.MaxInt64 {
		return 0, fmt.Errorf("a %q box longer than a length holds", movieHeaderBox)
	}
	return time.Duration(length), nil
}
