package gofile

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

var errorIncorrectPath = errors.New("The path given was provided in the incorrect format")

// CoreFs interface defines a wrapper around core filesystem so that it can be extended and mocked
type CoreFs interface {
	Open(name string) (File, error)
	Create(name string) (File, error)
	Stat(name string) (os.FileInfo, error)
	Copy(dst io.Writer, src io.Reader) (int64, error)
	MkdirAll(path string, perm os.FileMode) error
}

// osFS implements coreFs using the local disk.
type osFS struct{}

// Open calls the default os.Open
func (osFS) Open(name string) (File, error) { return os.Open(name) }

// Create calls the default os.Create
func (osFS) Create(name string) (File, error) { return os.Create(name) }

// Stat calls the default os.Stat
func (osFS) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }

// Copy calls io.Copy
func (osFS) Copy(dst io.Writer, src io.Reader) (int64, error) { return io.Copy(dst, src) }

// MkdirAll calls the default os.MkdirAll
func (osFS) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }

// OSFileSystem implements the FileSystem interface by calling the CoreFs
type OSFileSystem struct {
	os CoreFs
}

// NewOSFileSystem is a construct function that returns a pointer to a OSFileSystem
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{
		&osFS{},
	}
}

// Put creates a file with the given location, creating the directories as needed
func (fs *OSFileSystem) Put(src io.ReadSeeker, path string) (File, error) {
	path = SanitizePath(path)
	r := regexp.MustCompile("(.+\\/)*(.+)\\.(.+)$")
	t := regexp.MustCompile("^\\/")

	// if we don't have a leading slash then we need to add a dot in order to
	// faciliate relative path creation
	if !t.MatchString(path) {
		path = "." + string(filepath.Separator) + path
	}

	matches := r.FindStringSubmatch(path)

	if len(matches) < 1 {
		return new(os.File), errorIncorrectPath
	}

	fs.os.MkdirAll(matches[1], 0755)

	file, err := fs.os.Create(path)
	if err != nil {
		return file, err
	}

	_, err = fs.os.Copy(file, src)
	if err != nil {
		return file, err
	}

	return file, nil
}

// Get returns a file from the core os
func (fs *OSFileSystem) Get(key string) (File, error) {
	return fs.os.Open(key)
}
