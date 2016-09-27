/*
	Tests for downloadasset task

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package tasks

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TODO(mierdin): This unit test works, but you need other tests to test for problems
// with the i/o system, filesystem, or http call. (rainy day testing)

func TestTaskRun(t *testing.T) {

	// Test server that always responds with 200 code, and specific payload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/text")
		fmt.Fprintln(w, `responsetext`)
	}))
	defer server.Close()

	// Make a transport that reroutes all traffic to the example server
	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	// Make a http.Client with the transport
	httpClient := &http.Client{Transport: transport}

	// setup task object
	task := DownloadAssetTask{
		HTTPClient:   httpClient,
		Fs:           mockFS{},
		Ios:          mockIoSys{},
		CollectorDir: "/tmp",
		Assets: []string{
			"http://127.0.0.1:8080/factcollectors/test_collector_1",
			"http://127.0.0.1:8080/factcollectors/test_collector_2",
			"http://127.0.0.1:8080/factcollectors/test_collector_3", // write a test that removes the "factcollectors" and tests for an error
		},
	}

	// Run task
	err := task.Run()

	if err != nil {
		t.Fatalf("DownloadCollectors failed in some way and wasn't supposed to")
	}

}
