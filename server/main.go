package main

import (
	"crypto/tls"
	"fmt"
	"io/fs"
	"log"
	"net"
	pb "openabyss/proto/server"
	"openabyss/server/configuration"
	"openabyss/server/storage"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
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

// Attempts to load TLS certificates from the Certificates Pool path first,
// if the path was specified, or loads TLS from a single TLS Keypair.
// Returns:
//  - TLS Credentials. (pool / single keypair)
//  - Ok State.
func loadTLSCredentials() (credentials.TransportCredentials, bool) {
	// Append tls certs into the TLS Pool.
	if tlsPoolPath != "" {
		log.Println("[server.main] Searching for certificates in ", tlsPoolPath)

		// Prepare TLS Keypair.
		tlsKeyPairMap := make(map[string]struct {
			CertPath string
			KeyPath  string
		})

		// Prepare Regex filename matching.
		cert_re_str := `(.*)cert\.pem$`
		cert_re := regexp.MustCompile(cert_re_str)
		key_re_str := `(.*)key\.pem$`
		key_re := regexp.MustCompile(key_re_str)

		// Find and store Certificate and Key filename into  keypair Map.
		err := filepath.Walk(tlsPoolPath, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				log.Printf("[server.main] TLS Pool PathWalk error: %v\n", err)
				return err
			}

			// Add matching certificate/key filepath to tls keypair map.
			if cert_match := cert_re.FindStringSubmatch(info.Name()); len(cert_match) > 0 {
				certFilePrefix := cert_match[1]
				if val, ok := tlsKeyPairMap[certFilePrefix]; ok {
					val.CertPath = path
					tlsKeyPairMap[certFilePrefix] = val
				} else {
					tlsKeyPairMap[certFilePrefix] = struct {
						CertPath string
						KeyPath  string
					}{
						CertPath: path,
					}
				}
			} else if key_match := key_re.FindStringSubmatch(info.Name()); len(key_match) > 0 {
				keyFilePrefix := key_match[1]
				if val, ok := tlsKeyPairMap[keyFilePrefix]; ok {
					val.KeyPath = path
					tlsKeyPairMap[keyFilePrefix] = val
				} else {
					tlsKeyPairMap[keyFilePrefix] = struct {
						CertPath string
						KeyPath  string
					}{
						KeyPath: path,
					}
				}
			}

			return nil
		})
		if err != nil {
			return nil, false
		}

		// Iterate over and add tls keypair to pool.
		tlsCertificates := []tls.Certificate{}
		log.Printf("[server.main] Adding the following keypair to the TLS Pool (%s) matching '%s' and '%s':\n", tlsPoolPath, cert_re_str, key_re_str)
		for _, val := range tlsKeyPairMap {
			log.Println(" -", val.CertPath)

			if serverCA, err := tls.LoadX509KeyPair(val.CertPath, val.KeyPath); err != nil {
				log.Printf("[server.main] Failed to load Server CA Keypair: %s | %s\n", val.CertPath, val.KeyPath)
			} else {
				tlsCertificates = append(tlsCertificates, serverCA)
			}
		}

		// No certs found to be added to pool. Revert to single keypair.
		if len(tlsCertificates) == 0 {
			log.Printf("[server.main] No certificates found in pool path, %s, reverting to single keypair resolve.\n", tlsPoolPath)
		} else {
			tlsConfig := &tls.Config{
				Certificates: tlsCertificates,
			}
			creds := credentials.NewTLS(tlsConfig)
			return creds, true
		}
	}

	// Create TLS Credentials from a single keypair.
	creds, err := credentials.NewServerTLSFromFile(tlsCert, tlsKey)
	if err != nil {
		log.Printf("[server.main] failed to create new server tls: %v", err)
		return nil, false
	}

	log.Printf("[server.main] TLS loaded (cert=%s) (key=%s)\n", tlsCert, tlsKey)
	return creds, true
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
		if creds, ok := loadTLSCredentials(); ok {
			s = grpc.NewServer(grpc.Creds(creds))
		} else { // Resolve to insecure on Failure
			log.Println("[server.main] Resolving to insecure server")
			s = grpc.NewServer()
		}
	}
	pb.RegisterOpenAbyssServer(s, openabyss_server{})
	log.Printf("[server.main] server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("[server.main] failed to serve: %v", err)
	}
}
