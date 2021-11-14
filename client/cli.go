package main

import (
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

type Arguments struct {
	// KEYS/ENTRIES
	GetKeyNames        *bool
	GenerateKeys       *bool
	KeyPairName        *string
	KeyPairDescription *string
	KeyPairAlgo        *string
	KeyExpiration      *time.Duration

	// KEY MOD
	KeyIdMod                *string
	KeyPairNameMod          *string
	KeyPairDescriptionMod   *string
	KeyExpirationMod        *time.Duration
	KeyExpirationDisableMod *bool

	// KEY REMOVE
	KeyIdRem *string

	// KEY EXPORT/IMPORT
	KeyExportFilePath *string
	KeyExportKeyId    *string
	KeyImportFilePath *string
	KeyImportKeyId    *string

	// LIST
	ListStoragePath *string

	// ENCRYPT/DECRYPT
	EncryptFile      *string
	DecryptFile      *string
	FilePacketOutput *string
	StoragePath      *string
	EncryptKeyId     *string
	DecryptKeyId     *string

	// PATH
	ListPath      *bool
	RecursivePath *bool

	// BACKUP
	GetBackupManagerStatus *bool
	ToggleBackupManager    *bool // Toggle On/Off Backup Manager
	SetBackupRetention     *time.Duration
	SetBackupFrequency     *time.Duration
	RemoveBackup           *string
	ExportBackup           *string
	ExportFilePath         *string
	ImportBackup           *string
	RestoreFromBackup      *string

	// REMOVE/FORCE
	RemoveFile *string
	Force      *bool // Used to force overwrite

	// MISC.
	Verbose *bool
	Version *bool
}

func ParseArguments() (string, *Arguments) {
	args := Arguments{}

	// LIST
	listCmd := kingpin.Command("list", "List server internal data")

	// LIST: Keys
	listKeysCmd := listCmd.Command("keys", "Retrieves available keys with their name and public key")
	args.GetKeyNames = listKeysCmd.Flag("names", "Retrieves available key names only").Bool()

	// LIST: Internal Storage
	listStorageCmd := listCmd.Command("storage", "List an internal path")
	args.ListStoragePath = listStorageCmd.Flag("path", "Internal path to data").Default("/").String()
	args.RecursivePath = listStorageCmd.Flag("recursive", "Enabled recursive path listing").Bool()

	// KEY
	keyCmd := kingpin.Command("keys", "Key interaction sub-menu")

	// KEY: Modify
	keyModCmd := keyCmd.Command("modify", "Key modification sub-menu")
	args.KeyIdMod = keyModCmd.Flag("key-id", "Key name to modify").Required().String()
	args.KeyPairNameMod = keyModCmd.Flag("name", "Modify key name").Default("").String()
	args.KeyPairDescriptionMod = keyModCmd.Flag("description", "Modify key description").Default("").String()
	args.KeyExpirationMod = keyModCmd.Flag("expire", "Set expiration duration for given key, making the key read-only").Default("0s").Duration()
	args.KeyExpirationDisableMod = keyModCmd.Flag("no-expire", "Disable key expiration for given key").Default("false").Bool()

	// KEY: Remove
	keyRemCmd := keyCmd.Command("remove", "Key removal sub-menu")
	args.KeyIdRem = keyRemCmd.Flag("key-id", "Key name to remove").Required().String()

	// KEY: Generation
	keyGenerateCmd := keyCmd.Command("generate", "Generate Keypair given key metadata")
	args.KeyPairName = keyGenerateCmd.Flag("name", "Generated key's name").Required().String()
	args.KeyPairDescription = keyGenerateCmd.Flag("description", "Generated key's description").Default("").String()
	args.KeyPairAlgo = keyGenerateCmd.Flag("algorithm", "Generated key's algorithm").Default("rsa").String()
	args.KeyExpiration = keyGenerateCmd.Flag("expire", "Set expiration duration for generated key").Default("0").Duration()

	// KEY: Export
	keyExportCmd := keyCmd.Command("export", "Export key sub-menu")
	args.KeyExportFilePath = keyExportCmd.Flag("dest", "Destination of exported keys").Required().String()
	args.KeyExportKeyId = keyExportCmd.Flag("key-id", "Key's id/name to export").Required().String()

	// KEY: Import
	keyImportCmd := keyCmd.Command("import", "Import key sub-menu")
	args.KeyImportFilePath = keyImportCmd.Flag("path", "Path to key that will be imported").Required().String()
	args.KeyImportKeyId = keyImportCmd.Flag("key-id", "Key's id/name to be imported to").Required().String()

	// ENCRYPT
	encryptCmd := kingpin.Command("encrypt", "Encrypts given path, storing it in given storage path")
	args.EncryptFile = encryptCmd.Flag("path", "Path to the file to encrypt").Required().String()
	args.EncryptKeyId = encryptCmd.Flag("key-id", "Key's id/name used to encrypt").Required().String()
	args.StoragePath = encryptCmd.Flag("storage-path", "Internal path to store encrpyted data").Default("/").String()

	// DECRYPT
	decryptCmd := kingpin.Command("decrypt", "Decrypts file from given path, responding with file data")
	args.DecryptFile = decryptCmd.Flag("path", "Path to the file to decrypt on server").Required().String()
	args.DecryptKeyId = decryptCmd.Flag("key-id", "Key's id/name used to encrypt").Required().String()
	args.FilePacketOutput = decryptCmd.Flag("dest", "Destination for incoming file packet data. Default: Outputs to stdout").Default("").String()

	// BACKUP
	backupCmd := kingpin.Command("backup", "Backup Commands")
	backupCmd.Command("list", "Lists backed up internal storage")
	backupCmd.Command("invoke", "Creates a new backup of the internal storage")

	// BACKUP: Manager
	backupManagerCmd := backupCmd.Command("manager", "Backup Manager Sub-Commands")
	args.GetBackupManagerStatus = backupManagerCmd.Flag("status", "Returns the status of the Backup Manager on the server").Bool()
	args.ToggleBackupManager = backupManagerCmd.Flag("toggle", "Toggles On/Off Backup Manager on the server").Bool()
	args.SetBackupRetention = backupManagerCmd.Flag("set-retention", "Sets the backup retention period of the Backup Manager").Default("0").Duration()
	args.SetBackupFrequency = backupManagerCmd.Flag("set-frequency", "Sets the backup frequency of the Backup Manager").Default("0").Duration()

	// BACKUP: Remove
	backupRemoveCmd := backupCmd.Command("remove", "Removes stored backups from the server")
	args.RemoveBackup = backupRemoveCmd.Flag("name", "Backup name to remove").Required().String()

	// BACKUP: Export
	backupExportCmd := backupCmd.Command("export", "Exports stored backup from the server")
	args.ExportBackup = backupExportCmd.Flag("name", "Backup name to export").Required().String()
	args.ExportFilePath = backupExportCmd.Flag("path", "Path to export file to").Default("").String()

	// BACKUP: Import
	backupImportCmd := backupCmd.Command("import", "Import backups into server")
	args.ImportBackup = backupImportCmd.Flag("path", "Path to backup being imported into the server").Default("").String()

	// BACKUP: Restore
	backupRestoreCmd := backupImportCmd.Command("restore", "Restore backups from server storage")
	args.RestoreFromBackup = backupRestoreCmd.Flag("name", "Backup name used to restore").Default("").String()

	// REMOVE
	removeCmd := kingpin.Command("remove", "Remove an internal entry")
	args.RemoveFile = removeCmd.Flag("path", "Internal file to remove").Required().String()

	// FORCE (overwrite)
	args.Force = kingpin.Flag("force", "Forces supplied action").Bool()

	kingpin.Command("version", "Prints Client/Servier Versions")

	// OTHER
	args.Verbose = kingpin.Flag("verbose", "Enables verbose printing").Bool()

	subCmd := kingpin.Parse()

	return subCmd, &args
}
