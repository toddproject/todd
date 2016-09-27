/*
	Primary entry point for ToDD Agent

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/toddproject/todd/agent/cache"
	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/agent/facts"
	"github.com/toddproject/todd/agent/responses"
	"github.com/toddproject/todd/comms"
	"github.com/toddproject/todd/config"
	"github.com/toddproject/todd/hostresources"
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

	cfg, err := config.GetConfig(arg_config)
	if err != nil {
		os.Exit(1)
	}

	// Set up cache
	var ac = cache.NewAgentCache(cfg)
	ac.Init()

	// Generate UUID
	uuid := hostresources.GenerateUuid()
	ac.SetKeyValue("uuid", uuid)

	log.Infof("ToDD Agent Activated: %s", uuid)

	// Start test data reporting service
	go watchForFinishedTestRuns(cfg)

	// Construct comms package
	tc, err := comms.NewToDDComms(cfg)
	if err != nil {
		os.Exit(1)
	}

	// Spawn goroutine to listen for tasks issued by server
	go func() {
		for {
			err := tc.CommsPackage.ListenForTasks(uuid)
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

// watchForFinishedTestRuns simply watches the local cache for any test runs that have test data.
// It will periodically look at the table and send any present test data back to the server as a response.
// When the server has successfully received this data, it will send a task back to this specific agent
// to delete this row from the cache.
func watchForFinishedTestRuns(cfg config.Config) error {

	var ac = cache.NewAgentCache(cfg)

	agentUuid := ac.GetKeyValue("uuid")

	for {

		time.Sleep(5000 * time.Millisecond)

		testruns, err := ac.GetFinishedTestRuns()
		if err != nil {
			log.Error("Problem retrieving finished test runs")
			return errors.New("Problem retrieving finished test runs")
		}

		for testUuid, testData := range testruns {

			log.Debug("Found ripe testrun: ", testUuid)

			utdr := responses.UploadTestDataResponse{
				TestUuid: testUuid,
				TestData: testData,
			}
			utdr.AgentUuid = agentUuid
			utdr.Type = "TestData" //TODO(mierdin): This is an extra step. Maybe a factory function for the task could help here?

			tc, err := comms.NewToDDComms(cfg)
			if err != nil {
				return err
			}
			tc.CommsPackage.SendResponse(utdr)

		}

	}

	return nil

}
