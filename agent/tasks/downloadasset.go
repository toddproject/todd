/*
	ToDD task - download collectors

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package tasks

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// DownloadAssetTask defines this particular task. It contains definitions not only for the task message, but
// also HTTPClient, filesystem, and i/o system abstractions to be more conducive to unit testing.
type DownloadAssetTask struct {
	BaseTask
	HTTPClient   *http.Client `json:"-"`
	Fs           FileSystem   `json:"-"`
	Ios          IoSystem     `json:"-"`
	CollectorDir string       `json:"-"`
	Assets       []string     `json:"assets"`
	TestletDir   string       `json:"-"`
}

// Run contains the logic necessary to perform this task on the agent. This particular task will download all required assets,
// copy them into the appropriate directory, and ensure that the execute permission is given to each collector file.
func (dat DownloadAssetTask) Run() error {

	// Iterate over the slice of collectors and download them.
	for x := range dat.Assets {

		assetUrl := dat.Assets[x]

		assetDir := ""

		switch {
		case strings.Contains(assetUrl, "factcollectors"):
			assetDir = dat.CollectorDir
		case strings.Contains(assetUrl, "testlets"):
			assetDir = dat.TestletDir
		default:
			log.Error("Invalid asset download URL received")
			os.Exit(1)
		}

		err := dat.downloadAsset(assetUrl, assetDir)
		if err != nil {
			log.Error("Problem downloading asset ", assetUrl)
		}
	}

	return nil
}

// downloadAsset will download an asset at the specified URL, into the specified directory
func (dat DownloadAssetTask) downloadAsset(url, directory string) error {

	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]
	fileName = fmt.Sprintf("%s/%s", directory, fileName)
	log.Info("Downloading ", url, " to ", fileName)

	// TODO: check file existence first with io.IsExist
	output, err := dat.Fs.Create(fileName)
	if err != nil {
		return errors.New(fmt.Sprintf("Error while creating", fileName, "-", err))
	}
	defer output.Close()

	response, err := dat.HTTPClient.Get(url)
	if err != nil {
		return errors.New(fmt.Sprintf("Error while downloading", url, "-", err))
	}
	defer response.Body.Close()

	n, err := dat.Ios.Copy(output, response.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("Error while downloading", url, "-", err))
	}
	err = dat.Fs.Chmod(fileName, 0744)
	if err != nil {
		// For now, let's just throw a warning in the logs. It's not the end of the world if this fails; it may
		// be a limited issue. Worst case is the agent simply reports no facts.
		log.Warn("Problem setting execute permission on downloaded script")
	}

	log.Info(n, " bytes downloaded.")
	return nil
}
