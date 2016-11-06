package filesystem

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
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

func joinPath(path, fileType string) string {
	r := regexp.MustCompile("^.+/")
	extension := r.ReplaceAllString(fileType, "")
	return path + "." + extension
}

// the core filesystem which can be extended and mocked
type CoreFs interface {
	Open(name string) (file, error)
	Create(name string) (file, error)
	Stat(name string) (os.FileInfo, error)
	Copy(dst io.Writer, src io.Reader) (int64, error)
	MkdirAll(path string, perm os.FileMode) error
}

// the file which will be returned by the fs
type file interface {
	io.Closer
	io.ReadWriteSeeker
	Stat() (os.FileInfo, error)
}

// the wrapped for the filesystem interface which allows an fluent interface
type FileSystem interface {
	Put(src io.ReadSeeker, location, fileType string) (file, error)
	Get(key string) (file, error)
}
