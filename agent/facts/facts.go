/*
   	Fact-Gathering functions

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package facts

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/toddproject/todd/config"
)

// GetFacts is responsible for gathering facts on a system (runs on the agent).
// It does so by iterating over the collectors installed on this agent's system,
// executing them, and capturing their output. It will aggregate this output and
// return it all as a single map (keys are fact names)
func GetFacts(cfg *config.ToDDConfig) (map[string][]string, error) {
	var scripts []string
	// Find all collector scripts
	dir := filepath.Join(cfg.OptDir, "assets", "factcollectors")
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			log.Warnf("Error gathering facts: %v (%s)", err, path)
			return nil // skip but continue walking
		}
		if !f.IsDir() {
			scripts = append(scripts, path)
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "finding fact collectors")
	}

	retFactSet := make(map[string][]string)

	// Execute collector scripts
	for _, script := range scripts {
		out, err := exec.Command(script).Output()
		if err != nil {
			return nil, errors.Wrapf(err, "executing %q", script)
		}

		// We only care that the key is a string (this is the fact name)
		// The value for this key can be whatever
		fact := make(map[string][]string)

		// Unmarshal JSON into our fact map
		err = json.Unmarshal(out, &fact)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing %q output", script)
		}

		// We only expect a single key in the returned fact map. Only add to fact map if this is true.
		if len(fact) != 1 {
			continue
		}

		for factName, factValue := range fact {
			retFactSet[factName] = factValue
			log.Debugf("Results from collector '%s': %s", factName, factValue)
		}
	}

	return retFactSet, nil
}
