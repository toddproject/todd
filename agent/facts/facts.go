/*
   	Fact-Gathering functions

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package facts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/toddproject/todd/config"
)

// GetFacts is responsible for gathering facts on a system (runs on the agent).
// It does so by iterating over the collectors installed on this agent's system,
// executing them, and capturing their output. It will aggregate this output and
// return it all as a single map (keys are fact names)
func GetFacts(cfg config.Config) map[string][]string {

	retFactSet := make(map[string][]string)

	// this is the function that will do work on a single file during a walk
	execute_collector := func(path string, f os.FileInfo, err error) error {
		if f.IsDir() != true {
			cmd := exec.Command(path)

			// Stdout buffer
			cmdOutput := &bytes.Buffer{}
			// Attach buffer to command
			cmd.Stdout = cmdOutput
			// Execute collector
			cmd.Run()

			// We only care that the key is a string (this is the fact name)
			// The value for this key can be whatever
			fact := make(map[string][]string)

			// Unmarshal JSON into our fact map
			err = json.Unmarshal(cmdOutput.Bytes(), &fact)

			// We only expect a single key in the returned fact map. Only add to fact map if this is true.
			if len(fact) == 1 {
				for factName, factValue := range fact {
					retFactSet[factName] = factValue
					log.Debugf("Results from collector '%s': %s", factName, factValue)
				}
			}
		}
		return nil
	}

	// Perform above Walk function (execute_collector) on the collector directory
	err := filepath.Walk(fmt.Sprintf("%s/assets/factcollectors", cfg.LocalResources.OptDir), execute_collector)
	if err != nil {
		log.Error("Problem running fact-gathering collector")
	}

	return retFactSet
}
