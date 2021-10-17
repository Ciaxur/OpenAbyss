package main

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	pb "openabyss/proto/server"
	"openabyss/server/configuration"
	"openabyss/server/storage"
	"openabyss/utils"
	"os"
	"path"
	"path/filepath"
	"time"
)

// Lists internal storage backups
func (s openabyss_server) ListInternalBackups(ctx context.Context, in *pb.EmptyMessage) (*pb.BackupEntries, error) {
	// Construct full path to backups
	backupFullPath := path.Join(storage.InternalStoragePath, storage.BackupStoragePath)
	log.Printf("Walking through %s\n", backupFullPath)

	// Cosntruct response
	resp := pb.BackupEntries{
		Backups: []*pb.BackupEntry{},
	}

	filepath.Walk(backupFullPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Printf("[Internal Backup]: Walk error: %v\n", err)
			return err
		}

		if info.Name() != storage.BackupStoragePath {
			dt_mod_time := time.Now().UnixMilli() - info.ModTime().UnixMilli()
			time_till_expire := uint64(int64(configuration.LoadedConfig.Backup.RetentionPeriod) - dt_mod_time)

			resp.Backups = append(resp.Backups, &pb.BackupEntry{
				FileName:               info.Name(),
				CreatedUnixTimestamp:   uint64(info.ModTime().UnixMilli()),
				ExpiresInUnixTimestamp: time_till_expire,
			})
		}

		return nil
	})

	log.Printf("[Internal Backup]: Request handled %d internal backups\n", len(resp.Backups))
	return &resp, nil
}

// Invokes a new storage backup
func (s openabyss_server) InvokeNewStorageBackup(ctx context.Context, in *pb.EmptyMessage) (*pb.BackupEntry, error) {
	// Invoke new Backup
	backup_filepath := storage.InvokeNewBackup()

	// Generate Backup File Information
	time_now := time.Now().UnixMilli()
	dt_mod_time := time.Now().UnixMilli() - time_now
	time_till_expire := uint64(int64(configuration.LoadedConfig.Backup.RetentionPeriod) - dt_mod_time)

	// Return result of backup invokation
	return &pb.BackupEntry{
		FileName:               path.Base(backup_filepath),
		CreatedUnixTimestamp:   uint64(time_now),
		ExpiresInUnixTimestamp: time_till_expire,
	}, nil
}

// Returns current Backup Manager's Confifiguration
func (s openabyss_server) GetBackupManagerConfig(ctx context.Context, in *pb.EmptyMessage) (*pb.BackupManagerStatus, error) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("could not get cwd", err)
	}
	backup_dir := path.Join(wd, storage.InternalStoragePath, storage.BackupStoragePath)
	last_backup := int64(0)
	total_backups := 0

	filepath.Walk(backup_dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("[server_rpc_backups]: error walking path '%s': %v", backup_dir, err)
			return err
		}

		// Ignore root path
		if path != backup_dir {
			// Keep track of latest backup
			last_mod_milli := info.ModTime().UnixMilli()
			if last_mod_milli > last_backup {
				last_backup = last_mod_milli
			}

			// Keep track of how many backups are stored
			total_backups += 1
		}

		return nil
	})

	return &pb.BackupManagerStatus{
		IsEnabled:               configuration.LoadedConfig.Backup.Enable,
		LastBackupUnixTimestamp: uint64(last_backup),
		TotalBackups:            uint64(total_backups),
		RetentionPeriod:         configuration.LoadedConfig.Backup.RetentionPeriod,
		BackupFrequency:         configuration.LoadedConfig.Backup.BackupFrequency,
	}, nil
}

// Set Backup Manager Configuration
func (s openabyss_server) SetBackupManagerConfig(ctx context.Context, in *pb.BackupManagerStatus) (*pb.BackupManagerStatus, error) {
	// Re-init Backup Manager if there was a change in toggle
	if in.IsEnabled && !configuration.LoadedConfig.Backup.Enable {
		go storage.Init_Backup_Manager()
	}

	// Modify Backup Manager's Config
	configuration.LoadedConfig.Backup = configuration.BackupSubConfiguration{
		Enable:          in.IsEnabled,
		RetentionPeriod: in.RetentionPeriod,
		BackupFrequency: in.BackupFrequency,
	}

	return in, nil
}

