package gofile

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockReader struct {
	N int
}

func (m *MockReader) Read(p []byte) (n int, err error) {
	m.N = len(p)

	return m.N, nil
}

func (s *MockReader) Seek(offset int64, whence int) (int64, error) {
	return int64(1), nil
}

func TestOsFileSystemCopiesBytesToCreatedFileWithRelativePath(t *testing.T) {
	path := "sys/test.png"
	src := new(MockReader)

	corefs := new(MockCoreFs)
	mockFile := new(MockFile)

	fs := OSFileSystem{
		corefs,
	}

	corefs.On("MkdirAll", "./sys/", os.FileMode(0755)).Return(nil)
	corefs.On("Create", "./"+path).Return(mockFile, nil)
	corefs.On("Copy", mockFile, src).Return(int64(6), nil)

	file, err := fs.Put(src, path)

	assert.Equal(t, mockFile, file)
	assert.Nil(t, err)
}

func TestOsFileSystemCopiesBytesToCreatedFileWithAbsPath(t *testing.T) {
	path := "/sys/test.png"
	src := new(MockReader)

	corefs := new(MockCoreFs)
	mockFile := new(MockFile)

	fs := OSFileSystem{
		corefs,
	}

	corefs.On("Create", path).Return(mockFile, nil)
	corefs.On("MkdirAll", "/sys/", os.FileMode(0755)).Return(nil)
	corefs.On("Copy", mockFile, src).Return(int64(6), nil)

	file, err := fs.Put(src, path)

	assert.Equal(t, mockFile, file)
	assert.Nil(t, err)
}

func TestOsFileSystemCreateErrorPassedBack(t *testing.T) {
	path := "sys/test.png"
	src := new(MockReader)
	e := errors.New("err creating file")

	corefs := new(MockCoreFs)
	mockFile := new(MockFile)

	fs := OSFileSystem{
		corefs,
	}

	corefs.On("MkdirAll", "./sys/", os.FileMode(0755)).Return(nil)
	corefs.On("Create", "./"+path).Return(mockFile, e)

	corefs.AssertNotCalled(t, "Copy")

	file, err := fs.Put(src, path)

	assert.Equal(t, mockFile, file)
	assert.Equal(t, e, err)
}

func TestOsFileSystemCopyErrorPassedBack(t *testing.T) {
	path := "sys/test.png"
	src := new(MockReader)
	e := errors.New("err copying file")

	corefs := new(MockCoreFs)
	mockFile := new(MockFile)

	fs := OSFileSystem{
		corefs,
	}

	corefs.On("MkdirAll", "./sys/", os.FileMode(0755)).Return(nil)
	corefs.On("Create", "./"+path).Return(mockFile, nil)
	corefs.On("Copy", mockFile, src).Return(int64(0), e)

	file, err := fs.Put(src, path)

	assert.Equal(t, mockFile, file)
	assert.Equal(t, e, err)
}

func TestOsFileSystemPutReturnsErrorPathIncorrect(t *testing.T) {
	path := "\\]]]]sys  dfsdf-21321321;3;"

	src := new(MockReader)
	corefs := new(MockCoreFs)

	fs := OSFileSystem{
		corefs,
	}
	_, err := fs.Put(src, path)
	assert.Equal(t, errorIncorrectPath, err)
}
