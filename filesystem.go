package gofile

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"regexp"
	"strings"
)

// take an input of base64 bytes and strinp the encoding signature
// returns a bas64 decoder with std encoding
func Base64ToDecoder(src []byte) io.ReadSeeker {
	reader := bytes.NewReader(StripBaseEncoding(src))
	b, _ := ioutil.ReadAll(base64.NewDecoder(base64.StdEncoding, reader))
	return bytes.NewReader(b)
}

//	strip the base encoding prefece from an encoded image
func StripBaseEncoding(image []byte) []byte {
	r := regexp.MustCompile("data:(image/[^;]+);base64,")
	return r.ReplaceAll(image, []byte(""))
}

// determine the base 64 image type from an encoded image
func Base64ImageType(image []byte) string {
	r := regexp.MustCompile("data:(image/[^;]+);base64,")
	matches := r.FindSubmatch(image)

	return string(matches[1])
}

// remove whitespace from the path so that files can be persisted without error
func SanitizePath(path string) string {
	reg := regexp.MustCompile("\\s+")
	return reg.ReplaceAllString(strings.TrimSpace(path), "-")
}

// Get the mime type from a path to a file
func GetMIMETypeFromPath(path string) string {
	r := regexp.MustCompile("(\\.\\w+$)")
	extension := r.FindStringSubmatch(path)[0]

	return mime.TypeByExtension(extension)
}

func joinPath(path, fileType string) string {
	r := regexp.MustCompile("^.+/")
	extension := r.ReplaceAllString(fileType, "")
	return path + "." + extension
}

// the wrapped for the filesystem interface which allows an fluent interface
type FileSystem interface {
	// Put a file into a filesystem and return a File interface which
	// will give you information about the location and status of
	// the uploaded file the path must contain the file and full extension
	//
	// e.g. /path/to/my/funky/file.gif
	Put(src io.ReadSeeker, path string) (File, error)

	// get a file from the file system, return the File interface
	// so that all generic file interactions can be facilitated
	Get(path string) (File, error)
}

// the File which will be returned by the fs
type File interface {
	io.Closer
	io.ReadWriteSeeker
	Stat() (os.FileInfo, error)
}
