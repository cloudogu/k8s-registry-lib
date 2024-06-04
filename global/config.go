package global

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudogu/cesapp-lib/core"
	"github.com/cloudogu/cesapp-lib/registry"
	k8sErrs "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
		cnErr = fmt.Errorf("failed to set key in cluster native registry: %w", cnErr)
	}

	etcdErr := c.etcdRegistry.Set(key, value)
	if etcdErr != nil {
		etcdErr = fmt.Errorf("failed to set key in etcd registry: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) SetWithLifetime(ctx context.Context, key, value string, timeToLiveInSeconds int) error {
	cnErr := c.clusterNativeRegistry.SetWithLifetime(ctx, key, value, timeToLiveInSeconds)
	if cnErr != nil {
		cnErr = fmt.Errorf("failed to set key with lifetime in cluster native registry: %w", cnErr)
	}

	etcdErr := c.etcdRegistry.SetWithLifetime(key, value, timeToLiveInSeconds)
	if etcdErr != nil {
		etcdErr = fmt.Errorf("failed to set key with lifetime in etcd registry: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) Refresh(ctx context.Context, key string, timeToLiveInSeconds int) error {
	cnErr := c.clusterNativeRegistry.Refresh(ctx, key, timeToLiveInSeconds)
	if cnErr != nil {
		cnErr = fmt.Errorf("failed to refresh key lifetime in cluster native registry: %w", cnErr)
	}

	etcdErr := c.etcdRegistry.Refresh(key, timeToLiveInSeconds)
	if etcdErr != nil {
		etcdErr = fmt.Errorf("failed to refresh key lifetime in etcd registry: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) Delete(ctx context.Context, key string) error {
	cnErr := c.clusterNativeRegistry.Delete(ctx, key)
	if cnErr != nil {
		cnErr = fmt.Errorf("failed to delete key in cluster native registry: %w", cnErr)
	}

	etcdErr := c.etcdRegistry.Delete(key)
	if etcdErr != nil {
		etcdErr = fmt.Errorf("failed to delete key in etcd registry: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) DeleteRecursive(ctx context.Context, key string) error {
	cnErr := c.clusterNativeRegistry.DeleteRecursive(ctx, key)
	if cnErr != nil {
		cnErr = fmt.Errorf("failed to delete recursive in cluster native registry: %w", cnErr)
	}

	etcdErr := c.etcdRegistry.DeleteRecursive(key)
	if etcdErr != nil {
		etcdErr = fmt.Errorf("failed to delete recursive in etcd registry: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) RemoveAll(ctx context.Context) error {
	cnErr := c.clusterNativeRegistry.RemoveAll(ctx)
	if cnErr != nil {
		cnErr = fmt.Errorf("failed to remove all in cluster native registry: %w", cnErr)
	}

	etcdErr := c.etcdRegistry.RemoveAll()
	if etcdErr != nil {
		etcdErr = fmt.Errorf("failed to remove all in etcd registry: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) Get(ctx context.Context, key string) (string, error) {
	logger := log.FromContext(ctx).WithName("ConfigurationRegistry.Get")
	value, err := c.clusterNativeRegistry.Get(ctx, key)

	if k8sErrs.IsNotFound(err) {
		logger.Error(err, fmt.Sprintf("could not find key '%s' in cluster native registry, falling back to etcd", key))
		value, err = c.etcdRegistry.Get(key)
		if err != nil {
			return "", fmt.Errorf("failed to get key from etcd: %w", err)
		}
	} else if err != nil {
		return "", fmt.Errorf("failed to get key from cluster native registry: %w", err)
	}

	return value, nil
}

func (c Config) GetAll(ctx context.Context) (map[string]string, error) {
	logger := log.FromContext(ctx).WithName("ConfigurationRegistry.GetAll")
	value, err := c.clusterNativeRegistry.GetAll(ctx)

	if k8sErrs.IsNotFound(err) {
		logger.Error(err, "could not find all in cluster native registry, falling back to etcd")
		value, err = c.etcdRegistry.GetAll()
		if err != nil {
			return nil, fmt.Errorf("failed to get all from etcd: %w", err)
		}

	} else if err != nil {
		return nil, fmt.Errorf("failed to get all from cluster native registry: %w", err)
	}

	return value, nil
}

func (c Config) Exists(ctx context.Context, key string) (bool, error) {
	logger := log.FromContext(ctx).WithName("ConfigurationRegistry.Exists")
	exists, err := c.clusterNativeRegistry.Exists(ctx, key)

	if k8sErrs.IsNotFound(err) {
		logger.Error(err, fmt.Sprintf("could not find key '%s' in cluster native registry, falling back to etcd", key))
		exists, err = c.etcdRegistry.Exists(key)
		if err != nil {
			return false, fmt.Errorf("failed to read key from etcd: %w", err)
		}

	} else if err != nil {
		return false, fmt.Errorf("failed to read key from cluster native registry: %w", err)
	}

	return exists, nil
}

func (c Config) GetOrFalse(ctx context.Context, key string) (bool, string, error) {
	logger := log.FromContext(ctx).WithName("ConfigurationRegistry.GetOrFalse")
	exists, value, err := c.clusterNativeRegistry.GetOrFalse(ctx, key)

	if k8sErrs.IsNotFound(err) {
		logger.Error(err, fmt.Sprintf("could not find key '%s' in cluster native registry, falling back to etcd", key))
		exists, value, err = c.etcdRegistry.GetOrFalse(key)

		if err != nil {
			return false, "", fmt.Errorf("failed to get key from etcd: %w", err)
		}
	} else if err != nil {
		return false, "", fmt.Errorf("failed to get key from cluster native registry: %w", err)
	}

	return exists, value, nil
}
