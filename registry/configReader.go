package registry

import (
	"context"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
)

type configReader struct {
	repo configRepository
}

// Exists returns true if configuration key exists
func (cr configReader) Exists(ctx context.Context, key string) (bool, error) {
	cfg, err := cr.repo.get(ctx)
	if err != nil {
		return false, fmt.Errorf("could not read dogu config: %w", err)
	}

	return cfg.Exists(config.Key(key)), nil
}

// Get returns the configuration value for the given key.
// Returns an error if no values exists for the given key.
func (cr configReader) Get(ctx context.Context, key string) (string, error) {
	cfg, err := cr.repo.get(ctx)
	if err != nil {
		return "", fmt.Errorf("could not read dogu config: %w", err)
	}

	v, err := cfg.Get(config.Key(key))
	if err != nil {
		return "", fmt.Errorf("could not get value from config: %w", err)
	}

	return v.String(), nil
}

// GetAll returns a map of all key-value-pairs
func (cr configReader) GetAll(ctx context.Context) (map[string]string, error) {
	cfg, err := cr.repo.get(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not read dogu config: %w", err)
	}

	entries := cfg.GetAll()

	sEntries := make(map[string]string, len(entries))

	for k, v := range entries {
		sEntries[k.String()] = v.String()
	}

	return sEntries, nil
}
