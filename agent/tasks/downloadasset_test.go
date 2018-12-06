/*
	Tests for downloadasset task

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package tasks

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/toddproject/todd/config"
)

// TODO(mierdin): This unit test works, but you need other tests to test for problems
// with the i/o system, filesystem, or http call. (rainy day testing)

// TODO: write a test that removes the "factcollectors" and tests for an error

func TestTaskRun(t *testing.T) {
	responseData := []byte("responsetext")
	// Test server that always responds with 200 code, and specific payload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/text")
		w.Write(responseData)
	}))
	defer server.Close()

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	outPath := filepath.Join(tmpDir, "assets", "factcollectors")
	err = os.MkdirAll(outPath, 0755)
	if err != nil {
		t.Fatal(err)
	}

	assets := []string{
		server.URL + "/factcollectors/test_collector_1",
		server.URL + "/factcollectors/test_collector_2",
		server.URL + "/factcollectors/test_collector_3",
	}

	// setup task object
	task := NewDownloadAsset(assets)

	var cfg config.Config
	cfg.LocalResources.OptDir = tmpDir

	// Run task
	err = task.Run(&cfg, nil, nil)
	if err != nil {
		t.Fatal("DownloadCollectors failed:", err)
	}

	finfos, err := ioutil.ReadDir(outPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(finfos) != len(assets) {
		t.Errorf("wanted %d files, got %d", len(assets), len(finfos))
	}

	for _, asset := range assets {
		path := filepath.Join(outPath, filepath.Base(asset))
		data, err := ioutil.ReadFile(path)
		if os.IsNotExist(err) {
			t.Errorf("wanted %q to be created, but it wasn't", path)
			continue
		}
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(data, responseData) {
			t.Errorf("wanted %q to contains %q, but it was %q", path, string(responseData), string(data))
		}
	}
}
