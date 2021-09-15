package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"openabyss/entity"
	pb "openabyss/proto/server"
	"openabyss/server/storage"
	"openabyss/utils"
	"os"
	"path"
	"regexp"
)

// Reused Structures
var (
	emptyEncryptResult = pb.EncryptResult{
		FileStoragePath: "",
		FileId:          "",
	}
)

// Encrypts requested file, saving the location to an internal structure
func (s openabyss_server) EncryptFile(ctx context.Context, in *pb.FilePacket) (*pb.EncryptResult, error) {
	// Adjust root path
	storagePath := regexp.MustCompile(`^(\.*)/`).ReplaceAllString(in.StoragePath, "")
	log.Printf("[EncryptFile]: storagePath extracted: '%s' -'%s'\n", in.StoragePath, storagePath)

	// Verify no duplicates
	if _, err := storage.Internal.GetFileByPath(storagePath); err == nil {
		log.Printf("[EncryptFile]: Duplicate internal FilePath found '%s'\n", storagePath)
		return &emptyEncryptResult, errors.New("duplicte internal file path'" + storagePath + "'")
	}

	// Handle Storage Directory
	storageDir := path.Join("./", ".storage")
	if !utils.DirExists(storageDir) {
		os.Mkdir(storageDir, 0755)
	}

	// Get requested key
	sk, ok := entity.Store.Keys[in.KeyName]
	var err error = nil
	if ok {
		// Stored in internal storage for lookup
		storedStoragePath := path.Join(storageDir, storagePath)

		// Generate fileId based on path
		fileIdBuffer := sha256.Sum256([]byte(
			path.Join(storedStoragePath, in.FileName),
		))
		fileId := hex.EncodeToString(fileIdBuffer[:])

		// Encrypt the data
		actualStoredPath := path.Join(storageDir, fileId)
		log.Printf("[EncryptFile]: storing '%s' -> '%s'\n", in.FileName, actualStoredPath)
		err = entity.Encrypt(in.FileBytes, actualStoredPath, sk.PrivateKey)

		// Store data in internal storage
		if err := storage.Internal.Store(fileId, storagePath, storage.Type_File); err != nil {
			log.Printf("[EncryptFile]: Failed to store encrypted file internally: %v\n", err)
			return &emptyEncryptResult, errors.New("could not store data internally")
		} else {
			storage.Internal.WriteToFile()
			log.Println("[EncryptFile]: Successfully stored encrypted data internally")
		}

		return &pb.EncryptResult{
			FileStoragePath: storagePath,
			FileId:          fileId,
		}, err
	} else {
		err = errors.New("key id not found")
		return &emptyEncryptResult, err
	}
}
