package content

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"io"
	"io/fs"
	"slices"
	"testing"
	"testing/fstest"
)

// Where a case writes the sample post's picture, by the name its text gives it, and the fixture's own picture, which
// its post shows
const (
	samplePictureName  = "margin.png"
	samplePictureFile  = "posts/sample-post/" + samplePictureName
	fixturePictureFile = "posts/reading-slowly/margin.png"
)

// One byte past the most a picture may hold, twice that most, and the quality a test's JPEG is written at
const (
	pastTheLargest  = largestPicture + 1
	twiceTheLargest = 2 * largestPicture
	jpegQualityFull = 100
)

// A WebP's head, lossless, three pixels by two, which is all the load reads of one: RIFF's chunk of 18 bytes, WEBP,
// and a VP8L chunk of five bytes, its signature, then the width less one and the height less one, 14 bits each, with
// a byte of padding after it
const webpOfThreeByTwo = "RIFF\x12\x00\x00\x00WEBPVP8L\x05\x00\x00\x00\x2f\x02\x40\x00\x00\x00"

// pictureFields are the fields of a loaded picture a test reads: what it is sent as, and its size in pixels.
type pictureFields struct {
	MediaType string
	Width     int
	Height    int
}

// A post's picture loads by the name its text gives it, read whole from the post's folder, with the type it is sent as
// and its size in pixels: a raster's from its head, a drawing's from its width and height, or from its view box where
// it names no size of its own; wherever it stands as a paragraph of its own, a list's item and a quote among them.
func TestAPostsPictureLoadsWithItsTypeAndSize(test *testing.T) {
	cases := []struct {
		name    string
		body    string
		file    string
		picture []byte
		want    pictureFields
	}{
		{
			name: "a PNG", body: "![A page](margin.png)", file: "margin.png", picture: fixturePicture(test),
			want: pictureFields{MediaType: "image/png", Width: 16, Height: 8},
		},
		{
			name: "a JPEG", body: "![A page](margin.jpg)", file: "margin.jpg", picture: jpegOfFiveByThree(test),
			want: pictureFields{MediaType: "image/jpeg", Width: 5, Height: 3},
		},
		{
			name: "a WebP", body: "![A page](margin.webp)", file: "margin.webp", picture: []byte(webpOfThreeByTwo),
			want: pictureFields{MediaType: "image/webp", Width: 3, Height: 2},
		},
		{
			name: "a drawing of its own size", body: "![A page](margin.svg)", file: "margin.svg",
			picture: []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="40px" height="20.4" viewBox="0 0 4 2"/>`),
			want:    pictureFields{MediaType: "image/svg+xml", Width: 40, Height: 20},
		},
		{
			name: "a drawing sized by its view box", body: "![A page](margin.svg)", file: "margin.svg",
			picture: []byte(`<?xml version="1.0"?>` + "\n" + `<svg xmlns="http://www.w3.org/2000/svg" width="100%" viewBox="0 0 120 60"><rect width="1" height="1"/></svg>`),
			want:    pictureFields{MediaType: "image/svg+xml", Width: 120, Height: 60},
		},
		{
			name: "a drawing of its width alone, its height by its view box's shape", body: "![A page](margin.svg)", file: "margin.svg",
			picture: []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="40" height="2em" viewBox="0,0,4,2"/>`),
			want:    pictureFields{MediaType: "image/svg+xml", Width: 40, Height: 20},
		},
		{
			name: "a drawing of its height alone, its width by its view box's shape", body: "![A page](margin.svg)", file: "margin.svg",
			picture: []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="50%" height="10" viewBox="0 0 4 2"/>`),
			want:    pictureFields{MediaType: "image/svg+xml", Width: 20, Height: 10},
		},
		{
			name: "a picture under a caption", body: `![A page](margin.png "The margin")`, file: "margin.png", picture: fixturePicture(test),
			want: pictureFields{MediaType: "image/png", Width: 16, Height: 8},
		},
		{
			name: "a picture in a list's item, after its words", body: "- A page:\n\n  ![A page](margin.png)", file: "margin.png",
			picture: fixturePicture(test),
			want:    pictureFields{MediaType: "image/png", Width: 16, Height: 8},
		},
		{
			name: "a picture in a quote", body: "> ![A page](margin.png)", file: "margin.png", picture: fixturePicture(test),
			want: pictureFields{MediaType: "image/png", Width: 16, Height: 8},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			folder := writtenFile(fixtureFolder(test), samplePostFile, sampleWith(test, samplePost, "Its text.", testCase.body))
			snapshot, err := Load(writtenFile(folder, "posts/sample-post/"+testCase.file, string(testCase.picture)))
			if err != nil {
				test.Fatalf("Load with %q beside %q = %v, want no error", testCase.file, testCase.body, err)
			}
			pictures := postNamed(test, snapshot, "", samplePostSlug).Pictures
			picture, found := pictures[testCase.file]
			got := pictureFields{MediaType: picture.MediaType, Width: picture.Width, Height: picture.Height}
			if !found || got != testCase.want || !bytes.Equal(picture.Body, testCase.picture) || len(pictures) != 1 {
				test.Errorf("the pictures of %q = %d, %q as %+v, found %t; want it alone, as %+v, holding the file", testCase.body,
					len(pictures), testCase.file, got, found, testCase.want)
			}
		})
	}
}

// Every fault of a post's picture stands at the line of the text that shows it, or at the file where no line does: a
// picture with no alt text, or one holding a mark; one among words, or in a footnote, which no face draws as a figure;
// an address outside the post's folder, a remote one among them; a file missing, of a type the site serves no picture
// as, named otherwise than a slug, over the size a picture may hold, that does not read as its type, or a drawing of no
// size; a caption holding a character no value may hold; and a picture of the folder no line shows.
func TestEveryFaultOfAPictureStandsAtItsPlace(test *testing.T) {
	atTheLine := faultPlace{file: samplePostFile, field: "", line: 12}
	cases := []struct {
		name    string
		body    string
		file    string
		picture []byte
		want    faultPlace
	}{
		{name: "no alt text", body: "![](margin.png)", file: samplePictureName, picture: fixturePicture(test), want: atTheLine},
		{name: "alt text of spaces alone", body: "![  ](margin.png)", file: samplePictureName, picture: fixturePicture(test), want: atTheLine},
		{name: "a mark in the alt text", body: "![A *slow* page](margin.png)", file: samplePictureName, picture: fixturePicture(test), want: atTheLine},
		{name: "among words", body: "Its ![A page](margin.png) text.", file: samplePictureName, picture: fixturePicture(test), want: atTheLine},
		{name: "inside a link", body: "[![A page](margin.png)](https://example.test)", file: samplePictureName, picture: fixturePicture(test), want: atTheLine},
		{name: "opening a list's item", body: "- ![A page](margin.png)", file: samplePictureName, picture: fixturePicture(test), want: atTheLine},
		{
			name: "opening a loose list's item", body: "- ![A page](margin.png)\n\n  Its words.", file: samplePictureName, picture: fixturePicture(test),
			want: atTheLine,
		},
		{
			name: "in a footnote", body: "Its text.[^note]\n\n[^note]: ![A page](margin.png)", file: samplePictureName, picture: fixturePicture(test),
			want: faultPlace{file: samplePostFile, field: "", line: 14},
		},
		{name: "a file missing", body: "![A page](margin.png)", file: "other.png", picture: fixturePicture(test), want: atTheLine},
		{name: "a path out of the folder", body: "![A page](../reading-slowly/margin.png)", file: "", picture: nil, want: atTheLine},
		{name: "a path from the root", body: "![A page](/margin.png)", file: samplePictureName, picture: fixturePicture(test), want: atTheLine},
		{name: "a path into a folder", body: "![A page](figures/margin.png)", file: "figures/margin.png", picture: fixturePicture(test), want: atTheLine},
		{name: "a remote address", body: "![A page](https://example.test/margin.png)", file: samplePictureName, picture: fixturePicture(test), want: atTheLine},
		{name: "a type the site serves no picture as", body: "![A page](margin.gif)", file: "margin.gif", picture: fixturePicture(test), want: atTheLine},
		{name: "a name in capitals", body: "![A page](Margin.png)", file: "Margin.png", picture: fixturePicture(test), want: atTheLine},
		{
			name: "a file past the size", body: "![A page](margin.png)", file: samplePictureName,
			picture: append(fixturePicture(test), make([]byte, pastTheLargest)...), want: atTheLine,
		},
		{name: "a file that is not its type", body: "![A page](margin.png)", file: samplePictureName, picture: []byte(`<svg/>`), want: atTheLine},
		{name: "a JPEG named a PNG", body: "![A page](margin.png)", file: samplePictureName, picture: jpegOfFiveByThree(test), want: atTheLine},
		{
			name: "a drawing of no size", body: "![A page](margin.svg)", file: "margin.svg",
			picture: []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="4em" height="2em"/>`), want: atTheLine,
		},
		{
			name: "a drawing of a width below nothing", body: "![A page](margin.svg)", file: "margin.svg",
			picture: []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="-4" height="2"/>`), want: atTheLine,
		},
		{
			name: "a drawing wider than a size holds", body: "![A page](margin.svg)", file: "margin.svg",
			picture: []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="1e12" height="2"/>`), want: atTheLine,
		},
		{
			name: "a drawing of a width that is no number", body: "![A page](margin.svg)", file: "margin.svg",
			picture: []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="NaN" height="2"/>`), want: atTheLine,
		},
		{
			name: "a drawing outside SVG's namespace", body: "![A page](margin.svg)", file: "margin.svg",
			picture: []byte(`<svg width="4" height="2"/>`), want: atTheLine,
		},
		{name: "a drawing that is no SVG", body: "![A page](margin.svg)", file: "margin.svg", picture: []byte(`<html width="4" height="2"/>`), want: atTheLine},
		{name: "a caption holding a control character", body: `![A page](margin.png "A &#27; caption")`, file: samplePictureName, picture: fixturePicture(test), want: atTheLine},
		{
			name: "a picture no line shows", body: "Its text.", file: samplePictureName, picture: fixturePicture(test),
			want: faultPlace{file: samplePictureFile, field: "", line: 0},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			folder := writtenFile(fixtureFolder(test), samplePostFile, sampleWith(test, samplePost, "Its text.", testCase.body))
			if testCase.file != "" {
				folder = writtenFile(folder, "posts/sample-post/"+testCase.file, string(testCase.picture))
			}
			faults := faultsOf(test, folder)
			if !slices.ContainsFunc(faults, func(fault Fault) bool { return placeOf(fault) == testCase.want }) {
				test.Errorf("faults with %q beside %q = %q, want one at %+v", testCase.body, testCase.file, faults, testCase.want)
			}
		})
	}
}

// A picture that is no file of its own, a named pipe whose read would wait for a writer or a folder, is refused, never
// opened.
func TestAPictureThatIsNoRegularFileIsRefused(test *testing.T) {
	cases := []struct {
		name  string
		file  string
		entry *fstest.MapFile
	}{
		{name: "a named pipe", file: samplePictureFile, entry: &fstest.MapFile{Data: fixturePicture(test), Mode: fs.ModeNamedPipe}},
		{name: "a folder", file: samplePictureFile + "/inside.png", entry: &fstest.MapFile{Data: fixturePicture(test)}},
	}
	want := faultPlace{file: samplePostFile, field: "", line: 12}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			folder := writtenFile(fixtureFolder(test), samplePostFile, sampleWith(test, samplePost, "Its text.", "![A page](margin.png)"))
			folder[testCase.file] = testCase.entry
			if faults := faultsOf(test, folder); !slices.ContainsFunc(faults, func(fault Fault) bool { return placeOf(fault) == want }) {
				test.Errorf("faults with %s at %s = %q, want one at %+v", testCase.name, samplePictureFile, faults, want)
			}
		})
	}
}

// A picture is read no further than one byte past the most a picture may hold, however large its file, as one still
// growing while it is read, and refused.
func TestAPictureIsReadNoFurtherThanAPictureMayHold(test *testing.T) {
	folder := writtenFile(fixtureFolder(test), samplePostFile, sampleWith(test, samplePost, "Its text.", "![A page](margin.png)"))
	large := &countedFile{left: twiceTheLargest, read: 0}
	counted := map[string]*countedFile{samplePictureFile: large}
	_, err := Load(largeFolder{files: writtenFile(folder, samplePictureFile, ""), counted: counted})
	var faults Faults
	want := faultPlace{file: samplePostFile, field: "", line: 12}
	if !errors.As(err, &faults) || !slices.ContainsFunc(faults, func(fault Fault) bool { return placeOf(fault) == want }) ||
		large.read != pastTheLargest {
		test.Errorf("Load with a picture of %d bytes = %v, %d bytes read; want a fault at %+v, %d bytes read",
			twiceTheLargest, err, large.read, want, pastTheLargest)
	}
}

// largeFolder is a folder of files in memory, some of which, by their names, are read as counted files.
type largeFolder struct {
	files   fstest.MapFS
	counted map[string]*countedFile
}

func (folder largeFolder) Open(name string) (fs.File, error) {
	if counted, isCounted := folder.counted[name]; isCounted {
		return counted, nil
	}
	return folder.files.Open(name)
}

func (folder largeFolder) Stat(name string) (fs.FileInfo, error) {
	return folder.files.Stat(name)
}

// countedFile is a file of zeros, how many it has left, and how many bytes were read from it.
type countedFile struct {
	left int
	read int
}

func (file *countedFile) Read(buffer []byte) (int, error) {
	if file.left == 0 {
		return 0, io.EOF
	}
	count := min(len(buffer), file.left)
	clear(buffer[:count])
	file.left -= count
	file.read += count
	return count, nil
}

func (file *countedFile) Stat() (fs.FileInfo, error) {
	return nil, fs.ErrInvalid
}

func (file *countedFile) Close() error {
	return nil
}

// fixturePicture is the picture the fixture's post shows, a PNG.
func fixturePicture(test *testing.T) []byte {
	test.Helper()
	return slices.Clone(fixtureFolder(test)[fixturePictureFile].Data)
}

// jpegOfFiveByThree is a JPEG five pixels by three.
func jpegOfFiveByThree(test *testing.T) []byte {
	test.Helper()
	var written bytes.Buffer
	if err := jpeg.Encode(&written, image.NewGray(image.Rect(0, 0, 5, 3)), &jpeg.Options{Quality: jpegQualityFull}); err != nil {
		test.Fatalf("write a JPEG: %v", err)
	}
	return written.Bytes()
}
