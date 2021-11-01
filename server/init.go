package main

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"log"
	"openabyss/entity"
	"openabyss/server/configuration"
	"openabyss/server/storage"
	"openabyss/utils"
)

func mapAesCipherToEntity() {
	log.Println("[aes.mapping] Mapping Aes Ciphers to each Entity")
	for key, val := range storage.Internal.KeyMap {
		if ent, ok := entity.Store.Keys[key]; ok {
			ent.AesEncryptedKey = []byte(val.CipherEncKey)
			if cipherEncKey, err := base64.StdEncoding.DecodeString(val.CipherEncKey); err != nil {
				utils.HandleErr(err, "could not decode base64 encrypted cipher key")
			} else {
				// Decrypt the Cipher
				cipherKey := bytes.NewBufferString("")
				if err := entity.Decrypt(cipherEncKey, cipherKey, ent.PrivateKey); err != nil {
					utils.HandleErr(err, "failed to decrypt cipher key")
				}

				if cipherAes, err := aes.NewCipher(cipherKey.Bytes()); err != nil {
					utils.HandleErr(err, "failed to create new cipher")
				} else {
					ent.Cipher = &cipherAes
					log.Printf("[aes.mapping] AES Cipher Created for key[%s]: %d Block Size\n", key, cipherAes.BlockSize())
				}
			}

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
	mapAesCipherToEntity()

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
		insecure = true
	}
}
