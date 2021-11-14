package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"openabyss/entity"
	pb "openabyss/proto/server"
	"openabyss/server/storage"
	"openabyss/utils"
	"path"
)

// KeyTarPackage encapsulates required Key information for importing
//  and exporting keys
type KeyTarPackage struct {
	KeyStoreEntry storage.KeyStorage `json:"keyStoreEntry"`
	KeyEntity     entity.Entity      `json:"keyEntity"`
	RawPrivateKey []byte             `json:"privateKey"`
	RawPublicKey  []byte             `json:"publicKey"`
}

// Export existing keypair
func (s openabyss_server) ExportKey(ctx context.Context, in *pb.KeyExportRequest) (*pb.KeyExportResponse, error) {
	log.Printf("[ExportKey]: Export key '%s' requested\n", in.KeyId)

	// Try and find if the key is available
	if entry, ok := storage.Internal.KeyMap[in.KeyId]; !ok {
		log.Printf("[ExportKey]: Export key '%s' not found\n", in.KeyId)
		return nil, errors.New("requested key not found")
	} else {
		// Get Key Store's Data
		v := entity.Store.Keys[in.KeyId]

		// Load Key files as bytes
		skBuffer, errSk := ioutil.ReadFile(path.Join(entity.KeyStorePath, v.Name))
		pkBuffer, errPk := ioutil.ReadFile(path.Join(entity.KeyStorePath, v.Name) + ".pub")

		// Handle file read errors
		if errSk != nil {
			utils.HandleErr(errSk, "[ExportKey]: failed to read private key file")
			return nil, errors.New("internal error")
		}
		if errPk != nil {
			utils.HandleErr(errPk, "[ExportKey]: failed to read public key file")
			return nil, errors.New("internal error")
		}

		// Serialize package
		packageBuffer, err := json.Marshal(&KeyTarPackage{
			KeyStoreEntry: entry,
			KeyEntity:     v,
			RawPrivateKey: skBuffer,
			RawPublicKey:  pkBuffer,
		})
		if err != nil {
			utils.HandleErr(err, "[ExportKey]: failed to marshal Key Tar Package")
			return nil, errors.New("internal error")
		}

		// Gzip those bytes!
		compBuffer := bytes.NewBuffer(nil)
		writer := gzip.NewWriter(compBuffer)
		writer.Write(packageBuffer)
		writer.Close()

		// Respond with gziped data
		log.Printf("[ExportKey]: Exported key '%s'\n", in.KeyId)
		return &pb.KeyExportResponse{
			KeyGzip: compBuffer.Bytes(),
			KeyId:   entry.Name,
		}, nil
	}
}

// Import key to server
func (s openabyss_server) ImportKey(ctx context.Context, in *pb.KeyImportRequest) (*pb.KeyImportResponse, error) {
	log.Printf("[ImportKey]: Import key '%s' requested\n", in.KeyId)

	// Check if key exists
	if _, ok := storage.Internal.KeyMap[in.KeyId]; ok && !in.Force {
		log.Printf("[ImportKey]: Import key '%s' duplicate found\n", in.KeyId)
		return nil, errors.New("duplicate key found, issue force=true to overwrite duplicate")
	} else {
		// Unpack & unmarshal blob
		if reader, err := gzip.NewReader(bytes.NewBuffer(in.KeyGzip)); err != nil {
			log.Printf("[ImportKey]: Failed to unpack key '%s'\n", in.KeyId)
			return nil, errors.New("failed unpack archive")
		} else {
			var pkg KeyTarPackage
			buffer, _ := ioutil.ReadAll(reader)
			json.Unmarshal(buffer, &pkg)

			// Make sure package entries' key id matches what's intended
			pkg.KeyEntity.Name = in.KeyId
			pkg.KeyStoreEntry.Name = in.KeyId

			// Add Key to Key Store
			entity.Store.Keys[in.KeyId] = pkg.KeyEntity

			// Add Key to internal storage
			storage.Internal.KeyMap[in.KeyId] = pkg.KeyStoreEntry

			// Overwrite key if force requested
			if ok {
				log.Printf("[ImportKey]: Overwriting keys for '%s'\n", pkg.KeyEntity.Name)
			}

			// Save Private & Public key files
			skPath := path.Join(entity.KeyStorePath, pkg.KeyEntity.Name)
			ioutil.WriteFile(skPath, pkg.RawPrivateKey, 0644)
			log.Printf("[ImportKey]: Private key saved to '%s'\n", skPath)

			pkPath := path.Join(entity.KeyStorePath, pkg.KeyEntity.Name) + ".pub"
			ioutil.WriteFile(pkPath, pkg.RawPublicKey, 0644)
			log.Printf("[ImportKey]: Public key saved to '%s'\n", pkPath)

			return &pb.KeyImportResponse{}, nil
		}
	}
}
