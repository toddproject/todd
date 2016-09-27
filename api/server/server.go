/*
    ToDD API

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/toddproject/todd/db"

	"github.com/toddproject/todd/config"
)

type ToDDApi struct {
	cfg config.Config
	tdb db.DatabasePackage
}

func (tapi ToDDApi) Start(cfg config.Config) error {

	tapi.cfg = cfg

	tdb, err := db.NewToddDB(tapi.cfg)
	if err != nil {
		return err
	}

	tapi.tdb = tdb

	// TODO(mierdin): This needs a lot of work. Not only is the version very static
	// (which is okay for now until we hit a new version)
	// but the static routing is gross. Make this not suck
	http.HandleFunc("/v1/agent", tapi.Agent)
	http.HandleFunc("/v1/groups", tapi.Groups)
	http.HandleFunc("/v1/object/list", tapi.ListObjects)
	http.HandleFunc("/v1/object/group", tapi.ListObjects)
	http.HandleFunc("/v1/object/testrun", tapi.ListObjects)
	http.HandleFunc("/v1/object/create", tapi.CreateObject)
	http.HandleFunc("/v1/object/delete", tapi.DeleteObject)
	http.HandleFunc("/v1/testrun/run", tapi.Run)
	http.HandleFunc("/v1/testdata", tapi.TestData)

	serve_url := fmt.Sprintf("%s:%s", tapi.cfg.API.Host, tapi.cfg.API.Port)

	log.Infof("Serving ToDD Server API at: %s\n", serve_url)
	return http.ListenAndServe(serve_url, nil)
}
