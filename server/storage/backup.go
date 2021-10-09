package storage

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"openabyss/server/configuration"
	"openabyss/utils"
	"os"
	"path"
	"path/filepath"
	"time"
)

// Checks and handles backup retention
func check_retention_expiration(dirPath string) {
	filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("[backup_manager]: error walking path '%s': %v", dirPath, err)
			return err
		}

		// Ignore root path
		if path != dirPath {
			dt_mod_time := time.Now().UnixMilli() - info.ModTime().UnixMilli()

			// Check if retention time elapsed
			if dt_mod_time >= int64(configuration.LoadedConfig.Backup.RetentionPeriod) {
				fmt.Printf("[backup_manager]: removing backup '%s', retention expired. Dt[%d] Mod_time[%d]\n", info.Name(), dt_mod_time, info.ModTime().UnixMilli())
				if err := os.Remove(path); err != nil {
					fmt.Printf("[backup_manager]: error removing backup, %v\n", err)
				}
			}
		}

		return nil
	})
}

// Storage backup logic
func backup_current_storage(time_now_ms int64, storage_path string, backup_path string) {
	log.Printf("[backup_manager]: Backing up %s\n", InternalStoragePath)

	backup_zip_file, err := os.Create(path.Join(backup_path, fmt.Sprintf("storage_%d.zip", time_now_ms)))
	if err != nil {
		fmt.Printf("[backup_manager]: error creating file '%s': %v", backup_path, err)
		return
	}
	defer backup_zip_file.Close()
	gw := zip.NewWriter(backup_zip_file)
	defer gw.Close()

	filepath.WalkDir(storage_path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("[backup_manager]: error backing up '%s': %v", backup_path, err)
			return err
		}

		if !d.IsDir() {
			// Read file data
			data, _ := ioutil.ReadFile(path)

			log.Printf("[backup_manager]: Zipping up[%d]: %s\n", len(data), path)

			// Zip file with its data
			f, _ := gw.Create(path)
			f.Write(data)
		} else if d.Name() == BackupStoragePath {
			log.Printf("[backup_manager]: skipping backup storage directory %s\n", BackupStoragePath)
			return filepath.SkipDir
		}
		return nil
	})
}

// Periodically checks backup logic
func Init_Backup_Manager() {
	last_backup := time.Now().UnixMilli()

	// Keep track of Working Direcotory
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("could not get cwd", err)
	}

	for {
		if configuration.LoadedConfig.Backup.Enable {
			time.Sleep(time.Second)

			backup_dir := path.Join(wd, InternalStoragePath, BackupStoragePath)
			storage_dir := path.Join(wd, InternalStoragePath)
			if !utils.DirExists(backup_dir) {
				log.Printf("[backup_manager]: Creating backup directory: %s\n", backup_dir)
				os.MkdirAll(backup_dir, 0755)
			}
			check_retention_expiration(backup_dir)

			// Check and backup at set frequency
			time_now := time.Now().UnixMilli()
			dt_since_last_backup := time_now - last_backup
			if dt_since_last_backup >= int64(configuration.LoadedConfig.Backup.BackupFrequency) {
				backup_current_storage(time_now, storage_dir, backup_dir)
				last_backup = time_now
			}
		} else {
			log.Println("[backup_manager]: Backup is disabled. Exiting...")
			return
		}
	}
}
