package k8s

import (
	"context"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
)

type doguConfigMapRepo interface {
	GetDoguConfig(ctx context.Context, doguName string) (config.DoguConfig, error)
	DeleteDoguConfig(ctx context.Context, doguName string) error
	WriteDoguConfigMap(ctx context.Context, cfg config.DoguConfig) error
}

type DoguConfigRegistry struct {
	doguName string
	repo     doguConfigMapRepo
}

func CreateDoguConfigRegistry(configMapClient ConfigMapClient, doguName string) DoguConfigRegistry {
	return DoguConfigRegistry{
		doguName: doguName,
		repo:     CreateDoguConfigRepo(configMapClient),
	}
}

func (dr DoguConfigRegistry) Set(ctx context.Context, key, value string) error {
	doguConfig, err := dr.repo.GetDoguConfig(ctx, dr.doguName)
	if err != nil {
		return fmt.Errorf("could not read dogu config: %w", err)
	}

	doguConfig.Set(key, value)

	err = dr.repo.WriteDoguConfigMap(ctx, doguConfig)
	if err != nil {
		return fmt.Errorf("could not write dogu config after updating value: %w", err)
	}

	return nil
}

// Exists returns true if configuration key exists
func (dr DoguConfigRegistry) Exists(ctx context.Context, key string) (bool, error) {
	doguConfig, err := dr.repo.GetDoguConfig(ctx, dr.doguName)
	if err != nil {
		return false, fmt.Errorf("could not read dogu config: %w", err)
	}

	return doguConfig.Exists(key), nil
}

// Get returns the configuration value for the given key.
// Returns an error if no values exists for the given key.
func (dr DoguConfigRegistry) Get(ctx context.Context, key string) (string, error) {
	doguConfig, err := dr.repo.GetDoguConfig(ctx, dr.doguName)
	if err != nil {
		return "", fmt.Errorf("could not read dogu config: %w", err)
	}

	return doguConfig.Get(key)
}

// GetOrFalse returns false and an empty string when the configuration value does not exist.
// Otherwise, returns true and the configuration value, even when the configuration value is an empty string.
func (dr DoguConfigRegistry) GetOrFalse(ctx context.Context, key string) (bool, string, error) {
	doguConfig, err := dr.repo.GetDoguConfig(ctx, dr.doguName)
	if err != nil {
		return false, "", fmt.Errorf("could not read dogu config: %w", err)
	}

	value, ok := doguConfig.GetOrFalse(key)

	return ok, value, nil
}

// GetAll returns a map of all key-value-pairs
func (dr DoguConfigRegistry) GetAll(ctx context.Context) (map[string]string, error) {
	doguConfig, err := dr.repo.GetDoguConfig(ctx, dr.doguName)
	if err != nil {
		return nil, fmt.Errorf("could not read dogu config: %w", err)
	}

	return doguConfig.GetAll(), nil
}

// Delete removes the configuration key and value
func (dr DoguConfigRegistry) Delete(ctx context.Context, key string) error {
	doguConfig, err := dr.repo.GetDoguConfig(ctx, dr.doguName)
	if err != nil {
		return fmt.Errorf("could not read dogu config: %w", err)
	}

	err = doguConfig.Delete(key)
	if err != nil {
		return fmt.Errorf("could not delete value for key %s from dogu config: %w", key, err)
	}

	err = dr.repo.WriteDoguConfigMap(ctx, doguConfig)
	if err != nil {
		return fmt.Errorf("could not write dogu config after deleting key %s: %w", key, err)
	}

	return nil
}

// DeleteRecursive removes all configuration for the given key, including all configuration for sub-keys
func (dr DoguConfigRegistry) DeleteRecursive(ctx context.Context, key string) error {
	doguConfig, err := dr.repo.GetDoguConfig(ctx, dr.doguName)
	if err != nil {
		return fmt.Errorf("could not read dogu config: %w", err)
	}

	doguConfig.DeleteRecursive(key)

	err = dr.repo.WriteDoguConfigMap(ctx, doguConfig)
	if err != nil {
		return fmt.Errorf("could not write dogu config after recursively deleting key %s: %w", key, err)
	}

	return nil
}

func (dr DoguConfigRegistry) RemoveAll(ctx context.Context) error {
	doguConfig, err := dr.repo.GetDoguConfig(ctx, dr.doguName)
	if err != nil {
		return fmt.Errorf("could not read dogu config: %w", err)
	}

	doguConfig.RemoveAll()

	if lErr := dr.repo.DeleteDoguConfig(ctx, doguConfig.Name); lErr != nil {
		return fmt.Errorf("could not delete dogu config: %w", err)
	}

	if lErr := dr.repo.WriteDoguConfigMap(ctx, doguConfig); lErr != nil {
		return fmt.Errorf("could not write dogu config after deleting all keys: %w", err)
	}

	return nil
}
