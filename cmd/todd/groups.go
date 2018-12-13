/*
   ToDD API Client

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	yaml "gopkg.in/yaml.v2"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/toddproject/todd/api/exp/generated"
)

func ListGroups(conn *grpc.ClientConn) ([]*pb.Group, error) {

	defer conn.Close()
	client := pb.NewGroupsClient(conn)

	groupList, err := client.ListGroups(context.Background(), &empty.Empty{})
	if err != nil {
		return nil, err
	}

	return groupList.Groups, nil

}

func GetGroup(conn *grpc.ClientConn, groupName string) error {
	fmt.Printf("NOT IMPLEMENTED - would have retrieved group %s\n", groupName)
	return nil
}

func DeleteGroup(conn *grpc.ClientConn, groupName string) error {
	fmt.Printf("NOT IMPLEMENTED - would have deleted group %s\n", groupName)
	return nil
}

func CreateGroup(conn *grpc.ClientConn, group *pb.Group) error {

	defer conn.Close()
	client := pb.NewGroupsClient(conn)

	_, err := client.CreateGroup(context.Background(), group)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// marshalGroupFromFile creates a new Group instance from a file definition
func marshalGroupFromFile(absPath string) (*pb.Group, error) {
	yamlDef, _ := getYAMLDef(absPath)

	var groupObj *pb.Group
	err := yaml.Unmarshal(yamlDef, &groupObj)
	if err != nil {
		return nil, err
	}

	return groupObj, nil
}
