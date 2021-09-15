package storage

import (
	"path"
	"strings"
	"time"
)

// Internal helper function for creating sub-storages
func (fsMap *FileStorageMap) create_sub_storage(sub_storage string) error {
	// New storage map if none exist
	if fsMap.StorageMap == nil {
		fsMap.StorageMap = map[string]FileStorageMap{}
	}

	// New sub storage IF none exist
	if _, ok := fsMap.StorageMap[sub_storage]; !ok {
		fsMap.StorageMap[sub_storage] = FileStorageMap{
			CreatedAt_UnixTimestamp:  uint64(time.Now().Unix()),
			ModifiedAt_UnixTimestamp: uint64(time.Now().Unix()),
			StorageMap:               make(map[string]FileStorageMap),
			Storage:                  make(map[string]FileStorage),
		}
	}

	return nil
}

// Returns the given sub-storage at the map's root level
func (fsMap *FileStorageMap) GetSubStorage(subStorageName string) *FileStorageMap {
	if subStorage, ok := fsMap.StorageMap[subStorageName]; ok {
		return &subStorage
	}
	return nil
}

// Returns the given storage name at the map's root level
func (fsMap *FileStorageMap) GetStorage(storageName string) *FileStorage {
	if storageFile, ok := fsMap.Storage[storageName]; ok {
		return &storageFile
	}
	return nil
}

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
		fsPtr.create_sub_storage(p)
		fsPtr = fsPtr.GetSubStorage(p)
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
