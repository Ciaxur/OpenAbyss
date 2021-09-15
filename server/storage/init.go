package storage

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"openabyss/utils"
	"os"
	"path"
)

func Init() {
	// Check for available persistant data
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("could not get cwd", err)
	}

	InternalFilePath = path.Join(wd, ".storage", "internal.json")
	log.Println("internal path", InternalFilePath)

	// Check if internal file storeage exists
	if utils.FileExists(InternalFilePath) {
		log.Printf("[storage]: internal persistant file '%s' exists\n", InternalFilePath)
		fileBuffer, _ := ioutil.ReadFile(InternalFilePath)
		if err := json.Unmarshal(fileBuffer, &Internal); err != nil {
			log.Fatalln("internal storage unmarshal error:", err)
		}

	} else {
		// Create Storage directory
		log.Println("[storage]: no internal persistant file found")
		log.Println("[storage]: creating:", path.Dir(InternalFilePath))
		if err := os.MkdirAll(path.Dir(InternalFilePath), 0755); err != nil {
			log.Fatalf("could not created path '%s': %v\n", path.Dir(InternalFilePath), err)
		}

		// Write empty data
		if file, err := os.Create(InternalFilePath); err != nil {
			log.Fatalln("could not create internal storage file:", err)
		} else {
			data, _ := json.Marshal(Internal)
			file.Write(data)
			file.Close()
		}
	}

}

// Closes and cleans up internal data
func Close() error {
	_, err := Internal.WriteToFile()
	return err
}
