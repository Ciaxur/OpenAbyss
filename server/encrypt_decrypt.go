package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"log"
	"openabyss/entity"
	pb "openabyss/proto/server"
	"openabyss/server/storage"
	"openabyss/utils"
	"os"
	"path"
	"regexp"
)

// Encrypts requested file, saving the location to an internal structure
func (s openabyss_server) EncryptFile(ctx context.Context, in *pb.FilePacket) (*pb.EncryptResult, error) {
	// Verify Key provided
	if len(in.Options.KeyName) == 0 {
		return nil, errors.New("no key name provided")
	}

	// Adjust root path
	storagePath := regexp.MustCompile(`^(\.*)/`).ReplaceAllString(in.Options.StoragePath, "")
	log.Printf("[EncryptFile]: storagePath extracted: '%s' - '%s'\n", in.Options.StoragePath, storagePath)

	// Adjust for internal root path
	if storagePath == "" {
		storagePath = "/"
	}

	// Verify no duplicates
	if !in.Options.Overwrite {
		if _, err := storage.Internal.GetFileByPath(path.Join(storagePath, in.FileName)); err == nil {
			log.Printf("[EncryptFile]: Duplicate internal FilePath found '%s'\n", path.Join(storagePath, in.FileName))
			return &pb.EncryptResult{}, errors.New("duplicte internal file path'" + path.Join(storagePath, in.FileName) + "'")
		}
	}

	// Handle Storage Directory
	storageDir := path.Join("./", storage.InternalStoragePath)
	if !utils.DirExists(storageDir) {
		os.Mkdir(storageDir, 0755)
	}

	// Get requested key
	sk, ok := entity.Store.Keys[in.Options.KeyName]
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
		log.Printf("[EncryptFile]: storing '%s' -> '%s'\n", path.Join(storagePath, in.FileName), actualStoredPath)

		if destWriter, err := os.Create(actualStoredPath); err != nil {
			utils.HandleErr(err, "[EncryptFile]: failed to create file path")
		} else {
			if err := entity.CipherEncrypt(in.FileBytes, destWriter, &sk); err != nil {
				utils.HandleErr(err, "[EncryptFile]: failed to encrypt")
				destWriter.Close()
			}
		}

		// Store data in internal storage
		if _, err := storage.Internal.Store(fileId, path.Join(storagePath, in.FileName), uint64(in.SizeInBytes), storage.Type_File, in.Options.Overwrite); err != nil {
			log.Printf("[EncryptFile]: Failed to store encrypted file internally: %v\n", err)
			return &pb.EncryptResult{}, errors.New("could not store data internally")
		} else {
			storage.Internal.WriteToFile()
			log.Printf("[EncryptFile]: Successfully stored encrypted data, %d bytes, internally\n", in.SizeInBytes)
		}

		return &pb.EncryptResult{
			FileStoragePath: storagePath,
			FileId:          fileId,
		}, err
	} else {
		err = errors.New("key id not found")
		return &pb.EncryptResult{}, err
	}
}

// Encrypts requested file, saving the location to an internal structure
func (s openabyss_server) DecryptFile(ctx context.Context, in *pb.DecryptRequest) (*pb.FilePacket, error) {
	// Verify Key provided
	if len(in.PrivateKeyName) == 0 {
		return nil, errors.New("no key name provided")
	}

	// Get suplied entity based on name
	sk := entity.Store.Get(string(in.PrivateKeyName))
	if sk == nil {
		log.Printf("[DecryptFile]: private key '%s' not found\n", in.PrivateKeyName)
		return &pb.FilePacket{}, errors.New("supplied key name not found")
	}

	// Adjust root path
	storagePath := regexp.MustCompile(`^(\.*)`).ReplaceAllString(in.FilePath, "")
	log.Printf("[DecryptFile]: storagePath extracted: '%s' -'%s'\n", in.FilePath, storagePath)

	// Obtain from internal storage
	fsFile, err := storage.Internal.GetFileByPath(storagePath)
	if err != nil {
		log.Printf("[DecryptFile]: File '%s' not found: %v\n", in.FilePath, err)
		return &pb.FilePacket{}, errors.New("file '" + storagePath + "' not found")
	}

	// Decrypt the data
	encFilePath := path.Join(storage.InternalStoragePath, fsFile.Name)

	if fsBytes, err := ioutil.ReadFile(encFilePath); err != nil {
		log.Printf("[DecryptFile]: Failed to read '%s'\n", encFilePath)
		return &pb.FilePacket{}, err
	} else {
		destWriter := bytes.NewBuffer(nil)

		// Attempt to Decrypt data
		if err := entity.CipherDecrypt(fsBytes, destWriter, sk); err != nil {
			log.Printf("[DecryptFile]: Failed to decrypt file '%s'\n", encFilePath)
			return &pb.FilePacket{}, err
		}
		log.Printf("[DecryptFile]: Successfuly decrypted, %d bytes, file '%s'\n", fsFile.SizeInBytes, encFilePath)

		// Successful Response
		return &pb.FilePacket{
			FileBytes:   destWriter.Bytes(),
			SizeInBytes: int64(fsFile.SizeInBytes),
			FileName:    path.Base(fsFile.Path),
			Options: &pb.FileOptions{
				StoragePath: fsFile.Path,
				KeyName:     string(in.PrivateKeyName),
			},
		}, nil
	}
}
