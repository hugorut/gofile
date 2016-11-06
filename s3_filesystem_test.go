package filesystem

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
)

func TestGetReturnsAnFileWithUrlOfFileAndLastModified(t *testing.T) {
	bucket := "bucket"
	region := "region"
	config := getConfig(region)
	path := "some/file.jpg"
	body := "some body"

	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	}

	fs, caller, _ := setUpS3FileSystem(bucket, config)

	now := time.Now()
	recorder := httptest.NewRecorder()
	recorder.WriteString(body)

	response := new(s3.GetObjectOutput)
	response.LastModified = &now
	response.Body = recorder.Result().Body

	caller.On("NewSvc", []*aws.Config{config}).Return(caller)
	caller.On("GetObject", params).Return(response, nil)

	file, err := fs.Get(path)
	assert.Nil(t, err)

	info, _ := file.Stat()
	assert.Equal(t, "https://s3-"+region+".amazonaws.com/"+bucket+path, info.Name())
	assert.Equal(t, now, info.ModTime())

	b, _ := ioutil.ReadAll(file)
	assert.Equal(t, []byte(body), b)
}

func TestPutCallsS3AndWrapsReponseInFile(t *testing.T) {
	bucket := "bucket"
	region := "region"
	config := getConfig(region)

	location := "some/file"
	fileType := "image/jpg"
	path := "some/file.jpg"

	content := []byte("some content")

	params := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(path),
		Body:          bytes.NewReader(content),
		ContentLength: aws.Int64(int64(len(content))),
		ContentType:   aws.String(fileType),
	}

	fs, caller, timer := setUpS3FileSystem(bucket, config)

	now := time.Now()
	timer.On("Now").Return(now)

	caller.On("NewSvc", []*aws.Config{config}).Return(caller)
	caller.On("PutObject", params).Return(nil, nil)

	file, err := fs.Put(bytes.NewReader(content), location, fileType)
	assert.Nil(t, err)

	info, _ := file.Stat()
	assert.Equal(t, "https://s3-"+region+".amazonaws.com/"+bucket+path, info.Name())
	assert.Equal(t, now, info.ModTime())

	b, _ := ioutil.ReadAll(file)
	assert.Equal(t, content, b)
}

func TestPutS3ReponseErrorReturnsZeroFileAndError(t *testing.T) {
	bucket := "bucket"
	region := "region"
	config := getConfig(region)

	location := "some/file"
	fileType := "image/jpg"
	path := "some/file.jpg"

	content := []byte("some content")

	params := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(path),
		Body:          bytes.NewReader(content),
		ContentLength: aws.Int64(int64(len(content))),
		ContentType:   aws.String(fileType),
	}

	fs, caller, timer := setUpS3FileSystem(bucket, config)

	now := time.Now()
	timer.On("Now").Return(now)

	e := errors.New("s3 problem")

	caller.On("NewSvc", []*aws.Config{config}).Return(caller)
	caller.On("PutObject", params).Return(nil, e)

	file, err := fs.Put(bytes.NewReader(content), location, fileType)
	assert.Equal(t, e, err)
	assert.Equal(t, new(S3File), file)
}

func getConfig(region string) *aws.Config {
	return &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewEnvCredentials(),
	}
}

func setUpS3FileSystem(bucket string, config *aws.Config) (*S3FileSystem, *MockS3Caller, *MockTime) {
	caller := new(MockS3Caller)
	timer := new(MockTime)

	return &S3FileSystem{
		bucket,
		config,
		caller,
		timer,
	}, caller, timer
}
