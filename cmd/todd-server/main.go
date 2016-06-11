/*
    ToDD Server - primary entry point

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

	toddapi "github.com/Mierdin/todd/api/server"
	"github.com/Mierdin/todd/comms"
	"github.com/Mierdin/todd/config"
	"github.com/Mierdin/todd/db"
	"github.com/Mierdin/todd/server/grouping"
	log "github.com/Sirupsen/logrus"
)

// Command-line Arguments
var arg_config string

func init() {

	flag.Usage = func() {
		fmt.Print(`Usage: todd-server [OPTIONS] COMMAND [arg...]

    An extensible framework for providing natively distributed testing on demand

    Options:
      --config="/etc/todd/server.cfg"          Absolute path to ToDD server config file`, "\n\n")

		os.Exit(0)
	}

	flag.StringVar(&arg_config, "config", "/etc/todd/server.cfg", "ToDD server config file location")
	flag.Parse()

	// TODO(moswalt): Implement configurable loglevel in server and agent
	log.SetLevel(log.DebugLevel)
}

func main() {

	todd_version := "0.0.1"

	cfg, err := config.GetConfig(arg_config)
	if err != nil {
		os.Exit(1)
	}

	// Start serving collectors and testlets, and retrieve map of names and hashes
	assets := serveAssets(cfg)

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

	// Start listening for agent advertisements
	tc, err := comms.NewToDDComms(cfg)
	if err != nil {
		os.Exit(1)
	}

	go func() {
		for {
			err := tc.CommsPackage.ListenForAgent(assets)
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

	log.Infof("ToDD server v%s. Press any key to exit...\n", todd_version)

	// Sssh, sssh, only dreams now....
	for {
		time.Sleep(time.Second * 10)
	}
}
