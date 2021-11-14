package storage

import "time"

// Internal FileStorage
var (
	Internal FileStorageMap = FileStorageMap{
		CreatedAt_UnixTimestamp:  uint64(time.Now().Unix()),
		ModifiedAt_UnixTimestamp: uint64(time.Now().Unix()),
		StorageMap:               make(map[string]FileStorageMap),
		Storage:                  make(map[string]FileStorage),
		KeyMap:                   make(map[string]KeyStorage),
	}
	LastBackup          int64  = time.Now().UnixMilli()
	InternalStoragePath string = ".storage"
	InternalConfigPath  string
	BackupStoragePath   string = "backups" // InternalStoragePath/BackupStoragePath
)

// File Type "Enum" Mapping
const (
	Type_File = uint8(0)
	Type_Dir  = uint8(1)
)

// Mapped FileStorage Object
type FileStorageMap struct {
	ModifiedAt_UnixTimestamp uint64                    `json:"modified_at_unix_timestamp"`
	CreatedAt_UnixTimestamp  uint64                    `json:"created_at_unix_timestamp"`
	StorageMap               map[string]FileStorageMap `json:"sub_storage"`
	Storage                  map[string]FileStorage    `json:"storage"`
	KeyMap                   map[string]KeyStorage     `json:"keyStorage"`
}

// KeyStorage Structure for each Key
type KeyStorage struct {
	Name                     string `json:"name"`
	Description              string `json:"description"`
	Algorithm                string `json:"algorithm"`
	CipherEncKey             string `json:"cipherEncKey"`
	CipherAlgorithm          string `json:"cipherAlgorithm"`
	ExpiresAt_UnixTimestamp  uint64 `json:"expires_at_unix_timestamp"` // Expires the abilit to encrypt data, can still decrypt (becomes read-only)
	CreatedAt_UnixTimestamp  uint64 `json:"created_at_unix_timestamp"`
	ModifiedAt_UnixTimestamp uint64 `json:"modified_at_unix_timestamp"`
}

// FileStorage Structure for each Entry
type FileStorage struct {
	Path                     string `json:"path"`
	Name                     string `json:"name"`
	SizeInBytes              uint64 `json:"sizeInBytes"`
	Type                     uint8  `json:"type"`
	CreatedAt_UnixTimestamp  uint64 `json:"created_at_unix_timestamp"`
	ModifiedAt_UnixTimestamp uint64 `json:"modified_at_unix_timestamp"`
}
