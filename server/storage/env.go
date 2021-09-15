package storage

// Internal FileStorage
var (
	Internal FileStorageMap = FileStorageMap{
		StorageMap: make(map[string]FileStorageMap),
		Storage:    make(map[string]FileStorage),
	}
	InternalStoragePath string = ".storage"
	InternalConfigPath  string
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
}

// FileStorage Structure for each Entry
type FileStorage struct {
	Path                     string `json:"path"`
	Name                     string `json:"name"`
	Type                     uint8  `json:"type"`
	CreatedAt_UnixTimestamp  uint64 `json:"created_at_unix_timestamp"`
	ModifiedAt_UnixTimestamp uint64 `json:"modified_at_unix_timestamp"`
}
