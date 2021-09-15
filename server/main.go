package main

import (
	"context"
	"log"
	"net"
	pb "openabyss/proto/server"
	"openabyss/server/storage"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
)

// TODO:
func (s openabyss_server) ListPathContents(ctx context.Context, in *pb.ListPathContentRequest) (*pb.PathContent, error) {
	return &pb.PathContent{}, nil
}

func onSignalChannel_cleanup(sigChan chan os.Signal) {
	<-sigChan
	log.Println("Clean-up Signal Issued: Cleaning up...")

	log.Println("[Clean Up]: Closing up Internal Storage")
	if err := storage.Close(); err != nil {
		log.Println("[Clean Up]: Error closing up Internal Storage:", err)
	}

	os.Exit(0)
}

func main() {
	Init()

	// Register SIGINT listener
	sig_chan := make(chan os.Signal, 1)
	signal.Notify(sig_chan, syscall.SIGTERM, syscall.SIGINT)
	go onSignalChannel_cleanup(sig_chan)

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
