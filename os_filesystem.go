package filesystem

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
)

// osFS implements coreFs using the local disk.
type osFS struct{}

// wrapper functions to call the underlying core os and io utilities
func (osFS) Open(name string) (file, error)                   { return os.Open(name) }
func (osFS) Create(name string) (file, error)                 { return os.Create(name) }
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
func (fs *OSFileSystem) Put(src io.ReadSeeker, location, fileType string) (file, error) {
	location = SanitizePath(location)
	r := regexp.MustCompile("\\/.+$")

	path := "." + string(filepath.Separator) + r.ReplaceAllString(location, "")
	fs.os.MkdirAll(path, 0755)

	file, err := fs.os.Create(joinPath(location, fileType))
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
func (fs *OSFileSystem) Get(key string) (file, error) {
	return fs.os.Open(key)
}
