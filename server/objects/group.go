/*
   ToDD GroupObject definition

    Copyright 2016 Matt Oswalt. Use or modification of this
    source code is governed by the license provided here:
    https://github.com/toddproject/todd/blob/master/LICENSE
*/

package objects

import (
	"fmt"
)

// GroupObject is a specific implementation of BaseObject. It represents the "group" object in ToDD - which
// stores specific rules that are used when grouping agents together prior to running a test
type GroupObject struct {
	BaseObject `yaml:",inline"` // the ",inline" tag is necessary for the go-yaml package to properly see the outer struct fields
	Spec       struct {
		Group   string              `json:"group" yaml:"group"`
		Matches []map[string]string `json:"matches" yaml:"matches"`
	} `json:"spec" yaml:"spec"`
}

// GetSpec is a simple function to return the "Spec" attribute of a GroupObject
func (g GroupObject) GetSpec() string {
	return fmt.Sprint(g.Spec)
}
