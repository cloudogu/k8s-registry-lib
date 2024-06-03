package global

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudogu/cesapp-lib/core"
	"github.com/cloudogu/cesapp-lib/registry"
)

// ConfigurationRegistry is able to manage the configuration of a single context
type ConfigurationRegistry interface {
	// Set sets a configuration value in current context
	Set(ctx context.Context, key, value string) error
	// SetWithLifetime sets a configuration value in current context with the given lifetime
	SetWithLifetime(ctx context.Context, key, value string, timeToLiveInSeconds int) error
	// Refresh resets the time to live of a key
	Refresh(ctx context.Context, key string, timeToLiveInSeconds int) error
	// Get returns a configuration value from the current context
	Get(ctx context.Context, key string) (string, error)
	// GetAll returns a map of key value pairs
	GetAll(ctx context.Context) (map[string]string, error)
	// Delete removes a configuration key and value from the current context
	Delete(ctx context.Context, key string) error
	// DeleteRecursive removes a configuration key or directory from the current context
	DeleteRecursive(ctx context.Context, key string) error
	// Exists returns true if configuration key exists in the current context
	Exists(ctx context.Context, key string) (bool, error)
	// RemoveAll remove all configuration keys
	RemoveAll(ctx context.Context) error
	// GetOrFalse return false and empty string when the configuration value does not exist.
	// Otherwise, return true and the configuration value, even when the configuration value is an empty string.
	GetOrFalse(ctx context.Context, key string) (bool, string, error)
}

type Config struct {
	etcdRegistry          registry.ConfigurationContext
	clusterNativeRegistry ConfigurationRegistry
}

func NewConfig() (*Config, error) {
	return newCombinedRegistry("config/_global")
}

func newCombinedRegistry(prefix string) (*Config, error) {
	etcdRegistry, err := registry.New(core.Registry{}) // TODO IMPLEMENT PARAM
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd registry: %w", err)
	}

	var reg registry.ConfigurationContext
	isGlobalRegistry := true // TODO implement + if-else / switch-case

	if isGlobalRegistry {
		reg = etcdRegistry.GlobalConfig()
	}

	return &Config{
		etcdRegistry: reg,
		clusterNativeRegistry: &clusterNativeConfigRegistry{
			prefix: prefix,
		},
	}, nil
}

func (c Config) Set(ctx context.Context, key, value string) error {
	cnErr := c.clusterNativeRegistry.Set(ctx, key, value)
	if cnErr != nil {
		cnErr = fmt.Errorf("tbd: %w", cnErr)
	}

	etcdErr := c.etcdRegistry.Set(key, value)
	if etcdErr != nil {
		etcdErr = fmt.Errorf("tbd: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) SetWithLifetime(ctx context.Context, key, value string, timeToLiveInSeconds int) error {
	cnErr := c.clusterNativeRegistry.SetWithLifetime(ctx, key, value, timeToLiveInSeconds)
	if cnErr != nil {
		cnErr = fmt.Errorf("tbd: %w", cnErr)
	}

	etcdErr := c.etcdRegistry.SetWithLifetime(key, value, timeToLiveInSeconds)
	if etcdErr != nil {
		etcdErr = fmt.Errorf("tbd: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) Refresh(ctx context.Context, key string, timeToLiveInSeconds int) error {
	cnErr := c.clusterNativeRegistry.Refresh(ctx, key, timeToLiveInSeconds)
	if cnErr != nil {
		cnErr = fmt.Errorf("tbd: %w", cnErr)
	}

	etcdErr := c.etcdRegistry.Refresh(key, timeToLiveInSeconds)
	if etcdErr != nil {
		etcdErr = fmt.Errorf("tbd: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) Delete(ctx context.Context, key string) error {
	cnErr := c.clusterNativeRegistry.Delete(ctx, key)
	if cnErr != nil {
		cnErr = fmt.Errorf("tbd: %w", cnErr)
	}

	etcdErr := c.etcdRegistry.Delete(key)
	if etcdErr != nil {
		etcdErr = fmt.Errorf("tbd: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) DeleteRecursive(ctx context.Context, key string) error {
	cnErr := c.clusterNativeRegistry.DeleteRecursive(ctx, key)
	if cnErr != nil {
		cnErr = fmt.Errorf("tbd: %w", cnErr)
	}

	etcdErr := c.etcdRegistry.DeleteRecursive(key)
	if etcdErr != nil {
		etcdErr = fmt.Errorf("tbd: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) RemoveAll(ctx context.Context) error {
	cnErr := c.clusterNativeRegistry.RemoveAll(ctx)
	if cnErr != nil {
		cnErr = fmt.Errorf("tbd: %w", cnErr)
	}

	etcdErr := c.etcdRegistry.RemoveAll()
	if etcdErr != nil {
		etcdErr = fmt.Errorf("tbd: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) Get(ctx context.Context, key string) (string, error) {
	//logger := log.FromContext(ctx).
	//	WithName("CombinedLocalDoguRegistry.GetCurrent").
	//	WithValues("dogu.name", simpleDoguName)
	//dogu, err := c.clusterNativeRegistry.Get(key)
	//if k8sErrs.IsNotFound(err) {
	//	logger.Error(err, "current dogu.json not found in cluster-native local registry; falling back to ETCD")
	//
	//	dogu, err = cr.etcdRegistry.GetCurrent(ctx, simpleDoguName)
	//	if err != nil {
	//		return nil, fmt.Errorf("failed to get current dogu.json of %q from ETCD local registry (legacy/fallback): %w", simpleDoguName, err)
	//	}
	//
	//} else if err != nil {
	//	return nil, fmt.Errorf("failed to get current dogu.json of %q from cluster-native local registry: %w", simpleDoguName, err)
	//}
	//
	//return dogu, nil
	return "", nil
}

func (c Config) GetAll(ctx context.Context) (map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func (c Config) Exists(ctx context.Context, key string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (c Config) GetOrFalse(ctx context.Context, key string) (bool, string, error) {
	//TODO implement me
	panic("implement me")
}
