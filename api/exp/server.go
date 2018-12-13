/*
    ToDD API

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"net"
	"sync"

	// TODO need to fix this, I think it's part of stdlib now so you could kill all vendored deps for this
	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
	"github.com/toddproject/todd/persistence"
	"google.golang.org/grpc/credentials"

	pb "github.com/toddproject/todd/api/exp/generated"
	"github.com/toddproject/todd/config"
)

const (
	port = ":50099"
)

func StartAPI(cfg *config.ToDDConfig, p *persistence.Persistence) error {

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Errorf("failed to listen: %v", err)
	}

	creds, _ := credentials.NewServerTLSFromFile("/Users/mierdin/Code/GO/src/github.com/toddproject/todd/scripts/todd-cert.pem", "/Users/mierdin/Code/GO/src/github.com/toddproject/todd/scripts/todd-key.pem")
	s := grpc.NewServer(grpc.Creds(creds))

	apiServer := &server{
		cfg:         cfg,
		persistence: p,
		agentMut:    &sync.Mutex{},
		agents:      map[string]*agentInstance{},
	}

	pb.RegisterGroupsServer(s, apiServer)
	pb.RegisterAgentsServer(s, apiServer)

	// log.Infof("Serving ToDD API at: %s\n", serveURL)

	defer s.Stop()
	return s.Serve(lis)

}

type server struct {
	// Agents not stored in database, but rather kept in memory. So we need the map and a mutex
	agents   map[string]*agentInstance
	agentMut *sync.Mutex

	// Persistence used for all resource we persist to disk
	persistence *persistence.Persistence
	cfg         *config.ToDDConfig
}
