/*
   Unit testing for ToDD Client API - agents.go

   Copyright 2016 Matt Oswalt. Use or modification of this
   source code is governed by the license provided here:
   https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/toddproject/todd/agent/defs"
)

var testAgent = defs.AgentAdvert{
	UUID:        "aaaaaaaaaaaaaaaaaaaaaa",
	DefaultAddr: "192.168.0.1",
	Expires:     27000000000,
	LocalTime:   time.Now(),
	Facts: map[string][]string{
		"Hostname": []string{
			"testmachine",
		},
	},
	FactCollectors: map[string]string{
		"get_addresses": "aaaaaaaaaaaaaaaaaaaaaa",
		"get_hostname":  "aaaaaaaaaaaaaaaaaaaaaa",
	},
	Testlets: map[string]string{
		"iperf": "aaaaaaaaaaaaaaaaaaaaaa",
		"ping":  "aaaaaaaaaaaaaaaaaaaaaa",
	},
}

// TestAgents tests the ability for the Agents client API call to function correctly
func TestAgents(t *testing.T) {

	testAgentSlice := []defs.AgentAdvert{testAgent}

	agentJSON, err := json.Marshal(testAgentSlice)
	if err != nil {
		t.Error(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(agentJSON)
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Error(err)
	}

	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		t.Error(err)
	}

	var capi ClientAPI
	agents, err := capi.Agents(
		map[string]string{
			"host": host,
			"port": port,
		}, "",
	)
	if err != nil {
		t.Error(err)
	}

	if len(agents) != 1 {
		t.Error("Incorrect number of agents found")
	}
}

// agentTests is a "table" of test cases to apply to TestDisplayAgents
var agentTests = []struct {
	arg1 []defs.AgentAdvert
	arg2 bool
}{
	{[]defs.AgentAdvert{testAgent}, false},
	{[]defs.AgentAdvert{testAgent}, true},
	{[]defs.AgentAdvert{}, false},
}

// TestDisplayAgents iterates over the test cases and runs DisplayAgents on each
func TestDisplayAgents(t *testing.T) {
	var capi ClientAPI
	for _, test := range agentTests {
		err := capi.DisplayAgents(test.arg1, test.arg2)
		if err != nil {
			t.Error(err)
		}
	}
}
