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
	"fmt"
	"os"
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

	var testlet_path string

	// Determine if this is a valid native testlet
	//_, err := testlets.NewTestlet(itt.Tr.Testlet)
	isNative, newName := testlets.IsNativeTestlet(itt.Tr.Testlet)
	// Not a native testlet - attempt to run check mode on testlet in filesystem
	if isNative {
		// Nothing to do, as we're using a native testlet
		log.Infof("%s is a native testlet - installing testrun.", itt.Tr.Testlet)
		//itt.Tr.Testlet = newName
		testlet_path = newName

	} else {
		// Generate path to testlet and make sure it exists.
		testlet_path = fmt.Sprintf("%s/assets/testlets/%s", itt.Config.LocalResources.OptDir, itt.Tr.Testlet)

		if _, err := os.Stat(testlet_path); os.IsNotExist(err) {
			log.Errorf("Testlet %s does not exist on this agent", itt.Tr.Testlet)
			return errors.New("Error installing testrun - testlet doesn't exist on this agent.")
		}
	}

	// Run the testlet in check mode to verify that everything is okay to run this test
	log.Debug("Running testlet in check mode: ", testlet_path)
	cmd := exec.Command(testlet_path, "check")

	// Stdout buffer
	cmdOutput := &bytes.Buffer{}
	// Attach buffer to command
	cmd.Stdout = cmdOutput
	// Execute collector
	cmd.Run()

	// This is probably the best cross-platform way to see if check mode passed.
	if strings.Contains(string(cmdOutput.Bytes()), "Check mode PASSED") {
		log.Debugf("Check mode for %s passed", testlet_path)
	} else {
		log.Error("Testlet returned an error during check mode: ", string(cmdOutput.Bytes()))
		return errors.New("Testlet returned an error during check mode")
	}

	// Insert testrun into agent cache
	var ac = cache.NewAgentCache(itt.Config)
	err := ac.InsertTestRun(itt.Tr)
	if err != nil {
		log.Error(err)
		return errors.New("Problem installing test run into agent cache")
	}

	return nil
}
