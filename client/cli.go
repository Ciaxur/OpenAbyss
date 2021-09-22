package main

import (
	"flag"
)

type Arguments struct {
	// KEYS/ENTRIES
	GetKeyNames     bool
	GetKeys         bool
	GenerateKeyPair string
	FilePath        string

	// ENCRYPT/DECRYPT
	EncryptFile      bool
	DecryptFile      bool
	FilePacketOutput string
	StoragePath      string
	KeyId            string

	// PATH
	ListPath      bool
	RecursivePath bool

	// REMOVE/FORCE
	RemoveFile bool
	Force      bool // Used to force overwrite
}

func ParseArguments() Arguments {
	// KEYS/ENTRIES
	var flagGetKeyNames = flag.Bool("list-key-names", false, "Retrieves available key names")
	var flagGetKeys = flag.Bool("list-keys", false, "Retrieves available keys with their name and public key")
	var flagGenerateKeyPair = flag.String("generate-keypair", "", "Generate a keypair given the key's name")
	var flagFilePath = flag.String("filepath", "", "Path to file")
	flag.StringVar(flagFilePath, "f", "", "Path to file")

	// ENCRYPT/DECRYPT
	var flagEncryptFile = flag.Bool("encrypt", false, "Encrypts given path, storing it in given storage path")
	flag.BoolVar(flagEncryptFile, "e", false, "Encrypts given path, storing it in given storage path")

	var flagDecryptFile = flag.Bool("decrypt", false, "Decrypts file from given path, responding with file data")
	flag.BoolVar(flagDecryptFile, "d", false, "Decrypts file from given path, responding with file data")

	var flagFilePacketOutput = flag.String("file-packet-out", "", "Destination for incoming file packet data. Default: Outputs to stdout")
	flag.StringVar(flagFilePacketOutput, "o", "", "Destination for incoming file packet data. Default: Outputs to stdout")

	var flagStoragePath = flag.String("storage-path", "/", "Internal path of where to store data")
	flag.StringVar(flagStoragePath, "s", "/", "Internal path of where to store data")

	var flagKeyId = flag.String("key-id", "", "Key's id/name used to encrypt")
	flag.StringVar(flagKeyId, "k", "", "Key's id/name used to encrypt")

	// PATH
	var flagListPath = flag.Bool("list-path", false, "List an internal path given by the 'storage-path' argument | Root Storage by default")
	flag.BoolVar(flagListPath, "l", false, "List an internal path given by the 'storage-path' argument")
	var flagRecursive = flag.Bool("recursive", false, "Enabled recursive path listing")

	// REMOVE
	var flagRemoveFile = flag.Bool("remove", false, "Removes internal entry")
	flag.BoolVar(flagRemoveFile, "r", false, "Removes internal entry")

	// FORCE (overwrite)
	var flagForce = flag.Bool("force", false, "Forces supplied action")

	flag.Parse()
	return Arguments{
		GetKeyNames:      *flagGetKeyNames,
		GetKeys:          *flagGetKeys,
		GenerateKeyPair:  *flagGenerateKeyPair,
		FilePath:         *flagFilePath,
		EncryptFile:      *flagEncryptFile,
		DecryptFile:      *flagDecryptFile,
		FilePacketOutput: *flagFilePacketOutput,
		StoragePath:      *flagStoragePath,
		KeyId:            *flagKeyId,
		ListPath:         *flagListPath,
		RecursivePath:    *flagRecursive,
		RemoveFile:       *flagRemoveFile,
		Force:            *flagForce,
	}
}
