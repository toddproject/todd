package api

import (
	"context"

	pb "github.com/toddproject/todd/api/exp/generated"
)

func (s *server) AgentRegister(ctx context.Context, atr *pb.AgentTaskResult) (*pb.AgentTaskRequest, error) {

	// Perform grouping once, right here. None of this background process BS.
	// And no keepalives either. Let the long-running grpc streams handle this at the transport layer. Mark agents alive/not alive based on their connectivity.

	return nil, nil
}
