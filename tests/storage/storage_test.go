package storage_test

import (
	"crypto/sha256"
	"encoding/hex"
	"openabyss/server/storage"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileStorage_Store_RootFile_Success(t *testing.T) {
	filePath := "/filename"
	fileId := sha256.Sum256([]byte(filePath))
	hexFileId := hex.EncodeToString(fileId[:])
	storage.Internal = storage.FileStorageMap{}

	err := storage.Internal.Store(hexFileId, filePath, storage.Type_File)

	assert.Nil(t, err, "internal store should not have failed")
	assert.Greater(t, len(storage.Internal.Storage), 0, "internal storage should contain data")

	fStorage, ok := storage.Internal.Storage[path.Base(filePath)]
	assert.True(t, ok, "file not stored in root storage")

	// Check Storage Entries
	assert.Greater(t, fStorage.CreatedAt_UnixTimestamp, uint64(0), "filestorage should have a created timestamp")
	assert.Greater(t, fStorage.ModifiedAt_UnixTimestamp, uint64(0), "filestorage should have a modified timestamp")
	assert.Equal(t, storage.Type_File, fStorage.Type, "stored file should be a File Type")
	assert.Equal(t, hexFileId, fStorage.Name, "stored assigned filename should match SHA256 Sum")
	assert.Equal(t, filePath, fStorage.Path, "stored file should contain the client's full path")
}

func TestFileStorage_Store_RecursiveSubDir_Success(t *testing.T) {
	filePath := "/path/to/file"
	fileId := sha256.Sum256([]byte(filePath))
	hexFileId := hex.EncodeToString(fileId[:])
	storage.Internal = storage.FileStorageMap{}

	err := storage.Internal.Store(hexFileId, filePath, storage.Type_File)

	assert.Nil(t, err, "internal store should not have failed")
	assert.Greater(t, len(storage.Internal.StorageMap), 0, "internal storage map should contain data")

	// Root Level should contain "path" sub-storage
	rootSubStorage := storage.Internal.GetSubStorage("path")
	assert.NotNil(t, rootSubStorage, "path should be the root sub-storage")
	assert.NotNil(t, rootSubStorage.Storage, "root storage should be not be nil")

	// Check metadata
	assert.Greater(t, rootSubStorage.CreatedAt_UnixTimestamp, uint64(0), "root filestorage should have a created timestamp")
	assert.Greater(t, rootSubStorage.ModifiedAt_UnixTimestamp, uint64(0), "root filestorage should have a modified timestamp")

	// 2nd-Level Storage
	secLvlStorage := rootSubStorage.GetSubStorage("to")
	assert.NotNil(t, secLvlStorage, "2nd level should be the 'to' sub-storage")
	assert.NotNil(t, secLvlStorage.Storage, "2nd level storage should not be nil")

	// Check metadata
	assert.Greater(t, secLvlStorage.CreatedAt_UnixTimestamp, uint64(0), "2nd level filestorage should have a created timestamp")
	assert.Greater(t, secLvlStorage.ModifiedAt_UnixTimestamp, uint64(0), "2nd level filestorage should have a modified timestamp")

	// 3rd-Level Storage
	thirdLvlStorage := secLvlStorage.GetSubStorage("file")
	assert.Nil(t, thirdLvlStorage, "3rd level should be nil, since there is not sub-storage there")

	// Get storage file from 2nd level
	assert.Greater(t, len(secLvlStorage.Storage), 0, "2nd level should contain storage")
	storageFile := secLvlStorage.GetStorage("file")
	assert.NotNil(t, storageFile, "2nd level should return storage file")

	// Check metadata
	assert.Greater(t, storageFile.CreatedAt_UnixTimestamp, uint64(0), "basename filestorage should have a created timestamp")
	assert.Greater(t, storageFile.ModifiedAt_UnixTimestamp, uint64(0), "basename filestorage should have a modified timestamp")

	// Check file-specific metadata
	assert.Equal(t, storage.Type_File, storageFile.Type, "stored file should be a File Type")
	assert.Equal(t, hexFileId, storageFile.Name, "stored assigned filename should match SHA256 Sum")
	assert.Equal(t, filePath, storageFile.Path, "stored file should contain the client's full path")
}

func TestFileStorage_Store_GetRootFile_Success(t *testing.T) {
	filePath := "/path/to/file"
	fileId := sha256.Sum256([]byte(filePath))
	hexFileId := hex.EncodeToString(fileId[:])
	storage.Internal = storage.FileStorageMap{}

	// Store file internally
	err := storage.Internal.Store(hexFileId, filePath, storage.Type_File)

	assert.Nil(t, err, "internal store should not have failed")
	assert.Greater(t, len(storage.Internal.StorageMap), 0, "internal storage map should contain data")

	// Fetch file by path
	fileStorage, err := storage.Internal.GetFileByPath(filePath)

	assert.Nil(t, err, "internal storage GetFileByPath should not fail")
	assert.NotNil(t, fileStorage, "filestorage should contain file storage")

	// Validate FileStorage correctness
	assert.Equal(t, filePath, fileStorage.Path, "file storage meta: path does not match")
	assert.Equal(t, hexFileId, fileStorage.Name, "file storage meta: fileId does not match")
	assert.Equal(t, storage.Type_File, fileStorage.Type, "file storage meta: file type does not match")
}
