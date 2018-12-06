/*
	ToDD task - delete testrun data from cache

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package tasks

import (
	log "github.com/Sirupsen/logrus"

	"github.com/toddproject/todd/agent/cache"
	"github.com/toddproject/todd/config"
)

// DeleteTestData embeds BaseTask and adds the necessary fields to transport a
// DeleteTestData task through comms
type DeleteTestData struct {
	BaseTask
	TestUUID string `json:"key"`
}

// NewDeleteTestData returns a new DeleteTestData tasks.
func NewDeleteTestData(uuid string) *DeleteTestData {
	return &DeleteTestData{
		BaseTask: BaseTask{Type: TypeDeleteTestData},
		TestUUID: uuid,
	}
}

// Run contains the logic necessary to perform this task on the agent.
func (t *DeleteTestData) Run(_ *config.Config, ac *cache.AgentCache, _ Responder) error {
	err := ac.DeleteTestRun(t.TestUUID)
	if err == nil {
		log.Infof("DeleteTestDataTask successful - %s", t.TestUUID)
	}
	return err
}
