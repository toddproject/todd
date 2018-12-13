/*
    ToDD TSDB Functions

    This file holds the infrastructure for TSDB abstractions in ToDD.

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package tsdb

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/toddproject/todd/config"
)

// Package represents all of the behavior that a ToDD TSDB plugin must support
type Package interface {
	WriteData(string, string, string, map[string]map[string]map[string]interface{}) error
}

// toddTSDB is a struct to hold anything that satisfies the databasePackage interface
type toddTSDB struct {
	Package
}

// NewToddTSDB will create a new instance of toddTSDB, and load the desired
// databasePackage-compatible comms package into it.
func NewToddTSDB(cfg config.Config) *toddTSDB { // TODO: return Package instead of *struct embedding Package

	// Create toddTSDB instance
	var tsdb toddTSDB

	// Load the appropriate TSDB package based on config file
	switch cfg.TSDB.Plugin {
	case "influxdb":
		tsdb.Package = newInfluxDB(cfg)
	default:
		log.Error("Invalid DB plugin in config file")
		os.Exit(1)
	}

	return &tsdb
}
