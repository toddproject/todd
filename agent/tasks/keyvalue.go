/*
	ToDD task - set keyvalue pair in cache

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

// KeyValueTask defines this particular task.
type KeyValueTask struct {
	BaseTask
	Config config.Config `json:"-"`
	Key    string        `json:"key"`
	Value  string        `json:"value"`
}

// Run contains the logic necessary to perform this task on the agent. This particular task
// will simply pass a key/value pair to the agent cache to be set
func (kvt KeyValueTask) Run() error {

	var ac = cache.NewAgentCache(kvt.Config)

	err := ac.SetKeyValue(kvt.Key, kvt.Value)
	if err != nil {
		return errors.New(fmt.Sprintf("KeyValueTask failed - %s:%s", kvt.Key, kvt.Value))
	}
	log.Info(fmt.Sprintf("KeyValueTask successful - %s:%s", kvt.Key, kvt.Value))

	return nil
}
