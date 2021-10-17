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

	InternalConfigPath = path.Join(wd, InternalStoragePath, "internal.json")
	log.Println("internal path", InternalConfigPath)

	// Check if internal file storeage exists
	if utils.FileExists(InternalConfigPath) {
		log.Printf("[storage]: internal persistant file '%s' exists\n", InternalConfigPath)
		fileBuffer, _ := ioutil.ReadFile(InternalConfigPath)
		if err := json.Unmarshal(fileBuffer, &Internal); err != nil {
			log.Fatalln("internal storage unmarshal error:", err)
		}

	} else {
		// Create Storage directory
		log.Println("[storage]: no internal persistant file found")
		log.Println("[storage]: creating:", path.Dir(InternalConfigPath))
		if err := os.MkdirAll(path.Dir(InternalConfigPath), 0755); err != nil {
			log.Fatalf("could not created path '%s': %v\n", path.Dir(InternalConfigPath), err)
		}

		// Write empty data
		if file, err := os.Create(InternalConfigPath); err != nil {
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
	log.Println("[storage]: Closing internal storage, writing to file...")
	_, err := Internal.WriteToFile()
	return err
}
