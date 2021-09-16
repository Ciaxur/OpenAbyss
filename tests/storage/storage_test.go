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

	assert.Nil(t, err, "internal store failed")
	assert.Greater(t, len(storage.Internal.Storage), 0, "internal storage does not contain data")

	fStorage, ok := storage.Internal.Storage[path.Base(filePath)]
	assert.True(t, ok, "file not stored in root storage")

	// Check Storage Entries
	assert.Greater(t, fStorage.CreatedAt_UnixTimestamp, uint64(0), "filestorage did not create created timestamp")
	assert.Greater(t, fStorage.ModifiedAt_UnixTimestamp, uint64(0), "filestorage did not create modified timestamp")
	assert.Equal(t, storage.Type_File, fStorage.Type, "stored file is not a File Type")
	assert.Equal(t, hexFileId, fStorage.Name, "stored assigned filename does not match SHA256 Sum")
	assert.Equal(t, filePath, fStorage.Path, "stored file does not contain the client's full path")
}

func TestFileStorage_Store_RecursiveSubDir_Success(t *testing.T) {
	filePath := "/path/to/file"
	fileId := sha256.Sum256([]byte(filePath))
	hexFileId := hex.EncodeToString(fileId[:])
	storage.Internal = storage.FileStorageMap{}

	err := storage.Internal.Store(hexFileId, filePath, storage.Type_File)

	assert.Nil(t, err, "internal store failed")
	assert.Greater(t, len(storage.Internal.StorageMap), 0, "internal storage map does not contain data")

	// Root Level should contain "path" sub-storage
	rootSubStorage := storage.Internal.GetSubStorage("path")
	assert.NotNil(t, rootSubStorage, "path is not the root sub-storage")
	assert.NotNil(t, rootSubStorage.Storage, "root storage failed to create")

	// Check metadata
	assert.Greater(t, rootSubStorage.CreatedAt_UnixTimestamp, uint64(0), "root filestorage did not create created timestamp")
	assert.Greater(t, rootSubStorage.ModifiedAt_UnixTimestamp, uint64(0), "root filestorage did not create modified timestamp")

	// 2nd-Level Storage
	secLvlStorage := rootSubStorage.GetSubStorage("to")
	assert.NotNil(t, secLvlStorage, "2nd level should be the 'to' sub-storage")
	assert.NotNil(t, secLvlStorage.Storage, "2nd level storage failed to create")

	// Check metadata
	assert.Greater(t, secLvlStorage.CreatedAt_UnixTimestamp, uint64(0), "2nd level filestorage did not create created timestamp")
	assert.Greater(t, secLvlStorage.ModifiedAt_UnixTimestamp, uint64(0), "2nd level filestorage did not create modified timestamp")

	// 3rd-Level Storage
	thirdLvlStorage := secLvlStorage.GetSubStorage("file")
	assert.Nil(t, thirdLvlStorage, "3rd level contains sub-storage")

	// Get storage file from 2nd level
	assert.Greater(t, len(secLvlStorage.Storage), 0, "2nd level does not contain storage")
	storageFile := secLvlStorage.GetStorage("file")
	assert.NotNil(t, storageFile, "2nd level did not return storage file")

	// Check metadata
	assert.Greater(t, storageFile.CreatedAt_UnixTimestamp, uint64(0), "basename filestorage did not create created timestamp")
	assert.Greater(t, storageFile.ModifiedAt_UnixTimestamp, uint64(0), "basename filestorage did not create modified timestamp")

	// Check file-specific metadata
	assert.Equal(t, storage.Type_File, storageFile.Type, "stored file is not a File Type")
	assert.Equal(t, hexFileId, storageFile.Name, "stored assigned filename does not match SHA256 Sum")
	assert.Equal(t, filePath, storageFile.Path, "stored file does not contain the client's full path")
}

func TestFileStorage_Store_GetRootSubStorageByPath_Success(t *testing.T) {
	filePath := "/file"
	fileId := sha256.Sum256([]byte(filePath))
	hexFileId := hex.EncodeToString(fileId[:])
	storage.Internal = storage.FileStorageMap{}

	// Store file internally
	err := storage.Internal.Store(hexFileId, filePath, storage.Type_File)

	assert.Nil(t, err, "internal store failed")
	assert.Greater(t, len(storage.Internal.Storage), 0, "internal storage does not contain data")

	// Fetch Sub-Storage by path
	fsSubStorage, err := storage.Internal.GetSubStorageByPath(path.Dir(filePath))
	assert.Nil(t, err, "internal storage failed to get sub-storage")
	assert.NotNil(t, fsSubStorage, "internal storage did not return sub-storage")

	// Verify Sub-Storage is correct
	// The containing Sub-Storage should have 'file' stored
	fsFile := fsSubStorage.GetStorage(path.Base(filePath))
	assert.NotNil(t, fsFile, "sub-storage does not contain file as storage")
	assert.Equal(t, hexFileId, fsFile.Name, "sub-storage file name mismatch")
	assert.Equal(t, storage.Type_File, fsFile.Type, "sub-storage file type is not a file-type")
}

