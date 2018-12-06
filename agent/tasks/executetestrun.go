/*
	ToDD task - test run

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package tasks

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"

	"github.com/toddproject/todd/agent/cache"
	"github.com/toddproject/todd/agent/responses"
	"github.com/toddproject/todd/agent/testing"
	"github.com/toddproject/todd/config"
)

// ExecuteTestRun defines this particular task.
type ExecuteTestRun struct {
	BaseTask
	TestUUID  string `json:"test_uuid"`
	TimeLimit int    `json:"time_limit"`
}

// NewExecuteTestRun returns a new ExecuteTestRun task.
func NewExecuteTestRun(testUUID string, timeLimit int) *ExecuteTestRun {
	return &ExecuteTestRun{
		BaseTask:  BaseTask{Type: TypeExecuteTestRun},
		TestUUID:  testUUID,
		TimeLimit: timeLimit,
	}
}

type executeResult struct {
	target string
	data   json.RawMessage
}

// Run contains the logic necessary to perform this task on the agent. This particular task will execute a
// testrun that has already been installed into the local agent cache. In this context (single agent),
// a testrun will be executed once per target, all in parallel.
func (t *ExecuteTestRun) Run(cfg *config.Config, ac *cache.AgentCache, responder Responder) (err error) {
	// Retrieve UUID
	uuid, err := ac.GetKeyValue("uuid")
	if err != nil {
		return errors.Wrap(err, "retrieving UUID")
	}

	// Send status that the testing has begun, right now.
	response := responses.NewSetAgentStatus(uuid, t.TestUUID, testing.StatusTesting)
	responder(response)

	// Send fail status if error occurs
	defer func() {
		if err != nil {
			response.Status = testing.StatusFail
			responder(response)
		}
	}()

	// Waiting three seconds to ensure all the agents have their tasks before we potentially hammer the network
	//
	// TODO(mierdin): This is a temporary measure - in the future, testruns will be executed via time schedule,
	// making not only this sleep, but also the entire task unnecessary. Testruns will simply be installed, and
	// executed when the time is right. This is, in part tracked by https://github.com/toddproject/todd/issues/89
	time.Sleep(3000 * time.Millisecond)

	// Retrieve test from cache by UUID
	tr, err := ac.GetTestRun(t.TestUUID)
	if err != nil {
		return errors.Wrapf(err, "retrieving testrun from agent cache")
	}

	log.Debugf("IMMA FIRIN MAH LAZER (for test %s) ", t.TestUUID)

	testletPath, err := testing.GetTestletPath(tr.Testlet, cfg.LocalResources.OptDir)
	if err != nil {
		return errors.Wrapf(err, "getting testlet path")
	}

	// Execute testlets against all targets asynchronously
	results := make(chan executeResult)
	for _, target := range tr.Targets {
		go t.executeTestRun(target, testletPath, tr.Args, results)
	}

	// gatheredData represents test data from this agent for all targets.
	// Key is target name, value is JSON output from testlet for that target
	// This is reset to a blank map every time ExecuteTestRunTask is called
	gatheredData := map[string]*json.RawMessage{}
	for range tr.Targets {
		result := <-results
		if result.data != nil {
			gatheredData[result.target] = &result.data
		}
	}

	testdataJSON, err := json.Marshal(gatheredData)
	if err != nil {
		log.Errorf("Failed to marshal post-test data %v", err)
		// TODO(mierdin): This gets triggered for all iperf servers, because they don't provide metrics.
		// Need to figure out a more elegant way of handling this (instead of showing this message)
	}

	// Write test data to agent cache
	err = ac.UpdateTestRunData(t.TestUUID, string(testdataJSON))
	if err != nil {
		return errors.Wrapf(err, "adding test data to cache")
	}
	log.Debugf("Wrote combined post-test data for %s to cache", t.TestUUID)

	return nil
}

func (t *ExecuteTestRun) executeTestRun(target, path, args string, results chan<- executeResult) {
	r := executeResult{target: target}
	defer func() {
		// ensure result is sent regardless of return path
		results <- r
	}()

	log.Debugf("Full testlet command and args: '%s %s %s'", path, target, args)

	cmd := exec.Command(path, target, args)
	// Create a new process group so we can kill child processes if necessary.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Stdout buffer
	var out bytes.Buffer
	cmd.Stdout = &out

	// Execute testlet
	cmd.Start()

	done := make(chan error, 1) // buffer to prevent send blocking and leaking a goroutine
	go func() { done <- cmd.Wait() }()

	// This select statement will block until one of these two conditions are met:
	// - The testlet finishes, in which case the channel "done" will be receive a value
	// - The configured time limit is exceeded (expected for testlets running in server mode)
	select {
	// Testlet finished
	case err := <-done:
		if err != nil {
			log.Errorf("Testlet %s completed with error: %v", path, err)

			// TODO(mierdin): Handling testrun errors is on my plate as it is, and so
			// this should get addressed properly in the future. The current approach
			// of adding "error" to this map does nothing, as the status is tracked elsewhere
			// gatheredData[thisTarget] = "error"
		}

	// Timeout
	case <-time.After(time.Duration(t.TimeLimit) * time.Second):
		// Sending a kill signal to the process group as a negative number
		// will ensure child process are killed.
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
			log.Errorf("Failed to kill %s after timeout: %s", path, err)
			return
		}

		log.Debug("Successfully killed ", path)
		return
	}

	log.Debugf("Testlet %s completed without error", path)

	// Record test data
	r.data = json.RawMessage(out.Bytes())
}
