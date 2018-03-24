/*
    ToDD API - manages agent groups

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

func (tapi ToDDApi) Groups(w http.ResponseWriter, r *http.Request) {

	log.Info("Received request for group map")

	groupmap, err := tapi.tdb.GetGroupMap()
	if err != nil {
		log.Errorln(err)
		http.Error(w, "Internal Error", 500)
		return
	}

	response, err := json.MarshalIndent(groupmap, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Fprint(w, string(response))
}
