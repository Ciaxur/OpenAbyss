package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"openabyss/entity"
	pb "openabyss/proto/server"
	"openabyss/utils"
	"os"
	"path"
	"regexp"
)

// Encrypts requested file, saving the location to an internal structure
func (s openabyss_server) EncryptFile(ctx context.Context, in *pb.FilePacket) (*pb.EncryptResult, error) {
	// Adjust root path
	storagePath := regexp.MustCompile(`^(\.*)/`).ReplaceAllString(in.StoragePath, "")
	log.Printf("[EncryptFile]: storagePath extracted: '%s' -'%s'\n", in.StoragePath, storagePath)

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

		return &pb.EncryptResult{
			FileStoragePath: storedStoragePath,
			FileId:          fileId,
		}, err
	} else {
		err = errors.New("key id not found")
		return &pb.EncryptResult{
			FileStoragePath: "",
			FileId:          "",
		}, err
	}
}
