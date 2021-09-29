package configuration

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"openabyss/utils"
	"os"
	"path"
)

func Init() {
	// Assemble internal storage paths
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("could not get cwd", err)
	}

	InternalConfigPath = path.Join(wd, InternalConfigDirPath, "config.json")

	// Attempt to Load in Configuration
	// Load file if exists
	if utils.FileExists(InternalConfigPath) {
		log.Printf("[configuration]: internal persistant file '%s' exists\n", InternalConfigPath)
		fileBuffer, _ := ioutil.ReadFile(InternalConfigPath)
		if err := json.Unmarshal(fileBuffer, &LoadedConfig); err != nil {
			log.Fatalln("configuration unmarshal error:", err)
		}

	} else {
		// Create Configuration directory
		log.Println("[configuration]: no internal persistant file found")
		log.Println("[configuration]: creating:", path.Dir(InternalConfigPath))
		if err := os.MkdirAll(path.Dir(InternalConfigPath), 0755); err != nil {
			log.Fatalf("could not created path '%s': %v\n", path.Dir(InternalConfigPath), err)
		}

		// Write empty data
		if file, err := os.Create(InternalConfigPath); err != nil {
			log.Fatalln("could not create internal storage file:", err)
		} else {
			data, _ := json.Marshal(LoadedConfig)
			file.Write(data)
			file.Close()
		}
	}
}
