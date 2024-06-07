package registry

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudogu/cesapp-lib/keys"
	"github.com/cloudogu/k8s-registry-lib/internal/etcd"
	"github.com/cloudogu/k8s-registry-lib/k8s"
	k8sErrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type globalGetter interface {
	GlobalConfig() etcd.ConfigurationContext
}

type doguConfigGetter interface {
	DoguConfig(dogu string) etcd.ConfigurationContext
}

type globalAndDoguConfigGetter interface {
	globalGetter
	doguConfigGetter
}

type configRegistry struct {
	EtcdRegistry          etcd.ConfigurationContext
	ClusterNativeRegistry ConfigurationRegistry
}

type GlobalRegistry struct {
	configRegistry
}

type DoguRegistry struct {
	configRegistry
}

type SensitiveDoguRegistry struct {
	configRegistry
}

func NewGlobalConfigRegistry(etcdClient globalGetter, k8sClient k8s.ConfigMapClient) *GlobalRegistry {
	return &GlobalRegistry{configRegistry{
		EtcdRegistry:          etcdClient.GlobalConfig(),
		ClusterNativeRegistry: k8s.CreateGlobalConfigRegistry(k8sClient),
	}}
}

func NewDoguConfigRegistry(doguName string, etcdClient doguConfigGetter, k8sClient k8s.ConfigMapClient) *DoguRegistry {
	return &DoguRegistry{configRegistry{
		EtcdRegistry:          etcdClient.DoguConfig(doguName),
		ClusterNativeRegistry: k8s.CreateDoguConfigRegistry(k8sClient, doguName),
	}}
}

func NewSensitiveDoguRegistry(etcdReg globalAndDoguConfigGetter, secretClient k8s.SecretClient, doguName string) (*SensitiveDoguRegistry, error) {
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

	return &SensitiveDoguRegistry{configRegistry{
		EtcdRegistry:          etcd.NewEncryptedRegistry(etcdReg.DoguConfig(doguName), privateKey.Public(), privateKey.Private()),
		ClusterNativeRegistry: nil, // TODO implement
	}}, nil
}

func (c configRegistry) Set(ctx context.Context, key, value string) error {
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

func (c configRegistry) Delete(ctx context.Context, key string) error {
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

func (c configRegistry) DeleteRecursive(ctx context.Context, key string) error {
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

func (c configRegistry) RemoveAll(ctx context.Context) error {
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

func (c configRegistry) Get(ctx context.Context, key string) (string, error) {
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

func (c configRegistry) GetAll(ctx context.Context) (map[string]string, error) {
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

func (c configRegistry) Exists(ctx context.Context, key string) (bool, error) {
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

func (c configRegistry) GetOrFalse(ctx context.Context, key string) (bool, string, error) {
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
