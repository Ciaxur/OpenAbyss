package main

import pb "openabyss/proto/server"

type openabyss_server struct {
	pb.UnimplementedOpenAbyssServer
}

var (
	port     = uint16(50051)
	host     = "0.0.0.0"
	insecure = false // Secure by default
	tlsCert  = "cert/server.crt"
	tlsKey   = "cert/server.key"
)
