package configuration

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"openabyss/utils"
	"os"
	"path"
)

func EnableVerbose() {
	IsVerbose = true
}

func DisableVerbose() {
	IsVerbose = false
}

func Init() {
	// Assemble internal storage paths
	bin_path, err := os.Executable()
	wd := path.Dir(bin_path)
	if err != nil {
		log.Fatalln("could not get cwd", err)
	}

	InternalConfigPath = path.Join(wd, InternalConfigDirPath, ConfigFileName)

	// Attempt to Load in Configuration
	// Load file if exists
	if utils.FileExists(InternalConfigPath) {
		if IsVerbose {
			log.Printf("[configuration]: internal persistant file '%s' exists\n", InternalConfigPath)
		}
		fileBuffer, _ := ioutil.ReadFile(InternalConfigPath)
		if err := json.Unmarshal(fileBuffer, &LoadedConfig); err != nil {
			log.Fatalln("configuration unmarshal error:", err)
		}

	} else {
		// Create Configuration directory
		if IsVerbose {
			log.Println("[configuration]: no internal persistant file found")
			log.Println("[configuration]: creating:", path.Dir(InternalConfigPath))
		}
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
