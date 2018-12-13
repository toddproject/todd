/*
	ToDD Agent Asset management

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package main

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/toddproject/todd/config"
	"github.com/toddproject/todd/hostresources"
)

// GetLocalAssets gathers all currently installed assets, and generates a map of their names and hashes.
func GetLocalAssets(cfg *config.ToDDConfig) map[string]map[string]string {

	assetTypes := []string{
		"factcollectors",
		"testlets",
	}

	finalReturnAssets := make(map[string]map[string]string)

	for _, thisType := range assetTypes {
		foundAssets := make(map[string]string)

		// this is the function that will generate a hash for a file and add it to our asset map
		discoverAssets := func(path string, f os.FileInfo, err error) error {

			if !f.IsDir() {
				// Generate hash
				foundAssets[f.Name()] = hostresources.GetFileSHA256(path)
				log.Debugf("Asset found locally: %s (with hash %s)", f.Name(), hostresources.GetFileSHA256(path))
			}
			return nil
		}

		// set up a directory for this particular asset type
		thisAssetDir := fmt.Sprintf("%s/assets/%s", cfg.OptDir, thisType)

		// create fact collector directory if needed
		err := os.MkdirAll(thisAssetDir, 0777)
		if err != nil {
			log.Error("Problem creating asset directory")
			os.Exit(1)
		}

		// Perform above Walk function (discover_assets) on the collector directory
		err = filepath.Walk(thisAssetDir, discoverAssets)
		if err != nil {
			log.Error("Problem getting assets")
			os.Exit(1)
		}

		finalReturnAssets[thisType] = foundAssets

	}

	// Return collectors so that the calling function can pass this to the registry for enforcement
	return finalReturnAssets

}
