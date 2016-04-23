/*
    Asset management for ToDD server

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Mierdin/todd/assets"
	"github.com/Mierdin/todd/config"
	"github.com/Mierdin/todd/hostresources"
	log "github.com/Sirupsen/logrus"
)

// serveAssets is responsible for deriving embedded assets like collector files and testlets from the golang source generated by go-bindata
// These will be written to the appropriate directories, a hash (SHA256) will be generated, and these files will be served via HTTP
// This function is typically run on the ToDD server.
func serveAssets(cfg config.Config) map[string]map[string]string {

	// Initialize asset map
	final_asset_map := make(map[string]map[string]string)

	type toddAsset struct {
		name      string
		assetType string
		dir       string
		assetDir  string
		assets    []string
	}

	var collectors = toddAsset{
		name:     "factcollectors",
		dir:      fmt.Sprintf("%s/assets/factcollectors", cfg.LocalResources.OptDir),
		assetDir: "agent/facts/collectors/",
		assets: []string{
			"get_addresses", // TODO(moswalt): Temporary measure - should figure out a way to iterate over these in the Asset
			"get_hostname",
		},
	}

	var testlets = toddAsset{
		name:     "testlets",
		dir:      fmt.Sprintf("%s/assets/testlets", cfg.LocalResources.OptDir),
		assetDir: "agent/testing/testlets/",
		assets: []string{
			"ping",
			"iperf",
			"http",
		},
	}

	embeddedAssets := []toddAsset{collectors, testlets}

	for x := range embeddedAssets {

		thisAsset := embeddedAssets[x]

		final_asset_map[thisAsset.name] = make(map[string]string)

		// create asset directory if needed
		err := os.MkdirAll(thisAsset.dir, 0777) // TODO(mierdin): obviously overkill - get this down to a reasonable value
		if err != nil {
			log.Error("Problem creating asset directory ", thisAsset.dir)
			os.Exit(1)
		}

		for i := range thisAsset.assets {

			asset_name := thisAsset.assets[i]

			log.Debug("Loading asset - ", asset_name)

			// Retrieve Asset from embedded Go source
			data, err := assets.Asset(fmt.Sprintf("%s%s", thisAsset.assetDir, asset_name))
			if err != nil {
				log.Error("Error retrieving asset from embedded source")
				os.Exit(1)
			}

			// Derive full path to asset file, and write it
			file := fmt.Sprintf("%s/%s", thisAsset.dir, asset_name)
			err = ioutil.WriteFile(file, data, 0744)
			if err != nil {
				log.Error("Error writing asset file")
				os.Exit(1)
			}

			// Generate SHA256 for this asset file, and append to collector map
			final_asset_map[thisAsset.name][asset_name] = hostresources.GetFileSHA256(file)
		}

	}

	// Begin serving files to agents
	// TODO(moswalt): Handle error
	go http.ListenAndServe(fmt.Sprintf(":%s", cfg.Assets.Port), http.FileServer(http.Dir(fmt.Sprintf("%s/assets", cfg.LocalResources.OptDir))))

	return final_asset_map

}
