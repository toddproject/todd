/*
   testrun definition

    Copyright 2016 Matt Oswalt. Use or modification of this
    source code is governed by the license provided here:
    https://github.com/toddproject/todd/blob/master/LICENSE
*/

package defs

// TestRun is a struct  for a testrun command to be sent to the agent. This is not to be confused with the testrun object.
// Here, the "targets" property is a string slice - which means that the target type and targets have already been calculated by the server.
// This struct is sent to a specific agent so that it has the instructions it needs to perform a test.
type TestRun struct {
	Uuid    string   `json:"uuid"`
	Targets []string `json:"targets"`
	Testlet string   `json:"testlet"`
	Args    string   `json:"args"`
}
