/*
    ToDD API

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package api

import (
	"fmt"
	"net/http"

	"github.com/Mierdin/todd/db"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"

	"github.com/Mierdin/todd/config"
)

func handlers(t *ToDDApi) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/v1/agent", t.Agent).Methods("GET")
	r.HandleFunc("/v1/groups", t.Groups).Methods("GET")
	r.HandleFunc("/v1/object/list", t.ListObjects).Methods("GET")
	r.HandleFunc("/v1/object/group", t.ListObjects).Methods("GET")
	r.HandleFunc("/v1/object/testrun", t.ListObjects).Methods("GET")
	r.HandleFunc("/v1/object/create", t.CreateObject).Methods("POST")
	r.HandleFunc("/v1/object/delete", t.DeleteObject).Methods("POST")
	r.HandleFunc("/v1/testrun/run", t.Run).Methods("POST")
	r.HandleFunc("/v1/testdata", t.TestData).Methods("GET")

	return r
}

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

	serve_url := fmt.Sprintf("%s:%s", tapi.cfg.API.Host, tapi.cfg.API.Port)

	log.Infof("Serving ToDD Server API at: %s\n", serve_url)
	return http.ListenAndServe(serve_url, handlers(&tapi))
}
