package main

import (
	"flag"
)

type Arguments struct {
	// KEYS/ENTRIES
	GetKeyNames     bool
	GetKeys         bool
	GenerateKeyPair string

	// ENCRYPT/DECRYPT
	EncryptFile string
	StoragePath string
	KeyId       string
}

func ParseArguments() Arguments {
	// KEYS/ENTRIES
	var flagGetKeyNames = flag.Bool("get-key-names", false, "Retrieves available key names")
	var flagGetKeys = flag.Bool("get-keys", false, "Retrieves available keys with their name and public key")
	var flagGenerateKeyPair = flag.String("generate-keypair", "", "Generate a keypair given the key's name")

	// ENCRYPT/DECRYPT
	var flagEncryptFile = flag.String("encrypt", "", "Encrypts given path, storing it in given storage path")
	var flagStoragePath = flag.String("storage-path", "", "Internal path of where to store data")
	var flagKeyId = flag.String("key-id", "", "Key's id/name used to encrypt")

	flag.Parse()
	return Arguments{
		GetKeyNames:     *flagGetKeyNames,
		GetKeys:         *flagGetKeys,
		GenerateKeyPair: *flagGenerateKeyPair,
		EncryptFile:     *flagEncryptFile,
		StoragePath:     *flagStoragePath,
		KeyId:           *flagKeyId,
	}
}
