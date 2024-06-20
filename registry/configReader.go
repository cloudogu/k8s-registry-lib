package registry

import (
	"context"
	"fmt"
)

type configReader struct {
	repo configRepository
}

// Exists returns true if configuration key exists
func (cr configReader) Exists(ctx context.Context, key string) (bool, error) {
	config, err := cr.repo.get(ctx)
	if err != nil {
		return false, fmt.Errorf("could not read dogu config: %w", err)
	}

	return config.Exists(key), nil
}

// Get returns the configuration value for the given key.
// Returns an error if no values exists for the given key.
func (cr configReader) Get(ctx context.Context, key string) (string, error) {
	config, err := cr.repo.get(ctx)
	if err != nil {
		return "", fmt.Errorf("could not read dogu config: %w", err)
	}

	return config.Get(key)
}

// GetAll returns a map of all key-value-pairs
func (cr configReader) GetAll(ctx context.Context) (map[string]string, error) {
	config, err := cr.repo.get(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not read dogu config: %w", err)
	}

	return config.GetAll(), nil
}
