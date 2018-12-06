/*
    ToDD agent grouping mechanism

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package grouping

import (
	"net"
	"regexp"

	log "github.com/Sirupsen/logrus"

	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/agent/tasks"
	"github.com/toddproject/todd/comms"
	"github.com/toddproject/todd/config"
	"github.com/toddproject/todd/db"
	"github.com/toddproject/todd/server/objects"
)

// CalculateGroups is a function designed to ingest a list of group objects, a collection of agents, and
// return a map that contains the resulting group for each agent UUID.
func CalculateGroups(cfg config.Config) {

	tdb, err := db.NewToddDB(cfg)
	if err != nil {
		log.Fatalf("Error connecting to DB: %v", err)
	}

	// Retrieve all currently active agents
	agents, err := tdb.GetAgents()
	if err != nil {
		log.Fatalf("Error retrieving agents: %v", err)
	}

	// Retrieve all objects with type "group"
	groupObs, err := tdb.GetObjects("group")
	if err != nil {
		log.Fatalf("Error retrieving groups: %v", err)
	}

	// Cast retrieved slice of ToddObject interfaces to actual GroupObjects
	groups := make([]objects.GroupObject, len(groupObs))
	for i, gobj := range groupObs {
		groups[i] = gobj.(objects.GroupObject)
	}

	// groupmap contains the uuid-to-groupname mappings to be used for test runs
	groupmap := map[string]string{}

	// This slice will hold all of the agents that are sad because they didn't get into a group
	var lonelyAgents []defs.AgentAdvert

next:
	for x := range agents {

		for i := range groups {

			// See if this agent is in this group
			if isInGroup(groups[i].Spec.Matches, agents[x].Facts) {

				// Insert this group name ("Label") into groupmap under the key of the UUID for the agent that belongs to it
				log.Debugf("Agent %s is in group %s\n", agents[x].UUID, groups[i].Label)
				groupmap[agents[x].UUID] = groups[i].Label
				continue next

			}
		}

		// The "continue next" should prohibit all agents that have a group from getting to this point,
		// so the only ones left do not have a group.
		lonelyAgents = append(lonelyAgents, agents[x])

	}

	// Write results to database
	err = tdb.SetGroupMap(groupmap)
	if err != nil {
		log.Fatalf("Error setting group map: %v", err)
	}

	// Send notifications to each agent to let them know what group they're in, so they can cache it
	tc, err := comms.NewToDDComms(cfg)
	if err != nil {
		log.Fatalf("Error setting up ToDD Comms during group calculation")
	}

	for uuid, groupName := range groupmap {
		tc.Package.SendTask(uuid, tasks.NewSetGroup(groupName))
	}

	// need to send a message to all agents that weren't in groupmap to set their group to nothing
	for x := range lonelyAgents {
		tc.Package.SendTask(lonelyAgents[x].UUID, tasks.NewSetGroup(""))
	}
}

// isInGroup takes a set of match statements (typically present in a group object definition) and a map of a single agent's facts,
// and returns True if one of the match statements validated against this map of facts. In short, this function can tell you if an agent
// is in a given group. This means that ToDD stops at the first successful match.
//
// This function is currently written to statically provide two mechanisms for matching:
//
// - hostname
// - within_subnet
//
// In the future, this functionality will be made much more generic, in order to take advantage of any user-defined fact collectors.
func isInGroup(matchStatements []map[string]string, factmap map[string][]string) bool {

	// Iterate over the match statements
	for x := range matchStatements {

		// If a "hostname" statement is present, perform a check
		if _, ok := matchStatements[x]["hostname"]; ok {

			// Continue if no "Hostname" fact present
			if _, ok := factmap["Hostname"]; !ok {
				continue
			}

			exp, err := regexp.Compile(matchStatements[x]["hostname"])
			if err != nil {
				log.Warn("Unable to compile provided regular expression in group object")
				continue
			}

			// retrieved hostname from system
			//
			// TODO(mierdin): Currently, hostname is provided as single
			// string in slice, so we can go straight to zero index.
			// This is due to crappy implementation in the fact collecting
			// process, so this is an acceptable temporary workaround.
			hostname := factmap["Hostname"][0]

			if exp.Match([]byte(hostname)) {
				return true
			}

		}

		// If a "within_subnet" statement is present, perform a check
		if _, ok := matchStatements[x]["within_subnet"]; ok {

			// Continue if no "Addresses" fact present
			if _, ok := factmap["Addresses"]; !ok {
				continue
			}

			thisSubnet := matchStatements[x]["within_subnet"]

			// First, we retrieve the desired subnet from the grouping object, and convert to net.IPNet
			_, desiredNet, err := net.ParseCIDR(thisSubnet)
			if err != nil {
				log.Errorf("Problem parsing desired network in grouping object: %q", thisSubnet)
			}

			addresses := factmap["Addresses"]

			for y := range addresses {

				if desiredNet.Contains(net.ParseIP(addresses[y])) {
					return true
				}
			}

		}

	}

	return false
}
