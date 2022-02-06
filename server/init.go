package main

import (
	"log"
	"openabyss/entity"
	"openabyss/server/configuration"
	"openabyss/server/storage"
	"openabyss/utils"
)

func Init() {
	// Load Entity Store
	entity.Init()

	// Load Storage
	storage.Init()

	// Load Configuration
	configuration.Init()

	// Init Backup Manager
	go storage.Init_Backup_Manager()

	// Setup internal configuraiton
	port = configuration.LoadedConfig.GrpcPort
	host = configuration.LoadedConfig.GrpcHost
	tlsPoolPath = configuration.LoadedConfig.TLSCertPoolPath
	tlsCert = configuration.LoadedConfig.TLSCertPath
	tlsKey = configuration.LoadedConfig.TLSKeyPath
	insecure = configuration.LoadedConfig.Insecure

	// Validate TLS Files exist, otherwise override to Insecure
	if !insecure && tlsPoolPath == "" && !(utils.FileExists(tlsCert) || utils.FileExists(tlsKey)) {
		log.Println("[server.init] Insecure override, no files found")
		insecure = true
	}
}
