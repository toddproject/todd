/*
   ToDD API Client

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package main

import (
	"context"
	"flag"
	"fmt"

	"google.golang.org/grpc"
	yaml "gopkg.in/yaml.v2"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/toddproject/todd/api/exp/generated"
)

func ListGroups() ([]*pb.Group, error) {

	var (
		serverAddr = flag.String("server_addr", "127.0.0.1:50099", "The server address in the format of host:port")
	)

	// TODO(mierdin): Add security options
	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	client := pb.NewGroupsClient(conn)

	groupList, err := client.ListGroups(context.Background(), &empty.Empty{})
	if err != nil {
		return nil, err
	}

	return groupList.Groups, nil

}

func GetGroup(groupName string) error {
	fmt.Printf("NOT IMPLEMENTED - would have retrieved group %s\n", groupName)
	return nil
}

func DeleteGroup(groupName string) error {
	fmt.Printf("NOT IMPLEMENTED - would have deleted group %s\n", groupName)
	return nil
}

func CreateGroup(group *pb.Group) error {

	var (
		serverAddr = flag.String("server_addr", "127.0.0.1:50099", "The server address in the format of host:port")
	)

	// TODO(mierdin): Add security options
	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer conn.Close()
	client := pb.NewGroupsClient(conn)

	_, err = client.CreateGroup(context.Background(), group)
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
