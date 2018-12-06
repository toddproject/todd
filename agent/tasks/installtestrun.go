/*
	ToDD task - test run

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package tasks

import (
	"bytes"
	"os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"

	"github.com/toddproject/todd/agent/cache"
	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/agent/responses"
	"github.com/toddproject/todd/agent/testing"
	"github.com/toddproject/todd/config"
)

// InstallTestRun defines this particular task.
type InstallTestRun struct {
	BaseTask
	TR defs.TestRun `json:"test_run"`
}

// NewInstallTestRun returns a new InstallTestRun task.
func NewInstallTestRun(tr defs.TestRun) *InstallTestRun {
	return &InstallTestRun{
		BaseTask: BaseTask{Type: TypeInstallTestRun},
		TR:       tr,
	}
}

// Run contains the logic necessary to perform this task on the agent.
// This particular task will install, but not execute a testrun on this agent.
// The installation procedure will first run the referenced testlet in check mode
// to help ensure that the actual testrun execution can take place. If that
// succeeds, the testrun is installed in the agent cache.
func (t *InstallTestRun) Run(cfg *config.Config, ac *cache.AgentCache, responder Responder) (err error) {
	// Retrieve UUID
	uuid, err := ac.GetKeyValue("uuid")
	if err != nil {
		return errors.Wrap(err, "retrieving UUID")
	}

	// Send response regardless of how we return
	response := responses.NewSetAgentStatus(uuid, t.TR.UUID, testing.StatusReady)
	defer func() {
		if err != nil {
			response.Status = testing.StatusFail
		}
		responder(response)
	}()

	if t.TR.Testlet == "" {
		log.Error("Testlet parameter for this testrun is null")
		return errors.New("Testlet parameter for this testrun is null")
	}

	// Determine if this is a native testlet
	testletPath, err := testing.GetTestletPath(t.TR.Testlet, cfg.LocalResources.OptDir)
	if err != nil {
		return err
	}

	// Run the testlet in check mode to verify that everything is okay to run this test
	log.Debug("Running testlet in check mode:", testletPath)
	out, err := exec.Command(testletPath, "check").Output()
	if err != nil {
		return errors.Wrap(err, "executing testletin check mode")
	}

	// This is probably the best cross-platform way to see if check mode passed.
	if !bytes.Contains(out, []byte("Check mode PASSED")) {
		log.Error("Testlet returned an error during check mode: ", string(out))
		return errors.New("Testlet returned an error during check mode")
	}
	log.Debugf("Check mode for %s passed", testletPath)

	// Insert testrun into agent cache
	err = ac.InsertTestRun(t.TR)
	if err != nil {
		log.Error(err)
		return errors.Wrap(err, "installing testrun in cache")
	}

	return nil
}
