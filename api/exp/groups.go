/*
   ToDD API - manages todd objects

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"context"

	log "github.com/Sirupsen/logrus"

	pb "github.com/toddproject/todd/api/exp/generated"
)

// CreateGroups creates a new Group
func (s *server) CreateGroup(ctx context.Context, in *pb.Group) (*pb.GroupResponse, error) {
	s.groups = append(s.groups, in)
	return &pb.GroupResponse{Id: in.Id, Success: true}, nil
}

// GetGroups returns all groups by given filter
func (s *server) GetGroups(ctx context.Context, f *pb.GroupFilter) (*pb.GroupList, error) {

	objectList, err := s.tdb.GetObjects("group")
	if err != nil {
		log.Errorln(err)
	}

	log.Debug(objectList)

	theseGroups := []*pb.Group{
		{Id: 1, Name: "foobar", Matches: []*pb.Group_Match{}},
	}

	return &pb.GroupList{Groups: theseGroups}, nil

	// type Group struct {
	// 	Id        int32          `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	// 	Groupname string         `protobuf:"bytes,2,opt,name=groupname" json:"groupname,omitempty"`
	// 	Matches   []*Group_Match `protobuf:"bytes,3,rep,name=matches" json:"matches,omitempty"`
	// }

}
