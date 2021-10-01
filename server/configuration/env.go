package configuration

// Configuraiton Structure
type Configuration struct {
	DefaultKeyAlgorithm string `json:"defaultKeyAlgorithm"`
}

// Assigned Default Values
var (
	LoadedConfig = Configuration{
		DefaultKeyAlgorithm: "rsa",
	}
	InternalConfigDirPath string = ".config"            // Directory that holds configs
	ConfigFileName        string = "config-server.json" // Configuration JSON Filename
	InternalConfigPath    string                        // Absolute path to Configuration file
)
