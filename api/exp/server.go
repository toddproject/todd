/*
    ToDD API

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"net"

	// TODO need to fix this, I think it's part of stdlib now so you could kill all vendored deps for this
	"google.golang.org/grpc"

	log "github.com/Sirupsen/logrus"
	"github.com/toddproject/todd/persistence"

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

	s := grpc.NewServer()
	pb.RegisterGroupsServer(s, &server{
		cfg:         cfg,
		persistence: p,
	})

	// log.Infof("Serving ToDD API at: %s\n", serveURL)

	defer s.Stop()
	return s.Serve(lis)

}

type server struct {
	groups      []*pb.Group
	agents      map[string]*pb.Agent
	persistence *persistence.Persistence
	cfg         *config.ToDDConfig
}
