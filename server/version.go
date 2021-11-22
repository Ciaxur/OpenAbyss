package main

import (
	"context"
	"log"
	pb "openabyss/proto/server"
)

// TODO: Due for another update

// Returns Server's Version
func (s openabyss_server) GetServerVersion(ctx context.Context, in *pb.ServerVersionRequest) (*pb.ServerVersionResponse, error) {
	log.Printf("[GetServerVersion]: Server version requested: '%s'\n", version)
	return &pb.ServerVersionResponse{
		Version: version,
	}, nil
}
