package main

import (
	"context"
	"crypto/x509"
	"log"
	"openabyss/entity"
	pb "openabyss/proto/server"
)

// Obtains available stored Entity Keys
func (s openabyss_server) GetKeyNames(ctx context.Context, in *pb.EmptyRequest) (*pb.GetKeyNamesResponse, error) {
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
func (s openabyss_server) GetKeys(ctx context.Context, in *pb.EmptyRequest) (*pb.GetKeysResponse, error) {
	respObj := &pb.GetKeysResponse{
		entities: make([]pb.Entity, entity.Store.Length),
	}

	idx := 0
	for _, v := range entity.Store.Keys {
		respObj.entities[idx]
		respObj.Keys[idx].Name = v.Name
		respObj.Keys[idx].PublicKey = x509.MarshalPKCS1PublicKey(v.PublicKey)
		idx += 1
	}

	return respObj, nil
}

// Generate a keypair given a unique key name
func (s openabyss_server) GenerateKeyPair(ctx context.Context, in *pb.GenerateEntityRequest) (*pb.Entity, error) {
	e1, err := entity.GenerateKeys("keys", in.Name, 2048)
	if err == nil {
		log.Println("Generated Key:", e1.Name)
		entity.Store.Add(e1)
		entity.Store.Length += 1

		return &pb.Entity{
			Name:      e1.Name,
			PublicKey: x509.MarshalPKCS1PublicKey(e1.PublicKey),
		}, nil
	}
	return nil, err
}
