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
	"strings"
	"time"

	"openabyss/client/configuration"
	"openabyss/client/console"
	pb "openabyss/proto/server"
	"openabyss/utils"

	"github.com/fatih/color"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Encapsulates references for Client
type ClientContext struct {
	pbClient pb.OpenAbyssClient
	ctx      context.Context
	args     *Arguments
}

// Helper function that prints entity details
func printEntity(entity *pb.Entity) {
	created_at := time.UnixMilli(int64(entity.CreatedUnixTimestamp))
	modified_at := time.UnixMilli(int64(entity.ModifiedUnixTimestamp))
	expires_at := time.UnixMilli(int64(entity.ExpiresAtUnixTimestamp + entity.CreatedUnixTimestamp))

	console.Heading.Printf("== [%s] ==\n", entity.Name)
	console.Log.Println("- Description: ", entity.Description)
	console.Log.Println("- Algorithm: ", entity.Algorithm)

	console.Log.Println("- Created on: ", created_at.Local())
	console.Log.Println("- Modified on: ", modified_at.Local())

	if entity.ExpiresAtUnixTimestamp != 0 {
		console.Log.Println("- Expires on: ", expires_at.Local())
	} else {
		console.Log.Println("- Expires on: ", "NEVER")
	}

	console.Log.Println("- Public Key:")
	console.Log.Println(string(entity.PublicKeyName))
}

// Subcommand-Handler: Generate
func handleKeysSubCmd(actions []string, context *ClientContext) {
	switch actions[0] {
	case "generate":
		resp, err := context.pbClient.GenerateKeyPair(context.ctx, &pb.GenerateEntityRequest{
			Name:                   *context.args.KeyPairName,
			Description:            *context.args.KeyPairDescription,
			Algorithm:              *context.args.KeyPairAlgo,
			ExpiresInUnixTimestamp: uint64(context.args.KeyExpiration.Milliseconds()),
		})
		utils.HandleErr(err, "could not generate keypair for given name")

		if err == nil {
			console.Heading.Printf("Generated keypair for '%s':\n", color.WhiteString(resp.Name))
		}
	case "modify":
		modifyKeyExpiration := false
		if context.args.KeyExpirationMod.Milliseconds() != 0 || *context.args.KeyExpirationDisableMod {
			modifyKeyExpiration = true
		}

		resp, err := context.pbClient.ModifyKeyPair(context.ctx, &pb.EntityModifyRequest{
			Name:                   *context.args.KeyPairNameMod,
			Description:            *context.args.KeyPairDescriptionMod,
			KeyId:                  *context.args.KeyIdMod,
			ModifyKeyExpiration:    modifyKeyExpiration,
			ExpiresInUnixTimestamp: uint64(context.args.KeyExpirationMod.Milliseconds()),
		})
		utils.HandleErr(err, "could not modify key details for given key-id")

		if err == nil {
			console.Heading.Printf("Key Details modified for '%s':\n", color.WhiteString(*context.args.KeyIdMod))
			printEntity(resp)
		}
	case "remove":
		resp, err := context.pbClient.RemoveKeyPair(context.ctx, &pb.EntityRemoveRequest{
			KeyId: *context.args.KeyIdRem,
		})
		utils.HandleErr(err, "could not remove key for given key-id")

		if err == nil {
			console.Heading.Printf("Key '%s' successfully removed:\n", color.WhiteString(*context.args.KeyIdRem))
			printEntity(resp)
		}
	case "import":
		// Read that gzip file
		filePath := *context.args.KeyImportFilePath
		if buffer, err := ioutil.ReadFile(filePath); err != nil {
			utils.HandleErr(err, console.Info.Sprintf("failed to read file '%s'\n", filePath))
		} else {
			if _, err := context.pbClient.ImportKey(context.ctx, &pb.KeyImportRequest{
				KeyId:   *context.args.KeyImportKeyId,
				KeyGzip: buffer,
				Force:   *context.args.Force,
			}); err != nil {
				utils.HandleErr(err, "request error")
			} else {
				console.Info.Printf("Successfuly imported '%s'\n", *context.args.KeyImportKeyId)
			}
		}
	case "export":
		if resp, err := context.pbClient.ExportKey(context.ctx, &pb.KeyExportRequest{
			KeyId: *context.args.KeyExportKeyId,
		}); err != nil {
			utils.HandleErr(err, "could not export key for given key-id")
		} else {
			if err := ioutil.WriteFile(*context.args.KeyExportFilePath, resp.KeyGzip, 0644); err != nil {
				console.Error.Printf("Failed to export key '%s' to destination '%s'\n", color.WhiteString(*context.args.KeyExportKeyId), color.WhiteString(*context.args.KeyExportFilePath))
			} else {
				console.Info.Println("Successsfully export to", color.WhiteString(*context.args.KeyExportFilePath))
			}
		}
	}
}

