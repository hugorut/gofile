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

// Base64ToDecoder take an input of base64 bytes and strips the encoding signature.
// returns a bytes reader so as to conform with the ReadSeeker interface
func Base64ToDecoder(src []byte) io.ReadSeeker {
	reader := bytes.NewReader(StripBaseEncoding(src))
	b, _ := ioutil.ReadAll(base64.NewDecoder(base64.StdEncoding, reader))

	return bytes.NewReader(b)
}

// StripBaseEncoding strips the base encoding prefece from an encoded image.
func StripBaseEncoding(image []byte) []byte {
	r := regexp.MustCompile("data:(image/[^;]+);base64,")
	return r.ReplaceAll(image, []byte(""))
}

// Base64ImageType returns the extension of a base64 encoded image.
func Base64ImageType(image []byte) string {
	r := regexp.MustCompile("data:image/([^;]+);base64,")
	matches := r.FindSubmatch(image)

	if len(matches) < 2 {
		return "txt"
	}

	return string(matches[1])
}

// SanitizePath removes whitespace from the path so that files can be persisted without error.
func SanitizePath(path string) string {
	reg := regexp.MustCompile("\\s+")
	return reg.ReplaceAllString(strings.TrimSpace(path), "-")
}

// GetMIMETypeFromPath returns mime type from a path to a file.
func GetMIMETypeFromPath(path string) string {
	r := regexp.MustCompile("(\\.\\w+$)")
	extension := r.FindStringSubmatch(path)

	if len(extension) == 0 {
		return "text/plain"
	}

	return mime.TypeByExtension(extension[0])
}

// joinPath concatenates a path with a give path and extension.
func joinPath(path, ext string) string {
	r := regexp.MustCompile("^.+/")
	extension := r.ReplaceAllString(ext, "")

	return path + "." + extension
}

// FileSystem interface provides a fluent interface to file interactions.
type FileSystem interface {
	// Put creates a file into a filesystem and return a File interface which
	// will give you information about the location and status of
	// the uploaded file the path must contain the file and full extension.
	//
	// e.g. /path/to/my/funky/file.gif
	Put(src io.ReadSeeker, path string) (File, error)

	// get a file from the file system, return the File interface
	// so that all generic file interactions can be facilitated.
	Get(path string) (File, error)
}

// File interface is the generic interface that is returned from a FileSytem.
type File interface {
	io.Closer
	io.ReadWriteSeeker
	Stat() (os.FileInfo, error)
}
