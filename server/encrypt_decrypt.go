package main

import (
	"context"
	"errors"
	"openabyss/entity"
	pb "openabyss/proto/server"
	"openabyss/utils"
	"os"
	"path"
	"regexp"
)

// Encrypts requested file, saving the location to an internal structure
func (s openabyss_server) EncryptFile(ctx context.Context, in *pb.FilePacket) (*pb.EmptyMessage, error) {
	// Adjust root path
	storagePath := regexp.MustCompile(`^(\.*)/`).ReplaceAllString(in.StoragePath, "")

	// Handle Storage Directory
	storageDir := path.Join("./", "storage")
	if !utils.DirExists(storageDir) {
		os.Mkdir(storageDir, 0755)
	}

	// Get requested key
	sk, ok := entity.Store.Keys[in.KeyName]
	var err error = nil
	if ok {
		finalStoragePath := path.Join(storageDir, storagePath)

		// Handle Path and destination creation
		// Create directory path if not available
		//  - Case1: Create directory path for dest path being a directory
		//  - Case2: Create directory path for dest path's parent
		if isDirPath, _ := regexp.MatchString("/$", finalStoragePath); isDirPath {
			// Create path if doesn't exist
			if !utils.DirExists(finalStoragePath) {
				os.MkdirAll(finalStoragePath, 0755)
			}

			// Adjust Destination path with file end
			finalStoragePath = path.Join(finalStoragePath, in.FileName)
		} else if !utils.DirExists(path.Dir(finalStoragePath)) {
			os.MkdirAll(path.Dir(finalStoragePath), 0755)
		}

		// Encrypt the data
		err = entity.Encrypt(in.FileBytes, finalStoragePath, sk.PrivateKey)
	} else {
		err = errors.New("key id not found")
	}

	return &pb.EmptyMessage{}, err
}
