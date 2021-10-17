package storage_test

import (
	"io/ioutil"
	"openabyss/server/configuration"
	"openabyss/server/storage"
	"openabyss/utils"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStorageBackup_SimpleBackup_File_Success(t *testing.T) {
	configuration.LoadedConfig.Backup.BackupFrequency = 0
	configuration.LoadedConfig.Backup.Enable = true
	go func() {
		time.Sleep(250 * time.Millisecond)
		configuration.LoadedConfig.Backup.Enable = false
	}()

	// Modify internal storage
	storage.InternalStoragePath = ".storage-test"
	storage.Init_Backup_Manager()

	// Check backup was successful
	assert.True(t, utils.DirExists(storage.InternalStoragePath), "did not create internal test storage path")

	files, err := ioutil.ReadDir(storage.InternalStoragePath)
	assert.Nil(t, err, "internal test storage path read failed")
	assert.NotEmpty(t, files, "no backup created")

	// Clean up
	wd, _ := os.Getwd()
	full_storage_test_path := path.Join(wd, storage.InternalStoragePath)
	os.RemoveAll(full_storage_test_path)
}