// Subcommand-Handler: List Keys
func handleListKeysSubCmd(actions []string, context *ClientContext) {
	// List key names only
	if *context.args.GetKeyNames {
		resp, err := context.pbClient.GetKeyNames(context.ctx, &pb.EmptyMessage{})
		utils.HandleErr(err, "could no get names")
		if err == nil {
			console.Info.Println("Total Keys: ", len(resp.Keys))
			for idx, entryKey := range resp.Keys {
				console.Heading.Printf("%d) ", idx)
				console.Log.Println(entryKey)
			}
		}
		return
	}

	// List Keys with metadata
	resp, err := context.pbClient.GetKeys(context.ctx, &pb.EmptyMessage{})
	utils.HandleErr(err, "could no get keys")
	if err == nil {
		// Log there are no keys if none are returned
		if len(resp.Entities) == 0 {
			console.Info.Println("No keys found. Consider generating one.")
			return
		}

		for _, entry := range resp.Entities {
			printEntity(entry)
		}
	}
}

// Subcommand-Handler: List Storage
func handleListStorageSubCmd(actions []string, context *ClientContext) {
	// Issue request & handle response
	req := pb.ListPathContentRequest{
		Path:      *context.args.ListStoragePath,
		Recursive: *context.args.RecursivePath,
	}
	if resp, err := context.pbClient.ListPathContents(context.ctx, &req); err != nil {
		utils.HandleErr(err, "list path error")
		os.Exit(1)
	} else {
		if len(resp.Content) > 0 {
			console.Heading.Println("Internal Storage Content:")
			for _, entry := range resp.Content {
				createdDate := time.Unix(int64(entry.CreatedUnixTimestamp), 0).Format(time.RFC822)
				modifiedDate := time.Unix(int64(entry.ModifiedUnixTimestamp), 0).Format(time.RFC822)

				console.Log.Printf("[%s]: Created at '%s' | Last Modified at '%s'\n", entry.Path, createdDate, modifiedDate)
			}
		} else {
			console.Warning.Println("No internal content")
		}
	}
}

// Subcommand-Handler: Encrypt
func handleEncryptSubCmd(actions []string, context *ClientContext) {
	if !utils.PathExists(*context.args.EncryptFile) { // Validate Path
		console.Fatalf("given path '%s' does not exist\n", *context.args.EncryptFile)
	} else if len(*context.args.EncryptKeyId) == 0 { // No given key to encrypt with
		console.Fatalln("no given required keyId to use")
	} else { // Issue request
		// Read in the file
		if fileBytes, err := ioutil.ReadFile(*context.args.EncryptFile); err != nil {
			console.Fatalln("could not read in file:", err)
		} else {
			// Compress given data
			compBuffer := bytes.NewBuffer(nil)
			writer := gzip.NewWriter(compBuffer)
			writer.Write(fileBytes)
			writer.Close()

			resp, err := context.pbClient.EncryptFile(context.ctx, &pb.FilePacket{
				FileBytes:   compBuffer.Bytes(),
				SizeInBytes: int64(len(fileBytes)),
				FileName:    path.Base(*context.args.EncryptFile),
				Options: &pb.FileOptions{
					StoragePath: *context.args.StoragePath,
					KeyName:     *context.args.EncryptKeyId,
					Overwrite:   *context.args.Force,
				},
			})
			if err != nil {
				// Handle duplicate internal store file found
				isDuplicate := regexp.MustCompile("(?i)duplicte").MatchString(err.Error())
				if isDuplicate {
					console.Warning.Println("Duplicate stored file found. Use --force to overwrite")
				} else {
					utils.HandleErr(err, "failed to encrypt file")
				}

			} else {
				storedFilePath := path.Join(resp.FileStoragePath, resp.FileId)
				console.Info.Printf("Encrypted '%s' -> '%s' successfuly!\n", *context.args.EncryptFile, storedFilePath)
			}
		}
	}
}

