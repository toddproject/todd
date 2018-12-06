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
var argConfig string

func init() {

	flag.Usage = func() {
		fmt.Print(`Usage: todd-agent [OPTIONS] COMMAND [arg...]

    An extensible framework for providing natively distributed testing on demand

    Options:
      --config="/etc/todd/agent.cfg"          Absolute path to ToDD agent config file`, "\n\n")

		os.Exit(0)
	}

	flag.StringVar(&argConfig, "config", "/etc/todd/agent.cfg", "ToDD agent config file location")
	flag.Parse()

	// TODO(moswalt): Implement configurable loglevel in server and agent
	log.SetLevel(log.DebugLevel)
}

func main() {

	cfg, err := config.GetConfig(argConfig)
	if err != nil {
		os.Exit(1)
	}

	// Set up cache
	ac, err := cache.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer ac.Close()

	// Generate UUID
	uuid := hostresources.GenerateUUID()
	err = ac.SetKeyValue("uuid", uuid)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("ToDD Agent Activated: %s", uuid)

	// Start test data reporting service
	go watchForFinishedTestRuns(cfg, ac)

	// Construct comms package
	tc, err := comms.NewAgentComms(cfg, ac)
	if err != nil {
		os.Exit(1)
	}

	// Spawn goroutine to listen for tasks issued by server
	go func() {
		for {
			err := tc.Package.ListenForTasks(uuid)
			if err != nil {
				log.Warn("ListenForTasks reported a failure. Trying again...")
			}
		}
	}()

	// Watch for changes to group membership
	go tc.Package.WatchForGroup()

	// Get default IP address for the server.
	// This address is primarily used so that the server knows how to orchestrate tests.
	// (i.e. This agent publishes it's default address, and the server instructs other agents to target it in tests)
	defaultaddr, err := hostresources.GetDefaultInterfaceIP(cfg.LocalResources.DefaultInterface, cfg.LocalResources.IPAddrOverride)
	if err != nil {
		log.Fatalf("Unable to derive address from configured DefaultInterface: %v", err)
	}

	// Gather assets here as a map, and refer to a key in that map in the below struct
	gatheredAssets := GetLocalAssets(cfg)

	fcts, err := facts.GetFacts(cfg)
	if err != nil {
		log.Errorf("Error gathering facts: %v", err)
	}

	// Continually advertise agent status into message queue
	for {
		// Create an AgentAdvert instance to represent this particular agent
		me := defs.AgentAdvert{
			UUID:           uuid,
			DefaultAddr:    defaultaddr,
			FactCollectors: gatheredAssets["factcollectors"],
			Testlets:       gatheredAssets["testlets"],
			Facts:          fcts,
			LocalTime:      time.Now().UTC(),
		}

		// Advertise this agent
		err = tc.Package.AdvertiseAgent(me)
		if err != nil {
			log.Error("Failed to advertise agent after several retries")
		}

		time.Sleep(10 * time.Second) // TODO(moswalt): make configurable
	}

}

// watchForFinishedTestRuns simply watches the local cache for any test runs that have test data.
// It will periodically look at the table and send any present test data back to the server as a response.
// When the server has successfully received this data, it will send a task back to this specific agent
// to delete this row from the cache.
func watchForFinishedTestRuns(cfg config.Config, ac *cache.AgentCache) error {
	agentUUID, err := ac.GetKeyValue("uuid")
	if err != nil {
		return err
	}

	for {

		time.Sleep(5000 * time.Millisecond)

		testruns, err := ac.GetFinishedTestRuns()
		if err != nil {
			log.Error("Problem retrieving finished test runs")
			return errors.New("Problem retrieving finished test runs")
		}

		for testUUID, testData := range testruns {

			log.Debug("Found ripe testrun: ", testUUID)

			tc, err := comms.NewToDDComms(cfg)
			if err != nil {
				return err
			}

			utdr := responses.NewUploadTestData(agentUUID, testUUID, testData)
			tc.Package.SendResponse(utdr)

		}

	}
}
