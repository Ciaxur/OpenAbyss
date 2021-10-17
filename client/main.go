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
	"path/filepath"
	"regexp"
	"time"

	"openabyss/client/configuration"
	pb "openabyss/proto/server"
	"openabyss/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	// Init
	args := ParseArguments()

	if args.Verbose {
		configuration.EnableVerbose()
	}

	configuration.Init() // Client Config

	grpcHost := configuration.LoadedConfig.GrpcHost
	grpcPort := configuration.LoadedConfig.GrpcPort
	address := fmt.Sprintf("%s:%d", grpcHost, grpcPort)
	var conn *grpc.ClientConn
	var conn_err error = nil

	// Init TLS Credentials
	if !configuration.LoadedConfig.Insecure {
		creds, err := credentials.NewClientTLSFromFile("cert/server.crt", "")
		if err != nil {
			log.Fatalln("tls could not be read from file:", err)
		}

		conn, conn_err = grpc.Dial(
			address,
			grpc.WithTransportCredentials(creds),
		)
	} else {
		conn, conn_err = grpc.Dial(
			address,
			grpc.WithInsecure(),
		)
	}

	if conn_err != nil {
		log.Fatalf("did not connect: %v", conn_err)
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
	} else if args.ListBackups {
		// Issue request & handle response
		if resp, err := client.ListInternalBackups(ctx, &pb.EmptyMessage{}); err != nil {
			utils.HandleErr(err, "list backups error")
			os.Exit(1)
		} else {
			if len(resp.Backups) == 0 {
				log.Println("No existng backups")
			} else {
				log.Printf("%d backup entries found:\n", len(resp.Backups))
				for idx, elt := range resp.Backups {
					// Construct Expiration Time
					created_at := time.UnixMilli(int64(elt.CreatedUnixTimestamp))
					expires_at := time.Now().Add(time.Millisecond * time.Duration(elt.ExpiresInUnixTimestamp))

					log.Printf("[%d]: %s\n", idx, elt.FileName)
					log.Println("  - Created at: ", created_at.Local().String())
					log.Println("  - Expires at: ", expires_at.Local().String())
				}
			}
		}
	} else if args.InvokeBackup {
		// Issue backup invoke
		if resp, err := client.InvokeNewStorageBackup(ctx, &pb.EmptyMessage{}); err != nil {
			utils.HandleErr(err, "invoke new backup error")
			os.Exit(1)
		} else {
			// Construct Expiration Time
			expires_at := time.Now().Add(time.Millisecond * time.Duration(resp.ExpiresInUnixTimestamp))

			log.Println("Successfuly backed up internal storage")
			log.Println("  - Backup Filename: ", resp.FileName)
			log.Println("  - Backup Expires at: ", expires_at.Local().String())
		}
	} else if args.GetBackupManagerStatus {
		if resp, err := client.GetBackupManagerConfig(ctx, &pb.EmptyMessage{}); err != nil {
			utils.HandleErr(err, "get backup manager config error")
			os.Exit(1)
		} else {
			lastBackup := time.UnixMilli(int64(resp.LastBackupUnixTimestamp))
			backup_freq := time.UnixMilli(int64(resp.BackupFrequency))
			retention_period := time.UnixMilli(int64(resp.RetentionPeriod))

			log.Println("Backup Manager Configuration:")
			fmt.Printf("IsEnabled: %v\n", resp.IsEnabled)
			fmt.Printf("Total Backups: %d\n", resp.TotalBackups)

			if lastBackup.UnixMilli() == 0 {
				fmt.Println("Last Backup: NONE")
			} else {
				fmt.Printf("Last Backup: %s\n", lastBackup.Local().String())
			}

			fmt.Printf("Backup Frequency: %s\n", time.Duration(backup_freq.UnixNano()).String())
			fmt.Printf("Retention Period: %s\n", time.Duration(retention_period.UnixNano()).String())
		}
	} else if args.ToggleBackupManager {
		// Get current config
		resp, err := client.GetBackupManagerConfig(ctx, &pb.EmptyMessage{})
		if err != nil {
			utils.HandleErr(err, "could not get current backup manager's from server")
			os.Exit(1)
		}

		if resp, err := client.SetBackupManagerConfig(ctx, &pb.BackupManagerStatus{
			IsEnabled:       !resp.IsEnabled,
			RetentionPeriod: resp.RetentionPeriod,
			BackupFrequency: resp.BackupFrequency,
		}); err != nil {
			utils.HandleErr(err, "could not update backup manager's config")
			os.Exit(1)
		} else {
			log.Printf("Successfuly set Backup Manager to: %v\n", resp.IsEnabled)
		}
	} else if args.SetBackupRetention.Milliseconds() > 0 {
		// Get current config
		resp, err := client.GetBackupManagerConfig(ctx, &pb.EmptyMessage{})
		if err != nil {
			utils.HandleErr(err, "could not get current backup manager's from server")
			os.Exit(1)
		}

		if _, err := client.SetBackupManagerConfig(ctx, &pb.BackupManagerStatus{
			IsEnabled:       resp.IsEnabled,
			RetentionPeriod: uint64(args.SetBackupRetention.Milliseconds()),
			BackupFrequency: resp.BackupFrequency,
		}); err != nil {
			utils.HandleErr(err, "could not update backup manager's config")
			os.Exit(1)
		} else {
			log.Printf("Successfuly updated Backup Retention Period to: %v\n", args.SetBackupRetention.String())
		}
	} else if args.SetBackupFrequency.Milliseconds() > 0 {
		// Get current config
		resp, err := client.GetBackupManagerConfig(ctx, &pb.EmptyMessage{})
		if err != nil {
			utils.HandleErr(err, "could not get current backup manager's from server")
			os.Exit(1)
		}

		if _, err := client.SetBackupManagerConfig(ctx, &pb.BackupManagerStatus{
			IsEnabled:       resp.IsEnabled,
			RetentionPeriod: resp.RetentionPeriod,
			BackupFrequency: uint64(args.SetBackupFrequency.Milliseconds()),
		}); err != nil {
			utils.HandleErr(err, "could not update backup manager's config")
			os.Exit(1)
		} else {
			log.Printf("Successfuly updated Backup Frequency to: %v\n", args.SetBackupFrequency.String())
		}
	} else if len(args.RemoveBackup) > 0 {
		if resp, err := client.DeleteBackup(ctx, &pb.BackupEntryRequest{
			BackupFileName: args.RemoveBackup,
		}); err != nil {
			utils.HandleErr(err, "failed to remove backup")
			os.Exit(1)
		} else {
			log.Printf("Successfully removed \"%s\"\n", resp.FileName)
		}
	} else if len(args.ExportBackup) > 0 {
		// Requires file path to export TO
		if len(args.FilePath) == 0 {
			log.Fatalln("Export Backup requires a filepath to export to!")
		}

		// Request export
		if resp, err := client.ExportBackup(ctx, &pb.BackupEntryRequest{
			BackupFileName: args.ExportBackup,
		}); err != nil {
			log.Fatalln("Export Backup Error:", err)
		} else {
			// Write received file bytes to file
			if err := ioutil.WriteFile(args.FilePath, resp.FileData, 0664); err != nil {
				log.Fatalln("Error writing received backup to file:", err)
			} else {
				log.Printf("Successfuly export '%s' -> '%s'\n", resp.FileName, args.FilePath)
			}
		}
	} else if len(args.ImportBackup) > 0 {
		// Read in file import
		fileBuffer, err := os.ReadFile(args.ImportBackup)
		if err != nil {
			log.Fatalln("Error reading in file:", err)
		}

		// Issue import request
		if _, err := client.ImportBackup(ctx, &pb.ImportBackupRequest{
			FileName: filepath.Base(args.ImportBackup),
			FileData: fileBuffer,
		}); err != nil {
			log.Fatalln("Failed to import backup:", err)
		} else {
			log.Printf("Successfuly imported '%s'!\n", args.ImportBackup)
		}
	} else if len(args.RestoreFromBackup) > 0 {
		if resp, err := client.RestoreFromBackup(ctx, &pb.RestoreFromBackupRequest{
			FileName: args.RestoreFromBackup,
		}); err != nil {
			log.Fatalln("Failed to restore from backup:", err)
		} else {
			expires_at := time.Now().Add(time.Millisecond * time.Duration(resp.ExpiresInUnixTimestamp))

			log.Printf("Successfully restored from backup. Backup up previous storage'%s'\n", resp.FileName)
			log.Println("  - Expires at: ", expires_at.Local().String())
		}
	}
}
