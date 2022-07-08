package client

import (
	"fmt"
	"github.com/bytelang/kplayer/types/config"
	"google.golang.org/grpc"
)

func GrpcClientRequest(server *config.Server) (*grpc.ClientConn, error) {
	if !server.On {
		return nil, fmt.Errorf("rpc server not start up")
	}

	return grpc.Dial(fmt.Sprintf("%s:%d", server.Address, server.GrpcPort), grpc.WithInsecure())
}
