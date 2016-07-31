package testlets

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
	//"sync/atomic"

	log "github.com/Sirupsen/logrus"
)

var (
	testletsMu sync.RWMutex
	testlets   = make(map[string]Testlet)
	done       = make(chan error) // Used from within the goroutine to inform the infrastructure it has finished
	kill       = make(chan bool)  // Used from outside the goroutine to inform the goroutine to stop
)

// Testlet defines what a testlet should look like if built in native
// go and compiled with the agent
type Testlet interface {

	// Run is the "workflow" function for a testlet. It handles running
	// the RunTestlet function asynchronously and managing the state therein.
	//
	// Params are
	// target (string)
	// args ([]string)
	// timeLimit (int in seconds)
	//
	// Returns:
	// metrics (map[string]interface{})
	// (name of metric is key, value is metric value)
	//
	// Keep as much logic out of here as possible. All native testlets
	// must support a "Kill" method, so it's best to implement core testlet
	// logic in a separate function so that the Run and Kill commands can manage
	// execution of that logic in a goroutine
	Run(string, []string, int) (map[string]string, error)

	// RunTestlet is designed to be the one-stop shop for testlet logic.
	// The developer of a native testlet just needs to implement the testlet logic here,
	// without worrying about things like managing goroutines or channels. That's all
	// managed by the "Run" or "Kill" functions
	RunTestlet(string, []string, chan bool) (map[string]string, error)
	// TODO(mierdin): is this really the best name for it? Maybe something that's less confusing, less like "Run"

	// All testlets must be able to stop operation when sent a Kill command.
	Kill() error
}

type rtfunc func(target string, args []string, kill chan bool) (map[string]string, error)

type BaseTestlet struct {

	// rtfunc is a type that will store our RunTestlet function. It is the responsibility
	// of the "child" testlet to set this value upon creation
	RunFunction rtfunc
}

// Run takes care of running the testlet function and managing it's operation given the parameters provided
func (b BaseTestlet) Run(target string, args []string, timeLimit int) (map[string]string, error) {

	var metrics map[string]string

	// TODO(mierdin): ensure channel is nil
	// done = make(chan error)
	// kill = make(chan bool)

	// TODO(mierdin): Based on experimentation, this will keep running even if this function returns.
	// Need to be sure about how this ends. Also might want to evaluate the same for the existing non-native model, likely has the same issue
	go func() {
		theseMetrics, err := b.RunFunction(target, args, kill)
		metrics = theseMetrics //TODO(mierdin): Gross.
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
			return map[string]string{}, errors.New("testlet error") // TODO(mierdin): elaborate?
		} else {
			log.Debugf("Testlet <TESTLET> completed without error")
			return metrics, nil
		}
	}
}

func (b BaseTestlet) Kill() error {
	// TODO (mierdin): This will have to be coordinated with the task above. Basically
	// you need a way to kill this testlet (and that's really only possible when running
	// async)

	// Probably just want to set the channel  to something so the select within "Run" will execute

	return nil
}

// IsNativeTestlet polls the list of registered native testlets, and returns
// true if the referenced name exists
func IsNativeTestlet(name string) bool {
	if _, ok := testlets[name]; ok {
		return true
	} else {
		return false
	}
}

//NewTestlet produces a new testlet based on the "name" param
func NewTestlet(name string) (Testlet, error) {

	if testlet, ok := testlets[name]; ok {

		// testlet.runFunction = testlet.run

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
