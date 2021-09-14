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
