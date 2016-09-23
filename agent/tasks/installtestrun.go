/*
	ToDD task - test run

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package tasks

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/Mierdin/todd/agent/cache"
	"github.com/Mierdin/todd/agent/defs"
	"github.com/Mierdin/todd/agent/testing"
	"github.com/Mierdin/todd/config"
)

// InstallTestRunTask defines this particular task.
type InstallTestRunTask struct {
	BaseTask
	Config config.Config `json:"-"`
	Tr     defs.TestRun  `json:"testrun"`
}

// Run contains the logic necessary to perform this task on the agent.
// This particular task will install, but not execute a testrun on this agent.
// The installation procedure will first run the referenced testlet in check mode
// to help ensure that the actual testrun execution can take place. If that
// succeeds, the testrun is installed in the agent cache.
func (itt InstallTestRunTask) Run() error {

	if itt.Tr.Testlet == "" {
		log.Error("Testlet parameter for this testrun is null")
		return errors.New("Testlet parameter for this testrun is null")
	}

	// Determine if this is a native testlet
	testletPath, err := testing.GetTestletPath(itt.Tr.Testlet, itt.Config.LocalResources.OptDir)
	if err != nil {
		return err
	}

	// Run the testlet in check mode to verify that everything is okay to run this test
	log.Debug("Running testlet in check mode: ", testletPath)
	cmd := exec.Command(testletPath, "check")

	// Stdout buffer
	cmdOutput := &bytes.Buffer{}
	// Attach buffer to command
	cmd.Stdout = cmdOutput
	// Execute collector
	cmd.Run()

	// This is probably the best cross-platform way to see if check mode passed.
	if strings.Contains(string(cmdOutput.Bytes()), "Check mode PASSED") {
		log.Debugf("Check mode for %s passed", testletPath)
	} else {
		log.Error("Testlet returned an error during check mode: ", string(cmdOutput.Bytes()))
		return errors.New("Testlet returned an error during check mode")
	}

	// Insert testrun into agent cache
	var ac = cache.NewAgentCache(itt.Config)
	err = ac.InsertTestRun(itt.Tr)
	if err != nil {
		log.Error(err)
		return errors.New("Problem installing test run into agent cache")
	}

	return nil
}
