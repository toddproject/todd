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
	//
	// Params are
	// target (string)
	// args ([]string)
	// timeLimit (int in seconds)
	//
	// Returns:
	// metrics (map[string]string)
	// (name of metric is key, value is metric value)
	Run(string, []string, int) (map[string]string, error)
}

type rtfunc func(target string, args []string, timeout int) (map[string]string, error)

type BaseTestlet struct {

	// rtfunc is a type that will store our RunTestlet function. It is the responsibility
	// of the "child" testlet to set this value upon creation
	RunFunction rtfunc
}

// GetTestletPath generates whatever path is needed to reach the given testlet
// It first determines if the referenced testlet is native or not - if it is native,
// then only the name needs to be returned (all native testlets must be in the path).
// If it is a custom testlet, then it will generate the full path to the testlet,
// ensure it is a valid path, and if so, return that full path back to the caller.
func GetTestletPath(testletName, optDir string) (string, error) {

	if _, ok := nativeTestlets[testletName]; ok {
		log.Infof("%s is a native testlet", testletName)
		return nativeTestlets[testletName], nil
	} else {

		log.Infof("%s is a custom testlet", testletName)

		// Generate path to testlet and make sure it exists.
		testletPath := fmt.Sprintf("%s/assets/testlets/%s", optDir, testletName)
		if _, err := os.Stat(testletPath); os.IsNotExist(err) {
			log.Errorf("Testlet %s does not exist on this agent", testletName)
			return "", errors.New("Error installing testrun - testlet doesn't exist on this agent.")
		}

		return testletPath, nil
	}
}
