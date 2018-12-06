/*
   ToDD API

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

// ToDDResource represents any protocol-buffer based resource.
type ToDDResource interface {
	Reset()
	String() string
	ProtoMessage()
	Descriptor() ([]byte, []int)
}
