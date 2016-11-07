package gofile

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
)

// the core filesystem which can be extended and mocked
type CoreFs interface {
	Open(name string) (File, error)
	Create(name string) (File, error)
	Stat(name string) (os.FileInfo, error)
	Copy(dst io.Writer, src io.Reader) (int64, error)
	MkdirAll(path string, perm os.FileMode) error
}

// osFS implements coreFs using the local disk.
type osFS struct{}

// wrapper functions to call the underlying core os and io utilities
func (osFS) Open(name string) (File, error)                   { return os.Open(name) }
func (osFS) Create(name string) (File, error)                 { return os.Create(name) }
func (osFS) Stat(name string) (os.FileInfo, error)            { return os.Stat(name) }
func (osFS) Copy(dst io.Writer, src io.Reader) (int64, error) { return io.Copy(dst, src) }
func (osFS) MkdirAll(path string, perm os.FileMode) error     { return os.MkdirAll(path, perm) }

// filesystem which wraps the native os system
type OSFileSystem struct {
	os CoreFs
}

func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{
		&osFS{},
	}
}

// create a file with the given location and type and then copy
func (fs *OSFileSystem) Put(src io.ReadSeeker, path, extension string) (File, error) {
	path = SanitizePath(path)
	r := regexp.MustCompile("\\/.+$")

	location := "." + string(filepath.Separator) + r.ReplaceAllString(path, "")
	fs.os.MkdirAll(location, 0755)

	file, err := fs.os.Create(joinPath(path, extension))
	if err != nil {
		return file, err
	}

	_, err = fs.os.Copy(file, src)
	if err != nil {
		return file, err
	}

	return file, nil
}

// get a file from the core os
func (fs *OSFileSystem) Get(key string) (File, error) {
	return fs.os.Open(key)
}
