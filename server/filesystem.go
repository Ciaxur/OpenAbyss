package main

import (
	"context"
	pb "openabyss/proto/server"
	"openabyss/server/storage"
	"path"
)

// Lists internal filesystem storage contents
func (s openabyss_server) ListPathContents(ctx context.Context, in *pb.ListPathContentRequest) (*pb.PathResponse, error) {
	// Get "root" path from which to list path content of
	fsStorage, err := storage.Internal.GetSubStorageByPath(in.Path)
	if err != nil {
		return &pb.PathResponse{}, err
	}

	// Root directory keys (BFS Algorithm)
	dirQueue := []storage.FileStorageMap{*fsStorage}

	// Result
	internalStorage := &pb.PathResponse{
		Content: []*pb.ContentType{},
	}

	for ; len(dirQueue) != 0; dirQueue = dirQueue[1:] {
		// Enqueue
		fsSubStorage := dirQueue[0]

		// Add all sub-storages (for later)
		if in.Recursive {
			for _, sStorage := range fsSubStorage.StorageMap {
				dirQueue = append(dirQueue, sStorage)
			}
		}

		// Add all content to result
		for _, sContent := range fsSubStorage.Storage {
			internalStorage.Content = append(internalStorage.Content, &pb.ContentType{
				Name:                  path.Base(sContent.Path),
				Path:                  sContent.Path,
				SizeInBytes:           sContent.SizeInBytes,
				CreatedUnixTimestamp:  sContent.CreatedAt_UnixTimestamp,
				ModifiedUnixTimestamp: sContent.ModifiedAt_UnixTimestamp,
			})
		}
	}

	return internalStorage, nil
}
