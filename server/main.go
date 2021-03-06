package main

import (
	"fmt"
	"log"
	"net"
	pb "openabyss/proto/server"
	"openabyss/server/configuration"
	"openabyss/server/storage"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func onSignalChannel_cleanup(sigChan chan os.Signal) {
	<-sigChan
	log.Println("[Clean Up] Clean-up Signal Issued: Cleaning up...")

	log.Println("[Clean Up]: Closing up Internal Storage")
	if err := storage.Close(); err != nil {
		log.Println("[Clean Up]: Error closing up Internal Storage:", err)
	}

	log.Println("[Clean Up]: Closing up Server Configuration")
	configuration.Close()

	os.Exit(0)
}

func main() {
	Init()

	// Register SIGINT listener
	sig_chan := make(chan os.Signal, 1)
	signal.Notify(sig_chan, syscall.SIGTERM, syscall.SIGINT)
	go onSignalChannel_cleanup(sig_chan)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatalf("[server.main] failed to listen: %v", err)
	}

	// Instantiate Server (Insecure/Secure)
	var s *grpc.Server
	if insecure {
		log.Println("[server.main] no TLS")
		s = grpc.NewServer()
	} else {
		// Create TLS Credentials
		creds, err := credentials.NewServerTLSFromFile(tlsCert, tlsKey)
		if err != nil {
			log.Fatalf("[server.main] failed to create new server tls: %v", err)
		}

		log.Printf("[server.main] TLS loaded (cert=%s) (key=%s)\n", tlsCert, tlsKey)
		s = grpc.NewServer(grpc.Creds(creds))
	}
	pb.RegisterOpenAbyssServer(s, openabyss_server{})
	log.Printf("[server.main] server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("[server.main] failed to serve: %v", err)
	}
}
