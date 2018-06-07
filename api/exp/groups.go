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
	"github.com/toddproject/todd/server/objects"
)

// CreateGroup creates a new Group
func (s *server) CreateGroup(ctx context.Context, newGroup *pb.Group) (*pb.GroupResponse, error) {

	// If we wanted to hold them in memory, this might be useful
	// s.groups = append(s.groups, in)

	err := s.tdb.CreateGroup(newGroup)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}

	return &pb.GroupResponse{Id: newGroup.Id, Success: true}, nil
}

// GetGroups returns all groups by given filter
func (s *server) GetGroups(ctx context.Context, f *pb.GroupFilter) (*pb.GroupList, error) {

	groups, err := s.tdb.GetGroups()
	if err != nil {
		return nil, err
	}

	return &pb.GroupList{Groups: groups}, nil

	_, err = s.tdb.GetObjects("group")
	if err != nil {
		log.Errorln(err)
		return nil, err
	}

	// s.convertGroups(objectList)

	// "[{{blue group} {blue [map[hostname:todd-agent-.*[13579]]]}}]"

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

// TODO it may just be easier to find the right way to store these and put these additional functions in the DB
// layer and bypass the current way of doing things entirely
func (s *server) convertGroups(oldObjects []objects.GroupObject) {
	for i := range oldObjects {
		log.Debug(oldObjects[i].GetSpec())
		// {blue [map[hostname:todd-agent-.*[13579]]]}
		// Group   string              `json:"group" yaml:"group"`
		// Matches []map[string]string `json:"matches" yaml:"matches"`

		// groupName := oldObjects[i].Spec.Group
		// groupMatches := oldObjects[i].GetSpec()
	}
}
