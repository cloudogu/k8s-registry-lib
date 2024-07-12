package config

// GlobalConfig represents a global configuration.
type GlobalConfig struct {
	Config
}

// CreateGlobalConfig creates a new global configuration with the provided entries.
func CreateGlobalConfig(e Entries) GlobalConfig {
	return GlobalConfig{
		Config: CreateConfig(e),
	}
}
