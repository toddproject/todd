/*
    ToDD Database Functions

    This file holds the infrastructure for database abstractions in ToDD.

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package db

import (
	"errors"

	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/config"
	"github.com/toddproject/todd/server/objects"
)

var (
	ErrInvalidDBPlugin = errors.New("Invalid DB plugin in config file")
	ErrNotExist        = errors.New("Value does not exist")
)

// DatabasePackage represents all of the behavior that a ToDD database plugin must support
type DatabasePackage interface {

	// (no args)
	Init() error

	// (agent advertisement to set)
	SetAgent(defs.AgentAdvert) error

	GetAgent(string) (*defs.AgentAdvert, error)
	GetAgents() ([]defs.AgentAdvert, error)

	// (agent advertisement to remove)
	RemoveAgent(defs.AgentAdvert) error

	SetObject(objects.ToddObject) error
	GetObjects(string) ([]objects.ToddObject, error)
	DeleteObject(string, string) error

	GetGroupMap() (map[string]string, error)
	SetGroupMap(map[string]string) error

	// Testing
	InitTestRun(string, map[string]map[string]string) error
	SetAgentTestStatus(string, string, string) error
	GetTestStatus(string) (map[string]string, error)
	SetAgentTestData(string, string, string) error
	GetAgentTestData(string, string) (map[string]string, error)
	WriteCleanTestData(string, string) error
	GetCleanTestData(string) (string, error)
}

// NewToddDB will create a new instance of toddDatabase, and load the desired
// databasePackage-compatible comms package into it.
func NewToddDB(cfg config.Config) (DatabasePackage, error) {

	// Create toddDatabase instance
	var tdb DatabasePackage

	// Load the appropriate DB package based on config file
	switch cfg.DB.Plugin {
	case "etcd":
		tdb = newEtcdDB(cfg)
	default:
		return nil, ErrInvalidDBPlugin
	}

	return tdb, nil
}
