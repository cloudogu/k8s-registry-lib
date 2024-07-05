package config

// DoguConfig represents a Dogu-specific configuration.
type DoguConfig struct {
	DoguName SimpleDoguName
	Config
}

// CreateDoguConfig creates a new Dogu-specific configuration with the provided Dogu name and entries.
func CreateDoguConfig(dogu SimpleDoguName, e Entries) DoguConfig {
	return DoguConfig{
		DoguName: dogu,
		Config:   CreateConfig(e),
	}
}
