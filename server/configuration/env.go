package configuration

// Periodic Backup Sub-config
type BackupSubConfiguration struct {
	RetentionPeriod uint64 `json:"retentionPeriod"` // Retention Period for backups in Milliseconds
	BackupFrequency uint64 `json:"backupFrequency"` // How frequently to backup in Milliseconds
}

// Root Configuraiton Structure
type Configuration struct {
	DefaultKeyAlgorithm string                 `json:"defaultKeyAlgorithm"`
	Insecure            bool                   `json:"insecure"`
	GrpcPort            uint16                 `json:"grpcPort"`
	GrpcHost            string                 `json:"grpcHost"`
	TLSCertPath         string                 `json:"tlsCertPath"`
	TLSKeyPath          string                 `json:"tlsKeyPath"`
	Backup              BackupSubConfiguration `json:"backup"`
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
		Backup: BackupSubConfiguration{
			RetentionPeriod: 7 * 24 * 60 * 60 * 1000, // 7 Days by default
			BackupFrequency: 7 * 24 * 60 * 60 * 1000, // Daily backups by default
		},
	}
	InternalConfigDirPath string = ".config"            // Directory that holds configs
	ConfigFileName        string = "config-server.json" // Configuration JSON Filename
	InternalConfigPath    string                        // Absolute path to Configuration file
)
