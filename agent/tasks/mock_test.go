/*
   This file contains mock infrastructure for the tasks package

   Copyright 2016 Matt Oswalt. Use or modification of this
   source code is governed by the license provided here:
   https://github.com/toddproject/todd/blob/master/LICENSE
*/

package tasks

import (
	"io"
	"os"
	"time"
)

// mockFS implements fileSystem using mocked functions
type mockFS struct{}

func (mockFS) Open(name string) (file, error)            { return &os.File{}, nil }
func (mockFS) Stat(name string) (os.FileInfo, error)     { return fileinf{}, nil }
func (mockFS) Create(name string) (file, error)          { return &os.File{}, nil }
func (mockFS) Chmod(name string, mode os.FileMode) error { return nil }

// mockIoSys implements IoSystem using mocked functions.
type mockIoSys struct{}

func (mockIoSys) Copy(dst io.Writer, src io.Reader) (int64, error) { return 1, nil }

// fileinf is a struct to implement os.FileInfo for test purposes.
type fileinf struct{}

func (fileinf) Name() string       { return "test_name" }
func (fileinf) Size() int64        { return 37628 }
func (fileinf) Mode() os.FileMode  { return 0777 }
func (fileinf) ModTime() time.Time { return time.Now() }
func (fileinf) IsDir() bool        { return false }
func (fileinf) Sys() interface{}   { return "whatever" }
