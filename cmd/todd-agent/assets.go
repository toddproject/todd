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

	log "github.com/Sirupsen/logrus"

	"github.com/toddproject/todd/config"
	"github.com/toddproject/todd/hostresources"
)

// GetLocalAssets gathers all currently installed assets, and generates a map of their names and hashes.
func GetLocalAssets(cfg config.Config) map[string]map[string]string {

	asset_types := []string{
		"factcollectors",
		"testlets",
	}

	final_return_assets := make(map[string]map[string]string)

	for i := range asset_types {
		found_assets := make(map[string]string)

		thisType := asset_types[i]

		// this is the function that will generate a hash for a file and add it to our asset map
		discover_assets := func(path string, f os.FileInfo, err error) error {

			if f.IsDir() != true {
				// Generate hash
				found_assets[f.Name()] = hostresources.GetFileSHA256(path)
				log.Debugf("Asset found locally: %s (with hash %s)", f.Name(), hostresources.GetFileSHA256(path))
			}
			return nil
		}

		// set up a directory for this particular asset type
		thisAssetDir := fmt.Sprintf("%s/assets/%s", cfg.LocalResources.OptDir, thisType)

		// create fact collector directory if needed
		err := os.MkdirAll(thisAssetDir, 0777)
		if err != nil {
			log.Error("Problem creating asset directory")
			os.Exit(1)
		}

		// Perform above Walk function (discover_assets) on the collector directory
		err = filepath.Walk(thisAssetDir, discover_assets)
		if err != nil {
			log.Error("Problem getting assets")
			os.Exit(1)
		}

		final_return_assets[thisType] = found_assets

	}

	// Return collectors so that the calling function can pass this to the registry for enforcement
	return final_return_assets

}
