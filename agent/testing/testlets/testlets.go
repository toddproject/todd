package testlets

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	//"sync/atomic"
)

var (
	testletsMu sync.RWMutex
	testlets   = make(map[string]Testlet)
)

// Testlet defines what a testlet should look like if built in native
// go and compiled with the agent
type Testlet interface {

	// Params are
	// target (string)
	// args ([]string)
	//
	// Returns:
	// metrics (map[string]interface{})
	// (name of metric is key, value is metric value)
	Run(string, []string) (map[string]string, error)

	Kill() error

	Test() string //TODO(mierdin): Remove me
}

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

func unregisterAllDrivers() {
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
