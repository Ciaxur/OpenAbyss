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
	for _, v := range entity.Store.Keys {
		// Encode Public Key
		publicKeyBuffer := bytes.NewBuffer(nil)
		pem.Encode(publicKeyBuffer, &pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(v.PublicKey),
		})

		// Construct restponse for the entry
		respObj.Entities[idx] = &pb.Entity{
			Name:          v.Name,
			PublicKeyName: publicKeyBuffer.Bytes(),
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
		log.Println("Generated Key:", e1.Name)
		entity.Store.Add(e1)

		return &pb.Entity{
			Name:          e1.Name,
			PublicKeyName: x509.MarshalPKCS1PublicKey(e1.PublicKey),
		}, nil
	} else {
		log.Printf("[GenerateKeyPair]: Could not generate KeyPair for '%s' key\n", in.Name)
	}
	return nil, err
}
