/*
	ToDD task - set group

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

// SetGroupTask defines this particular task.
type SetGroupTask struct {
	BaseTask
	Config    config.ToDDConfig `json:"-"`
	GroupName string            `json:"groupName"`
}

// TODO (mierdin): Could this not be condensed with the generic "keyvalue" task?

// Run contains the logic necessary to perform this task on the agent.
func (sgt SetGroupTask) Run(ac *cache.AgentCache) error {
	// First, see what the current group is. If it matches what this task is instructing, we don't need to do anything.
	groupName, err := ac.GetKeyValue("group")
	if err != nil {
		return err
	}

	if groupName != sgt.GroupName {
		err := ac.SetKeyValue("group", sgt.GroupName)
		if err != nil {
			return fmt.Errorf("Failed to set keyvalue pair - %s:%s", "group", sgt.GroupName)
		}
		err = ac.SetKeyValue("unackedGroup", "true")
		if err != nil {
			return fmt.Errorf("Failed to set keyvalue pair - %s:%s", "unackedGroup", "true")
		}
	} else {
		log.Info("Already in the group being dictated by the server: ", sgt.GroupName)
	}

	return nil
}
