/*
	ToDD task - download collectors

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package tasks

import (
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/toddproject/todd/agent/cache"
	"github.com/toddproject/todd/config"
)

// DownloadAsset defines this particular task. It contains definitions not only for the task message, but
// also HTTPClient, filesystem, and i/o system abstractions to be more conducive to unit testing.
type DownloadAsset struct {
	BaseTask
	Assets []string `json:"assets"`
}

// NewDownloadAsset returns a new DownloadAsset task.
func NewDownloadAsset(assets []string) *DownloadAsset {
	return &DownloadAsset{
		BaseTask: BaseTask{Type: TypeDownloadAsset},
		Assets:   assets,
	}
}

// Run contains the logic necessary to perform this task on the agent. This particular task will download all required assets,
// copy them into the appropriate directory, and ensure that the execute permission is given to each collector file.
func (t *DownloadAsset) Run(cfg *config.Config, _ *cache.AgentCache, _ Responder) error {
	// Iterate over the slice of collectors and download them.
	baseAssetDir := filepath.Join(cfg.LocalResources.OptDir, "assets")
	for _, assetURL := range t.Assets {
		var assetDir string
		switch {
		case strings.Contains(assetURL, "factcollectors"):
			assetDir = filepath.Join(baseAssetDir, "factcollectors")
		case strings.Contains(assetURL, "testlets"):
			assetDir = filepath.Join(baseAssetDir, "testlets")
		default:
			return errors.New("invalid asset download URL received")
		}

		err := t.downloadAsset(assetURL, assetDir)
		if err != nil {
			return errors.Wrapf(err, "downloading %q to %s", assetURL, assetDir)
		}
	}

	return nil
}

// downloadAsset will download an asset at the specified URL, into the specified directory
func (t *DownloadAsset) downloadAsset(url, directory string) error {
	path := filepath.Join(directory, path.Base(url))
	log.Infof("Downloading %q to %s.", url, path)

	// TODO: What if this already exists? Consider checking file existence first with io.IsExist?
	output, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0744)
	if err != nil {
		return errors.Wrapf(err, "creating %s", path)
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		// If we have a problem retrieving the testlet, we want to return immediately,
		// instead of writing an empty file to disk
		return errors.Wrapf(err, "making HTTP call")
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.Errorf("%s response", response.Status)
	}

	n, err := io.Copy(output, response.Body)
	if err != nil {
		return errors.Wrapf(err, "writing to %q", path)
	}

	log.Infof("%d bytes downloaded.", n)
	return nil
}
