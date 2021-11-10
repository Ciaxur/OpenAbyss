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

func (s openabyss_server) ImportKey(ctx context.Context, in *pb.KeyImportRequest) (*pb.KeyImportResponse, error) {
	// TODO:
	return &pb.KeyImportResponse{}, nil
}
