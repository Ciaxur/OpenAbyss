package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"time"

	pb "openabyss/proto/server"
	"openabyss/utils"

	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "OpenAbyss-Client"
)

func main() {
	args := ParseArguments()

	conn, err := grpc.Dial(
		address,
		grpc.WithInsecure(),
	)

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
		utils.HandleErr(err, "could no get keys")

		if err == nil {
			log.Printf("Response[%d]:\n", len(resp.Entities))
			for _, entry := range resp.Entities {
				log.Printf("== [%s] ==\n", entry.Name)
				log.Println(string(entry.PublicKeyName))
			}
		}
	} else if args.GetKeyNames {
		resp, err := client.GetKeyNames(ctx, &pb.EmptyMessage{})
		utils.HandleErr(err, "could no get names")
		if err == nil {
			log.Printf("Response[%d]:\n", len(resp.Keys))
			for idx, entryKey := range resp.Keys {
				log.Printf("%d) %s\n", idx, entryKey)
			}
		}
	} else if len(args.GenerateKeyPair) > 0 {
		resp, err := client.GenerateKeyPair(ctx, &pb.GenerateEntityRequest{
			Name: args.GenerateKeyPair,
		})
		utils.HandleErr(err, "could not generate keypair for given name")

		if err == nil {
			log.Println("Response:")
			log.Printf("Generated keypair for '%s':\n", resp.Name)
		}
	} else if args.EncryptFile {
		// Check: Required File Path
		if len(args.FilePath) == 0 {
			log.Fatalf("no given required file to encrypt. filepath argument is required")
		}

		if len(args.StoragePath) == 0 { // Validate acompaning output destination
			log.Fatal("no given required storage path argument")
		} else if !utils.PathExists(args.FilePath) { // Validate Path
			log.Fatalf("given path '%s' does not exist\n", args.FilePath)
		} else if len(args.KeyId) == 0 { // No given key to encrypt with
			log.Fatal("no given required keyId to use")
		} else { // Issue request
			// Read in the file
			if fileBytes, err := ioutil.ReadFile(args.FilePath); err != nil {
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
					FileName:    path.Base(args.FilePath),
					Options: &pb.FileOptions{
						StoragePath: args.StoragePath,
						KeyName:     args.KeyId,
						Overwrite:   args.Force,
					},
				})
				if err != nil {
					// Handle duplicate internal store file found
					isDuplicate := regexp.MustCompile("(?i)duplicte").MatchString(err.Error())
					if isDuplicate {
						log.Println("Duplicate stored file found. Use -force to overwrite")
					} else {
						utils.HandleErr(err, "failed to encrypt file")
					}

				} else {
					storedFilePath := path.Join(resp.FileStoragePath, resp.FileId)
					log.Printf("Encrypted '%s' -> '%s' successfuly!\n", args.FilePath, storedFilePath)
				}
			}
		}
	} else if args.DecryptFile {
		// Check: Required Key-Id Argument
		if len(args.KeyId) == 0 {
			log.Fatalln("KeyId is required to decrypt the file")
		}

		// Check: Required Internal File path
		if len(args.FilePath) == 0 {
			log.Fatalln("FilePath is required to decrypt the file")
		}

		// Issue request
		resp, err := client.DecryptFile(ctx, &pb.DecryptRequest{
			FilePath:       args.FilePath,
			PrivateKeyName: []byte(args.KeyId),
		})

		// Handle resposne
		if err != nil {
			utils.HandleErr(err, "could no decrypt file")
		} else {
			fileBuffer := make([]byte, resp.SizeInBytes)
			gReader, err := gzip.NewReader(bytes.NewBuffer(resp.FileBytes))
			if err != nil {
				utils.HandleErr(err, "gzip failed to extract data")
				os.Exit(1)
			}
			gReader.Read(fileBuffer)

			// Output to a file
			if len(args.FilePacketOutput) > 0 {
				log.Println("Response:")
				fmt.Printf("File Name: %s\n", resp.FileName)
				fmt.Printf("File Size in Bytes: %d Bytes\n", resp.SizeInBytes)

				if fd, err := os.Create(args.FilePacketOutput); err != nil {
					utils.HandleErr(err, "failed to create file")
				} else {
					fd.Write(fileBuffer)
					fd.Close()

					fmt.Println("Data saved to:", args.FilePacketOutput)
				}

			} else { // Output to stdout
				fmt.Print(string(fileBuffer))
			}

		}
	} else if args.RemoveFile {
		// Check: Required argument filepath
		if len(args.FilePath) == 0 {
			log.Fatalln("FilePath is required to remove it")
		}

		// Issue request
		_, err := client.ModifyEntity(ctx, &pb.EntityMod{
			FilePath: args.FilePath,
			Remove:   args.RemoveFile,
		})

		// Check status
		if err != nil {
			utils.HandleErr(err, "failed to modify entity")
			os.Exit(1)
		} else {
			log.Printf("Successfuly removed '%s'\n", args.FilePath)
		}
	} else if args.ListPath {
		if len(args.StoragePath) == 0 {
			log.Fatalln("Storage path argument missing")
		}

		// Issue request & handle response
		req := pb.ListPathContentRequest{
			Path:      args.StoragePath,
			Recursive: args.RecursivePath,
		}
		if resp, err := client.ListPathContents(ctx, &req); err != nil {
			utils.HandleErr(err, "list path error")
			os.Exit(1)
		} else {
			if len(resp.Content) > 0 {
				log.Println("Internal Storage Content:")
				for _, entry := range resp.Content {
					createdDate := time.Unix(int64(entry.CreatedUnixTimestamp), 0).Format(time.RFC822)
					modifiedDate := time.Unix(int64(entry.ModifiedUnixTimestamp), 0).Format(time.RFC822)

					log.Printf("[%s]: Created at '%s' | Last Modified at '%s'\n", entry.Path, createdDate, modifiedDate)
				}
			} else {
				log.Println("No internal content")
			}
		}
	}

}