func TestFileStorage_Store_GetRecursiveSubStorageByPath_Success(t *testing.T) {
	filePath := "/path/to/file"
	fileId := sha256.Sum256([]byte(filePath))
	hexFileId := hex.EncodeToString(fileId[:])
	storage.Internal = storage.FileStorageMap{}

	// Store file internally
	err := storage.Internal.Store(hexFileId, filePath, storage.Type_File)

	assert.Nil(t, err, "internal store failed")
	assert.Greater(t, len(storage.Internal.StorageMap), 0, "internal storage map does not contain data")

	// Fetch Sub-Storage by path
	fsSubStorage, err := storage.Internal.GetSubStorageByPath(path.Dir(filePath))
	assert.Nil(t, err, "internal storage failed to get sub-storage")
	assert.NotNil(t, fsSubStorage, "internal storage did not return sub-storage")

	// Verify Sub-Storage is correct
	// The containing Sub-Storage should have 'file' stored
	fsFile := fsSubStorage.GetStorage(path.Base(filePath))
	assert.NotNil(t, fsFile, "sub-storage does not contain file as storage")
	assert.Equal(t, hexFileId, fsFile.Name, "sub-storage file name mismatch")
	assert.Equal(t, storage.Type_File, fsFile.Type, "sub-storage file type is not a file-type")
}

func TestFileStorage_Store_GetRootFileByPath_Success(t *testing.T) {
	filePath := "/file"
	fileId := sha256.Sum256([]byte(filePath))
	hexFileId := hex.EncodeToString(fileId[:])
	storage.Internal = storage.FileStorageMap{}

	// Store file internally
	err := storage.Internal.Store(hexFileId, filePath, storage.Type_File)

	assert.Nil(t, err, "internal store failed")
	assert.Greater(t, len(storage.Internal.Storage), 0, "internal storage does not contain data")

	// Fetch file by path
	fileStorage, err := storage.Internal.GetFileByPath(filePath)

	assert.Nil(t, err, "fetching file by path failed")
	assert.NotNil(t, fileStorage, "fetching filestorage failed")

	// Validate FileStorage correctness
	assert.Equal(t, filePath, fileStorage.Path, "file storage meta: path does not match")
	assert.Equal(t, hexFileId, fileStorage.Name, "file storage meta: fileId does not match")
	assert.Equal(t, storage.Type_File, fileStorage.Type, "file storage meta: file type does not match")
}

func TestFileStorage_Store_GetRecursiveStorageByPath_Success(t *testing.T) {
	filePath := "/path/to/file"
	fileId := sha256.Sum256([]byte(filePath))
	hexFileId := hex.EncodeToString(fileId[:])
	storage.Internal = storage.FileStorageMap{}

	// Store file internally
	err := storage.Internal.Store(hexFileId, filePath, storage.Type_File)

	assert.Nil(t, err, "internal store failed")
	assert.Greater(t, len(storage.Internal.StorageMap), 0, "internal storage map does not contain data")

	// Fetch file by path
	fileStorage, err := storage.Internal.GetFileByPath(filePath)

	assert.Nil(t, err, "fetching file by path failed")
	assert.NotNil(t, fileStorage, "fetching filestorage failed")

	// Validate FileStorage correctness
	assert.Equal(t, filePath, fileStorage.Path, "file storage meta: path does not match")
	assert.Equal(t, hexFileId, fileStorage.Name, "file storage meta: fileId does not match")
	assert.Equal(t, storage.Type_File, fileStorage.Type, "file storage meta: file type does not match")
}

func TestFileStorage_Remove_RootFile_Success(t *testing.T) {
	filePath := "/file"
	fileId := sha256.Sum256([]byte(filePath))
	hexFileId := hex.EncodeToString(fileId[:])
	storage.Internal = storage.FileStorageMap{}

	// Store file internally
	err := storage.Internal.Store(hexFileId, filePath, storage.Type_File)

	assert.Nil(t, err, "internal store failed")
	assert.Greater(t, len(storage.Internal.Storage), 0, "internal storage does not contain data")

	_, ok := storage.Internal.Storage[path.Base(filePath)]
	assert.True(t, ok, "file not stored in root storage")

	// Remove internal file
	internalFilepath, err := storage.Internal.RemoveStorage(filePath)
	assert.Nil(t, err, "internal storage failed to remove storage")
	assert.NotEmpty(t, internalFilepath, "internal storage failed to return internal file path after removal")
	assert.Equal(t, path.Join(storage.InternalStoragePath, hexFileId), internalFilepath, "incorrect internal storage path returned")

	// Verify removal
	fsFile, err := storage.Internal.GetFileByPath(filePath)
	assert.NotNil(t, err, "internal storage did not fail to find internal file")
	assert.Nil(t, fsFile, "internal storage found removed internal file")
}

func TestFileStorage_Remove_RecursiveSubStorage_Success(t *testing.T) {
	filePath := "/path/to/file"
	fileId := sha256.Sum256([]byte(filePath))
	hexFileId := hex.EncodeToString(fileId[:])
	storage.Internal = storage.FileStorageMap{}

	// Store file internally
	err := storage.Internal.Store(hexFileId, filePath, storage.Type_File)

	assert.Nil(t, err, "internal store failed")
	assert.Greater(t, len(storage.Internal.StorageMap), 0, "internal storage map does not contain data")

	// Remove internal file
	internalFilepath, err := storage.Internal.RemoveStorage(filePath)
	assert.Nil(t, err, "internal storage failed to remove storage")
	assert.NotEmpty(t, internalFilepath, "internal storage failed to return internal file path after removal")
	assert.Equal(t, path.Join(storage.InternalStoragePath, hexFileId), internalFilepath, "incorrect internal storage path returned")

	// Verify removal
	fsFile, err := storage.Internal.GetFileByPath(filePath)
	assert.NotNil(t, err, "internal storage did not fail to find internal file")
	assert.Nil(t, fsFile, "internal storage found removed internal file")
}
