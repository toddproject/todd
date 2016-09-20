/*
   ToDD task - set keyvalue pair in cache

   Copyright 2016 Matt Oswalt. Use or modification of this
   source code is governed by the license provided here:
   https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package testing

import (
	"sync"
	// log "github.com/Sirupsen/logrus"
)

// NOTE
//
// Early efforts to build native-Go testlets involved the embedding of testlet logic into the
// ToDD agent itself. As a result, it was important to build some reusable infrastructure so that goroutines
// running testlet code inside the agent could be controlled, and that new testlets could benefit from this
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

<<<<<<< 287aa03c127fa5a60db0684c4e4340d6cec19f62:agent/testing/testlets.go
// Run takes care of running the testlet function and managing it's operation given the parameters provided
func (b BaseTestlet) Run(target string, args []string, timeLimit int) (map[string]string, error) {

	var metrics map[string]string

	// Ensure control channels are empty
	done := make(chan error)
	kill := make(chan bool)

	go func() {
		metrics, err := b.RunFunction(target, args, kill)
<<<<<<< 5056f30dcb3a2cf7a20e251b620c8220217f2f4b:agent/testing/testlets.go
=======

		// TODO(mierdin): avoiding a "declared and not used" error for now
		// If this code is ever actually used, it should be modified to make "done" a channel that returns the metrics, so it's actually used (just an idea)
		log.Error(metrics)

>>>>>>> Updated testlets.go:agent/testing/testlets/testlets.go
		done <- err
	}()

	// This select statement will block until one of these two conditions are met:
	// - The testlet finishes, in which case the channel "done" will be receive a value
	// - The configured time limit is exceeded (expected for testlets running in server mode)
	select {
	case <-time.After(time.Duration(timeLimit) * time.Second):
		log.Debug("Successfully killed <TESTLET>")
		return map[string]string{}, nil

	case err := <-done:
		if err != nil {
			return map[string]string{}, errors.New("testlet error")
		} else {
			log.Debugf("Testlet <TESTLET> completed without error")
			return metrics, nil
		}
	}
}

// Kill is currently unimplemented. This will have to be coordinated with "Run". Basically
// you need a way to kill this testlet (and that's really only possible when running
// async) Probably just want to set the channel to something so the select within "Run" will execute
func (b BaseTestlet) Kill() error {
	return nil
}

=======
>>>>>>> removed all old work from testlets (moved to archive):agent/testing/testlets/testlets.go
// IsNativeTestlet polls the list of registered native testlets, and returns
// true if the referenced name exists
func IsNativeTestlet(name string) (bool, string) {
	if _, ok := nativeTestlets[name]; ok {
		return true, nativeTestlets[name]
	} else {
		return false, ""
	}
}
<<<<<<< 287aa03c127fa5a60db0684c4e4340d6cec19f62:agent/testing/testlets.go

//NewTestlet produces a new testlet based on the "name" param
func NewTestlet(name string) (Testlet, error) {

	if testlet, ok := testlets[name]; ok {
		return testlet, nil
	} else {
		return nil, errors.New(
			fmt.Sprintf("'%s' not currently supported as a native testlet"),
		)
	}
}

// Register makes a testlet available by the provided name.
// If Register is called twice with the same name or if testlet is nil,
// it will return an error
func Register(name string, testlet Testlet) error {
	testletsMu.Lock()
	defer testletsMu.Unlock()
	if testlet == nil {
		return errors.New("Register testlet is nil")
	}
	if _, dup := testlets[name]; dup {
		return errors.New("Register called twice for testlet " + name)
	}
	testlets[name] = testlet
	return nil
}

func unregisterAllTestlets() {
	testletsMu.Lock()
	defer testletsMu.Unlock()
	// For tests.
	testlets = make(map[string]Testlet)
}

// Testlets returns a sorted list of the names of the registered testlets.
func Testlets() []string {
	testletsMu.RLock()
	defer testletsMu.RUnlock()
	var list []string
	for name := range testlets {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}
=======
>>>>>>> removed all old work from testlets (moved to archive):agent/testing/testlets/testlets.go
