/*
   ToDD API Client

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package main

import (
	"context"

	"google.golang.org/grpc"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/toddproject/todd/api/exp/generated"
)

func ListAgents(conn *grpc.ClientConn) ([]*pb.Agent, error) {

	defer conn.Close()
	client := pb.NewAgentsClient(conn)

	agentList, err := client.ListAgents(context.Background(), &empty.Empty{})
	if err != nil {
		return nil, err
	}

	return agentList.Agents, nil

}
