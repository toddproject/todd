package toddping

// NOTE //////////////////
//
// This is a built-in testlet. Currently, the approach is to have each
// testlet under it's own package, which is explicitly imported under
// the ToDD agent's 'main' package.
//
// Currently, each testlet is stored within the ToDD repo in order to
// vet out the architecture, which means they are awkwardly placed
// underneath the "testlets" directory together. However, this should be
// a tempoarary holding place, as the main effort around native testlets
// is being implemented so they can be broken out into their own repos.

import (
	"github.com/Mierdin/todd/agent/testing/testlets"
)

type PingTestlet struct{}

func init() {
	testlets.Register("ping", &PingTestlet{})
}

// TODO(mierdin): Maybe consider running these asyc by default? Basically
// the "Run" function kicks back a channel of type map[string]string so when
// it's populated, it contains the metrics and you know it can stop

func (p PingTestlet) Run(target string, args []string) (map[string]string, error) {

	// TODO(mierdin): Implement ping test here

	return map[string]string{}, nil
}

func (p PingTestlet) Kill() error {
	// TODO (mierdin): This will have to be coordinated with the task above. Basically
	// you need a way to kill this testlet (and that's really only possible when running
	// async)

	return nil
}

func (p PingTestlet) Test() string {
	return "trolololol"
}
