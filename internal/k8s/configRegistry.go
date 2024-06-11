package k8s

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
)

const globalConfigMapName = "global"

func createConfigName(doguName string) string {
	return fmt.Sprintf("%s-config", doguName)
}

type ConfigRepository interface {
	get(ctx context.Context) (config.Config, error)
	delete(ctx context.Context) error
	write(ctx context.Context, cfg config.Config) error
}

type ConfigRegistry struct {
	repo ConfigRepository
}

func CreateGlobalConfigRegistry(cmClient ConfigMapClient) ConfigRegistry {
	return ConfigRegistry{
		repo: newConfigRepo(globalConfigMapName, &configMapClient{cmClient}, doguConfigType),
	}
}

func CreateDoguConfigRegistry(cmClient ConfigMapClient, doguName string) ConfigRegistry {
	return ConfigRegistry{
		repo: newConfigRepo(createConfigName(doguName), &configMapClient{cmClient}, doguConfigType),
	}
}

func CreateSensitiveDoguConfigRegistry(sc SecretClient, doguName string) ConfigRegistry {
	return ConfigRegistry{
		repo: newConfigRepo(createConfigName(doguName), &secretClient{sc}, doguConfigType),
	}
}

func (dr ConfigRegistry) Set(ctx context.Context, key, value string) error {
	doguConfig, err := dr.repo.get(ctx)
	if err != nil {
		if !errors.Is(err, ErrConfigNotFound) {
			return fmt.Errorf("could not read dogu config: %w", err)
		}

		//create new, empty doguConfig
		doguConfig = config.CreateConfig(make(config.Data))
	}

	doguConfig.Set(key, value)

	err = dr.repo.write(ctx, doguConfig)
	if err != nil {
		return fmt.Errorf("could not write dogu config after updating value: %w", err)
	}

	return nil
}

// Exists returns true if configuration key exists
func (dr ConfigRegistry) Exists(ctx context.Context, key string) (bool, error) {
	doguConfig, err := dr.repo.get(ctx)
	if err != nil {
		return false, fmt.Errorf("could not read dogu config: %w", err)
	}

	return doguConfig.Exists(key), nil
}

// Get returns the configuration value for the given key.
// Returns an error if no values exists for the given key.
func (dr ConfigRegistry) Get(ctx context.Context, key string) (string, error) {
	doguConfig, err := dr.repo.get(ctx)
	if err != nil {
		return "", fmt.Errorf("could not read dogu config: %w", err)
	}

	return doguConfig.Get(key)
}

// GetOrFalse returns false and an empty string when the configuration value does not exist.
// Otherwise, returns true and the configuration value, even when the configuration value is an empty string.
func (dr ConfigRegistry) GetOrFalse(ctx context.Context, key string) (bool, string, error) {
	doguConfig, err := dr.repo.get(ctx)
	if err != nil {
		return false, "", fmt.Errorf("could not read dogu config: %w", err)
	}

	value, ok := doguConfig.GetOrFalse(key)

	return ok, value, nil
}

// GetAll returns a map of all key-value-pairs
func (dr ConfigRegistry) GetAll(ctx context.Context) (map[string]string, error) {
	doguConfig, err := dr.repo.get(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not read dogu config: %w", err)
	}

	return doguConfig.GetAll(), nil
}

// Delete removes the configuration key and value
func (dr ConfigRegistry) Delete(ctx context.Context, key string) error {
	doguConfig, err := dr.repo.get(ctx)
	if err != nil {
		return fmt.Errorf("could not read dogu config: %w", err)
	}

	err = doguConfig.Delete(key)
	if err != nil {
		return fmt.Errorf("could not delete value for key %s from dogu config: %w", key, err)
	}

	err = dr.repo.write(ctx, doguConfig)
	if err != nil {
		return fmt.Errorf("could not write dogu config after deleting key %s: %w", key, err)
	}

	return nil
}

// DeleteRecursive removes all configuration for the given key, including all configuration for sub-keys
func (dr ConfigRegistry) DeleteRecursive(ctx context.Context, key string) error {
	doguConfig, err := dr.repo.get(ctx)
	if err != nil {
		return fmt.Errorf("could not read dogu config: %w", err)
	}

	doguConfig.DeleteRecursive(key)

	err = dr.repo.write(ctx, doguConfig)
	if err != nil {
		return fmt.Errorf("could not write dogu config after recursively deleting key %s: %w", key, err)
	}

	return nil
}

func (dr ConfigRegistry) RemoveAll(ctx context.Context) error {
	doguConfig, err := dr.repo.get(ctx)
	if err != nil {
		return fmt.Errorf("could not read dogu config: %w", err)
	}

	doguConfig.RemoveAll()

	if lErr := dr.repo.delete(ctx); lErr != nil {
		return fmt.Errorf("could not delete dogu config: %w", err)
	}

	if lErr := dr.repo.write(ctx, doguConfig); lErr != nil {
		return fmt.Errorf("could not write dogu config after deleting all keys: %w", err)
	}

	return nil
}
