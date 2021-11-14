package main

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log"
	"openabyss/entity"
	pb "openabyss/proto/server"
	"openabyss/server/storage"
	"os"
	"path"
	"strings"
	"time"
)

// Obtains available stored Entity Keys
func (s openabyss_server) GetKeyNames(ctx context.Context, in *pb.EmptyMessage) (*pb.GetKeyNamesResponse, error) {
	log.Printf("[GetKeyNames]: Total Entities in Store: %d\n", entity.Store.Length)

	keyResp := &pb.GetKeyNamesResponse{
		Keys: make([]string, entity.Store.Length),
	}

	idx := 0
	for _, v := range entity.Store.Keys {
		keyResp.Keys[idx] = v.Name
		idx += 1
	}

	return keyResp, nil
}

// Obtains available stored Entities without the Private Keys
func (s openabyss_server) GetKeys(ctx context.Context, in *pb.EmptyMessage) (*pb.GetKeysResponse, error) {
	log.Printf("[GetKeys]: Total Entities in Store: %d\n", entity.Store.Length)

	respObj := &pb.GetKeysResponse{
		Entities: make([]*pb.Entity, entity.Store.Length),
	}

	idx := 0
	for k, v := range entity.Store.Keys {
		// Encode Public Key
		publicKeyBuffer := bytes.NewBuffer(nil)
		pem.Encode(publicKeyBuffer, &pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(v.PublicKey),
		})

		// Construct response for the entry
		_key := storage.Internal.KeyMap[k]
		respObj.Entities[idx] = &pb.Entity{
			Name:                   _key.Name,
			PublicKeyName:          publicKeyBuffer.Bytes(),
			Description:            _key.Description,
			Algorithm:              _key.Algorithm,
			CreatedUnixTimestamp:   _key.CreatedAt_UnixTimestamp,
			ModifiedUnixTimestamp:  _key.ModifiedAt_UnixTimestamp,
			ExpiresAtUnixTimestamp: _key.ExpiresAt_UnixTimestamp,
		}
		idx += 1
	}

	return respObj, nil
}

// Generate a keypair given a unique key name
func (s openabyss_server) GenerateKeyPair(ctx context.Context, in *pb.GenerateEntityRequest) (*pb.Entity, error) {
	// Early return: Keypair name already exists
	if entity.Store.Has(in.Name) {
		log.Printf("[GenerateKeyPair]: Could not generate. KeyPair '%s' already exists\n", in.Name)
		return nil, errors.New("keypair name already exists")
	}

	log.Printf("[GenerateKeyPair]: Generating KeyPair for '%s' key\n", in.Name)
	e1, err := entity.GenerateKeys(entity.KeyStorePath, in.Name, 2048)
	if err == nil {
		// Construct Key Expiration
		keyExpiresAt := uint64(time.Now().UnixMilli()) + in.ExpiresInUnixTimestamp
		if in.ExpiresInUnixTimestamp == 0 {
			keyExpiresAt = 0
		}

		log.Println("Generated Key:", e1.Name)
		entity.Store.Add(e1)
		storage.Internal.KeyMap[e1.Name] = storage.KeyStorage{
			Name:                     e1.Name,
			Description:              in.Description,
			Algorithm:                "rsa", // TODO: Change me when other algos are supported
			CipherEncKey:             string(e1.AesEncryptedKey),
			CipherAlgorithm:          "aes", // TODO: Change me when other algos are supported
			CreatedAt_UnixTimestamp:  uint64(time.Now().UnixMilli()),
			ModifiedAt_UnixTimestamp: uint64(time.Now().UnixMilli()),
			ExpiresAt_UnixTimestamp:  uint64(keyExpiresAt),
		}

		return &pb.Entity{
			Name:                   e1.Name,
			Description:            in.Description,
			Algorithm:              "rsa'", // TODO: Change me when other algos are supported
			CreatedUnixTimestamp:   uint64(time.Now().UnixMilli()),
			ModifiedUnixTimestamp:  uint64(time.Now().UnixMilli()),
			PublicKeyName:          x509.MarshalPKCS1PublicKey(e1.PublicKey),
			ExpiresAtUnixTimestamp: uint64(time.Now().UnixMilli()) + in.ExpiresInUnixTimestamp,
		}, nil
	} else {
		log.Printf("[GenerateKeyPair]: Could not generate KeyPair for '%s' key\n", in.Name)
	}
	return nil, err
}

