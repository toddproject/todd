package api

import (
	"context"
	"io"

	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	pb "github.com/toddproject/todd/api/exp/generated"
)

// agentInstance is used to bundle an instance of an agent with the stream
// used to communicate with it. We store this in memory so we can access each agent's
// properties, and it's communication, all in one place as an API server property
type agentInstance struct {
	agent  *pb.Agent
	stream *pb.Agents_AgentRegisterServer
}

// AgentRegister is responsible for receiving an initial Agent connection request, and maintaining that Agent's state within the in-memory
// map. It receives incoming messages in a goroutine, and the function is held open as long as the Agent connection remains active
func (s *server) AgentRegister(stream pb.Agents_AgentRegisterServer) error {

	// Perform grouping once, right here. None of this background process BS.
	//
	// Create wait channel to detect disconnect of agent from this stream
	waitc := stream.Context().Done()

	// Receive initial message from agent synchronously
	in, err := stream.Recv()
	if err == io.EOF {
		log.Error(err)
		return err
	}
	if err != nil {
		log.Error(err)
		return err
	}

	agentID := in.Agent.Id

	group, err := s.assignGroup(in.Agent)
	if err != nil {
		return err
	}

	in.Agent.Group = group

	// Register agent with in-memory map
	s.agentMut.Lock()
	s.agents[agentID] = &agentInstance{
		agent:  in.Agent,
		stream: &stream,
	}
	s.agentMut.Unlock()
	log.Infof("Agent %s has registered in group %s.", agentID, group)

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
			log.Infof("Received additional message from %s: %v", agentID, in)

			// if err := stream.Send(&pb.ServerMessage{}); err != nil {
			// 	log.Error("FOO3")
			// 	return
			// }

		}
	}()

	// Block until agent disconnects.
	<-waitc

	// Agent disconnected; clean it up
	if agentID != "" {
		log.Infof("Agent %s has disconnected.", agentID)
		s.agentMut.Lock()
		delete(s.agents, agentID)
		s.agentMut.Unlock()
	}

	return nil
}

func (s *server) GetAgent(context.Context, *pb.AgentFilter) (*pb.Agent, error) {
	return nil, nil
}

func (s *server) ListAgents(context.Context, *empty.Empty) (*pb.AgentList, error) {

	agents := []*pb.Agent{}

	for _, ai := range s.agents {
		agents = append(agents, ai.agent)
	}

	return &pb.AgentList{
		Agents: agents,
	}, nil
}
