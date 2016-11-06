#Gofile

Gofile provides a consistent and simple interface to deal with differing filesystems. It provides great flexibility and allows you to easily mock out and unit tests file interactions.

Currently Gofile supports interactions with the core *OS* and *Amazon S3* filesystems, with the *Google Cloud* application in development.

### Usage

Install Gofile via `go get`

```
go get github/hugorut/gofile.git
```

The interface across the different filesystems is as follows:

```go
type FileSystem interface {
    Put(src io.ReadSeeker, path, extension string) (file, error)
    Get(path string) (file, error)
}
```

with each method returning a `gofile.file` interface:

```go
type file interface {
    io.Closer
    io.ReadWriteSeeker
    Stat() (os.FileInfo, error)
}
```

#### OS Filesystem

**Put**
```go
reader := bytes.NewReader([]byte("my file contents"))

filesys := gofile.NewOSFileSystem()
file, err := filesys.Put(reader, "my/path/to-file", "text/txt")
```

**Get**
```go
filesys := gofile.NewOSFileSystem()
file, err := filesys.Get("my/path/to-file.txt")
```
#### S3 Filesystem

**Put**
```go
reader := bytes.NewReader([]byte("my file contents"))
region := "eu-west-1"
bucket := "my-trusty-bucket"

filesys := gofile.NewS3FileSystem(region, bucket, &aws.EnvProvider{})
file, err := filesys.Put(reader, "my/path/to-file", "text/txt")
```

**Get**
```go
region := "eu-west-1"
bucket := "my-trusty-bucket"

filesys := gofile.NewS3FileSystem(region, bucket, &aws.EnvProvider{})
file, err := filesys.Get("my/path/to-file.txt")
```