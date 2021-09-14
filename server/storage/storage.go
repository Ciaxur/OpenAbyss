package storage

import (
	"path"
	"strings"
	"time"
)

// Handles splitting up file path & storing given fileId under split file paths
//  where filePath MUST include the filename
func (fsMap *FileStorageMap) Store(fileId string, filePath string, fileType uint8) error {
	fsPtr := fsMap

	// NOTE: string.Split could create empty strings since the root
	//  split creates an empty string when split ie. '/file'
	subdirs := strings.Split(path.Dir(filePath), "/")

	// Create sub-directories
	for _, p := range subdirs {
		// Ignore empty paths
		if p == "" {
			continue
		}

		// Create & traverse Storage Map
		if fsPtr.StorageMap == nil {
			fsPtr.StorageMap = make(map[string]FileStorageMap)
			fsPtr.StorageMap[p] = FileStorageMap{
				CreatedAt_UnixTimestamp:  uint64(time.Now().Unix()),
				ModifiedAt_UnixTimestamp: uint64(time.Now().Unix()),
				StorageMap:               make(map[string]FileStorageMap),
				Storage:                  make(map[string]FileStorage),
			}
		}

		// Traverse sub-storage
		fsPtr_inter := fsPtr.StorageMap[p]
		fsPtr = &fsPtr_inter
	}

	// Create storage entry
	if fsPtr.Storage == nil {
		fsPtr.Storage = make(map[string]FileStorage)
	}

	// Store file data
	fsPtr.Storage[path.Base(filePath)] = FileStorage{
		Path:                     filePath,
		Name:                     fileId,
		Type:                     fileType,
		CreatedAt_UnixTimestamp:  uint64(time.Now().Unix()),
		ModifiedAt_UnixTimestamp: uint64(time.Now().Unix()),
	}

	return nil
}
