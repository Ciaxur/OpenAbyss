package main

import (
	"context"
	"io/fs"
	"log"
	pb "openabyss/proto/server"
	"openabyss/server/configuration"
	"openabyss/server/storage"
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
