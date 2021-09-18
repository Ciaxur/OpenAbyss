package storage

import (
	"encoding/json"
	"errors"
	"io/ioutil"
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
func (fsMap *FileStorageMap) Store(fileId string, filePath string, fileByteSize uint64, fileType uint8) error {
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
		SizeInBytes:              fileByteSize,
		Type:                     fileType,
		CreatedAt_UnixTimestamp:  uint64(time.Now().Unix()),
		ModifiedAt_UnixTimestamp: uint64(time.Now().Unix()),
	}

	return nil
}

// Handles removing storage entry retuning the actual file path storage if successful
func (fsMap *FileStorageMap) RemoveStorage(ssPath string) (string, error) {
	// Obtain Sub Storage by path
	if fsMap, err := fsMap.GetSubStorageByPath(path.Dir(ssPath)); err != nil {
		return "", err
	} else {
		fsStorage := fsMap.GetStorage(path.Base(ssPath))
		if fsStorage == nil {
			return "", errors.New("storage entry not found")
		}

		// Store internal path & remove entry
		internalFilepath := path.Join(InternalStoragePath, fsStorage.Name)
		delete(fsMap.Storage, path.Base(ssPath))

		return internalFilepath, nil
	}
}

// Handles fetching given Sub-Storage from internal store
// Returns an error if not found
func (fsMap *FileStorageMap) GetSubStorageByPath(filePath string) (*FileStorageMap, error) {
	fsPtr := fsMap

	// NOTE: string.Split could create empty strings since the root
	//  split creates an empty string when split ie. '/file'
	subdirs := strings.Split(filePath, "/")

	// Traverse sub-storage
	for _, p := range subdirs {
		// Ignore empty paths
		if p == "" {
			continue
		}

		// Traverse Storage Map
		fsPtr = fsPtr.GetSubStorage(p)
		if fsPtr == nil {
			return nil, errors.New("sub-storage '" + p + "' not found")
		}
	}

	return fsPtr, nil
}

// Handles fetching given file from internal store
// Returns an error if not found
func (fsMap *FileStorageMap) GetFileByPath(filePath string) (*FileStorage, error) {
	// Get internal Sub-Storage
	fsMap, err := fsMap.GetSubStorageByPath(path.Dir(filePath))
	if err != nil {
		return nil, err
	}

	// Create storage entry
	if fileStorage := fsMap.GetStorage(path.Base(filePath)); fileStorage != nil {
		return fileStorage, nil
	} else {
		return nil, errors.New("file storage '" + path.Base(filePath) + "' not found")
	}
}

// Writes internal data to file
func (fsMap *FileStorageMap) WriteToFile() (int, error) {
	// Open & Save data
	data, _ := json.Marshal(Internal)
	if err := ioutil.WriteFile(InternalConfigPath, data, 0644); err != nil {
		return 0, err
	}
	return len(data), nil
}
