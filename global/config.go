package global

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudogu/cesapp-lib/core"
	"github.com/cloudogu/cesapp-lib/keys"
	"github.com/cloudogu/cesapp-lib/registry"
	k8sErrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ConfigurationRegistry is able to manage the configuration of a single context
type ConfigurationRegistry interface {
	// Set sets a configuration value in current context
	Set(ctx context.Context, key, value string) error
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
	EtcdRegistry          registry.ConfigurationContext
	ClusterNativeRegistry ConfigurationRegistry
}

func NewGlobalConfig(regConfig core.Registry) (*Config, error) {
	etcdRegistry, err := registry.New(regConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd registry: %w", err)
	}

	return &Config{
		EtcdRegistry: etcdRegistry.GlobalConfig(),
		ClusterNativeRegistry: &clusterNativeConfigRegistry{
			prefix: "/config/_global",
		},
	}, nil
}

func NewDoguConfig(regConfig core.Registry, doguName string) (*Config, error) {
	etcdRegistry, err := registry.New(regConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd regist ry: %w", err)
	}

	return &Config{
		EtcdRegistry: etcdRegistry.DoguConfig(doguName),
		ClusterNativeRegistry: &clusterNativeConfigRegistry{
			prefix: fmt.Sprintf("/config/%s", doguName),
		},
	}, nil
}

func NewEncryptedDoguConfig(etcdReg registry.Registry, secretClient corev1client.SecretInterface, doguName string) (*Config, error) {
	keyType, err := etcdReg.GlobalConfig().Get("key_provider")
	if err != nil {
		return nil, fmt.Errorf("failed to get key provider type: %w", err)
	}

	keyProvider, err := keys.NewKeyProvider(keyType)
	if err != nil {
		return nil, fmt.Errorf("failed to create key provider: %w", err)
	}

	secret, err := secretClient.Get(context.Background(), fmt.Sprintf("private-%s", doguName), metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get private pem from dogu secret: %w", err)
	}

	privateKeyString := secret.Data["private.pem"]

	privateKey, err := keyProvider.FromPrivateKey(privateKeyString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &Config{
		EtcdRegistry:          newEncryptedEtcdRegistry(etcdReg.DoguConfig(doguName), privateKey.Public(), privateKey.Private()),
		ClusterNativeRegistry: nil, // TODO implement
	}, nil
}

func (c Config) Set(ctx context.Context, key, value string) error {
	cnErr := c.ClusterNativeRegistry.Set(ctx, key, value)
	if cnErr != nil {
		cnErr = fmt.Errorf("failed to set key in cluster native registry: %w", cnErr)
	}

	etcdErr := c.EtcdRegistry.Set(key, value)
	if etcdErr != nil {
		etcdErr = fmt.Errorf("failed to set key in etcd registry: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) Delete(ctx context.Context, key string) error {
	cnErr := c.ClusterNativeRegistry.Delete(ctx, key)
	if cnErr != nil {
		cnErr = fmt.Errorf("failed to delete key in cluster native registry: %w", cnErr)
	}

	etcdErr := c.EtcdRegistry.Delete(key)
	if etcdErr != nil {
		etcdErr = fmt.Errorf("failed to delete key in etcd registry: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) DeleteRecursive(ctx context.Context, key string) error {
	cnErr := c.ClusterNativeRegistry.DeleteRecursive(ctx, key)
	if cnErr != nil {
		cnErr = fmt.Errorf("failed to delete recursive in cluster native registry: %w", cnErr)
	}

	etcdErr := c.EtcdRegistry.DeleteRecursive(key)
	if etcdErr != nil {
		etcdErr = fmt.Errorf("failed to delete recursive in etcd registry: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) RemoveAll(ctx context.Context) error {
	cnErr := c.ClusterNativeRegistry.RemoveAll(ctx)
	if cnErr != nil {
		cnErr = fmt.Errorf("failed to remove all in cluster native registry: %w", cnErr)
	}

	etcdErr := c.EtcdRegistry.RemoveAll()
	if etcdErr != nil {
		etcdErr = fmt.Errorf("failed to remove all in etcd registry: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) Get(ctx context.Context, key string) (string, error) {
	logger := log.FromContext(ctx).WithName("ConfigurationRegistry.Get")
	value, err := c.ClusterNativeRegistry.Get(ctx, key)

	if k8sErrs.IsNotFound(err) {
		logger.Error(err, fmt.Sprintf("could not find key '%s' in cluster native registry, falling back to etcd", key))
		value, err = c.EtcdRegistry.Get(key)
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
	value, err := c.ClusterNativeRegistry.GetAll(ctx)

	if k8sErrs.IsNotFound(err) {
		logger.Error(err, "could not find all in cluster native registry, falling back to etcd")
		value, err = c.EtcdRegistry.GetAll()
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
	exists, err := c.ClusterNativeRegistry.Exists(ctx, key)

	if k8sErrs.IsNotFound(err) {
		logger.Error(err, fmt.Sprintf("could not find key '%s' in cluster native registry, falling back to etcd", key))
		exists, err = c.EtcdRegistry.Exists(key)
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
	exists, value, err := c.ClusterNativeRegistry.GetOrFalse(ctx, key)

	if k8sErrs.IsNotFound(err) {
		logger.Error(err, fmt.Sprintf("could not find key '%s' in cluster native registry, falling back to etcd", key))
		exists, value, err = c.EtcdRegistry.GetOrFalse(key)

		if err != nil {
			return false, "", fmt.Errorf("failed to get key from etcd: %w", err)
		}
	} else if err != nil {
		return false, "", fmt.Errorf("failed to get key from cluster native registry: %w", err)
	}

	return exists, value, nil
}
