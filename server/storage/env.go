package storage

// Internal FileStorage
var (
	Internal         *FileStorageMap = &FileStorageMap{}
	InternalFilePath string
)

// File Type "Enum" Mapping
const (
	Type_File = 0
	Type_Dir  = 1
)

// Mapped FileStorage Object
type FileStorageMap struct {
	ModifiedAt_UnixTimestamp uint64                 `json:"modified_at_unix_timestamp"`
	CreatedAt_UnixTimestamp  uint64                 `json:"created_at_unix_timestamp"`
	Data                     map[string]FileStorage `json:"data"`
}

// FileStorage Structure for each Entry
type FileStorage struct {
	Path                     string `json:"path"`
	Name                     string `json:"name"`
	Type                     uint8  `json:"type"`
	CreatedAt_UnixTimestamp  uint64 `json:"created_at_unix_timestamp"`
	ModifiedAt_UnixTimestamp uint64 `json:"modified_at_unix_timestamp"`
}