// Subcommand-Handler: Decrypt
func handleDecryptSubCmd(actions []string, context *ClientContext) {
	// Issue request
	resp, err := context.pbClient.DecryptFile(context.ctx, &pb.DecryptRequest{
		FilePath:       *context.args.DecryptFile,
		PrivateKeyName: []byte(*context.args.DecryptKeyId),
	})

	// Handle response
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
		if len(*context.args.FilePacketOutput) > 0 {
			console.Log.Printf("File Name: %s\n", resp.FileName)
			console.Log.Printf("File Size in Bytes: %d Bytes\n", resp.SizeInBytes)

			if fd, err := os.Create(*context.args.FilePacketOutput); err != nil {
				utils.HandleErr(err, "failed to create file")
			} else {
				fd.Write(fileBuffer)
				fd.Close()

				console.Log.Println("Data saved to:", *context.args.FilePacketOutput)
			}

		} else { // Output to stdout
			console.Log.Print(string(fileBuffer))
		}
	}
}

// Subcommand-Handler: Remove
func handleRemoveSubCmd(actions []string, context *ClientContext) {
	// Issue request
	_, err := context.pbClient.ModifyEntity(context.ctx, &pb.EntityMod{
		FilePath: *context.args.RemoveFile,
		Remove:   true,
	})

	// Check status
	if err != nil {
		utils.HandleErr(err, "failed to modify entity")
		os.Exit(1)
	} else {
		console.Log.Printf("Successfuly removed '%s'\n", *context.args.RemoveFile)
	}
}

// Subcommand-Handler: Backup -> Manager
func handleBackupManagerSubCmd(actions []string, context *ClientContext) {
	if *context.args.ToggleBackupManager {
		// Get current config
		resp, err := context.pbClient.GetBackupManagerConfig(context.ctx, &pb.EmptyMessage{})
		if err != nil {
			utils.HandleErr(err, "could not get current backup manager's from server")
			os.Exit(1)
		}

		if resp, err := context.pbClient.SetBackupManagerConfig(context.ctx, &pb.BackupManagerStatus{
			IsEnabled:       !resp.IsEnabled,
			RetentionPeriod: resp.RetentionPeriod,
			BackupFrequency: resp.BackupFrequency,
		}); err != nil {
			utils.HandleErr(err, "could not update backup manager's config")
			os.Exit(1)
		} else {
			console.Heading.Printf("Successfuly set Backup Manager to: %v\n", color.WhiteString(fmt.Sprintf("%v", resp.IsEnabled)))
		}
	} else if context.args.SetBackupRetention.Milliseconds() > 0 {
		// Get current config
		resp, err := context.pbClient.GetBackupManagerConfig(context.ctx, &pb.EmptyMessage{})
		if err != nil {
			utils.HandleErr(err, "could not get current backup manager's from server")
			os.Exit(1)
		}

		if _, err := context.pbClient.SetBackupManagerConfig(context.ctx, &pb.BackupManagerStatus{
			IsEnabled:       resp.IsEnabled,
			RetentionPeriod: uint64(context.args.SetBackupRetention.Milliseconds()),
			BackupFrequency: resp.BackupFrequency,
		}); err != nil {
			utils.HandleErr(err, "could not update backup manager's config")
			os.Exit(1)
		} else {
			console.Heading.Printf("Successfuly updated Backup Retention Period to: %v\n", color.WhiteString(context.args.SetBackupRetention.String()))
		}
	} else if context.args.SetBackupFrequency.Milliseconds() > 0 {
		// Get current config
		resp, err := context.pbClient.GetBackupManagerConfig(context.ctx, &pb.EmptyMessage{})
		if err != nil {
			utils.HandleErr(err, "could not get current backup manager's from server")
			os.Exit(1)
		}

		if _, err := context.pbClient.SetBackupManagerConfig(context.ctx, &pb.BackupManagerStatus{
			IsEnabled:       resp.IsEnabled,
			RetentionPeriod: resp.RetentionPeriod,
			BackupFrequency: uint64(context.args.SetBackupFrequency.Milliseconds()),
		}); err != nil {
			utils.HandleErr(err, "could not update backup manager's config")
			os.Exit(1)
		} else {
			console.Heading.Printf("Successfuly updated Backup Frequency to: %v\n", color.WhiteString(context.args.SetBackupFrequency.String()))
		}
	} else { // Default is --status
		if resp, err := context.pbClient.GetBackupManagerConfig(context.ctx, &pb.EmptyMessage{}); err != nil {
			utils.HandleErr(err, "get backup manager config error")
			os.Exit(1)
		} else {
			lastBackup := time.UnixMilli(int64(resp.LastBackupUnixTimestamp))
			backup_freq := time.UnixMilli(int64(resp.BackupFrequency))
			retention_period := time.UnixMilli(int64(resp.RetentionPeriod))

			console.Heading.Println("Backup Manager Configuration:")
			console.Log.Printf("- IsEnabled: %v\n", resp.IsEnabled)
			console.Log.Printf("- Total Backups: %d\n", resp.TotalBackups)

			if lastBackup.UnixMilli() == 0 {
				console.Log.Println("- Last Backup: NONE")
			} else {
				console.Log.Printf("- Last Backup: %s\n", lastBackup.Local().String())
			}

			console.Log.Printf("- Backup Frequency: %s\n", time.Duration(backup_freq.UnixNano()).String())
			console.Log.Printf("- Retention Period: %s\n", time.Duration(retention_period.UnixNano()).String())
		}
	}
}

