/*
	ToDD task - set keyvalue pair in cache

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

// KeyValue defines this particular task.
type KeyValue struct {
	BaseTask
	Key   string `json:"key"`
	Value string `json:"value"`
}

// NewKeyValue returns a new KeyValue task.
func NewKeyValue(key, value string) *KeyValue {
	return &KeyValue{
		BaseTask: BaseTask{Type: TypeKeyValue},
		Key:      key,
		Value:    value,
	}
}

// Run contains the logic necessary to perform this task on the agent. This particular task
// will simply pass a key/value pair to the agent cache to be set
func (t *KeyValue) Run(_ *config.Config, ac *cache.AgentCache, _ Responder) error {
	err := ac.SetKeyValue(t.Key, t.Value)
	if err != nil {
		return errors.Wrapf(err, "setting %q:%q", t.Key, t.Value)
	}

	log.Infof("KeyValueTask successful - %s:%s", t.Key, t.Value)
	return nil
}
