/*
	ToDD task - delete testrun data from cache

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package tasks

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/toddproject/todd/agent/cache"
	"github.com/toddproject/todd/config"
)

// DeleteTestDataTask embeds BaseTask and adds the necessary fields to transport a
// DeleteTestData task through comms
type DeleteTestDataTask struct {
	BaseTask
	Config   config.Config `json:"-"`
	TestUuid string        `json:"key"`
}

// Run contains the logic necessary to perform this task on the agent.
func (dtdt DeleteTestDataTask) Run() error {

	var ac = cache.NewAgentCache(dtdt.Config)

	err := ac.DeleteTestRun(dtdt.TestUuid)
	if err != nil {
		return errors.New(fmt.Sprintf("DeleteTestDataTask failed - %s", dtdt.TestUuid))
	}
	log.Info(fmt.Sprintf("DeleteTestDataTask successful - %s", dtdt.TestUuid))

	return nil
}
