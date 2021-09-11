package main

import (
	"context"
	"log"
	"net"
	pb "openabyss/proto/server"

	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

type openabyss_server struct {
	pb.UnimplementedOpenAbyssServer
}

func (s openabyss_server) GetKeyNames(ctx context.Context, in *pb.EmptyRequest) (*pb.GetKeyNamesResponse, error) {
	return &pb.GetKeyNamesResponse{
		Keys: []string{
			"key1", // Example
		},
	}, nil
}

func (s openabyss_server) ListPathContents(ctx context.Context, in *pb.ListPathContentRequest) (*pb.PathContent, error) {
	return &pb.PathContent{}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterOpenAbyssServer(s, openabyss_server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
