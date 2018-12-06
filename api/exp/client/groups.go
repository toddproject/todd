/*
   ToDD API Client

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"context"
	"flag"
	"fmt"

	// TODO need to fix this, I think it's part of stdlib now so you could kill all vendored deps for this
	"google.golang.org/grpc"

	pb "github.com/toddproject/todd/api/exp/generated"
)

// ListGroups retrieves all groups
func (capi APIExpClient) ListGroups(conf map[string]string) ([]*pb.Group, error) {

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

	groupList, err := client.GetGroups(context.Background(), &pb.GroupFilter{})
	if err != nil {
		return nil, err
	}

	return groupList.Groups, nil

}

// GetGroup retrieves a specific group by name
func (capi APIExpClient) GetGroup(groupName string) error {
	fmt.Printf("NOT IMPLEMENTED - would have retrieved group %s\n", groupName)
	return nil
}

// DeleteGroup deletes a specific group by name
func (capi APIExpClient) DeleteGroup(groupName string) error {
	fmt.Printf("NOT IMPLEMENTED - would have deleted group %s\n", groupName)
	return nil
}

// CreateGroup creates a new group
func (capi APIExpClient) CreateGroup(group *pb.Group) error {

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
