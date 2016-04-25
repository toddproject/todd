/*
	Primary entry point for ToDD Agent

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Mierdin/todd/agent/cache"
	"github.com/Mierdin/todd/agent/defs"
	"github.com/Mierdin/todd/agent/facts"
	"github.com/Mierdin/todd/agent/testing"
	"github.com/Mierdin/todd/comms"
	"github.com/Mierdin/todd/config"
	"github.com/Mierdin/todd/hostresources"
	log "github.com/Sirupsen/logrus"
)

// Command-line Arguments
var arg_config string

func init() {

	flag.Usage = func() {
		fmt.Print(`Usage: todd-agent [OPTIONS] COMMAND [arg...]

    An extensible framework for providing natively distributed testing on demand

    Options:
      --config="/etc/todd/agent.cfg"          Absolute path to ToDD agent config file`, "\n\n")

		os.Exit(0)
	}

	flag.StringVar(&arg_config, "config", "/etc/todd/agent.cfg", "ToDD agent config file location")
	flag.Parse()

	// TODO(moswalt): Implement configurable loglevel in server and agent
	log.SetLevel(log.DebugLevel)
}

func main() {

	cfg := config.GetConfig(arg_config)

	// Set up cache
	var ac = cache.NewAgentCache(cfg)
	ac.Init()

	// Generate UUID
	uuid := hostresources.GenerateUuid()
	ac.SetKeyValue("uuid", uuid)

	log.Infof("ToDD Agent Activated: %s", uuid)

	// Start test data reporting service
	go testing.WatchForFinishedTestRuns(cfg)

	// Construct comms package
	var tc = comms.NewToDDComms(cfg)

	// Spawn goroutine to listen for tasks issued by server
	go func() {
		for {
			tc.CommsPackage.ListenForTasks(uuid)
			if err != nil {
				log.Warn("ListenForTasks reported a failure. Trying again...")
			}
		}
	}()

	// Watch for changes to group membership
	go tc.CommsPackage.WatchForGroup()

	// Continually advertise agent status into message queue
	for {

		// Gather assets here as a map, and refer to a key in that map in the below struct
		gatheredAssets := GetLocalAssets(cfg)

		var defaultaddr string
		if cfg.LocalResources.IPAddrOverride != "" {
			defaultaddr = cfg.LocalResources.IPAddrOverride
		} else {
			defaultaddr = hostresources.GetIPOfInt(cfg.LocalResources.DefaultInterface).String()
		}

		// Create an AgentAdvert instance to represent this particular agent
		me := defs.AgentAdvert{
			Uuid:           uuid,
			DefaultAddr:    defaultaddr,
			FactCollectors: gatheredAssets["factcollectors"],
			Testlets:       gatheredAssets["testlets"],
			Facts:          facts.GetFacts(cfg),
			LocalTime:      time.Now().UTC(),
		}

		// Advertise this agent
		err := tc.CommsPackage.AdvertiseAgent(me)
		if err != nil {
			log.Error("Failed to advertise agent after several retries")
		}

		time.Sleep(15 * time.Second) // TODO(moswalt): make configurable
	}

}
