package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"io/ioutil"
	"log"
	"path"
	"time"

	pb "openabyss/proto/server"
	"openabyss/utils"

	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "OpenAbyss-Client"
)

func handleError(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

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

	// Client Reqeust
	if args.GetKeys {
		resp, err := client.GetKeys(ctx, &pb.EmptyMessage{})
		handleError(err, "could no get keys")

		if err == nil {
			log.Println("Response:")
			for _, entry := range resp.Entities {
				log.Printf("== [%s] ==\n", entry.Name)
				log.Println(string(entry.PublicKey))
			}
		}
	} else if args.GetKeyNames {
		resp, err := client.GetKeyNames(ctx, &pb.EmptyMessage{})
		handleError(err, "could no get names")
		if err == nil {
			log.Println("Response:")
			for idx, entryKey := range resp.Keys {
				log.Printf("%d) %s\n", idx, entryKey)
			}
		}
	} else if len(args.GenerateKeyPair) > 0 {
		resp, err := client.GenerateKeyPair(ctx, &pb.GenerateEntityRequest{
			Name: args.GenerateKeyPair,
		})
		handleError(err, "could not generate keypair for given name")

		if err == nil {
			log.Println("Response:")
			log.Printf("Generated keypair for '%s':\n", resp.Name)
			log.Println(resp.PublicKey)
		}
	} else if len(args.EncryptFile) > 0 {
		if len(args.StoragePath) == 0 { // Validate acompaning output destination
			log.Fatal("no given required storage path argument")
		} else if !utils.PathExists(args.EncryptFile) { // Validate Path
			log.Fatalf("given path '%s' does not exist\n", args.EncryptFile)
		} else if len(args.KeyId) == 0 { // No given key to encrypt with
			log.Fatal("no given required keyId to use")
		} else { // Issue request
			// Read in the file
			if fileBytes, err := ioutil.ReadFile(args.EncryptFile); err != nil {
				log.Fatalln("could not read in file:", err)
			} else {

				// Compress given data
				compBuffer := bytes.NewBuffer(nil)
				writer := gzip.NewWriter(compBuffer)
				writer.Write(fileBytes)
				writer.Close()

				resp, err := client.EncryptFile(ctx, &pb.FilePacket{
					FileBytes:   compBuffer.Bytes(),
					SizeInBytes: int64(compBuffer.Len()),
					FileName:    path.Base(args.EncryptFile),
					StoragePath: args.StoragePath,
					KeyName:     args.KeyId,
				})
				if err != nil {
					log.Fatalln("failed to encrypt file:", err)
				} else {
					storedFilePath := path.Join(resp.FileStoragePath, resp.FileId)
					log.Printf("Encrypted '%s' -> '%s' successfuly!\n", args.EncryptFile, storedFilePath)
				}
			}
		}
	}

}
