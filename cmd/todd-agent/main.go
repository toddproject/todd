/*
	Primary entry point for ToDD Agent

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/credentials"

	pb "github.com/toddproject/todd/api/exp/generated"

	"github.com/toddproject/todd/config"
)

// Command-line Arguments
var argConfig string

func init() {

	flag.Usage = func() {
		fmt.Print(`Usage: todd-agent [OPTIONS] COMMAND [arg...]

    An extensible framework for providing natively distributed testing on demand

    Options:
      --config="/etc/todd/agent.cfg"          Absolute path to ToDD agent config file`, "\n\n")

		os.Exit(0)
	}

	flag.StringVar(&argConfig, "config", "/etc/todd/agent.cfg", "ToDD agent config file location")
	flag.Parse()

	// TODO(moswalt): Implement configurable loglevel in server and agent
	log.SetLevel(log.DebugLevel)
}

func main() {
	_, err := config.LoadConfigFromEnv()
	if err != nil {
		os.Exit(1)
	}

	log.Infof("ToDD agent (version %s) started.\n", config.BuildInfo["buildSha"])

	// Generate UUID for this agent.
	agentId := GenerateUUID()
	log.Debugf("UUID is %s", agentId)

	for {

		log.Info("Attempting to find server...")
		creds, _ := credentials.NewClientTLSFromFile("/Users/mierdin/Code/GO/src/github.com/toddproject/todd/scripts/todd-cert.pem", "")
		conn, err := grpc.Dial("127.0.0.1:50099", grpc.WithTransportCredentials(creds), grpc.WithBlock(), grpc.WithTimeout(15*time.Second))
		if err != nil {
			log.Error(err)
			continue
		}
		defer conn.Close()

		client := pb.NewAgentsClient(conn)
		stream, err := client.AgentRegister(context.Background())
		if err != nil {
			log.Error(err)
		}

		if err := stream.Send(&pb.AgentMessage{
			Agent: &pb.Agent{
				Id: agentId,
				Facts: &pb.AgentFacts{
					Hostname: "uraj",
				},
			},
		}); err != nil {
			log.Error(err)
		}

		log.Info("Connected and registered to todd-server.")

		// Create wait channel to detect disconnect of agent from this stream
		waitc := stream.Context().Done()

		// Asynchronously handle all additional messages from this agent
		go func() {
			for {
				in, err := stream.Recv()
				if err == io.EOF {
					log.Errorf("Detected EOF in stream: %s", err)
					return
				}
				if err != nil {
					log.Error(err)
					return
				}

				log.Debug(in)

				// if err := stream.Send(&pb.ServerMessage{}); err != nil {
				// 	log.Error("FOO3")
				// 	return
				// }

			}
		}()

		// Block until agent disconnects.
		<-waitc

		// Should detect server disconnect on this end as well, and reset to "searching" mode when disconnected
		// Also should think about how to handle test data that didn't get uploaded
		// Do GRPC messages get delivered reliably?
	}
}
