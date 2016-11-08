package gofile

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// The S3 filesystem provides a consistent interface around the aws golang sdk
// again the s3 filesystem supports just two methods,
// Put and Get which return File interfaces
type S3FileSystem struct {
	bucket string
	config *aws.Config
	caller S3Caller
	time   Time
}

// NewS3FileSystem is a construct function takes both the region, bucket, and credential provider of your s3 filesystem
// the region and bucket parameters are self explanatory and represent configuration that you can find in you aws dashboard
// the final argument, the aws provider, this is a struct which is in charge of getting your aws credentials
// it is recommended to use the aws.EnvProvider with the filesystem
func NewS3FileSystem(region, bucket string, provider credentials.Provider) *S3FileSystem {
	return &S3FileSystem{
		bucket,
		&aws.Config{
			Region:      aws.String(region),
			Credentials: credentials.NewCredentials(provider),
		},
		new(S3Call),
		new(OSTime),
	}
}

// Get finds and return a File using a specific s3 key
func (fs *S3FileSystem) Get(path string) (File, error) {
	svc := fs.caller.NewSvc(fs.config)

	params := &s3.GetObjectInput{
		Bucket: aws.String(fs.bucket),
		Key:    aws.String(path),
	}

	resp, err := svc.GetObject(params)

	if err != nil {
		return &S3File{}, err
	}

	r, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return NewS3File(r, fs.FileUrl(path), resp.LastModified, fs), nil
}

// Put uploads a a readers contents to a specific s3 key
// the function is in charge of starting a session and sending the a structured request to the api
// it returns a File interface from the response which can be used to get information about the upload
func (fs *S3FileSystem) Put(src io.ReadSeeker, path string) (File, error) {
	svc := fs.caller.NewSvc(fs.config)

	path = SanitizePath(path)
	mimeType := GetMIMETypeFromPath(path)

	content, _ := ioutil.ReadAll(src)
	params := &s3.PutObjectInput{
		Bucket:        aws.String(fs.bucket),
		Key:           aws.String(path),
		Body:          bytes.NewReader(content),
		ContentLength: aws.Int64(int64(len(content))),
		ContentType:   aws.String(mimeType),
	}

	_, err := svc.PutObject(params)
	if err != nil {
		return new(S3File), err
	}

	now := fs.time.Now()
	return NewS3File(content, fs.FileUrl(path), &now, fs), nil
}

// FileUrl takes a path and formats its to a url to the corresponding file
func (fs *S3FileSystem) FileUrl(path string) string {
	return "https://s3-" + *fs.config.Region + ".amazonaws.com/" + fs.bucket + "/" + path
}

// S3Caller interface defines a wrapper around s3 interactions allowing calls can be safely mocked
type S3Caller interface {
	PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error)
	GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
	NewSvc(cfgs ...*aws.Config) S3Caller
}

// S3Call strcut implements the S3Caller interface by delegating calls to the svc pointer
type S3Call struct {
	svc *s3.S3
}

// NewSvc sets the aws session instance so that calls can be made to the service
func (s *S3Call) NewSvc(cfgs ...*aws.Config) S3Caller {
	sess := session.New(cfgs...)
	s.svc = s3.New(sess)

	return s
}

// PutObject uploads object to s3 using an PutObjectInput struct
func (s *S3Call) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return s.svc.PutObject(input)
}

// GetObject from the s3 api using an GetObjectInput struct
func (s *S3Call) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return s.svc.GetObject(input)
}

// S3File conforms to the File interface defining all of the generic file handling
type S3File struct {
	r    io.ReadSeeker
	info *S3FileInfo
	fs   *S3FileSystem
}

// NewS3File is a contruct function to generate a s3 file pointer
// It takes the contents that the file holds aswell and the path to its location on s3
// as well as mod which represents the time modified, the final argument is the s3 filesystem
// that the File was created under, this is used for functions such as Create
func NewS3File(contents []byte, path string, mod *time.Time, fs *S3FileSystem) *S3File {
	return &S3File{
		bytes.NewReader(contents),
		&S3FileInfo{
			path,
			contents,
			mod,
		},
		fs,
	}
}

// Close method has no functionality but is used to conform to the interface
func (s *S3File) Close() error {
	return nil
}

// Stat returns the file info of the s3 file
func (s *S3File) Stat() (os.FileInfo, error) {
	return s.info, nil
}

// Read defines how the file should be read, delegates to the implanted readseeker in the struct
func (s *S3File) Read(p []byte) (n int, err error) {
	return s.r.Read(p)
}

// Seek delegates to the implanted readseeker in the struct
func (s *S3File) Seek(offset int64, whence int) (int64, error) {
	return s.r.Seek(offset, whence)
}

// Write writes bytes to the path location by calling the implanted filesystem
func (s *S3File) Write(p []byte) (n int, err error) {
	read := len(p)

	info, _ := s.Stat()
	_, err = s.fs.Put(bytes.NewReader(p), info.Name())
	return read, err
}

// S3FileInfo is A struct which conforms to the file interface which provides information about the s3 file
type S3FileInfo struct {
	path    string
	content []byte
	mod     *time.Time
}

// Name gets the base path of the file
func (s *S3FileInfo) Name() string {
	return s.path
}

// Size returns the length in bytes of the file
func (s *S3FileInfo) Size() int64 {
	return int64(len(s.content))
}

// IsDir returns false as s3 file is assumed not to be a directory
func (s *S3FileInfo) IsDir() bool {
	return false
}

// Mode returns a os.ModePerm as the file is assumed perm
func (s *S3FileInfo) Mode() os.FileMode {
	return os.ModePerm
}

// Sys underlying data source which should return nil
func (s *S3FileInfo) Sys() interface{} {
	return nil
}

// ModTime returns modification time
func (s *S3FileInfo) ModTime() time.Time {
	return *s.mod
}
