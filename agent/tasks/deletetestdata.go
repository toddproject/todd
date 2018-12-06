/*
	ToDD task - delete testrun data from cache

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package tasks

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/toddproject/todd/agent/cache"
	"github.com/toddproject/todd/config"
)

// DeleteTestDataTask embeds BaseTask and adds the necessary fields to transport a
// DeleteTestData task through comms
type DeleteTestDataTask struct {
	BaseTask
	Config   config.ToDDConfig `json:"-"`
	TestUUID string            `json:"key"`
}

// Run contains the logic necessary to perform this task on the agent.
func (dtdt DeleteTestDataTask) Run(ac *cache.AgentCache) error {
	err := ac.DeleteTestRun(dtdt.TestUUID)
	if err != nil {
		return fmt.Errorf("DeleteTestDataTask failed - %s", dtdt.TestUUID)
	}
	log.Infof("DeleteTestDataTask successful - %s", dtdt.TestUUID)

	return nil
}
