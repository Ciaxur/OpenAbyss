package configuration

// Configuraiton Structure
type Configuration struct {
	GrpcName string `json:"grpcName"`
	GrpcHost string `json:"grpcHost"`
	GrpcPort uint16 `json:"grpcPort"`
	Insecure bool   `json:"insecure"`
}

// Assigned Default Values
var (
	LoadedConfig = Configuration{
		GrpcName: "OpenAbyss-Client",
		GrpcHost: "localhost",
		GrpcPort: uint16(50051),
	}
	InternalConfigDirPath string = ".config"            // Directory that holds configs
	ConfigFileName        string = "config-client.json" // Configuration JSON Filename
	InternalConfigPath    string                        // Absolute path to Configuration file
	IsVerbose             bool   = false                // Verbose Printing
)
