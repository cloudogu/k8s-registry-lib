package config

import "github.com/cloudogu/ces-commons-lib/dogu"

// DoguConfig represents a Dogu-specific configuration.
type DoguConfig struct {
	DoguName dogu.SimpleName
	Config
}

// CreateDoguConfig creates a new Dogu-specific configuration with the provided Dogu name and entries.
func CreateDoguConfig(dogu dogu.SimpleName, e Entries) DoguConfig {
	return DoguConfig{
		DoguName: dogu,
		Config:   CreateConfig(e),
	}
}
