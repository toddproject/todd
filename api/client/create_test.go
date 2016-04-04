/*
   Unit testing for Create (Client API)

   Copyright 2016 Matt Oswalt. Use or modification of this
   source code is governed by the license provided here:
   https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Mierdin/todd/agent/defs"
)

// TestAgents tests the ability for the Create client API call to function correctly
func TestAgents(t *testing.T) {

	agentJSON := `
[
  {
    "Uuid": "f9f1ad4e54b2b686088a2d347465e19034bb1533f403abfe44d4bf276030d30a",
    "DefaultAddr": "10.12.0.5",
    "Expires": 27000000000,
    "LocalTime": "2016-03-25T08:03:11.378211992Z",
    "Facts": {
      "Hostname": [
        "uraj"
      ]
    },
    "FactCollectors": {
      "get_addresses": "ff83697f546df213963f73e5b6af0fb462ad1ada96ca5c9760129750fed34b0b",
      "get_hostname": "ff1270753854f8f51a1d9233ab37a58e01280d3acf9e7a571ab65816688cf73b"
    },
    "Testlets": {
      "iperf": "0cd8877ed71dda64bc89e420ec0998c664fc274878dfedc6446aa34f31544abb",
      "ping": "992e1570967052660f944d4b06d298f45556dbd010ec10d01880eedfe65d41a7"
    }
  }
]
`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, agentJSON)
	}))
	defer ts.Close()

	// userJson := `{"username": "dennis", "balance": 200}`
	// reader = strings.NewReader(userJson) //Convert string to reader
	// request, err := http.NewRequest("GET", agentsUrl, reader) //Create request with JSON body

	agentsUrl := fmt.Sprintf("%s/v1/agent", ts.URL)

	var capi ClientApi
	err, agents := capi.Agents(
		map[string]string{
			"host": strings.Split(strings.Replace(agentsUrl, "http://", "", 1), ":")[0],
			"port": strings.Split(strings.Replace(agentsUrl, "http://", "", 1), ":")[1],
		}, "",
	)
	if err != nil {
		t.Error(err)
	}

	if len(agents) != 1 {
		t.Error("Incorrect number of agents found")
	}
}

// This was derived from the godoc on httptest (server example)
func TestDisplayAgents(t *testing.T) {
	var capi ClientApi
	err := capi.DisplayAgents([]defs.AgentAdvert{}, false)
	if err != nil {
		t.Error(err)
	}
}
