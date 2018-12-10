/*
   ToDD API Client

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package main

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/toddproject/todd/api/exp/generated"
	"google.golang.org/grpc/credentials"
)

func ListAgents() ([]*pb.Agent, error) {

	serverAddr := "127.0.0.1:50099"

	creds, err := credentials.NewClientTLSFromFile("/Users/mierdin/Code/GO/src/github.com/toddproject/todd/scripts/todd-cert.pem", "")
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to reach todd-server at %s\n", serverAddr))
	}
	defer conn.Close()
	client := pb.NewAgentsClient(conn)

	agentList, err := client.ListAgents(context.Background(), &empty.Empty{})
	if err != nil {
		return nil, err
	}

	return agentList.Agents, nil

}
