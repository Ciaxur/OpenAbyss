package main

import (
	"context"
	"errors"
	"log"
	pb "openabyss/proto/server"
	"openabyss/server/storage"
	"openabyss/utils"
	"os"
)

func (s openabyss_server) ModifyEntity(ctx context.Context, in *pb.EntityMod) (*pb.EmptyMessage, error) {
	// Check type of modification
	if in.Remove {
		log.Println("[ModifyEntity]: Requesed removal of", in.FilePath)

		// Remove internal storage
		if internalFilepath, err := storage.Internal.RemoveStorage(in.FilePath); err != nil {
			log.Printf("[ModifyEntity]: failed to remove internal storage '%s'\n", in.FilePath)
			return &pb.EmptyMessage{}, err
		} else {
			if _, err := storage.Internal.WriteToFile(); err != nil {
				utils.HandleErr(err, "failed to save internal storage to file after entry removal")
			}

			// Remove actual file
			if err := os.Remove(internalFilepath); err != nil {
				log.Printf("[ModifyEntity]: failed to remove actual file '%s'\n", internalFilepath)

				// Client doesn't need to know, since it'll be re-writen IFF this issue occured
				// This is a server-side issue only
				return &pb.EmptyMessage{}, nil
			}

			log.Println("[ModifyEntity]: Successfuly removed", internalFilepath)
			return &pb.EmptyMessage{}, nil
		}
	}

	return &pb.EmptyMessage{}, errors.New("unhandled request")
}
