package main

import (
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

type Arguments struct {
	// KEYS/ENTRIES
	GetKeyNames        *bool
	GetKeys            *bool
	GenerateKeyPair    *string
	KeyPairDescription *string
	KeyPairAlgo        *string
	FilePath           *string

	// ENCRYPT/DECRYPT
	EncryptFile      *string
	DecryptFile      *string
	FilePacketOutput *string
	StoragePath      *string
	KeyId            *string

	// PATH
	ListPath      *bool
	RecursivePath *bool

	// BACKUP
	ListBackups            *bool
	InvokeBackup           *bool
	BackupIndex            *int64
	GetBackupManagerStatus *bool
	ToggleBackupManager    *bool // Toggle On/Off Backup Manager
	SetBackupRetention     *time.Duration
	SetBackupFrequency     *time.Duration
	RemoveBackup           *string
	ExportBackup           *string
	ImportBackup           *string
	RestoreFromBackup      *string

	// REMOVE/FORCE
	RemoveFile *string
	Force      *bool // Used to force overwrite

	// MISC.
	Verbose *bool
}

func ParseArguments() (string, *Arguments) {
	args := Arguments{}

	// GENREAL
	args.FilePath = kingpin.Flag("filepath", "Path to file").Default("").String()

	// LIST
	listCmd := kingpin.Command("list", "List server internal data")
	args.GetKeyNames = listCmd.Flag("key-names", "Retrieves available key names").Bool()
	args.GetKeys = listCmd.Flag("keys", "Retrieves available keys with their name and public key").Bool()
	args.ListPath = listCmd.Flag("internal-storage", "List an internal path given by the 'storage-path' argument | Root Storage by default").Bool()

	// GENERATE
	generateCmd := kingpin.Command("generate", "Generate Keypair given key metadata")
	args.GenerateKeyPair = generateCmd.Flag("name", "Generated key's name").Required().String()
	args.KeyPairDescription = generateCmd.Flag("description", "Generated key's description").Default("").String()
	args.KeyPairAlgo = generateCmd.Flag("algorithm", "Generated key's algorithm").Default("rsa").String()

	// ENCRYPT/DECRYPT
	encryptCmd := kingpin.Command("encrypt", "Encrypts given path, storing it in given storage path")
	args.EncryptFile = encryptCmd.Flag("path", "Path to the file to encrypt").Required().String()

	decryptCmd := kingpin.Command("decrypt", "Decrypts file from given path, responding with file data")
	args.DecryptFile = decryptCmd.Flag("path", "Path to the file to decrypt on server").Required().String()

	args.FilePacketOutput = kingpin.Flag("file-packet-out", "Destination for incoming file packet data. Default: Outputs to stdout").Default("").String()

	args.StoragePath = kingpin.Flag("storage-path", "Internal path to data").Default("/").String()
	args.KeyId = kingpin.Flag("key-id", "Key's id/name used to encrypt").Default("").String()

	// PATH
	args.RecursivePath = kingpin.Flag("recursive", "Enabled recursive path listing").Bool()

	// BACKUP
	backupCmd := kingpin.Command("backup", "Backup Commands")
	args.ListBackups = backupCmd.Flag("list", "Lists backed up internal storage").Bool()
	args.BackupIndex = backupCmd.Flag("index", "Index of the stored backup").Default("-1").Int64()
	args.InvokeBackup = backupCmd.Flag("invoke-backup", "Creates a new backup of the internal storage").Bool()
	args.GetBackupManagerStatus = backupCmd.Flag("manager-status", "Returns the status of the Backup Manager on the server").Bool()
	args.ToggleBackupManager = backupCmd.Flag("toggle-backup-manager", "Toggles On/Off Backup Manager on the server").Bool()
	args.SetBackupRetention = backupCmd.Flag("set-backup-retention", "Sets the backup retention period of the Backup Manager").Default("0").Duration()
	args.SetBackupFrequency = backupCmd.Flag("set-backup-frequency", "Sets the backup frequency of the Backup Manager").Default("0").Duration()
	args.RemoveBackup = backupCmd.Flag("remove", "Removes stored backup from the server").Default("").String()
	args.ExportBackup = backupCmd.Flag("export", "Exports stored backup from the server to given filepath").Default("").String()
	args.ImportBackup = backupCmd.Flag("import", "Imported given backup path to the server").Default("").String()
	args.RestoreFromBackup = backupCmd.Flag("restore", "Restores server storage from given backup name").Default("").String()

	// REMOVE
	removeCmd := kingpin.Command("remove", "Remove an internal entry")
	args.RemoveFile = removeCmd.Flag("path", "Internal file to remove").Default("").String()

	// // FORCE (overwrite)
	args.Force = kingpin.Flag("force", "Forces supplied action").Bool()

	// OTHER
	args.Verbose = kingpin.Flag("verbose", "Enables verbose printing").Bool()

	subCmd := kingpin.Parse()
	return subCmd, &args
}
