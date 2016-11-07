package gofile

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase64ToDecoder(t *testing.T) {
	im := []byte("data:image/gif;base64,R0lGODlhAQABAIAAAP///////yH5BAEKAAEALAAAAAABAAEAAAICTAEAOw==")
	decoded := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x1, 0x0, 0x1, 0x0, 0x80, 0x0, 0x0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x21, 0xf9, 0x4, 0x1, 0xa, 0x0, 0x1, 0x0, 0x2c, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x1, 0x0, 0x0, 0x2, 0x2, 0x4c, 0x1, 0x0, 0x3b}
	r := Base64ToDecoder(im)
	b, err := ioutil.ReadAll(r)

	assert.Nil(t, err)
	assert.Equal(t, decoded, b)
}

func TestBase64ImageType(t *testing.T) {
	im := []byte("data:image/gif;base64,R0lGODlhAQABAIAAAP///////yH5BAEKAAEALAAAAAABAAEAAAICTAEAOw==")
	ty := Base64ImageType(im)
	assert.Equal(t, "image/gif", ty)
}

func TestSanitizePath(t *testing.T) {
	paths := map[string]string{
		"my funky path   d": "my-funky-path-d",
		"my funky path   ":  "my-funky-path",
		"my  f  path":       "my-f-path",
	}

	for input, expected := range paths {
		actual := SanitizePath(input)
		assert.Equal(t, expected, actual)
	}
}

func TestStripBaseEncoding(t *testing.T) {
	im := []byte("data:image/gif;base64,R0lGODlhAQABAIAAAP///////yH5BAEKAAEALAAAAAABAAEAAAICTAEAOw==")
	expected := []byte("R0lGODlhAQABAIAAAP///////yH5BAEKAAEALAAAAAABAAEAAAICTAEAOw==")

	actual := StripBaseEncoding(im)
	assert.Equal(t, expected, actual)
}

func TestGetMIMETypeFromPath(t *testing.T) {
	paths := map[string]string{
		"/my/path/here/image.jpg": "image/jpeg",
		"image.jpg":               "image/jpeg",
		"path/image.jpg":          "image/jpeg",
	}

	for input, expected := range paths {
		actual := GetMIMETypeFromPath(input)
		assert.Equal(t, expected, actual, "input: "+input)
	}
}
