package filesystem

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
// again the s3 filesystem supports just two methods, Put and Get which return
// file interfaces
type S3FileSystem struct {
	bucket string
	config *aws.Config
}

// Generate a pointer to a new s3 filesystem
// the constructor takes both the region and bucket that you wish to operate as your filesystem
// the final argument is an aws provider, this is a struct which is in charge of getting
// your aws credentials, it is recommended to use the aws.EnvProvider with the filesystem
func NewS3FileSystem(region, bucket string, provider credentials.Provider) *S3FileSystem {
	return &S3FileSystem{
		bucket,
		&aws.Config{
			Region:      aws.String(region),
			Credentials: credentials.NewCredentials(provider),
		},
	}
}

// get a file using a specific s3 key and return an instance of the file interface
func (fs *S3FileSystem) Get(key string) (file, error) {
	sess := session.New(fs.config)
	svc := s3.New(sess)

	params := &s3.GetObjectInput{
		Bucket: aws.String(fs.bucket),
		Key:    aws.String(key),
	}

	resp, err := svc.GetObject(params)

	if err != nil {
		return &S3File{}, err
	}

	r, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return NewS3File(r, fs.FileUrl(key), resp.LastModified, fs), nil
}

// put a file into a key with an s3 bucket, start a session and make a request to put an object
// returns a file interface from the response
func (fs *S3FileSystem) Put(src io.ReadSeeker, location string, fileType string) (file, error) {
	sess := session.New(fs.config)
	svc := s3.New(sess)

	location = SanitizePath(location)
	path := joinPath(location, fileType)

	content, _ := ioutil.ReadAll(src)
	params := &s3.PutObjectInput{
		Bucket:        aws.String(fs.bucket),
		Key:           aws.String(path),
		Body:          bytes.NewReader(content),
		ContentLength: aws.Int64(int64(len(content))),
		ContentType:   aws.String(fileType),
	}

	_, err := svc.PutObject(params)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return NewS3File(content, fs.FileUrl(path), &now, fs), nil
}

// return a url to the corresponding file
func (fs *S3FileSystem) FileUrl(path string) string {
	return "https://s3-" + *fs.config.Region + ".amazonaws.com/" + fs.bucket + path
}

// A struct which conforms to the file interface
type S3File struct {
	r    io.ReadSeeker
	info *S3FileInfo
	fs   *S3FileSystem
}

// helper function to generate a s3 file pointer by taking the
// byte representation of the the file
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

// close method has no functionality
func (s *S3File) Close() error {
	return nil
}

// return the file info
func (s *S3File) Stat() (os.FileInfo, error) {
	return s.info, nil
}

// delegate to the implanted reader to read bytes
func (s *S3File) Read(p []byte) (n int, err error) {
	return s.r.Read(p)
}

// delegate to the implanted seeker to read bytes
func (s *S3File) Seek(offset int64, whence int) (int64, error) {
	return s.r.Seek(offset, whence)
}

// write to the file location
func (s *S3File) Write(p []byte) (n int, err error) {
	read := len(p)
	// file path and name needed
	info, _ := s.Stat()
	_, err = s.fs.Put(bytes.NewReader(p), info.Name(), info.Name())
	return read, err
}

// A struct which conforms to the file interface
type S3FileInfo struct {
	path    string
	content []byte
	mod     *time.Time
}

// base name of the file
func (s *S3FileInfo) Name() string {
	return s.path
}

// length in bytes
func (s *S3FileInfo) Size() int64 {
	return int64(len(s.content))
}

// s3 file is assumed not to be a directory
func (s *S3FileInfo) IsDir() bool {
	return false
}

// file mode bits
func (s *S3FileInfo) Mode() os.FileMode {
	return os.ModePerm
}

// underlying data source which should return nil
func (s *S3FileInfo) Sys() interface{} {
	return nil
}

// modification time
func (s *S3FileInfo) ModTime() time.Time {
	return *s.mod
}