// Subcommand-Handler: Backup
func handleBackupSubCmd(actions []string, context *ClientContext) {
	switch actions[0] {
	case "manager":
		handleBackupManagerSubCmd(actions[1:], context)
	case "list":
		// Issue request & handle response
		if resp, err := context.pbClient.ListInternalBackups(context.ctx, &pb.EmptyMessage{}); err != nil {
			utils.HandleErr(err, "list backups error")
			os.Exit(1)
		} else {
			if len(resp.Backups) == 0 {
				console.Warning.Println("No existng backups")
			} else {
				console.Heading.Printf("%d backup entries found:\n", len(resp.Backups))
				for idx, elt := range resp.Backups {
					// Construct Expiration Time
					created_at := time.UnixMilli(int64(elt.CreatedUnixTimestamp))
					expires_at := time.Now().Add(time.Millisecond * time.Duration(elt.ExpiresInUnixTimestamp))

					console.Heading.Printf("[%d]: %s\n", idx, elt.FileName)
					console.Log.Println("  - Created at: ", created_at.Local().String())
					console.Log.Println("  - Expires at: ", expires_at.Local().String())
				}
			}
		}
	case "invoke":
		// Issue backup invoke
		if resp, err := context.pbClient.InvokeNewStorageBackup(context.ctx, &pb.EmptyMessage{}); err != nil {
			utils.HandleErr(err, "invoke new backup error")
			os.Exit(1)
		} else {
			// Construct Expiration Time
			expires_at := time.Now().Add(time.Millisecond * time.Duration(resp.ExpiresInUnixTimestamp))

			console.Heading.Println("Successfuly backed up internal storage")
			console.Log.Println("  - Backup Filename: ", resp.FileName)
			console.Log.Println("  - Backup Expires at: ", expires_at.Local().String())
		}
	case "remove":
		if resp, err := context.pbClient.DeleteBackup(context.ctx, &pb.BackupEntryRequest{
			BackupFileName: *context.args.RemoveBackup,
		}); err != nil {
			utils.HandleErr(err, "failed to remove backup")
			os.Exit(1)
		} else {
			console.Heading.Printf("Successfully removed \"%s\"\n", color.WhiteString(resp.FileName))
		}
	case "export":
		// Request export
		if resp, err := context.pbClient.ExportBackup(context.ctx, &pb.BackupEntryRequest{
			BackupFileName: *context.args.ExportBackup,
		}); err != nil {
			console.Fatalln("Export Backup Error:", err)
		} else {
			// Write received file bytes to file
			if err := ioutil.WriteFile(*context.args.ExportFilePath, resp.FileData, 0664); err != nil {
				console.Fatalln("Error writing received backup to file:", err)
			} else {
				console.Heading.Printf("Successfuly export '%s' -> '%s'\n", color.WhiteString(resp.FileName), color.WhiteString(*context.args.ExportFilePath))
			}
		}
	case "import":
		// Read in file import
		fileBuffer, err := os.ReadFile(*context.args.ImportBackup)
		if err != nil {
			console.Fatalln("Error reading in file:", err)
		}

		// Issue import request
		if _, err := context.pbClient.ImportBackup(context.ctx, &pb.ImportBackupRequest{
			FileName: filepath.Base(*context.args.ImportBackup),
			FileData: fileBuffer,
		}); err != nil {
			console.Fatalln("Failed to import backup:", err)
		} else {
			console.Heading.Printf("Successfuly imported '%s'!\n", color.WhiteString(*context.args.ImportBackup))
		}
	case "restore":
		if resp, err := context.pbClient.RestoreFromBackup(context.ctx, &pb.RestoreFromBackupRequest{
			FileName: *context.args.RestoreFromBackup,
		}); err != nil {
			console.Fatalln("Failed to restore from backup:", err)
		} else {
			expires_at := time.Now().Add(time.Millisecond * time.Duration(resp.ExpiresInUnixTimestamp))

			console.Heading.Printf("Successfully restored from backup. Backup up previous storage'%s'\n", color.WhiteString(resp.FileName))
			console.Log.Println("  - Expires at: ", expires_at.Local().String())
		}
	}
}

