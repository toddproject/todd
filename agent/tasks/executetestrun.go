/*
	ToDD task - test run

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package tasks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/Mierdin/todd/agent/cache"
	"github.com/Mierdin/todd/agent/testing/testlets"
	"github.com/Mierdin/todd/config"
)

var (
	// gatheredData represents test data from this agent for all targets.
	// Key is target name, value is JSON output from testlet for that target
	// This is reset to a blank map every time ExecuteTestRunTask is called
	gatheredData = make(map[string]string)

	// Use a wait group to ensure that all of the testlets have a chance to finish
	wg sync.WaitGroup
)

// ExecuteTestRunTask defines this particular task.
type ExecuteTestRunTask struct {
	BaseTask
	Config    config.Config `json:"-"`
	TestUuid  string        `json:"testuuid"`
	TimeLimit int           `json:"timelimit"`
}

// Run contains the logic necessary to perform this task on the agent. This particular task will execute a
// testrun that has already been installed into the local agent cache. In this context (single agent),
// a testrun will be executed once per target, all in parallel.
func (ett ExecuteTestRunTask) Run() error {

	// Make sure we're working with a clean slate
	gatheredData = map[string]string{}

	// Waiting three seconds to ensure all the agents have their tasks before we potentially hammer the network
	// TODO(mierdin): This is a bit of a copout. I would like to do something a little more robust than simply waiting
	// for a few seconds in the future.
	time.Sleep(3000 * time.Millisecond)

	// Retrieve test from cache by UUID
	var ac = cache.NewAgentCache(ett.Config)
	tr, err := ac.GetTestRun(ett.TestUuid)
	if err != nil {
		log.Error(err)
		return errors.New("Problem retrieving testrun from agent cache")
	}

	log.Debugf("IMMA FIRIN MAH LAZER (for test %s) ", ett.TestUuid)

	// Specify size of wait group
	wg.Add(len(tr.Targets))

	// Determine if this is a native testlet
	nativeTestlet, err := testlets.NewTestlet(tr.Testlet)
	native := (err == nil)

	// Execute testlets against all targets asynchronously
	for i := range tr.Targets {

		thisTarget := tr.Targets[i]

		if native {
			go func() {

				// TODO(mierdin): Just calling a temporary method to ensure we can get to the native testlet.
				// (it works!)
				log.Error(nativeTestlet.Test())
			}()
		} else {
			// Generate path to testlet and make sure it exists.
			testlet_path := fmt.Sprintf("%s/assets/testlets/%s", ett.Config.LocalResources.OptDir, tr.Testlet)
			if _, err := os.Stat(testlet_path); os.IsNotExist(err) {
				log.Errorf("Testlet %s does not exist on this agent", tr.Testlet)
				return errors.New("Error executing testrun - testlet doesn't exist on this agent.")
			}

			go func() {

				defer wg.Done()

				log.Debugf("Full testlet command and args: '%s %s %s'", testlet_path, thisTarget, tr.Args)
				cmd := exec.Command(testlet_path, thisTarget, tr.Args)

				// Stdout buffer
				cmdOutput := &bytes.Buffer{}
				// Attach buffer to command
				cmd.Stdout = cmdOutput

				// Execute collector
				cmd.Start()

				done := make(chan error, 1)
				go func() {
					done <- cmd.Wait()
				}()

				// This select statement will block until one of these two conditions are met:
				// - The testlet finishes, in which case the channel "done" will be receive a value
				// - The configured time limit is exceeded (expected for testlets running in server mode)
				select {
				case <-time.After(time.Duration(ett.TimeLimit) * time.Second):
					if err := cmd.Process.Kill(); err != nil {
						log.Errorf("Failed to kill %s after timeout: %s", testlet_path, err)
					} else {
						log.Debug("Successfully killed ", testlet_path)
					}
				case err := <-done:
					if err != nil {
						log.Errorf("Testlet %s completed with error '%s'", testlet_path, err)
						gatheredData[thisTarget] = "error"
					} else {
						log.Debugf("Testlet %s completed without error", testlet_path)
					}
				}

				// Record test data
				gatheredData[thisTarget] = string(cmdOutput.Bytes())

			}()
		}
	}

	wg.Wait()

	testdata_json, err := json.Marshal(gatheredData)
	if err != nil {
		log.Fatal("Failed to marshal post-test data")
		os.Exit(1)
	}

	// Write test data to agent cache
	err = ac.UpdateTestRunData(ett.TestUuid, string(testdata_json))
	if err != nil {
		log.Fatal("Failed to install post-test data into cache")
		os.Exit(1)
	} else {
		log.Debugf("Wrote combined post-test data for %s to cache", ett.TestUuid)
	}

	return nil
}
