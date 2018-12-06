/*
    ToDD Server - primary entry point

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	toddapi "github.com/toddproject/todd/api/_old/server"
	api "github.com/toddproject/todd/api/exp"

	"github.com/toddproject/todd/comms"
	"github.com/toddproject/todd/config"
	"github.com/toddproject/todd/db"
	"github.com/toddproject/todd/hostresources"
	"github.com/toddproject/todd/server/grouping"
)

var (
	toddVersion = "0.0.1"
	// Command-line Arguments
	argVersion string
)

func init() {

	flag.Usage = func() {
		fmt.Print(`Usage: todd-server [OPTIONS] COMMAND [arg...]

    An extensible framework for providing natively distributed testing on demand

    Options:
      --config="/etc/todd/server.cfg"          Absolute path to ToDD server config file`, "\n\n")

		os.Exit(0)
	}

	flag.StringVar(&argVersion, "config", "/etc/todd/server.cfg", "ToDD server config file location")
	flag.Parse()

	// TODO(moswalt): Implement configurable loglevel in server and agent
	log.SetLevel(log.DebugLevel)
}

func main() {

	cfg, err := config.GetConfig(argVersion)
	if err != nil {
		log.Fatalf("Problem getting configuration: %v\n", err)
	}

	// Start serving collectors and testlets, and retrieve map of names and hashes
	assets := newAssetConfig(cfg)

	// Perform database initialization tasks
	tdb, err := db.NewToddDB(cfg)
	if err != nil {
		log.Fatalf("Error setting up database: %v\n", err)
	}

	if err := tdb.Init(); err != nil {
		log.Fatalf("Error initializing database: %v\n", err)
	}

	// Initialize API
	var tapi toddapi.ToDDApi
	go func() {
		log.Fatal(tapi.Start(cfg))
	}()

	var api api.ToDDApiExp
	go func() {
		log.Fatal(api.Start(cfg))
	}()

	// Start listening for agent advertisements
	tc, err := comms.NewToDDComms(cfg)
	if err != nil {
		log.Fatalf("Problem connecting to comms: %v\n", err)
	}

	// Get default IP address for the server.
	// This address is primarily used to inform the agents of the URL they should use to download assets
	defaultaddr, err := hostresources.GetDefaultInterfaceIP(cfg.LocalResources.DefaultInterface, cfg.LocalResources.IPAddrOverride)
	if err != nil {
		log.Fatalf("Unable to derive address from configured DefaultInterface: %v", err)
	}

	assetURLPrefix := fmt.Sprintf("http://%s:%s", defaultaddr, cfg.Assets.Port)

	go func() {
		for {
			err := tc.Package.ListenForAgent(assets, assetURLPrefix)
			if err != nil {
				log.Fatalf("Error listening for ToDD Agents")
			}
		}
	}()

	// Kick off group calculation in background
	go func() {
		for {
			log.Info("Beginning group calculation")
			grouping.CalculateGroups(cfg)
			time.Sleep(time.Second * time.Duration(cfg.Grouping.Interval))
		}
	}()

	log.Infof("ToDD server v%s. Press any key to exit...\n", toddVersion)

	// Sssh, sssh, only dreams now....
	for {
		time.Sleep(time.Second * 10) // TODO: Replace with select{}, blocks forever without interrupt
	}
}
