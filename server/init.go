package main

import (
	"log"
	"openabyss/entity"
	"openabyss/server/configuration"
	"openabyss/server/storage"
	"openabyss/utils"
)

// Maps Encrypted AES Key to Entity
func mapAesKeyToEntity() {
	log.Println("[aes.mapping] Mapping Aes Ciphers to each Entity")
	for key, val := range storage.Internal.KeyMap {
		if ent, ok := entity.Store.Keys[key]; ok {
			ent.AesEncryptedKey = []byte(val.CipherEncKey)

			// Re-apply entity to Object
			entity.Store.Keys[key] = ent
		}
	}
}

func Init() {
	// Load Entity Store
	entity.Init()

	// Load Storage
	storage.Init()
	mapAesKeyToEntity()

	// Load Configuration
	configuration.Init()

	// Init Backup Manager
	go storage.Init_Backup_Manager()

	// Setup internal configuraiton
	port = configuration.LoadedConfig.GrpcPort
	host = configuration.LoadedConfig.GrpcHost
	tlsCert = configuration.LoadedConfig.TLSCertPath
	tlsKey = configuration.LoadedConfig.TLSKeyPath
	insecure = configuration.LoadedConfig.Insecure

	// Validate TLS Files exist, otherwise override to Insecure
	if !insecure && !(utils.FileExists(tlsCert) || utils.FileExists(tlsKey)) {
		log.Println("[server.init] Insecure override, no files found")
		insecure = true
	}
}
