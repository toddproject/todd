/*
   ToDD API - manages todd objects

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/toddproject/todd/api/exp/generated"

	log "github.com/Sirupsen/logrus"
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

	return &pb.GroupResponse{Id: newGroup.Id, Success: true}, nil
}

func (s *server) ListGroups(ctx context.Context, _ *empty.Empty) (*pb.GroupList, error) {

	groups, err := s.persistence.ListGroups()
	if err != nil {
		return nil, err
	}

	log.Warn(groups)

	return &pb.GroupList{Groups: groups}, nil
}
