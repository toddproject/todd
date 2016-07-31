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
// a temporary holding place, as the main effort around native testlets
// is being implemented so they can be broken out into their own repos.

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/Mierdin/todd/agent/testing/testlets"
)

type PingTestlet struct {
	testlets.BaseTestlet
}

func init() {

	var pt = PingTestlet{}

	// Ensure the RunFunction attribute is set correctly.
	// This allows the underlying testlet infrastructure
	// to know what function to call at runtime
	pt.RunFunction = pt.RunTestlet

	// This is important - register the name of this testlet
	// (the name the user will use in a testrun definition)
	testlets.Register("ping", &pt)
}

// RunTestlet implements the core logic of the testlet. Don't worry about running asynchronously,
// that's handled by the infrastructure.
func (p PingTestlet) RunTestlet(target string, args []string, kill chan (bool)) (map[string]string, error) {

	// Get number of pings
	count := 3 //TODO(mierdin): need to parse from 'args', or if omitted, use a default value

	log.Error(args)

	var latencies []float32
	var replies int

	// Execute ping once per count
	i := 0
	for i < count {
		select {
		case <-kill:
			// Terminating early; return empty metrics
			return map[string]string{}, nil
		default:

			//log.Debugf("Executing ping #%d", i)

			// Mocked ping logic
			latency, replyReceived := pingTemp(count)

			latencies = append(latencies, latency)

			if replyReceived {
				replies += 1
			}

			i += 1
			time.Sleep(1000 * time.Millisecond)

		}
	}

	// Calculate metrics
	var latencyTotal float32 = 0
	for _, value := range latencies {
		latencyTotal += value
	}
	avg_latency_ms := latencyTotal / float32(len(latencies))
	packet_loss := (float32(count) - float32(replies)) / float32(count)

	return map[string]string{
		"avg_latency_ms": fmt.Sprintf("%.2f", avg_latency_ms),
		"packet_loss":    fmt.Sprintf("%.2f", packet_loss),
	}, nil

}

func pingTemp(count int) (float32, bool) {
	return float32(count) * 4.234, true
}
