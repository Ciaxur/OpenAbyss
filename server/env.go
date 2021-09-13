package main

import pb "openabyss/proto/server"

type openabyss_server struct {
	pb.UnimplementedOpenAbyssServer
}

const (
	port = ":50051"
)
