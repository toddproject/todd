/*
    ToDD Client API Calls for "todd groups"

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/Mierdin/todd/hostresources"
)

// Groups will query ToDD for a map containing current agent-to-group mappings
func (capi ClientApi) Groups(conf map[string]string) error {

	url := fmt.Sprintf("http://%s:%s/v1/groups", conf["host"], conf["port"])

	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Send the request via a client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Defer the closing of the body
	defer resp.Body.Close()
	// Read the content into a byte array
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Marshal API data into map
	var groupmap map[string]string
	err = json.Unmarshal(body, &groupmap)
	if err != nil {
		return err
	}

	w := new(tabwriter.Writer)

	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintln(w, "UUID\tGROUP NAME")

	for agent_uuid, group_name := range groupmap {
		fmt.Fprintf(
			w,
			"%s\t%s\n",
			hostresources.TruncateID(agent_uuid),
			group_name,
		)
	}
	fmt.Fprintln(w)
	w.Flush()

	return nil

}
