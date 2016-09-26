/*
   ToDD testing package

   Contains infrastructure running testlets as well as maintaining
   conformance for other native-Go testlet projects

   Copyright 2016 Matt Oswalt. Use or modification of this
   source code is governed by the license provided here:
   https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package testing

import (
	"errors"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
)

var (

	// This map provides name redirection so that the native testlets can use names that don't
	// conflict with existing system tools (i.e. using "toddping" instead of "ping") but users
	// can still refer to the testlets using simple names.
	//
	// In short, users refer to the testlet by <key> and this map will redirect to the
	// actual binary name <value>
	nativeTestlets = map[string]string{
		"ping": "toddping",
	}
)

// Testlet defines what a testlet should look like if built in native
// go and compiled with the agent
type Testlet interface {

	// Run is the "workflow" function for a testlet. All testing takes place here
	// (or in a function called within)
	Run(target string, args []string, timeLimit int) (metrics map[string]string, err error)
}

// GetTestletPath generates whatever path is needed to reach the given testlet
// It first determines if the referenced testlet is native or not - if it is native,
// then only the name needs to be returned (all native testlets must be in the path).
// If it is a custom testlet, then it will generate the full path to the testlet,
// ensure it is a valid path, and if so, return that full path back to the caller.
func GetTestletPath(testletName, optDir string) (string, error) {

	if _, ok := nativeTestlets[testletName]; ok {
		log.Debugf("%s is a native testlet", testletName)
		return nativeTestlets[testletName], nil
	}

	log.Debugf("%s is a custom testlet", testletName)

	// Generate path to testlet and make sure it exists.
	testletPath := fmt.Sprintf("%s/assets/testlets/%s", optDir, testletName)
	if _, err := os.Stat(testletPath); err != nil {
		log.Errorf("Problem accessing testlet %q  on this agent", testletName)
		return "", errors.New("Error installing testrun - problem accessing testrun on agent")
	}

	return testletPath, nil

}
