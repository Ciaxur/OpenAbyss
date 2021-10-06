package main

import (
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

	// Setup internal configuraiton
	port = configuration.LoadedConfig.GrpcPort
	tlsCert = configuration.LoadedConfig.TLSCertPath
	tlsKey = configuration.LoadedConfig.TLSKeyPath
	insecure = configuration.LoadedConfig.Insecure

	// Validate TLS Files exist, otherwise override to Insecure
	if !insecure && !(utils.FileExists(tlsCert) || utils.FileExists(tlsKey)) {
		insecure = true
	}
}
