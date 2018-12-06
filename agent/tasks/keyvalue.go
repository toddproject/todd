/*
	ToDD task - set keyvalue pair in cache

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

// KeyValueTask defines this particular task.
type KeyValueTask struct {
	BaseTask
	Config config.ToDDConfig `json:"-"`
	Key    string            `json:"key"`
	Value  string            `json:"value"`
}

// Run contains the logic necessary to perform this task on the agent. This particular task
// will simply pass a key/value pair to the agent cache to be set
func (kvt KeyValueTask) Run(ac *cache.AgentCache) error {
	err := ac.SetKeyValue(kvt.Key, kvt.Value)
	if err != nil {
		return fmt.Errorf("KeyValueTask failed - %s:%s", kvt.Key, kvt.Value)
	}
	log.Infof("KeyValueTask successful - %s:%s", kvt.Key, kvt.Value)

	return nil
}
