#Gofile

Gofile provides a consistent and simple interface to deal with differing file systems. It provides great flexibility and allows you to easily mock out and unit test file interactions.

Currently Gofile supports interactions with the core *OS* and *Amazon S3* file systems, with the *Google Cloud* application in development.

### Installation

Install Gofile via `go get`

```
go get github.com/hugorut/gofile
```

### Usage

The interface across the different filesystems is as follows:

```go
type FileSystem interface {
    // Put a file into a filesystem and return a File interface which
    // will give you information about the location and status of
    // the uploaded file the path must contain the file and full extension
    //
    // e.g. /path/to/my/funky/file.gif
    Put(src io.ReadSeeker, path string) (File, error)

    // get a file from the file system, return the File interface
    // so that all generic file interactions can be facilitated
    Get(path string) (File, error)
}
```

with each method returning a `gofile.File` interface:

```go
type File interface {
    io.Closer
    io.ReadWriteSeeker
    Stat() (os.FileInfo, error)
}
```

Getting up an running with Gofile is easy. The package provides simple construction functions which give you access to the different file sytem implementations. Here's a quick tour of the usage of in the different filesystems.

#### S3 File system

**Put**
```go
reader := bytes.NewReader([]byte("my file contents"))
region := "eu-west-1"
bucket := "my-trusty-bucket"

filesys := gofile.NewS3FileSystem(region, bucket, &aws.EnvProvider{})
file, err := filesys.Put(reader, "my/path/to-file.txt")
```

**Get**
```go
region := "eu-west-1"
bucket := "my-trusty-bucket"

filesys := gofile.NewS3FileSystem(region, bucket, &aws.EnvProvider{})
file, err := filesys.Get("my/path/to-file.txt")
```

#### OS File system

**Put**
```go
reader := bytes.NewReader([]byte("my file contents"))

filesys := gofile.NewOSFileSystem()
file, err := filesys.Put(reader, "my/path/to-file.txt")
```

**Get**
```go
filesys := gofile.NewOSFileSystem()
file, err := filesys.Get("my/path/to-file.txt")
```

###Examples

Here's a quick example of how easy it is to work with gofile. The example below outlines a implementation of a http handler which uploads an image to s3 from a request which holds a base64 encoded image.

```go
package main

import (
    "encoding/json"
    "net/http"

    "github.com/hugorut/gofile"
)

var region = "eu-west-1"
var bucket = "my-trusty-bucket"

//Create the filesystem with the package constrcut function, we'll use environmnet variables
//To set the configuration i.e. AWS api secrets
var filesys = gofile.NewS3FileSystem(region, bucket, &aws.EnvProvider{})

// The request struct to marshal the json into
type ImageStoreRequest struct {
    Image []byte "json:image"
}

//a Http Handler which performs an image upload from a request 
// which holds an base63 encoded image
func UploadImage(w http.ResponseWriter, r *http.Request) {
    var imageRequest ImageStoreRequest
    err := json.NewDecoder(r.Body).Decode(&imageRequest)

    if err != nil {
        errorResponse(w, err)
    }
    
    //Lets get the extension of the file through a helper function
    //and then call the Put function in order to upload the file
    ext := gofile.Base64ImageType(imageRequest.Image)
    file, err := filesys.Put(
        gofile.Base64ToDecoder(imageRequest.Image),
        "my/path/to/thefile."+ext,
    )

    if err != nil {
        errorResponse(w, err)
    }
    
    // lets now get the file information from the uploaded file
    // and pass that back to the client
    info, _ := file.Stat()
    json.NewEncoder(w).Encode(map[string]string{"path": info.Name()})
}

// handle the error and make a http response for the purpose of demo and simpility
// we'll just spit the error back to the client in a map
func errorResponse(w http.ResponseWriter, err error) {
    w.WriteHeader(400)
    json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
```

###Testing

Gofile was built with testing in mind and provides a `MockFileSystem` to facilitate ease of testing. Access to the mock filesystem is given through the conventional construct function

```go
mockFilesystem := gofile.NewMockFilesystem()
```

This returns an implementation of the `Filesystem` interface which is a `testify/mock.Mock` and faciliates method expectation such as below:

```go
    file := new(sys.MockFile)
    info := new(sys.MockFileInfo)

    reader := sys.Base64ToDecoder([]byte(im))
    mockFilesystem.On("Put", reader, path).Return(file, nil)
```

To read more about testifies mock library see [here](https://godoc.org/github.com/stretchr/testify/mock)