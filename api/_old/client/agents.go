/*
    ToDD Client API Calls for "todd agents"

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"text/tabwriter"
	"text/template"

	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/hostresources"
)

// Agents will query the ToDD server for a list of currently registered agents, and will display
// a list of them to the user. Optionally, the user can provide a subargument containing the UUID of
// a registered agent, and this function will output more detailed information about that agent.
func (capi ClientAPI) Agents(conf map[string]string, agentUUID string) ([]defs.AgentAdvert, error) {

	var agents []defs.AgentAdvert

	var url string

	if agentUUID != "" {
		url = fmt.Sprintf("http://%s:%s/v1/agent?uuid=%s", conf["host"], conf["port"], agentUUID)
	} else {
		url = fmt.Sprintf("http://%s:%s/v1/agent", conf["host"], conf["port"])
	}

	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return agents, err
	}

	// Send the request via a client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return agents, err
	}

	// Defer the closing of the body
	defer resp.Body.Close()
	// Read the content into a byte array
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return agents, err
	}

	// Marshal API data into object
	err = json.Unmarshal(body, &agents)
	return agents, err
}

// DisplayAgents is responsible for displaying a set of Agents to the terminal
func (capi ClientAPI) DisplayAgents(agents []defs.AgentAdvert, detail bool) error {

	if len(agents) == 0 {
		fmt.Println("No agents found.")
		return nil
	}

	if agents[0].UUID == "" {
		fmt.Println("No agents found.")
		return nil
	}

	if detail {

		// TODO(moswalt): if nothing found, API should return either null or empty slice, and client should handle this
		tmpl, err := template.New("test").Parse(
			`Agent UUID:  {{.UUID}}
Expires:  {{.Expires}}
Collector Summary: {{.CollectorSummary}}
Facts:
{{.PPFacts}}` + "\n")

		if err != nil {
			return err
		}

		// Output retrieved data
		for i := range agents {
			err = tmpl.Execute(os.Stdout, agents[i])
			if err != nil {
				return err
			}
		}

	} else {
		w := new(tabwriter.Writer)

		// Format in tab-separated columns with a tab stop of 8.
		w.Init(os.Stdout, 0, 8, 0, '\t', 0)
		fmt.Fprintln(w, "UUID\tEXPIRES\tADDR\tFACT SUMMARY\tCOLLECTOR SUMMARY")

		for i := range agents {
			fmt.Fprintf(
				w,
				"%s\t%s\t%s\t%s\t%s\n",
				hostresources.TruncateID(agents[i].UUID),
				agents[i].Expires,
				agents[i].DefaultAddr,
				agents[i].FactSummary(),
				agents[i].CollectorSummary(),
			)
		}
		fmt.Fprintln(w)
		w.Flush()

	}

	return nil
}
