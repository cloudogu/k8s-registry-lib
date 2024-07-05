package config

// SensitiveDoguConfig represents a Dogu-specific configuration that contains sensitive data for a dogu.
type SensitiveDoguConfig struct {
	DoguName SimpleDoguName
	Config
}

// CreateSensitiveDoguConfig creates a new Dogu-specific configuration with the provided Dogu name and sensitive entries.
func CreateSensitiveDoguConfig(dogu SimpleDoguName, e Entries) SensitiveDoguConfig {
	return SensitiveDoguConfig{
		DoguName: dogu,
		Config:   CreateConfig(e),
	}
}
