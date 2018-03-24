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
	"github.com/toddproject/todd/db"

	pb "github.com/toddproject/todd/api/exp/generated"
	"github.com/toddproject/todd/config"
)

type ToDDApiExp struct {
	cfg config.Config
	tdb db.DatabasePackage
}

const (
	port = ":50099"
)

func (tapi ToDDApiExp) Start(cfg config.Config) error {

	tapi.cfg = cfg

	tdb, err := db.NewToddDB(tapi.cfg)
	if err != nil {
		return err
	}

	tapi.tdb = tdb

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Errorf("failed to listen: %v", err)
	}
	// Creates a new gRPC server
	s := grpc.NewServer()
	pb.RegisterGroupsServer(s, &server{tdb: tdb})

	// log.Infof("Serving ToDD Server API at: %s\n", serveURL)

	defer s.Stop()
	return s.Serve(lis)

}

// server is used to implement customer.CustomerServer.
type server struct {
	groups []*pb.Group
	tdb    db.DatabasePackage
}