// Deletes Stored backup based on Index
func (s openabyss_server) DeleteBackup(ctx context.Context, in *pb.BackupEntryRequest) (*pb.BackupEntry, error) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("could not get cwd", err)
	}
	backup_dir := path.Join(wd, storage.InternalStoragePath, storage.BackupStoragePath)
	backup_path := path.Join(backup_dir, in.BackupFileName)
	stat, err := os.Stat(backup_path)
	if err != nil {
		log.Println("[rpc_delete_backup]: Stat failed: ", err)
		return nil, fmt.Errorf("given backup file '%s' not found", in.BackupFileName)
	}

	// Attempt to Remove file
	if err := os.Remove(path.Join(backup_dir, in.BackupFileName)); err != nil {
		log.Println("[rpc_delete_backup]: File removal failed: ", err)
		return nil, fmt.Errorf("failed to remove '%s'", in.BackupFileName)
	}

	log.Println("[rpc_delete_backup]: Successfuly removed backup ", in.BackupFileName)
	return &pb.BackupEntry{
		FileName:               stat.Name(),
		CreatedUnixTimestamp:   uint64(stat.ModTime().UnixMilli()),
		ExpiresInUnixTimestamp: 0,
	}, nil
}

// Exports given backup to the client
func (s openabyss_server) ExportBackup(ctx context.Context, in *pb.BackupEntryRequest) (*pb.ExportedBackupResponse, error) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("could not get cwd", err)
	}
	backup_dir := path.Join(wd, storage.InternalStoragePath, storage.BackupStoragePath)
	backup_path := path.Join(backup_dir, in.BackupFileName)
	stat, err := os.Stat(backup_path)
	if err != nil {
		log.Println("[rpc_export_backup]: Stat failed: ", err)
		return nil, fmt.Errorf("given backup file '%s' not found", in.BackupFileName)
	}

	// Read backup file
	fileData, err := os.ReadFile(backup_path)
	if err != nil {
		log.Println("[rpc_export_backup]: failed to read file: ", err)
		return nil, fmt.Errorf("given backup file '%s' read error", in.BackupFileName)
	}

	log.Printf("[rpc_export_backup]: Successfuly exported '%s' backup file\n", in.BackupFileName)
	return &pb.ExportedBackupResponse{
		FileName:             stat.Name(),
		CreatedUnixTimestamp: uint64(stat.ModTime().UnixMilli()),
		FileData:             fileData,
	}, nil
}

// Imports given backup to server backups
func (s openabyss_server) ImportBackup(ctx context.Context, in *pb.ImportBackupRequest) (*pb.EmptyMessage, error) {
	log.Printf("[rpc_import_backup]: Attempting to import '%s'...\n", in.FileName)

	// Construct path to import to
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("could not get cwd", err)
	}
	backup_dir := path.Join(wd, storage.InternalStoragePath, storage.BackupStoragePath)
	backup_path := path.Join(backup_dir, in.FileName)

	// Verify no duplicates
	if utils.FileExists(backup_path) {
		log.Printf("[rpc_import_backup]: Duplicate backup found '%s': %v\n", backup_path, err)
		return &pb.EmptyMessage{}, fmt.Errorf("duplicate backup already exists '%s'", in.FileName)
	}

	// Store Backup
	if err := ioutil.WriteFile(backup_path, in.FileData, 0664); err != nil {
		log.Printf("[rpc_import_backup]: Failed to write imported data to '%s': %v\n", backup_path, err)
		return &pb.EmptyMessage{}, fmt.Errorf("failed to import '%s'", in.FileName)
	}

	log.Println("[rpc_import_backup]: Successfully imported ", backup_path)
	return &pb.EmptyMessage{}, nil
}
