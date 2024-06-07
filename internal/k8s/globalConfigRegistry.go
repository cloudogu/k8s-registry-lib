package k8s

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
)

type globalConfigMapRepo interface {
	GetGlobalConfig(context.Context) (config.GlobalConfig, error)
	WriteGlobalConfigMap(context.Context, config.GlobalConfig) error
}

type GlobalConfigRegistry struct {
	repo globalConfigMapRepo
}

func CreateGlobalConfigRegistry(configMapClient ConfigMapClient) GlobalConfigRegistry {
	return GlobalConfigRegistry{
		repo: CreateGlobalConfigRepo(configMapClient),
	}
}

func (gr GlobalConfigRegistry) Set(ctx context.Context, key, value string) error {
	globalConfig, err := gr.repo.GetGlobalConfig(ctx)
	if err != nil {
		return fmt.Errorf("could not read global config: %w", err)
	}

	globalConfig.Set(key, value)

	err = gr.repo.WriteGlobalConfigMap(ctx, globalConfig)
	if err != nil {
		return fmt.Errorf("could not write global config after updating value: %w", err)
	}

	return nil
}

// Exists returns true if configuration key exists
func (gr GlobalConfigRegistry) Exists(ctx context.Context, key string) (bool, error) {
	globalConfig, err := gr.repo.GetGlobalConfig(ctx)
	if err != nil {
		return false, fmt.Errorf("could not read global config: %w", err)
	}

	return globalConfig.Exists(key), nil
}

// Get returns the configuration value for the given key.
// Returns an error if no values exists for the given key.
func (gr GlobalConfigRegistry) Get(ctx context.Context, key string) (string, error) {
	globalConfig, err := gr.repo.GetGlobalConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("could not read global config: %w", err)
	}

	return globalConfig.Get(key)
}

// GetOrFalse returns false and an empty string when the configuration value does not exist.
// Otherwise, returns true and the configuration value, even when the configuration value is an empty string.
func (gr GlobalConfigRegistry) GetOrFalse(ctx context.Context, key string) (bool, string, error) {
	globalConfig, err := gr.repo.GetGlobalConfig(ctx)
	if err != nil {
		return false, "", fmt.Errorf("could not read global config: %w", err)
	}

	value, ok := globalConfig.GetOrFalse(key)

	return ok, value, nil
}

// GetAll returns a map of all key-value-pairs
func (gr GlobalConfigRegistry) GetAll(ctx context.Context) (map[string]string, error) {
	globalConfig, err := gr.repo.GetGlobalConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not read global config: %w", err)
	}

	return globalConfig.GetAll(), nil
}

// Delete removes the configuration key and value
func (gr GlobalConfigRegistry) Delete(ctx context.Context, key string) error {
	globalConfig, err := gr.repo.GetGlobalConfig(ctx)
	if err != nil {
		return fmt.Errorf("could not read global config: %w", err)
	}

	err = globalConfig.Delete(key)
	if err != nil {
		return fmt.Errorf("could not delete value for key %s from global config: %w", key, err)
	}

	err = gr.repo.WriteGlobalConfigMap(ctx, globalConfig)
	if err != nil {
		return fmt.Errorf("could not write global config after deleting key %s: %w", key, err)
	}

	return nil
}

// DeleteRecursive removes all configuration for the given key, including all configuration for sub-keys
func (gr GlobalConfigRegistry) DeleteRecursive(ctx context.Context, key string) error {
	globalConfig, err := gr.repo.GetGlobalConfig(ctx)
	if err != nil {
		return fmt.Errorf("could not read global config: %w", err)
	}

	globalConfig.DeleteRecursive(key)

	err = gr.repo.WriteGlobalConfigMap(ctx, globalConfig)
	if err != nil {
		return fmt.Errorf("could not write global config after recursively deleting key %s: %w", key, err)
	}

	return nil
}

// RemoveAll TODO: Implement
func (gr GlobalConfigRegistry) RemoveAll(ctx context.Context) error {
	return errors.New("NOT IMPLEMENTED")
}