// Subcommand-Handler: Version
func handleVersionSubCmnd(actions []string, context *ClientContext) {
	// Get current config
	resp, err := context.pbClient.GetServerVersion(context.ctx, &pb.ServerVersionRequest{})
	if err != nil {
		utils.HandleErr(err, "could not get server's version")
		os.Exit(1)
	}

	console.Heading.Printf("Server Version: %s\n", color.HiWhiteString(resp.Version))
	console.Heading.Printf("Client Version: %s\n", color.HiWhiteString(version))
}

func main() {
	// Init: Parse Arguments
	rawSubCmd, args := ParseArguments()
	subCmds := strings.Split(rawSubCmd, " ")
	directSubCmd, actions := subCmds[0], subCmds[1:]

	if *args.Verbose {
		configuration.EnableVerbose()
	}

	configuration.Init() // Client Config
	console.Init()

	grpcHost := configuration.LoadedConfig.GrpcHost
	grpcPort := configuration.LoadedConfig.GrpcPort
	address := fmt.Sprintf("%s:%d", grpcHost, grpcPort)
	var conn *grpc.ClientConn
	var conn_err error = nil

	// Init TLS Credentials
	insecure := configuration.LoadedConfig.Insecure
	tlsCertPath := configuration.LoadedConfig.TLSCertPath

	if !insecure {
		creds, err := credentials.NewClientTLSFromFile(tlsCertPath, "")
		if err != nil {
			console.Fatalln("tls could not be read from file:", err)
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

	// Construct Context
	context := ClientContext{
		pbClient: client,
		ctx:      ctx,
		args:     args,
	}

	// Client Reqeust
	switch directSubCmd {
	case "list":
		if actions[0] == "keys" {
			handleListKeysSubCmd(actions[1:], &context)
		} else if actions[0] == "storage" {
			handleListStorageSubCmd(actions[1:], &context)
		}
	case "keys":
		handleKeysSubCmd(actions, &context)
	case "encrypt":
		handleEncryptSubCmd(actions, &context)
	case "decrypt":
		handleDecryptSubCmd(actions, &context)
	case "remove":
		handleRemoveSubCmd(actions, &context)
	case "backup":
		handleBackupSubCmd(actions, &context)
	case "version":
		handleVersionSubCmnd(actions, &context)
	}
}
