package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
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
	"time"
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

	// Check internal key found and validate signature
	internalKey, ok := storage.Internal.KeyMap[in.Options.KeyName]
	if !ok {
		log.Printf("[EncryptFile]: Key '%s' not found\n", in.FileName)
		return nil, errors.New("key id not found")
	}
	if internalKey.Algorithm == "ed25519" {
		pk_pem, _ := base64.StdEncoding.DecodeString(internalKey.SigningPublicKey_pem)
		pk := utils.PEM_to_ed25519(pk_pem)
		if !ed25519.Verify(pk, in.FileBytes, in.FileSignature) {
			log.Println("[EncryptFile]: File signature invalid")
			return nil, errors.New("invalid signature")
		}
		log.Println("[EncryptFile]: File signature validated")
	}

	// Get requested encryption key
	sk, key_store_found := entity.Store.Keys[in.Options.KeyName]

	// Verify key has not expired (if expires | none zero)
	expires_in := time.Now().UnixMilli() - int64(internalKey.ExpiresAt_UnixTimestamp)
	if internalKey.ExpiresAt_UnixTimestamp != 0 && expires_in > 0 {
		log.Printf("[EncryptFile]: Key '%s' expired '%d'ms ago\n", in.Options.KeyName, expires_in)
		return nil, errors.New("failed to encrypt, key expired")
	}

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
		return nil, errors.New("internal storage failure")
	} else if key_store_found {
		// Encrypt the file by decrypting the cipher using rsa and
		//  then encrypting the file
		if err := entity.CipherEncrypt(in.FileBytes, destWriter, &sk, internalKey.CipherEncKey); err != nil {
			utils.HandleErr(err, "[EncryptFile]: failed to encrypt")
			destWriter.Close()
		}
	} else {
		// Do not enrypt the file internally, but signature is requried
		//  which is checked above.
		// TODO: In the future, add AES encryption to be used if signature
		/// is verified. Then add another condition here to do so.
		destWriter.Write(in.FileBytes)
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
	}, nil

}

// Encrypts requested file, saving the location to an internal structure
func (s openabyss_server) DecryptFile(ctx context.Context, in *pb.DecryptRequest) (*pb.FilePacket, error) {
	// Verify Key provided
	if len(in.KeyName) == 0 {
		return nil, errors.New("no key name provided")
	}

	// Check file signature prior to request completion
	internalKey, ok := storage.Internal.KeyMap[string(in.KeyName)]
	if !ok {
		log.Printf("[DecryptFile]: key '%s' not found\n", in.KeyName)
		return nil, errors.New("supplied key name not found")
	}
	if internalKey.Algorithm == "ed25519" {
		pk_pem, _ := base64.StdEncoding.DecodeString(internalKey.SigningPublicKey_pem)
		pk := utils.PEM_to_ed25519(pk_pem)
		if !ed25519.Verify(pk, []byte(in.FilePath), in.FilePathSignature) {
			log.Println("[EncryptFile]: File signature invalid")
			return nil, errors.New("invalid signature")
		}
		log.Println("[EncryptFile]: File signature validated")
	}

	// Get suplied entity based on name
	sk, key_store_found := entity.Store.Keys[string(in.KeyName)]

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
		return nil, err
	} else {
		destWriter := bytes.NewBuffer(nil)

		// Attempt to Decrypt data
		if key_store_found {
			if err := entity.CipherDecrypt(fsBytes, destWriter, &sk, internalKey.CipherEncKey); err != nil {
				log.Printf("[DecryptFile]: Failed to decrypt file '%s'\n", encFilePath)
				return nil, err
			}
			log.Printf("[DecryptFile]: Successfuly decrypted, %d bytes, file '%s'\n", fsFile.SizeInBytes, encFilePath)
		} else {
			// Data was not encrypted but signature was verified, return the
			//  data from storage.
			// TODO: When sig + aes is added, make sure to decrypt using cipher here.
			destWriter.Write(fsBytes)
		}

		// Successful Response
		return &pb.FilePacket{
			FileBytes:   destWriter.Bytes(),
			SizeInBytes: int64(fsFile.SizeInBytes),
			FileName:    path.Base(fsFile.Path),
			Options: &pb.FileOptions{
				StoragePath: fsFile.Path,
				KeyName:     string(in.KeyName),
			},
		}, nil
	}
}
