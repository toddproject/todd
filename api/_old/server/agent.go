/*
    ToDD API - manages agents

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/toddproject/todd/agent/defs"
)

func (tapi ToDDApi) Agent(w http.ResponseWriter, r *http.Request) {

	agentList, err := tapi.tdb.GetAgents()
	if err != nil {
		log.Errorln(err)
		http.Error(w, "Internal Error", 500)
		return
	}

	// Make sure UUID string is provided
	if uuid := r.URL.Query().Get("uuid"); uuid != "" {
		// Let's use the full list so we can identify the right agent if the user specified a short
		for i := range agentList {
			if strings.HasPrefix(agentList[i].UUID, uuid) {
				// Replace agentList with first match and break
				agentList = []defs.AgentAdvert{agentList[i]}
				break
			}
		}
	}

	response, err := json.MarshalIndent(agentList, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Fprint(w, string(response))
}
