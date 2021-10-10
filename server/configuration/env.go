package configuration

// Configuraiton Structure
type Configuration struct {
	DefaultKeyAlgorithm string `json:"defaultKeyAlgorithm"`
	Insecure            bool   `json:"insecure"`
	GrpcPort            uint16 `json:"grpcPort"`
	GrpcHost            string `json:"grpcHost"`
	TLSCertPath         string `json:"tlsCertPath"`
	TLSKeyPath          string `json:"tlsKeyPath"`
}

// Assigned Default Values
var (
	LoadedConfig = Configuration{
		DefaultKeyAlgorithm: "rsa",
		Insecure:            false,
		GrpcHost:            "0.0.0.0",
		GrpcPort:            50051,
		TLSCertPath:         "cert/server.crt",
		TLSKeyPath:          "cert/server.key",
	}
	InternalConfigDirPath string = ".config"            // Directory that holds configs
	ConfigFileName        string = "config-server.json" // Configuration JSON Filename
	InternalConfigPath    string                        // Absolute path to Configuration file
)
