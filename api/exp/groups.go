/*
   ToDD API - manages todd objects

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"context"
	"errors"
	"fmt"

	"net"
	"regexp"

	log "github.com/sirupsen/logrus"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/toddproject/todd/api/exp/generated"
)

func (s *server) CreateGroup(ctx context.Context, newGroup *pb.Group) (*pb.GroupResponse, error) {

	err := newGroup.Validate()
	if err != nil {
		return nil, err
	}

	err = s.persistence.CreateGroup(newGroup)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}

	// TODO(mierdin): should we trigger a re-evaluation of agents' groups here? Should there be an endpoint for resetting group memberships?

	return &pb.GroupResponse{Id: newGroup.Id, Success: true}, nil
}

func (s *server) ListGroups(ctx context.Context, _ *empty.Empty) (*pb.GroupList, error) {

	groups, err := s.persistence.ListGroups()
	if err != nil {
		return nil, err
	}

	return &pb.GroupList{Groups: groups}, nil
}

func (s *server) assignGroup(agent *pb.Agent) (string, error) {

	// TODO check agent.Facts for nil value

	groups, err := s.persistence.ListGroups()
	if err != nil {
		return "", err
	}

	for i := range groups {

		// See if this agent is in this group
		if isInGroup(groups[i], agent) {
			return groups[i].Name, nil
		}
	}

	return "", errors.New(fmt.Sprintf("Unable to determine group for agent %d", agent.Id))
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
// func isInGroup(matchStatements []map[string]string, factmap map[string][]string) bool {
func isInGroup(group *pb.Group, agent *pb.Agent) bool {

	for x := range group.Matches {

		if group.Matches[x].Type == "hostname" {
			if agent.Facts.Hostname == "" {
				continue
			}

			exp, err := regexp.Compile(group.Matches[x].Statement)
			if err != nil {
				log.Warn("Unable to compile provided regular expression in group object")
				continue
			}

			if exp.Match([]byte(agent.Facts.Hostname)) {
				return true
			}
		}

		if group.Matches[x].Type == "subnet" {

			if len(agent.Facts.Addresses) == 0 {
				continue
			}

			thisSubnet := group.Matches[x].Statement

			// First, we retrieve the desired subnet from the grouping object, and convert to net.IPNet
			_, desiredNet, err := net.ParseCIDR(thisSubnet)
			if err != nil {
				log.Errorf("Problem parsing desired network in grouping object: %q", thisSubnet)
			}

			for y := range agent.Facts.Addresses {

				if desiredNet.Contains(net.ParseIP(agent.Facts.Addresses[y])) {
					return true
				}
			}

		}

	}

	return false
}
