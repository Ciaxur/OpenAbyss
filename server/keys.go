package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
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
	log.Printf("[GetKeyNames]: Total Entities in Store: %d\n", len(storage.Internal.KeyMap))

	keyResp := &pb.GetKeyNamesResponse{
		Keys: make([]string, len(storage.Internal.KeyMap)),
	}

	idx := 0
	for _, v := range storage.Internal.KeyMap {
		keyResp.Keys[idx] = v.Name
		idx += 1
	}

	return keyResp, nil
}

// Obtains available stored Entities without the Private Keys
func (s openabyss_server) GetKeys(ctx context.Context, in *pb.EmptyMessage) (*pb.GetKeysResponse, error) {
	log.Printf("[GetKeys]: Total Entities in Store: %d\n", len(storage.Internal.KeyMap))

	respObj := &pb.GetKeysResponse{
		Entities: make([]*pb.Entity, len(storage.Internal.KeyMap)),
	}

	idx := 0
	for key, value := range storage.Internal.KeyMap {
		// Encode Public Key
		publicKeyBuffer := bytes.NewBuffer(nil)
		if entity.Store.Has(key) {
			v := entity.Store.Get(key)
			pem.Encode(publicKeyBuffer, &pem.Block{
				Type:  "RSA PUBLIC KEY",
				Bytes: x509.MarshalPKCS1PublicKey(v.PublicKey),
			})
		}

		// Construct response for the entry
		respObj.Entities[idx] = &pb.Entity{
			Name:                   value.Name,
			PublicKeyName:          publicKeyBuffer.Bytes(),
			Description:            value.Description,
			Algorithm:              value.Algorithm,
			CreatedUnixTimestamp:   value.CreatedAt_UnixTimestamp,
			ModifiedUnixTimestamp:  value.ModifiedAt_UnixTimestamp,
			ExpiresAtUnixTimestamp: value.ExpiresAt_UnixTimestamp,
			SigningPublicKeyPem:    value.SigningPublicKey_pem,
		}
		idx += 1
	}

	return respObj, nil
}

// Generates AEK Key used for Cipher block
func GenerateAESKey() []byte {
	// Generate a random 32-bit AES Key to use for Encrypting & Decrypting Data
	aesKey := make([]byte, 32)
	rand.Reader.Read(aesKey)
	return aesKey
}

