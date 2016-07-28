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
	"time"

	"github.com/Mierdin/todd/agent/testing/testlets"
)

type PingTestlet struct{}

func init() {

	// This is important - register the name of this testlet
	// (the name the user will use in a testrun definition)
	testlets.Register("ping", &PingTestlet{})
}

func (p PingTestlet) Run(target string, args []string) (map[string]string, error) {

	// TODO(mierdin): Implement ping test here - this is just a mock
	time.Sleep(3000 * time.Millisecond)
	return map[string]string{
		"avg_latency_ms":         "25.144",
		"packet_loss_percentage": "0",
	}, nil

}

func (p PingTestlet) Kill() error {
	// TODO (mierdin): This will have to be coordinated with the task above. Basically
	// you need a way to kill this testlet (and that's really only possible when running
	// async)

	return nil
}
