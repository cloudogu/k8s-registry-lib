package registry

import (
	"context"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
)

const globalConfigMapName = "global"

type configRepository interface {
	get(ctx context.Context) (config.Config, error)
	delete(ctx context.Context) error
	write(ctx context.Context, cfg config.Config) error
	watch(ctx context.Context) (*configWatch, error)
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

type configRegistry struct {
	configReader
	configWriter
	configWatcher
}

func NewGlobalConfigRegistry(ctx context.Context, k8sClient ConfigMapClient) (*GlobalRegistry, error) {
	repo, _ := newConfigRepo(withGeneralName(globalConfigMapName), createConfigMapClient(k8sClient, globalConfigType))

	if lErr := repo.write(ctx, config.CreateConfig(make(config.Entries))); lErr != nil {
		return nil, fmt.Errorf("could not create initial global config: %w", lErr)
	}

	return &GlobalRegistry{configRegistry{
		configReader{repo: repo},
		configWriter{repo: repo},
		configWatcher{repo: repo},
	}}, nil
}

func NewDoguConfigRegistry(ctx context.Context, doguName string, k8sClient ConfigMapClient) (*DoguRegistry, error) {
	repo, _ := newConfigRepo(withDoguName(doguName), createConfigMapClient(k8sClient, doguConfigType))

	if lErr := repo.write(ctx, config.CreateConfig(make(config.Entries))); lErr != nil {
		return nil, fmt.Errorf("could not create initial dogu config %s: %w", doguName, lErr)
	}

	return &DoguRegistry{configRegistry{
		configReader{repo: repo},
		configWriter{repo: repo},
		configWatcher{repo: repo},
	}}, nil
}

func NewSensitiveDoguRegistry(ctx context.Context, doguName string, sc SecretClient) (*SensitiveDoguRegistry, error) {
	repo, _ := newConfigRepo(withDoguName(doguName), createSecretClient(sc, sensitiveConfigType))

	if lErr := repo.write(ctx, config.CreateConfig(make(config.Entries))); lErr != nil {
		return nil, fmt.Errorf("could not create initial sensitive dogu config %s: %w", doguName, lErr)
	}

	return &SensitiveDoguRegistry{configRegistry{
		configReader{repo: repo},
		configWriter{repo: repo},
		configWatcher{repo: repo},
	}}, nil
}

type GlobalReader struct {
	configReader
}

type DoguReader struct {
	configReader
}

type SensitiveDoguReader struct {
	configReader
}

func NewGlobalConfigReader(ctx context.Context, k8sClient ConfigMapClient) (*GlobalReader, error) {
	repo, _ := newConfigRepo(withGeneralName(globalConfigMapName), createConfigMapClient(k8sClient, globalConfigType))

	if lErr := repo.write(ctx, config.CreateConfig(make(config.Entries))); lErr != nil {
		return nil, fmt.Errorf("could not create initial global config reader: %w", lErr)
	}

	return &GlobalReader{
		configReader{repo: repo},
	}, nil
}

func NewDoguConfigReader(ctx context.Context, doguName string, k8sClient ConfigMapClient) (*DoguReader, error) {
	repo, _ := newConfigRepo(withDoguName(doguName), createConfigMapClient(k8sClient, doguConfigType))

	if lErr := repo.write(ctx, config.CreateConfig(make(config.Entries))); lErr != nil {
		return nil, fmt.Errorf("could not create initial dogu config for reader %s: %w", doguName, lErr)
	}

	return &DoguReader{
		configReader{repo: repo},
	}, nil
}

func NewSensitiveDoguReader(ctx context.Context, doguName string, sc SecretClient) (*SensitiveDoguReader, error) {
	repo, _ := newConfigRepo(withDoguName(doguName), createSecretClient(sc, sensitiveConfigType))

	if lErr := repo.write(ctx, config.CreateConfig(make(config.Entries))); lErr != nil {
		return nil, fmt.Errorf("could not create initial sensitive dogu config %s: %w", doguName, lErr)
	}

	return &SensitiveDoguReader{
		configReader{repo: repo},
	}, nil
}