// Generate a keypair given a unique key name
func (s openabyss_server) GenerateKeyPair(ctx context.Context, in *pb.GenerateEntityRequest) (*pb.Entity, error) {
	// Early return: Keypair name already exists
	if _, ok := storage.Internal.KeyMap[in.Name]; ok {
		log.Printf("[GenerateKeyPair]: Could not generate. KeyPair '%s' already exists\n", in.Name)
		return nil, errors.New("keypair name already exists")
	}

	// Generate requested key by algorithm
	log.Printf("[GenerateKeyPair]: Generating KeyPair[%s] for '%s' key\n", in.Algorithm, in.Name)

	// Shared data between algorithms
	aesKey := GenerateAESKey()
	keyExpiresAt := uint64(time.Now().UnixMilli()) + in.ExpiresInUnixTimestamp
	if in.ExpiresInUnixTimestamp == 0 {
		keyExpiresAt = 0
	}

	// Construct Key Entries with pre-set shared values
	keyStorage := storage.KeyStorage{
		Name:                     in.Name,
		Description:              in.Description,
		Algorithm:                in.Algorithm,
		CipherAlgorithm:          "aes", // NOTE: Move to specific entry unless shared
		CreatedAt_UnixTimestamp:  uint64(time.Now().UnixMilli()),
		ModifiedAt_UnixTimestamp: uint64(time.Now().UnixMilli()),
		ExpiresAt_UnixTimestamp:  uint64(keyExpiresAt),
	}
	response := &pb.Entity{
		Name:                   in.Name,
		Description:            in.Description,
		Algorithm:              in.Algorithm,
		CreatedUnixTimestamp:   uint64(time.Now().UnixMilli()),
		ModifiedUnixTimestamp:  uint64(time.Now().UnixMilli()),
		ExpiresAtUnixTimestamp: uint64(time.Now().UnixMilli()) + in.ExpiresInUnixTimestamp,
	}

	// Generate key based on given Algorithm
	switch in.Algorithm {
	case "rsa":
		e1, err := entity.GenerateKeys(entity.KeyStorePath, in.Name, 2048, aesKey)
		if err == nil {
			log.Println("Generated Key:", e1.Name)

			// Modify entries to represent RSA algorithm option
			keyStorage.Name = e1.Name
			keyStorage.CipherEncKey = string(e1.AesEncryptedKey)

			entity.Store.Add(e1)
			storage.Internal.KeyMap[e1.Name] = keyStorage

			response.Name = e1.Name
			response.PublicKeyName = x509.MarshalPKCS1PublicKey(e1.PublicKey)
		} else {
			log.Printf("[GenerateKeyPair]: Could not generate KeyPair[%s] for '%s' key\n", in.Algorithm, in.Name)
			return nil, err
		}
	case "ed25519": // Signature
		// Generate a random 32-bit AES Key to use for Encrypting & Decrypting Data
		aesKey := make([]byte, 32)
		rand.Reader.Read(aesKey)

		// Generate Public/Private Sig keys | Convert to base64 and store them
		//  respectively. Private key goes to user, public key goes to both
		if pk, sk, err := ed25519.GenerateKey(rand.Reader); err != nil {
			log.Println("[GenerateKeyPair]: Failed to generate signing algorithm key")
			return nil, errors.New("internal error: failed to genreate signing keys")
		} else {
			b, _ := x509.MarshalPKIXPublicKey(pk)
			block := &pem.Block{
				Type:  "PUBLIC KEY",
				Bytes: b,
			}
			pk_pem := pem.EncodeToMemory(block)

			keyStorage.SigningPublicKey_pem = base64.StdEncoding.EncodeToString(pk_pem)
			response.SigningPrivateKeySeed = base64.StdEncoding.EncodeToString(sk.Seed())
			response.SigningPublicKeyPem = keyStorage.SigningPublicKey_pem
		}

		// Modify & Add Key to store
		keyStorage.CipherEncKey = base64.StdEncoding.EncodeToString(aesKey)
		storage.Internal.KeyMap[in.Name] = keyStorage
	case "none": // No Encryption | AES-Only
		// Generate a random 32-bit AES Key to use for Encrypting & Decrypting Data
		aesKey := make([]byte, 32)
		rand.Reader.Read(aesKey)

		return nil, errors.New("wip; not implemented yet")
	default:
		log.Printf("[GenerateKeyPair]: Algorithm '%s' not supported\n", in.Algorithm)
		return nil, errors.New("algorithm not supported")
	}

	return response, nil
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

		// Generate Public Key Buffer (RSA) and then remove rsa key entry data
		publicKeyBuffer := bytes.NewBufferString("")
		if entry.Algorithm == "rsa" {
			v := entity.Store.Keys[in.KeyId]
			publicKeyBuffer = bytes.NewBuffer(nil)
			pem.Encode(publicKeyBuffer, &pem.Block{
				Type:  "RSA PUBLIC KEY",
				Bytes: x509.MarshalPKCS1PublicKey(v.PublicKey),
			})

			// Clean up RSA keys
			delete(entity.Store.Keys, in.KeyId)
			entity.Store.Length -= 1
			os.Remove(path.Join(entity.KeyStorePath, in.KeyId+".pub"))
			os.Remove(path.Join(entity.KeyStorePath, in.KeyId))
		}

		// Remove Key from Key Store and Internal Storage
		delete(storage.Internal.KeyMap, in.KeyId)

		return &pb.Entity{
			Name:                   entry.Name,
			Description:            entry.Description,
			PublicKeyName:          publicKeyBuffer.Bytes(),
			Algorithm:              entry.Algorithm,
			CreatedUnixTimestamp:   entry.CreatedAt_UnixTimestamp,
			ModifiedUnixTimestamp:  entry.ModifiedAt_UnixTimestamp, // Last modified
			ExpiresAtUnixTimestamp: entry.ExpiresAt_UnixTimestamp,
			SigningPublicKeyPem:    entry.SigningPublicKey_pem,
		}, nil
	}
}
