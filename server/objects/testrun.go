/*
    ToDD TestRun definition

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package objects

import (
	"fmt"
)

// TestRunObject is a specific implementation of BaseObject. It represents the "testrun" object in ToDD, which
// is a set of parameters under which a test should be run
type TestRunObject struct {
	BaseObject `yaml:",inline"` // the ",inline" tag is necessary for the go-yaml package to properly see the outer struct fields
	Spec       struct {
		TargetType string            `json:"targettype" yaml:"targettype"`
		Source     map[string]string `json:"source" yaml:"source"`
		Target     interface{}       `json:"target" yaml:"target"` // This is an empty interface because targettype of "group" uses this as a map, targettype of "uncontrolled" uses this as a slice.
		//App        string            `json:"app" yaml:"app"`  //TODO(mierdin): temporarily commenting out because App is defined in Source and Target now.
	} `json:"spec" yaml:"spec"`
}

// GetSpec is a simple function to return the "Spec" attribute of a TestRunObject
func (t TestRunObject) GetSpec() string {
	return fmt.Sprint(t.Spec)
}