// Modify existing keypair
func (s openabyss_server) ModifyKeyPair(ctx context.Context, in *pb.EntityModifyRequest) (*pb.Entity, error) {
	// Trim spaces
	newName := strings.Trim(in.Name, " ")
	newDesc := strings.Trim(in.Description, " ")

	// Get entry to be modified
	if entry, ok := storage.Internal.KeyMap[in.KeyId]; !ok {
		log.Printf("[ModifyKeyPair]: '%s' key not found\n", in.KeyId)
		return nil, errors.New("entity key-id not found")
	} else {
		log.Printf("[ModifyKeyPair]: Modifying '%s' key\n", in.KeyId)

		// Verify no Duplicates
		if _, ok := storage.Internal.KeyMap[newName]; len(newName) > 0 && ok {
			return nil, errors.New("new name for key already exists")
		}

		// Start modifying
		if len(newName) > 0 {
			log.Printf("[ModifyKeyPair]: Modifying name '%s' -> '%s'\n", entry.Name, newName)
			entry.Name = newName
		}
		if len(newDesc) > 0 {
			log.Printf("[ModifyKeyPair]: Modifying description '%s' -> '%s'\n", entry.Description, newDesc)
			entry.Description = newDesc
		}

		// Modify new Data
		entry.ModifiedAt_UnixTimestamp = uint64(time.Now().UnixMilli())

		// Modify entity expiration
		if in.ModifyKeyExpiration {
			log.Printf("[ModifyKeyPair]: Modifying expiration from '%d' -> '%d' for key '%s'\n", in.ExpiresInUnixTimestamp, entry.ExpiresAt_UnixTimestamp, in.KeyId)

			// Construct Key Expiration
			keyExpiresAt := uint64(time.Now().UnixMilli()) + in.ExpiresInUnixTimestamp
			if in.ExpiresInUnixTimestamp == 0 {
				keyExpiresAt = 0
			}
			entry.ExpiresAt_UnixTimestamp = keyExpiresAt
		}

		// Modify Key name & old map entries
		if len(newName) > 0 && in.KeyId != newName {
			// Store internally with new key
			storage.Internal.KeyMap[newName] = entry

			// Modify Entity Key Store
			newStoreKey := entity.Store.Keys[in.KeyId]
			newStoreKey.Name = newName
			entity.Store.Keys[newName] = newStoreKey
			os.Rename(path.Join(entity.KeyStorePath, in.KeyId), path.Join(entity.KeyStorePath, newName))
			os.Rename(path.Join(entity.KeyStorePath, in.KeyId+".pub"), path.Join(entity.KeyStorePath, newName+".pub"))

			// Remove old Keys
			delete(entity.Store.Keys, in.KeyId)
			delete(storage.Internal.KeyMap, in.KeyId)
		} else { // Store new metadata
			storage.Internal.KeyMap[in.KeyId] = entry
			newName = in.KeyId
		}
	}

	// No new name change,
	if len(newName) == 0 {
		newName = in.KeyId
	}

	// Generate Public Key Buffer
	v := entity.Store.Keys[newName]
	publicKeyBuffer := bytes.NewBuffer(nil)
	pem.Encode(publicKeyBuffer, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(v.PublicKey),
	})

	entity := storage.Internal.KeyMap[newName]
	return &pb.Entity{
		Name:                   entity.Name,
		Description:            entity.Description,
		PublicKeyName:          publicKeyBuffer.Bytes(),
		Algorithm:              entity.Algorithm,
		CreatedUnixTimestamp:   entity.CreatedAt_UnixTimestamp,
		ModifiedUnixTimestamp:  entity.ModifiedAt_UnixTimestamp,
		ExpiresAtUnixTimestamp: entity.ExpiresAt_UnixTimestamp,
	}, nil
}

// Remove existing keypair
func (s openabyss_server) RemoveKeyPair(ctx context.Context, in *pb.EntityRemoveRequest) (*pb.Entity, error) {

	// Get entry to be removed
	if entry, ok := storage.Internal.KeyMap[in.KeyId]; !ok {
		log.Printf("[RemoveKeyPair]: Key '%s' not found\n", in.KeyId)
		return nil, errors.New("key-id not found")
	} else {
		log.Printf("[RemoveKeyPair]: Removing '%s' key\n", in.KeyId)

		// Generate Public Key Buffer
		v := entity.Store.Keys[in.KeyId]
		publicKeyBuffer := bytes.NewBuffer(nil)
		pem.Encode(publicKeyBuffer, &pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(v.PublicKey),
		})

		// Remove Key from Key Store and Internal Storage
		delete(entity.Store.Keys, in.KeyId)
		delete(storage.Internal.KeyMap, in.KeyId)
		os.Remove(path.Join(entity.KeyStorePath, in.KeyId+".pub"))
		os.Remove(path.Join(entity.KeyStorePath, in.KeyId))
		entity.Store.Length -= 1

		return &pb.Entity{
			Name:                   entry.Name,
			Description:            entry.Description,
			PublicKeyName:          publicKeyBuffer.Bytes(),
			Algorithm:              entry.Algorithm,
			CreatedUnixTimestamp:   entry.CreatedAt_UnixTimestamp,
			ModifiedUnixTimestamp:  entry.ModifiedAt_UnixTimestamp, // Last modified
			ExpiresAtUnixTimestamp: entry.ExpiresAt_UnixTimestamp,
		}, nil
	}
}
