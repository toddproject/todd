/*
    ToDD comms functions

    This file holds the infrastructure for agent-server communication abstractions in ToDD.

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package comms

import (
	"errors"

	log "github.com/Sirupsen/logrus"

	"github.com/toddproject/todd/agent/cache"
	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/agent/responses"
	"github.com/toddproject/todd/agent/tasks"
	"github.com/toddproject/todd/config"
)

type assetProvider interface {
	Assets() map[string]map[string]string
}

// Package will ensure that whatever specific comms struct is loaded at compile time will support
// all of the necessary features/functions that we need to make ToDD work. In short, this interface
// represents a list of things that the server and agents do on the message queue.
type Package interface {

	// TODO(mierdin) best way to document interface or function args? I've tried to document
	// them minimally below, but would like a better way to document the meaning behind
	// the arguments defined here.

	// (agent advertisement to advertise)
	AdvertiseAgent(defs.AgentAdvert) error

	// (map of assets:hashes, base asset URL)
	ListenForAgent(assetProvider, string) error

	// (uuid)
	ListenForTasks(string) error

	// (queuename, task)
	SendTask(string, tasks.Task) error

	// watches for new group membership instructions in the cache and reregisters
	WatchForGroup() error

	ListenForGroupTasks(string, chan bool) error

	ListenForResponses(*chan bool) error
	SendResponse(responses.Response) error

	// adds a cache agent to the comms package. temporary until
	// comms package is refactored
	setAgentCache(*cache.AgentCache)
}

// toddComms is a struct to hold anything that satisfies the CommsPackage interface
type toddComms struct {
	Package
}

// NewToDDComms will create a new instance of toddComms, and load the desired
// CommsPackage-compatible comms package into it.
func NewToDDComms(cfg config.Config) (*toddComms, error) { // TODO: Return Package instead of *struct embedding Package

	var tc toddComms

	// Load the appropriate comms package based on config file
	switch cfg.Comms.Plugin {
	case "rabbitmq":
		tc.Package = newRabbitMQComms(cfg)
	case "amazon_mq":
		tc.Package = newAMQComms(cfg)
	default:
		log.Error("Invalid comms plugin in config file")
		return nil, errors.New("Invalid comms plugin in config file")
	}

	return &tc, nil

}

// NewAgentComms returns a comms instance configured for agent usages.
//
// TODO: accept an interface for cache instead of concrete type
func NewAgentComms(cfg config.Config, ac *cache.AgentCache) (*toddComms, error) {
	comms, err := NewToDDComms(cfg)
	if err == nil {
		comms.setAgentCache(ac)
	}
	return comms, err
}
