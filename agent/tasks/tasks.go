/*
	ToDD tasks (agent communication)

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package tasks

import (
	"io"
	"os"
)

// Task is an interface to define task behavior This is used for functions like those in comms
// that need to work with all specific task types, not one specific task.
type Task interface {

	// TODO(mierdin): This works but is a little "meh". Basically, each task has to have a "Run()" function, as enforced
	// by this interface. If the task needs some additional data, it gets these through struct properties. This works but
	// doesn't quite feel right. Come back to this and see if there's a better way.
	Run() error
}

// BaseTask is a struct that is intended to be embedded by specific task structs. Both of these in conjunction
// are used primarily to house the JSON message for passing tasks over the comms package (i.e. message queue), but may also contain important
// dependencies of the task, such as an HTTP handler.
type BaseTask struct {
	Type string `json:"type"`
}

// FileSystem is an interface to abstract the "os" calls, to properly mock out functions that work with the filesystem.
type FileSystem interface {
	Open(name string) (file, error)
	Stat(name string) (os.FileInfo, error)
	Create(name string) (file, error)
	Chmod(name string, mode os.FileMode) error
}

type file interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Writer
	Stat() (os.FileInfo, error)
}

// OsFS implements fileSystem using the local disk.
type OsFS struct{}

func (OsFS) Open(name string) (file, error)            { return os.Open(name) }
func (OsFS) Stat(name string) (os.FileInfo, error)     { return os.Stat(name) }
func (OsFS) Create(name string) (file, error)          { return os.Create(name) }
func (OsFS) Chmod(name string, mode os.FileMode) error { return os.Chmod(name, mode) }

// IoSystem is an interface to abstract the "ios" calls, to properly mock out functions that work with the filesystem.
type IoSystem interface {
	Copy(dst io.Writer, src io.Reader) (int64, error)
}

// IoSys implements IoSystem using the local disk.
type IoSys struct{}

func (IoSys) Copy(dst io.Writer, src io.Reader) (int64, error) { return io.Copy(dst, src) }
