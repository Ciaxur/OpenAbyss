package main

import (
	"context"
	"log"
	"time"

	pb "openabyss/proto/server"

	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "OpenAbyss-Client"
)

func main() {
	args := ParseArguments()

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewOpenAbyssClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.GetKeyNames(ctx, &pb.EmptyRequest{})
	if err != nil {
		log.Fatalf("could no get names: %v", err)
	}
	log.Println("Get Names Response:", resp)
}
