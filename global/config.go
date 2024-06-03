package global

import (
	"errors"
	"fmt"
	"github.com/cloudogu/cesapp-lib/core"
	"github.com/cloudogu/cesapp-lib/registry"
	"github.com/cloudogu/k8s-registry-lib/dogu/local"
)

type Config struct {
	etcdRegistry          local.ConfigurationRegistry
	clusterNativeRegistry local.ConfigurationRegistry
}

func NewConfig() (*Config, error) {
	return newCombinedRegistry("config/_global")
}

func newCombinedRegistry(prefix string) (*Config, error) {
	etcdRegistry, err := registry.New(core.Registry{}) // TODO IMPLEMENT PARAM
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd registry: %w", err)
	}

	var reg local.ConfigurationRegistry
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

func (c Config) Set(key, value string) error {
	cnErr := c.clusterNativeRegistry.Set(key, value)
	if cnErr != nil {
		cnErr = fmt.Errorf("tbd: %w", cnErr)
	}

	etcdErr := c.etcdRegistry.Set(key, value)
	if etcdErr != nil {
		etcdErr = fmt.Errorf("tbd: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) SetWithLifetime(key, value string, timeToLiveInSeconds int) error {
	cnErr := c.clusterNativeRegistry.SetWithLifetime(key, value, timeToLiveInSeconds)
	if cnErr != nil {
		cnErr = fmt.Errorf("tbd: %w", cnErr)
	}

	etcdErr := c.etcdRegistry.SetWithLifetime(key, value, timeToLiveInSeconds)
	if etcdErr != nil {
		etcdErr = fmt.Errorf("tbd: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) Refresh(key string, timeToLiveInSeconds int) error {
	cnErr := c.clusterNativeRegistry.Refresh(key, timeToLiveInSeconds)
	if cnErr != nil {
		cnErr = fmt.Errorf("tbd: %w", cnErr)
	}

	etcdErr := c.etcdRegistry.Refresh(key, timeToLiveInSeconds)
	if etcdErr != nil {
		etcdErr = fmt.Errorf("tbd: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) Delete(key string) error {
	cnErr := c.clusterNativeRegistry.Delete(key)
	if cnErr != nil {
		cnErr = fmt.Errorf("tbd: %w", cnErr)
	}

	etcdErr := c.etcdRegistry.Delete(key)
	if etcdErr != nil {
		etcdErr = fmt.Errorf("tbd: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) DeleteRecursive(key string) error {
	cnErr := c.clusterNativeRegistry.DeleteRecursive(key)
	if cnErr != nil {
		cnErr = fmt.Errorf("tbd: %w", cnErr)
	}

	etcdErr := c.etcdRegistry.DeleteRecursive(key)
	if etcdErr != nil {
		etcdErr = fmt.Errorf("tbd: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) RemoveAll() error {
	cnErr := c.clusterNativeRegistry.RemoveAll()
	if cnErr != nil {
		cnErr = fmt.Errorf("tbd: %w", cnErr)
	}

	etcdErr := c.etcdRegistry.RemoveAll()
	if etcdErr != nil {
		etcdErr = fmt.Errorf("tbd: %w", etcdErr)
	}

	return errors.Join(cnErr, etcdErr)
}

func (c Config) Get(key string) (string, error) {
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
}

func (c Config) GetAll() (map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func (c Config) Exists(key string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (c Config) GetOrFalse(key string) (bool, string, error) {
	//TODO implement me
	panic("implement me")
}
