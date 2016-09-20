/*
   ToDD task - set keyvalue pair in cache

   Copyright 2016 Matt Oswalt. Use or modification of this
   source code is governed by the license provided here:
   https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package testlets

import (
	"sync"
	// log "github.com/Sirupsen/logrus"
)

var (
	testletsMu sync.RWMutex
	testlets   = make(map[string]Testlet)
	done       = make(chan error) // Used from within the goroutine to inform the infrastructure it has finished
	kill       = make(chan bool)  // Used from outside the goroutine to inform the goroutine to stop

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

// NOTE
//
// Early efforts to build native-Go testlets involved the embedding of testlet logic into the
// ToDD agent itself. As a result, it was important to build some reusable infrastructure so that goroutines
// running testlet code within the agent could be controlled, and that new testlets could benefit from this
// infrastructure.
//
// Since then, the decision was made to keep testlets as their own separate binaries.
//
// These testlets are in their own repositories, and they do actually use some of the logic below, just not as meaningfully
// and comprehensively as they would have if they were baked in to the agent.  The development standard for all "blessed"
// testlets will still ensure that they use this interface, so that if we decide to bake them into the agent in the future,
// they'll already conform.
//
// (The vast majority of this code was inspired by the database drivers implementation in the stdlib)

type rtfunc func(target string, args []string, timeout int) (map[string]string, error)

type BaseTestlet struct {

	// rtfunc is a type that will store our RunTestlet function. It is the responsibility
	// of the "child" testlet to set this value upon creation
	RunFunction rtfunc
}

// IsNativeTestlet polls the list of registered native testlets, and returns
// true if the referenced name exists
func IsNativeTestlet(name string) (bool, string) {
	if _, ok := nativeTestlets[name]; ok {
		return true, nativeTestlets[name]
	} else {
		return false, ""
	}
}
