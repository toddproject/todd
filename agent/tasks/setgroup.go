/*
	ToDD task - set group

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package tasks

import (
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"

	"github.com/toddproject/todd/agent/cache"
	"github.com/toddproject/todd/config"
)

// SetGroup defines this particular task.
type SetGroup struct {
	BaseTask
	GroupName string `json:"groupName"`
}

// NewSetGroup returns a new SetGroup task.
func NewSetGroup(groupName string) *SetGroup {
	return &SetGroup{
		BaseTask:  BaseTask{Type: TypeKeyValue},
		GroupName: groupName,
	}
}

// TODO (mierdin): Could this not be condensed with the generic "keyvalue" task?

// Run contains the logic necessary to perform this task on the agent.
func (t *SetGroup) Run(_ *config.Config, ac *cache.AgentCache, _ Responder) error {
	// First, see what the current group is. If it matches what this task is instructing, we don't need to do anything.
	groupName, err := ac.GetKeyValue("group")
	if err != nil {
		return errors.Wrap(err, "retrieving existing group")
	}

	if groupName == t.GroupName {
		log.Info("Already in the group being dictated by the server: ", t.GroupName)
		return nil
	}

	err = ac.SetKeyValue("group", t.GroupName)
	if err != nil {
		return errors.Wrapf(err, "setting group to %q", t.GroupName)
	}

	err = ac.SetKeyValue("unackedGroup", "true")
	if err != nil {
		return errors.Wrap(err, "setting unackedGroup to true")
	}

	return nil
}
